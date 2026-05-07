package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/potoo/potoo/internal/db/sqlc"
)

type ContactRepo struct {
	q *db.Queries
}

func NewContactRepo(pool *pgxpool.Pool) *ContactRepo {
	return &ContactRepo{q: db.New(pool)}
}

type CreateContactParams struct {
	OrgID string
	Email string
	Name  *string
	Phone *string
	Tag   *string
}

func (r *ContactRepo) Create(ctx context.Context, p CreateContactParams) (*db.Contact, error) {
	row, err := r.q.CreateContact(ctx, db.CreateContactParams{
		ID:    newID(),
		OrgID: p.OrgID,
		Email: p.Email,
		Name:  p.Name,
		Phone: p.Phone,
		Tag:   p.Tag,
	})
	if err != nil {
		return nil, fmt.Errorf("create contact: %w", err)
	}
	return &row, nil
}

func (r *ContactRepo) Get(ctx context.Context, id, orgID string) (*db.Contact, error) {
	row, err := r.q.GetContact(ctx, db.GetContactParams{ID: id, OrgID: orgID})
	if err != nil {
		return nil, fmt.Errorf("get contact: %w", err)
	}
	return &row, nil
}

func (r *ContactRepo) List(ctx context.Context, orgID string, tag *string) ([]db.Contact, error) {
	if tag != nil {
		rows, err := r.q.ListContactsByTag(ctx, db.ListContactsByTagParams{OrgID: orgID, Tag: tag})
		if err != nil {
			return nil, fmt.Errorf("list contacts by tag: %w", err)
		}
		return rows, nil
	}
	rows, err := r.q.ListContacts(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	return rows, nil
}

func (r *ContactRepo) ListTags(ctx context.Context, orgID string) ([]string, error) {
	rows, err := r.q.ListContactTags(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list contact tags: %w", err)
	}
	tags := make([]string, 0, len(rows))
	for _, t := range rows {
		if t != nil {
			tags = append(tags, *t)
		}
	}
	return tags, nil
}

type UpdateContactParams struct {
	ID    string
	OrgID string
	Name  *string
	Phone *string
	Tag   *string
}

func (r *ContactRepo) Update(ctx context.Context, p UpdateContactParams) (*db.Contact, error) {
	row, err := r.q.UpdateContact(ctx, db.UpdateContactParams{
		ID:    p.ID,
		OrgID: p.OrgID,
		Name:  p.Name,
		Phone: p.Phone,
		Tag:   p.Tag,
	})
	if err != nil {
		return nil, fmt.Errorf("update contact: %w", err)
	}
	return &row, nil
}

func (r *ContactRepo) Delete(ctx context.Context, id, orgID string) error {
	if err := r.q.DeleteContact(ctx, db.DeleteContactParams{ID: id, OrgID: orgID}); err != nil {
		return fmt.Errorf("delete contact: %w", err)
	}
	return nil
}
