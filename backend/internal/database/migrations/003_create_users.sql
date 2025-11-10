-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TABLE IF NOT EXISTS users (
                                     id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email          CITEXT UNIQUE NOT NULL,
    name           TEXT NOT NULL,
    role           TEXT NOT NULL DEFAULT 'end_user',  -- end_user / agent / supervisor / admin
    password_h  TEXT NOT NULL,
    active         BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
    );

-- Optional: enforce valid roles
ALTER TABLE users
    ADD CONSTRAINT users_role_check
        CHECK (role IN ('end_user', 'agent', 'supervisor', 'admin'));

CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- +goose Down
DROP TABLE IF EXISTS users;