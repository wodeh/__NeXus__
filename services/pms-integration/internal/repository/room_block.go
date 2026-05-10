package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
)

// RoomBlockRepository manages room block storage.
type RoomBlockRepository struct {
	mu     sync.RWMutex
	blocks map[string]*domain.RoomBlock
	seq    int64
}

// NewRoomBlockRepository creates a new room block repository.
func NewRoomBlockRepository() *RoomBlockRepository {
	return &RoomBlockRepository{
		blocks: make(map[string]*domain.RoomBlock),
		seq:    1000,
	}
}

// ListByTenant returns all blocks for a tenant, optionally filtered by status.
func (r *RoomBlockRepository) ListByTenant(ctx context.Context, tenant, status string, limit, offset int) ([]*domain.RoomBlock, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*domain.RoomBlock
	for _, b := range r.blocks {
		if b.TenantID != tenant {
			continue
		}
		if status != "" && string(b.Status) != status {
			continue
		}
		results = append(results, b)
	}
	total := len(results)
	if offset >= total {
		return []*domain.RoomBlock{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return results[offset:end], total, nil
}

// GetByID returns a specific block.
func (r *RoomBlockRepository) GetByID(ctx context.Context, tenant, id string) (*domain.RoomBlock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.blocks[id]
	if !ok || b.TenantID != tenant {
		return nil, fmt.Errorf("room block not found")
	}
	return b, nil
}

// Create stores a new block.
func (r *RoomBlockRepository) Create(ctx context.Context, b *domain.RoomBlock) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	b.ID = fmt.Sprintf("rb-%d", r.seq)
	b.CreatedAt = time.Now()
	b.UpdatedAt = b.CreatedAt
	if b.Status == "" {
		b.Status = domain.RoomBlockActive
	}
	r.blocks[b.ID] = b
	return nil
}

// UpdateStatus changes a block's status.
func (r *RoomBlockRepository) UpdateStatus(ctx context.Context, tenant, id string, status domain.RoomBlockStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.blocks[id]
	if !ok || b.TenantID != tenant {
		return fmt.Errorf("room block not found")
	}
	b.Status = status
	b.UpdatedAt = time.Now()
	return nil
}

// Delete removes a block.
func (r *RoomBlockRepository) Delete(ctx context.Context, tenant, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.blocks[id]
	if !ok || b.TenantID != tenant {
		return fmt.Errorf("room block not found")
	}
	delete(r.blocks, id)
	return nil
}

// SeedRoomBlocks pre-populates demo data.
func (r *RoomBlockRepository) SeedRoomBlocks() {
	r.Create(context.Background(), &domain.RoomBlock{
		TenantID:   "demo",
		PropertyID: "p-001",
		RoomIDs:    []string{"r-104"},
		StartDate:  time.Now(),
		EndDate:    time.Now().AddDate(0, 0, 7),
		Reason:     "VIP Hold - Mr. Smith",
		Type:       domain.RoomBlockVIP,
		Status:     domain.RoomBlockActive,
		CreatedBy:  "manager@demo.com",
	})
	r.Create(context.Background(), &domain.RoomBlock{
		TenantID:   "demo",
		PropertyID: "p-001",
		RoomIDs:    []string{"r-302", "r-303", "r-304"},
		StartDate:  time.Now().AddDate(0, 0, 14),
		EndDate:    time.Now().AddDate(0, 0, 16),
		Reason:     "Wedding Party - Johnson Group",
		Type:       domain.RoomBlockGroup,
		Status:     domain.RoomBlockActive,
		CreatedBy:  "sales@demo.com",
	})
}
