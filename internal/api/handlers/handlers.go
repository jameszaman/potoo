package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/notifylayer/notifylayer/internal/auth"
	"github.com/notifylayer/notifylayer/internal/db/repo"
	"github.com/notifylayer/notifylayer/internal/db/sqlc"
	api "github.com/notifylayer/notifylayer/internal/gen/openapi"
	tmpl "github.com/notifylayer/notifylayer/internal/template"
)

type Handlers struct {
	templates *repo.TemplateRepo
}

func New(pool *pgxpool.Pool) *Handlers {
	return &Handlers{
		templates: repo.NewTemplateRepo(pool),
	}
}

// --- System ---

func (h *Handlers) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return api.GetHealth200JSONResponse{Status: api.Ok}, nil
}

// --- Notifications (stub — implemented in Phase 6) ---

func (h *Handlers) SendNotification(ctx context.Context, _ api.SendNotificationRequestObject) (api.SendNotificationResponseObject, error) {
	if _, ok := auth.TenantFromContext(ctx); !ok {
		return api.SendNotification400JSONResponse(unauthorizedError(ctx)), nil
	}
	return api.SendNotification202JSONResponse{}, nil
}

func (h *Handlers) GetNotification(ctx context.Context, _ api.GetNotificationRequestObject) (api.GetNotificationResponseObject, error) {
	if _, ok := auth.TenantFromContext(ctx); !ok {
		return api.GetNotification404JSONResponse(notFoundError(ctx)), nil
	}
	return api.GetNotification404JSONResponse(notImplError(ctx)), nil
}

// --- Deliveries (stub — implemented in Phase 6) ---

func (h *Handlers) GetDelivery(ctx context.Context, _ api.GetDeliveryRequestObject) (api.GetDeliveryResponseObject, error) {
	if _, ok := auth.TenantFromContext(ctx); !ok {
		return api.GetDelivery404JSONResponse(notFoundError(ctx)), nil
	}
	return api.GetDelivery404JSONResponse(notImplError(ctx)), nil
}

func (h *Handlers) GetDeliveryEvents(ctx context.Context, _ api.GetDeliveryEventsRequestObject) (api.GetDeliveryEventsResponseObject, error) {
	if _, ok := auth.TenantFromContext(ctx); !ok {
		return api.GetDeliveryEvents404JSONResponse(notFoundError(ctx)), nil
	}
	return api.GetDeliveryEvents404JSONResponse(notImplError(ctx)), nil
}

// --- Templates ---

func (h *Handlers) ListTemplates(ctx context.Context, _ api.ListTemplatesRequestObject) (api.ListTemplatesResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.ListTemplates200JSONResponse{Data: []api.Template{}}, nil
	}

	rows, err := h.templates.List(ctx, tenant.OrganizationID, tenant.ProjectID)
	if err != nil {
		return nil, err
	}

	out := make([]api.Template, len(rows))
	for i, r := range rows {
		out[i] = dbTemplateToAPI(r)
	}
	return api.ListTemplates200JSONResponse{Data: out}, nil
}

func (h *Handlers) CreateTemplate(ctx context.Context, req api.CreateTemplateRequestObject) (api.CreateTemplateResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.CreateTemplate400JSONResponse(unauthorizedError(ctx)), nil
	}

	channel := db.NotificationChannel(req.Body.Channel)
	row, err := h.templates.Create(ctx, tenant.OrganizationID, tenant.ProjectID, req.Body.Key, req.Body.Name, channel)
	if err != nil {
		if strings.Contains(err.Error(), "unique") {
			return api.CreateTemplate409JSONResponse(errBody(ctx, "conflict", "A template with that key already exists")), nil
		}
		return nil, err
	}
	return api.CreateTemplate201JSONResponse(dbTemplateToAPI(*row)), nil
}

func (h *Handlers) GetTemplate(ctx context.Context, req api.GetTemplateRequestObject) (api.GetTemplateResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.GetTemplate404JSONResponse(notFoundError(ctx)), nil
	}

	row, err := h.templates.GetByKey(ctx, tenant.OrganizationID, tenant.ProjectID, req.TemplateKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetTemplate404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}
	return api.GetTemplate200JSONResponse(dbTemplateToAPI(*row)), nil
}

func (h *Handlers) CreateTemplateVersion(ctx context.Context, req api.CreateTemplateVersionRequestObject) (api.CreateTemplateVersionResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.CreateTemplateVersion404JSONResponse(notFoundError(ctx)), nil
	}

	tpl, err := h.templates.GetByKey(ctx, tenant.OrganizationID, tenant.ProjectID, req.TemplateKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.CreateTemplateVersion404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}

	var schema map[string]string
	if req.Body.VariablesSchema != nil {
		schema = *req.Body.VariablesSchema
	}

	version, err := h.templates.CreateVersion(ctx, repo.CreateVersionParams{
		TemplateID:      tpl.ID,
		Subject:         req.Body.Subject,
		HtmlBody:        req.Body.HtmlBody,
		TextBody:        req.Body.TextBody,
		SmsBody:         req.Body.SmsBody,
		VariablesSchema: schema,
	})
	if err != nil {
		return nil, err
	}
	return api.CreateTemplateVersion201JSONResponse(dbVersionToAPI(*version)), nil
}

