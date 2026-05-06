package handlers

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/argon2"

	openapi_types "github.com/oapi-codegen/runtime/types"

	api "github.com/notifylayer/notifylayer/internal/gen/openapi"
	"github.com/notifylayer/notifylayer/internal/db/repo"
	"github.com/notifylayer/notifylayer/internal/jwtutil"
)

// --- Register (via strict handler — no cookie needed) ---

func (h *Handlers) Register(ctx context.Context, req api.RegisterRequestObject) (api.RegisterResponseObject, error) {
	body := req.Body
	email := strings.ToLower(string(body.Email))

	if len(body.Password) < 8 {
		return api.Register400JSONResponse(errBody(ctx, "validation_failed", "password must be at least 8 characters")), nil
	}

	hash, err := hashPassword(body.Password)
	if err != nil {
		return nil, err
	}

	user, err := h.users.Create(ctx, email, hash)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			return api.Register409JSONResponse(errBody(ctx, "conflict", "An account with that email already exists")), nil
		}
		return nil, err
	}

	return api.Register201JSONResponse(api.AuthUser{
		Id:        user.ID,
		Email:     openapi_types.Email(email),
		CreatedAt: user.CreatedAt.Time,
	}), nil
}

func (h *Handlers) RegisterHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}

	resp, err := h.Register(ctx, api.RegisterRequestObject{Body: &body})
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	switch v := resp.(type) {
	case api.Register201JSONResponse:
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(v)
	case api.Register400JSONResponse:
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(v)
	case api.Register409JSONResponse:
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(v)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// --- Login (plain HTTP handler — needs to set cookies) ---

// Login satisfies the strict interface — real work is done in LoginHTTP.
func (h *Handlers) Login(_ context.Context, _ api.LoginRequestObject) (api.LoginResponseObject, error) {
	return nil, nil
}

func (h *Handlers) LoginHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body api.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSONError(w, r, http.StatusBadRequest, "invalid_request", "invalid JSON")
		return
	}

	email := strings.ToLower(string(body.Email))

	user, err := h.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSONError(w, r, http.StatusUnauthorized, "auth_failed", "Invalid email or password")
			return
		}
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error")
		return
	}

	if !verifyPassword(body.Password, user.PasswordHash) {
		writeJSONError(w, r, http.StatusUnauthorized, "auth_failed", "Invalid email or password")
		return
	}

	// Resolve the org for this session — pick the first org the user belongs to.
	orgs, err := h.members.ListUserOrgs(ctx, user.ID)
	if err != nil || len(orgs) == 0 {
		writeJSONError(w, r, http.StatusForbidden, "no_org", "User is not a member of any organization")
		return
	}

	accessToken, refreshToken, err := h.createSession(ctx, user.ID, orgs[0].ID)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "could not create session")
		return
	}

	setAuthCookies(w, accessToken, refreshToken, time.Now().Add(jwtutil.RefreshTTL))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(api.AuthUser{
		Id:        user.ID,
		Email:     body.Email,
		CreatedAt: user.CreatedAt.Time,
	})
}

// --- Refresh (plain HTTP handler — rotates refresh token, issues new access token) ---

// RefreshSession satisfies the strict interface — real work is done in RefreshSessionHTTP.
func (h *Handlers) RefreshSession(_ context.Context, _ api.RefreshSessionRequestObject) (api.RefreshSessionResponseObject, error) {
	return nil, nil
}

func (h *Handlers) RefreshSessionHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie(jwtutil.RefreshCookieName)
	if err != nil {
		writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "No refresh token")
		return
	}

	incomingToken := cookie.Value
	hash := repo.HashRefreshToken(incomingToken)

	session, err := h.sessions.GetByRefreshTokenHash(ctx, hash)
	if err != nil {
		// Hash not found — either expired+deleted or a replayed token.
		// We cannot identify which session to nuke without the record, so just reject.
		writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "Invalid or expired refresh token")
		return
	}

	if session.ExpiresAt.Time.Before(time.Now()) {
		_ = h.sessions.Delete(ctx, session.ID)
		clearAuthCookies(w)
		writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "Refresh token expired")
		return
	}

	// Rotate: replace refresh token hash in the same session row.
	newRefreshToken, newHash, err := repo.GenerateRefreshToken()
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "could not rotate session")
		return
	}
	newExpiry := time.Now().Add(jwtutil.RefreshTTL)

	if _, err := h.sessions.Rotate(ctx, session.ID, newHash, newExpiry); err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "could not rotate session")
		return
	}

	accessToken, err := jwtutil.SignAccess(session.UserID, session.ID)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "could not issue access token")
		return
	}

	user, err := h.users.GetByID(ctx, session.UserID)
	if err != nil {
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "user not found")
		return
	}

	setAuthCookies(w, accessToken, newRefreshToken, newExpiry)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(api.AuthUser{
		Id:        user.ID,
		Email:     openapi_types.Email(user.Email),
		CreatedAt: user.CreatedAt.Time,
	})
}

