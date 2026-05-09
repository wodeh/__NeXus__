package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/backend-core/internal/db"
)

// Tenant represents a hospitality group / brand in the platform.
type Tenant struct {
	ID         uuid.UUID `json:"id"`
	ExternalID string    `json:"external_id"`
	Name       string    `json:"name"`
	Region     string    `json:"region"`
	Tier       string    `json:"tier"`
	Config     map[string]interface{} `json:"config"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	Version    int       `json:"version"`
}

// TenantProperty represents a single hotel/property within a tenant.
type TenantProperty struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	PropertyID string    `json:"property_id"`
	Name       string    `json:"name"`
	Timezone   string    `json:"timezone"`
	Locale     string    `json:"locale"`
	Currency   string    `json:"currency"`
	Config     map[string]interface{} `json:"config"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	Version    int       `json:"version"`
}

// TenantRepository provides tenant-scoped data access with explicit SQL.
type TenantRepository struct {
	pool *db.Pool
}

// NewTenantRepository creates a tenant repository backed by a connection pool.
func NewTenantRepository(pool *db.Pool) *TenantRepository {
	return &TenantRepository{pool: pool}
}

// GetByExternalID retrieves a tenant by its external slug. Returns ErrTenantNotFound if absent.
func (r *TenantRepository) GetByExternalID(ctx context.Context, externalID string) (*Tenant, error) {
	if externalID == "" {
		return nil, fmt.Errorf("external_id required")
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, external_id, name, region, tier, config, created_at, updated_at, deleted_at, version
		FROM tenants
		WHERE external_id = $1 AND deleted_at IS NULL
	`, externalID)

	var t Tenant
	var configJSON []byte
	if err := row.Scan(&t.ID, &t.ExternalID, &t.Name, &t.Region, &t.Tier, &configJSON, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &t.Version); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("tenant %s: %w", externalID, ErrTenantNotFound)
		}
		return nil, fmt.Errorf("get tenant: %w", err)
	}

	return &t, nil
}

// Create inserts a new tenant. Uses explicit SQL with RETURNING.
func (r *TenantRepository) Create(ctx context.Context, tx pgx.Tx, t *Tenant) error {
	if t.ExternalID == "" || t.Name == "" {
		return fmt.Errorf("external_id and name required")
	}
	if t.ID == uuid.Nil {
		t.ID = uuid.Must(uuid.NewRandom())
	}
	now := time.Now().UTC()
	t.CreatedAt = now
	t.UpdatedAt = now

	// Use tx if provided, otherwise pool
	execer := r.execer(tx)
	_, err := execer.Exec(ctx, `
		INSERT INTO tenants (id, external_id, name, region, tier, config, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, t.ID, t.ExternalID, t.Name, t.Region, t.Tier, t.Config, t.CreatedAt, t.UpdatedAt, t.Version)
	if err != nil {
		return fmt.Errorf("insert tenant: %w", err)
	}
	return nil
}

// SoftDelete marks a tenant as deleted (optimistic locking via version).
func (r *TenantRepository) SoftDelete(ctx context.Context, tx pgx.Tx, tenantID uuid.UUID, expectedVersion int) error {
	execer := r.execer(tx)
	cmdTag, err := execer.Exec(ctx, `
		UPDATE tenants
		SET deleted_at = NOW(), version = version + 1
		WHERE id = $1 AND deleted_at IS NULL AND version = $2
	`, tenantID, expectedVersion)
	if err != nil {
		return fmt.Errorf("soft delete tenant: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("tenant %s version %d: %w", tenantID, expectedVersion, ErrOptimisticLock)
	}
	return nil
}

// ListProperties returns all active properties for a tenant.
func (r *TenantRepository) ListProperties(ctx context.Context, tenantID uuid.UUID) ([]TenantProperty, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, name, timezone, locale, currency, config, created_at, updated_at, deleted_at, version
		FROM tenant_properties
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY name
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list properties: %w", err)
	}
	defer rows.Close()

	var props []TenantProperty
	for rows.Next() {
		var p TenantProperty
		var configJSON []byte
		if err := rows.Scan(&p.ID, &p.TenantID, &p.PropertyID, &p.Name, &p.Timezone, &p.Locale, &p.Currency, &configJSON, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt, &p.Version); err != nil {
			return nil, fmt.Errorf("scan property: %w", err)
		}
		props = append(props, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("property rows: %w", err)
	}
	return props, nil
}

// execer returns the transactional executor if tx is non-nil, otherwise the pool.
func (r *TenantRepository) execer(tx pgx.Tx) interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
} {
	if tx != nil {
		return tx
	}
	return r.pool
}

// Repository errors.
var (
	ErrTenantNotFound  = fmt.Errorf("tenant not found")
	ErrOptimisticLock  = fmt.Errorf("optimistic locking conflict")
)
