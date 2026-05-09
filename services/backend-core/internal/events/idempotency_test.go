package events

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryIdempotencyStore(t *testing.T) {
	store := NewInMemoryIdempotencyStore(1 * time.Hour)
	ctx := context.Background()

	seen, err := store.IsProcessed(ctx, "evt-1")
	if err != nil {
		t.Fatalf("is processed: %v", err)
	}
	if seen {
		t.Fatal("expected evt-1 to be unseen")
	}

	if err := store.MarkProcessed(ctx, "evt-1", 0); err != nil {
		t.Fatalf("mark processed: %v", err)
	}

	seen, err = store.IsProcessed(ctx, "evt-1")
	if err != nil {
		t.Fatalf("is processed after mark: %v", err)
	}
	if !seen {
		t.Fatal("expected evt-1 to be seen")
	}
}

func TestInMemoryIdempotencyStore_Expiry(t *testing.T) {
	store := NewInMemoryIdempotencyStore(1 * time.Millisecond)
	ctx := context.Background()

	if err := store.MarkProcessed(ctx, "evt-2", 1*time.Millisecond); err != nil {
		t.Fatalf("mark processed: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	seen, err := store.IsProcessed(ctx, "evt-2")
	if err != nil {
		t.Fatalf("is processed: %v", err)
	}
	if seen {
		t.Fatal("expected evt-2 to be expired")
	}
}
