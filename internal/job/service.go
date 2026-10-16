package job

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

type Publisher interface {
	Publish(context.Context, JobMessage) error
}

type Repository interface {
	CreateJob(context.Context, CreateJobRequest) (JobMessage, error)
	ListJobs(context.Context, JobFilter) ([]Job, error)
	GetJob(context.Context, uuid.UUID) (Job, error)
	UpdateJob(context.Context, uuid.UUID, JobStatus) (Job, error)
	GetJobByIdemKey(context.Context, string) (JobMessage, error)
}

type Service struct {
	repo  Repository
	queue Publisher
}

func NewService(repo Repository, queue Publisher) *Service {
	return &Service{repo: repo, queue: queue}
}

func (s *Service) CreateJob(ctx context.Context, req CreateJobRequest) (JobMessage, error) {
	req.Type = strings.TrimSpace(req.Type)

	if req.Type == "" {
		return JobMessage{}, ErrInvalidJob
	}

	if len(req.Payload) == 0 {
		req.Payload = json.RawMessage(`{}`)
	}

	if !json.Valid(req.Payload) {
		return JobMessage{}, ErrInvalidJob
	}

	if req.IdempotencyKey != "" {
		job, err := s.repo.GetJobByIdemKey(ctx, req.IdempotencyKey)

		switch {
		case err == nil:
			return job, nil

		case errors.Is(err, ErrJobNotFound):

		default:
			return JobMessage{}, fmt.Errorf("checking idempotency key: %w", err)
		}
	}

	job, err := s.repo.CreateJob(ctx, req)
	if err != nil {
		return JobMessage{}, err
	}

	if err = s.queue.Publish(ctx, job); err != nil {
		return JobMessage{}, err
	}

	return job, nil
}

func (s *Service) ListJobs(ctx context.Context, filter JobFilter) ([]Job, error) {
	filter.Type = strings.TrimSpace(filter.Type)
	filter.Status = JobStatus(strings.TrimSpace(string(filter.Status)))

	switch filter.Status {
	case "",
		JobStatusQueued,
		JobStatusRunning,
		JobStatusSucceeded,
		JobStatusFailed:
	default:
		return nil, ErrInvalidJobFilter
	}

	return s.repo.ListJobs(ctx, filter)
}

func (s *Service) GetJob(ctx context.Context, id string) (Job, error) {
	id = strings.TrimSpace(id)

	jobID, err := uuid.Parse(id)
	if err != nil {
		return Job{}, ErrInvalidJobID
	}

	return s.repo.GetJob(ctx, jobID)
}

func (s *Service) UpdateJobStatus(ctx context.Context, id string, status JobStatus) (Job, error) {
	id = strings.TrimSpace(id)

	jobID, err := uuid.Parse(id)
	if err != nil {
		return Job{}, ErrInvalidJobID
	}

	status = JobStatus(strings.TrimSpace(string(status)))

	switch status {
	case JobStatusQueued,
		JobStatusRunning,
		JobStatusSucceeded,
		JobStatusFailed:
	default:
		return Job{}, ErrInvalidJobStatus
	}

	return s.repo.UpdateJob(ctx, jobID, status)
}
