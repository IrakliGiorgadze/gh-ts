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

	r.Use(middleware.WithAuth(log, cfg))

	// Repos & services
	userRepo := postgres.NewUserRepo(db)
	authSvc := service.NewAuthService(userRepo, cfg.SessionSecret)
	auth := handlers.NewAuthHTTP(authSvc)

	r.Use(middleware.RequestLogger(log))
	r.Use(middleware.Recoverer(log))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.Origin},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))
	r.Use(httprate.LimitByIP(200, time.Minute))

	// Health
	r.Get("/healthz", handlers.Health())
	r.Get("/api/healthz", handlers.Health())

	// Repos + handlers
	ticketRepo := postgres.NewTicketRepo(db)
	th := handlers.NewTicketHTTP(ticketRepo)

	r.Route("/api/tickets", func(r chi.Router) {
		r.Get("/", th.List())
		r.Post("/", th.Create())
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", th.Get())
			r.Patch("/", th.Update())
			r.Post("/comments", th.AddComment())
		})
	})

	// Auth routes
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", auth.Register())
		r.Post("/login", auth.Login(cfg.SessionSecret))
		r.Post("/logout", auth.Logout())
		r.Get("/me", auth.Me())
	})

	return r
}
