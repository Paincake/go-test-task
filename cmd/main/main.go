package main

import (
	"context"
	"github.com/Paincake/go-test-task/internal/http-server/handlers/person/get"
	"github.com/Paincake/go-test-task/internal/http-server/handlers/person/save"
	"github.com/Paincake/go-test-task/internal/lib/api/filter"
	"github.com/Paincake/go-test-task/internal/lib/api/sort"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/exp/slog"

	"github.com/Paincake/go-test-task/internal/config"
	"github.com/Paincake/go-test-task/internal/database/postgres"
	"github.com/Paincake/go-test-task/internal/http-server/middleware/logger"
	"github.com/Paincake/go-test-task/internal/lib/logger/sl"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg, server := config.MustLoad()
	log := setupLogger(cfg.Env)
	log = log.With(slog.String("env", cfg.Env))
	log.Info("initializing server", slog.String("address", server.Address))
	log.Debug("initializing logger in debug mode")
	database, err := postgres.New(cfg.Name, cfg.User, cfg.Password, cfg.Host, cfg.Port)
	if err != nil {
		log.Error("failed to initialize database", sl.Err(err))
	}

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)
	router.Use(logger.New(log))
	router.Use(sort.Middleware)
	router.Use(filter.Middleware)

	router.Post("/person", save.Save(log, database))
	router.Get("/persons", get.GetAll(log, database))

	log.Info("starting server", slog.String("address", server.Address))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	srv := &http.Server{
		Addr:         server.Address,
		Handler:      router,
		ReadTimeout:  server.Timeout,
		WriteTimeout: server.Timeout,
		IdleTimeout:  server.IdleTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error("failed to start server")
		}
	}()

	log.Info("server started")
	<-done
	log.Info("stopping server")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))

		return
	}

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
	return log
}
