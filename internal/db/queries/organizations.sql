-- name: CreateOrganization :one
INSERT INTO organizations (id, name, slug, type)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE id = $1;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations WHERE slug = $1;

-- name: GetPlatformOrg :one
SELECT * FROM organizations WHERE type = 'platform' LIMIT 1;

-- name: PlatformOrgExists :one
SELECT EXISTS(SELECT 1 FROM organizations WHERE type = 'platform') AS exists;

-- name: ListOrganizations :many
SELECT * FROM organizations ORDER BY name;
