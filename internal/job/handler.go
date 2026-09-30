package job

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req CreateJobRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		http.Error(w, "Request body must contain a single JSON object", http.StatusBadRequest)
		return
	}

	job, err := h.service.CreateJob(r.Context(), req)

	switch {
	case errors.Is(err, ErrInvalidJob):
		http.Error(w, "Invalid job", http.StatusBadRequest)
		return

	case err != nil:
		log.Printf("CreateJob internal error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, job)
}

func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	filter := JobFilter{
		Type:   r.URL.Query().Get("type"),
		Status: JobStatus(r.URL.Query().Get("status")),
	}

	jobs, err := h.service.ListJobs(r.Context(), filter)
	if err != nil {
		log.Printf("ListJobs internal error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, jobs)
}

func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, "Missing job id", http.StatusBadRequest)
		return
	}

	job, err := h.service.GetJob(r.Context(), id)

	switch {
	case errors.Is(err, ErrJobNotFound):
		http.Error(w, "Job not found", http.StatusNotFound)
		return

	case err != nil:
		log.Printf("GetJob internal error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, job)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
