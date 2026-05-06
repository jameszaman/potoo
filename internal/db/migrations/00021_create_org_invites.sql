-- +goose Up
CREATE TABLE org_invites (
    id         TEXT        NOT NULL PRIMARY KEY,
    org_id     TEXT        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    created_by TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT        NOT NULL UNIQUE,
    used_at    TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX org_invites_token_idx  ON org_invites (token);
CREATE INDEX org_invites_org_id_idx ON org_invites (org_id);

-- +goose Down
DROP TABLE org_invites;
