package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/notifylayer/notifylayer/internal/db/sqlc"
)

type ProviderConnectionRepo struct {
	q *db.Queries
}

func NewProviderConnectionRepo(pool *pgxpool.Pool) *ProviderConnectionRepo {
	return &ProviderConnectionRepo{q: db.New(pool)}
}

type CreateProviderConnectionParams struct {
	OrganizationID  string
	ProjectID       string
	EnvironmentID   string
	ProviderType    db.ProviderType
	Channel         db.ProviderChannel
	DisplayName     string
	EncryptedConfig string
	IsDefault       bool
}

func (r *ProviderConnectionRepo) Create(ctx context.Context, p CreateProviderConnectionParams) (*db.ProviderConnection, error) {
	row, err := r.q.CreateProviderConnection(ctx, db.CreateProviderConnectionParams{
		ID:              newID(),
		OrganizationID:  p.OrganizationID,
		ProjectID:       p.ProjectID,
		EnvironmentID:   p.EnvironmentID,
		ProviderType:    p.ProviderType,
		Channel:         p.Channel,
		DisplayName:     p.DisplayName,
		EncryptedConfig: p.EncryptedConfig,
		IsDefault:       p.IsDefault,
	})
	if err != nil {
		return nil, fmt.Errorf("create provider connection: %w", err)
	}
	return &row, nil
}

func (r *ProviderConnectionRepo) Get(ctx context.Context, id, orgID string) (*db.ProviderConnection, error) {
	row, err := r.q.GetProviderConnection(ctx, db.GetProviderConnectionParams{
		ID:             id,
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get provider connection: %w", err)
	}
	return &row, nil
}

func (r *ProviderConnectionRepo) List(ctx context.Context, orgID, projectID, envID string) ([]db.ProviderConnection, error) {
	rows, err := r.q.ListProviderConnections(ctx, db.ListProviderConnectionsParams{
		OrganizationID: orgID,
		ProjectID:      projectID,
		EnvironmentID:  envID,
	})
	if err != nil {
		return nil, fmt.Errorf("list provider connections: %w", err)
	}
	return rows, nil
}

func (r *ProviderConnectionRepo) Delete(ctx context.Context, id, orgID string) error {
	if err := r.q.DeleteProviderConnection(ctx, db.DeleteProviderConnectionParams{
		ID:             id,
		OrganizationID: orgID,
	}); err != nil {
		return fmt.Errorf("delete provider connection: %w", err)
	}
	return nil
}

func (r *ProviderConnectionRepo) GetDefault(ctx context.Context, orgID, projectID, envID string, channel db.ProviderChannel) (*db.ProviderConnection, error) {
	row, err := r.q.GetDefaultProvider(ctx, db.GetDefaultProviderParams{
		OrganizationID: orgID,
		ProjectID:      projectID,
		EnvironmentID:  envID,
		Channel:        channel,
	})
	if err != nil {
		return nil, fmt.Errorf("get default provider: %w", err)
	}
	return &row, nil
}
