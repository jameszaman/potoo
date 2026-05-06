-- +goose Up
ALTER TYPE provider_type ADD VALUE 'smtp';

-- +goose Down
-- Postgres does not support removing enum values; this migration is irreversible.
-- To roll back: recreate the enum without 'smtp' and migrate all data manually.
