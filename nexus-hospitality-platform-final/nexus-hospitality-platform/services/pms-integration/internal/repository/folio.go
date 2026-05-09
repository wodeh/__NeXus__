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

// FolioRepository persists and retrieves Folio aggregates.
type FolioRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.Folio, error)
	GetByReservation(ctx context.Context, tenantID, reservationID string) (*domain.Folio, error)
	Create(ctx context.Context, f *domain.Folio) error
	Update(ctx context.Context, f *domain.Folio) error
	AddCharge(ctx context.Context, tenantID, folioID string, charge domain.Charge) error
	AddPayment(ctx context.Context, tenantID, folioID string, payment domain.Payment) error
	SoftDelete(ctx context.Context, tenantID, id string) error
}

// PostgresFolioRepository is the PostgreSQL implementation of FolioRepository.
type PostgresFolioRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewFolioRepository creates a new PostgreSQL folio repository.
func NewFolioRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresFolioRepository {
	return &PostgresFolioRepository{pool: pool, metrics: metrics}
}

func (r *PostgresFolioRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// GetByID retrieves a folio by ID with tenant isolation.
func (r *PostgresFolioRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Folio, error) {
	start := time.Now()
	r.metrics.IncQuery("folio", "get_by_id")
	defer r.metrics.ObserveDuration("folio", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("folio", "get_by_id", "tenant_bind")
		return nil, err
	}

	var f domain.Folio
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, reservation_id, guest_id, status, balance, currency_code, created_at, updated_at, version
		FROM folios
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID).Scan(
		&f.ID, &f.TenantID, &f.PropertyID, &f.ReservationID, &f.GuestID, &f.Status, &f.Balance, &f.Currency, &f.CreatedAt, &f.UpdatedAt, &f.Version,
	)
	if err != nil {
		r.metrics.IncError("folio", "get_by_id", "query")
		return nil, fmt.Errorf("get folio: %w", err)
	}

	charges, payments, err := r.loadTransactions(ctx, tenantID, id)
	if err != nil {
		r.metrics.IncError("folio", "get_by_id", "transactions")
		return nil, fmt.Errorf("load folio transactions: %w", err)
	}
	f.Charges = charges
	f.Payments = payments

	return &f, nil
}

// GetByReservation retrieves a folio by reservation ID.
func (r *PostgresFolioRepository) GetByReservation(ctx context.Context, tenantID, reservationID string) (*domain.Folio, error) {
	start := time.Now()
	r.metrics.IncQuery("folio", "get_by_reservation")
	defer r.metrics.ObserveDuration("folio", "get_by_reservation", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("folio", "get_by_reservation", "tenant_bind")
		return nil, err
	}

	var f domain.Folio
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, reservation_id, guest_id, status, balance, currency_code, created_at, updated_at, version
		FROM folios
		WHERE reservation_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, reservationID, tenantID).Scan(
		&f.ID, &f.TenantID, &f.PropertyID, &f.ReservationID, &f.GuestID, &f.Status, &f.Balance, &f.Currency, &f.CreatedAt, &f.UpdatedAt, &f.Version,
	)
	if err != nil {
		r.metrics.IncError("folio", "get_by_reservation", "query")
		return nil, fmt.Errorf("get folio by reservation: %w", err)
	}

	charges, payments, err := r.loadTransactions(ctx, tenantID, string(f.ID))
	if err != nil {
		r.metrics.IncError("folio", "get_by_reservation", "transactions")
		return nil, fmt.Errorf("load folio transactions: %w", err)
	}
	f.Charges = charges
	f.Payments = payments

	return &f, nil
}

