package handlers

import (
	"context"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/notifylayer/notifylayer/internal/auth"
	api "github.com/notifylayer/notifylayer/internal/gen/openapi"
)

type Handlers struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Handlers {
	return &Handlers{pool: pool}
}

func (h *Handlers) GetHealth(_ context.Context, _ api.GetHealthRequestObject) (api.GetHealthResponseObject, error) {
	return api.GetHealth200JSONResponse{Status: api.Ok}, nil
}

func (h *Handlers) SendNotification(ctx context.Context, _ api.SendNotificationRequestObject) (api.SendNotificationResponseObject, error) {
	if _, ok := auth.TenantFromContext(ctx); !ok {
		return api.SendNotification400JSONResponse(authError(ctx)), nil
	}
	return api.SendNotification202JSONResponse{}, nil
}

func (h *Handlers) GetNotification(ctx context.Context, _ api.GetNotificationRequestObject) (api.GetNotificationResponseObject, error) {
	if _, ok := auth.TenantFromContext(ctx); !ok {
		return api.GetNotification404JSONResponse(notFoundError(ctx)), nil
	}
	return api.GetNotification404JSONResponse(notImplError(ctx)), nil
}

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

func authError(ctx context.Context) api.Error {
	return errBody(ctx, "auth_invalid_api_key", "Invalid or missing API key")
}

func notFoundError(ctx context.Context) api.Error {
	return errBody(ctx, "not_found", "Resource not found")
}

func notImplError(ctx context.Context) api.Error {
	return errBody(ctx, "not_implemented", "Not implemented")
}

// Ensure Handlers satisfies the generated interface at compile time.
var _ interface {
	GetHealth(context.Context, api.GetHealthRequestObject) (api.GetHealthResponseObject, error)
} = (*Handlers)(nil)
