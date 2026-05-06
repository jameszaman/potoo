package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/notifylayer/notifylayer/internal/db/sqlc"
)

type APIKeyRepo struct {
	q *db.Queries
}

func NewAPIKeyRepo(pool *pgxpool.Pool) *APIKeyRepo {
	return &APIKeyRepo{q: db.New(pool)}
}

func (r *APIKeyRepo) GetByHash(ctx context.Context, hash string) (*db.ApiKey, error) {
	key, err := r.q.GetAPIKeyByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("get api key: %w", err)
	}
	return &key, nil
}

func (r *APIKeyRepo) Touch(ctx context.Context, id string) error {
	if err := r.q.TouchAPIKey(ctx, id); err != nil {
		return fmt.Errorf("touch api key: %w", err)
	}
	return nil
}
