-- +goose Up
CREATE TYPE notification_status AS ENUM (
    'queued',
    'processing',
    'sent',
    'delivered',
    'failed_temporary',
    'failed_permanent',
    'bounced',
    'complained',
    'suppressed',
    'cancelled'
);

CREATE TYPE notification_channel AS ENUM ('email', 'sms', 'push', 'whatsapp');

CREATE TABLE notifications (
    id              TEXT                  NOT NULL PRIMARY KEY,
    organization_id TEXT                  NOT NULL REFERENCES organizations(id),
    project_id      TEXT                  NOT NULL REFERENCES projects(id),
    environment_id  TEXT                  NOT NULL REFERENCES environments(id),
    external_id     TEXT,
    template_key    TEXT,
    channel         notification_channel  NOT NULL,
    recipient_ref   TEXT,
    status          notification_status   NOT NULL DEFAULT 'queued',
    metadata        JSONB                 NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ           NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ           NOT NULL DEFAULT NOW()
);

CREATE INDEX notifications_organization_id_idx ON notifications (organization_id);
CREATE INDEX notifications_status_idx          ON notifications (status);

-- +goose Down
DROP INDEX notifications_status_idx;
DROP INDEX notifications_organization_id_idx;
DROP TABLE notifications;
DROP TYPE notification_channel;
DROP TYPE notification_status;
