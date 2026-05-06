package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/potoo/potoo/internal/db/repo"
	db "github.com/potoo/potoo/internal/db/sqlc"
	"github.com/potoo/potoo/internal/providers/email"
	"github.com/potoo/potoo/internal/providers/email/resend"
	"github.com/potoo/potoo/internal/providers/email/sendgrid"
)

// IngestEmailWebhookHTTP is a plain http.HandlerFunc registered outside the strict
// handler. We need raw bytes before JSON decoding so signature verification works.
func (h *Handlers) IngestEmailWebhookHTTP(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeWebhookError(w, r, "read_error", "could not read request body")
		return
	}

	headers := lowerHeaders(r)
	ctx := r.Context()

	events, err := h.verifyAndParse(ctx, provider, headers, body)
	if err != nil {
		slog.Warn("webhook rejected", "provider", provider, "err", err)
		writeWebhookError(w, r, "invalid_webhook", err.Error())
		return
	}

	for _, evt := range events {
		if err := h.processWebhookEvent(ctx, provider, evt); err != nil {
			slog.Error("webhook event processing failed", "provider", provider, "err", err)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handlers) verifyAndParse(ctx context.Context, provider string, headers map[string]string, body []byte) ([]email.WebhookEvent, error) {
	switch provider {
	case "resend":
		conn, err := h.providers.GetAnyByType(ctx, db.ProviderTypeResend)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		_, apiKey, webhookKey := parseCreds(conn.EncryptedConfig)
		a := resend.New(apiKey, webhookKey)
		if err := a.VerifyWebhook(ctx, headers, body); err != nil {
			return nil, err
		}
		return a.ParseWebhook(ctx, headers, body)

	case "sendgrid":
		conn, err := h.providers.GetAnyByType(ctx, db.ProviderTypeSendgrid)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, nil
			}
			return nil, err
		}
		_, apiKey, webhookKey := parseCreds(conn.EncryptedConfig)
		a := sendgrid.New(apiKey, webhookKey)
		if err := a.VerifyWebhook(ctx, headers, body); err != nil {
			return nil, err
		}
		return a.ParseWebhook(ctx, headers, body)

	default:
		return nil, nil
	}
}

func (h *Handlers) processWebhookEvent(ctx context.Context, provider string, evt email.WebhookEvent) error {
	if evt.ProviderMessageID == "" {
		return nil
	}

	delivery, err := h.deliveries.GetByProviderMessageID(ctx, evt.ProviderMessageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	providerStr := provider
	h.recordDeliveryEvent(ctx, delivery.ID, db.DeliveryEventType(evt.EventType), &providerStr, nilIfEmpty(evt.EventID), evt.OccurredAt)

	switch evt.EventType {
	case "delivered":
		_, _ = h.deliveries.MarkDelivered(ctx, delivery.ID)
		_, _ = h.notifications.UpdateStatus(ctx, delivery.NotificationID, db.NotificationStatusDelivered)
	case "bounced":
		_, _ = h.deliveries.MarkBounced(ctx, delivery.ID)
		_, _ = h.notifications.UpdateStatus(ctx, delivery.NotificationID, db.NotificationStatusBounced)
	case "complained":
		_, _ = h.deliveries.UpdateStatus(ctx, delivery.ID, db.NotificationStatusComplained)
		_, _ = h.notifications.UpdateStatus(ctx, delivery.NotificationID, db.NotificationStatusComplained)
	}

	return nil
}

func (h *Handlers) recordDeliveryEvent(ctx context.Context, deliveryID string, eventType db.DeliveryEventType, providerType *string, providerEventID *string, occurredAt time.Time) {
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

func lowerHeaders(r *http.Request) map[string]string {
	out := make(map[string]string, len(r.Header))
	for k, v := range r.Header {
		if len(v) > 0 {
			out[strings.ToLower(k)] = v[0]
		}
	}
	return out
}

func parseCreds(encryptedConfig string) (map[string]string, string, string) {
	var creds map[string]string
	_ = json.Unmarshal([]byte(encryptedConfig), &creds)
	return creds, creds["api_key"], creds["webhook_key"]
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func writeWebhookError(w http.ResponseWriter, r *http.Request, code, message string) {
	rid := r.Header.Get("X-Request-Id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":       code,
			"message":    message,
			"request_id": rid,
		},
	})
}
