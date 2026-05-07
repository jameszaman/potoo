package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/potoo/potoo/internal/db/repo"
	"github.com/potoo/potoo/internal/db/sqlc"
	"github.com/potoo/potoo/internal/providers/email"
	"github.com/potoo/potoo/internal/providers/email/resend"
	"github.com/potoo/potoo/internal/providers/email/sendgrid"
	emailsmtp "github.com/potoo/potoo/internal/providers/email/smtp"
	"github.com/potoo/potoo/internal/queue"
	tmpl "github.com/potoo/potoo/internal/template"
)

type EmailHandler struct {
	notifications *repo.NotificationRepo
	deliveries    *repo.DeliveryRepo
	events        *repo.DeliveryEventRepo
	providers     *repo.ProviderConnectionRepo
	templates     *repo.TemplateRepo
}

func NewEmailHandler(pool *pgxpool.Pool) *EmailHandler {
	return &EmailHandler{
		notifications: repo.NewNotificationRepo(pool),
		deliveries:    repo.NewDeliveryRepo(pool),
		events:        repo.NewDeliveryEventRepo(pool),
		providers:     repo.NewProviderConnectionRepo(pool),
		templates:     repo.NewTemplateRepo(pool),
	}
}

func (h *EmailHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	p, err := queue.ParseSendEmailPayload(t)
	if err != nil {
		return fmt.Errorf("parse payload: %w", err)
	}

	delivery, err := h.deliveries.Get(ctx, p.DeliveryID, p.OrganizationID)
	if err != nil {
		return fmt.Errorf("get delivery: %w", err)
	}

	notification, err := h.notifications.Get(ctx, p.NotificationID, p.OrganizationID)
	if err != nil {
		return fmt.Errorf("get notification: %w", err)
	}

	// Mark processing.
	if _, err := h.deliveries.UpdateStatus(ctx, delivery.ID, db.NotificationStatusProcessing); err != nil {
		return fmt.Errorf("mark processing: %w", err)
	}

	result, sendErr := h.send(ctx, notification, delivery)

	if sendErr != nil {
		slog.Error("delivery failed", "delivery_id", delivery.ID, "err", sendErr)
		if _, err := h.deliveries.UpdateStatus(ctx, delivery.ID, db.NotificationStatusFailedTemporary); err != nil {
			slog.Error("update delivery status", "err", err)
		}
		h.recordEvent(ctx, delivery.ID, db.DeliveryEventTypeFailed, nil, nil, time.Now())
		// Return the error so Asynq retries the task.
		return sendErr
	}

	if _, err := h.deliveries.MarkSent(ctx, delivery.ID, &result.ProviderType, &result.ProviderMessageID); err != nil {
		slog.Error("mark delivery sent", "err", err)
	}
	if _, err := h.notifications.UpdateStatus(ctx, notification.ID, db.NotificationStatusSent); err != nil {
		slog.Error("update notification status", "err", err)
	}
	h.recordEvent(ctx, delivery.ID, db.DeliveryEventTypeSent, &result.ProviderType, nil, time.Now())

	slog.Info("email sent",
		"delivery_id", delivery.ID,
		"provider", result.ProviderType,
		"message_id", result.ProviderMessageID,
	)
	return nil
}

type sendResult struct {
	ProviderType      string
	ProviderMessageID string
}

func (h *EmailHandler) send(ctx context.Context, n *db.Notification, d *db.Delivery) (sendResult, error) {
	// Resolve provider connection — permanent if missing.
	conn, err := h.providers.GetDefault(ctx, n.OrganizationID, n.ProjectID, n.EnvironmentID, db.ProviderChannelEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sendResult{}, fmt.Errorf("%w: no default email provider configured — add one on the Providers page", asynq.SkipRetry)
		}
		return sendResult{}, fmt.Errorf("resolve provider: %w", err)
	}

	// Bad stored credentials — permanent.
	provider, err := buildProvider(conn)
	if err != nil {
		return sendResult{}, fmt.Errorf("%w: invalid provider credentials: %s", asynq.SkipRetry, err)
	}

	var creds map[string]string
	_ = json.Unmarshal([]byte(conn.EncryptedConfig), &creds)

	// Render template — permanent if template missing or misconfigured.
	input, err := h.buildSendInput(ctx, n, creds)
	if err != nil {
		return sendResult{}, fmt.Errorf("%w: %s", asynq.SkipRetry, err)
	}

	result, err := provider.Send(ctx, input)
	if err != nil {
		// 4xx from the provider = permanent (bad API key, unverified domain, invalid request).
		// 5xx or network errors = transient, worth retrying.
		msg := err.Error()
		if isPermanentProviderError(msg) {
			return sendResult{}, fmt.Errorf("%w: %s", asynq.SkipRetry, msg)
		}
		return sendResult{}, fmt.Errorf("provider send: %w", err)
	}

	return sendResult{
		ProviderType:      string(conn.ProviderType),
		ProviderMessageID: result.ProviderMessageID,
	}, nil
}

