package job

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db     *pgxpool.Pool
	logger *slog.Logger
}

func NewPostgresRepository(db *pgxpool.Pool, logger *slog.Logger) *PostgresRepository {
	return &PostgresRepository{
		db:     db,
		logger: logger,
	}
}

func (r *PostgresRepository) CreateJob(ctx context.Context, req CreateJobRequest) (JobMessage, error) {
	const query = `
		INSERT INTO jobs (type, payload, idempotency_key)
		VALUES ($1, $2, $3)
		RETURNING id, type
	`

	var jobMessage JobMessage

	err := r.db.QueryRow(ctx, query, req.Type, []byte(req.Payload), req.IdempotencyKey).Scan(
		&jobMessage.JobId, &jobMessage.Type,
	)

	if err != nil {
		return JobMessage{}, fmt.Errorf("insert job: %w", err)
	}

	return jobMessage, nil
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

func (r *PostgresRepository) GetJob(ctx context.Context, id uuid.UUID) (Job, error) {
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

func (r *PostgresRepository) UpdateJob(ctx context.Context, id uuid.UUID, status JobStatus) (Job, error) {
	const query = `
		UPDATE jobs
		SET status = $2
		WHERE id = $1
		RETURNING id, type, payload, status, created_at
	`

	var job Job

	err := r.db.QueryRow(ctx, query, id, status).Scan(
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
		return Job{}, fmt.Errorf("update job: %w", err)
	}

	return job, nil
}

func (r *PostgresRepository) GetJobByIdemKey(ctx context.Context, key string) (JobMessage, error) {
	query := `SELECT  id, type
	FROM jobs 
	WHERE idempotency_key = $1`

	var jobMessage JobMessage

	err := r.db.QueryRow(ctx, query, key).Scan(
		&jobMessage.JobId, &jobMessage.Type,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return JobMessage{}, ErrJobNotFound
	}

	if err != nil {
		return JobMessage{}, fmt.Errorf("get job by idem key: %w", err)
	}

	return jobMessage, nil
}
