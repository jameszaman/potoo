-- +goose Up
CREATE TABLE projects (
    id              TEXT        NOT NULL PRIMARY KEY,
    organization_id TEXT        NOT NULL REFERENCES organizations(id),
    name            TEXT        NOT NULL,
    slug            TEXT        NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, slug)
);

-- +goose Down
DROP TABLE projects;
