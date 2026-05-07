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

	"github.com/potoo/potoo/internal/auth"
	"github.com/potoo/potoo/internal/db/repo"
	"github.com/potoo/potoo/internal/db/sqlc"
	api "github.com/potoo/potoo/internal/gen/openapi"
	"github.com/potoo/potoo/internal/queue"
	"github.com/potoo/potoo/internal/storage"
	tmpl "github.com/potoo/potoo/internal/template"
)

type Handlers struct {
	templates     *repo.TemplateRepo
	providers     *repo.ProviderConnectionRepo
	notifications *repo.NotificationRepo
	deliveries    *repo.DeliveryRepo
	events        *repo.DeliveryEventRepo
	users         *repo.UserRepo
	sessions      *repo.SessionRepo
	orgs          *repo.OrgRepo
	members       *repo.OrgMemberRepo
	apiKeys       *repo.APIKeyRepo
	invites       *repo.OrgInviteRepo
	queue         *queue.Client
	storage       storage.Driver
}

func New(pool *pgxpool.Pool, q *queue.Client, store storage.Driver) *Handlers {
	return &Handlers{
		templates:     repo.NewTemplateRepo(pool),
		providers:     repo.NewProviderConnectionRepo(pool),
		notifications: repo.NewNotificationRepo(pool),
		deliveries:    repo.NewDeliveryRepo(pool),
		events:        repo.NewDeliveryEventRepo(pool),
		users:         repo.NewUserRepo(pool),
		sessions:      repo.NewSessionRepo(pool),
		orgs:          repo.NewOrgRepo(pool),
		members:       repo.NewOrgMemberRepo(pool),
		apiKeys:       repo.NewAPIKeyRepo(pool),
		invites:       repo.NewOrgInviteRepo(pool),
		queue:         q,
		storage:       store,
	}
}

// --- System ---

func (h *Handlers) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return api.GetHealth200JSONResponse{Status: api.Ok}, nil
}

// --- Notifications (stub — implemented in Phase 6) ---

func (h *Handlers) SendNotification(ctx context.Context, req api.SendNotificationRequestObject) (api.SendNotificationResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.SendNotification400JSONResponse(unauthorizedError(ctx)), nil
	}

	body := req.Body
	channel := db.NotificationChannel(body.Channel)

	var templateKey *string
	if body.Template.Key != "" {
		templateKey = &body.Template.Key
	}

	var recipientRef *string
	if body.Recipient.ExternalId != nil {
		recipientRef = body.Recipient.ExternalId
	} else if body.Recipient.Email != nil {
		s := string(*body.Recipient.Email)
		recipientRef = &s
	}

	meta := map[string]string{}
	if body.Metadata != nil {
		meta = *body.Metadata
	}
	// Merge template data into metadata so the worker can render it.
	if body.Template.Data != nil {
		for k, v := range *body.Template.Data {
			if s, ok := v.(string); ok {
				meta[k] = s
			}
		}
	}

	notification, err := h.notifications.Create(ctx, repo.CreateNotificationParams{
		OrganizationID: tenant.OrganizationID,
		ProjectID:      tenant.ProjectID,
		EnvironmentID:  tenant.EnvironmentID,
		Channel:        channel,
		TemplateKey:    templateKey,
		RecipientRef:   recipientRef,
		Metadata:       meta,
	})
	if err != nil {
		return nil, err
	}

	delivery, err := h.deliveries.Create(ctx, notification.ID, tenant.OrganizationID, tenant.ProjectID, tenant.EnvironmentID, channel)
	if err != nil {
		return nil, err
	}

	task, err := queue.NewSendEmailTask(queue.SendEmailPayload{
		NotificationID: notification.ID,
		DeliveryID:     delivery.ID,
		OrganizationID: tenant.OrganizationID,
		ProjectID:      tenant.ProjectID,
		EnvironmentID:  tenant.EnvironmentID,
	})
	if err != nil {
		return nil, err
	}
	if err := h.queue.Enqueue(ctx, task); err != nil {
		return nil, err
	}

	return api.SendNotification202JSONResponse{
		NotificationId: notification.ID,
		DeliveryId:     delivery.ID,
		Status:         api.Queued,
	}, nil
}

func (h *Handlers) GetNotification(ctx context.Context, req api.GetNotificationRequestObject) (api.GetNotificationResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.GetNotification404JSONResponse(notFoundError(ctx)), nil
	}
	row, err := h.notifications.Get(ctx, req.NotificationId, tenant.OrganizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetNotification404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}
	return api.GetNotification200JSONResponse(dbNotificationToAPI(*row)), nil
}

// --- Deliveries ---

func (h *Handlers) GetDelivery(ctx context.Context, req api.GetDeliveryRequestObject) (api.GetDeliveryResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.GetDelivery404JSONResponse(notFoundError(ctx)), nil
	}
	row, err := h.deliveries.Get(ctx, req.DeliveryId, tenant.OrganizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetDelivery404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}
	return api.GetDelivery200JSONResponse(dbDeliveryToAPI(*row)), nil
}

