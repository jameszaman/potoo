-- +goose Up
CREATE TABLE templates (
    id              TEXT        NOT NULL PRIMARY KEY,
    organization_id TEXT        NOT NULL REFERENCES organizations(id),
    project_id      TEXT        NOT NULL REFERENCES projects(id),
    key             TEXT        NOT NULL,
    name            TEXT        NOT NULL,
    channel         notification_channel NOT NULL,
    active_version  INT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (organization_id, project_id, key)
);

CREATE TYPE template_version_status AS ENUM ('draft', 'active', 'archived');

CREATE TABLE template_versions (
    id               TEXT                   NOT NULL PRIMARY KEY,
    template_id      TEXT                   NOT NULL REFERENCES templates(id),
    version_number   INT                    NOT NULL,
    subject          TEXT,
    html_body        TEXT,
    text_body        TEXT,
    sms_body         TEXT,
    variables_schema JSONB                  NOT NULL DEFAULT '{}',
    status           template_version_status NOT NULL DEFAULT 'draft',
    created_at       TIMESTAMPTZ            NOT NULL DEFAULT NOW(),
    UNIQUE (template_id, version_number)
);

CREATE INDEX template_versions_template_id_idx ON template_versions (template_id);

-- +goose Down
DROP INDEX template_versions_template_id_idx;
DROP TABLE template_versions;
DROP TYPE template_version_status;
DROP TABLE templates;
