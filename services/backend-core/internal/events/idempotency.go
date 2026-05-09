package events

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// IdempotencyStore tracks processed event IDs to prevent duplicate handling.
type IdempotencyStore interface {
	IsProcessed(ctx context.Context, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, eventID string, ttl time.Duration) error
}

// InMemoryIdempotencyStore is a TTL-backed in-memory store suitable for single-instance consumers.
// For distributed consumers, replace with Redis or a database implementation.
type InMemoryIdempotencyStore struct {
	entries map[string]time.Time
	mu      sync.RWMutex
	ttl     time.Duration
}

// NewInMemoryIdempotencyStore creates an in-memory idempotency store with a default TTL.
func NewInMemoryIdempotencyStore(defaultTTL time.Duration) *InMemoryIdempotencyStore {
	store := &InMemoryIdempotencyStore{
		entries: make(map[string]time.Time),
		ttl:     defaultTTL,
	}
	go store.cleanupLoop()
	return store
}

// IsProcessed returns true if the event ID has been seen within the TTL window.
func (s *InMemoryIdempotencyStore) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	_ = ctx
	s.mu.RLock()
	expires, exists := s.entries[eventID]
	s.mu.RUnlock()
	if !exists {
		return false, nil
	}
	if time.Now().After(expires) {
		// Expired — treat as not processed and remove.
		s.mu.Lock()
		delete(s.entries, eventID)
		s.mu.Unlock()
		return false, nil
	}
	return true, nil
}

// MarkProcessed records an event ID with the given TTL.
func (s *InMemoryIdempotencyStore) MarkProcessed(ctx context.Context, eventID string, ttl time.Duration) error {
	_ = ctx
	if ttl <= 0 {
		ttl = s.ttl
	}
	s.mu.Lock()
	s.entries[eventID] = time.Now().Add(ttl)
	s.mu.Unlock()
	return nil
}

func (s *InMemoryIdempotencyStore) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, expires := range s.entries {
			if now.After(expires) {
				delete(s.entries, id)
			}
		}
		s.mu.Unlock()
	}
}

// ErrDuplicateEvent is returned when an event has already been processed.
var ErrDuplicateEvent = fmt.Errorf("duplicate event")