func (h *Handlers) GetDeliveryEvents(ctx context.Context, req api.GetDeliveryEventsRequestObject) (api.GetDeliveryEventsResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.GetDeliveryEvents404JSONResponse(notFoundError(ctx)), nil
	}
	// Verify the delivery belongs to this tenant.
	if _, err := h.deliveries.Get(ctx, req.DeliveryId, tenant.OrganizationID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetDeliveryEvents404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}
	rows, err := h.events.List(ctx, req.DeliveryId)
	if err != nil {
		return nil, err
	}
	out := make([]api.DeliveryEvent, len(rows))
	for i, r := range rows {
		out[i] = dbDeliveryEventToAPI(r)
	}
	return api.GetDeliveryEvents200JSONResponse{Data: out}, nil
}

// --- Webhooks ---

// IngestEmailWebhook satisfies the strict server interface but is never called —
// the real handler (IngestEmailWebhookHTTP) is registered directly on the router
// so it can read raw bytes for signature verification.
func (h *Handlers) IngestEmailWebhook(_ context.Context, _ api.IngestEmailWebhookRequestObject) (api.IngestEmailWebhookResponseObject, error) {
	return api.IngestEmailWebhook200Response{}, nil
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

	out := dbTemplateToAPI(*row)
	if row.ActiveVersion != nil {
		av, err := h.templates.GetActiveVersion(ctx, tenant.OrganizationID, tenant.ProjectID, req.TemplateKey)
		if err == nil && av != nil {
			if av.Subject != nil {
				out.ActiveVersionSubject = av.Subject
			}
			if av.HtmlBody != nil {
				out.ActiveVersionHtmlBody = av.HtmlBody
			}
			out.ActiveVersionBlocks = av.EditorBlocks
		}
	}
	return api.GetTemplate200JSONResponse(out), nil
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
		EditorBlocks:    req.Body.EditorBlocks,
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

// --- Provider connections ---

func (h *Handlers) ListProviderConnections(ctx context.Context, _ api.ListProviderConnectionsRequestObject) (api.ListProviderConnectionsResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.ListProviderConnections200JSONResponse{Data: []api.ProviderConnection{}}, nil
	}
	rows, err := h.providers.List(ctx, tenant.OrganizationID, tenant.ProjectID, tenant.EnvironmentID)
	if err != nil {
		return nil, err
	}
	out := make([]api.ProviderConnection, len(rows))
	for i, r := range rows {
		out[i] = dbProviderToAPI(r)
	}
	return api.ListProviderConnections200JSONResponse{Data: out}, nil
}

func (h *Handlers) CreateProviderConnection(ctx context.Context, req api.CreateProviderConnectionRequestObject) (api.CreateProviderConnectionResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.CreateProviderConnection400JSONResponse(unauthorizedError(ctx)), nil
	}

	creds, err := json.Marshal(req.Body.Credentials)
	if err != nil {
		return api.CreateProviderConnection400JSONResponse(errBody(ctx, "validation_failed", "invalid credentials")), nil
	}

	isDefault := req.Body.IsDefault != nil && *req.Body.IsDefault

	row, err := h.providers.Create(ctx, repo.CreateProviderConnectionParams{
		OrganizationID:  tenant.OrganizationID,
		ProjectID:       tenant.ProjectID,
		EnvironmentID:   tenant.EnvironmentID,
		ProviderType:    db.ProviderType(req.Body.ProviderType),
		Channel:         db.ProviderChannel(req.Body.Channel),
		DisplayName:     req.Body.DisplayName,
		EncryptedConfig: string(creds),
		IsDefault:       isDefault,
	})
	if err != nil {
		return nil, err
	}
	return api.CreateProviderConnection201JSONResponse(dbProviderToAPI(*row)), nil
}

func (h *Handlers) GetProviderConnection(ctx context.Context, req api.GetProviderConnectionRequestObject) (api.GetProviderConnectionResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.GetProviderConnection404JSONResponse(notFoundError(ctx)), nil
	}
	row, err := h.providers.Get(ctx, req.ConnectionId, tenant.OrganizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.GetProviderConnection404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}
	return api.GetProviderConnection200JSONResponse(dbProviderToAPI(*row)), nil
}

