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

func TestRequireTenantID_Success(t *testing.T) {
	ctx := context.Background()
	ctx = WithTenantID(ctx, "tenant-abc")

	tenantID, err := RequireTenantID(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenantID != "tenant-abc" {
		t.Fatalf("expected tenant-abc, got %s", tenantID)
	}
}

func TestRequireTenantID_Missing(t *testing.T) {
	ctx := context.Background()

	tenantID, err := RequireTenantID(ctx)
	if err == nil {
		t.Fatal("expected error for missing tenant_id, got nil")
	}
	if tenantID != "" {
		t.Fatalf("expected empty tenantID on error, got %s", tenantID)
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
