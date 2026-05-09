// Package repository provides tenant-aware data access for channel manager.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ChannelIntegrationRepository persists OTA channel connections.
type ChannelIntegrationRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.ChannelIntegration, error)
	GetByProperty(ctx context.Context, tenantID, propertyID string) ([]*domain.ChannelIntegration, error)
	Create(ctx context.Context, ci *domain.ChannelIntegration) error
	Update(ctx context.Context, ci *domain.ChannelIntegration) error
	Delete(ctx context.Context, tenantID, id string) error
	ListByTenant(ctx context.Context, tenantID string) ([]*domain.ChannelIntegration, error)
}

type PostgresChannelIntegrationRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewChannelIntegrationRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresChannelIntegrationRepository {
	return &PostgresChannelIntegrationRepository{pool: pool, metrics: metrics}
}

func (r *PostgresChannelIntegrationRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresChannelIntegrationRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.ChannelIntegration, error) {
	start := time.Now()
	r.metrics.IncQuery("channel_integration", "get_by_id")
	defer r.metrics.ObserveDuration("channel_integration", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, channel_type, name, api_key, api_secret,
		       hotel_id, rate_plan_map, room_type_map, commission_pct,
		       is_active, status, last_sync_at, last_error, last_error_at, created_at, updated_at
		FROM channel_integrations
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)
	return scanChannelIntegration(row)
}

func (r *PostgresChannelIntegrationRepository) GetByProperty(ctx context.Context, tenantID, propertyID string) ([]*domain.ChannelIntegration, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, channel_type, name, api_key, api_secret,
		       hotel_id, rate_plan_map, room_type_map, commission_pct,
		       is_active, status, last_sync_at, last_error, last_error_at, created_at, updated_at
		FROM channel_integrations
		WHERE tenant_id = $1 AND property_id = $2 AND is_active = TRUE
	`, tenantID, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanChannelIntegrations(rows)
}

func (r *PostgresChannelIntegrationRepository) Create(ctx context.Context, ci *domain.ChannelIntegration) error {
	start := time.Now()
	r.metrics.IncQuery("channel_integration", "create")
	defer r.metrics.ObserveDuration("channel_integration", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, ci.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO channel_integrations (id, tenant_id, property_id, channel_type, name, api_key, api_secret,
		                                 hotel_id, rate_plan_map, room_type_map, commission_pct,
		                                 is_active, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, ci.ID, ci.TenantID, ci.PropertyID, ci.ChannelType, ci.Name, ci.APIKey, ci.APISecret,
		ci.HotelID, ci.RatePlanMap, ci.RoomTypeMap, ci.CommissionPct,
		ci.IsActive, ci.Status, ci.CreatedAt, ci.UpdatedAt)
	return err
}

func (r *PostgresChannelIntegrationRepository) Update(ctx context.Context, ci *domain.ChannelIntegration) error {
	start := time.Now()
	r.metrics.IncQuery("channel_integration", "update")
	defer r.metrics.ObserveDuration("channel_integration", "update", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, ci.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE channel_integrations
		SET property_id = $1, name = $2, api_key = $3, api_secret = $4,
		    hotel_id = $5, rate_plan_map = $6, room_type_map = $7, commission_pct = $8,
		    is_active = $9, status = $10, last_sync_at = $11, last_error = $12, last_error_at = $13,
		    updated_at = $14
		WHERE id = $15 AND tenant_id = $16
	`, ci.PropertyID, ci.Name, ci.APIKey, ci.APISecret,
		ci.HotelID, ci.RatePlanMap, ci.RoomTypeMap, ci.CommissionPct,
		ci.IsActive, ci.Status, ci.LastSyncAt, ci.LastError, ci.LastErrorAt,
		ci.UpdatedAt, ci.ID, ci.TenantID)
	return err
}

func (r *PostgresChannelIntegrationRepository) Delete(ctx context.Context, tenantID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		DELETE FROM channel_integrations WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)
	return err
}

func (r *PostgresChannelIntegrationRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.ChannelIntegration, error) {
	start := time.Now()
	r.metrics.IncQuery("channel_integration", "list")
	defer r.metrics.ObserveDuration("channel_integration", "list", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, channel_type, name, api_key, api_secret,
		       hotel_id, rate_plan_map, room_type_map, commission_pct,
		       is_active, status, last_sync_at, last_error, last_error_at, created_at, updated_at
		FROM channel_integrations
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanChannelIntegrations(rows)
}

func scanChannelIntegration(row pgx.Row) (*domain.ChannelIntegration, error) {
	var ci domain.ChannelIntegration
	var lastSync, lastErrorAt *time.Time
	err := row.Scan(
		&ci.ID, &ci.TenantID, &ci.PropertyID, &ci.ChannelType, &ci.Name, &ci.APIKey, &ci.APISecret,
		&ci.HotelID, &ci.RatePlanMap, &ci.RoomTypeMap, &ci.CommissionPct,
		&ci.IsActive, &ci.Status, &lastSync, &ci.LastError, &lastErrorAt, &ci.CreatedAt, &ci.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("channel integration not found: %w", err)
	}
	ci.LastSyncAt = lastSync
	ci.LastErrorAt = lastErrorAt
	return &ci, nil
}

func scanChannelIntegrations(rows pgx.Rows) ([]*domain.ChannelIntegration, error) {
	var integrations []*domain.ChannelIntegration
	for rows.Next() {
		var ci domain.ChannelIntegration
		var lastSync, lastErrorAt *time.Time
		err := rows.Scan(
			&ci.ID, &ci.TenantID, &ci.PropertyID, &ci.ChannelType, &ci.Name, &ci.APIKey, &ci.APISecret,
			&ci.HotelID, &ci.RatePlanMap, &ci.RoomTypeMap, &ci.CommissionPct,
			&ci.IsActive, &ci.Status, &lastSync, &ci.LastError, &lastErrorAt, &ci.CreatedAt, &ci.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		ci.LastSyncAt = lastSync
		ci.LastErrorAt = lastErrorAt
		integrations = append(integrations, &ci)
	}
	return integrations, rows.Err()
}

// ==================== CHANNEL RESERVATION REPOSITORY ====================

type ChannelReservationRepository interface {
	Create(ctx context.Context, cr *domain.ChannelReservation) error
	GetByOTAConf(ctx context.Context, tenantID, otaConf string) (*domain.ChannelReservation, error)
	ListPending(ctx context.Context, tenantID string) ([]*domain.ChannelReservation, error)
	UpdateStatus(ctx context.Context, tenantID, id, status string, reservationID *string) error
}

type PostgresChannelReservationRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewChannelReservationRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresChannelReservationRepository {
	return &PostgresChannelReservationRepository{pool: pool, metrics: metrics}
}

func (r *PostgresChannelReservationRepository) Create(ctx context.Context, cr *domain.ChannelReservation) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO channel_reservations (id, tenant_id, channel_integration_id, ota_confirmation_number,
		                                  reservation_id, status, guest_name, guest_email, guest_phone,
		                                  room_type_id, rate_plan_id, check_in_date, check_out_date,
		                                  adults, children, total_amount, currency_code, commission_amount, raw_data,
		                                  created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`, cr.ID, cr.TenantID, cr.ChannelIntegrationID, cr.OTAConfirmationNumber,
		cr.ReservationID, cr.Status, cr.GuestName, cr.GuestEmail, cr.GuestPhone,
		cr.RoomTypeID, cr.RatePlanID, cr.CheckInDate, cr.CheckOutDate,
		cr.Adults, cr.Children, cr.TotalAmount, cr.CurrencyCode, cr.CommissionAmount, cr.RawData,
		cr.CreatedAt, cr.UpdatedAt)
	return err
}

func (r *PostgresChannelReservationRepository) GetByOTAConf(ctx context.Context, tenantID, otaConf string) (*domain.ChannelReservation, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, channel_integration_id, ota_confirmation_number,
		       reservation_id, status, guest_name, guest_email, guest_phone,
		       room_type_id, rate_plan_id, check_in_date, check_out_date,
		       adults, children, total_amount, currency_code, commission_amount, raw_data,
		       created_at, updated_at
		FROM channel_reservations
		WHERE tenant_id = $1 AND ota_confirmation_number = $2
	`, tenantID, otaConf)

	var cr domain.ChannelReservation
	err := row.Scan(
		&cr.ID, &cr.TenantID, &cr.ChannelIntegrationID, &cr.OTAConfirmationNumber,
		&cr.ReservationID, &cr.Status, &cr.GuestName, &cr.GuestEmail, &cr.GuestPhone,
		&cr.RoomTypeID, &cr.RatePlanID, &cr.CheckInDate, &cr.CheckOutDate,
		&cr.Adults, &cr.Children, &cr.TotalAmount, &cr.CurrencyCode, &cr.CommissionAmount, &cr.RawData,
		&cr.CreatedAt, &cr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("channel reservation not found: %w", err)
	}
	return &cr, nil
}

func (r *PostgresChannelReservationRepository) ListPending(ctx context.Context, tenantID string) ([]*domain.ChannelReservation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, channel_integration_id, ota_confirmation_number,
		       reservation_id, status, guest_name, guest_email, guest_phone,
		       room_type_id, rate_plan_id, check_in_date, check_out_date,
		       adults, children, total_amount, currency_code, commission_amount, raw_data,
		       created_at, updated_at
		FROM channel_reservations
		WHERE tenant_id = $1 AND status = 'pending'
		ORDER BY created_at ASC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []*domain.ChannelReservation
	for rows.Next() {
		var cr domain.ChannelReservation
		err := rows.Scan(
			&cr.ID, &cr.TenantID, &cr.ChannelIntegrationID, &cr.OTAConfirmationNumber,
			&cr.ReservationID, &cr.Status, &cr.GuestName, &cr.GuestEmail, &cr.GuestPhone,
			&cr.RoomTypeID, &cr.RatePlanID, &cr.CheckInDate, &cr.CheckOutDate,
			&cr.Adults, &cr.Children, &cr.TotalAmount, &cr.CurrencyCode, &cr.CommissionAmount, &cr.RawData,
			&cr.CreatedAt, &cr.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, &cr)
	}
	return reservations, rows.Err()
}

func (r *PostgresChannelReservationRepository) UpdateStatus(ctx context.Context, tenantID, id, status string, reservationID *string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE channel_reservations
		SET status = $1, reservation_id = $2, updated_at = NOW()
		WHERE id = $3 AND tenant_id = $4
	`, status, reservationID, id, tenantID)
	return err
}
