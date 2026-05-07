-- name: CreateOrgInvite :one
INSERT INTO org_invites (id, org_id, created_by, token, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetOrgInviteByToken :one
SELECT * FROM org_invites WHERE token = $1;

-- name: GetOrgInviteByID :one
SELECT * FROM org_invites WHERE id = $1;

-- name: MarkOrgInviteUsed :one
UPDATE org_invites SET used_at = NOW() WHERE id = $1 RETURNING *;

-- name: DeleteOrgInvite :one
UPDATE org_invites SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL RETURNING *;

-- name: ListOrgInvites :many
SELECT * FROM org_invites WHERE org_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC;
