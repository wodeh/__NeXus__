package auth

import (
	"context"
	"testing"
)

func TestTenantContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithTenantID(ctx, "tenant-abc")

	tenantID, ok := TenantIDFromContext(ctx)
	if !ok || tenantID != "tenant-abc" {
		t.Fatalf("expected tenant-abc, got %s", tenantID)
	}
}

func TestUserContext(t *testing.T) {
	ctx := context.Background()
	ctx = WithUserID(ctx, "user-123")

	userID, ok := UserIDFromContext(ctx)
	if !ok || userID != "user-123" {
		t.Fatalf("expected user-123, got %s", userID)
	}
}
