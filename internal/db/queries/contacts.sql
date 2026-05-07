-- name: CreateContact :one
INSERT INTO contacts (id, org_id, email, name, phone, tag)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetContact :one
SELECT * FROM contacts
WHERE id = $1 AND org_id = $2;

-- name: ListContacts :many
SELECT * FROM contacts
WHERE org_id = $1
ORDER BY created_at DESC;

-- name: UpdateContact :one
UPDATE contacts
SET name = $2, phone = $3, tag = $4, updated_at = NOW()
WHERE id = $1 AND org_id = $5
RETURNING *;

-- name: DeleteContact :exec
DELETE FROM contacts
WHERE id = $1 AND org_id = $2;

-- name: ListContactsByTag :many
SELECT * FROM contacts
WHERE org_id = $1 AND tag = $2
ORDER BY created_at DESC;

-- name: ListContactTags :many
SELECT DISTINCT tag FROM contacts
WHERE org_id = $1 AND tag IS NOT NULL
ORDER BY tag;
