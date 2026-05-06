package repo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/potoo/potoo/internal/db/sqlc"
)

type SessionRepo struct {
	q *db.Queries
}

func NewSessionRepo(pool *pgxpool.Pool) *SessionRepo {
	return &SessionRepo{q: db.New(pool)}
}

// GenerateRefreshToken returns a cryptographically random opaque token and its SHA-256 hash.
func GenerateRefreshToken() (token, hash string, err error) {
	b := make([]byte, 32)
	if _, err = rand.Read(b); err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	token = hex.EncodeToString(b)
	h := sha256.Sum256([]byte(token))
	hash = hex.EncodeToString(h[:])
	return token, hash, nil
}

func HashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

func (r *SessionRepo) Create(ctx context.Context, userID, orgID, refreshTokenHash string, expiresAt time.Time) (*db.Session, error) {
	row, err := r.q.CreateSession(ctx, db.CreateSessionParams{
		ID:               newID(),
		UserID:           userID,
		OrgID:            orgID,
		RefreshTokenHash: refreshTokenHash,
		ExpiresAt:        toPgTimestamptz(expiresAt),
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &row, nil
}

func (r *SessionRepo) GetByID(ctx context.Context, id string) (*db.Session, error) {
	row, err := r.q.GetSessionByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get session by id: %w", err)
	}
	return &row, nil
}

func (r *SessionRepo) GetByRefreshTokenHash(ctx context.Context, hash string) (*db.Session, error) {
	row, err := r.q.GetSessionByRefreshTokenHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return &row, nil
}

func (r *SessionRepo) Rotate(ctx context.Context, sessionID, newHash string, newExpiry time.Time) (*db.Session, error) {
	row, err := r.q.RotateSession(ctx, db.RotateSessionParams{
		ID:               sessionID,
		RefreshTokenHash: newHash,
		ExpiresAt:        toPgTimestamptz(newExpiry),
	})
	if err != nil {
		return nil, fmt.Errorf("rotate session: %w", err)
	}
	return &row, nil
}

func (r *SessionRepo) Delete(ctx context.Context, sessionID string) error {
	if err := r.q.DeleteSession(ctx, sessionID); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (r *SessionRepo) DeleteAllForUser(ctx context.Context, userID string) error {
	if err := r.q.DeleteAllUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("delete all user sessions: %w", err)
	}
	return nil
}

func (r *SessionRepo) DeleteAllForUserOrg(ctx context.Context, userID, orgID string) error {
	if err := r.q.DeleteAllUserOrgSessions(ctx, db.DeleteAllUserOrgSessionsParams{
		UserID: userID,
		OrgID:  orgID,
	}); err != nil {
		return fmt.Errorf("delete user org sessions: %w", err)
	}
	return nil
}
