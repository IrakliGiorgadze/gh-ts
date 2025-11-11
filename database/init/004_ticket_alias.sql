-- Add human-friendly alias for tickets in format TKT-YYYY-#####.
-- Safe to run multiple times; uses IF NOT EXISTS guards.

CREATE SEQUENCE IF NOT EXISTS ticket_alias_seq;

ALTER TABLE tickets
ADD COLUMN IF NOT EXISTS alias TEXT;

-- Ensure uniqueness.
CREATE UNIQUE INDEX IF NOT EXISTS tickets_alias_key
ON tickets ((lower(alias)))
WHERE alias IS NOT NULL;

-- Trigger to automatically populate alias on insert.
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

DROP TRIGGER IF EXISTS set_ticket_alias ON tickets;
CREATE TRIGGER set_ticket_alias
BEFORE INSERT ON tickets
FOR EACH ROW
EXECUTE FUNCTION ticket_alias_default();

-- Backfill existing rows that lack an alias.
UPDATE tickets
SET alias = 'TKT-' || to_char(COALESCE(created_at, now()), 'YYYY')
  || '-' || lpad(nextval('ticket_alias_seq')::text, 5, '0')
WHERE alias IS NULL OR alias = '';