// Create inserts a new folio.
func (r *PostgresFolioRepository) Create(ctx context.Context, f *domain.Folio) error {
	start := time.Now()
	r.metrics.IncQuery("folio", "create")
	defer r.metrics.ObserveDuration("folio", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, f.TenantID); err != nil {
		r.metrics.IncError("folio", "create", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO folios (id, tenant_id, property_id, reservation_id, guest_id, status, balance, currency_code, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, f.ID, f.TenantID, f.PropertyID, f.ReservationID, f.GuestID, f.Status, f.Balance, f.Currency, f.CreatedAt, f.UpdatedAt, f.Version)
	if err != nil {
		r.metrics.IncError("folio", "create", "query")
		return fmt.Errorf("create folio: %w", err)
	}
	return nil
}

// Update modifies a folio with optimistic locking.
func (r *PostgresFolioRepository) Update(ctx context.Context, f *domain.Folio) error {
	start := time.Now()
	r.metrics.IncQuery("folio", "update")
	defer r.metrics.ObserveDuration("folio", "update", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, f.TenantID); err != nil {
		r.metrics.IncError("folio", "update", "tenant_bind")
		return err
	}

	f.UpdatedAt = time.Now().UTC()
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE folios
		SET property_id = $1, reservation_id = $2, guest_id = $3, status = $4, balance = $5, currency_code = $6, updated_at = $7, version = version + 1
		WHERE id = $8 AND tenant_id = $9 AND version = $10 AND deleted_at IS NULL
	`, f.PropertyID, f.ReservationID, f.GuestID, f.Status, f.Balance, f.Currency, f.UpdatedAt, f.ID, f.TenantID, f.Version)
	if err != nil {
		r.metrics.IncError("folio", "update", "query")
		return fmt.Errorf("update folio: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		r.metrics.IncError("folio", "update", "optimistic_lock")
		return fmt.Errorf("folio update conflict: id=%s version=%d", f.ID, f.Version)
	}
	f.Version++
	return nil
}

// AddCharge appends a charge to a folio.
func (r *PostgresFolioRepository) AddCharge(ctx context.Context, tenantID, folioID string, charge domain.Charge) error {
	start := time.Now()
	r.metrics.IncQuery("folio", "add_charge")
	defer r.metrics.ObserveDuration("folio", "add_charge", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("folio", "add_charge", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO folio_transactions (id, tenant_id, folio_id, transaction_type, description, amount, currency_code, posting_date, posted_at)
		VALUES ($1, $2, $3, 'charge', $4, $5, $6, $7, $8)
	`, charge.ID, tenantID, folioID, charge.Description, charge.Amount, "USD", charge.PostedAt.Format("2006-01-02"), charge.PostedAt)
	if err != nil {
		r.metrics.IncError("folio", "add_charge", "query")
		return fmt.Errorf("add charge: %w", err)
	}
	return nil
}

// AddPayment appends a payment to a folio.
func (r *PostgresFolioRepository) AddPayment(ctx context.Context, tenantID, folioID string, payment domain.Payment) error {
	start := time.Now()
	r.metrics.IncQuery("folio", "add_payment")
	defer r.metrics.ObserveDuration("folio", "add_payment", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("folio", "add_payment", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO folio_transactions (id, tenant_id, folio_id, transaction_type, description, amount, currency_code, posting_date, posted_at)
		VALUES ($1, $2, $3, 'payment', $4, $5, $6, $7, $8)
	`, payment.ID, tenantID, folioID, payment.Method, payment.Amount, "USD", payment.PostedAt.Format("2006-01-02"), payment.PostedAt)
	if err != nil {
		r.metrics.IncError("folio", "add_payment", "query")
		return fmt.Errorf("add payment: %w", err)
	}
	return nil
}

// SoftDelete marks a folio as deleted.
func (r *PostgresFolioRepository) SoftDelete(ctx context.Context, tenantID, id string) error {
	start := time.Now()
	r.metrics.IncQuery("folio", "soft_delete")
	defer r.metrics.ObserveDuration("folio", "soft_delete", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("folio", "soft_delete", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE folios SET deleted_at = NOW() WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID)
	if err != nil {
		r.metrics.IncError("folio", "soft_delete", "query")
		return fmt.Errorf("soft delete folio: %w", err)
	}
	return nil
}

func (r *PostgresFolioRepository) loadTransactions(ctx context.Context, tenantID, folioID string) ([]domain.Charge, []domain.Payment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, transaction_type, description, amount, posted_at
		FROM folio_transactions
		WHERE folio_id = $1 AND tenant_id = $2 AND is_voided = FALSE
		ORDER BY posted_at DESC
	`, folioID, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("query transactions: %w", err)
	}
	defer rows.Close()

	var charges []domain.Charge
	var payments []domain.Payment
	for rows.Next() {
		var id, txType, description string
		var amount float64
		var postedAt time.Time
		if err := rows.Scan(&id, &txType, &description, &amount, &postedAt); err != nil {
			return nil, nil, fmt.Errorf("scan transaction: %w", err)
		}
		if txType == "charge" {
			charges = append(charges, domain.Charge{ID: id, Description: description, Amount: amount, PostedAt: postedAt})
		} else if txType == "payment" {
			payments = append(payments, domain.Payment{ID: id, Method: description, Amount: amount, PostedAt: postedAt})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("rows error: %w", err)
	}
	return charges, payments, nil
}
