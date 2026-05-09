// Package repository provides tenant-aware data access for rate management.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// RatePlanRepository persists and retrieves RatePlan aggregates.
type RatePlanRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.RatePlan, error)
	Create(ctx context.Context, rp *domain.RatePlan) error
	Update(ctx context.Context, rp *domain.RatePlan) error
	SoftDelete(ctx context.Context, tenantID, id string) error
	List(ctx context.Context, tenantID, propertyID, roomTypeID string, limit, offset int) ([]*domain.RatePlan, error)
}

// PostgresRatePlanRepository is the PostgreSQL implementation.
type PostgresRatePlanRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewRatePlanRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresRatePlanRepository {
	return &PostgresRatePlanRepository{pool: pool, metrics: metrics}
}

func (r *PostgresRatePlanRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresRatePlanRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.RatePlan, error) {
	start := time.Now()
	r.metrics.IncQuery("rate_plan", "get_by_id")
	defer r.metrics.ObserveDuration("rate_plan", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, room_type_id, code, name, description,
		       base_rate, currency_code, min_los, max_los, advance_booking_days,
		       is_active, created_at, updated_at, version
		FROM rate_plans
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID)

	var rp domain.RatePlan
	var rateDate time.Time
	err := row.Scan(
		&rp.ID, &rp.TenantID, &rp.PropertyID, &rp.RoomTypeID,
		&rp.Code, &rp.Name, &rp.Description,
		&rp.BaseRate, &rp.CurrencyCode, &rp.MinLOS, &rp.MaxLOS, &rp.AdvanceBookingDays,
		&rp.IsActive, &rp.CreatedAt, &rp.UpdatedAt, &rp.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("rate plan not found: %w", err)
	}
	_ = rateDate
	return &rp, nil
}

func (r *PostgresRatePlanRepository) Create(ctx context.Context, rp *domain.RatePlan) error {
	start := time.Now()
	r.metrics.IncQuery("rate_plan", "create")
	defer r.metrics.ObserveDuration("rate_plan", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, rp.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO rate_plans (id, tenant_id, property_id, room_type_id, code, name, description,
		                       base_rate, currency_code, min_los, max_los, advance_booking_days,
		                       is_active, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, rp.ID, rp.TenantID, rp.PropertyID, rp.RoomTypeID, rp.Code, rp.Name, rp.Description,
		rp.BaseRate, rp.CurrencyCode, rp.MinLOS, rp.MaxLOS, rp.AdvanceBookingDays,
		rp.IsActive, rp.Version, rp.CreatedAt, rp.UpdatedAt)
	return err
}

func (r *PostgresRatePlanRepository) Update(ctx context.Context, rp *domain.RatePlan) error {
	start := time.Now()
	r.metrics.IncQuery("rate_plan", "update")
	defer r.metrics.ObserveDuration("rate_plan", "update", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, rp.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE rate_plans
		SET name = $1, description = $2, base_rate = $3, currency_code = $4,
		    min_los = $5, max_los = $6, advance_booking_days = $7,
		    is_active = $8, version = $9, updated_at = $10
		WHERE id = $11 AND tenant_id = $12 AND version = $13 AND deleted_at IS NULL
	`, rp.Name, rp.Description, rp.BaseRate, rp.CurrencyCode,
		rp.MinLOS, rp.MaxLOS, rp.AdvanceBookingDays,
		rp.IsActive, rp.Version, rp.UpdatedAt,
		rp.ID, rp.TenantID, rp.Version-1)
	return err
}

func (r *PostgresRatePlanRepository) SoftDelete(ctx context.Context, tenantID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE rate_plans SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID)
	return err
}

func (r *PostgresRatePlanRepository) List(ctx context.Context, tenantID, propertyID, roomTypeID string, limit, offset int) ([]*domain.RatePlan, error) {
	start := time.Now()
	r.metrics.IncQuery("rate_plan", "list")
	defer r.metrics.ObserveDuration("rate_plan", "list", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	query := `
		SELECT id, tenant_id, property_id, room_type_id, code, name, description,
		       base_rate, currency_code, min_los, max_los, advance_booking_days,
		       is_active, created_at, updated_at, version
		FROM rate_plans
		WHERE tenant_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{tenantID}
	argCount := 1

	if propertyID != "" {
		argCount++
		query += fmt.Sprintf(" AND property_id = $%d", argCount)
		args = append(args, propertyID)
	}
	if roomTypeID != "" {
		argCount++
		query += fmt.Sprintf(" AND room_type_id = $%d", argCount)
		args = append(args, roomTypeID)
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount+1, argCount+2)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plans []*domain.RatePlan
	for rows.Next() {
		var rp domain.RatePlan
		err := rows.Scan(
			&rp.ID, &rp.TenantID, &rp.PropertyID, &rp.RoomTypeID,
			&rp.Code, &rp.Name, &rp.Description,
			&rp.BaseRate, &rp.CurrencyCode, &rp.MinLOS, &rp.MaxLOS, &rp.AdvanceBookingDays,
			&rp.IsActive, &rp.CreatedAt, &rp.UpdatedAt, &rp.Version,
		)
		if err != nil {
			return nil, err
		}
		plans = append(plans, &rp)
	}
	return plans, rows.Err()
}

// ==================== DAILY RATE ====================

// DailyRateRepository persists and retrieves DailyRate aggregates.
type DailyRateRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.DailyRate, error)
	Create(ctx context.Context, dr *domain.DailyRate) error
	Update(ctx context.Context, dr *domain.DailyRate) error
	Delete(ctx context.Context, tenantID, id string) error
	ListByPlanAndDateRange(ctx context.Context, tenantID, ratePlanID string, start, end time.Time) ([]*domain.DailyRate, error)
}

