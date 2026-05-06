package auth

import "context"

type tenantKey struct{}
type sessionKey struct{}

// TenantContext carries the resolved tenant for an API-key-authenticated request.
type TenantContext struct {
	OrganizationID string
	ProjectID      string
	EnvironmentID  string
	APIKeyID       string
}

// SessionContext carries the resolved session for a cookie-authenticated dashboard request.
type SessionContext struct {
	UserID    string
	SessionID string
	OrgID     string
}

func withTenant(ctx context.Context, t TenantContext) context.Context {
	return context.WithValue(ctx, tenantKey{}, t)
}

// TenantFromContext returns the tenant and whether the request is API-key authenticated.
func TenantFromContext(ctx context.Context) (TenantContext, bool) {
	t, ok := ctx.Value(tenantKey{}).(TenantContext)
	return t, ok
}

func WithSession(ctx context.Context, s SessionContext) context.Context {
	return context.WithValue(ctx, sessionKey{}, s)
}

// SessionFromContext returns the session and whether the request is session-authenticated.
func SessionFromContext(ctx context.Context) (SessionContext, bool) {
	s, ok := ctx.Value(sessionKey{}).(SessionContext)
	return s, ok
}
