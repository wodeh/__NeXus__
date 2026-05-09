// Package repository provides tenant-aware data access for PMS integration.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// GuestRepository persists and retrieves Guest aggregates.
type GuestRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.Guest, error)
	Create(ctx context.Context, g *domain.Guest) error
	Update(ctx context.Context, g *domain.Guest) error
	SoftDelete(ctx context.Context, tenantID, id string) error
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Guest, error)
	GetByEmail(ctx context.Context, tenantID, email string) (*domain.Guest, error)
}

// PostgresGuestRepository is the PostgreSQL implementation of GuestRepository.
type PostgresGuestRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewGuestRepository creates a new PostgreSQL guest repository.
func NewGuestRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresGuestRepository {
	return &PostgresGuestRepository{pool: pool, metrics: metrics}
}

func (r *PostgresGuestRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// GetByID retrieves a guest by ID with tenant isolation.
func (r *PostgresGuestRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Guest, error) {
	start := time.Now()
	r.metrics.IncQuery("guest", "get_by_id")
	defer r.metrics.ObserveDuration("guest", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("guest", "get_by_id", "tenant_bind")
		return nil, err
	}

	var g domain.Guest
	var preferences map[string]string
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, first_name, last_name, email, phone, loyalty_member_id, preferences, created_at, updated_at, version
		FROM guests
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID).Scan(
		&g.ID, &g.TenantID, &g.PropertyID, &g.FirstName, &g.LastName, &g.Email, &g.Phone,
		&g.LoyaltyID, &preferences, &g.CreatedAt, &g.UpdatedAt, &g.Version,
	)
	if err != nil {
		r.metrics.IncError("guest", "get_by_id", "query")
		return nil, fmt.Errorf("get guest: %w", err)
	}
	g.Preferences = preferences
	return &g, nil
}

// Create inserts a new guest.
func (r *PostgresGuestRepository) Create(ctx context.Context, g *domain.Guest) error {
	start := time.Now()
	r.metrics.IncQuery("guest", "create")
	defer r.metrics.ObserveDuration("guest", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, g.TenantID); err != nil {
		r.metrics.IncError("guest", "create", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO guests (id, tenant_id, property_id, first_name, last_name, email, phone, loyalty_member_id, preferences, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, g.ID, g.TenantID, g.PropertyID, g.FirstName, g.LastName, g.Email, g.Phone, g.LoyaltyID, g.Preferences, g.CreatedAt, g.UpdatedAt, g.Version)
	if err != nil {
		r.metrics.IncError("guest", "create", "query")
		return fmt.Errorf("create guest: %w", err)
	}
	return nil
}

// Update modifies a guest with optimistic locking.
func (r *PostgresGuestRepository) Update(ctx context.Context, g *domain.Guest) error {
	start := time.Now()
	r.metrics.IncQuery("guest", "update")
	defer r.metrics.ObserveDuration("guest", "update", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, g.TenantID); err != nil {
		r.metrics.IncError("guest", "update", "tenant_bind")
		return err
	}

	g.UpdatedAt = time.Now().UTC()
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE guests
		SET property_id = $1, first_name = $2, last_name = $3, email = $4, phone = $5, loyalty_member_id = $6, preferences = $7, updated_at = $8, version = version + 1
		WHERE id = $9 AND tenant_id = $10 AND version = $11 AND deleted_at IS NULL
	`, g.PropertyID, g.FirstName, g.LastName, g.Email, g.Phone, g.LoyaltyID, g.Preferences, g.UpdatedAt, g.ID, g.TenantID, g.Version)
	if err != nil {
		r.metrics.IncError("guest", "update", "query")
		return fmt.Errorf("update guest: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		r.metrics.IncError("guest", "update", "optimistic_lock")
		return fmt.Errorf("guest update conflict: id=%s version=%d", g.ID, g.Version)
	}
	g.Version++
	return nil
}

// SoftDelete marks a guest as deleted.
func (r *PostgresGuestRepository) SoftDelete(ctx context.Context, tenantID, id string) error {
	start := time.Now()
	r.metrics.IncQuery("guest", "soft_delete")
	defer r.metrics.ObserveDuration("guest", "soft_delete", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("guest", "soft_delete", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE guests SET deleted_at = NOW() WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID)
	if err != nil {
		r.metrics.IncError("guest", "soft_delete", "query")
		return fmt.Errorf("soft delete guest: %w", err)
	}
	return nil
}

// ListByTenant returns paginated guests for a tenant.
func (r *PostgresGuestRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Guest, error) {
	start := time.Now()
	r.metrics.IncQuery("guest", "list_by_tenant")
	defer r.metrics.ObserveDuration("guest", "list_by_tenant", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("guest", "list_by_tenant", "tenant_bind")
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, first_name, last_name, email, phone, loyalty_member_id, preferences, created_at, updated_at, version
		FROM guests
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)
	if err != nil {
		r.metrics.IncError("guest", "list_by_tenant", "query")
		return nil, fmt.Errorf("list guests: %w", err)
	}
	defer rows.Close()

	return scanGuests(rows)
}

// GetByEmail retrieves a guest by email with tenant isolation.
func (r *PostgresGuestRepository) GetByEmail(ctx context.Context, tenantID, email string) (*domain.Guest, error) {
	start := time.Now()
	r.metrics.IncQuery("guest", "get_by_email")
	defer r.metrics.ObserveDuration("guest", "get_by_email", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("guest", "get_by_email", "tenant_bind")
		return nil, err
	}

	var g domain.Guest
	var preferences map[string]string
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, first_name, last_name, email, phone, loyalty_member_id, preferences, created_at, updated_at, version
		FROM guests
		WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL
	`, tenantID, email).Scan(
		&g.ID, &g.TenantID, &g.PropertyID, &g.FirstName, &g.LastName, &g.Email, &g.Phone,
		&g.LoyaltyID, &preferences, &g.CreatedAt, &g.UpdatedAt, &g.Version,
	)
	if err != nil {
		r.metrics.IncError("guest", "get_by_email", "query")
		return nil, fmt.Errorf("get guest by email: %w", err)
	}
	g.Preferences = preferences
	return &g, nil
}

func scanGuests(rows pgx.Rows) ([]*domain.Guest, error) {
	var guests []*domain.Guest
	for rows.Next() {
		var g domain.Guest
		var preferences map[string]string
		if err := rows.Scan(
			&g.ID, &g.TenantID, &g.PropertyID, &g.FirstName, &g.LastName, &g.Email, &g.Phone,
			&g.LoyaltyID, &preferences, &g.CreatedAt, &g.UpdatedAt, &g.Version,
		); err != nil {
			return nil, fmt.Errorf("scan guest: %w", err)
		}
		g.Preferences = preferences
		guests = append(guests, &g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return guests, nil
}
