package job

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
)

var ErrInvalidJob = errors.New("invalid job")

type Service struct {
	repo *Repository
}

func NewService(reo *Repository) *Service {
	return &Service{repo: reo}
}

func (s *Service) Create(ctx context.Context, req CreateJobRequest) (Job, error) {
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
	return s.repo.Create(ctx, req)
}
