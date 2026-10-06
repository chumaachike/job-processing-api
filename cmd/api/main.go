package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chumaachike/job-processing-api/internal/config"
	"github.com/chumaachike/job-processing-api/internal/database"
	"github.com/chumaachike/job-processing-api/internal/job"
	"github.com/chumaachike/job-processing-api/internal/queue"
	"github.com/chumaachike/job-processing-api/internal/server"
	"github.com/chumaachike/job-processing-api/internal/worker"
	appmetrics "github.com/chumaachike/job-processing-api/metrics"
)

func main() {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)
	slog.SetDefault(logger)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		return
	}

	// Connect to database
	db, err := database.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		return
	}
	defer db.Close()

	jobRepository := job.NewPostgresRepository(db)
	jobQueue := queue.NewJobQUeue(100)
	jobService := job.NewService(jobRepository, jobQueue)

	m := appmetrics.New()

	jobHandler := job.NewHandler(
		jobService,
		m,
		logger,
	)

	workerPool := worker.New(4, jobQueue)

	workerPool.Start(ctx)

	router := server.NewRouter(jobHandler)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// Start server
	go func() {
		slog.Info("server started", "port", cfg.Port)

		if err := srv.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			slog.Error("server failed", "error", err)
		}
	}()

	// Wait here until SIGINT or SIGTERM
	<-ctx.Done()

	slog.Info("shutdown signal received")

	// Give active requests time to finish
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("graceful shutdown failed", "error", err)
		return
	}

	slog.Info("server stopped gracefully")
}
