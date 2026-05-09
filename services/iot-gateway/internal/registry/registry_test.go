package registry

import (
	"context"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	r, err := New(nil)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if r == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestRegisterAndGet(t *testing.T) {
	r, _ := New(nil)
	dev := &Device{
		ID:       "dev-1",
		TenantID: "tenant-a",
		RoomID:   "101",
		Protocol: "zigbee",
	}
	if err := r.Register(context.Background(), dev); err != nil {
		t.Fatalf("Register error = %v", err)
	}
	got, err := r.Get("dev-1")
	if err != nil {
		t.Fatalf("Get error = %v", err)
	}
	if got.ID != "dev-1" {
		t.Errorf("ID = %q, want %q", got.ID, "dev-1")
	}
}

func TestRegister_Duplicate(t *testing.T) {
	r, _ := New(nil)
	dev := &Device{ID: "dev-1", TenantID: "tenant-a"}
	_ = r.Register(context.Background(), dev)
	err := r.Register(context.Background(), dev)
	if err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestUpdateStatus(t *testing.T) {
	r, _ := New(nil)
	dev := &Device{ID: "dev-1", TenantID: "tenant-a"}
	_ = r.Register(context.Background(), dev)
	if err := r.UpdateStatus("dev-1", "offline"); err != nil {
		t.Fatalf("UpdateStatus error = %v", err)
	}
	got, _ := r.Get("dev-1")
	if got.Status != "offline" {
		t.Errorf("Status = %q, want %q", got.Status, "offline")
	}
}

func TestUpdateState(t *testing.T) {
	r, _ := New(nil)
	dev := &Device{ID: "dev-1", TenantID: "tenant-a"}
	_ = r.Register(context.Background(), dev)
	state := map[string]interface{}{"temperature": 22.5}
	if err := r.UpdateState("dev-1", state); err != nil {
		t.Fatalf("UpdateState error = %v", err)
	}
	got, _ := r.Get("dev-1")
	if got.State["temperature"] != 22.5 {
		t.Errorf("State = %v, want temperature 22.5", got.State)
	}
}

func TestListByTenant_Isolation(t *testing.T) {
	r, _ := New(nil)
	_ = r.Register(context.Background(), &Device{ID: "d1", TenantID: "tenant-a"})
	_ = r.Register(context.Background(), &Device{ID: "d2", TenantID: "tenant-b"})
	_ = r.Register(context.Background(), &Device{ID: "d3", TenantID: "tenant-a"})

	aDevices := r.ListByTenant("tenant-a")
	if len(aDevices) != 2 {
		t.Errorf("len(tenant-a) = %d, want 2", len(aDevices))
	}

	bDevices := r.ListByTenant("tenant-b")
	if len(bDevices) != 1 {
		t.Errorf("len(tenant-b) = %d, want 1", len(bDevices))
	}
}

func TestGet_NotFound(t *testing.T) {
	r, _ := New(nil)
	_, err := r.Get("missing")
	if err == nil {
		t.Error("expected error for missing device")
	}
}

func TestRegister_SetsTimestamps(t *testing.T) {
	r, _ := New(nil)
	dev := &Device{ID: "dev-1", TenantID: "tenant-a"}
	_ = r.Register(context.Background(), dev)
	if dev.RegisteredAt.IsZero() {
		t.Error("RegisteredAt not set")
	}
	if dev.Status != "online" {
		t.Errorf("Status = %q, want online", dev.Status)
	}
}
