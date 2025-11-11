-- +goose Up
-- Introduces human-friendly ticket aliases in the format TKT-YYYY-#####.
CREATE SEQUENCE IF NOT EXISTS ticket_alias_seq;

ALTER TABLE tickets
ADD COLUMN IF NOT EXISTS alias TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS tickets_alias_key
ON tickets ((lower(alias)))
WHERE alias IS NOT NULL;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION ticket_alias_default()
RETURNS trigger AS $$
BEGIN
  IF NEW.alias IS NULL OR NEW.alias = '' THEN
    NEW.alias := 'TKT-' || to_char(COALESCE(NEW.created_at, now()), 'YYYY')
      || '-' || lpad(nextval('ticket_alias_seq')::text, 5, '0');
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS set_ticket_alias ON tickets;
CREATE TRIGGER set_ticket_alias
BEFORE INSERT ON tickets
FOR EACH ROW
EXECUTE FUNCTION ticket_alias_default();

-- +goose StatementBegin
UPDATE tickets
SET alias = 'TKT-' || to_char(COALESCE(created_at, now()), 'YYYY')
  || '-' || lpad(nextval('ticket_alias_seq')::text, 5, '0')
WHERE alias IS NULL OR alias = '';
-- +goose StatementEnd

-- +goose Down
DROP TRIGGER IF EXISTS set_ticket_alias ON tickets;
-- +goose StatementBegin
DROP FUNCTION IF EXISTS ticket_alias_default();
-- +goose StatementEnd

DROP INDEX IF EXISTS tickets_alias_key;

ALTER TABLE tickets
DROP COLUMN IF EXISTS alias;

DROP SEQUENCE IF EXISTS ticket_alias_seq;

