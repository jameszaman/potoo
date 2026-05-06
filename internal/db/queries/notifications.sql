-- name: CreateNotification :one
INSERT INTO notifications (
    id, organization_id, project_id, environment_id,
    external_id, template_key, channel, recipient_ref,
    status, metadata
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetNotification :one
SELECT * FROM notifications
WHERE id = $1 AND organization_id = $2
LIMIT 1;

-- name: UpdateNotificationStatus :one
UPDATE notifications
SET status = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;
