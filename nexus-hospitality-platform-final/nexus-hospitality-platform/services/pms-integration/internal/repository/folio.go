// Package repository provides tenant-aware data access for financial operations.
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
	RecalculateBalance(ctx context.Context, folioID string) error
	ListByTenant(ctx context.Context, tenantID string, status string, limit, offset int) ([]*domain.Folio, error)
	CloseFolio(ctx context.Context, tenantID, id string) error
}

// PostgresFolioRepository is the PostgreSQL implementation.
type PostgresFolioRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewFolioRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresFolioRepository {
	return &PostgresFolioRepository{pool: pool, metrics: metrics}
}

func (r *PostgresFolioRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresFolioRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Folio, error) {
	start := time.Now()
	r.metrics.IncQuery("folio", "get_by_id")
	defer r.metrics.ObserveDuration("folio", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, reservation_id, guest_id, status, balance,
		       total_charges, total_payments, total_adjustments, currency_code,
		       is_master, master_folio_id, notes, created_at, updated_at, version
		FROM folios
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID)

	var f domain.Folio
	var masterFolioID *string
	err := row.Scan(
		&f.ID, &f.TenantID, &f.ReservationID, &f.GuestID, &f.Status, &f.Balance,
		&f.TotalCharges, &f.TotalPayments, &f.TotalAdjustments, &f.CurrencyCode,
		&f.IsMaster, &masterFolioID, &f.Notes, &f.CreatedAt, &f.UpdatedAt, &f.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("folio not found: %w", err)
	}
	f.MasterFolioID = masterFolioID
	return &f, nil
}

func (r *PostgresFolioRepository) GetByReservation(ctx context.Context, tenantID, reservationID string) (*domain.Folio, error) {
	start := time.Now()
	r.metrics.IncQuery("folio", "get_by_reservation")
	defer r.metrics.ObserveDuration("folio", "get_by_reservation", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, reservation_id, guest_id, status, balance,
		       total_charges, total_payments, total_adjustments, currency_code,
		       is_master, master_folio_id, notes, created_at, updated_at, version
		FROM folios
		WHERE reservation_id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, reservationID, tenantID)

	var f domain.Folio
	var masterFolioID *string
	err := row.Scan(
		&f.ID, &f.TenantID, &f.ReservationID, &f.GuestID, &f.Status, &f.Balance,
		&f.TotalCharges, &f.TotalPayments, &f.TotalAdjustments, &f.CurrencyCode,
		&f.IsMaster, &masterFolioID, &f.Notes, &f.CreatedAt, &f.UpdatedAt, &f.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("folio not found: %w", err)
	}
	f.MasterFolioID = masterFolioID
	return &f, nil
}

