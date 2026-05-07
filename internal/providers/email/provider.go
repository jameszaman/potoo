package email

import (
	"context"
	"encoding/json"
	"time"
)

type SendInput struct {
	To      string
	From    string
	ReplyTo string
	CC      []string
	BCC     []string
	Subject string
	HTML    string
	Text    string
}

type SendResult struct {
	ProviderMessageID string
	Raw               json.RawMessage
}

type WebhookEvent struct {
	EventID           string
	ProviderMessageID string
	EventType         string
	OccurredAt        time.Time
	Raw               json.RawMessage
}

// Provider is the interface every email adapter must implement.
type Provider interface {
	Send(ctx context.Context, input SendInput) (SendResult, error)
	VerifyWebhook(ctx context.Context, headers map[string]string, body []byte) error
	ParseWebhook(ctx context.Context, headers map[string]string, body []byte) ([]WebhookEvent, error)
}
