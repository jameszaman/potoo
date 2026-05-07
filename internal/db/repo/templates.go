package repo

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/potoo/potoo/internal/db/sqlc"
)

type TemplateRepo struct {
	q *db.Queries
}

func NewTemplateRepo(pool *pgxpool.Pool) *TemplateRepo {
	return &TemplateRepo{q: db.New(pool)}
}

func (r *TemplateRepo) Create(ctx context.Context, orgID, projectID, key, name string, channel db.NotificationChannel) (*db.Template, error) {
	row, err := r.q.CreateTemplate(ctx, db.CreateTemplateParams{
		ID:             newID(),
		OrganizationID: orgID,
		ProjectID:      projectID,
		Key:            key,
		Name:           name,
		Channel:        channel,
	})
	if err != nil {
		return nil, fmt.Errorf("create template: %w", err)
	}
	return &row, nil
}

func (r *TemplateRepo) GetByKey(ctx context.Context, orgID, projectID, key string) (*db.Template, error) {
	row, err := r.q.GetTemplateByKey(ctx, db.GetTemplateByKeyParams{
		OrganizationID: orgID,
		ProjectID:      projectID,
		Key:            key,
	})
	if err != nil {
		return nil, fmt.Errorf("get template: %w", err)
	}
	return &row, nil
}

func (r *TemplateRepo) List(ctx context.Context, orgID, projectID string) ([]db.Template, error) {
	rows, err := r.q.ListTemplates(ctx, db.ListTemplatesParams{
		OrganizationID: orgID,
		ProjectID:      projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	return rows, nil
}

type CreateVersionParams struct {
	TemplateID      string
	Subject         string
	HtmlBody        string
	TextBody        *string
	SmsBody         *string
	VariablesSchema map[string]string
	EditorBlocks    *string
}

func (r *TemplateRepo) CreateVersion(ctx context.Context, p CreateVersionParams) (*db.TemplateVersion, error) {
	next, err := r.q.NextVersionNumber(ctx, p.TemplateID)
	if err != nil {
		return nil, fmt.Errorf("next version number: %w", err)
	}

	schema, err := json.Marshal(p.VariablesSchema)
	if err != nil {
		return nil, fmt.Errorf("marshal variables schema: %w", err)
	}

	row, err := r.q.CreateTemplateVersion(ctx, db.CreateTemplateVersionParams{
		ID:              newID(),
		TemplateID:      p.TemplateID,
		VersionNumber:   next,
		Subject:         &p.Subject,
		HtmlBody:        &p.HtmlBody,
		TextBody:        p.TextBody,
		SmsBody:         p.SmsBody,
		VariablesSchema: schema,
		EditorBlocks:    p.EditorBlocks,
	})
	if err != nil {
		return nil, fmt.Errorf("create template version: %w", err)
	}
	return &row, nil
}

func (r *TemplateRepo) Activate(ctx context.Context, templateID string, versionNumber int32) (*db.Template, error) {
	if _, err := r.q.GetTemplateVersion(ctx, db.GetTemplateVersionParams{
		TemplateID:    templateID,
		VersionNumber: versionNumber,
	}); err != nil {
		return nil, fmt.Errorf("version not found: %w", err)
	}

	if err := r.q.ArchiveTemplateVersions(ctx, templateID); err != nil {
		return nil, fmt.Errorf("archive versions: %w", err)
	}

	if _, err := r.q.ActivateTemplateVersion(ctx, db.ActivateTemplateVersionParams{
		TemplateID:    templateID,
		VersionNumber: versionNumber,
	}); err != nil {
		return nil, fmt.Errorf("activate version: %w", err)
	}

	row, err := r.q.SetActiveVersion(ctx, db.SetActiveVersionParams{
		ID:            templateID,
		ActiveVersion: &versionNumber,
	})
	if err != nil {
		return nil, fmt.Errorf("set active version: %w", err)
	}
	return &row, nil
}

func (r *TemplateRepo) GetActiveVersion(ctx context.Context, orgID, projectID, key string) (*db.TemplateVersion, error) {
	row, err := r.q.GetActiveTemplateVersion(ctx, db.GetActiveTemplateVersionParams{
		OrganizationID: orgID,
		ProjectID:      projectID,
		Key:            key,
	})
	if err != nil {
		return nil, fmt.Errorf("get active version: %w", err)
	}
	return &row, nil
}
