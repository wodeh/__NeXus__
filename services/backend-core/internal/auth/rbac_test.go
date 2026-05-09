package auth

import (
	"testing"
)

func TestRBACEngine_HasPermission(t *testing.T) {
	engine := NewRBACEngine()

	if !engine.HasPermission("admin", "read:tenant") {
		t.Fatal("admin should have wildcard permission")
	}
	if !engine.HasPermission("manager", "read:tenant") {
		t.Fatal("manager should have read:tenant")
	}
	if engine.HasPermission("manager", "read:analytics") {
		t.Fatal("manager should not have read:analytics")
	}
	if engine.HasPermission("unknown", "read:tenant") {
		t.Fatal("unknown role should have no permissions")
	}
}

func TestRBACEngine_Enforce(t *testing.T) {
	engine := NewRBACEngine()
	ctx := WithTenantID(nil, "tenant-1")

	if err := engine.Enforce(ctx, "guest", "read:own_reservation"); err != nil {
		t.Fatalf("guest should be allowed: %v", err)
	}
	if err := engine.Enforce(ctx, "guest", "write:reservation"); err == nil {
		t.Fatal("guest should not be allowed write:reservation")
	}
}
