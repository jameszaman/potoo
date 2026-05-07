-- +goose Up
ALTER TABLE notifications ADD COLUMN scheduled_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE notifications DROP COLUMN scheduled_at;