// --- Logout (plain HTTP handler — deletes session, clears cookies) ---

// Logout satisfies the strict interface — real work is done in LogoutHTTP.
func (h *Handlers) Logout(_ context.Context, _ api.LogoutRequestObject) (api.LogoutResponseObject, error) {
	return nil, nil
}

func (h *Handlers) LogoutHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if cookie, err := r.Cookie(jwtutil.RefreshCookieName); err == nil {
		hash := repo.HashRefreshToken(cookie.Value)
		if session, err := h.sessions.GetByRefreshTokenHash(ctx, hash); err == nil {
			_ = h.sessions.Delete(ctx, session.ID)
		}
	}

	clearAuthCookies(w)
	w.WriteHeader(http.StatusNoContent)
}

// --- Me (plain HTTP handler — verifies access token cookie) ---

// GetMe satisfies the strict interface — real work is done in GetMeHTTP.
func (h *Handlers) GetMe(_ context.Context, _ api.GetMeRequestObject) (api.GetMeResponseObject, error) {
	return nil, nil
}

func (h *Handlers) GetMeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	cookie, err := r.Cookie(jwtutil.AccessCookieName)
	if err != nil {
		writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "Not signed in")
		return
	}

	claims, err := jwtutil.Verify(cookie.Value)
	if err != nil {
		writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "Invalid or expired session")
		return
	}

	user, err := h.users.GetByID(ctx, claims.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSONError(w, r, http.StatusUnauthorized, "unauthenticated", "User not found")
			return
		}
		writeJSONError(w, r, http.StatusInternalServerError, "internal_error", "unexpected error")
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

// --- Session helpers ---

func (h *Handlers) createSession(ctx context.Context, userID, orgID string) (accessToken, refreshToken string, err error) {
	refreshToken, hash, err := repo.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}
	session, err := h.sessions.Create(ctx, userID, orgID, hash, time.Now().Add(jwtutil.RefreshTTL))
	if err != nil {
		return "", "", err
	}
	accessToken, err = jwtutil.SignAccess(userID, session.ID)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func setAuthCookies(w http.ResponseWriter, accessToken, refreshToken string, refreshExpiry time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     jwtutil.AccessCookieName,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(jwtutil.AccessTTL.Seconds()),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     jwtutil.RefreshCookieName,
		Value:    refreshToken,
		Path:     "/v1/auth",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(time.Until(refreshExpiry).Seconds()),
	})
}

func clearAuthCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: jwtutil.AccessCookieName, Value: "", Path: "/", HttpOnly: true, MaxAge: -1})
	http.SetCookie(w, &http.Cookie{Name: jwtutil.RefreshCookieName, Value: "", Path: "/v1/auth", HttpOnly: true, MaxAge: -1})
}

// --- Argon2id password hashing ---

const (
	argonTime    = 1
	argonMemory  = 64 * 1024
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

func hashPassword(password string) (string, error) {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	encoded := fmt.Sprintf("$argon2id$s=%s$h=%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
	return encoded, nil
}

func verifyPassword(password, encoded string) bool {
	parts := strings.Split(encoded, "$")
	// format: $argon2id$s=<salt>$h=<hash>  →  ["", "argon2id", "s=...", "h=..."]
	if len(parts) != 4 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(parts[2], "s="))
	if err != nil {
		return false
	}
	expected, err := base64.RawStdEncoding.DecodeString(strings.TrimPrefix(parts[3], "h="))
	if err != nil {
		return false
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return subtle.ConstantTimeCompare(hash, expected) == 1
}

// --- HTTP helpers ---

func writeJSONError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	rid := r.Header.Get("X-Request-Id")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":       code,
			"message":    message,
			"request_id": rid,
		},
	})
}
