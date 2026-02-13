package router

import (
	"net/http"
	"time"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/config"
	"github.com/cfegela/flyhalf/internal/handler"
	"github.com/cfegela/flyhalf/internal/middleware"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"golang.org/x/time/rate"
)

type Router struct {
	healthHandler    *handler.HealthHandler
	metricsHandler   *handler.MetricsHandler
	authHandler      *handler.AuthHandler
	adminHandler     *handler.AdminHandler
	teamHandler      *handler.TeamHandler
	leagueHandler    *handler.LeagueHandler
	ticketHandler    *handler.TicketHandler
	projectHandler   *handler.ProjectHandler
	sprintHandler    *handler.SprintHandler
	retroItemHandler *handler.RetroItemHandler
	authMiddleware   *auth.AuthMiddleware
	cfg              *config.Config
}

func New(
	healthHandler *handler.HealthHandler,
	metricsHandler *handler.MetricsHandler,
	authHandler *handler.AuthHandler,
	adminHandler *handler.AdminHandler,
	teamHandler *handler.TeamHandler,
	leagueHandler *handler.LeagueHandler,
	ticketHandler *handler.TicketHandler,
	projectHandler *handler.ProjectHandler,
	sprintHandler *handler.SprintHandler,
	retroItemHandler *handler.RetroItemHandler,
	authMiddleware *auth.AuthMiddleware,
	cfg *config.Config,
) *Router {
	return &Router{
		healthHandler:    healthHandler,
		metricsHandler:   metricsHandler,
		authHandler:      authHandler,
		adminHandler:     adminHandler,
		teamHandler:      teamHandler,
		leagueHandler:    leagueHandler,
		ticketHandler:    ticketHandler,
		projectHandler:   projectHandler,
		sprintHandler:    sprintHandler,
		retroItemHandler: retroItemHandler,
		authMiddleware:   authMiddleware,
		cfg:              cfg,
	}
}

