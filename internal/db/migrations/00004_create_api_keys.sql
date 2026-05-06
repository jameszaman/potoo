-- +goose Up
CREATE TABLE api_keys (
    id              TEXT        NOT NULL PRIMARY KEY,
    organization_id TEXT        NOT NULL REFERENCES organizations(id),
    project_id      TEXT        NOT NULL REFERENCES projects(id),
    environment_id  TEXT        NOT NULL REFERENCES environments(id),
    name            TEXT        NOT NULL,
    key_prefix      TEXT        NOT NULL,
    key_hash        TEXT        NOT NULL UNIQUE,
    scopes          TEXT[]      NOT NULL DEFAULT '{}',
    last_used_at    TIMESTAMPTZ,
    revoked_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX api_keys_key_hash_idx ON api_keys (key_hash);

-- +goose Down
DROP INDEX api_keys_key_hash_idx;
DROP TABLE api_keys;
