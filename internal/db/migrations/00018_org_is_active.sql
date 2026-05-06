-- +goose Up
ALTER TABLE organizations
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE organizations DROP COLUMN is_active;