func (rt *Router) Setup() http.Handler {
	r := chi.NewRouter()

	// Global middleware applied to all routes
	r.Use(chiMiddleware.Logger)      // Log all requests
	r.Use(chiMiddleware.Recoverer)   // Recover from panics
	r.Use(chiMiddleware.RequestID)   // Add unique request ID for tracing
	r.Use(chiMiddleware.RealIP)      // Get real client IP (handles proxy headers)
	r.Use(middleware.RequestSizeLimit(1 * 1024 * 1024)) // Limit request body to 1MB
	r.Use(middleware.SecurityHeaders) // Add security headers (X-Frame-Options, etc.)

	// CORS configuration - allows the frontend to make cross-origin requests
	// Configured via ALLOWED_ORIGIN environment variable (default: http://localhost:3000)
	// Production should be set to https://demo.flyhalf.app
	r.Use(middleware.CORS(&middleware.CORSConfig{
		AllowedOrigins: rt.cfg.Server.AllowedOrigins,
	}))

	// Rate limiter for authentication endpoints (5 requests per second, burst of 10)
	authRateLimiter := middleware.NewRateLimiter(rate.Limit(5), 10)

	r.Get("/health", rt.healthHandler.Check)
	r.Get("/metrics", rt.metricsHandler.GetMetrics)

	r.Route("/api/v1", func(r chi.Router) {
		// Add 10 second timeout for all API requests
		r.Use(middleware.Timeout(10 * time.Second))

		r.Route("/auth", func(r chi.Router) {
			// Apply rate limiting to login and refresh endpoints
			r.With(authRateLimiter.Limit).Post("/login", rt.authHandler.Login)
			r.With(authRateLimiter.Limit).Post("/refresh", rt.authHandler.Refresh)

			r.Group(func(r chi.Router) {
				r.Use(rt.authMiddleware.Authenticate)
				r.Post("/logout", rt.authHandler.Logout)
				r.Get("/me", rt.authHandler.Me)
				r.Put("/password", rt.authHandler.ChangePassword)
			})
		})

		r.Route("/tickets", func(r chi.Router) {
			r.Use(rt.authMiddleware.Authenticate)

			r.Get("/", rt.ticketHandler.ListTickets)
			r.Post("/", rt.ticketHandler.CreateTicket)
			r.Get("/{id}", rt.ticketHandler.GetTicket)
			r.Put("/{id}", rt.ticketHandler.UpdateTicket)
			r.Delete("/{id}", rt.ticketHandler.DeleteTicket)
			r.Post("/{id}/promote", rt.ticketHandler.PromoteTicket)
			r.Patch("/{id}/priority", rt.ticketHandler.UpdateTicketPriority)
			r.Patch("/{id}/sprint-order", rt.ticketHandler.UpdateTicketSprintOrder)
			r.Patch("/{id}/acceptance-criteria/{criteriaId}", rt.ticketHandler.UpdateAcceptanceCriteriaCompleted)
			r.Delete("/{id}/acceptance-criteria/{criteriaId}", rt.ticketHandler.DeleteAcceptanceCriteria)
			r.Post("/{id}/updates", rt.ticketHandler.CreateTicketUpdate)
			r.Delete("/{id}/updates/{updateId}", rt.ticketHandler.DeleteTicketUpdate)
		})

		r.Route("/projects", func(r chi.Router) {
			r.Use(rt.authMiddleware.Authenticate)

			r.Get("/", rt.projectHandler.ListProjects)
			r.Post("/", rt.projectHandler.CreateProject)
			r.Get("/{id}", rt.projectHandler.GetProject)
			r.Put("/{id}", rt.projectHandler.UpdateProject)
			r.Delete("/{id}", rt.projectHandler.DeleteProject)
		})

		r.Route("/sprints", func(r chi.Router) {
			r.Use(rt.authMiddleware.Authenticate)

			r.Get("/", rt.sprintHandler.ListSprints)
			r.Post("/", rt.sprintHandler.CreateSprint)
			r.Get("/{id}", rt.sprintHandler.GetSprint)
			r.Get("/{id}/report", rt.sprintHandler.GetSprintReport)
			r.Put("/{id}", rt.sprintHandler.UpdateSprint)
			r.Delete("/{id}", rt.sprintHandler.DeleteSprint)
			r.Post("/{id}/close", rt.sprintHandler.CloseSprint)
			r.Get("/{sprintId}/retro", rt.retroItemHandler.ListRetroItems)
			r.Post("/{sprintId}/retro", rt.retroItemHandler.CreateRetroItem)
		})

		r.Route("/retro-items", func(r chi.Router) {
			r.Use(rt.authMiddleware.Authenticate)

			r.Put("/{id}", rt.retroItemHandler.UpdateRetroItem)
			r.Delete("/{id}", rt.retroItemHandler.DeleteRetroItem)
			r.Post("/{id}/vote", rt.retroItemHandler.VoteRetroItem)
			r.Delete("/{id}/vote", rt.retroItemHandler.UnvoteRetroItem)
		})

		// Public endpoint for getting users for assignment (all authenticated users)
		r.Group(func(r chi.Router) {
			r.Use(rt.authMiddleware.Authenticate)
			r.Get("/users", rt.adminHandler.ListUsersForAssignment)
		})

		r.Route("/admin", func(r chi.Router) {
			r.Use(rt.authMiddleware.Authenticate)
			r.Use(rt.authMiddleware.RequireRole(model.RoleAdmin))

			r.Route("/users", func(r chi.Router) {
				r.Get("/", rt.adminHandler.ListUsers)
				r.Post("/", rt.adminHandler.CreateUser)
				r.Get("/{id}", rt.adminHandler.GetUser)
				r.Put("/{id}", rt.adminHandler.UpdateUser)
				r.Delete("/{id}", rt.adminHandler.DeleteUser)
			})

			r.Route("/teams", func(r chi.Router) {
				r.Get("/", rt.teamHandler.ListTeams)
				r.Post("/", rt.teamHandler.CreateTeam)
				r.Get("/{id}", rt.teamHandler.GetTeam)
				r.Put("/{id}", rt.teamHandler.UpdateTeam)
				r.Delete("/{id}", rt.teamHandler.DeleteTeam)
			})

			r.Route("/leagues", func(r chi.Router) {
				r.Get("/", rt.leagueHandler.ListLeagues)
				r.Post("/", rt.leagueHandler.CreateLeague)
				r.Get("/{id}", rt.leagueHandler.GetLeague)
				r.Put("/{id}", rt.leagueHandler.UpdateLeague)
				r.Delete("/{id}", rt.leagueHandler.DeleteLeague)
			})

			r.Post("/reset-demo", rt.adminHandler.ResetDemo)
			r.Post("/reseed-demo", rt.adminHandler.ReseedDemo)
		})
	})

	return r
}
