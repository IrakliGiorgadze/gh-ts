-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto"; -- for gen_random_uuid()

CREATE TABLE IF NOT EXISTS tickets (
                                       id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title       TEXT        NOT NULL,
    description TEXT        NOT NULL DEFAULT '',
    category    TEXT        NOT NULL DEFAULT 'Software',
    priority    TEXT        NOT NULL DEFAULT 'Low',      -- Low/Medium/High/Critical
    status      TEXT        NOT NULL DEFAULT 'New',      -- New/Open/In Progress/Pending/Resolved/Closed
    assignee    TEXT        NULL,
    department  TEXT        NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
    );

CREATE INDEX IF NOT EXISTS idx_tickets_status     ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_priority   ON tickets(priority);
CREATE INDEX IF NOT EXISTS idx_tickets_category   ON tickets(category);
CREATE INDEX IF NOT EXISTS idx_tickets_updated_at ON tickets(updated_at DESC);

-- +goose Down
DROP TABLE IF EXISTS tickets;