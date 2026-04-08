package main

import (
	"context"
	"log"
	"net/http"
	"os"

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

	// Optional: run balancer subsystem inside the same process.
	// ENABLE_BALANCER=1 starts it in background; balancer is configured via its own env vars
	// (ROLE, POSTGRES_DSN, REDIS_ADDR, RABBIT_URL, etc).
	//
	// NOTE: balancer uses its own Postgres DSN env (POSTGRES_DSN). You can point it to the same DB as CRUD-service.
	var balancerErrCh <-chan error
	if os.Getenv("ENABLE_BALANCER") == "1" {
		ctx := context.Background()
		balancerErrCh = balancer.StartFromEnvInBackground(ctx)
		log.Printf("balancer enabled (ENABLE_BALANCER=1)")
	}

	// Database connection
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
	workflowRepo := workflow.NewWorkflowRepository(database)

	// Services
	employeeSvc := service.NewEmployeeService(database, employeeRepo, slotRepo)
	slotSvc := service.NewSlotService(database, slotRepo)
	subthemeSvc := service.NewSubthemeService(database, subthemeRepo)
	clientSvc := service.NewClientService(database, clientRepo)
	themeSvc := service.NewThemeService(database, themeRepo)
	teamSvc := service.NewTeamService(database, teamRepo)
	workflowSvc := workflow.NewWorkflowService(workflowRepo, teamSvc)
	appealSvc := service.NewAppealService(database, appealRepo, teamRepo, clientRepo, slotRepo, workflowSvc, teamSvc)

	// Handler & routes
	h := handler.New(employeeSvc, slotSvc, appealSvc, subthemeSvc, clientSvc, themeSvc, teamSvc, workflowSvc)
	router := h.InitRoutes()

	log.Printf("Starting server on %s", cfg.ServerAddr)

	// If balancer is enabled, we want to crash the process if it fails.
	if balancerErrCh != nil {
		go func() {
			if err := <-balancerErrCh; err != nil && err != context.Canceled {
				log.Fatalf("balancer stopped: %v", err)
			}
		}()
	}

	if err = http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
