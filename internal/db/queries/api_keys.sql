-- name: GetAPIKeyByHash :one
SELECT * FROM api_keys
WHERE key_hash = $1
  AND revoked_at IS NULL
LIMIT 1;

-- name: TouchAPIKey :exec
UPDATE api_keys
SET last_used_at = NOW()
WHERE id = $1;
