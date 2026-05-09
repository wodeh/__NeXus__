//go:build !nvml

package gpu

import (
	"context"
	"testing"
	"time"
)

func TestStubManager_New(t *testing.T) {
	m, err := NewManager(0, 0.9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil {
		t.Fatal("expected manager, got nil")
	}
	m.Shutdown()
}

func TestStubManager_AcquireTimeout(t *testing.T) {
	m, err := NewManager(0, 0.9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer m.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = m.Acquire(ctx, 1)
	if err == nil {
		t.Fatal("expected timeout error with zero devices")
	}
}

func TestStubManager_GetStats(t *testing.T) {
	m, err := NewManager(2, 0.9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer m.Shutdown()

	stats := m.GetStats()
	if stats.AvailableCount != 2 {
		t.Fatalf("expected available count 2, got %d", stats.AvailableCount)
	}
}
