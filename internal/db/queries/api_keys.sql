-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys
WHERE key_hash = $1
  AND revoked_at IS NULL
LIMIT 1;

-- name: TouchAPIKey :exec
UPDATE api_keys
SET last_used_at = NOW()
WHERE id = $1;

-- name: CreateAPIKey :one
INSERT INTO api_keys (id, organization_id, name, key_prefix, key_hash, scopes)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListAPIKeysByOrg :many
SELECT id, organization_id, name, key_prefix, scopes, last_used_at, revoked_at, created_at
FROM api_keys
WHERE organization_id = $1
ORDER BY created_at DESC;

-- name: RevokeAPIKey :exec
UPDATE api_keys
SET revoked_at = NOW()
WHERE id = $1 AND organization_id = $2 AND revoked_at IS NULL;
