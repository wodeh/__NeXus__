// Package repository provides tenant-aware data access for smart locks.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// SmartLockRepository persists and retrieves SmartLock aggregates.
type SmartLockRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.SmartLock, error)
	GetByDeviceID(ctx context.Context, tenantID, deviceID string) (*domain.SmartLock, error)
	GetByRoom(ctx context.Context, tenantID, roomID string) (*domain.SmartLock, error)
	Create(ctx context.Context, lock *domain.SmartLock) error
	Update(ctx context.Context, lock *domain.SmartLock) error
	UpdateStatus(ctx context.Context, tenantID, id string, isOnline bool, batteryLevel int) error
	ListByProperty(ctx context.Context, tenantID, propertyID string, limit, offset int) ([]*domain.SmartLock, error)
}

// PostgresSmartLockRepository is the PostgreSQL implementation.
type PostgresSmartLockRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewSmartLockRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresSmartLockRepository {
	return &PostgresSmartLockRepository{pool: pool, metrics: metrics}
}

func (r *PostgresSmartLockRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresSmartLockRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.SmartLock, error) {
	start := time.Now()
	r.metrics.IncQuery("smart_lock", "get_by_id")
	defer r.metrics.ObserveDuration("smart_lock", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, room_id, device_id, device_model,
		       firmware_version, battery_level, is_online, last_seen_at, is_active,
		       created_at, updated_at, version
		FROM smart_locks
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)

	var lock domain.SmartLock
	err := row.Scan(
		&lock.ID, &lock.TenantID, &lock.PropertyID, &lock.RoomID, &lock.DeviceID, &lock.DeviceModel,
		&lock.FirmwareVersion, &lock.BatteryLevel, &lock.IsOnline, &lock.LastSeenAt, &lock.IsActive,
		&lock.CreatedAt, &lock.UpdatedAt, &lock.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("smart lock not found: %w", err)
	}
	return &lock, nil
}

func (r *PostgresSmartLockRepository) GetByDeviceID(ctx context.Context, tenantID, deviceID string) (*domain.SmartLock, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, room_id, device_id, device_model,
		       firmware_version, battery_level, is_online, last_seen_at, is_active,
		       created_at, updated_at, version
		FROM smart_locks
		WHERE device_id = $1 AND tenant_id = $2
	`, deviceID, tenantID)

	var lock domain.SmartLock
	err := row.Scan(
		&lock.ID, &lock.TenantID, &lock.PropertyID, &lock.RoomID, &lock.DeviceID, &lock.DeviceModel,
		&lock.FirmwareVersion, &lock.BatteryLevel, &lock.IsOnline, &lock.LastSeenAt, &lock.IsActive,
		&lock.CreatedAt, &lock.UpdatedAt, &lock.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("smart lock not found: %w", err)
	}
	return &lock, nil
}

func (r *PostgresSmartLockRepository) GetByRoom(ctx context.Context, tenantID, roomID string) (*domain.SmartLock, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, room_id, device_id, device_model,
		       firmware_version, battery_level, is_online, last_seen_at, is_active,
		       created_at, updated_at, version
		FROM smart_locks
		WHERE room_id = $1 AND tenant_id = $2 AND is_active = TRUE
	`, roomID, tenantID)

	var lock domain.SmartLock
	err := row.Scan(
		&lock.ID, &lock.TenantID, &lock.PropertyID, &lock.RoomID, &lock.DeviceID, &lock.DeviceModel,
		&lock.FirmwareVersion, &lock.BatteryLevel, &lock.IsOnline, &lock.LastSeenAt, &lock.IsActive,
		&lock.CreatedAt, &lock.UpdatedAt, &lock.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("smart lock not found: %w", err)
	}
	return &lock, nil
}

