package balancer

import (
	"context"
	"log"

	"github.com/hibiken/asynq"
)

type Workers struct {
	Server *asynq.Server
	Mux    *asynq.ServeMux
}

func NewAsynqClient(redisAddr string) *asynq.Client {
	return asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
}

func NewWorkers(redisAddr string, balancerUpdate *BalancerUpdateService, matcher *Matcher, assigner *Assigner) *Workers {
	srv := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{
			Concurrency: 10, // Сколько job'ов выполняется конкурентно.
			Queues: map[string]int{
				"state-high":   6,
				"state-medium": 3,
				"state-low":    1,
				"assign":       4,
				"dist":         2,
			},
		},
	)

	mux := asynq.NewServeMux()
	mux.HandleFunc(TaskTypeBatchUpdate, balancerUpdate.HandleBatchUpdateTask)
	mux.HandleFunc(TaskTypeDistributionTick, matcher.HandleDistributionTick)
	mux.HandleFunc(TaskTypeAssign, assigner.HandleAssignTask)

	return &Workers{Server: srv, Mux: mux}
}

func (w *Workers) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		if err := w.Server.Run(w.Mux); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		w.Server.Shutdown()
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}

func EnqueueDistributionTickOnce(ctx context.Context, c *asynq.Client) {
	_, err := c.EnqueueContext(ctx, NewDistributionTickTask(), asynq.Queue("dist"), asynq.MaxRetry(0))
	if err != nil {
		log.Printf("enqueue initial distribution tick failed: %v", err)
	}
}
