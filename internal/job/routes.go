package job

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func Routes(h *Handler) http.Handler {
	r := chi.NewRouter()

	r.Post("/", h.Create)
	r.Get("/", h.listJobs)
	r.Get("/{id}", h.listJobs)

	return r
}
