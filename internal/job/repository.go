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

func (r *Repository) CreateJob(ctx context.Context, req CreateJobRequest) (Job, error) {
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

func (r *Repository) ListJobs(ctx context.Context, filter JobFilter) ([]Job, error) {
	query := `SELECT id, type, payload, status, created_at
	FROM jobs
	WHERE ($1 = '' OR type =$1)
	AND ($2 = '' OR status =$2)
	ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, filter.Type, filter.Status)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()

	var jobs []Job

	for rows.Next() {
		var job Job
		if err := rows.Scan(
			&job.ID,
			&job.Type,
			&job.Payload,
			&job.Status,
			&job.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("Scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate jobs: %w", err)
	}

	return jobs, nil
}