func (h *EmailHandler) buildSendInput(ctx context.Context, n *db.Notification, creds map[string]string) (email.SendInput, error) {
	if n.TemplateKey == nil {
		return email.SendInput{}, fmt.Errorf("notification has no template key")
	}

	version, err := h.templates.GetActiveVersion(ctx, n.OrganizationID, n.ProjectID, *n.TemplateKey)
	if err != nil {
		return email.SendInput{}, fmt.Errorf("get active template version: %w", err)
	}

	// Parse metadata as template data.
	data := map[string]any{}
	if len(n.Metadata) > 0 {
		var meta map[string]any
		if err := json.Unmarshal(n.Metadata, &meta); err == nil {
			data = meta
		}
	}

	subject, err := tmpl.Render(strVal(version.Subject), data)
	if err != nil {
		return email.SendInput{}, fmt.Errorf("render subject: %w", err)
	}
	htmlBody, err := tmpl.Render(strVal(version.HtmlBody), data)
	if err != nil {
		return email.SendInput{}, fmt.Errorf("render html body: %w", err)
	}
	textBody, _ := tmpl.Render(strVal(version.TextBody), data)

	recipient := strVal(n.RecipientRef)

	from := creds["from_email"]
	if from == "" {
		return email.SendInput{}, fmt.Errorf("%w: provider connection is missing 'from_email' — set it on the Providers page", asynq.SkipRetry)
	}
	fromName := creds["from_name"]
	if fromName != "" {
		from = fmt.Sprintf("%s <%s>", fromName, from)
	}

	var cc, bcc []string
	if raw, ok := data["__cc"]; ok {
		if s, ok := raw.(string); ok {
			_ = json.Unmarshal([]byte(s), &cc)
		}
	}
	if raw, ok := data["__bcc"]; ok {
		if s, ok := raw.(string); ok {
			_ = json.Unmarshal([]byte(s), &bcc)
		}
	}

	return email.SendInput{
		To:      recipient,
		From:    from,
		CC:      cc,
		BCC:     bcc,
		Subject: subject,
		HTML:    htmlBody,
		Text:    textBody,
	}, nil
}

func (h *EmailHandler) recordEvent(ctx context.Context, deliveryID string, eventType db.DeliveryEventType, providerType *string, providerEventID *string, occurredAt time.Time) {
	_, err := h.events.Create(ctx, repo.CreateDeliveryEventParams{
		DeliveryID:      deliveryID,
		EventType:       eventType,
		ProviderType:    providerType,
		ProviderEventID: providerEventID,
		Payload:         map[string]any{},
		OccurredAt:      occurredAt,
	})
	if err != nil {
		slog.Error("record delivery event", "err", err)
	}
}

func buildProvider(conn *db.ProviderConnection) (email.Provider, error) {
	var creds map[string]string
	if err := json.Unmarshal([]byte(conn.EncryptedConfig), &creds); err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}
	apiKey := creds["api_key"]
	webhookKey := creds["webhook_key"]

	switch conn.ProviderType {
	case db.ProviderTypeResend:
		return resend.New(apiKey, webhookKey), nil
	case db.ProviderTypeSendgrid:
		return sendgrid.New(apiKey, webhookKey), nil
	case db.ProviderTypeSmtp:
		return emailsmtp.NewFromCreds(creds)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", conn.ProviderType)
	}
}

// isPermanentProviderError returns true for 4xx provider errors (bad API key,
// unverified domain, invalid request) that will never succeed on retry.
func isPermanentProviderError(msg string) bool {
	return strings.Contains(msg, "error 400") ||
		strings.Contains(msg, "error 401") ||
		strings.Contains(msg, "error 403") ||
		strings.Contains(msg, "error 404") ||
		strings.Contains(msg, "error 422")
}

func strVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
