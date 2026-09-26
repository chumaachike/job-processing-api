package job

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(
	ctx context.Context,
	jobType string,
	payload json.RawMessage,
) (Job, error) {
	query := `
		INSERT INTO jobs (
			id,
			type,
			payload
		)
		VALUES (
			gen_random_uuid(),
			$1,
			$2::jsonb
		)
		RETURNING id::text, status
	`

	var job Job

	job.Type = jobType
	job.Payload = payload

	err := r.db.QueryRow(
		ctx,
		query,
		jobType,
		string(payload),
	).Scan(
		&job.ID,
		&job.Status,
	)

	if err != nil {
		return Job{}, fmt.Errorf("create job: %w", err)
	}

	return job, nil
}
