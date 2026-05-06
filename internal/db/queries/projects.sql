-- name: CreateProject :one
INSERT INTO projects (id, organization_id, name, slug, is_default)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetDefaultProject :one
SELECT * FROM projects
WHERE organization_id = $1 AND is_default = true
LIMIT 1;
