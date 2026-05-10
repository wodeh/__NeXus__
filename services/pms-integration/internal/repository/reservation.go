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

// ReservationRepository persists and retrieves Reservation aggregates.
type ReservationRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.Reservation, error)
	Create(ctx context.Context, r *domain.Reservation) error
	Update(ctx context.Context, r *domain.Reservation) error
	SoftDelete(ctx context.Context, tenantID, id string) error
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Reservation, error)
	ListByGuest(ctx context.Context, tenantID, guestID string, limit, offset int) ([]*domain.Reservation, error)
	ListByStatus(ctx context.Context, tenantID string, status domain.ReservationStatus, limit, offset int) ([]*domain.Reservation, error)
}

// PostgresReservationRepository is the PostgreSQL implementation of ReservationRepository.
type PostgresReservationRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewReservationRepository creates a new PostgreSQL reservation repository.
func NewReservationRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresReservationRepository {
	return &PostgresReservationRepository{pool: pool, metrics: metrics}
}

func (r *PostgresReservationRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// GetByID retrieves a reservation by ID with tenant isolation.
func (r *PostgresReservationRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Reservation, error) {
	start := time.Now()
	r.metrics.IncQuery("reservation", "get_by_id")
	defer r.metrics.ObserveDuration("reservation", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("reservation", "get_by_id", "tenant_bind")
		return nil, err
	}

	var res domain.Reservation
	var specialRequests []string
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, guest_id, room_id, check_in_date, check_out_date, status, special_requests, created_at, updated_at, version
		FROM reservations
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID).Scan(
		&res.ID, &res.TenantID, &res.PropertyID, &res.GuestID, &res.RoomID,
		&res.CheckInDate, &res.CheckOutDate, &res.Status, &specialRequests, &res.CreatedAt, &res.UpdatedAt, &res.Version,
	)
	if err != nil {
		r.metrics.IncError("reservation", "get_by_id", "query")
		return nil, fmt.Errorf("get reservation: %w", err)
	}
	res.SpecialRequests = specialRequests
	return &res, nil
}

