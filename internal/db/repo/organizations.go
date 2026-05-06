package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/notifylayer/notifylayer/internal/db/sqlc"
)

type OrgRepo struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewOrgRepo(pool *pgxpool.Pool) *OrgRepo {
	return &OrgRepo{q: db.New(pool), pool: pool}
}

// CreateWithDefaults creates an org, a default project, and a production environment
// in a single transaction. This is the preferred call site for all org creation.
func (r *OrgRepo) CreateWithDefaults(ctx context.Context, name, slug string, orgType db.OrgType, website, description *string) (*db.Organization, error) {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	q := db.New(tx)

	org, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{
		ID:          newID(),
		Name:        name,
		Slug:        slug,
		Type:        orgType,
		Website:     website,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}

	project, err := q.CreateProject(ctx, db.CreateProjectParams{
		ID:             newID(),
		OrganizationID: org.ID,
		Name:           "default",
		Slug:           "default",
		IsDefault:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("create default project: %w", err)
	}

	_, err = q.CreateEnvironment(ctx, db.CreateEnvironmentParams{
		ID:        newID(),
		ProjectID: project.ID,
		Name:      db.EnvironmentNameProduction,
	})
	if err != nil {
		return nil, fmt.Errorf("create production environment: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &org, nil
}

// Create creates an org only — no default project/env. Use CreateWithDefaults for
// all normal org creation; this is kept for internal/test use.
func (r *OrgRepo) Create(ctx context.Context, name, slug string, orgType db.OrgType) (*db.Organization, error) {
	row, err := r.q.CreateOrganization(ctx, db.CreateOrganizationParams{
		ID:   newID(),
		Name: name,
		Slug: slug,
		Type: orgType,
	})
	if err != nil {
		return nil, fmt.Errorf("create organization: %w", err)
	}
	return &row, nil
}

func (r *OrgRepo) GetByID(ctx context.Context, id string) (*db.Organization, error) {
	row, err := r.q.GetOrganizationByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get organization: %w", err)
	}
	return &row, nil
}

func (r *OrgRepo) GetBySlug(ctx context.Context, slug string) (*db.Organization, error) {
	row, err := r.q.GetOrganizationBySlug(ctx, slug)
	if err != nil {
		return nil, fmt.Errorf("get organization by slug: %w", err)
	}
	return &row, nil
}

func (r *OrgRepo) GetPlatformOrg(ctx context.Context) (*db.Organization, error) {
	row, err := r.q.GetPlatformOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("get platform org: %w", err)
	}
	return &row, nil
}

func (r *OrgRepo) PlatformOrgExists(ctx context.Context) (bool, error) {
	exists, err := r.q.PlatformOrgExists(ctx)
	if err != nil {
		return false, fmt.Errorf("check platform org: %w", err)
	}
	return exists, nil
}

func (r *OrgRepo) ListAll(ctx context.Context) ([]db.Organization, error) {
	rows, err := r.q.ListOrganizations(ctx)
	if err != nil {
		return nil, fmt.Errorf("list organizations: %w", err)
	}
	return rows, nil
}

func (r *OrgRepo) GetDefaultProject(ctx context.Context, orgID string) (*db.Project, error) {
	row, err := r.q.GetDefaultProject(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get default project: %w", err)
	}
	return &row, nil
}

func (r *OrgRepo) GetProductionEnvironment(ctx context.Context, projectID string) (*db.Environment, error) {
	row, err := r.q.GetProductionEnvironment(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("get production environment: %w", err)
	}
	return &row, nil
}

func (r *OrgRepo) SetActive(ctx context.Context, orgID string, active bool) (*db.Organization, error) {
	row, err := r.q.SetOrgActive(ctx, db.SetOrgActiveParams{
		ID:       orgID,
		IsActive: active,
	})
	if err != nil {
		return nil, fmt.Errorf("set org active: %w", err)
	}
	return &row, nil
}

func (r *OrgRepo) GetStats(ctx context.Context, orgID string) (*db.GetOrgStatsRow, error) {
	row, err := r.q.GetOrgStats(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("get org stats: %w", err)
	}
	return &row, nil
}