func (r *PostgresFolioRepository) Create(ctx context.Context, f *domain.Folio) error {
	start := time.Now()
	r.metrics.IncQuery("folio", "create")
	defer r.metrics.ObserveDuration("folio", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, f.TenantID); err != nil {
		return err
	}

	var masterFolioID interface{}
	if f.MasterFolioID != nil {
		masterFolioID = *f.MasterFolioID
	} else {
		masterFolioID = nil
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO folios (id, tenant_id, reservation_id, guest_id, status, balance,
		                   total_charges, total_payments, total_adjustments, currency_code,
		                   is_master, master_folio_id, notes, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, f.ID, f.TenantID, f.ReservationID, f.GuestID, f.Status, f.Balance,
		f.TotalCharges, f.TotalPayments, f.TotalAdjustments, f.CurrencyCode,
		f.IsMaster, masterFolioID, f.Notes, f.Version, f.CreatedAt, f.UpdatedAt)
	return err
}

func (r *PostgresFolioRepository) Update(ctx context.Context, f *domain.Folio) error {
	start := time.Now()
	r.metrics.IncQuery("folio", "update")
	defer r.metrics.ObserveDuration("folio", "update", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, f.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE folios
		SET status = $1, balance = $2, total_charges = $3, total_payments = $4,
		    total_adjustments = $5, notes = $6, version = $7, updated_at = $8
		WHERE id = $9 AND tenant_id = $10 AND version = $11 AND deleted_at IS NULL
	`, f.Status, f.Balance, f.TotalCharges, f.TotalPayments,
		f.TotalAdjustments, f.Notes, f.Version, f.UpdatedAt,
		f.ID, f.TenantID, f.Version-1)
	return err
}

func (r *PostgresFolioRepository) RecalculateBalance(ctx context.Context, folioID string) error {
	_, err := r.pool.Exec(ctx, `SELECT recalculate_folio_balance($1)`, folioID)
	return err
}

func (r *PostgresFolioRepository) ListByTenant(ctx context.Context, tenantID, status string, limit, offset int) ([]*domain.Folio, error) {
	start := time.Now()
	r.metrics.IncQuery("folio", "list")
	defer r.metrics.ObserveDuration("folio", "list", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	var rows pgx.Rows
	var err error
	if status != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, reservation_id, guest_id, status, balance,
			       total_charges, total_payments, total_adjustments, currency_code,
			       is_master, master_folio_id, notes, created_at, updated_at, version
			FROM folios
			WHERE tenant_id = $1 AND status = $2 AND deleted_at IS NULL
			ORDER BY created_at DESC LIMIT $3 OFFSET $4
		`, tenantID, status, limit, offset)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, reservation_id, guest_id, status, balance,
			       total_charges, total_payments, total_adjustments, currency_code,
			       is_master, master_folio_id, notes, created_at, updated_at, version
			FROM folios
			WHERE tenant_id = $1 AND deleted_at IS NULL
			ORDER BY created_at DESC LIMIT $2 OFFSET $3
		`, tenantID, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folios []*domain.Folio
	for rows.Next() {
		var f domain.Folio
		var masterFolioID *string
		err := rows.Scan(
			&f.ID, &f.TenantID, &f.ReservationID, &f.GuestID, &f.Status, &f.Balance,
			&f.TotalCharges, &f.TotalPayments, &f.TotalAdjustments, &f.CurrencyCode,
			&f.IsMaster, &masterFolioID, &f.Notes, &f.CreatedAt, &f.UpdatedAt, &f.Version,
		)
		if err != nil {
			return nil, err
		}
		f.MasterFolioID = masterFolioID
		folios = append(folios, &f)
	}
	return folios, rows.Err()
}

func (r *PostgresFolioRepository) CloseFolio(ctx context.Context, tenantID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE folios
		SET status = 'closed', updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2 AND status = 'open' AND balance = 0
	`, id, tenantID)
	return err
}

// ==================== CHARGE REPOSITORY ====================

type ChargeRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.Charge, error)
	Create(ctx context.Context, c *domain.Charge) error
	ListByFolio(ctx context.Context, tenantID, folioID string, limit, offset int) ([]*domain.Charge, error)
	VoidCharge(ctx context.Context, tenantID, id, voidedBy, reason string) error
}

type PostgresChargeRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewChargeRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresChargeRepository {
	return &PostgresChargeRepository{pool: pool, metrics: metrics}
}

func (r *PostgresChargeRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresChargeRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Charge, error) {
	start := time.Now()
	r.metrics.IncQuery("charge", "get_by_id")
	defer r.metrics.ObserveDuration("charge", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, folio_id, reservation_id, room_id, charge_type, category,
		       description, quantity, unit_price, total_amount, tax_amount, tax_rate,
		       currency_code, charge_date, posting_date, posted_by, is_revenue,
		       revenue_center, source, is_voided, voided_at, voided_by, void_reason, metadata,
		       created_at, updated_at
		FROM charges
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)

	var c domain.Charge
	var resID, roomID, voidedAt *string
	err := row.Scan(
		&c.ID, &c.TenantID, &c.FolioID, &resID, &roomID, &c.ChargeType, &c.Category,
		&c.Description, &c.Quantity, &c.UnitPrice, &c.TotalAmount, &c.TaxAmount, &c.TaxRate,
		&c.CurrencyCode, &c.ChargeDate, &c.PostingDate, &c.PostedBy, &c.IsRevenue,
		&c.RevenueCenter, &c.Source, &c.IsVoided, &voidedAt, &c.VoidedBy, &c.VoidReason, &c.Metadata,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("charge not found: %w", err)
	}
	c.ReservationID = resID
	c.RoomID = roomID
	if voidedAt != nil {
		t, _ := time.Parse(time.RFC3339, *voidedAt)
		c.VoidedAt = &t
	}
	return &c, nil
}

