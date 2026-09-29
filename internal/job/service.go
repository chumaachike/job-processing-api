package job

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

var ErrInvalidJob = errors.New("invalid job")
var ErrInvalidJobFilter = errors.New("invalid job fileter")

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
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
