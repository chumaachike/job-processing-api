package job

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", listJobs)
	r.Post("/", createJob)

	r.Get("/{id}", getJob)

	return r
}
