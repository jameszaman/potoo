-- +goose Up
CREATE TYPE org_type AS ENUM ('platform', 'customer');

ALTER TABLE organizations
    ADD COLUMN type org_type NOT NULL DEFAULT 'customer';

-- +goose Down
ALTER TABLE organizations DROP COLUMN type;
DROP TYPE org_type;
