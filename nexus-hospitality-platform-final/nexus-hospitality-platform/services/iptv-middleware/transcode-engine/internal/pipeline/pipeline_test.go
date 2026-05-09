package pipeline

import (
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("expected non-nil manager")
	}
}

func TestCreateAndGet(t *testing.T) {
	m := NewManager()
	p, err := m.Create("p1", "stream-1", "tenant-a")
	if err != nil {
		t.Fatalf("Create error = %v", err)
	}
	if p.ID != "p1" {
		t.Errorf("ID = %q, want %q", p.ID, "p1")
	}

	got, err := m.Get("p1")
	if err != nil {
		t.Fatalf("Get error = %v", err)
	}
	if got.StreamID != "stream-1" {
		t.Errorf("StreamID = %q, want %q", got.StreamID, "stream-1")
	}
}

func TestCreate_Duplicate(t *testing.T) {
	m := NewManager()
	_, _ = m.Create("p1", "stream-1", "tenant-a")
	_, err := m.Create("p1", "stream-2", "tenant-a")
	if err == nil {
		t.Error("expected error for duplicate pipeline")
	}
}

func TestListByTenant(t *testing.T) {
	m := NewManager()
	_, _ = m.Create("p1", "s1", "tenant-a")
	_, _ = m.Create("p2", "s2", "tenant-b")
	_, _ = m.Create("p3", "s3", "tenant-a")

	aPipes := m.ListByTenant("tenant-a")
	if len(aPipes) != 2 {
		t.Errorf("len = %d, want 2", len(aPipes))
	}
}

func TestUpdateStatus(t *testing.T) {
	m := NewManager()
	_, _ = m.Create("p1", "s1", "tenant-a")
	if err := m.UpdateStatus("p1", "running"); err != nil {
		t.Fatalf("UpdateStatus error = %v", err)
	}
	p, _ := m.Get("p1")
	if p.Status != "running" {
		t.Errorf("Status = %q, want running", p.Status)
	}
}

func TestUpdateStatus_NotFound(t *testing.T) {
	m := NewManager()
	err := m.UpdateStatus("missing", "running")
	if err == nil {
		t.Error("expected error for missing pipeline")
	}
}
