CREATE TABLE IF NOT EXISTS failed_jobs (
    id         SERIAL PRIMARY KEY,
    job_type   TEXT NOT NULL,
    payload    JSONB NOT NULL,
    error      TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    retried_at TIMESTAMP
);