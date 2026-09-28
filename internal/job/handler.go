package job

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req CreateJobRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		log.Printf("failed to decode job, %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	job, err := h.service.Create(r.Context(), req)

	if errors.Is(err, ErrInvalidJob) {
		log.Printf("CreateJob failed: %v", err)
		http.Error(w, "Invalid job", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Printf("CreateJob internal error: %v", err)

		http.Error(
			w,
			"Internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(job)
}

func (h *Handler) listJobs(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("List jobs"))
}

func (h *Handler) getJob(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Get job"))
}