// Create inserts a new reservation.
func (r *PostgresReservationRepository) Create(ctx context.Context, res *domain.Reservation) error {
	start := time.Now()
	r.metrics.IncQuery("reservation", "create")
	defer r.metrics.ObserveDuration("reservation", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, res.TenantID); err != nil {
		r.metrics.IncError("reservation", "create", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO reservations (id, tenant_id, property_id, guest_id, room_id, check_in_date, check_out_date, status, special_requests, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, res.ID, res.TenantID, res.PropertyID, res.GuestID, res.RoomID, res.CheckInDate, res.CheckOutDate, res.Status, res.SpecialRequests, res.CreatedAt, res.UpdatedAt, res.Version)
	if err != nil {
		r.metrics.IncError("reservation", "create", "query")
		return fmt.Errorf("create reservation: %w", err)
	}
	return nil
}

// Update modifies a reservation with optimistic locking.
func (r *PostgresReservationRepository) Update(ctx context.Context, res *domain.Reservation) error {
	start := time.Now()
	r.metrics.IncQuery("reservation", "update")
	defer r.metrics.ObserveDuration("reservation", "update", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, res.TenantID); err != nil {
		r.metrics.IncError("reservation", "update", "tenant_bind")
		return err
	}

	res.UpdatedAt = time.Now().UTC()
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE reservations
		SET property_id = $1, guest_id = $2, room_id = $3, check_in_date = $4, check_out_date = $5, status = $6, special_requests = $7, updated_at = $8, version = version + 1
		WHERE id = $9 AND tenant_id = $10 AND version = $11 AND deleted_at IS NULL
	`, res.PropertyID, res.GuestID, res.RoomID, res.CheckInDate, res.CheckOutDate, res.Status, res.SpecialRequests, res.UpdatedAt, res.ID, res.TenantID, res.Version)
	if err != nil {
		r.metrics.IncError("reservation", "update", "query")
		return fmt.Errorf("update reservation: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		r.metrics.IncError("reservation", "update", "optimistic_lock")
		return fmt.Errorf("reservation update conflict: id=%s version=%d", res.ID, res.Version)
	}
	res.Version++
	return nil
}

// SoftDelete marks a reservation as deleted.
func (r *PostgresReservationRepository) SoftDelete(ctx context.Context, tenantID, id string) error {
	start := time.Now()
	r.metrics.IncQuery("reservation", "soft_delete")
	defer r.metrics.ObserveDuration("reservation", "soft_delete", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("reservation", "soft_delete", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE reservations SET deleted_at = NOW() WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID)
	if err != nil {
		r.metrics.IncError("reservation", "soft_delete", "query")
		return fmt.Errorf("soft delete reservation: %w", err)
	}
	return nil
}

// ListByTenant returns paginated reservations for a tenant.
func (r *PostgresReservationRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Reservation, error) {
	start := time.Now()
	r.metrics.IncQuery("reservation", "list_by_tenant")
	defer r.metrics.ObserveDuration("reservation", "list_by_tenant", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("reservation", "list_by_tenant", "tenant_bind")
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, guest_id, room_id, check_in_date, check_out_date, status, special_requests, created_at, updated_at, version
		FROM reservations
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)
	if err != nil {
		r.metrics.IncError("reservation", "list_by_tenant", "query")
		return nil, fmt.Errorf("list reservations: %w", err)
	}
	defer rows.Close()

	return scanReservations(rows)
}

// ListByGuest returns paginated reservations for a guest.
func (r *PostgresReservationRepository) ListByGuest(ctx context.Context, tenantID, guestID string, limit, offset int) ([]*domain.Reservation, error) {
	start := time.Now()
	r.metrics.IncQuery("reservation", "list_by_guest")
	defer r.metrics.ObserveDuration("reservation", "list_by_guest", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("reservation", "list_by_guest", "tenant_bind")
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, guest_id, room_id, check_in_date, check_out_date, status, special_requests, created_at, updated_at, version
		FROM reservations
		WHERE tenant_id = $1 AND guest_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, tenantID, guestID, limit, offset)
	if err != nil {
		r.metrics.IncError("reservation", "list_by_guest", "query")
		return nil, fmt.Errorf("list reservations by guest: %w", err)
	}
	defer rows.Close()

	return scanReservations(rows)
}

// ListByStatus returns paginated reservations by status.
func (r *PostgresReservationRepository) ListByStatus(ctx context.Context, tenantID string, status domain.ReservationStatus, limit, offset int) ([]*domain.Reservation, error) {
	start := time.Now()
	r.metrics.IncQuery("reservation", "list_by_status")
	defer r.metrics.ObserveDuration("reservation", "list_by_status", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("reservation", "list_by_status", "tenant_bind")
		return nil, err
	}

	if limit <= 0 || limit > 100 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, guest_id, room_id, check_in_date, check_out_date, status, special_requests, created_at, updated_at, version
		FROM reservations
		WHERE tenant_id = $1 AND status = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, tenantID, status, limit, offset)
	if err != nil {
		r.metrics.IncError("reservation", "list_by_status", "query")
		return nil, fmt.Errorf("list reservations by status: %w", err)
	}
	defer rows.Close()

	return scanReservations(rows)
}

func scanReservations(rows pgx.Rows) ([]*domain.Reservation, error) {
	var reservations []*domain.Reservation
	for rows.Next() {
		var res domain.Reservation
		var specialRequests []string
		if err := rows.Scan(
			&res.ID, &res.TenantID, &res.PropertyID, &res.GuestID, &res.RoomID,
			&res.CheckInDate, &res.CheckOutDate, &res.Status, &specialRequests, &res.CreatedAt, &res.UpdatedAt, &res.Version,
		); err != nil {
			return nil, fmt.Errorf("scan reservation: %w", err)
		}
		res.SpecialRequests = specialRequests
		reservations = append(reservations, &res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return reservations, nil
}