func (r *PostgresSmartLockRepository) Create(ctx context.Context, lock *domain.SmartLock) error {
	start := time.Now()
	r.metrics.IncQuery("smart_lock", "create")
	defer r.metrics.ObserveDuration("smart_lock", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, lock.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO smart_locks (id, tenant_id, property_id, room_id, device_id, device_model,
		                        firmware_version, battery_level, is_online, last_seen_at, is_active,
		                        version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, lock.ID, lock.TenantID, lock.PropertyID, lock.RoomID, lock.DeviceID, lock.DeviceModel,
		lock.FirmwareVersion, lock.BatteryLevel, lock.IsOnline, lock.LastSeenAt, lock.IsActive,
		lock.Version, lock.CreatedAt, lock.UpdatedAt)
	return err
}

func (r *PostgresSmartLockRepository) Update(ctx context.Context, lock *domain.SmartLock) error {
	start := time.Now()
	r.metrics.IncQuery("smart_lock", "update")
	defer r.metrics.ObserveDuration("smart_lock", "update", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, lock.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE smart_locks
		SET device_model = $1, firmware_version = $2, battery_level = $3,
		    is_online = $4, last_seen_at = $5, is_active = $6,
		    version = $7, updated_at = $8
		WHERE id = $9 AND tenant_id = $10 AND version = $11
	`, lock.DeviceModel, lock.FirmwareVersion, lock.BatteryLevel,
		lock.IsOnline, lock.LastSeenAt, lock.IsActive,
		lock.Version, lock.UpdatedAt,
		lock.ID, lock.TenantID, lock.Version-1)
	return err
}

func (r *PostgresSmartLockRepository) UpdateStatus(ctx context.Context, tenantID, id string, isOnline bool, batteryLevel int) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE smart_locks
		SET is_online = $1, battery_level = $2, last_seen_at = NOW(), updated_at = NOW(), version = version + 1
		WHERE id = $3 AND tenant_id = $4
	`, isOnline, batteryLevel, id, tenantID)
	return err
}

func (r *PostgresSmartLockRepository) ListByProperty(ctx context.Context, tenantID, propertyID string, limit, offset int) ([]*domain.SmartLock, error) {
	start := time.Now()
	r.metrics.IncQuery("smart_lock", "list_by_property")
	defer r.metrics.ObserveDuration("smart_lock", "list_by_property", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, room_id, device_id, device_model,
		       firmware_version, battery_level, is_online, last_seen_at, is_active,
		       created_at, updated_at, version
		FROM smart_locks
		WHERE tenant_id = $1 AND property_id = $2
		ORDER BY created_at DESC LIMIT $3 OFFSET $4
	`, tenantID, propertyID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanSmartLocks(rows)
}

func scanSmartLocks(rows pgx.Rows) ([]*domain.SmartLock, error) {
	var locks []*domain.SmartLock
	for rows.Next() {
		var lock domain.SmartLock
		err := rows.Scan(
			&lock.ID, &lock.TenantID, &lock.PropertyID, &lock.RoomID, &lock.DeviceID, &lock.DeviceModel,
			&lock.FirmwareVersion, &lock.BatteryLevel, &lock.IsOnline, &lock.LastSeenAt, &lock.IsActive,
			&lock.CreatedAt, &lock.UpdatedAt, &lock.Version,
		)
		if err != nil {
			return nil, err
		}
		locks = append(locks, &lock)
	}
	return locks, rows.Err()
}

// ==================== DIGITAL KEY REPOSITORY ====================

type DigitalKeyRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.DigitalKey, error)
	GetByReservation(ctx context.Context, tenantID, reservationID string) (*domain.DigitalKey, error)
	Create(ctx context.Context, key *domain.DigitalKey) error
	Update(ctx context.Context, key *domain.DigitalKey) error
	Revoke(ctx context.Context, tenantID, id, revokedBy, reason string) error
	ListBySmartLock(ctx context.Context, tenantID, smartLockID string, status string, limit, offset int) ([]*domain.DigitalKey, error)
	RecordUse(ctx context.Context, tenantID, id string) error
}

type PostgresDigitalKeyRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewDigitalKeyRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresDigitalKeyRepository {
	return &PostgresDigitalKeyRepository{pool: pool, metrics: metrics}
}

func (r *PostgresDigitalKeyRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresDigitalKeyRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.DigitalKey, error) {
	start := time.Now()
	r.metrics.IncQuery("digital_key", "get_by_id")
	defer r.metrics.ObserveDuration("digital_key", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, smart_lock_id, reservation_id, guest_id, guest_email, guest_phone,
		       key_code, pin_code, nfc_token, ble_token, status, valid_from, valid_until,
		       max_uses, uses_remaining, last_used_at, issued_by, issued_at,
		       revoked_at, revoked_by, revoke_reason, metadata, created_at, updated_at, version
		FROM digital_keys
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)

	var key domain.DigitalKey
	var lastUsedAt, revokedAt *time.Time
	err := row.Scan(
		&key.ID, &key.TenantID, &key.SmartLockID, &key.ReservationID, &key.GuestID, &key.GuestEmail, &key.GuestPhone,
		&key.KeyCode, &key.PinCode, &key.NfcToken, &key.BleToken, &key.Status, &key.ValidFrom, &key.ValidUntil,
		&key.MaxUses, &key.UsesRemaining, &lastUsedAt, &key.IssuedBy, &key.IssuedAt,
		&revokedAt, &key.RevokedBy, &key.RevokeReason, &key.Metadata, &key.CreatedAt, &key.UpdatedAt, &key.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("digital key not found: %w", err)
	}
	key.LastUsedAt = lastUsedAt
	key.RevokedAt = revokedAt
	return &key, nil
}

func (r *PostgresDigitalKeyRepository) GetByReservation(ctx context.Context, tenantID, reservationID string) (*domain.DigitalKey, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, smart_lock_id, reservation_id, guest_id, guest_email, guest_phone,
		       key_code, pin_code, nfc_token, ble_token, status, valid_from, valid_until,
		       max_uses, uses_remaining, last_used_at, issued_by, issued_at,
		       revoked_at, revoked_by, revoke_reason, metadata, created_at, updated_at, version
		FROM digital_keys
		WHERE reservation_id = $1 AND tenant_id = $2 AND status = 'active'
		ORDER BY created_at DESC LIMIT 1
	`, reservationID, tenantID)

	var key domain.DigitalKey
	var lastUsedAt, revokedAt *time.Time
	err := row.Scan(
		&key.ID, &key.TenantID, &key.SmartLockID, &key.ReservationID, &key.GuestID, &key.GuestEmail, &key.GuestPhone,
		&key.KeyCode, &key.PinCode, &key.NfcToken, &key.BleToken, &key.Status, &key.ValidFrom, &key.ValidUntil,
		&key.MaxUses, &key.UsesRemaining, &lastUsedAt, &key.IssuedBy, &key.IssuedAt,
		&revokedAt, &key.RevokedBy, &key.RevokeReason, &key.Metadata, &key.CreatedAt, &key.UpdatedAt, &key.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("digital key not found: %w", err)
	}
	key.LastUsedAt = lastUsedAt
	key.RevokedAt = revokedAt
	return &key, nil
}

func (r *PostgresDigitalKeyRepository) Create(ctx context.Context, key *domain.DigitalKey) error {
	start := time.Now()
	r.metrics.IncQuery("digital_key", "create")
	defer r.metrics.ObserveDuration("digital_key", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, key.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO digital_keys (id, tenant_id, smart_lock_id, reservation_id, guest_id, guest_email, guest_phone,
		                        key_code, pin_code, nfc_token, ble_token, status, valid_from, valid_until,
		                        max_uses, uses_remaining, issued_by, issued_at, metadata,
		                        version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22)
	`, key.ID, key.TenantID, key.SmartLockID, key.ReservationID, key.GuestID, key.GuestEmail, key.GuestPhone,
		key.KeyCode, key.PinCode, key.NfcToken, key.BleToken, key.Status, key.ValidFrom, key.ValidUntil,
		key.MaxUses, key.UsesRemaining, key.IssuedBy, key.IssuedAt, key.Metadata,
		key.Version, key.CreatedAt, key.UpdatedAt)
	return err
}

func (r *PostgresDigitalKeyRepository) Update(ctx context.Context, key *domain.DigitalKey) error {
	start := time.Now()
	r.metrics.IncQuery("digital_key", "update")
	defer r.metrics.ObserveDuration("digital_key", "update", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, key.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE digital_keys
		SET status = $1, key_code = $2, pin_code = $3, nfc_token = $4, ble_token = $5,
		    valid_from = $6, valid_until = $7, max_uses = $8, uses_remaining = $9,
		    last_used_at = $10, metadata = $11, version = $12, updated_at = $13
		WHERE id = $14 AND tenant_id = $15 AND version = $16
	`, key.Status, key.KeyCode, key.PinCode, key.NfcToken, key.BleToken,
		key.ValidFrom, key.ValidUntil, key.MaxUses, key.UsesRemaining,
		key.LastUsedAt, key.Metadata, key.Version, key.UpdatedAt,
		key.ID, key.TenantID, key.Version-1)
	return err
}

func (r *PostgresDigitalKeyRepository) Revoke(ctx context.Context, tenantID, id, revokedBy, reason string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE digital_keys
		SET status = 'revoked', revoked_at = $1, revoked_by = $2, revoke_reason = $3,
		    updated_at = $1, version = version + 1
		WHERE id = $4 AND tenant_id = $5 AND status != 'revoked'
	`, now, revokedBy, reason, id, tenantID)
	return err
}

func (r *PostgresDigitalKeyRepository) ListBySmartLock(ctx context.Context, tenantID, smartLockID string, status string, limit, offset int) ([]*domain.DigitalKey, error) {
	start := time.Now()
	r.metrics.IncQuery("digital_key", "list_by_lock")
	defer r.metrics.ObserveDuration("digital_key", "list_by_lock", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	var rows pgx.Rows
	var err error
	if status != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, smart_lock_id, reservation_id, guest_id, guest_email, guest_phone,
			       key_code, pin_code, nfc_token, ble_token, status, valid_from, valid_until,
			       max_uses, uses_remaining, last_used_at, issued_by, issued_at,
			       revoked_at, revoked_by, revoke_reason, metadata, created_at, updated_at, version
			FROM digital_keys
			WHERE smart_lock_id = $1 AND tenant_id = $2 AND status = $3
			ORDER BY created_at DESC LIMIT $4 OFFSET $5
		`, smartLockID, tenantID, status, limit, offset)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, smart_lock_id, reservation_id, guest_id, guest_email, guest_phone,
			       key_code, pin_code, nfc_token, ble_token, status, valid_from, valid_until,
			       max_uses, uses_remaining, last_used_at, issued_by, issued_at,
			       revoked_at, revoked_by, revoke_reason, metadata, created_at, updated_at, version
			FROM digital_keys
			WHERE smart_lock_id = $1 AND tenant_id = $2
			ORDER BY created_at DESC LIMIT $3 OFFSET $4
		`, smartLockID, tenantID, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*domain.DigitalKey
	for rows.Next() {
		var key domain.DigitalKey
		var lastUsedAt, revokedAt *time.Time
		err := rows.Scan(
			&key.ID, &key.TenantID, &key.SmartLockID, &key.ReservationID, &key.GuestID, &key.GuestEmail, &key.GuestPhone,
			&key.KeyCode, &key.PinCode, &key.NfcToken, &key.BleToken, &key.Status, &key.ValidFrom, &key.ValidUntil,
			&key.MaxUses, &key.UsesRemaining, &lastUsedAt, &key.IssuedBy, &key.IssuedAt,
			&revokedAt, &key.RevokedBy, &key.RevokeReason, &key.Metadata, &key.CreatedAt, &key.UpdatedAt, &key.Version,
		)
		if err != nil {
			return nil, err
		}
		key.LastUsedAt = lastUsedAt
		key.RevokedAt = revokedAt
		keys = append(keys, &key)
	}
	return keys, rows.Err()
}

func (r *PostgresDigitalKeyRepository) RecordUse(ctx context.Context, tenantID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}

	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE digital_keys
		SET uses_remaining = uses_remaining - 1, last_used_at = $1, updated_at = $1
		WHERE id = $2 AND tenant_id = $3 AND status = 'active' AND uses_remaining > 0
	`, now, id, tenantID)
	return err
}
