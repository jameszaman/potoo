package repo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/potoo/potoo/internal/db/sqlc"
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

type GeneratedKey struct {
	Raw    string // full key — shown to user once, never stored
	Prefix string // first 12 chars, safe to display
	Hash   string // sha256 hex — stored in DB
}

// GenerateAPIKey creates a cryptographically random key with an "nlk_" prefix.
func GenerateAPIKey() (GeneratedKey, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return GeneratedKey{}, fmt.Errorf("generate key: %w", err)
	}
	raw := "nlk_" + base64.RawURLEncoding.EncodeToString(b)
	prefix := raw[:12]
	h := sha256.Sum256([]byte(raw))
	hash := fmt.Sprintf("%x", h)
	return GeneratedKey{Raw: raw, Prefix: prefix, Hash: hash}, nil
}

func (r *APIKeyRepo) Create(ctx context.Context, orgID, name string, scopes []string) (*db.ApiKey, string, error) {
	gen, err := GenerateAPIKey()
	if err != nil {
		return nil, "", err
	}
	if scopes == nil {
		scopes = []string{}
	}
	key, err := r.q.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		ID:             newID(),
		OrganizationID: orgID,
		Name:           name,
		KeyPrefix:      gen.Prefix,
		KeyHash:        gen.Hash,
		Scopes:         scopes,
	})
	if err != nil {
		return nil, "", fmt.Errorf("create api key: %w", err)
	}
	return &key, gen.Raw, nil
}

func (r *APIKeyRepo) ListByOrg(ctx context.Context, orgID string) ([]db.ListAPIKeysByOrgRow, error) {
	rows, err := r.q.ListAPIKeysByOrg(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list api keys: %w", err)
	}
	return rows, nil
}

func (r *APIKeyRepo) Revoke(ctx context.Context, keyID, orgID string) error {
	if err := r.q.RevokeAPIKey(ctx, db.RevokeAPIKeyParams{
		ID:             keyID,
		OrganizationID: orgID,
	}); err != nil {
		return fmt.Errorf("revoke api key: %w", err)
	}
	return nil
}
