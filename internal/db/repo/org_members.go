package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/potoo/potoo/internal/db/sqlc"
)

type OrgMemberRepo struct {
	q *db.Queries
}

func NewOrgMemberRepo(pool *pgxpool.Pool) *OrgMemberRepo {
	return &OrgMemberRepo{q: db.New(pool)}
}

func (r *OrgMemberRepo) Create(ctx context.Context, orgID, userID string, role db.OrgRole) (*db.OrganizationMember, error) {
	row, err := r.q.CreateOrgMember(ctx, db.CreateOrgMemberParams{
		ID:     newID(),
		OrgID:  orgID,
		UserID: userID,
		Role:   role,
	})
	if err != nil {
		return nil, fmt.Errorf("create org member: %w", err)
	}
	return &row, nil
}

func (r *OrgMemberRepo) Get(ctx context.Context, orgID, userID string) (*db.OrganizationMember, error) {
	row, err := r.q.GetOrgMember(ctx, db.GetOrgMemberParams{
		OrgID:  orgID,
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("get org member: %w", err)
	}
	return &row, nil
}

func (r *OrgMemberRepo) ListUserOrgs(ctx context.Context, userID string) ([]db.Organization, error) {
	rows, err := r.q.ListUserOrgs(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list user orgs: %w", err)
	}
	return rows, nil
}

func (r *OrgMemberRepo) ListOrgMembers(ctx context.Context, orgID string) ([]db.OrganizationMember, error) {
	rows, err := r.q.ListOrgMembers(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list org members: %w", err)
	}
	return rows, nil
}

func (r *OrgMemberRepo) Delete(ctx context.Context, orgID, userID string) error {
	if err := r.q.DeleteOrgMember(ctx, db.DeleteOrgMemberParams{
		OrgID:  orgID,
		UserID: userID,
	}); err != nil {
		return fmt.Errorf("delete org member: %w", err)
	}
	return nil
}
