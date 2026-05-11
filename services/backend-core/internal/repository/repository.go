// Package repository provides tenant-aware data access for backend-core.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/nexus-platform/backend-core/internal/db"
)

// Store is the root repository backed by a PostgreSQL connection pool.
type Store struct {
	pool      *db.Pool
	Tenants   *TenantRepository
	Admin     *AdminRepository
	Channel   *ChannelRepository
}

// NewStore creates a repository instance.
func NewStore(pool *db.Pool) *Store {
	return &Store{
		pool:    pool,
		Tenants: NewTenantRepository(pool),
		Admin:   NewAdminRepository(pool),
		Channel: NewChannelRepository(pool),
	}
}

// Close releases repository resources.
func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// Ping verifies database connectivity.
func (s *Store) Ping(ctx context.Context) error {
	if s.pool == nil {
		return fmt.Errorf("pool not initialized")
	}
	return s.pool.Ping(ctx)
}

// SetTenant binds the underlying pool to a tenant for RLS-aware queries.
func (s *Store) SetTenant(ctx context.Context, tenantID string) error {
	return s.pool.SetTenant(ctx, tenantID)
}

// ResetTenant clears the tenant binding.
func (s *Store) ResetTenant(ctx context.Context) error {
	return s.pool.ResetTenant(ctx)
}

// Pool returns the underlying database pool for direct repository creation.
func (s *Store) Pool() *db.Pool {
	return s.pool
}
// Deprecated: use TenantRepository.GetByExternalID for full tenant records.
type TenantConfig struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Region    string    `json:"region"`
	Tier      string    `json:"tier"`
	CreatedAt time.Time `json:"created_at"`
}

// GetTenantConfig returns tenant configuration.
// Phase 1 backward-compatibility: returns synthetic config when DB is unreachable.
func (s *Store) GetTenantConfig(ctx context.Context, tenantID string) (*TenantConfig, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id required")
	}
	if s.pool == nil {
		return &TenantConfig{
			ID:        tenantID,
			Name:      tenantID,
			Region:    "us-east-1",
			Tier:      "enterprise",
			CreatedAt: time.Now().UTC(),
		}, nil
	}
	return nil, fmt.Errorf("GetTenantConfig not yet implemented against real schema (use Tenants.GetByExternalID)")
}
