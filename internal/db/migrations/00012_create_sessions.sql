-- +goose Up
CREATE TABLE sessions (
    id                  TEXT        NOT NULL PRIMARY KEY,
    user_id             TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    refresh_token_hash  TEXT        NOT NULL UNIQUE,
    expires_at          TIMESTAMPTZ NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX sessions_user_id_idx ON sessions (user_id);

-- +goose Down
DROP TABLE sessions;
