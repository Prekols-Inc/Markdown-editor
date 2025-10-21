\connect postgres;

ALTER USER postgres WITH PASSWORD 'password';

\connect auth_db;

CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- uuid extension

DROP TABLE IF EXISTS users CASCADE;

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

INSERT INTO users (username, password_hash)
VALUES ('admin', '$2a$10$KsMkClK0bvgBZVSGTh76E.iEwg9VWFEpFTbPuwKCZZG3822DHiiSa') -- bcrypt hash for 'password'
ON CONFLICT (username) DO NOTHING;