func (h *Handlers) UpdateProviderConnection(ctx context.Context, req api.UpdateProviderConnectionRequestObject) (api.UpdateProviderConnectionResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.UpdateProviderConnection404JSONResponse(notFoundError(ctx)), nil
	}

	// Fetch existing connection to merge credentials.
	existing, err := h.providers.Get(ctx, req.ConnectionId, tenant.OrganizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.UpdateProviderConnection404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}

	// Merge new credentials over existing ones so omitted fields (e.g. api_key) are preserved.
	var merged map[string]string
	_ = json.Unmarshal([]byte(existing.EncryptedConfig), &merged)
	if merged == nil {
		merged = map[string]string{}
	}
	for k, v := range req.Body.Credentials {
		if v != "" {
			merged[k] = v
		}
	}
	credsJSON, err := json.Marshal(merged)
	if err != nil {
		return api.UpdateProviderConnection404JSONResponse(errBody(ctx, "validation_failed", "invalid credentials")), nil
	}

	row, err := h.providers.Update(ctx, req.ConnectionId, tenant.OrganizationID, req.Body.DisplayName, string(credsJSON))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.UpdateProviderConnection404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}
	return api.UpdateProviderConnection200JSONResponse(dbProviderToAPI(*row)), nil
}

func (h *Handlers) SetProviderDefault(ctx context.Context, req api.SetProviderDefaultRequestObject) (api.SetProviderDefaultResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.SetProviderDefault404JSONResponse(notFoundError(ctx)), nil
	}
	row, err := h.providers.SetDefault(ctx, req.ConnectionId, tenant.OrganizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.SetProviderDefault404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}
	return api.SetProviderDefault200JSONResponse(dbProviderToAPI(*row)), nil
}

func (h *Handlers) DeleteProviderConnection(ctx context.Context, req api.DeleteProviderConnectionRequestObject) (api.DeleteProviderConnectionResponseObject, error) {
	tenant, ok := auth.TenantFromContext(ctx)
	if !ok {
		return api.DeleteProviderConnection404JSONResponse(notFoundError(ctx)), nil
	}
	conn, err := h.providers.Get(ctx, req.ConnectionId, tenant.OrganizationID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return api.DeleteProviderConnection404JSONResponse(notFoundError(ctx)), nil
		}
		return nil, err
	}
	if conn.IsDefault {
		return api.DeleteProviderConnection400JSONResponse(errBody(ctx, "cannot_delete_default", "Cannot delete the default provider. Set another provider as default first.")), nil
	}
	if err := h.providers.Delete(ctx, req.ConnectionId, tenant.OrganizationID); err != nil {
		return nil, err
	}
	return api.DeleteProviderConnection204Response{}, nil
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

func dbProviderToAPI(p db.ProviderConnection) api.ProviderConnection {
	out := api.ProviderConnection{
		Id:          p.ID,
		ProviderType: api.ProviderType(p.ProviderType),
		Channel:     api.ProviderConnectionChannel(p.Channel),
		DisplayName: p.DisplayName,
		IsDefault:   p.IsDefault,
		IsActive:    p.IsActive,
		CreatedAt:   p.CreatedAt.Time,
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
		EditorBlocks:  v.EditorBlocks,
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

func dbNotificationToAPI(n db.Notification) api.Notification {
	var meta map[string]string
	if len(n.Metadata) > 0 {
		_ = json.Unmarshal(n.Metadata, &meta)
	}
	out := api.Notification{
		Id:          n.ID,
		Channel:     string(n.Channel),
		Status:      api.NotificationStatus(n.Status),
		TemplateKey: n.TemplateKey,
		RecipientRef: n.RecipientRef,
		CreatedAt:   n.CreatedAt.Time,
	}
	if meta != nil {
		out.Metadata = &meta
	}
	if n.UpdatedAt.Valid {
		out.UpdatedAt = &n.UpdatedAt.Time
	}
	return out
}

func dbDeliveryToAPI(d db.Delivery) api.Delivery {
	out := api.Delivery{
		Id:               d.ID,
		NotificationId:   d.NotificationID,
		Channel:          string(d.Channel),
		Status:           api.NotificationStatus(d.Status),
		AttemptCount:     int(d.AttemptCount),
		ProviderType:     d.ProviderType,
		ProviderMessageId: d.ProviderMessageID,
		LastErrorCode:    d.LastErrorCode,
		LastErrorMessage: d.LastErrorMessage,
		CreatedAt:        d.CreatedAt.Time,
	}
	if d.SentAt.Valid {
		out.SentAt = &d.SentAt.Time
	}
	if d.DeliveredAt.Valid {
		out.DeliveredAt = &d.DeliveredAt.Time
	}
	if d.FailedAt.Valid {
		out.FailedAt = &d.FailedAt.Time
	}
	if d.UpdatedAt.Valid {
		out.UpdatedAt = &d.UpdatedAt.Time
	}
	return out
}

func dbDeliveryEventToAPI(e db.DeliveryEvent) api.DeliveryEvent {
	out := api.DeliveryEvent{
		Id:              e.ID,
		DeliveryId:      e.DeliveryID,
		EventType:       api.DeliveryEventEventType(e.EventType),
		ProviderType:    e.ProviderType,
		ProviderEventId: e.ProviderEventID,
		OccurredAt:      e.OccurredAt.Time,
		CreatedAt:       e.CreatedAt.Time,
	}
	return out
}

// compile-time interface check
var _ api.StrictServerInterface = (*Handlers)(nil)

// suppress unused import warning — time is used via db model conversions
var _ = time.Now
