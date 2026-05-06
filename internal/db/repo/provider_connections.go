package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/potoo/potoo/internal/db/sqlc"
)

type ProviderConnectionRepo struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

func NewProviderConnectionRepo(pool *pgxpool.Pool) *ProviderConnectionRepo {
	return &ProviderConnectionRepo{pool: pool, q: db.New(pool)}
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
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := r.q.WithTx(tx)
	if p.IsDefault {
		if err := qtx.ClearDefaultProviders(ctx, db.ClearDefaultProvidersParams{
			OrganizationID: p.OrganizationID,
			ProjectID:      p.ProjectID,
			EnvironmentID:  p.EnvironmentID,
			Channel:        p.Channel,
		}); err != nil {
			return nil, fmt.Errorf("clear defaults: %w", err)
		}
	}
	row, err := qtx.CreateProviderConnection(ctx, db.CreateProviderConnectionParams{
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
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
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

func (r *ProviderConnectionRepo) GetAnyByType(ctx context.Context, providerType db.ProviderType) (*db.ProviderConnection, error) {
	row, err := r.q.GetProviderConnectionByType(ctx, providerType)
	if err != nil {
		return nil, fmt.Errorf("get provider by type: %w", err)
	}
	return &row, nil
}

func (r *ProviderConnectionRepo) Update(ctx context.Context, id, orgID, displayName, encryptedConfig string) (*db.ProviderConnection, error) {
	row, err := r.q.UpdateProviderConnection(ctx, db.UpdateProviderConnectionParams{
		ID:              id,
		OrganizationID:  orgID,
		DisplayName:     displayName,
		EncryptedConfig: encryptedConfig,
	})
	if err != nil {
		return nil, fmt.Errorf("update provider connection: %w", err)
	}
	return &row, nil
}

// SetDefault clears is_default on all connections for the same org/project/env/channel
// then sets it on the given connection, atomically.
func (r *ProviderConnectionRepo) SetDefault(ctx context.Context, id, orgID string) (*db.ProviderConnection, error) {
	conn, err := r.q.GetProviderConnection(ctx, db.GetProviderConnectionParams{
		ID:             id,
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get provider connection: %w", err)
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	qtx := r.q.WithTx(tx)
	if err := qtx.ClearDefaultProviders(ctx, db.ClearDefaultProvidersParams{
		OrganizationID: orgID,
		ProjectID:      conn.ProjectID,
		EnvironmentID:  conn.EnvironmentID,
		Channel:        conn.Channel,
	}); err != nil {
		return nil, fmt.Errorf("clear defaults: %w", err)
	}
	updated, err := qtx.SetProviderDefault(ctx, db.SetProviderDefaultParams{
		ID:             id,
		OrganizationID: orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("set default: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}
	return &updated, nil
}
