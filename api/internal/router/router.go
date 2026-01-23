package router

import (
	"net/http"

	"github.com/cfegela/flyhalf/internal/auth"
	"github.com/cfegela/flyhalf/internal/config"
	"github.com/cfegela/flyhalf/internal/handler"
	"github.com/cfegela/flyhalf/internal/middleware"
	"github.com/cfegela/flyhalf/internal/model"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
)

type Router struct {
	authHandler    *handler.AuthHandler
	adminHandler   *handler.AdminHandler
	ticketHandler  *handler.TicketHandler
	epicHandler    *handler.EpicHandler
	sprintHandler  *handler.SprintHandler
	authMiddleware *auth.AuthMiddleware
	cfg            *config.Config
}

func New(
	authHandler *handler.AuthHandler,
	adminHandler *handler.AdminHandler,
	ticketHandler *handler.TicketHandler,
	epicHandler *handler.EpicHandler,
	sprintHandler *handler.SprintHandler,
	authMiddleware *auth.AuthMiddleware,
	cfg *config.Config,
) *Router {
	return &Router{
		authHandler:    authHandler,
		adminHandler:   adminHandler,
		ticketHandler:  ticketHandler,
		epicHandler:    epicHandler,
		sprintHandler:  sprintHandler,
		authMiddleware: authMiddleware,
		cfg:            cfg,
	}
}

func (rt *Router) Setup() http.Handler {
	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORS(&middleware.CORSConfig{
		AllowedOrigins: rt.cfg.Server.AllowedOrigins,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/login", rt.authHandler.Login)
			r.Post("/refresh", rt.authHandler.Refresh)

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
		})

		r.Route("/epics", func(r chi.Router) {
			r.Use(rt.authMiddleware.Authenticate)

			r.Get("/", rt.epicHandler.ListEpics)
			r.Post("/", rt.epicHandler.CreateEpic)
			r.Get("/{id}", rt.epicHandler.GetEpic)
			r.Put("/{id}", rt.epicHandler.UpdateEpic)
			r.Delete("/{id}", rt.epicHandler.DeleteEpic)
		})

		r.Route("/sprints", func(r chi.Router) {
			r.Use(rt.authMiddleware.Authenticate)

			r.Get("/", rt.sprintHandler.ListSprints)
			r.Post("/", rt.sprintHandler.CreateSprint)
			r.Get("/{id}", rt.sprintHandler.GetSprint)
			r.Put("/{id}", rt.sprintHandler.UpdateSprint)
			r.Delete("/{id}", rt.sprintHandler.DeleteSprint)
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
		})
	})

	return r
}
