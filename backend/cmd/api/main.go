package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gh-ts/internal/config"
	"gh-ts/internal/database"
	"gh-ts/internal/router"
	"gh-ts/pkg/logger"
)

func main() {
	// config + logger
	cfg := config.Load()
	l := logger.New(cfg.Env)

	// db
	pool, err := database.Open(context.Background(), cfg)
	if err != nil {
		l.Fatal().Err(err).Msg("db connect failed")
	}
	defer pool.Close()

	// http
	r := router.New(l, pool, cfg)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		l.Info().Str("addr", srv.Addr).Msg("api listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Fatal().Err(err).Msg("server error")
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	l.Info().Msg("shutdown complete")
}
