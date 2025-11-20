package router

import (
	"context"
	"net/http"
	"strings"
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

	// Core middleware (order: recover -> logging -> cors -> body-limit -> rate-limit -> env -> auth)
	r.Use(middleware.Recoverer(log))
	r.Use(middleware.RequestLogger(log))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.Origin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
	r.Use(middleware.BodyLimit) // Limit request body size to 1MB
	// Global rate limit (applies to all routes except /api/auth which has its own)
	// Create the global rate limiter middleware
	globalRateLimit := httprate.LimitByIP(200, time.Minute)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip global rate limit for auth routes (they have their own stricter limit)
			if strings.HasPrefix(r.URL.Path, "/api/auth") {
				next.ServeHTTP(w, r)
				return
			}
			// Apply global rate limit for all other routes
			globalRateLimit(next).ServeHTTP(w, r)
		})
	})
	// Add environment to context for error sanitization
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, "env", cfg.Env)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
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
		r.With(middleware.RequireRoles("admin")).Post("/", userH.Create())
		r.With(middleware.RequireRoles("admin")).Get("/", userH.List())
		r.With(middleware.RequireRoles("admin")).Patch("/{id}/role", userH.UpdateRole())
		r.With(middleware.RequireRoles("admin")).Patch("/{id}/active", userH.SetActive())

		// Self-service (any authenticated user can update own basic info/password)
		r.With(middleware.RequireAuth).Patch("/{id}/basic", userH.UpdateBasic())
		r.With(middleware.RequireAuth).Patch("/{id}/password", userH.UpdatePassword())
	})

	// Auth with stricter rate limiting
	r.Route("/api/auth", func(r chi.Router) {
		// Rate limit for login/register/logout: 10 requests per minute (stricter for security)
		r.Route("/", func(r chi.Router) {
			r.Use(httprate.LimitByIP(10, time.Minute))
			r.Post("/register", authH.Register())
			r.Post("/login", authH.Login(cfg))
			r.Post("/logout", authH.Logout(cfg))
		})
		// Me endpoint: 60 requests per minute (more lenient for frequent auth checks)
		r.With(httprate.LimitByIP(60, time.Minute)).Get("/me", authH.Me())
	})

	return r
}
