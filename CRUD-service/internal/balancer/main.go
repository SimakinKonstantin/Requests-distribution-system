package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg := LoadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	db, err := NewDB(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("db init failed: %v", err)
	}
	defer db.Close()

	if err := applyMigrations(ctx, db); err != nil {
		log.Fatalf("apply migrations failed: %v", err)
	}

	asynqClient := NewAsynqClient(cfg.RedisAddr)
	defer asynqClient.Close()

	updateSvc := NewBalancerUpdateService(db, asynqClient, cfg)
	matcher := NewMatcher(db, asynqClient, cfg)
	assigner := NewAssigner(db)
	workers := NewWorkers(cfg.RedisAddr, updateSvc, matcher, assigner)

	switch cfg.Role {
	case "event-handler":
		log.Printf("starting role=event-handler")
		svc := NewEventHandlerService(cfg, asynqClient)
		if err := svc.Run(ctx); err != nil && err != context.Canceled {
			log.Fatalf("event-handler stopped: %v", err)
		}
	case "worker":
		log.Printf("starting role=worker")
		EnqueueDistributionTickOnce(ctx, asynqClient)
		go matcher.RunTicker(ctx)
		if err := workers.Run(ctx); err != nil && err != context.Canceled {
			log.Fatalf("workers stopped: %v", err)
		}
	case "all":
		log.Printf("starting role=all")
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
		case err := <-errCh:
			if err != nil && err != context.Canceled {
				log.Fatalf("service stopped: %v", err)
			}
		}
	default:
		log.Fatalf("unknown ROLE=%q (use: all|event-handler|worker)", cfg.Role)
	}
}

