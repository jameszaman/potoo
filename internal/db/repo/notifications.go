package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/oklog/ulid/v2"

	"github.com/potoo/potoo/internal/db/sqlc"
)

type NotificationRepo struct {
	q *db.Queries
}

func NewNotificationRepo(pool *pgxpool.Pool) *NotificationRepo {
	return &NotificationRepo{q: db.New(pool)}
}

type CreateNotificationParams struct {
	OrganizationID string
	ProjectID      string
	EnvironmentID  string
	ExternalID     *string
	TemplateKey    *string
	Channel        db.NotificationChannel
	RecipientRef   *string
	Metadata       map[string]string
}

func (r *NotificationRepo) Create(ctx context.Context, p CreateNotificationParams) (*db.Notification, error) {
	meta, err := json.Marshal(p.Metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}

	row, err := r.q.CreateNotification(ctx, db.CreateNotificationParams{
		ID:             newID(),
		OrganizationID: p.OrganizationID,
		ProjectID:      p.ProjectID,
		EnvironmentID:  p.EnvironmentID,
		ExternalID:     p.ExternalID,
		TemplateKey:    p.TemplateKey,
		Channel:        p.Channel,
		RecipientRef:   p.RecipientRef,
		Status:         db.NotificationStatusQueued,
		Metadata:       meta,
	})
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	return &row, nil
}

func (r *NotificationRepo) Get(ctx context.Context, id, orgID string) (*db.Notification, error) {
	row, err := r.q.GetNotification(ctx, db.GetNotificationParams{
		ID:             id,
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get notification: %w", err)
	}
	return &row, nil
}

func (r *NotificationRepo) UpdateStatus(ctx context.Context, id string, status db.NotificationStatus) (*db.Notification, error) {
	row, err := r.q.UpdateNotificationStatus(ctx, db.UpdateNotificationStatusParams{
		ID:     id,
		Status: status,
	})
	if err != nil {
		return nil, fmt.Errorf("update notification status: %w", err)
	}
	return &row, nil
}

func newID() string {
	return ulid.Make().String()
}

func toPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

// timestampNow is a convenience alias used across repo files.
var timestampNow = func() time.Time { return time.Now().UTC() }
