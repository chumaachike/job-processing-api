package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/chumaachike/job-processing-api/internal/config"
	"github.com/chumaachike/job-processing-api/internal/database"
	"github.com/chumaachike/job-processing-api/internal/job"
	"github.com/chumaachike/job-processing-api/internal/server"
)

func main() {

	ctx := context.Background()

	//Load configurations
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.NewPostgres(ctx, cfg.DatabaseURL)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	// Initilize dependencies
	jobRepository := job.NewRepository(db)

	jobService := job.NewService(jobRepository)

	jobHandler := job.NewHandler(jobService)

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
	log.Printf("Server running on port %s", cfg.Port)

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
