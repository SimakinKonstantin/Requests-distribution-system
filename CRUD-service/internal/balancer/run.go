package balancer

import (
	"context"
	"log"
	"time"

	"crud-service/internal/crud/repository"
	"crud-service/internal/crud/service"

	"github.com/hibiken/asynq"
)

type Services struct {
	AppealService  service.AppealService
	SlotService    service.SlotService
	EmployeeRepo   repository.EmployeeRepository
}

func Run(ctx context.Context, cfg Config, svc Services) error {
	asynqClient := NewAsynqClient(cfg.RedisAddr)
	defer asynqClient.Close()

	updateSvc := NewBalancerUpdateService(svc.AppealService, asynqClient)
	matcher := NewMatcher(asynqClient, cfg, svc.AppealService, svc.EmployeeRepo, svc.SlotService)
	assigner := NewAssigner(svc.AppealService)
	workers := NewWorkers(cfg.RedisAddr, updateSvc, matcher, assigner)

	switch cfg.BalancerRole {
	case "event-handler":
		log.Printf("balancer: starting role=event-handler")
		eventSvc := NewEventHandlerService(cfg, asynqClient)
		if err := eventSvc.Run(ctx); err != nil && err != context.Canceled {
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

		eventSvc := NewEventHandlerService(cfg, asynqClient)

		errCh := make(chan error, 2)
		go func() {
			go matcher.RunTicker(ctx)
			errCh <- workers.Run(ctx)
		}()
		go func() {
			errCh <- eventSvc.Run(ctx)
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

// StartInBackground запускает балансировщик в горутине и возвращает канал с ошибкой завершения.
func StartInBackground(parent context.Context, cfg Config, svc Services) <-chan error {
	errCh := make(chan error, 1)
	go func() {
		errCh <- Run(parent, cfg, svc)
	}()
	return errCh
}

type ConfigError struct{ Msg string }

func (e *ConfigError) Error() string { return e.Msg }

var _ = asynq.MaxRetry // keep dependency explicit for go mod tidy stability
