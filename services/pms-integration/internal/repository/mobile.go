package repository

import (
	"context"

	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// MobileDeviceRepository defines device storage.
type MobileDeviceRepository interface {
	Create(ctx context.Context, d *domain.MobileDevice) error
	GetByGuest(ctx context.Context, tenantID, guestID string) ([]*domain.MobileDevice, error)
	UpdateToken(ctx context.Context, tenantID, deviceID, token string) error
	Delete(ctx context.Context, tenantID, deviceID string) error
}

// GuestSelfServiceRepository defines self-service request storage.
type GuestSelfServiceRepository interface {
	Create(ctx context.Context, req *domain.GuestSelfServiceRequest) error
	ListByGuest(ctx context.Context, tenantID, guestID string, limit, offset int) ([]*domain.GuestSelfServiceRequest, error)
	ListByReservation(ctx context.Context, tenantID, reservationID string) ([]*domain.GuestSelfServiceRequest, error)
	UpdateStatus(ctx context.Context, tenantID, id, status string) error
}

// PostgresMobileDeviceRepository is a PostgreSQL implementation.
type PostgresMobileDeviceRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewMobileDeviceRepository creates a new repository.
func NewMobileDeviceRepository(pool *db.Pool, metrics *RepositoryMetrics) MobileDeviceRepository {
	return &PostgresMobileDeviceRepository{pool: pool, metrics: metrics}
}

func (r *PostgresMobileDeviceRepository) Create(ctx context.Context, d *domain.MobileDevice) error {
	defer r.metrics.ObserveQuery("mobile_device_create")()
	query := `INSERT INTO mobile_devices (tenant_id, guest_id, device_token, platform, app_version, os_version, device_model) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, last_active_at, created_at`
	return r.pool.QueryRow(ctx, query, d.TenantID, d.GuestID, d.DeviceToken, d.Platform, d.AppVersion, d.OSVersion, d.DeviceModel).Scan(&d.ID, &d.LastActiveAt, &d.CreatedAt)
}

func (r *PostgresMobileDeviceRepository) GetByGuest(ctx context.Context, tenantID, guestID string) ([]*domain.MobileDevice, error) {
	defer r.metrics.ObserveQuery("mobile_device_list")()
	query := `SELECT id, tenant_id, guest_id, device_token, platform, app_version, os_version, device_model, last_active_at, created_at FROM mobile_devices WHERE tenant_id = $1 AND guest_id = $2 ORDER BY last_active_at DESC`
	rows, err := r.pool.Query(ctx, query, tenantID, guestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []*domain.MobileDevice
	for rows.Next() {
		var d domain.MobileDevice
		if err := rows.Scan(&d.ID, &d.TenantID, &d.GuestID, &d.DeviceToken, &d.Platform, &d.AppVersion, &d.OSVersion, &d.DeviceModel, &d.LastActiveAt, &d.CreatedAt); err != nil {
			return nil, err
		}
		devices = append(devices, &d)
	}
	return devices, rows.Err()
}

func (r *PostgresMobileDeviceRepository) UpdateToken(ctx context.Context, tenantID, deviceID, token string) error {
	defer r.metrics.ObserveQuery("mobile_device_update_token")()
	_, err := r.pool.Exec(ctx, `UPDATE mobile_devices SET device_token = $1, last_active_at = NOW() WHERE tenant_id = $2 AND id = $3`, token, tenantID, deviceID)
	return err
}

func (r *PostgresMobileDeviceRepository) Delete(ctx context.Context, tenantID, deviceID string) error {
	defer r.metrics.ObserveQuery("mobile_device_delete")()
	_, err := r.pool.Exec(ctx, `DELETE FROM mobile_devices WHERE tenant_id = $1 AND id = $2`, tenantID, deviceID)
	return err
}

// PostgresGuestSelfServiceRepository is a PostgreSQL implementation.
type PostgresGuestSelfServiceRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewGuestSelfServiceRepository creates a new repository.
func NewGuestSelfServiceRepository(pool *db.Pool, metrics *RepositoryMetrics) GuestSelfServiceRepository {
	return &PostgresGuestSelfServiceRepository{pool: pool, metrics: metrics}
}

func (r *PostgresGuestSelfServiceRepository) Create(ctx context.Context, req *domain.GuestSelfServiceRequest) error {
	defer r.metrics.ObserveQuery("self_service_create")()
	query := `INSERT INTO guest_self_service_requests (tenant_id, guest_id, reservation_id, type, status, details) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query, req.TenantID, req.GuestID, req.ReservationID, req.Type, req.Status, req.Details).Scan(&req.ID, &req.CreatedAt, &req.UpdatedAt)
}

func (r *PostgresGuestSelfServiceRepository) ListByGuest(ctx context.Context, tenantID, guestID string, limit, offset int) ([]*domain.GuestSelfServiceRequest, error) {
	defer r.metrics.ObserveQuery("self_service_list_guest")()
	query := `SELECT id, tenant_id, guest_id, reservation_id, type, status, details, created_at, updated_at FROM guest_self_service_requests WHERE tenant_id = $1 AND guest_id = $2 ORDER BY created_at DESC LIMIT $3 OFFSET $4`
	rows, err := r.pool.Query(ctx, query, tenantID, guestID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []*domain.GuestSelfServiceRequest
	for rows.Next() {
		var req domain.GuestSelfServiceRequest
		if err := rows.Scan(&req.ID, &req.TenantID, &req.GuestID, &req.ReservationID, &req.Type, &req.Status, &req.Details, &req.CreatedAt, &req.UpdatedAt); err != nil {
			return nil, err
		}
		reqs = append(reqs, &req)
	}
	return reqs, rows.Err()
}

func (r *PostgresGuestSelfServiceRepository) ListByReservation(ctx context.Context, tenantID, reservationID string) ([]*domain.GuestSelfServiceRequest, error) {
	defer r.metrics.ObserveQuery("self_service_list_reservation")()
	query := `SELECT id, tenant_id, guest_id, reservation_id, type, status, details, created_at, updated_at FROM guest_self_service_requests WHERE tenant_id = $1 AND reservation_id = $2 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, tenantID, reservationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reqs []*domain.GuestSelfServiceRequest
	for rows.Next() {
		var req domain.GuestSelfServiceRequest
		if err := rows.Scan(&req.ID, &req.TenantID, &req.GuestID, &req.ReservationID, &req.Type, &req.Status, &req.Details, &req.CreatedAt, &req.UpdatedAt); err != nil {
			return nil, err
		}
		reqs = append(reqs, &req)
	}
	return reqs, rows.Err()
}

func (r *PostgresGuestSelfServiceRepository) UpdateStatus(ctx context.Context, tenantID, id, status string) error {
	defer r.metrics.ObserveQuery("self_service_update_status")()
	_, err := r.pool.Exec(ctx, `UPDATE guest_self_service_requests SET status = $1, updated_at = NOW() WHERE tenant_id = $2 AND id = $3`, status, tenantID, id)
	return err
}
