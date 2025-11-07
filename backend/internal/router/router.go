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
	ticketH := handlers.NewTicketHTTP(ticketRepo)

	// Reports (uses ticketRepo counters when available, else falls back)
	reportsH := handlers.NewReportsHTTP(ticketRepo)

	// Tickets
	r.Route("/api/tickets", func(r chi.Router) {
		// If you want to enforce auth for writes:
		// r.With(middleware.RequireAuth).Post("/", ticketH.Create())
		// r.Route("/{id}", func(r chi.Router) {
		//   r.Get("/", ticketH.Get())
		//   r.With(middleware.RequireAuth).Patch("/", ticketH.Update())
		//   r.With(middleware.RequireAuth).Post("/comments", ticketH.AddComment())
		//   return
		// })

		r.Get("/", ticketH.List())
		r.Post("/", ticketH.Create())
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", ticketH.Get())
			r.Patch("/", ticketH.Update())
			r.Post("/comments", ticketH.AddComment())
		})
	})

	// Reports
	r.Route("/api/reports", func(r chi.Router) {
		r.Get("/summary", reportsH.Summary())
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
