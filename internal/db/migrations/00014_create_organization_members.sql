-- +goose Up
CREATE TYPE org_role AS ENUM ('owner', 'member');

CREATE TABLE organization_members (
    id          TEXT        NOT NULL PRIMARY KEY,
    org_id      TEXT        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id     TEXT        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role        org_role    NOT NULL DEFAULT 'member',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, user_id)
);

CREATE INDEX organization_members_user_id_idx ON organization_members (user_id);
CREATE INDEX organization_members_org_id_idx ON organization_members (org_id);

-- +goose Down
DROP TABLE organization_members;
DROP TYPE org_role;
