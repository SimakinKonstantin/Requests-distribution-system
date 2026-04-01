package main

import (
	"log"
	"net/http"

	"crud-service/internal/config"
	"crud-service/internal/crud/repository"
	"crud-service/internal/crud/service"
	"crud-service/internal/db"
	"crud-service/internal/handler"
	"crud-service/internal/workflow"
)

func main() {
	cfg := config.Load()

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
	appealSvc := service.NewAppealService(database, appealRepo, teamRepo, clientRepo, slotRepo)
	subthemeSvc := service.NewSubthemeService(database, subthemeRepo)
	clientSvc := service.NewClientService(database, clientRepo)
	themeSvc := service.NewThemeService(database, themeRepo)
	teamSvc := service.NewTeamService(database, teamRepo)
	workflowSvc := workflow.NewWorkflowService(workflowRepo)

	// Handler & routes
	h := handler.New(employeeSvc, slotSvc, appealSvc, subthemeSvc, clientSvc, themeSvc, teamSvc, workflowSvc)
	router := h.InitRoutes()

	log.Printf("Starting server on %s", cfg.ServerAddr)
	if err = http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
