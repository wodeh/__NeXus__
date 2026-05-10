package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
)

// GroupReservationRepository manages group bookings.
type GroupReservationRepository struct {
	mu     sync.RWMutex
	groups map[string]*domain.GroupReservation
	seq    int64
}

// NewGroupReservationRepository creates a new group repository.
func NewGroupReservationRepository() *GroupReservationRepository {
	return &GroupReservationRepository{
		groups: make(map[string]*domain.GroupReservation),
		seq:    500,
	}
}

// ListByTenant returns all groups for a tenant, optionally filtered by status.
func (r *GroupReservationRepository) ListByTenant(ctx context.Context, tenant, status string, limit, offset int) ([]*domain.GroupReservation, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*domain.GroupReservation
	for _, g := range r.groups {
		if g.TenantID != tenant {
			continue
		}
		if status != "" && g.Status != status {
			continue
		}
		results = append(results, g)
	}
	total := len(results)
	if offset >= total {
		return []*domain.GroupReservation{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return results[offset:end], total, nil
}

// GetByID returns a specific group.
func (r *GroupReservationRepository) GetByID(ctx context.Context, tenant, id string) (*domain.GroupReservation, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	g, ok := r.groups[id]
	if !ok || g.TenantID != tenant {
		return nil, fmt.Errorf("group reservation not found")
	}
	return g, nil
}

// Create stores a new group.
func (r *GroupReservationRepository) Create(ctx context.Context, g *domain.GroupReservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	g.ID = fmt.Sprintf("grp-%d", r.seq)
	g.CreatedAt = time.Now()
	g.UpdatedAt = g.CreatedAt
	if g.Status == "" {
		g.Status = domain.GroupStatusTentative
	}
	r.groups[g.ID] = g
	return nil
}

// Update modifies an existing group.
func (r *GroupReservationRepository) Update(ctx context.Context, tenant, id string, updates *domain.GroupReservation) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	g, ok := r.groups[id]
	if !ok || g.TenantID != tenant {
		return fmt.Errorf("group reservation not found")
	}
	if updates.GroupName != "" {
		g.GroupName = updates.GroupName
	}
	if updates.Status != "" {
		g.Status = updates.Status
	}
	if updates.NumRooms > 0 {
		g.NumRooms = updates.NumRooms
	}
	g.UpdatedAt = time.Now()
	return nil
}

// Delete removes a group.
func (r *GroupReservationRepository) Delete(ctx context.Context, tenant, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	g, ok := r.groups[id]
	if !ok || g.TenantID != tenant {
		return fmt.Errorf("group reservation not found")
	}
	delete(r.groups, id)
	return nil
}

// SeedGroupReservations pre-populates demo data.
func (r *GroupReservationRepository) SeedGroupReservations() {
	r.Create(context.Background(), &domain.GroupReservation{
		TenantID:     "demo",
		PropertyID:   "p-001",
		GroupName:    "Wedding Party - Johnson",
		ContactName:  "Sarah Johnson",
		ContactEmail: "sarah.j@example.com",
		ContactPhone: "+1-555-0100",
		NumRooms:     5,
		NumGuests:    12,
		CheckIn:      time.Now().AddDate(0, 0, 14),
		CheckOut:     time.Now().AddDate(0, 0, 16),
		Status:       domain.GroupStatusDefinite,
		RateCode:     "GROUP-WEDDING",
	})
	r.Create(context.Background(), &domain.GroupReservation{
		TenantID:     "demo",
		PropertyID:   "p-001",
		GroupName:    "Corporate Retreat - TechCorp",
		ContactName:  "Mike Chen",
		ContactEmail: "mike@techcorp.com",
		ContactPhone: "+1-555-0200",
		NumRooms:     8,
		NumGuests:    15,
		CheckIn:      time.Now().AddDate(0, 0, 7),
		CheckOut:     time.Now().AddDate(0, 0, 9),
		Status:       domain.GroupStatusTentative,
		RateCode:     "CORP-RETREAT",
	})
}
