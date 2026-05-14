package auth

<<<<<<< HEAD
import "context"
=======
import (
	"context"
	"fmt"
)
>>>>>>> phase1/security-stability

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	tenantIDKey contextKey = "tenant_id"
)

// WithUserID attaches a user ID to the context.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// UserIDFromContext extracts the user ID from the context.
func UserIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(userIDKey).(string)
	return v, ok
}

// WithTenantID attaches a tenant ID to the context.
func WithTenantID(ctx context.Context, tenantID string) context.Context {
	return context.WithValue(ctx, tenantIDKey, tenantID)
}

// TenantIDFromContext extracts the tenant ID from the context.
func TenantIDFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(tenantIDKey).(string)
	return v, ok
}

<<<<<<< HEAD
// RequireTenantID returns the tenant ID or panics if missing.
// Use only in handlers where middleware guarantees presence.
func RequireTenantID(ctx context.Context) string {
	v, ok := TenantIDFromContext(ctx)
	if !ok {
		panic("tenant_id missing from context")
	}
	return v
=======
// RequireTenantID returns the tenant ID or an error if missing.
// Use in handlers where middleware guarantees presence; callers must handle the error.
func RequireTenantID(ctx context.Context) (string, error) {
	v, ok := TenantIDFromContext(ctx)
	if !ok {
		return "", fmt.Errorf("tenant_id missing from context")
	}
	return v, nil
>>>>>>> phase1/security-stability
}
