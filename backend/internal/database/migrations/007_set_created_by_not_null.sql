-- +goose Up
ALTER TABLE tickets
    ALTER COLUMN created_by SET NOT NULL;

-- +goose Down
ALTER TABLE tickets
    ALTER COLUMN created_by DROP NOT NULL;