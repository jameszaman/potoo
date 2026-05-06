-- +goose Up
-- Truncate any existing sessions — they predate org scoping and cannot be migrated.
DELETE FROM sessions;

ALTER TABLE sessions
    ADD COLUMN org_id TEXT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE;

CREATE INDEX sessions_org_id_idx ON sessions (org_id);

-- +goose Down
ALTER TABLE sessions DROP COLUMN org_id;
