package queue

import (
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"
)

const TypeSendEmail = "email:send"

type SendEmailPayload struct {
	NotificationID string
	DeliveryID     string
	OrganizationID string
	ProjectID      string
	EnvironmentID  string
}

func NewSendEmailTask(p SendEmailPayload) (*asynq.Task, error) {
	body, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}
	return asynq.NewTask(TypeSendEmail, body, asynq.MaxRetry(3)), nil
}

func ParseSendEmailPayload(t *asynq.Task) (SendEmailPayload, error) {
	var p SendEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return p, fmt.Errorf("unmarshal payload: %w", err)
	}
	return p, nil
}
