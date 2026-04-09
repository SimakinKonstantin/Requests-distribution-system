package balancer

import (
	"context"
	"log"
	"time"

	"github.com/hibiken/asynq"
)

// Run starts the balancer services according to cfg.BalancerRole ("all" | "event-handler" | "worker").
// It blocks until ctx is canceled or a component returns an error.
func Run(ctx context.Context, cfg Config) error {
	db, err := NewDB(ctx, cfg.BalancerPostgresDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	asynqClient := NewAsynqClient(cfg.RedisAddr)
	defer asynqClient.Close()

	updateSvc := NewBalancerUpdateService(db, asynqClient, cfg)
	matcher := NewMatcher(db, asynqClient, cfg)
	assigner := NewAssigner(db)
	workers := NewWorkers(cfg.RedisAddr, updateSvc, matcher, assigner)

	switch cfg.BalancerRole {
	case "event-handler":
		log.Printf("balancer: starting role=event-handler")
		svc := NewEventHandlerService(cfg, asynqClient)
		if err := svc.Run(ctx); err != nil && err != context.Canceled {
			return err
		}
		return nil
	case "worker":
		log.Printf("balancer: starting role=worker")
		EnqueueDistributionTickOnce(ctx, asynqClient)
		go matcher.RunTicker(ctx)
		if err := workers.Run(ctx); err != nil && err != context.Canceled {
			return err
		}
		return nil
	case "all":
		log.Printf("balancer: starting role=all")
		EnqueueDistributionTickOnce(ctx, asynqClient)

		svc := NewEventHandlerService(cfg, asynqClient)

		errCh := make(chan error, 2)
		go func() {
			go matcher.RunTicker(ctx)
			errCh <- workers.Run(ctx)
		}()
		go func() {
			errCh <- svc.Run(ctx)
		}()

		select {
		case <-ctx.Done():
			time.Sleep(200 * time.Millisecond)
			return ctx.Err()
		case err := <-errCh:
			if err != nil && err != context.Canceled {
				return err
			}
			return nil
		}
	default:
		return &ConfigError{Msg: `unknown ROLE (use: all|event-handler|worker)`}
	}
}

// StartInBackground starts the balancer in a goroutine and returns a channel
// with the terminal error (if any). Caller provides the already-loaded config.
func StartInBackground(parent context.Context, cfg Config) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(parent, cfg)
	}()
	return errCh
}

type ConfigError struct{ Msg string }

func (e *ConfigError) Error() string { return e.Msg }

var _ = asynq.MaxRetry // keep dependency explicit for go mod tidy stability
