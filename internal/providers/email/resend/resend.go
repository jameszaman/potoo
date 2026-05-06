package resend

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/notifylayer/notifylayer/internal/providers/email"
)

const baseURL = "https://api.resend.com"

type Adapter struct {
	apiKey     string
	webhookKey string
	httpClient *http.Client
}

func New(apiKey, webhookKey string) *Adapter {
	return &Adapter{
		apiKey:     apiKey,
		webhookKey: webhookKey,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (a *Adapter) Send(ctx context.Context, input email.SendInput) (email.SendResult, error) {
	body, _ := json.Marshal(map[string]any{
		"from":    input.From,
		"to":      []string{input.To},
		"subject": input.Subject,
		"html":    input.HTML,
		"text":    input.Text,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/emails", bytes.NewReader(body))
	if err != nil {
		return email.SendResult{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return email.SendResult{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return email.SendResult{}, fmt.Errorf("resend error %d: %s", resp.StatusCode, raw)
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(raw, &result); err != nil {
		return email.SendResult{}, fmt.Errorf("decode response: %w", err)
	}

	return email.SendResult{
		ProviderMessageID: result.ID,
		Raw:               raw,
	}, nil
}

func (a *Adapter) VerifyWebhook(_ context.Context, headers map[string]string, body []byte) error {
	if a.webhookKey == "" {
		return nil
	}
	sig := headers["svix-signature"]
	id := headers["svix-id"]
	ts := headers["svix-timestamp"]
	if sig == "" || id == "" || ts == "" {
		return fmt.Errorf("missing svix headers")
	}
	mac := hmac.New(sha256.New, []byte(a.webhookKey))
	fmt.Fprintf(mac, "%s.%s.%s", id, ts, body)
	expected := "v1," + hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(sig)) {
		return fmt.Errorf("invalid webhook signature")
	}
	return nil
}

func (a *Adapter) ParseWebhook(_ context.Context, _ map[string]string, body []byte) ([]email.WebhookEvent, error) {
	var payload struct {
		Type      string `json:"type"`
		CreatedAt string `json:"created_at"`
		Data      struct {
			EmailID string `json:"email_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("decode webhook: %w", err)
	}

	occurredAt, _ := time.Parse(time.RFC3339, payload.CreatedAt)

	return []email.WebhookEvent{{
		EventType:         normalizeEventType(payload.Type),
		ProviderMessageID: payload.Data.EmailID,
		OccurredAt:        occurredAt,
		Raw:               body,
	}}, nil
}

func normalizeEventType(t string) string {
	switch t {
	case "email.sent":
		return "sent"
	case "email.delivered":
		return "delivered"
	case "email.bounced":
		return "bounced"
	case "email.complained":
		return "complained"
	case "email.opened":
		return "opened"
	case "email.clicked":
		return "clicked"
	default:
		return t
	}
}
