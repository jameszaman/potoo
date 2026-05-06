package auth

import "context"

type contextKey struct{}

// TenantContext carries the resolved tenant for an authenticated request.
type TenantContext struct {
	OrganizationID string
	ProjectID      string
	EnvironmentID  string
	APIKeyID       string
}

func withTenant(ctx context.Context, t TenantContext) context.Context {
	return context.WithValue(ctx, contextKey{}, t)
}

// TenantFromContext returns the tenant and whether the request is authenticated.
func TenantFromContext(ctx context.Context) (TenantContext, bool) {
	t, ok := ctx.Value(contextKey{}).(TenantContext)
	return t, ok
}
