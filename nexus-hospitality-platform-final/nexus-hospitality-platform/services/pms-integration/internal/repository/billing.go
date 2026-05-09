// Package repository provides tenant-aware data access for billing.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// SubscriptionPlanRepository persists subscription plans.
type SubscriptionPlanRepository interface {
	GetByID(ctx context.Context, id string) (*domain.SubscriptionPlan, error)
	GetBySlug(ctx context.Context, slug string) (*domain.SubscriptionPlan, error)
	GetDefault(ctx context.Context) (*domain.SubscriptionPlan, error)
	Create(ctx context.Context, p *domain.SubscriptionPlan) error
	ListActive(ctx context.Context) ([]*domain.SubscriptionPlan, error)
}

type PostgresSubscriptionPlanRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewSubscriptionPlanRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresSubscriptionPlanRepository {
	return &PostgresSubscriptionPlanRepository{pool: pool, metrics: metrics}
}

func (r *PostgresSubscriptionPlanRepository) GetByID(ctx context.Context, id string) (*domain.SubscriptionPlan, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, description, price_cents, currency_code, interval,
		       max_properties, max_rooms, max_users, features,
		       stripe_price_id, stripe_product_id, is_active, is_default, trial_days,
		       created_at, updated_at
		FROM subscription_plans WHERE id = $1
	`, id)
	return scanPlan(row)
}

func (r *PostgresSubscriptionPlanRepository) GetBySlug(ctx context.Context, slug string) (*domain.SubscriptionPlan, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, description, price_cents, currency_code, interval,
		       max_properties, max_rooms, max_users, features,
		       stripe_price_id, stripe_product_id, is_active, is_default, trial_days,
		       created_at, updated_at
		FROM subscription_plans WHERE slug = $1
	`, slug)
	return scanPlan(row)
}

func (r *PostgresSubscriptionPlanRepository) GetDefault(ctx context.Context) (*domain.SubscriptionPlan, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, description, price_cents, currency_code, interval,
		       max_properties, max_rooms, max_users, features,
		       stripe_price_id, stripe_product_id, is_active, is_default, trial_days,
		       created_at, updated_at
		FROM subscription_plans WHERE is_default = TRUE AND is_active = TRUE LIMIT 1
	`)
	return scanPlan(row)
}

func (r *PostgresSubscriptionPlanRepository) Create(ctx context.Context, p *domain.SubscriptionPlan) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO subscription_plans (id, slug, name, description, price_cents, currency_code, interval,
		                               max_properties, max_rooms, max_users, features,
		                               stripe_price_id, stripe_product_id, is_active, is_default, trial_days,
		                               created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`, p.ID, p.Slug, p.Name, p.Description, p.PriceCents, p.CurrencyCode, p.Interval,
		p.MaxProperties, p.MaxRooms, p.MaxUsers, p.Features,
		p.StripePriceID, p.StripeProductID, p.IsActive, p.IsDefault, p.TrialDays,
		p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *PostgresSubscriptionPlanRepository) ListActive(ctx context.Context) ([]*domain.SubscriptionPlan, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, description, price_cents, currency_code, interval,
		       max_properties, max_rooms, max_users, features,
		       stripe_price_id, stripe_product_id, is_active, is_default, trial_days,
		       created_at, updated_at
		FROM subscription_plans WHERE is_active = TRUE ORDER BY price_cents ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPlans(rows)
}

func scanPlan(row pgx.Row) (*domain.SubscriptionPlan, error) {
	var p domain.SubscriptionPlan
	err := row.Scan(
		&p.ID, &p.Slug, &p.Name, &p.Description, &p.PriceCents, &p.CurrencyCode, &p.Interval,
		&p.MaxProperties, &p.MaxRooms, &p.MaxUsers, &p.Features,
		&p.StripePriceID, &p.StripeProductID, &p.IsActive, &p.IsDefault, &p.TrialDays,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("plan not found: %w", err)
	}
	return &p, nil
}

func scanPlans(rows pgx.Rows) ([]*domain.SubscriptionPlan, error) {
	var plans []*domain.SubscriptionPlan
	for rows.Next() {
		var p domain.SubscriptionPlan
		err := rows.Scan(
			&p.ID, &p.Slug, &p.Name, &p.Description, &p.PriceCents, &p.CurrencyCode, &p.Interval,
			&p.MaxProperties, &p.MaxRooms, &p.MaxUsers, &p.Features,
			&p.StripePriceID, &p.StripeProductID, &p.IsActive, &p.IsDefault, &p.TrialDays,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, &p)
	}
	return plans, rows.Err()
}

// ==================== SUBSCRIPTION REPOSITORY ====================

type SubscriptionRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Subscription, error)
	GetByTenant(ctx context.Context, tenantID string) (*domain.Subscription, error)
	Create(ctx context.Context, s *domain.Subscription) error
	Update(ctx context.Context, s *domain.Subscription) error
}

type PostgresSubscriptionRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewSubscriptionRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresSubscriptionRepository {
	return &PostgresSubscriptionRepository{pool: pool, metrics: metrics}
}

func (r *PostgresSubscriptionRepository) GetByID(ctx context.Context, id string) (*domain.Subscription, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, plan_id, status, stripe_subscription_id, stripe_customer_id,
		       current_period_start, current_period_end, trial_start, trial_end,
		       canceled_at, cancel_at_period_end, created_at, updated_at
		FROM subscriptions WHERE id = $1
	`, id)
	return scanSubscription(row)
}

func (r *PostgresSubscriptionRepository) GetByTenant(ctx context.Context, tenantID string) (*domain.Subscription, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, plan_id, status, stripe_subscription_id, stripe_customer_id,
		       current_period_start, current_period_end, trial_start, trial_end,
		       canceled_at, cancel_at_period_end, created_at, updated_at
		FROM subscriptions WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT 1
	`, tenantID)
	return scanSubscription(row)
}

