package auth

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/notifylayer/notifylayer/internal/db/repo"
)

// Authenticate is HTTP middleware that validates the Bearer API key and
// stores the resolved tenant in the request context.
func Authenticate(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return authenticate(pool, nil)
}

// AuthenticateExcept is like Authenticate but skips the given paths.
func AuthenticateExcept(skipPaths ...string) func(*pgxpool.Pool) func(http.Handler) http.Handler {
	return func(pool *pgxpool.Pool) func(http.Handler) http.Handler {
		skip := make(map[string]struct{}, len(skipPaths))
		for _, p := range skipPaths {
			skip[p] = struct{}{}
		}
		return authenticate(pool, skip)
	}
}

func authenticate(pool *pgxpool.Pool, skip map[string]struct{}) func(http.Handler) http.Handler {
	keys := repo.NewAPIKeyRepo(pool)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, skipped := skip[r.URL.Path]; skipped {
				next.ServeHTTP(w, r)
				return
			}

			raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if raw == "" {
				writeUnauthorized(w, r, "missing_api_key", "Authorization header is required")
				return
			}

			hash := hashKey(raw)
			key, err := keys.GetByHash(r.Context(), hash)
			if err != nil {
				writeUnauthorized(w, r, "invalid_api_key", "API key is invalid or revoked")
				return
			}

			// Fire-and-forget: update last_used_at without blocking the request.
			go func() {
				_ = keys.Touch(r.Context(), key.ID)
			}()

			ctx := withTenant(r.Context(), TenantContext{
				OrganizationID: key.OrganizationID,
				ProjectID:      key.ProjectID,
				EnvironmentID:  key.EnvironmentID,
				APIKeyID:       key.ID,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func hashKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":       code,
			"message":    message,
			"request_id": middleware.GetReqID(r.Context()),
		},
	})
}