func (r *PostgresChargeRepository) Create(ctx context.Context, c *domain.Charge) error {
	start := time.Now()
	r.metrics.IncQuery("charge", "create")
	defer r.metrics.ObserveDuration("charge", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, c.TenantID); err != nil {
		return err
	}

	var resID, roomID interface{}
	if c.ReservationID != nil { resID = *c.ReservationID }
	if c.RoomID != nil { roomID = *c.RoomID }

	_, err := r.pool.Exec(ctx, `
		INSERT INTO charges (id, tenant_id, folio_id, reservation_id, room_id, charge_type, category,
		                    description, quantity, unit_price, total_amount, tax_amount, tax_rate,
		                    currency_code, charge_date, posting_date, posted_by, is_revenue,
		                    revenue_center, source, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
	`, c.ID, c.TenantID, c.FolioID, resID, roomID, c.ChargeType, c.Category,
		c.Description, c.Quantity, c.UnitPrice, c.TotalAmount, c.TaxAmount, c.TaxRate,
		c.CurrencyCode, c.ChargeDate, c.PostingDate, c.PostedBy, c.IsRevenue,
		c.RevenueCenter, c.Source, c.Metadata, c.CreatedAt, c.UpdatedAt)
	return err
}

func (r *PostgresChargeRepository) ListByFolio(ctx context.Context, tenantID, folioID string, limit, offset int) ([]*domain.Charge, error) {
	start := time.Now()
	r.metrics.IncQuery("charge", "list_by_folio")
	defer r.metrics.ObserveDuration("charge", "list_by_folio", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, folio_id, reservation_id, room_id, charge_type, category,
		       description, quantity, unit_price, total_amount, tax_amount, tax_rate,
		       currency_code, charge_date, posting_date, posted_by, is_revenue,
		       revenue_center, source, is_voided, voided_at, voided_by, void_reason, metadata,
		       created_at, updated_at
		FROM charges
		WHERE folio_id = $1 AND tenant_id = $2
		ORDER BY created_at DESC LIMIT $3 OFFSET $4
	`, folioID, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var charges []*domain.Charge
	for rows.Next() {
		var c domain.Charge
		var resID, roomID, voidedAt *string
		err := rows.Scan(
			&c.ID, &c.TenantID, &c.FolioID, &resID, &roomID, &c.ChargeType, &c.Category,
			&c.Description, &c.Quantity, &c.UnitPrice, &c.TotalAmount, &c.TaxAmount, &c.TaxRate,
			&c.CurrencyCode, &c.ChargeDate, &c.PostingDate, &c.PostedBy, &c.IsRevenue,
			&c.RevenueCenter, &c.Source, &c.IsVoided, &voidedAt, &c.VoidedBy, &c.VoidReason, &c.Metadata,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		c.ReservationID = resID
		c.RoomID = roomID
		if voidedAt != nil {
			t, _ := time.Parse(time.RFC3339, *voidedAt)
			c.VoidedAt = &t
		}
		charges = append(charges, &c)
	}
	return charges, rows.Err()
}

func (r *PostgresChargeRepository) VoidCharge(ctx context.Context, tenantID, id, voidedBy, reason string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE charges
		SET is_voided = TRUE, voided_at = NOW(), voided_by = $1, void_reason = $2, updated_at = NOW()
		WHERE id = $3 AND tenant_id = $4 AND is_voided = FALSE
	`, voidedBy, reason, id, tenantID)
	return err
}

// ==================== PAYMENT REPOSITORY ====================

type PaymentRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.Payment, error)
	Create(ctx context.Context, p *domain.Payment) error
	ListByFolio(ctx context.Context, tenantID, folioID string, limit, offset int) ([]*domain.Payment, error)
	Refund(ctx context.Context, tenantID, id string, amount float64, reason string) error
}

type PostgresPaymentRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewPaymentRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{pool: pool, metrics: metrics}
}

