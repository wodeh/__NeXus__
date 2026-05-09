package domain

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// IdempotencyKey is a strongly typed key for deduplication.
type IdempotencyKey string

// IdempotencyStore tracks processed keys to prevent duplicate side effects.
type IdempotencyStore interface {
	IsProcessed(ctx context.Context, key IdempotencyKey) (bool, error)
	MarkProcessed(ctx context.Context, key IdempotencyKey, ttl time.Duration) error
}

// InMemoryIdempotencyStore is a TTL-backed in-memory store.
type InMemoryIdempotencyStore struct {
	entries map[string]time.Time
	mu      sync.RWMutex
	ttl     time.Duration
}

// NewInMemoryIdempotencyStore creates an in-memory store with the given default TTL.
func NewInMemoryIdempotencyStore(defaultTTL time.Duration) *InMemoryIdempotencyStore {
	store := &InMemoryIdempotencyStore{
		entries: make(map[string]time.Time),
		ttl:     defaultTTL,
	}
	go store.cleanupLoop()
	return store
}

// IsProcessed returns true if the key has been seen within the TTL window.
func (s *InMemoryIdempotencyStore) IsProcessed(ctx context.Context, key IdempotencyKey) (bool, error) {
	_ = ctx
	s.mu.RLock()
	expires, exists := s.entries[string(key)]
	s.mu.RUnlock()
	if !exists {
		return false, nil
	}
	if time.Now().After(expires) {
		s.mu.Lock()
		delete(s.entries, string(key))
		s.mu.Unlock()
		return false, nil
	}
	return true, nil
}

// MarkProcessed records a key with the given TTL.
func (s *InMemoryIdempotencyStore) MarkProcessed(ctx context.Context, key IdempotencyKey, ttl time.Duration) error {
	_ = ctx
	if ttl <= 0 {
		ttl = s.ttl
	}
	s.mu.Lock()
	s.entries[string(key)] = time.Now().Add(ttl)
	s.mu.Unlock()
	return nil
}

func (s *InMemoryIdempotencyStore) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for k, expires := range s.entries {
			if now.After(expires) {
				delete(s.entries, k)
			}
		}
		s.mu.Unlock()
	}
}

// ErrDuplicateIdempotencyKey is returned when a key has already been processed.
var ErrDuplicateIdempotencyKey = fmt.Errorf("duplicate idempotency key")
