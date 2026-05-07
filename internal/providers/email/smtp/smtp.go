package smtp

import (
	"context"
	"fmt"
	"strconv"

	mail "github.com/wneessen/go-mail"

	"github.com/potoo/potoo/internal/providers/email"
)

type Adapter struct {
	host      string
	port      int
	username  string
	password  string
	fromEmail string
	fromName  string
}

func New(host string, port int, username, password, fromEmail, fromName string) *Adapter {
	return &Adapter{
		host:      host,
		port:      port,
		username:  username,
		password:  password,
		fromEmail: fromEmail,
		fromName:  fromName,
	}
}

// NewFromCreds builds an Adapter from the credential map stored in provider_connections.
// Required keys: host, port, username, password. Optional: from_email, from_name.
func NewFromCreds(creds map[string]string) (*Adapter, error) {
	host := creds["host"]
	if host == "" {
		return nil, fmt.Errorf("smtp: missing required credential: host")
	}
	portStr := creds["port"]
	if portStr == "" {
		return nil, fmt.Errorf("smtp: missing required credential: port")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("smtp: invalid port %q: %w", portStr, err)
	}
	username := creds["username"]
	if username == "" {
		return nil, fmt.Errorf("smtp: missing required credential: username")
	}
	password := creds["password"]
	if password == "" {
		return nil, fmt.Errorf("smtp: missing required credential: password")
	}
	return New(host, port, username, password, creds["from_email"], creds["from_name"]), nil
}

func (a *Adapter) Send(ctx context.Context, input email.SendInput) (email.SendResult, error) {
	m := mail.NewMsg()

	if a.fromName != "" {
		if err := m.FromFormat(a.fromName, a.fromEmail); err != nil {
			return email.SendResult{}, fmt.Errorf("smtp: set from: %w", err)
		}
	} else {
		if err := m.From(a.fromEmail); err != nil {
			return email.SendResult{}, fmt.Errorf("smtp: set from: %w", err)
		}
	}

	if err := m.To(input.To); err != nil {
		return email.SendResult{}, fmt.Errorf("smtp: set to: %w", err)
	}
	m.Subject(input.Subject)

	// multipart/alternative: clients pick the last part they support.
	// Plain text must come first, HTML last, so HTML is preferred.
	if input.HTML != "" && input.Text != "" {
		m.SetBodyString(mail.TypeTextPlain, input.Text)
		m.AddAlternativeString(mail.TypeTextHTML, input.HTML)
	} else if input.HTML != "" {
		m.SetBodyString(mail.TypeTextHTML, input.HTML)
	} else if input.Text != "" {
		m.SetBodyString(mail.TypeTextPlain, input.Text)
	}

	// Determine TLS policy based on port:
	// 465 = implicit TLS (TLSMandatory), 587 = STARTTLS (TLSOpportunistic)
	tlsPolicy := mail.TLSOpportunistic
	if a.port == 465 {
		tlsPolicy = mail.TLSMandatory
	}

	client, err := mail.NewClient(a.host,
		mail.WithPort(a.port),
		mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover),
		mail.WithUsername(a.username),
		mail.WithPassword(a.password),
		mail.WithTLSPortPolicy(tlsPolicy),
	)
	if err != nil {
		return email.SendResult{}, fmt.Errorf("smtp: create client: %w", err)
	}

	if err := client.DialAndSendWithContext(ctx, m); err != nil {
		return email.SendResult{}, fmt.Errorf("smtp: send: %w", err)
	}

	// SMTP provides no message ID — delivery stays at "sent" with no further events.
	return email.SendResult{ProviderMessageID: ""}, nil
}

// VerifyWebhook is a no-op — SMTP has no webhook delivery receipts.
func (a *Adapter) VerifyWebhook(_ context.Context, _ map[string]string, _ []byte) error {
	return nil
}

// ParseWebhook is a no-op — SMTP has no webhook delivery receipts.
func (a *Adapter) ParseWebhook(_ context.Context, _ map[string]string, _ []byte) ([]email.WebhookEvent, error) {
	return nil, nil
}
