-- +goose Up
ALTER TABLE template_versions ADD COLUMN editor_blocks TEXT;

-- +goose Down
ALTER TABLE template_versions DROP COLUMN editor_blocks;
