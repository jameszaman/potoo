-- name: CreateProviderConnection :one
INSERT INTO provider_connections (
    id, organization_id, project_id, environment_id,
    provider_type, channel, display_name, encrypted_config, is_default
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetProviderConnection :one
SELECT * FROM provider_connections
WHERE id = $1 AND organization_id = $2
LIMIT 1;

-- name: ListProviderConnections :many
SELECT * FROM provider_connections
WHERE organization_id = $1 AND project_id = $2 AND environment_id = $3
ORDER BY created_at DESC;

-- name: DeleteProviderConnection :exec
DELETE FROM provider_connections
WHERE id = $1 AND organization_id = $2;

-- name: GetDefaultProvider :one
SELECT * FROM provider_connections
WHERE organization_id = $1
  AND project_id      = $2
  AND environment_id  = $3
  AND channel         = $4
  AND is_default      = true
  AND is_active       = true
LIMIT 1;

-- name: GetProviderConnectionByType :one
SELECT * FROM provider_connections
WHERE provider_type = $1
  AND is_active     = true
LIMIT 1;

-- name: UpdateProviderConnection :one
UPDATE provider_connections
SET display_name     = $3,
    encrypted_config = $4,
    updated_at       = NOW()
WHERE id = $1 AND organization_id = $2
RETURNING *;

-- name: ClearDefaultProviders :exec
UPDATE provider_connections
SET is_default = false
WHERE organization_id = $1
  AND project_id      = $2
  AND environment_id  = $3
  AND channel         = $4;

-- name: SetProviderDefault :one
UPDATE provider_connections
SET is_default = true
WHERE id = $1 AND organization_id = $2
RETURNING *;
