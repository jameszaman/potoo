-- +goose Up
CREATE TYPE environment_name AS ENUM ('development', 'staging', 'production');

CREATE TABLE environments (
    id          TEXT             NOT NULL PRIMARY KEY,
    project_id  TEXT             NOT NULL REFERENCES projects(id),
    name        environment_name NOT NULL,
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    UNIQUE (project_id, name)
);

-- +goose Down
DROP TABLE environments;
DROP TYPE environment_name;
