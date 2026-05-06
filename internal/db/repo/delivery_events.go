package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/notifylayer/notifylayer/internal/db/sqlc"
)

type DeliveryEventRepo struct {
	q *db.Queries
}

func NewDeliveryEventRepo(pool *pgxpool.Pool) *DeliveryEventRepo {
	return &DeliveryEventRepo{q: db.New(pool)}
}

type CreateDeliveryEventParams struct {
	DeliveryID      string
	EventType       db.DeliveryEventType
	ProviderType    *string
	ProviderEventID *string
	Payload         map[string]any
	OccurredAt      time.Time
}

func (r *DeliveryEventRepo) Create(ctx context.Context, p CreateDeliveryEventParams) (*db.DeliveryEvent, error) {
	payload, err := json.Marshal(p.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	row, err := r.q.CreateDeliveryEvent(ctx, db.CreateDeliveryEventParams{
		ID:              newID(),
		DeliveryID:      p.DeliveryID,
		EventType:       p.EventType,
		ProviderType:    p.ProviderType,
		ProviderEventID: p.ProviderEventID,
		Payload:         payload,
		OccurredAt:      pgtype.Timestamptz{Time: p.OccurredAt, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("create delivery event: %w", err)
	}
	return &row, nil
}

func (r *DeliveryEventRepo) List(ctx context.Context, deliveryID string) ([]db.DeliveryEvent, error) {
	rows, err := r.q.GetDeliveryEvents(ctx, deliveryID)
	if err != nil {
		return nil, fmt.Errorf("list delivery events: %w", err)
	}
	return rows, nil
}
