package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/potoo/potoo/internal/auth"
	db "github.com/potoo/potoo/internal/db/sqlc"
	api "github.com/potoo/potoo/internal/gen/openapi"
)

// CreateInvite satisfies the strict interface — real work is done in CreateInviteHTTP.
func (h *Handlers) CreateInvite(_ context.Context, _ api.CreateInviteRequestObject) (api.CreateInviteResponseObject, error) {
	return nil, nil
}

// ListInvites satisfies the strict interface — real work is done in ListInvitesHTTP.
func (h *Handlers) ListInvites(_ context.Context, _ api.ListInvitesRequestObject) (api.ListInvitesResponseObject, error) {
	return nil, nil
}

// CreateOrgInvite satisfies the strict interface — real work is done in CreateInviteHTTP.
func (h *Handlers) CreateOrgInvite(_ context.Context, _ api.CreateOrgInviteRequestObject) (api.CreateOrgInviteResponseObject, error) {
	return nil, nil
}

// ListOrgInvites satisfies the strict interface — real work is done in ListInvitesHTTP.
func (h *Handlers) ListOrgInvites(_ context.Context, _ api.ListOrgInvitesRequestObject) (api.ListOrgInvitesResponseObject, error) {
	return nil, nil
}

func (h *Handlers) DeleteOrgInvite(ctx context.Context, req api.DeleteOrgInviteRequestObject) (api.DeleteOrgInviteResponseObject, error) {
	sess, ok := auth.SessionFromContext(ctx)
	if !ok {
		return api.DeleteOrgInvite404JSONResponse(notFoundError(ctx)), nil
	}
	invite, err := h.invites.GetByID(ctx, req.InviteId)
	if err != nil || invite.OrgID != sess.OrgID {
		return api.DeleteOrgInvite404JSONResponse(notFoundError(ctx)), nil
	}
	if err := h.invites.Delete(ctx, req.InviteId); err != nil {
		return nil, err
	}
	return api.DeleteOrgInvite204Response{}, nil
}

// GetInvite satisfies the strict interface — real work is done in GetInviteHTTP.
func (h *Handlers) GetInvite(_ context.Context, _ api.GetInviteRequestObject) (api.GetInviteResponseObject, error) {
	return nil, nil
}

// --- Plain HTTP handlers ---

func (h *Handlers) GetInviteHTTP(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	invite, err := h.invites.GetByToken(r.Context(), token)
	if err != nil {
		writeJSONError(w, r, http.StatusNotFound, "not_found", "Invite not found")
		return
	}
	if invite.UsedAt.Valid {
		writeJSONError(w, r, http.StatusGone, "invite_used", "This invite has already been used")
		return
	}
	if invite.ExpiresAt.Time.Before(time.Now()) {
		writeJSONError(w, r, http.StatusGone, "invite_expired", "This invite has expired")
		return
	}

	org, err := h.orgs.GetByID(r.Context(), invite.OrgID)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(api.InvitePreview{
		OrgName:   org.Name,
		ExpiresAt: invite.ExpiresAt.Time,
	})
}

func (h *Handlers) CreateInviteHTTP(w http.ResponseWriter, r *http.Request) {
	sess, ok := auth.SessionFromContext(r.Context())
	if !ok {
		writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "Sign in required")
		return
	}

	org, err := h.orgs.GetByID(r.Context(), sess.OrgID)
	if err != nil {
		writeJSONError(w, r, http.StatusNotFound, "not_found", "Organization not found")
		return
	}

	invite, _, err := h.invites.Create(r.Context(), org.ID, sess.UserID)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "could not create invite")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dbInviteToAPI(*invite, org.Name))
}

func (h *Handlers) ListInvitesHTTP(w http.ResponseWriter, r *http.Request) {
	sess, ok := auth.SessionFromContext(r.Context())
	if !ok {
		writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "Sign in required")
		return
	}

	org, err := h.orgs.GetByID(r.Context(), sess.OrgID)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}

	invites, err := h.invites.ListByOrg(r.Context(), org.ID)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}

	out := make([]api.InviteResponse, len(invites))
	for i, inv := range invites {
		out[i] = dbInviteToAPI(inv, org.Name)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
}

func dbInviteToAPI(inv db.OrgInvite, orgName string) api.InviteResponse {
	out := api.InviteResponse{
		Id:        inv.ID,
		OrgId:     inv.OrgID,
		OrgName:   &orgName,
		Token:     inv.Token,
		ExpiresAt: inv.ExpiresAt.Time,
		CreatedAt: inv.CreatedAt.Time,
	}
	if inv.UsedAt.Valid {
		t := inv.UsedAt.Time
		out.UsedAt = &t
	}
	if inv.DeletedAt.Valid {
		t := inv.DeletedAt.Time
		out.DeletedAt = &t
	}
	return out
}
