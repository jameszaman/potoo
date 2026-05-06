package handlers

import (
	"context"

	api "github.com/notifylayer/notifylayer/internal/gen/openapi"
)

type Handlers struct{}

func New() *Handlers {
	return &Handlers{}
}

func (h *Handlers) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return api.GetHealth200JSONResponse{Status: api.Ok}, nil
}

func (h *Handlers) SendNotification(_ context.Context, _ api.SendNotificationRequestObject) (api.SendNotificationResponseObject, error) {
	return api.SendNotification202JSONResponse{}, nil
}

func (h *Handlers) GetNotification(_ context.Context, _ api.GetNotificationRequestObject) (api.GetNotificationResponseObject, error) {
	return api.GetNotification404JSONResponse(notImplError()), nil
}

func (h *Handlers) GetDelivery(_ context.Context, _ api.GetDeliveryRequestObject) (api.GetDeliveryResponseObject, error) {
	return api.GetDelivery404JSONResponse(notImplError()), nil
}

func (h *Handlers) GetDeliveryEvents(_ context.Context, _ api.GetDeliveryEventsRequestObject) (api.GetDeliveryEventsResponseObject, error) {
	return api.GetDeliveryEvents404JSONResponse(notImplError()), nil
}

func notImplError() api.Error {
	return api.Error{
		Error: struct {
			Code      string  `json:"code"`
			DocsUrl   *string `json:"docs_url,omitempty"`
			Message   string  `json:"message"`
			RequestId string  `json:"request_id"`
		}{
			Code:    "not_implemented",
			Message: "not implemented",
		},
	}
}
