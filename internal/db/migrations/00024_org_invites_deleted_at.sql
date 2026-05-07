-- +goose Up
ALTER TABLE org_invites ADD COLUMN deleted_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE org_invites DROP COLUMN deleted_at;
