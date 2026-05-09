package store

import (
	"context"
	"testing"
	"time"

	"github.com/nexus-platform/pms-integration/internal/saga"
)

func TestNewInMemorySagaStore(t *testing.T) {
	s := NewInMemorySagaStore()
	if s == nil {
		t.Fatal("expected non-nil store")
	}
}

func TestSaveAndGet(t *testing.T) {
	s := NewInMemorySagaStore()
	inst := &saga.SagaInstance{
		ID:       "saga-1",
		Type:     "guest_checkin",
		TenantID: "tenant-a",
		Status:   saga.SagaStatusRunning,
		Steps: []saga.SagaStep{
			{Name: "step-1", Status: saga.StepStatusCompleted},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.Save(context.Background(), inst); err != nil {
		t.Fatalf("Save error = %v", err)
	}

	got, err := s.Get(context.Background(), "saga-1")
	if err != nil {
		t.Fatalf("Get error = %v", err)
	}
	if got.ID != "saga-1" {
		t.Errorf("ID = %q, want %q", got.ID, "saga-1")
	}
}

func TestSave_Nil(t *testing.T) {
	s := NewInMemorySagaStore()
	err := s.Save(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil instance")
	}
}

func TestGet_NotFound(t *testing.T) {
	s := NewInMemorySagaStore()
	_, err := s.Get(context.Background(), "missing")
	if err == nil {
		t.Error("expected error for missing saga")
	}
}

func TestListByTenant(t *testing.T) {
	s := NewInMemorySagaStore()
	_ = s.Save(context.Background(), &saga.SagaInstance{ID: "s1", TenantID: "tenant-a", Status: saga.SagaStatusRunning, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	_ = s.Save(context.Background(), &saga.SagaInstance{ID: "s2", TenantID: "tenant-b", Status: saga.SagaStatusRunning, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	_ = s.Save(context.Background(), &saga.SagaInstance{ID: "s3", TenantID: "tenant-a", Status: saga.SagaStatusRunning, CreatedAt: time.Now(), UpdatedAt: time.Now()})

	list := s.ListByTenant(context.Background(), "tenant-a")
	if len(list) != 2 {
		t.Errorf("len = %d, want 2", len(list))
	}
}
