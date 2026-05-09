package repository

import (
	"context"
	"testing"
)

func TestStore_GetTenantConfig_MissingID(t *testing.T) {
	store := NewStore(nil)
	_, err := store.GetTenantConfig(context.Background(), "")
	if err == nil {
		t.Fatal("expected error when tenant_id is empty")
	}
}

func TestStore_GetTenantConfig_Success(t *testing.T) {
	store := NewStore(nil)
	cfg, err := store.GetTenantConfig(context.Background(), "tenant-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ID != "tenant-123" {
		t.Fatalf("expected ID tenant-123, got %s", cfg.ID)
	}
	if cfg.Tier != "enterprise" {
		t.Fatalf("expected tier enterprise, got %s", cfg.Tier)
	}
}
