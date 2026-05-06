-- name: CreateDelivery :one
INSERT INTO deliveries (
    id, notification_id, organization_id, project_id, environment_id,
    channel, status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetDelivery :one
SELECT * FROM deliveries
WHERE id = $1 AND organization_id = $2
LIMIT 1;

-- name: GetDeliveriesByNotification :many
SELECT * FROM deliveries
WHERE notification_id = $1 AND organization_id = $2
ORDER BY created_at ASC;

-- name: UpdateDeliveryStatus :one
UPDATE deliveries
SET status = $2, attempt_count = attempt_count + 1, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDeliverySent :one
UPDATE deliveries
SET
    status              = 'sent',
    provider_type       = $2,
    provider_message_id = $3,
    sent_at             = NOW(),
    updated_at          = NOW()
WHERE id = $1
RETURNING *;

-- name: GetDeliveryByProviderMessageID :one
SELECT * FROM deliveries
WHERE provider_message_id = $1
LIMIT 1;

-- name: UpdateDeliveryDelivered :one
UPDATE deliveries
SET status = 'delivered', delivered_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateDeliveryBounced :one
UPDATE deliveries
SET status = 'bounced', failed_at = NOW(), updated_at = NOW()
WHERE id = $1
RETURNING *;
