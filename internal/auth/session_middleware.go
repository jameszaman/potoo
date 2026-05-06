package auth

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/potoo/potoo/internal/db/repo"
	"github.com/potoo/potoo/internal/jwtutil"
)

// AuthenticateSession validates the pt_access cookie, looks up the session in DB
// to resolve org_id, and stores a SessionContext in the request context.
// org_id is NEVER taken from the JWT — always from the DB session row.
func AuthenticateSession(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	sessions := repo.NewSessionRepo(pool)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(jwtutil.AccessCookieName)
			if err != nil {
				writeUnauthorized(w, r, "unauthenticated", "Sign in required")
				return
			}

			claims, err := jwtutil.Verify(cookie.Value)
			if err != nil {
				writeUnauthorized(w, r, "unauthenticated", "Invalid or expired session")
				return
			}

			if claims.SessionID == "" {
				writeUnauthorized(w, r, "unauthenticated", "Invalid session token")
				return
			}

			session, err := sessions.GetByID(r.Context(), claims.SessionID)
			if err != nil {
				writeUnauthorized(w, r, "unauthenticated", "Session not found")
				return
			}

			ctx := WithSession(r.Context(), SessionContext{
				UserID:    session.UserID,
				SessionID: session.ID,
				OrgID:     session.OrgID,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequirePlatformOwner is middleware that asserts the authenticated session belongs
// to the platform org and that the user is an owner. Must be used after AuthenticateSession.
func RequirePlatformOwner(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	orgs := repo.NewOrgRepo(pool)
	members := repo.NewOrgMemberRepo(pool)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sess, ok := SessionFromContext(r.Context())
			if !ok {
				writeUnauthorized(w, r, "unauthenticated", "Sign in required")
				return
			}

			org, err := orgs.GetByID(r.Context(), sess.OrgID)
			if err != nil || string(org.Type) != "platform" {
				writeForbidden(w, r, "forbidden", "Platform access required")
				return
			}

			member, err := members.Get(r.Context(), sess.OrgID, sess.UserID)
			if err != nil || string(member.Role) != "owner" {
				writeForbidden(w, r, "forbidden", "Owner role required")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
