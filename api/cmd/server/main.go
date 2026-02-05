package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/config"
	"github.com/cfegela/flyhalf/internal/database"
	"github.com/cfegela/flyhalf/internal/handler"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/cfegela/flyhalf/internal/router"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.JWT.AccessSecret == "" || cfg.JWT.RefreshSecret == "" {
		return fmt.Errorf("JWT secrets must be set")
	}

	// Set bcrypt cost from configuration
	auth.SetBcryptCost(cfg.Security.BcryptCost)

	ctx := context.Background()

	pool, err := database.Connect(&cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer pool.Close()

	log.Println("Running database migrations...")
	if err := database.RunMigrations(ctx, pool); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	log.Println("Database migrations completed successfully")

	userRepo := model.NewUserRepository(pool)
	teamRepo := model.NewTeamRepository(pool)
	ticketRepo := model.NewTicketRepository(pool)
	projectRepo := model.NewProjectRepository(pool)
	sprintRepo := model.NewSprintRepository(pool)
	retroItemRepo := model.NewRetroItemRepository(pool)
	criteriaRepo := model.NewAcceptanceCriteriaRepository(pool)
	updateRepo := model.NewTicketUpdateRepository(pool)

	jwtService := auth.NewJWTService(&cfg.JWT)
	authMiddleware := auth.NewAuthMiddleware(jwtService)

	isProduction := cfg.Server.Environment == "production"
	healthHandler := handler.NewHealthHandler(pool)
	metricsHandler := handler.NewMetricsHandler(pool)
	authHandler := handler.NewAuthHandler(userRepo, jwtService, isProduction)
	adminHandler := handler.NewAdminHandler(userRepo, ticketRepo, sprintRepo, projectRepo)
	teamHandler := handler.NewTeamHandler(teamRepo)
	ticketHandler := handler.NewTicketHandler(ticketRepo, criteriaRepo, updateRepo, pool)
	projectHandler := handler.NewProjectHandler(projectRepo)
	sprintHandler := handler.NewSprintHandler(sprintRepo, ticketRepo)
	retroItemHandler := handler.NewRetroItemHandler(retroItemRepo, userRepo)

	rt := router.New(healthHandler, metricsHandler, authHandler, adminHandler, teamHandler, ticketHandler, projectHandler, sprintHandler, retroItemHandler, authMiddleware, cfg)
	httpHandler := rt.Setup()

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      httpHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("Received signal %v, starting graceful shutdown", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			srv.Close()
			return fmt.Errorf("failed to gracefully shutdown server: %w", err)
		}

		log.Println("Server stopped gracefully")
	}

	return nil
}
