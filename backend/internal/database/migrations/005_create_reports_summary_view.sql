-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE VIEW reports_summary AS
SELECT
    COUNT(*) FILTER (WHERE status NOT IN ('Resolved', 'Closed')) AS open,
  COUNT(*) FILTER (
    WHERE status = 'Resolved'
      AND updated_at > NOW() - INTERVAL '7 days'
  ) AS resolved7d,
  COUNT(*) FILTER (
    WHERE priority IN ('High', 'Critical')
      AND status NOT IN ('Resolved', 'Closed')
  ) AS high_critical_open
FROM tickets;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP VIEW IF EXISTS reports_summary;
-- +goose StatementEnd