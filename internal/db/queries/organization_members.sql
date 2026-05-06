-- name: CreateOrgMember :one
INSERT INTO organization_members (id, org_id, user_id, role)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOrgMember :one
SELECT * FROM organization_members
WHERE org_id = $1 AND user_id = $2;

-- name: ListUserOrgs :many
SELECT o.* FROM organizations o
JOIN organization_members m ON m.org_id = o.id
WHERE m.user_id = $1
ORDER BY o.name;

-- name: ListOrgMembers :many
SELECT * FROM organization_members
WHERE org_id = $1
ORDER BY created_at;

-- name: DeleteOrgMember :exec
DELETE FROM organization_members WHERE org_id = $1 AND user_id = $2;
