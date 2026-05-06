package repo

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/notifylayer/notifylayer/internal/db/sqlc"
)

const inviteTTL = 7 * 24 * time.Hour

type OrgInviteRepo struct {
	q *db.Queries
}

func NewOrgInviteRepo(pool *pgxpool.Pool) *OrgInviteRepo {
	return &OrgInviteRepo{q: db.New(pool)}
}

func generateInviteToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (r *OrgInviteRepo) Create(ctx context.Context, orgID, createdBy string) (*db.OrgInvite, string, error) {
	token, err := generateInviteToken()
	if err != nil {
		return nil, "", fmt.Errorf("generate token: %w", err)
	}
	expiresAt := pgtype.Timestamptz{Time: time.Now().Add(inviteTTL), Valid: true}
	row, err := r.q.CreateOrgInvite(ctx, db.CreateOrgInviteParams{
		ID:        newID(),
		OrgID:     orgID,
		CreatedBy: createdBy,
		Token:     token,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, "", fmt.Errorf("create invite: %w", err)
	}
	return &row, token, nil
}

func (r *OrgInviteRepo) GetByToken(ctx context.Context, token string) (*db.OrgInvite, error) {
	row, err := r.q.GetOrgInviteByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("get invite: %w", err)
	}
	return &row, nil
}

func (r *OrgInviteRepo) MarkUsed(ctx context.Context, id string) error {
	_, err := r.q.MarkOrgInviteUsed(ctx, id)
	return err
}

func (r *OrgInviteRepo) ListByOrg(ctx context.Context, orgID string) ([]db.OrgInvite, error) {
	rows, err := r.q.ListOrgInvites(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list invites: %w", err)
	}
	return rows, nil
}
