package job

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID        uuid.UUID       `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Status    JobStatus       `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
}

type CreateJobRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type JobMessage struct {
	JobId uuid.UUID `json:"id"`
	Type  string    `json:"type"`
}

type JobStatus string

const (
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusSucceeded JobStatus = "succeeded"
	JobStatusFailed    JobStatus = "failed"
)

type JobFilter struct {
	Type   string
	Status JobStatus
}

type UpdateJobStatusRequest struct {
	Status JobStatus `json:"status"`
}

var (
	ErrInvalidJob       = errors.New("invalid job")
	ErrInvalidJobFilter = errors.New("invalid job filter")
	ErrInvalidJobID     = errors.New("invalid job id")
	ErrJobNotFound      = errors.New("job not found")
	ErrInvalidJobStatus = errors.New("Invalid job status")
	ErrQueueFull        = errors.New("job queue full")
)
