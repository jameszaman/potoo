package auth

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/potoo/potoo/internal/db/repo"
	"github.com/potoo/potoo/internal/jwtutil"
)

// Authenticate validates a Bearer API key, resolves the tenant's default project
// and production environment, and stores TenantContext in the request context.
func Authenticate(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	keys := repo.NewAPIKeyRepo(pool)
	orgs := repo.NewOrgRepo(pool)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

			// Check the org is active before doing anything else.
			org, err := orgs.GetByID(r.Context(), key.OrganizationID)
			if err != nil {
				writeErrorResponse(w, r, http.StatusUnauthorized, "invalid_api_key", "API key is invalid or revoked")
				return
			}
			if !org.IsActive {
				writeErrorResponse(w, r, http.StatusForbidden, "org_inactive", "This organization has been deactivated")
				return
			}

			// Resolve default project and production environment for this org.
			project, err := orgs.GetDefaultProject(r.Context(), key.OrganizationID)
			if err != nil {
				writeErrorResponse(w, r, http.StatusInternalServerError, "missing_default_project",
					"Organization has no default project — contact support")
				return
			}
			env, err := orgs.GetProductionEnvironment(r.Context(), project.ID)
			if err != nil {
				writeErrorResponse(w, r, http.StatusInternalServerError, "missing_production_env",
					"Organization has no production environment — contact support")
				return
			}

			// Fire-and-forget: update last_used_at without blocking the request.
			go func() {
				_ = keys.Touch(r.Context(), key.ID)
			}()

			ctx := withTenant(r.Context(), TenantContext{
				OrganizationID: key.OrganizationID,
				ProjectID:      project.ID,
				EnvironmentID:  env.ID,
				APIKeyID:       key.ID,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthenticateEither accepts either a Bearer API key or a session cookie.
// Either way it resolves a TenantContext so all downstream handlers work identically.
// When a session cookie is used it also stores a SessionContext.
func AuthenticateEither(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	apiKeys := repo.NewAPIKeyRepo(pool)
	orgs := repo.NewOrgRepo(pool)
	sessions := repo.NewSessionRepo(pool)

	resolveTenantForOrg := func(w http.ResponseWriter, r *http.Request, orgID string) (TenantContext, bool) {
		org, err := orgs.GetByID(r.Context(), orgID)
		if err != nil {
			writeUnauthorized(w, r, "invalid_auth", "Organization not found")
			return TenantContext{}, false
		}
		if !org.IsActive {
			writeErrorResponse(w, r, http.StatusForbidden, "org_inactive", "This organization has been deactivated")
			return TenantContext{}, false
		}
		project, err := orgs.GetDefaultProject(r.Context(), orgID)
		if err != nil {
			writeErrorResponse(w, r, http.StatusInternalServerError, "missing_default_project", "Organization has no default project")
			return TenantContext{}, false
		}
		env, err := orgs.GetProductionEnvironment(r.Context(), project.ID)
		if err != nil {
			writeErrorResponse(w, r, http.StatusInternalServerError, "missing_production_env", "Organization has no production environment")
			return TenantContext{}, false
		}
		return TenantContext{OrganizationID: orgID, ProjectID: project.ID, EnvironmentID: env.ID}, true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try Bearer API key first.
			raw := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if raw != "" {
				hash := hashKey(raw)
				key, err := apiKeys.GetByHash(r.Context(), hash)
				if err != nil {
					writeUnauthorized(w, r, "invalid_api_key", "API key is invalid or revoked")
					return
				}
				tenant, ok := resolveTenantForOrg(w, r, key.OrganizationID)
				if !ok {
					return
				}
				tenant.APIKeyID = key.ID
				go func() { _ = apiKeys.Touch(r.Context(), key.ID) }()
				next.ServeHTTP(w, r.WithContext(withTenant(r.Context(), tenant)))
				return
			}

			// Fall back to session cookie.
			cookie, err := r.Cookie("pt_access")
			if err != nil {
				writeUnauthorized(w, r, "unauthenticated", "Authorization header or session cookie required")
				return
			}
			claims, err := jwtutil.Verify(cookie.Value)
			if err != nil || claims.SessionID == "" {
				writeUnauthorized(w, r, "unauthenticated", "Invalid or expired session")
				return
			}
			session, err := sessions.GetByID(r.Context(), claims.SessionID)
			if err != nil {
				writeUnauthorized(w, r, "unauthenticated", "Session not found")
				return
			}
			tenant, ok := resolveTenantForOrg(w, r, session.OrgID)
			if !ok {
				return
			}
			ctx := WithSession(r.Context(), SessionContext{
				UserID:    session.UserID,
				SessionID: session.ID,
				OrgID:     session.OrgID,
			})
			ctx = withTenant(ctx, tenant)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func hashKey(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("%x", sum)
}

func writeUnauthorized(w http.ResponseWriter, r *http.Request, code, message string) {
	writeErrorResponse(w, r, http.StatusUnauthorized, code, message)
}

func writeForbidden(w http.ResponseWriter, r *http.Request, code, message string) {
	writeErrorResponse(w, r, http.StatusForbidden, code, message)
}

func writeErrorResponse(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":       code,
			"message":    message,
			"request_id": middleware.GetReqID(r.Context()),
		},
	})
}
