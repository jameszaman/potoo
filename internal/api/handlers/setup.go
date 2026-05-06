package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/potoo/potoo/internal/auth"
	db "github.com/potoo/potoo/internal/db/sqlc"
	api "github.com/potoo/potoo/internal/gen/openapi"
	"github.com/potoo/potoo/internal/jwtutil"
)

// GetPlatformStatusHTTP returns whether the platform has been bootstrapped.
func (h *Handlers) GetPlatformStatusHTTP(w http.ResponseWriter, r *http.Request) {
	exists, err := h.orgs.PlatformOrgExists(r.Context())
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"is_configured": exists})
}

// GetPlatformStatus satisfies the strict interface — real work is done in GetPlatformStatusHTTP.
func (h *Handlers) GetPlatformStatus(_ context.Context, _ api.GetPlatformStatusRequestObject) (api.GetPlatformStatusResponseObject, error) {
	return nil, nil
}

// SetupHTTP is the plain HTTP handler for one-time platform bootstrap.
func (h *Handlers) SetupHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}

	exists, err := h.orgs.PlatformOrgExists(ctx)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}
	if exists {
		writeJSONError(w, r, http.StatusGone, "already_configured", "Platform is already configured")
		return
	}

	if body.FirstName == "" || body.LastName == "" {
		writeJSONError(w, r, http.StatusBadRequest, "validation_failed", "first_name and last_name are required")
		return
	}
	if len(body.Password) < 8 {
		writeJSONError(w, r, http.StatusBadRequest, "validation_failed", "password must be at least 8 characters")
		return
	}
	if body.OrgDescription != nil && len(*body.OrgDescription) > 200 {
		writeJSONError(w, r, http.StatusBadRequest, "validation_failed", "org_description must be 200 characters or fewer")
		return
	}

	email := strings.ToLower(string(body.Email))
	hash, err := hashPassword(body.Password)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}

	phone := ""
	if body.Phone != nil {
		phone = *body.Phone
	}

	user, err := h.users.Create(ctx, email, hash, body.FirstName, body.LastName, phone)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			writeJSONError(w, r, http.StatusConflict, "conflict", "An account with that email already exists")
			return
		}
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}

	slug := strings.ToLower(strings.ReplaceAll(body.OrgName, " ", "-"))
	org, err := h.orgs.CreateWithDefaults(ctx, body.OrgName, slug, db.OrgTypePlatform, body.OrgWebsite, body.OrgDescription)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "could not create org")
		return
	}

	if _, err := h.members.Create(ctx, org.ID, user.ID, db.OrgRoleOwner); err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "could not create membership")
		return
	}

	accessToken, refreshToken, err := h.createSession(ctx, user.ID, org.ID)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "could not create session")
		return
	}

	setAuthCookies(w, accessToken, refreshToken, time.Now().Add(jwtutil.RefreshTTL))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(api.AuthUser{
		Id:        user.ID,
		Email:     openapi_types.Email(email),
		CreatedAt: user.CreatedAt.Time,
	})
}

// Setup satisfies the strict interface — real work is done in SetupHTTP.
func (h *Handlers) Setup(_ context.Context, _ api.SetupRequestObject) (api.SetupResponseObject, error) {
	return nil, nil
}

// UpdateOrg satisfies the strict interface — real work is done in UpdateOrgHTTP.
func (h *Handlers) UpdateOrg(_ context.Context, _ api.UpdateOrgRequestObject) (api.UpdateOrgResponseObject, error) {
	return nil, nil
}

// --- Org management (platform owner only, enforced by middleware) ---

func (h *Handlers) ListOrgs(ctx context.Context, _ api.ListOrgsRequestObject) (api.ListOrgsResponseObject, error) {
	rows, err := h.orgs.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]api.OrgSummary, len(rows))
	for i, r := range rows {
		out[i] = dbOrgToAPI(r, db.OrgRoleOwner)
	}
	return api.ListOrgs200JSONResponse{Data: out}, nil
}

func (h *Handlers) CreateOrg(ctx context.Context, req api.CreateOrgRequestObject) (api.CreateOrgResponseObject, error) {
	org, err := h.orgs.CreateWithDefaults(ctx, req.Body.Name, req.Body.Slug, db.OrgTypeCustomer, nil, nil)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return api.CreateOrg409JSONResponse(errBody(ctx, "conflict", "An organization with that slug already exists")), nil
		}
		return nil, err
	}
	return api.CreateOrg201JSONResponse(dbOrgToAPI(*org, db.OrgRoleOwner)), nil
}

// --- Per-user org actions ---

