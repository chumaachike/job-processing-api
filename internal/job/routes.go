package job

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Post("/", h.CreateJob)
	r.Get("/", h.ListJobs)
	r.Get("/{id}", h.GetJob)
	r.Patch("/{id}", h.UpdateJobStatus)

	return r
}
