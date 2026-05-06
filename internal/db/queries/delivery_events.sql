-- name: CreateDeliveryEvent :one
INSERT INTO delivery_events (
    id, delivery_id, event_type, provider_type,
    provider_event_id, payload, occurred_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
ON CONFLICT (provider_type, provider_event_id) DO NOTHING
RETURNING *;

-- name: GetDeliveryEvents :many
SELECT * FROM delivery_events
WHERE delivery_id = $1
ORDER BY occurred_at ASC;
