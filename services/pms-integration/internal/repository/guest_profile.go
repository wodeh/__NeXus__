package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
)

// GuestProfileRepository manages enriched guest CRM data.
type GuestProfileRepository struct {
	mu       sync.RWMutex
	profiles map[string]*domain.GuestProfile
	seq      int64
}

// NewGuestProfileRepository creates a new guest profile repository.
func NewGuestProfileRepository() *GuestProfileRepository {
	return &GuestProfileRepository{
		profiles: make(map[string]*domain.GuestProfile),
		seq:      2000,
	}
}

// ListByTenant returns all guest profiles for a tenant.
func (r *GuestProfileRepository) ListByTenant(ctx context.Context, tenant string, limit, offset int) ([]*domain.GuestProfile, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var results []*domain.GuestProfile
	for _, p := range r.profiles {
		if p.TenantID != tenant {
			continue
		}
		results = append(results, p)
	}
	total := len(results)
	if offset >= total {
		return []*domain.GuestProfile{}, total, nil
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return results[offset:end], total, nil
}

// GetByID returns a specific profile.
func (r *GuestProfileRepository) GetByID(ctx context.Context, tenant, id string) (*domain.GuestProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.profiles[id]
	if !ok || p.TenantID != tenant {
		return nil, fmt.Errorf("guest profile not found")
	}
	return p, nil
}

// GetByEmail looks up a profile by email.
func (r *GuestProfileRepository) GetByEmail(ctx context.Context, tenant, email string) (*domain.GuestProfile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, p := range r.profiles {
		if p.TenantID == tenant && p.Email == email {
			return p, nil
		}
	}
	return nil, fmt.Errorf("guest profile not found")
}

// Create stores a new profile.
func (r *GuestProfileRepository) Create(ctx context.Context, p *domain.GuestProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.seq++
	p.ID = fmt.Sprintf("gp-%d", r.seq)
	p.CreatedAt = time.Now()
	p.UpdatedAt = p.CreatedAt
	p.VIPStatus = domain.VIPNone
	r.profiles[p.ID] = p
	return nil
}

// Update modifies an existing profile.
func (r *GuestProfileRepository) Update(ctx context.Context, tenant, id string, updates *domain.GuestProfile) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.profiles[id]
	if !ok || p.TenantID != tenant {
		return fmt.Errorf("guest profile not found")
	}
	if updates.VIPStatus != "" {
		p.VIPStatus = updates.VIPStatus
	}
	if updates.LoyaltyTier != "" {
		p.LoyaltyTier = updates.LoyaltyTier
	}
	if updates.Notes != "" {
		p.Notes = updates.Notes
	}
	p.UpdatedAt = time.Now()
	return nil
}

// AddCommunication appends a log entry.
func (r *GuestProfileRepository) AddCommunication(ctx context.Context, tenant, id string, entry domain.CommunicationEntry) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.profiles[id]
	if !ok || p.TenantID != tenant {
		return fmt.Errorf("guest profile not found")
	}
	p.CommunicationLog = append(p.CommunicationLog, entry)
	p.UpdatedAt = time.Now()
	return nil
}

// SeedGuestProfiles pre-populates demo data.
func (r *GuestProfileRepository) SeedGuestProfiles() {
	now := time.Now()
	r.Create(context.Background(), &domain.GuestProfile{
		TenantID:       "demo",
		FirstName:      "Alice",
		LastName:       "Chen",
		Email:          "alice.chen@example.com",
		Phone:          "+1-555-0101",
		VIPStatus:      domain.VIPGold,
		LoyaltyTier:    "Gold",
		TotalStays:     12,
		TotalNights:    45,
		TotalRevenue:   12500.00,
		AverageDailyRate: 277.78,
		LastStayDate:   &now,
		Preferences: domain.GuestPreferences{
			RoomType: "Deluxe King",
			Floor:    "3",
			BedType:  "King",
			Amenities: []string{"extra pillows", "late checkout"},
		},
		Notes: "Prefers quiet rooms. Allergic to down feathers.",
	})
	r.Create(context.Background(), &domain.GuestProfile{
		TenantID:       "demo",
		FirstName:      "Bob",
		LastName:       "Jones",
		Email:          "bob.jones@example.com",
		Phone:          "+1-555-0102",
		VIPStatus:      domain.VIPSilver,
		LoyaltyTier:    "Silver",
		TotalStays:     5,
		TotalNights:    15,
		TotalRevenue:   3500.00,
		AverageDailyRate: 233.33,
		LastStayDate:   &now,
		Preferences: domain.GuestPreferences{
			RoomType: "Standard",
			Floor:    "2",
			Smoking:  false,
		},
		Notes: "Business traveler. Early riser.",
	})
}
