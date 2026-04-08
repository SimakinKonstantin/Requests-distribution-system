package balancer

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
)

// Run starts the balancer services according to cfg.Role ("all" | "event-handler" | "worker").
// It blocks until ctx is canceled or a component returns an error.
func Run(ctx context.Context, cfg Config) error {
	db, err := NewDB(ctx, cfg.PostgresDSN)
	if err != nil {
		return err
	}
	defer db.Close()

	if err := applyMigrations(ctx, db); err != nil {
		return err
	}

	asynqClient := NewAsynqClient(cfg.RedisAddr)
	defer asynqClient.Close()

	updateSvc := NewBalancerUpdateService(db, asynqClient, cfg)
	matcher := NewMatcher(db, asynqClient, cfg)
	assigner := NewAssigner(db)
	workers := NewWorkers(cfg.RedisAddr, updateSvc, matcher, assigner)

	switch cfg.Role {
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

// RunFromEnv is a convenience wrapper around LoadConfig + Run.
func RunFromEnv(ctx context.Context) error {
	return Run(ctx, LoadConfig())
}

// StartFromEnvInBackground starts the balancer in a goroutine and returns a channel with the terminal error (if any).
// Useful for integrating into another main() without blocking.
func StartFromEnvInBackground(parent context.Context) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- RunFromEnv(parent)
	}()
	return errCh
}

// WithSignalCancel creates a context that cancels on SIGINT/SIGTERM (Windows supports os.Interrupt).
func WithSignalCancel(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	go func() {
		<-sig
		cancel()
	}()
	return ctx, cancel
}

type ConfigError struct{ Msg string }

func (e *ConfigError) Error() string { return e.Msg }

var _ = asynq.MaxRetry // keep dependency explicit for go mod tidy stability