func (h *Handlers) ActivateTemplateVersion(ctx context.Context, req api.ActivateTemplateVersionRequestObject) (api.ActivateTemplateVersionResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.ActivateTemplateVersion404JSONResponse(notFoundError(ctx)), nil
	}

	tpl, err := h.templates.GetByKey(ctx, tenant.OrganizationID, tenant.ProjectID, req.TemplateKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.ActivateTemplateVersion404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}

	updated, err := h.templates.Activate(ctx, tpl.ID, int32(req.Body.VersionNumber))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), "not found") {
			return api.ActivateTemplateVersion404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}
	return api.ActivateTemplateVersion200JSONResponse(dbTemplateToAPI(*updated)), nil
}

func (h *Handlers) RenderTemplate(ctx context.Context, req api.RenderTemplateRequestObject) (api.RenderTemplateResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.RenderTemplate404JSONResponse(notFoundError(ctx)), nil
	}

	version, err := h.templates.GetActiveVersion(ctx, tenant.OrganizationID, tenant.ProjectID, req.TemplateKey)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.RenderTemplate404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}

	data := map[string]any{}
	if req.Body.Data != nil {
		for k, v := range req.Body.Data {
			data[k] = v
		}
	}

	subject, err := renderField(version.Subject, data)
	if err != nil {
		return api.RenderTemplate400JSONResponse(errBody(ctx, "template_missing_variable", err.Error())), nil
	}
	htmlBody, err := renderField(version.HtmlBody, data)
	if err != nil {
		return api.RenderTemplate400JSONResponse(errBody(ctx, "template_missing_variable", err.Error())), nil
	}
	textBody, err := renderField(version.TextBody, data)
	if err != nil {
		return api.RenderTemplate400JSONResponse(errBody(ctx, "template_missing_variable", err.Error())), nil
	}
	smsBody, err := renderField(version.SmsBody, data)
	if err != nil {
		return api.RenderTemplate400JSONResponse(errBody(ctx, "template_missing_variable", err.Error())), nil
	}

	resp := api.RenderTemplateResponse{
		Subject:  subject,
		HtmlBody: htmlBody,
	}
	if textBody != "" {
		resp.TextBody = &textBody
	}
	if smsBody != "" {
		resp.SmsBody = &smsBody
	}
	return api.RenderTemplate200JSONResponse(resp), nil
}

// --- Helpers ---

func renderField(field *string, data map[string]any) (string, error) {
	if field == nil || *field == "" {
		return "", nil
	}
	return tmpl.Render(*field, data)
}

func dbTemplateToAPI(t db.Template) api.Template {
	out := api.Template{
		Id:        t.ID,
		Key:       t.Key,
		Name:      t.Name,
		Channel:   api.TemplateChannel(t.Channel),
		CreatedAt: t.CreatedAt.Time,
	}
	if t.UpdatedAt.Valid {
		out.UpdatedAt = &t.UpdatedAt.Time
	}
	if t.ActiveVersion != nil {
		v := int(*t.ActiveVersion)
		out.ActiveVersion = &v
	}
	return out
}

func dbVersionToAPI(v db.TemplateVersion) api.TemplateVersion {
	var schema map[string]string
	if len(v.VariablesSchema) > 0 {
		_ = json.Unmarshal(v.VariablesSchema, &schema)
	}
	out := api.TemplateVersion{
		Id:            v.ID,
		TemplateId:    v.TemplateID,
		VersionNumber: int(v.VersionNumber),
		Subject:       v.Subject,
		HtmlBody:      v.HtmlBody,
		TextBody:      v.TextBody,
		SmsBody:       v.SmsBody,
		Status:        api.TemplateVersionStatus(v.Status),
		CreatedAt:     v.CreatedAt.Time,
	}
	if schema != nil {
		out.VariablesSchema = &schema
	}
	return out
}

func reqID(ctx context.Context) string {
	return middleware.GetReqID(ctx)
}

func errBody(ctx context.Context, code, message string) api.Error {
	rid := reqID(ctx)
	return api.Error{
		Error: struct {
			Code      string  `json:"code"`
			DocsUrl   *string `json:"docs_url,omitempty"`
			Message   string  `json:"message"`
			RequestId string  `json:"request_id"`
		}{
			Code:      code,
			Message:   message,
			RequestId: rid,
		},
	}
}

func unauthorizedError(ctx context.Context) api.Error {
	return errBody(ctx, "auth_invalid_api_key", "Invalid or missing API key")
}

func notFoundError(ctx context.Context) api.Error {
	return errBody(ctx, "not_found", "Resource not found")
}

func notImplError(ctx context.Context) api.Error {
	return errBody(ctx, "not_implemented", "Not implemented")
}

// compile-time interface check
var _ api.StrictServerInterface = (*Handlers)(nil)

// suppress unused import warning — time is used via db model conversions
var _ = time.Now
