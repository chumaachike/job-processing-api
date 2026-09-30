package job

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

func (r *PostgresRepository) CreateJob(
	ctx context.Context,
	req CreateJobRequest,
) (Job, error) {
	const query = `
		INSERT INTO jobs (type, payload)
		VALUES ($1, $2)
		RETURNING id, type, payload, status, created_at
	`

	var job Job

	err := r.db.QueryRow(
		ctx,
		query,
		req.Type,
		[]byte(req.Payload),
	).Scan(
		&job.ID,
		&job.Type,
		&job.Payload,
		&job.Status,
		&job.CreatedAt,
	)

	if err != nil {
		return Job{}, fmt.Errorf("insert job: %w", err)
	}

	return job, nil
}

func (r *PostgresRepository) ListJobs(ctx context.Context, filter JobFilter) ([]Job, error) {
	const query = `
		SELECT id, type, payload, status, created_at
		FROM jobs
		WHERE ($1 = '' OR type = $1)
		  AND ($2 = '' OR status = $2)
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(
		ctx,
		query,
		filter.Type,
		filter.Status,
	)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]Job, 0)

	for rows.Next() {
		var job Job

		if err := rows.Scan(
			&job.ID,
			&job.Type,
			&job.Payload,
			&job.Status,
			&job.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan job: %w", err)
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate jobs: %w", err)
	}

	return jobs, nil
}

func (r *PostgresRepository) GetJob(ctx context.Context, id int64) (Job, error) {
	const query = `
		SELECT id, type, payload, status, created_at
		FROM jobs
		WHERE id = $1
	`

	var job Job

	err := r.db.QueryRow(ctx, query, id).Scan(
		&job.ID,
		&job.Type,
		&job.Payload,
		&job.Status,
		&job.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return Job{}, ErrJobNotFound
	}

	if err != nil {
		return Job{}, fmt.Errorf("get job: %w", err)
	}

	return job, nil
}
