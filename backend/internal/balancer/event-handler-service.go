package balancer

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/hibiken/asynq"
	amqp "github.com/rabbitmq/amqp091-go"
)

type EventHandlerService struct {
	cfg   Config
	asynq *asynq.Client

	mu        sync.Mutex
	batch     map[string]ProcessedEvent
	lastFlush time.Time
}

func NewEventHandlerService(cfg Config, asynqClient *asynq.Client) *EventHandlerService {
	return &EventHandlerService{
		cfg:       cfg,
		asynq:     asynqClient,
		batch:     make(map[string]ProcessedEvent, cfg.BatchSize),
		lastFlush: time.Now(),
	}
}

func (s *EventHandlerService) Run(ctx context.Context) error {
	conn, err := amqp.Dial(s.cfg.RabbitURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		s.cfg.RabbitQueue,
		true,  // постоянная очередь
		false, // без автоудаления
		false, // не эксклюзивная
		false, // без ожидания ответа
		nil,
	)
	if err != nil {
		return err
	}

	if err := ch.Qos(50, 0, false); err != nil {
		return err
	}

	msgs, err := ch.Consume(
		s.cfg.RabbitQueue,
		"",
		false, // ручное подтверждение
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(s.cfg.BatchTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.flush(ctx)
			return ctx.Err()
		case <-ticker.C:
			s.flush(ctx)
		case msg, ok := <-msgs:
			if !ok {
				return context.Canceled
			}
			if err := s.handleDelivery(ctx, msg); err != nil {
				log.Printf("event-handler: handle failed: %v", err)
				_ = msg.Nack(false, true)
				continue
			}
			_ = msg.Ack(false)
		}
	}
}

func (s *EventHandlerService) handleDelivery(ctx context.Context, msg amqp.Delivery) error {
	var e RabbitEvent
	if err := json.Unmarshal(msg.Body, &e); err != nil {
		return err
	}

	pe := ProcessedEvent{
		RabbitEvent: e,
		ProcessedAt: time.Now().UTC(),
	}
	s.addEventToBatch(ctx, pe)
	return nil
}

func (s *EventHandlerService) addEventToBatch(ctx context.Context, e ProcessedEvent) {
	key := s.createJobName(e)

	s.mu.Lock()
	s.batch[key] = e
	shouldFlush := len(s.batch) >= s.cfg.BatchSize
	s.mu.Unlock()

	if shouldFlush {
		s.flush(ctx)
	}
}

func (s *EventHandlerService) flush(ctx context.Context) {
	s.mu.Lock()
	if len(s.batch) == 0 {
		s.mu.Unlock()
		return
	}
	events := make([]ProcessedEvent, 0, len(s.batch))
	for _, v := range s.batch {
		events = append(events, v)
	}
	s.batch = make(map[string]ProcessedEvent, s.cfg.BatchSize)
	s.lastFlush = time.Now()
	s.mu.Unlock()

	grouped := groupEventsByType(events)
	for batchType, events := range grouped {
		if len(events) == 0 {
			continue
		}
		task, err := NewBatchUpdateTask(BatchUpdatePayload{Type: batchType, Events: events})
		if err != nil {
			log.Printf("event-handler: make task failed: %v", err)
			continue
		}
		_, err = s.asynq.EnqueueContext(ctx, task,
			asynq.MaxRetry(3),
			asynq.Queue(queueForBatchType(batchType)),
		)
		if err != nil {
			log.Printf("event-handler: enqueue failed: %v", err)
		}
	}
}

func (s *EventHandlerService) createJobName(e ProcessedEvent) string {
	return string(e.Name) + "_" + strconv.Itoa(e.Payload.AppealID)
}

func groupEventsByType(events []ProcessedEvent) map[JobBatchType][]ProcessedEvent {
	out := map[JobBatchType][]ProcessedEvent{
		BatchAppealStatusChanges: {},

		// Этот ивент не используется.
		BatchManagerAvailabilityChange: {},
		BatchDistributionRequests:      {},
	}
	for _, e := range events {
		switch e.Name {
		case EventAppealNeedsDistribution:
			out[BatchDistributionRequests] = append(out[BatchDistributionRequests], e)
		default:
			out[BatchDistributionRequests] = append(out[BatchDistributionRequests], e)
		}
	}
	return out
}

func queueForBatchType(t JobBatchType) string {
	switch t {
	case BatchAppealStatusChanges:
		return "state-high"

	// Этот ивент не используется.
	case BatchManagerAvailabilityChange:
		return "state-medium"
	case BatchDistributionRequests:
		return "state-low"
	default:
		return "state-low"
	}
}
