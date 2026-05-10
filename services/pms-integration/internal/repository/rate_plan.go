package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
)

// RatePlanRepository manages rate plans.
type RatePlanRepository struct {
	mu        sync.RWMutex
	plans     map[string]*domain.RatePlanSeasonal
	seq       int64
}

// NewRatePlanRepository creates a new rate plan repository.
func NewRatePlanRepository() *RatePlanRepository {
	return &RatePlanRepository{
		plans: make(map[string]*domain.RatePlanSeasonal),
		seq:   3000,
	}
}

// ListByTenant returns all rate plans for a tenant.
func (r *RatePlanRepository) ListByTenant(ctx context.Context, tenant string, limit, offset int) ([]*domain.RatePlanSeasonal, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*domain.RatePlanSeasonal
	for _, p := range r.plans {
		if p.TenantID != tenant {
			continue
		}
		results = append(results, p)
	}
	total := len(results)
	if offset >= total {
		return []*domain.RatePlanSeasonal{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return results[offset:end], total, nil
}

// GetByID returns a specific rate plan.
func (r *RatePlanRepository) GetByID(ctx context.Context, tenant, id string) (*domain.RatePlanSeasonal, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.plans[id]
	if !ok || p.TenantID != tenant {
		return nil, fmt.Errorf("rate plan not found")
	}
	return p, nil
}

// Create stores a new rate plan.
func (r *RatePlanRepository) Create(ctx context.Context, p *domain.RatePlanSeasonal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	p.ID = fmt.Sprintf("rp-%d", r.seq)
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	p.IsActive = true
	r.plans[p.ID] = p
	return nil
}

// Update modifies an existing rate plan.
func (r *RatePlanRepository) Update(ctx context.Context, tenant, id string, updates *domain.RatePlanSeasonal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.plans[id]
	if !ok || p.TenantID != tenant {
		return fmt.Errorf("rate plan not found")
	}
	if updates.Name != "" {
		p.Name = updates.Name
	}
	if updates.Description != "" {
		p.Description = updates.Description
	}
	if updates.IsActive != p.IsActive {
		p.IsActive = updates.IsActive
	}
	if len(updates.SeasonalRates) > 0 {
		p.SeasonalRates = updates.SeasonalRates
	}
	p.UpdatedAt = time.Now()
	return nil
}

// Delete removes a rate plan.
func (r *RatePlanRepository) Delete(ctx context.Context, tenant, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.plans[id]
	if !ok || p.TenantID != tenant {
		return fmt.Errorf("rate plan not found")
	}
	delete(r.plans, id)
	return nil
}

// SeedRatePlans pre-populates demo data.
func (r *RatePlanRepository) SeedRatePlans() {
	r.Create(context.Background(), &domain.RatePlanSeasonal{
		TenantID:    "demo",
		PropertyID:  "p-001",
		Name:        "Summer 2026 Standard",
		Code:        "SUM-STD",
		Description: "Standard room summer rate",
		SeasonalRates: []domain.SeasonalRate{
			{RoomTypeID: "rt-001", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 3, 0), BasePrice: 129.00},
			{RoomTypeID: "rt-002", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 3, 0), BasePrice: 189.00},
		},
		Restrictions: domain.RateRestriction{MinStay: 1, MaxStay: 14},
		IsActive:     true,
	})
	r.Create(context.Background(), &domain.RatePlanSeasonal{
		TenantID:    "demo",
		PropertyID:  "p-001",
		Name:        "Corporate Weekend",
		Code:        "CORP-WKD",
		Description: "Corporate negotiated weekend rate",
		SeasonalRates: []domain.SeasonalRate{
			{RoomTypeID: "rt-001", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 90), BasePrice: 99.00},
			{RoomTypeID: "rt-003", StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 90), BasePrice: 249.00},
		},
		Restrictions: domain.RateRestriction{MinStay: 2, ClosedToArrival: false, ClosedToDeparture: false},
		IsActive:     true,
	})
}