// PostgresDailyRateRepository is the PostgreSQL implementation.
type PostgresDailyRateRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewDailyRateRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresDailyRateRepository {
	return &PostgresDailyRateRepository{pool: pool, metrics: metrics}
}

func (r *PostgresDailyRateRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresDailyRateRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.DailyRate, error) {
	start := time.Now()
	r.metrics.IncQuery("daily_rate", "get_by_id")
	defer r.metrics.ObserveDuration("daily_rate", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, rate_plan_id, rate_date, rate, availability,
		       min_stay, close_to_arrival, close_to_departure, stop_sell, created_at, updated_at
		FROM daily_rates
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)

	var dr domain.DailyRate
	err := row.Scan(
		&dr.ID, &dr.TenantID, &dr.RatePlanID, &dr.RateDate, &dr.Rate, &dr.Availability,
		&dr.MinStay, &dr.CloseToArrival, &dr.CloseToDeparture, &dr.StopSell,
		&dr.CreatedAt, &dr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("daily rate not found: %w", err)
	}
	return &dr, nil
}

func (r *PostgresDailyRateRepository) Create(ctx context.Context, dr *domain.DailyRate) error {
	start := time.Now()
	r.metrics.IncQuery("daily_rate", "create")
	defer r.metrics.ObserveDuration("daily_rate", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, dr.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO daily_rates (id, tenant_id, rate_plan_id, rate_date, rate, availability,
		                         min_stay, close_to_arrival, close_to_departure, stop_sell,
		                         created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (rate_plan_id, rate_date) DO UPDATE SET
			rate = EXCLUDED.rate,
			availability = EXCLUDED.availability,
			min_stay = EXCLUDED.min_stay,
			close_to_arrival = EXCLUDED.close_to_arrival,
			close_to_departure = EXCLUDED.close_to_departure,
			stop_sell = EXCLUDED.stop_sell,
			updated_at = NOW()
	`, dr.ID, dr.TenantID, dr.RatePlanID, dr.RateDate, dr.Rate, dr.Availability,
		dr.MinStay, dr.CloseToArrival, dr.CloseToDeparture, dr.StopSell,
		dr.CreatedAt, dr.UpdatedAt)
	return err
}

func (r *PostgresDailyRateRepository) Update(ctx context.Context, dr *domain.DailyRate) error {
	start := time.Now()
	r.metrics.IncQuery("daily_rate", "update")
	defer r.metrics.ObserveDuration("daily_rate", "update", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, dr.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE daily_rates
		SET rate = $1, availability = $2, min_stay = $3,
		    close_to_arrival = $4, close_to_departure = $5, stop_sell = $6, updated_at = $7
		WHERE id = $8 AND tenant_id = $9
	`, dr.Rate, dr.Availability, dr.MinStay,
		dr.CloseToArrival, dr.CloseToDeparture, dr.StopSell, dr.UpdatedAt,
		dr.ID, dr.TenantID)
	return err
}

func (r *PostgresDailyRateRepository) Delete(ctx context.Context, tenantID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `DELETE FROM daily_rates WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

func (r *PostgresDailyRateRepository) ListByPlanAndDateRange(ctx context.Context, tenantID, ratePlanID string, start, end time.Time) ([]*domain.DailyRate, error) {
	startT := time.Now()
	r.metrics.IncQuery("daily_rate", "list_by_plan_and_date")
	defer r.metrics.ObserveDuration("daily_rate", "list_by_plan_and_date", time.Since(startT).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	var rows pgx.Rows
	var err error

	if !start.IsZero() && !end.IsZero() {
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, rate_plan_id, rate_date, rate, availability,
			       min_stay, close_to_arrival, close_to_departure, stop_sell, created_at, updated_at
			FROM daily_rates
			WHERE tenant_id = $1 AND rate_plan_id = $2 AND rate_date BETWEEN $3 AND $4
			ORDER BY rate_date ASC
		`, tenantID, ratePlanID, start, end)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, rate_plan_id, rate_date, rate, availability,
			       min_stay, close_to_arrival, close_to_departure, stop_sell, created_at, updated_at
			FROM daily_rates
			WHERE tenant_id = $1 AND rate_plan_id = $2
			ORDER BY rate_date ASC
		`, tenantID, ratePlanID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []*domain.DailyRate
	for rows.Next() {
		var dr domain.DailyRate
		err := rows.Scan(
			&dr.ID, &dr.TenantID, &dr.RatePlanID, &dr.RateDate, &dr.Rate, &dr.Availability,
			&dr.MinStay, &dr.CloseToArrival, &dr.CloseToDeparture, &dr.StopSell,
			&dr.CreatedAt, &dr.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		rates = append(rates, &dr)
	}
	return rates, rows.Err()
}
