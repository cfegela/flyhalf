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
	authHandler     *handler.AuthHandler
	adminHandler    *handler.AdminHandler
	resourceHandler *handler.ResourceHandler
	authMiddleware  *auth.AuthMiddleware
	cfg             *config.Config
}

func New(
	authHandler *handler.AuthHandler,
	adminHandler *handler.AdminHandler,
	resourceHandler *handler.ResourceHandler,
	authMiddleware *auth.AuthMiddleware,
	cfg *config.Config,
) *Router {
	return &Router{
		authHandler:     authHandler,
		adminHandler:    adminHandler,
		resourceHandler: resourceHandler,
		authMiddleware:  authMiddleware,
		cfg:             cfg,
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
			})
		})

		r.Route("/resources", func(r chi.Router) {
			r.Use(rt.authMiddleware.Authenticate)

			r.Get("/", rt.resourceHandler.ListResources)
			r.Post("/", rt.resourceHandler.CreateResource)
			r.Get("/{id}", rt.resourceHandler.GetResource)
			r.Put("/{id}", rt.resourceHandler.UpdateResource)
			r.Delete("/{id}", rt.resourceHandler.DeleteResource)
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
