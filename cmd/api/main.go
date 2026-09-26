package main

import (
	"log"
	"net/http"

	"github.com/chumaachike/job-processing-api/internal/server"
)

func main() {
	router := server.NewRouter()

	log.Println("Server is running on :8080")

	log.Fatal(
		http.ListenAndServe(":8080", router),
	)
}
