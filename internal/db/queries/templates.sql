-- name: CreateTemplate :one
INSERT INTO templates (id, organization_id, project_id, key, name, channel)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetTemplateByKey :one
SELECT * FROM templates
WHERE organization_id = $1 AND project_id = $2 AND key = $3
LIMIT 1;

-- name: ListTemplates :many
SELECT * FROM templates
WHERE organization_id = $1 AND project_id = $2
ORDER BY created_at DESC;

-- name: SetActiveVersion :one
UPDATE templates
SET active_version = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: CreateTemplateVersion :one
INSERT INTO template_versions (
    id, template_id, version_number,
    subject, html_body, text_body, sms_body, variables_schema, editor_blocks
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetTemplateVersion :one
SELECT * FROM template_versions
WHERE template_id = $1 AND version_number = $2
LIMIT 1;

-- name: GetActiveTemplateVersion :one
SELECT tv.* FROM template_versions tv
JOIN templates t ON t.id = tv.template_id
WHERE t.organization_id = $1
  AND t.project_id      = $2
  AND t.key             = $3
  AND tv.version_number = t.active_version
LIMIT 1;

-- name: NextVersionNumber :one
SELECT COALESCE(MAX(version_number), 0) + 1 AS next
FROM template_versions
WHERE template_id = $1;

-- name: ArchiveTemplateVersions :exec
UPDATE template_versions
SET status = 'archived'
WHERE template_id = $1 AND status = 'active';

-- name: ActivateTemplateVersion :one
UPDATE template_versions
SET status = 'active'
WHERE template_id = $1 AND version_number = $2
RETURNING *;
