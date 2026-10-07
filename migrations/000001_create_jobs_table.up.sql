CREATE TABLE jobs (
    id BIGSERIAL PRIMARY KEY,

    type TEXT NOT NULL,

    payload JSONB NOT NULL DEFAULT '{}'::jsonb,

    status TEXT NOT NULL DEFAULT 'queued',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT valid_job_status CHECK (
        status IN (
            'queued',
            'running',
            'succeeded',
            'failed'
        )
    )
);