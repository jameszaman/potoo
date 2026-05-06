-- +goose Up
ALTER TABLE users
    ADD COLUMN first_name TEXT NOT NULL,
    ADD COLUMN last_name  TEXT NOT NULL,
    ADD COLUMN phone      TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE users
    DROP COLUMN first_name,
    DROP COLUMN last_name,
    DROP COLUMN phone;
