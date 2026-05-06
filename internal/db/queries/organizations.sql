-- name: CreateOrganization :one
INSERT INTO organizations (id, name, slug, type, website, description)
VALUES ($1, $2, $3, $4, $5, $6)
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

-- name: SetOrgActive :one
UPDATE organizations
SET is_active  = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: GetOrgStats :one
SELECT
    (SELECT COUNT(*) FROM api_keys      WHERE api_keys.organization_id = $1 AND revoked_at IS NULL) AS api_key_count,
    (SELECT COUNT(*) FROM notifications WHERE notifications.organization_id = $1)                    AS call_count;
