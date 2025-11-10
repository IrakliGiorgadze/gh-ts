-- +goose Up
-- Install pg_trgm for trigram indexes (wrap for goose)
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;
-- +goose StatementEnd

-- Keep ONLY the indexes that are not already created in 001:
-- Assignee (new) + GIN trigram indexes for fast ILIKE search
CREATE INDEX IF NOT EXISTS idx_tickets_assignee        ON tickets (assignee);
CREATE INDEX IF NOT EXISTS idx_tickets_title_trgm      ON tickets USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_tickets_description_trgm ON tickets USING gin (description gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_tickets_description_trgm;
DROP INDEX IF EXISTS idx_tickets_title_trgm;
DROP INDEX IF EXISTS idx_tickets_assignee;

-- (Normally we don't drop extensions in Down because others may rely on them.)
-- If you want symmetry, you can uncomment:
-- -- +goose StatementBegin
-- DROP EXTENSION IF EXISTS pg_trgm;
-- -- +goose StatementEnd