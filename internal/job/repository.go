package job

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, req CreateJobRequest) (Job, error) {
	query := `
		INSERT INTO jobs (type, payload)
		VALUES ($1, $2)
		RETURNING id, type, payload, status, created_at
	`

	var job Job

	if err := r.db.QueryRow(ctx, query, req.Type, []byte(req.Payload)).Scan(&job.ID, &job.Type, &job.Payload, &job.Status, &job.CreatedAt); err != nil {
		return Job{}, fmt.Errorf(
			"insert job: %w", err,
		)
	}

	return job, nil

}
