-- +goose Up
CREATE TYPE delivery_event_type AS ENUM (
    'queued', 'sent', 'delivered', 'opened', 'clicked',
    'bounced', 'complained', 'failed'
);

CREATE TABLE delivery_events (
    id                TEXT               NOT NULL PRIMARY KEY,
    delivery_id       TEXT               NOT NULL REFERENCES deliveries(id),
    event_type        delivery_event_type NOT NULL,
    provider_type     TEXT,
    provider_event_id TEXT,
    payload           JSONB              NOT NULL DEFAULT '{}',
    occurred_at       TIMESTAMPTZ        NOT NULL,
    created_at        TIMESTAMPTZ        NOT NULL DEFAULT NOW(),
    UNIQUE (provider_type, provider_event_id)
);

CREATE INDEX delivery_events_delivery_id_idx ON delivery_events (delivery_id);

-- +goose Down
DROP INDEX delivery_events_delivery_id_idx;
DROP TABLE delivery_events;
DROP TYPE delivery_event_type;