func (h *Handlers) ListMyOrgs(ctx context.Context, _ api.ListMyOrgsRequestObject) (api.ListMyOrgsResponseObject, error) {
	sess, ok := auth.SessionFromContext(ctx)
	if !ok {
		return api.ListMyOrgs401JSONResponse(errBody(ctx, "unauthenticated", "Sign in required")), nil
	}

	orgs, err := h.members.ListUserOrgs(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}

	out := make([]api.OrgSummary, 0, len(orgs))
	for _, org := range orgs {
		member, err := h.members.Get(ctx, org.ID, sess.UserID)
		if err != nil {
			continue
		}
		out = append(out, dbOrgToAPI(org, member.Role))
	}
	return api.ListMyOrgs200JSONResponse{Data: out}, nil
}

// SelectOrg satisfies the strict interface — real work is done in SelectOrgHTTP.
func (h *Handlers) SelectOrg(_ context.Context, _ api.SelectOrgRequestObject) (api.SelectOrgResponseObject, error) {
	return nil, nil
}

func (h *Handlers) SelectOrgHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie(jwtutil.AccessCookieName)
	if err != nil {
		writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "Sign in required")
		return
	}
	claims, err := jwtutil.Verify(cookie.Value)
	if err != nil {
		writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "Invalid or expired session")
		return
	}

	var body api.SelectOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}

	// Verify the user is a member of the requested org.
	member, err := h.members.Get(ctx, body.OrgId, claims.UserID)
	if err != nil {
		writeJSONError(w, r, http.StatusForbidden, "forbidden", "Not a member of this organization")
		return
	}

	// Delete the old session.
	if claims.SessionID != "" {
		_ = h.sessions.Delete(ctx, claims.SessionID)
	}

	accessToken, refreshToken, err := h.createSession(ctx, claims.UserID, member.OrgID)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "could not create session")
		return
	}

	setAuthCookies(w, accessToken, refreshToken, time.Now().Add(jwtutil.RefreshTTL))

	user, err := h.users.GetByID(ctx, claims.UserID)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "user not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(api.AuthUser{
		Id:        user.ID,
		Email:     openapi_types.Email(user.Email),
		CreatedAt: user.CreatedAt.Time,
	})
}

// --- Plain HTTP wrappers for session-authenticated routes ---

func (h *Handlers) ListMyOrgsHTTP(w http.ResponseWriter, r *http.Request) {
	sess, ok := auth.SessionFromContext(r.Context())
	if !ok {
		writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "Sign in required")
		return
	}

	orgs, err := h.members.ListUserOrgs(r.Context(), sess.UserID)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}

	out := make([]api.OrgSummary, 0, len(orgs))
	for _, org := range orgs {
		member, err := h.members.Get(r.Context(), org.ID, sess.UserID)
		if err != nil {
			continue
		}
		out = append(out, dbOrgToAPI(org, member.Role))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
}

func (h *Handlers) ListOrgsHTTP(w http.ResponseWriter, r *http.Request) {
	rows, err := h.orgs.ListAll(r.Context())
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}
	out := make([]api.OrgSummary, 0, len(rows))
	for _, row := range rows {
		summary := dbOrgToAPI(row, db.OrgRoleOwner)
		if stats, err := h.orgs.GetStats(r.Context(), row.ID); err == nil {
			keyCount := int(stats.ApiKeyCount)
			callCount := int(stats.CallCount)
			summary.ApiKeyCount = &keyCount
			summary.CallCount = &callCount
		}
		out = append(out, summary)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"data": out})
}

func (h *Handlers) UpdateOrgHTTP(w http.ResponseWriter, r *http.Request) {
	orgID := chi.URLParam(r, "orgId")

	var body api.UpdateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}

	if body.IsActive == nil {
		writeJSONError(w, r, http.StatusBadRequest, "validation_failed", "no fields to update")
		return
	}

	org, err := h.orgs.SetActive(r.Context(), orgID, *body.IsActive)
	if err != nil {
		writeJSONError(w, r, http.StatusNotFound, "not_found", "Organization not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dbOrgToAPI(*org, db.OrgRoleOwner))
}

func (h *Handlers) CreateOrgHTTP(w http.ResponseWriter, r *http.Request) {
	var body api.CreateOrgRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}

	org, err := h.orgs.CreateWithDefaults(r.Context(), body.Name, body.Slug, db.OrgTypeCustomer, nil, nil)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			writeJSONError(w, r, http.StatusConflict, "conflict", "An organization with that slug already exists")
			return
		}
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(dbOrgToAPI(*org, db.OrgRoleOwner))
}

// --- Helpers ---

func dbOrgToAPI(org db.Organization, role db.OrgRole) api.OrgSummary {
	return api.OrgSummary{
		Id:       org.ID,
		Name:     org.Name,
		Slug:     org.Slug,
		Type:     api.OrgSummaryType(org.Type),
		Role:     api.OrgSummaryRole(role),
		IsActive: org.IsActive,
	}
}
