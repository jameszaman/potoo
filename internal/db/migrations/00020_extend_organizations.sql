-- +goose Up
ALTER TABLE organizations
    ADD COLUMN website     TEXT,
    ADD COLUMN description TEXT CHECK (char_length(description) <= 200);

-- +goose Down
ALTER TABLE organizations
    DROP COLUMN website,
    DROP COLUMN description;
