package main

import (
	"log"
	"net/http"

	"crud-service/internal/config"
	"crud-service/internal/db"
	"crud-service/internal/handler"
	"crud-service/internal/repository"
	"crud-service/internal/service"
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

	// Services
	employeeSvc := service.NewEmployeeService(employeeRepo, slotRepo)
	slotSvc := service.NewSlotService(slotRepo)
	appealSvc := service.NewAppealService(appealRepo)
	subthemeSvc := service.NewSubthemeService(subthemeRepo)
	clientSvc := service.NewClientService(clientRepo)
	themeSvc := service.NewThemeService(themeRepo)

	// Handler & routes
	h := handler.New(employeeSvc, slotSvc, appealSvc, subthemeSvc, clientSvc, themeSvc)
	router := h.InitRoutes()

	log.Printf("Starting server on %s", cfg.ServerAddr)
	if err = http.ListenAndServe(cfg.ServerAddr, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
