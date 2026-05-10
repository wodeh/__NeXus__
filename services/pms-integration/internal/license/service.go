package license

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
)

// Service manages tenant configurations and capability checks.
type Service struct {
	mu      sync.RWMutex
	configs map[string]*domain.TenantConfig
}

// NewService creates a license service with demo data.
func NewService() *Service {
	s := &Service{configs: make(map[string]*domain.TenantConfig)}
	// Seed demo tenants with different tiers for testing
	s.seedDemoTenants()
	return s
}

func (s *Service) seedDemoTenants() {
	s.configs["demo"] = domain.NewTenantConfig("demo", "Demo Hotel", domain.PropertyTypeBoutique, domain.LicenseTierEnterprise)
	s.configs["demo"].LicenseStatus = domain.LicenseStatusActive
	s.configs["demo"].LicenseExpires = time.Now().AddDate(1, 0, 0)

	s.configs["grand-plaza"] = domain.NewTenantConfig("grand-plaza", "Grand Plaza", domain.PropertyTypeResort, domain.LicenseTierRevenue)
	s.configs["grand-plaza"].LicenseStatus = domain.LicenseStatusActive

	s.configs["sunset-inn"] = domain.NewTenantConfig("sunset-inn", "Sunset Inn", domain.PropertyTypeMotel, domain.LicenseTierCore)
	s.configs["sunset-inn"].LicenseStatus = domain.LicenseStatusActive
}

// GetConfig returns tenant configuration by ID.
func (s *Service) GetConfig(ctx context.Context, tenantID string) (*domain.TenantConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg, ok := s.configs[tenantID]
	if !ok {
		return nil, fmt.Errorf("tenant not found: %s", tenantID)
	}
	return cfg, nil
}

// UpsertConfig creates or updates a tenant configuration.
func (s *Service) UpsertConfig(ctx context.Context, cfg *domain.TenantConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cfg.UpdatedAt = time.Now()
	if cfg.CreatedAt.IsZero() {
		cfg.CreatedAt = cfg.UpdatedAt
	}
	s.configs[cfg.ID] = cfg
	return nil
}

// CheckCapability verifies a tenant has the required capability.
func (s *Service) CheckCapability(ctx context.Context, tenantID string, cap domain.Capability) (bool, error) {
	cfg, err := s.GetConfig(ctx, tenantID)
	if err != nil {
		return false, err
	}
	if cfg.LicenseStatus == domain.LicenseStatusExpired || cfg.LicenseStatus == domain.LicenseStatusSuspended {
		return false, fmt.Errorf("license %s", cfg.LicenseStatus)
	}
	return cfg.HasCapability(cap), nil
}

// CheckAnyCapability verifies a tenant has at least one of the required capabilities.
func (s *Service) CheckAnyCapability(ctx context.Context, tenantID string, caps ...domain.Capability) (bool, error) {
	cfg, err := s.GetConfig(ctx, tenantID)
	if err != nil {
		return false, err
	}
	if cfg.LicenseStatus == domain.LicenseStatusExpired || cfg.LicenseStatus == domain.LicenseStatusSuspended {
		return false, fmt.Errorf("license %s", cfg.LicenseStatus)
	}
	for _, cap := range caps {
		if cfg.HasCapability(cap) {
			return true, nil
		}
	}
	return false, nil
}

// ListTenants returns all configured tenants.
func (s *Service) ListTenants(ctx context.Context) ([]*domain.TenantConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var list []*domain.TenantConfig
	for _, cfg := range s.configs {
		list = append(list, cfg)
	}
	return list, nil
}

// PropertyTypes returns available property types.
func (s *Service) PropertyTypes() []map[string]string {
	return []map[string]string{
		{"id": string(domain.PropertyTypeBoutique), "name": "Boutique Hotel", "description": "1-50 rooms, personalized service"},
		{"id": string(domain.PropertyTypeMotel), "name": "Motel", "description": "Highway roadside, automated check-in"},
		{"id": string(domain.PropertyTypeResort), "name": "Resort", "description": "100-500+ rooms, leisure focused"},
		{"id": string(domain.PropertyTypeHostel), "name": "Hostel", "description": "Shared rooms, social spaces"},
		{"id": string(domain.PropertyTypeAparthotel), "name": "Aparthotel", "description": "Extended stay with kitchenettes"},
		{"id": string(domain.PropertyTypeBnB), "name": "Bed & Breakfast", "description": "Small owner-operated"},
	}
}

// LicenseTiers returns available license tiers with pricing.
func (s *Service) LicenseTiers() []map[string]interface{} {
	return []map[string]interface{}{
		{"id": string(domain.LicenseTierCore), "name": "Core", "price_monthly": 99, "max_rooms": 50, "max_users": 5},
		{"id": string(domain.LicenseTierOperations), "name": "Operations", "price_monthly": 199, "max_rooms": 150, "max_users": 20},
		{"id": string(domain.LicenseTierRevenue), "name": "Revenue", "price_monthly": 349, "max_rooms": 500, "max_users": 50},
		{"id": string(domain.LicenseTierEnterprise), "name": "Enterprise", "price_monthly": 599, "max_rooms": 5000, "max_users": 200},
	}
}