func (r *PostgresPaymentRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresPaymentRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Payment, error) {
	start := time.Now()
	r.metrics.IncQuery("payment", "get_by_id")
	defer r.metrics.ObserveDuration("payment", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, folio_id, reservation_id, guest_id, payment_method,
		       payment_type, amount, currency_code, exchange_rate, reference_number,
		       transaction_id, status, processor, last_four_digits, card_brand,
		       expiry_month, expiry_year, authorization_code, captured_at,
		       refunded_at, refund_amount, refund_reason, processed_by, metadata,
		       created_at, updated_at
		FROM payments
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)

	var p domain.Payment
	var resID, guestID, capturedAt, refundedAt *string
	err := row.Scan(
		&p.ID, &p.TenantID, &p.FolioID, &resID, &guestID, &p.PaymentMethod,
		&p.PaymentType, &p.Amount, &p.CurrencyCode, &p.ExchangeRate, &p.ReferenceNumber,
		&p.TransactionID, &p.Status, &p.Processor, &p.LastFourDigits, &p.CardBrand,
		&p.ExpiryMonth, &p.ExpiryYear, &p.AuthorizationCode, &capturedAt,
		&refundedAt, &p.RefundAmount, &p.RefundReason, &p.ProcessedBy, &p.Metadata,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("payment not found: %w", err)
	}
	p.ReservationID = resID
	p.GuestID = guestID
	return &p, nil
}

func (r *PostgresPaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	start := time.Now()
	r.metrics.IncQuery("payment", "create")
	defer r.metrics.ObserveDuration("payment", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, p.TenantID); err != nil {
		return err
	}

	var resID, guestID interface{}
	if p.ReservationID != nil { resID = *p.ReservationID }
	if p.GuestID != nil { guestID = *p.GuestID }

	_, err := r.pool.Exec(ctx, `
		INSERT INTO payments (id, tenant_id, folio_id, reservation_id, guest_id, payment_method,
		                     payment_type, amount, currency_code, exchange_rate, reference_number,
		                     transaction_id, status, processor, last_four_digits, card_brand,
		                     expiry_month, expiry_year, authorization_code, processed_by,
		                     metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
	`, p.ID, p.TenantID, p.FolioID, resID, guestID, p.PaymentMethod,
		p.PaymentType, p.Amount, p.CurrencyCode, p.ExchangeRate, p.ReferenceNumber,
		p.TransactionID, p.Status, p.Processor, p.LastFourDigits, p.CardBrand,
		p.ExpiryMonth, p.ExpiryYear, p.AuthorizationCode, p.ProcessedBy,
		p.Metadata, p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *PostgresPaymentRepository) ListByFolio(ctx context.Context, tenantID, folioID string, limit, offset int) ([]*domain.Payment, error) {
	start := time.Now()
	r.metrics.IncQuery("payment", "list_by_folio")
	defer r.metrics.ObserveDuration("payment", "list_by_folio", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, folio_id, reservation_id, guest_id, payment_method,
		       payment_type, amount, currency_code, exchange_rate, reference_number,
		       transaction_id, status, processor, last_four_digits, card_brand,
		       expiry_month, expiry_year, authorization_code, captured_at,
		       refunded_at, refund_amount, refund_reason, processed_by, metadata,
		       created_at, updated_at
		FROM payments
		WHERE folio_id = $1 AND tenant_id = $2
		ORDER BY created_at DESC LIMIT $3 OFFSET $4
	`, folioID, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var p domain.Payment
		var resID, guestID, capturedAt, refundedAt *string
		err := rows.Scan(
			&p.ID, &p.TenantID, &p.FolioID, &resID, &guestID, &p.PaymentMethod,
			&p.PaymentType, &p.Amount, &p.CurrencyCode, &p.ExchangeRate, &p.ReferenceNumber,
			&p.TransactionID, &p.Status, &p.Processor, &p.LastFourDigits, &p.CardBrand,
			&p.ExpiryMonth, &p.ExpiryYear, &p.AuthorizationCode, &capturedAt,
			&refundedAt, &p.RefundAmount, &p.RefundReason, &p.ProcessedBy, &p.Metadata,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		p.ReservationID = resID
		p.GuestID = guestID
		payments = append(payments, &p)
	}
	return payments, rows.Err()
}

func (r *PostgresPaymentRepository) Refund(ctx context.Context, tenantID, id string, amount float64, reason string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE payments
		SET refunded_at = $1, refund_amount = $2, refund_reason = $3,
		    status = CASE WHEN refund_amount + $2 = amount THEN 'refunded' ELSE 'partial_refund' END,
		    updated_at = $1
		WHERE id = $4 AND tenant_id = $5 AND refunded_at IS NULL
	`, now, amount, reason, id, tenantID)
	return err
}
