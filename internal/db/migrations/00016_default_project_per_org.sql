-- +goose Up
ALTER TABLE projects ADD COLUMN is_default BOOLEAN NOT NULL DEFAULT false;

CREATE UNIQUE INDEX projects_org_default_idx ON projects (organization_id)
    WHERE is_default = true;

-- Backfill: mark the earliest project per org as default.
UPDATE projects p
SET is_default = true
WHERE id = (
    SELECT id FROM projects
    WHERE organization_id = p.organization_id
    ORDER BY created_at
    LIMIT 1
);

-- +goose Down
DROP INDEX projects_org_default_idx;
ALTER TABLE projects DROP COLUMN is_default;
