package router

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"gh-ts/internal/config"
	"gh-ts/internal/handlers"
	"gh-ts/internal/middleware"
	"gh-ts/internal/repository/postgres"
	"gh-ts/internal/service"
)

func New(log zerolog.Logger, db *pgxpool.Pool, cfg config.Config) http.Handler {
	r := chi.NewRouter()

	// Core middleware (order: recover -> logging -> cors -> rate-limit -> auth)
	r.Use(middleware.Recoverer(log))
	r.Use(middleware.RequestLogger(log))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.Origin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
	r.Use(httprate.LimitByIP(200, time.Minute))
	r.Use(middleware.WithAuth(log, cfg)) // attaches user id/role to context if cookie present

	// Health
	r.Get("/healthz", handlers.Health())
	r.Get("/api/healthz", handlers.Health())

	// Repos & services
	userRepo := postgres.NewUserRepo(db)
	authSvc := service.NewAuthService(userRepo, cfg.SessionSecret)
	authH := handlers.NewAuthHTTP(authSvc, userRepo)

	ticketRepo := postgres.NewTicketRepo(db)
	// Pass userRepo into TicketHTTP for auto-assignment logic
	ticketH := handlers.NewTicketHTTP(ticketRepo, userRepo)

	// Reports (uses ticketRepo counters when available, else falls back)
	reportsH := handlers.NewReportsHTTP(ticketRepo)

	// Tickets (RBAC-enforced)
	r.Route("/api/tickets", func(r chi.Router) {
		// List is open (optionally protect with RequireAuth)
		r.Get("/", ticketH.List())

		// Create requires authentication
		r.With(middleware.RequireAuth).Post("/", ticketH.Create())

		r.Route("/{id}", func(r chi.Router) {
			// Get single ticket
			r.Get("/", ticketH.Get())

			// Update restricted to admin/agent/supervisor
			r.With(middleware.RequireRoles("admin", "agent", "supervisor")).
				Patch("/", ticketH.Update())

			// Comments allowed for authenticated users
			r.With(middleware.RequireAuth).
				Post("/comments", ticketH.AddComment())
		})
	})

	// Reports
	r.Route("/api/reports", func(r chi.Router) {
		r.Get("/summary", reportsH.Summary())
	})

	// Users (admin-only listing & admin ops; self-service updates require auth)
	userH := handlers.NewUserHTTP(userRepo)
	r.Route("/api/users", func(r chi.Router) {
		// Admin-only endpoints
		r.With(middleware.RequireRoles("admin")).Get("/", userH.List())
		r.With(middleware.RequireRoles("admin")).Patch("/{id}/role", userH.UpdateRole())
		r.With(middleware.RequireRoles("admin")).Patch("/{id}/active", userH.SetActive())

		// Self-service (any authenticated user can update own basic info/password)
		r.With(middleware.RequireAuth).Patch("/{id}/basic", userH.UpdateBasic())
		r.With(middleware.RequireAuth).Patch("/{id}/password", userH.UpdatePassword())
	})

	// Auth
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authH.Register())
		r.Post("/login", authH.Login(cfg.SessionSecret))
		r.Post("/logout", authH.Logout())
		r.Get("/me", authH.Me())
	})

	return r
}
