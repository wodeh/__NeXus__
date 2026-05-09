package domain

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryIdempotencyStore(t *testing.T) {
	store := NewInMemoryIdempotencyStore(1 * time.Hour)
	ctx := context.Background()

	seen, err := store.IsProcessed(ctx, IdempotencyKey("key-1"))
	if err != nil {
		t.Fatalf("is processed: %v", err)
	}
	if seen {
		t.Fatal("expected key-1 to be unseen")
	}

	if err := store.MarkProcessed(ctx, IdempotencyKey("key-1"), 0); err != nil {
		t.Fatalf("mark processed: %v", err)
	}

	seen, err = store.IsProcessed(ctx, IdempotencyKey("key-1"))
	if err != nil {
		t.Fatalf("is processed after mark: %v", err)
	}
	if !seen {
		t.Fatal("expected key-1 to be seen")
	}
}
