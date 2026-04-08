package balancer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
)

type BalancerUpdateService struct {
	db   *DB
	asynq *asynq.Client
	cfg  Config
}

func NewBalancerUpdateService(db *DB, asynqClient *asynq.Client, cfg Config) *BalancerUpdateService {
	return &BalancerUpdateService{db: db, asynq: asynqClient, cfg: cfg}
}

func (s *BalancerUpdateService) HandleBatchUpdateTask(ctx context.Context, t *asynq.Task) error {
	var p BatchUpdatePayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return err
	}

	switch p.Type {
	case BatchAppealStatusChanges:
		return s.processBatchAppealUpdates(ctx, p.Events)
	case BatchDistributionRequests:
		// For demo we piggy-back on appeal updates: ensure appeal is in pending set.
		return s.processBatchAppealUpdates(ctx, p.Events)
	case BatchManagerAvailabilityChange:
		// Not used in simplified demo.
		return nil
	default:
		return nil
	}
}

func (s *BalancerUpdateService) processBatchAppealUpdates(ctx context.Context, events []ProcessedEvent) error {
	// Deduplicate by appealId (similar to BaseStateWorker.processBatchAppealUpdates).
	unique := make(map[int]ProcessedEvent, len(events))
	for _, e := range events {
		if e.Payload.AppealID != 0 {
			unique[e.Payload.AppealID] = e
		}
	}

	var errs []error
	for appealID, e := range unique {
		switch e.Name {
		case EventAppealNeedsDistribution:
			if err := s.db.UpsertPendingAppealByID(ctx, appealID); err != nil {
				log.Printf("balancer-update: upsert pending appeal %d failed: %v", appealID, err)
				errs = append(errs, fmt.Errorf("appeal %d needs_distribution: %w", appealID, err))
			}
		case EventAppealClosed:
			if err := s.db.CloseAppeal(ctx, appealID); err != nil {
				log.Printf("balancer-update: close appeal %d failed: %v", appealID, err)
				errs = append(errs, fmt.Errorf("appeal %d closed: %w", appealID, err))
			}
		default:
			// ignore
		}
	}

	return errors.Join(errs...)
}

