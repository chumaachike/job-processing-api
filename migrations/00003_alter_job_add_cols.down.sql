ALTER TABLE jobs
    DROP COLUMN idempotency_key,
    DROP COLUMN result,
    DROP COLUMN error_message,
    DROP COLUMN started_at,
    DROP COLUMN completed_at;