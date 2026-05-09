// Package store provides persistence implementations for saga state.
package store

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexus-platform/pms-integration/internal/saga"
)

// InMemorySagaStore persists saga instances in memory with tenant isolation.
type InMemorySagaStore struct {
	instances map[string]*saga.SagaInstance
	mu        sync.RWMutex
}

// NewInMemorySagaStore creates an in-memory saga store.
func NewInMemorySagaStore() *InMemorySagaStore {
	return &InMemorySagaStore{
		instances: make(map[string]*saga.SagaInstance),
	}
}

// Save persists a saga instance.
func (s *InMemorySagaStore) Save(ctx context.Context, instance *saga.SagaInstance) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if instance == nil {
		return fmt.Errorf("cannot save nil saga instance")
	}
	instance.UpdatedAt = time.Now()
	s.instances[instance.ID] = instance
	return nil
}

// Get retrieves a saga instance by ID.
func (s *InMemorySagaStore) Get(ctx context.Context, sagaID string) (*saga.SagaInstance, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inst, ok := s.instances[sagaID]
	if !ok {
		return nil, fmt.Errorf("saga %s not found", sagaID)
	}
	return inst, nil
}

// ListByTenant returns all saga instances for a tenant.
func (s *InMemorySagaStore) ListByTenant(ctx context.Context, tenantID string) []*saga.SagaInstance {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*saga.SagaInstance
	for _, inst := range s.instances {
		if inst.TenantID == tenantID {
			out = append(out, inst)
		}
	}
	return out
}
