CREATE TABLE IF NOT EXISTS entries (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    text TEXT,
    mood INTEGER,
    category TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);