func (r *PostgresSubscriptionRepository) Create(ctx context.Context, s *domain.Subscription) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO subscriptions (id, tenant_id, plan_id, status, stripe_subscription_id, stripe_customer_id,
		                         current_period_start, current_period_end, trial_start, trial_end,
		                         canceled_at, cancel_at_period_end, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, s.ID, s.TenantID, s.PlanID, s.Status, s.StripeSubscriptionID, s.StripeCustomerID,
		s.CurrentPeriodStart, s.CurrentPeriodEnd, s.TrialStart, s.TrialEnd,
		s.CanceledAt, s.CancelAtPeriodEnd, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *PostgresSubscriptionRepository) Update(ctx context.Context, s *domain.Subscription) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE subscriptions
		SET plan_id = $1, status = $2, stripe_subscription_id = $3, stripe_customer_id = $4,
		    current_period_start = $5, current_period_end = $6, trial_start = $7, trial_end = $8,
		    canceled_at = $9, cancel_at_period_end = $10, updated_at = $11
		WHERE id = $12
	`, s.PlanID, s.Status, s.StripeSubscriptionID, s.StripeCustomerID,
		s.CurrentPeriodStart, s.CurrentPeriodEnd, s.TrialStart, s.TrialEnd,
		s.CanceledAt, s.CancelAtPeriodEnd, s.UpdatedAt, s.ID)
	return err
}

func scanSubscription(row pgx.Row) (*domain.Subscription, error) {
	var s domain.Subscription
	var trialStart, trialEnd, canceledAt *time.Time
	err := row.Scan(
		&s.ID, &s.TenantID, &s.PlanID, &s.Status, &s.StripeSubscriptionID, &s.StripeCustomerID,
		&s.CurrentPeriodStart, &s.CurrentPeriodEnd, &trialStart, &trialEnd,
		&canceledAt, &s.CancelAtPeriodEnd, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("subscription not found: %w", err)
	}
	s.TrialStart = trialStart
	s.TrialEnd = trialEnd
	s.CanceledAt = canceledAt
	return &s, nil
}

// ==================== INVOICE REPOSITORY ====================

type InvoiceRepository interface {
	Create(ctx context.Context, i *domain.Invoice) error
	GetByStripeID(ctx context.Context, stripeID string) (*domain.Invoice, error)
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Invoice, error)
	Update(ctx context.Context, i *domain.Invoice) error
}

type PostgresInvoiceRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewInvoiceRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresInvoiceRepository {
	return &PostgresInvoiceRepository{pool: pool, metrics: metrics}
}

func (r *PostgresInvoiceRepository) Create(ctx context.Context, i *domain.Invoice) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO invoices (id, tenant_id, subscription_id, stripe_invoice_id, status,
		                    amount_due_cents, amount_paid_cents, currency_code,
		                    period_start, period_end, due_date, paid_at, pdf_url, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, i.ID, i.TenantID, i.SubscriptionID, i.StripeInvoiceID, i.Status,
		i.AmountDueCents, i.AmountPaidCents, i.CurrencyCode,
		i.PeriodStart, i.PeriodEnd, i.DueDate, i.PaidAt, i.PdfUrl, i.CreatedAt)
	return err
}

func (r *PostgresInvoiceRepository) GetByStripeID(ctx context.Context, stripeID string) (*domain.Invoice, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, subscription_id, stripe_invoice_id, status,
		       amount_due_cents, amount_paid_cents, currency_code,
		       period_start, period_end, due_date, paid_at, pdf_url, created_at
		FROM invoices WHERE stripe_invoice_id = $1
	`, stripeID)
	return scanInvoice(row)
}

func (r *PostgresInvoiceRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Invoice, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, subscription_id, stripe_invoice_id, status,
		       amount_due_cents, amount_paid_cents, currency_code,
		       period_start, period_end, due_date, paid_at, pdf_url, created_at
		FROM invoices WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanInvoices(rows)
}

func (r *PostgresInvoiceRepository) Update(ctx context.Context, i *domain.Invoice) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE invoices
		SET status = $1, amount_due_cents = $2, amount_paid_cents = $3, paid_at = $4, pdf_url = $5
		WHERE id = $6
	`, i.Status, i.AmountDueCents, i.AmountPaidCents, i.PaidAt, i.PdfUrl, i.ID)
	return err
}

func scanInvoice(row pgx.Row) (*domain.Invoice, error) {
	var i domain.Invoice
	var dueDate, paidAt *time.Time
	err := row.Scan(
		&i.ID, &i.TenantID, &i.SubscriptionID, &i.StripeInvoiceID, &i.Status,
		&i.AmountDueCents, &i.AmountPaidCents, &i.CurrencyCode,
		&i.PeriodStart, &i.PeriodEnd, &dueDate, &paidAt, &i.PdfUrl, &i.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("invoice not found: %w", err)
	}
	i.DueDate = dueDate
	i.PaidAt = paidAt
	return &i, nil
}

func scanInvoices(rows pgx.Rows) ([]*domain.Invoice, error) {
	var invoices []*domain.Invoice
	for rows.Next() {
		var i domain.Invoice
		var dueDate, paidAt *time.Time
		err := rows.Scan(
			&i.ID, &i.TenantID, &i.SubscriptionID, &i.StripeInvoiceID, &i.Status,
			&i.AmountDueCents, &i.AmountPaidCents, &i.CurrencyCode,
			&i.PeriodStart, &i.PeriodEnd, &dueDate, &paidAt, &i.PdfUrl, &i.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		i.DueDate = dueDate
		i.PaidAt = paidAt
		invoices = append(invoices, &i)
	}
	return invoices, rows.Err()
}
