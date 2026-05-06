package sendgrid

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/potoo/potoo/internal/providers/email"
)

const baseURL = "https://api.sendgrid.com/v3"

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
		"personalizations": []map[string]any{
			{"to": []map[string]string{{"email": input.To}}},
		},
		"from":    map[string]string{"email": input.From},
		"subject": input.Subject,
		"content": []map[string]string{
			{"type": "text/html", "value": input.HTML},
			{"type": "text/plain", "value": input.Text},
		},
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/mail/send", bytes.NewReader(body))
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
		return email.SendResult{}, fmt.Errorf("sendgrid error %d: %s", resp.StatusCode, raw)
	}

	// SendGrid returns the message ID in the X-Message-Id header, not the body.
	return email.SendResult{
		ProviderMessageID: resp.Header.Get("X-Message-Id"),
		Raw:               raw,
	}, nil
}

// VerifyWebhook verifies the SendGrid event webhook signature using ECDSA.
func (a *Adapter) VerifyWebhook(_ context.Context, headers map[string]string, body []byte) error {
	if a.webhookKey == "" {
		return nil
	}
	sig := headers["x-twilio-email-event-webhook-signature"]
	ts := headers["x-twilio-email-event-webhook-timestamp"]
	if sig == "" || ts == "" {
		return fmt.Errorf("missing sendgrid webhook headers")
	}

	block, _ := pem.Decode([]byte(a.webhookKey))
	if block == nil {
		return fmt.Errorf("invalid webhook public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse public key: %w", err)
	}
	ecKey, ok := pub.(*ecdsa.PublicKey)
	if !ok {
		return fmt.Errorf("webhook key is not ECDSA")
	}

	sigBytes, err := base64.StdEncoding.DecodeString(sig)
	if err != nil || len(sigBytes) < 64 {
		return fmt.Errorf("decode signature: %w", err)
	}
	r := new(big.Int).SetBytes(sigBytes[:32])
	s := new(big.Int).SetBytes(sigBytes[32:])

	payload := append([]byte(ts), body...)
	hash := hashPayload(payload)
	if !ecdsa.Verify(ecKey, hash, r, s) {
		return fmt.Errorf("invalid webhook signature")
	}
	return nil
}

func (a *Adapter) ParseWebhook(_ context.Context, _ map[string]string, body []byte) ([]email.WebhookEvent, error) {
	var events []struct {
		Event     string  `json:"event"`
		Timestamp float64 `json:"timestamp"`
		MessageID string  `json:"sg_message_id"`
		EventID   string  `json:"sg_event_id"`
	}
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("decode webhook: %w", err)
	}

	out := make([]email.WebhookEvent, 0, len(events))
	for _, e := range events {
		out = append(out, email.WebhookEvent{
			EventID:           e.EventID,
			ProviderMessageID: e.MessageID,
			EventType:         normalizeEventType(e.Event),
			OccurredAt:        time.Unix(int64(e.Timestamp), 0),
			Raw:               body,
		})
	}
	return out, nil
}

func normalizeEventType(t string) string {
	switch t {
	case "delivered":
		return "delivered"
	case "bounce", "blocked":
		return "bounced"
	case "spamreport":
		return "complained"
	case "open":
		return "opened"
	case "click":
		return "clicked"
	default:
		return t
	}
}
