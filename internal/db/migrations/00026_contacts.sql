-- +goose Up
CREATE TABLE contacts (
    id            TEXT        NOT NULL PRIMARY KEY,
    org_id        TEXT        NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email         TEXT        NOT NULL,
    name          TEXT,
    phone         TEXT,
    tag           TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (org_id, email)
);

CREATE INDEX contacts_org_id_idx ON contacts (org_id);
CREATE INDEX contacts_org_tag_idx ON contacts (org_id, tag);

-- +goose Down
DROP TABLE IF EXISTS contacts;
