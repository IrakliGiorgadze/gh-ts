-- +goose Up
-- Add column if missing
ALTER TABLE tickets
    ADD COLUMN IF NOT EXISTS created_by UUID;

-- Create the FK only if it doesn't already exist
-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint
    WHERE conname = 'fk_tickets_created_by'
  ) THEN
ALTER TABLE tickets
    ADD CONSTRAINT fk_tickets_created_by
        FOREIGN KEY (created_by) REFERENCES users(id)
            ON DELETE SET NULL;
END IF;
END $$;
-- +goose StatementEnd

-- Optional: index for querying by creator
CREATE INDEX IF NOT EXISTS idx_tickets_created_by ON tickets(created_by);

-- +goose Down
ALTER TABLE tickets DROP CONSTRAINT IF EXISTS fk_tickets_created_by;
DROP INDEX IF EXISTS idx_tickets_created_by;
ALTER TABLE tickets DROP COLUMN IF EXISTS created_by;