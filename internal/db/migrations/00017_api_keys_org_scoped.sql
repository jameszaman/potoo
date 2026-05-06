-- +goose Up
ALTER TABLE api_keys
    DROP COLUMN project_id,
    DROP COLUMN environment_id;

-- +goose Down
ALTER TABLE api_keys
    ADD COLUMN project_id    TEXT REFERENCES projects(id),
    ADD COLUMN environment_id TEXT REFERENCES environments(id);
