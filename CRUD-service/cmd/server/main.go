package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"crud-service/internal/balancer"
	"crud-service/internal/config"
	"crud-service/internal/crud/repository"
	"crud-service/internal/crud/service"
	"crud-service/internal/db"
	"crud-service/internal/handler"
	"crud-service/internal/workflow"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown on SIGINT / SIGTERM
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		log.Println("shutdown signal received")
		cancel()
	}()

	// Database connection (sqlx, used by CRUD-service)
	database, err := db.New(cfg.ConnectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Repositories
	employeeRepo := repository.NewEmployeeRepository(database)
	slotRepo := repository.NewSlotRepository(database)
	appealRepo := repository.NewAppealRepository(database)
	subthemeRepo := repository.NewSubthemeRepository(database)
	clientRepo := repository.NewClientRepository(database)
	themeRepo := repository.NewThemeRepository(database)
	teamRepo := repository.NewTeamRepository(database)
	pendingAppealRepo := repository.NewPendingAppealRepository(database)
	workflowRepo := workflow.NewWorkflowRepository(database)

	// Services
	employeeSvc := service.NewEmployeeService(database, employeeRepo, slotRepo)
	slotSvc := service.NewSlotService(database, slotRepo)
	subthemeSvc := service.NewSubthemeService(database, subthemeRepo)
	clientSvc := service.NewClientService(database, clientRepo)
	themeSvc := service.NewThemeService(database, themeRepo)
	teamSvc := service.NewTeamService(database, teamRepo)
	workflowSvc := workflow.NewWorkflowService(workflowRepo, teamSvc)
	var appealEventPub service.BalancerEventPublisher
	if cfg.RabbitURL != "" {
		appealEventPub = balancer.NewRabbitPublisher(cfg.RabbitURL, cfg.RabbitQueue)
	}
	appealSvc := service.NewAppealService(database, appealRepo, teamRepo, clientRepo, slotRepo, pendingAppealRepo, &workflowSvc, teamSvc, appealEventPub)

	// Handler & routes
	h := handler.New(employeeSvc, slotSvc, appealSvc, subthemeSvc, clientSvc, themeSvc, teamSvc, workflowSvc)
	router := h.InitRoutes()

	// Balancer subsystem - запускается всегда когда задан RABBIT_URL.
	// Если RABBIT_URL не задан (локальная разработка без RabbitMQ), балансировщик пропускается.
	if cfg.RabbitURL == "" {
		log.Println("balancer: RABBIT_URL not set, skipping balancer startup")
	} else {
		balancerSvc := balancer.Services{
			AppealService: appealSvc,
			SlotService:   slotSvc,
			EmployeeRepo:  employeeRepo,
		}
		balancerErrCh := balancer.StartInBackground(ctx, *cfg, balancerSvc)
		go func() {
			if err := <-balancerErrCh; err != nil && err != context.Canceled {
				log.Fatalf("balancer stopped with error: %v", err)
			}
		}()
		log.Printf("balancer: started (role=%s, queue=%s)", cfg.BalancerRole, cfg.RabbitQueue)
	}

	log.Printf("Starting server on %s", cfg.ServerAddr)
	if err = http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
