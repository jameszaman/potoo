package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/notifylayer/notifylayer/internal/auth"
	db "github.com/notifylayer/notifylayer/internal/db/sqlc"
	api "github.com/notifylayer/notifylayer/internal/gen/openapi"
)

// ListAPIKeys satisfies the strict interface — real work is done in ListAPIKeysHTTP.
func (h *Handlers) ListAPIKeys(_ context.Context, _ api.ListAPIKeysRequestObject) (api.ListAPIKeysResponseObject, error) {
	return nil, nil
}

// CreateAPIKey satisfies the strict interface — real work is done in CreateAPIKeyHTTP.
func (h *Handlers) CreateAPIKey(_ context.Context, _ api.CreateAPIKeyRequestObject) (api.CreateAPIKeyResponseObject, error) {
	return nil, nil
}

// RevokeAPIKey satisfies the strict interface — real work is done in RevokeAPIKeyHTTP.
func (h *Handlers) RevokeAPIKey(_ context.Context, _ api.RevokeAPIKeyRequestObject) (api.RevokeAPIKeyResponseObject, error) {
	return nil, nil
}

// --- Plain HTTP handlers (session-authenticated) ---

func (h *Handlers) ListAPIKeysHTTP(w http.ResponseWriter, r *http.Request) {
	sess, ok := auth.SessionFromContext(r.Context())
	if !ok {
		writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "Sign in required")
		return
	}

	rows, err := h.apiKeys.ListByOrg(r.Context(), sess.OrgID)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}

	out := make([]api.ApiKey, len(rows))
	for i, row := range rows {
		out[i] = listRowToAPIKey(row)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
}

func (h *Handlers) CreateAPIKeyHTTP(w http.ResponseWriter, r *http.Request) {
	sess, ok := auth.SessionFromContext(r.Context())
	if !ok {
		writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "Sign in required")
		return
	}

	var body api.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}

	scopes := []string{}
	if body.Scopes != nil {
		scopes = *body.Scopes
	}

	key, rawKey, err := h.apiKeys.Create(r.Context(), sess.OrgID, body.Name, scopes)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "could not create key")
		return
	}

	out := api.CreateAPIKeyResponse{
		Id:        key.ID,
		Name:      key.Name,
		KeyPrefix: key.KeyPrefix,
		Scopes:    key.Scopes,
		Key:       rawKey,
		CreatedAt: key.CreatedAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(out)
}

func (h *Handlers) RevokeAPIKeyHTTP(w http.ResponseWriter, r *http.Request) {
	sess, ok := auth.SessionFromContext(r.Context())
	if !ok {
		writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "Sign in required")
		return
	}

	keyID := chi.URLParam(r, "keyId")

	if err := h.apiKeys.Revoke(r.Context(), keyID, sess.OrgID); err != nil {
		writeJSONError(w, r, http.StatusNotFound, "not_found", "Key not found or already revoked")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Converters ---

func listRowToAPIKey(row db.ListAPIKeysByOrgRow) api.ApiKey {
	out := api.ApiKey{
		Id:        row.ID,
		Name:      row.Name,
		KeyPrefix: row.KeyPrefix,
		Scopes:    row.Scopes,
		CreatedAt: row.CreatedAt.Time,
	}
	if row.LastUsedAt.Valid {
		t := row.LastUsedAt.Time
		out.LastUsedAt = &t
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		out.RevokedAt = &t
	}
	return out
}

// suppress unused import
var _ = time.Now
