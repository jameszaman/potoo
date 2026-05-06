-- name: CreateSession :one
INSERT INTO sessions (id, user_id, org_id, refresh_token_hash, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE id = $1;

-- name: GetSessionByRefreshTokenHash :one
SELECT * FROM sessions
WHERE refresh_token_hash = $1
LIMIT 1;

-- name: RotateSession :one
UPDATE sessions
SET refresh_token_hash = $2,
    expires_at         = $3,
    updated_at         = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSession :exec
DELETE FROM sessions WHERE id = $1;

-- name: DeleteSessionByRefreshTokenHash :exec
DELETE FROM sessions WHERE refresh_token_hash = $1;

-- name: DeleteAllUserSessions :exec
DELETE FROM sessions WHERE user_id = $1;

-- name: DeleteAllUserOrgSessions :exec
DELETE FROM sessions WHERE user_id = $1 AND org_id = $2;
