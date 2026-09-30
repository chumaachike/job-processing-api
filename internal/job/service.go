package job

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
)

type Repository interface {
	CreateJob(ctx context.Context, req CreateJobRequest) (Job, error)
	ListJobs(ctx context.Context, filter JobFilter) ([]Job, error)
	GetJob(ctx context.Context, id int64) (Job, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateJob(ctx context.Context, req CreateJobRequest) (Job, error) {
	req.Type = strings.TrimSpace(req.Type)

	if req.Type == "" {
		return Job{}, ErrInvalidJob
	}

	if len(req.Payload) == 0 {
		req.Payload = json.RawMessage(`{}`)
	}

	if !json.Valid(req.Payload) {
		return Job{}, ErrInvalidJob
	}

	return s.repo.CreateJob(ctx, req)
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

	jobID, err := strconv.ParseInt(id, 10, 64)
	if err != nil || jobID <= 0 {
		return Job{}, ErrInvalidJobID
	}

	return s.repo.GetJob(ctx, jobID)
}
