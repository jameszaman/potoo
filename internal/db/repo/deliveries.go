package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/notifylayer/notifylayer/internal/db/sqlc"
)

type DeliveryRepo struct {
	q *db.Queries
}

func NewDeliveryRepo(pool *pgxpool.Pool) *DeliveryRepo {
	return &DeliveryRepo{q: db.New(pool)}
}

func (r *DeliveryRepo) Create(ctx context.Context, notificationID, orgID, projectID, envID string, channel db.NotificationChannel) (*db.Delivery, error) {
	row, err := r.q.CreateDelivery(ctx, db.CreateDeliveryParams{
		ID:             newID(),
		NotificationID: notificationID,
		OrganizationID: orgID,
		ProjectID:      projectID,
		EnvironmentID:  envID,
		Channel:        channel,
		Status:         db.NotificationStatusQueued,
	})
	if err != nil {
		return nil, fmt.Errorf("create delivery: %w", err)
	}
	return &row, nil
}

func (r *DeliveryRepo) Get(ctx context.Context, id, orgID string) (*db.Delivery, error) {
	row, err := r.q.GetDelivery(ctx, db.GetDeliveryParams{
		ID:             id,
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get delivery: %w", err)
	}
	return &row, nil
}

func (r *DeliveryRepo) ListByNotification(ctx context.Context, notificationID, orgID string) ([]db.Delivery, error) {
	rows, err := r.q.GetDeliveriesByNotification(ctx, db.GetDeliveriesByNotificationParams{
		NotificationID: notificationID,
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("list deliveries: %w", err)
	}
	return rows, nil
}

func (r *DeliveryRepo) UpdateStatus(ctx context.Context, id string, status db.NotificationStatus) (*db.Delivery, error) {
	row, err := r.q.UpdateDeliveryStatus(ctx, db.UpdateDeliveryStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return nil, fmt.Errorf("update delivery status: %w", err)
	}
	return &row, nil
}

func (r *DeliveryRepo) MarkSent(ctx context.Context, id string, providerType, providerMessageID *string) (*db.Delivery, error) {
	row, err := r.q.UpdateDeliverySent(ctx, db.UpdateDeliverySentParams{
		ID:                id,
		ProviderType:      providerType,
		ProviderMessageID: providerMessageID,
	})
	if err != nil {
		return nil, fmt.Errorf("mark delivery sent: %w", err)
	}
	return &row, nil
}
