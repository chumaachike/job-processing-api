package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/chumaachike/job-processing-api/internal/config"
	"github.com/chumaachike/job-processing-api/internal/database"
	"github.com/chumaachike/job-processing-api/internal/job"
	"github.com/chumaachike/job-processing-api/internal/server"
	"github.com/chumaachike/job-processing-api/metrics"
)

func main() {

	ctx := context.Background()

	logger := slog.New(
		slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}),
	)
	slog.SetDefault(logger)

	//Load configurations
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", "err")
	}

	db, err := database.NewPostgres(ctx, cfg.DatabaseURL)

	if err != nil {
		slog.Error("failed to load database", "error", err)
	}

	defer db.Close()

	// Initilize dependencies
	jobRepository := job.NewPostgresRepository(db)

	jobService := job.NewService(jobRepository)

	metrics := metrics.New()

	jobHandler := job.NewHandler(jobService, metrics, logger)

	// Build router
	router := server.NewRouter(jobHandler)

	// Configure HTTP Server
	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// 6. Start server
	slog.Info("server started", "port", cfg.Port)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
