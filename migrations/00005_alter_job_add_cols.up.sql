ALTER TABLE jobs
    ADD COLUMN idempotency_key TEXT,
    ADD COLUMN result JSONB,
    ADD COLUMN error_message TEXT,
    ADD COLUMN started_at TIMESTAMPTZ,
    ADD COLUMN completed_at TIMESTAMPTZ;