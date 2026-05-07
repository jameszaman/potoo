package handlers

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/potoo/potoo/internal/auth"
	"github.com/potoo/potoo/internal/db/repo"
	db "github.com/potoo/potoo/internal/db/sqlc"
	api "github.com/potoo/potoo/internal/gen/openapi"
)

func (h *Handlers) ListContacts(ctx context.Context, req api.ListContactsRequestObject) (api.ListContactsResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.ListContacts200JSONResponse{Data: []api.Contact{}}, nil
	}

	var tag *string
	if req.Params.Tag != nil && *req.Params.Tag != "" {
		tag = req.Params.Tag
	}

	rows, err := h.contacts.List(ctx, tenant.OrganizationID, tag)
	if err != nil {
		return nil, err
	}

	out := make([]api.Contact, len(rows))
	for i, c := range rows {
		out[i] = dbContactToAPI(c)
	}
	return api.ListContacts200JSONResponse{Data: out}, nil
}

func (h *Handlers) CreateContact(ctx context.Context, req api.CreateContactRequestObject) (api.CreateContactResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.CreateContact400JSONResponse(unauthorizedError(ctx)), nil
	}

	body := req.Body
	if body.Email == "" {
		return api.CreateContact400JSONResponse(validationError(ctx, "email is required")), nil
	}

	var name, phone, tag *string
	if body.Name != nil && *body.Name != "" {
		name = body.Name
	}
	if body.Phone != nil && *body.Phone != "" {
		phone = body.Phone
	}
	if body.Tag != nil && *body.Tag != "" {
		tag = body.Tag
	}

	c, err := h.contacts.Create(ctx, repo.CreateContactParams{
		OrgID: tenant.OrganizationID,
		Email: string(body.Email),
		Name:  name,
		Phone: phone,
		Tag:   tag,
	})
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return api.CreateContact409JSONResponse(conflictError(ctx, "A contact with that email already exists")), nil
		}
		return nil, err
	}

	out := dbContactToAPI(*c)
	return api.CreateContact201JSONResponse(out), nil
}

func (h *Handlers) GetContact(ctx context.Context, req api.GetContactRequestObject) (api.GetContactResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.GetContact404JSONResponse(notFoundError(ctx)), nil
	}

	c, err := h.contacts.Get(ctx, req.ContactId, tenant.OrganizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetContact404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}

	out := dbContactToAPI(*c)
	return api.GetContact200JSONResponse(out), nil
}

func (h *Handlers) UpdateContact(ctx context.Context, req api.UpdateContactRequestObject) (api.UpdateContactResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.UpdateContact404JSONResponse(notFoundError(ctx)), nil
	}

	body := req.Body
	var name, phone, tag *string
	if body.Name != nil {
		name = body.Name
	}
	if body.Phone != nil {
		phone = body.Phone
	}
	if body.Tag != nil {
		tag = body.Tag
	}

	c, err := h.contacts.Update(ctx, repo.UpdateContactParams{
		ID:    req.ContactId,
		OrgID: tenant.OrganizationID,
		Name:  name,
		Phone: phone,
		Tag:   tag,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.UpdateContact404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}

	out := dbContactToAPI(*c)
	return api.UpdateContact200JSONResponse(out), nil
}

func (h *Handlers) DeleteContact(ctx context.Context, req api.DeleteContactRequestObject) (api.DeleteContactResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.DeleteContact404JSONResponse(notFoundError(ctx)), nil
	}

	if err := h.contacts.Delete(ctx, req.ContactId, tenant.OrganizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.DeleteContact404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}

	return api.DeleteContact204Response{}, nil
}

func (h *Handlers) ListContactTags(ctx context.Context, req api.ListContactTagsRequestObject) (api.ListContactTagsResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.ListContactTags200JSONResponse{Data: []string{}}, nil
	}

	tags, err := h.contacts.ListTags(ctx, tenant.OrganizationID)
	if err != nil {
		return nil, err
	}

	return api.ListContactTags200JSONResponse{Data: tags}, nil
}

func dbContactToAPI(c db.Contact) api.Contact {
	contact := api.Contact{
		Id:        c.ID,
		OrgId:     c.OrgID,
		Email:     openapi_types.Email(c.Email),
		CreatedAt: c.CreatedAt.Time,
		UpdatedAt: c.UpdatedAt.Time,
	}
	if c.Name != nil {
		contact.Name = c.Name
	}
	if c.Phone != nil {
		contact.Phone = c.Phone
	}
	if c.Tag != nil {
		contact.Tag = c.Tag
	}
	return contact
}
