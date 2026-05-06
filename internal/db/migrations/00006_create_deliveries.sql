-- +goose Up
CREATE TABLE deliveries (
    id                      TEXT                NOT NULL PRIMARY KEY,
    notification_id         TEXT                NOT NULL REFERENCES notifications(id),
    organization_id         TEXT                NOT NULL REFERENCES organizations(id),
    project_id              TEXT                NOT NULL REFERENCES projects(id),
    environment_id          TEXT                NOT NULL REFERENCES environments(id),
    channel                 notification_channel NOT NULL,
    provider_type           TEXT,
    provider_message_id     TEXT,
    status                  notification_status  NOT NULL DEFAULT 'queued',
    attempt_count           INT                 NOT NULL DEFAULT 0,
    last_error_code         TEXT,
    last_error_message      TEXT,
    scheduled_at            TIMESTAMPTZ,
    sent_at                 TIMESTAMPTZ,
    delivered_at            TIMESTAMPTZ,
    failed_at               TIMESTAMPTZ,
    created_at              TIMESTAMPTZ         NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ         NOT NULL DEFAULT NOW()
);

CREATE INDEX deliveries_notification_id_idx ON deliveries (notification_id);
CREATE INDEX deliveries_organization_id_idx ON deliveries (organization_id);
CREATE INDEX deliveries_status_idx          ON deliveries (status);

-- +goose Down
DROP INDEX deliveries_status_idx;
DROP INDEX deliveries_organization_id_idx;
DROP INDEX deliveries_notification_id_idx;
DROP TABLE deliveries;
