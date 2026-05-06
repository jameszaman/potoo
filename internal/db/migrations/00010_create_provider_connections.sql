-- +goose Up
CREATE TYPE provider_type AS ENUM ('resend', 'sendgrid', 'ses', 'twilio', 'fcm');
CREATE TYPE provider_channel AS ENUM ('email', 'sms', 'push');

CREATE TABLE provider_connections (
    id              TEXT             NOT NULL PRIMARY KEY,
    organization_id TEXT             NOT NULL REFERENCES organizations(id),
    project_id      TEXT             NOT NULL REFERENCES projects(id),
    environment_id  TEXT             NOT NULL REFERENCES environments(id),
    provider_type   provider_type    NOT NULL,
    channel         provider_channel NOT NULL,
    display_name    TEXT             NOT NULL,
    encrypted_config TEXT            NOT NULL,
    is_default      BOOLEAN          NOT NULL DEFAULT false,
    is_active       BOOLEAN          NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE INDEX provider_connections_org_project_env_idx
    ON provider_connections (organization_id, project_id, environment_id);

-- +goose Down
DROP INDEX provider_connections_org_project_env_idx;
DROP TABLE provider_connections;
DROP TYPE provider_channel;
DROP TYPE provider_type;
