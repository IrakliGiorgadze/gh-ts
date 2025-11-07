-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS idx_tickets_status        ON tickets (status);
CREATE INDEX IF NOT EXISTS idx_tickets_priority      ON tickets (priority);
CREATE INDEX IF NOT EXISTS idx_tickets_category      ON tickets (category);
CREATE INDEX IF NOT EXISTS idx_tickets_assignee      ON tickets (assignee);
CREATE INDEX IF NOT EXISTS idx_tickets_updated_at    ON tickets (updated_at DESC);

-- Optional trigram (fast ILIKE search on title/description)
CREATE INDEX IF NOT EXISTS idx_tickets_title_trgm       ON tickets USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_tickets_description_trgm ON tickets USING gin (description gin_trgm_ops);

-- +goose Down
DROP INDEX IF EXISTS idx_tickets_description_trgm;
DROP INDEX IF EXISTS idx_tickets_title_trgm;
DROP INDEX IF EXISTS idx_tickets_updated_at;
DROP INDEX IF EXISTS idx_tickets_assignee;
DROP INDEX IF EXISTS idx_tickets_category;
DROP INDEX IF EXISTS idx_tickets_priority;
DROP INDEX IF EXISTS idx_tickets_status;
-- Note: we usually don't drop the extension in Down (it may be used elsewhere).
-- If you really want symmetry, uncomment the next two lines:
-- -- +goose StatementBegin
-- DROP EXTENSION IF EXISTS pg_trgm;
-- -- +goose StatementEnd