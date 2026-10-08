package job

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/chumaachike/job-processing-api/metrics"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
	metrics *metrics.Metrics
	logger  *slog.Logger
}

func NewHandler(service *Service, metrics *metrics.Metrics, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		metrics: metrics,
		logger:  logger,
	}
}

func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req CreateJobRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		h.logger.Error("invlid request", "Error", err)
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
	case errors.Is(err, ErrQueueFull):
		http.Error(w, "server busy, retry later", http.StatusServiceUnavailable)

	case err != nil:
		h.logger.Error("CreateJob internal error", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	h.metrics.JobCreated.Inc()
	writeJSON(w, http.StatusCreated, job, h.logger)
}

func (h *Handler) ListJobs(w http.ResponseWriter, r *http.Request) {
	filter := JobFilter{
		Type:   r.URL.Query().Get("type"),
		Status: JobStatus(r.URL.Query().Get("status")),
	}

	jobs, err := h.service.ListJobs(r.Context(), filter)
	if err != nil {
		h.logger.Error("ListJobs internal error", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, jobs, h.logger)
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
		h.logger.Error("Get job internal error", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, job, h.logger)
}

func (h *Handler) UpdateJobStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if id == "" {
		http.Error(w, "Missing job id", http.StatusBadRequest)
		return
	}

	var req UpdateJobStatusRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job, err := h.service.UpdateJobStatus(
		r.Context(),
		id,
		req.Status,
	)

	switch {
	case errors.Is(err, ErrInvalidJobID):
		http.Error(w, "Invalid job id", http.StatusBadRequest)
		return

	case errors.Is(err, ErrInvalidJobStatus):
		http.Error(w, "Invalid job status", http.StatusBadRequest)
		return

	case errors.Is(err, ErrJobNotFound):
		http.Error(w, "Job not found", http.StatusNotFound)
		return

	case err != nil:
		h.logger.Error("UpdateJobStatus internal error", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, job, h.logger)
}

func writeJSON(w http.ResponseWriter, status int, v any, logger *slog.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Error("failed to encode response", "error", err)
	}
}
