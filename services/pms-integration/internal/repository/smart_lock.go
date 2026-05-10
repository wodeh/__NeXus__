package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// SmartLockRepository provides CRUD for smart lock entities.
type SmartLockRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewSmartLockRepository creates a new smart lock repository.
func NewSmartLockRepository(pool *db.Pool, metrics *RepositoryMetrics) *SmartLockRepository {
	return &SmartLockRepository{pool: pool, metrics: metrics}
}

func (r *SmartLockRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// ==================== SMART LOCKS ====================

func (r *SmartLockRepository) CreateLock(ctx context.Context, lock *domain.SmartLock) error {
	start := time.Now()
	r.metrics.IncQuery("smart_lock", "create")
	defer r.metrics.ObserveDuration("smart_lock", "create", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, lock.TenantID); err != nil {
		r.metrics.IncError("smart_lock", "create", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO smart_locks (tenant_id, property_id, room_id, device_name, device_serial, manufacturer, model,
		firmware_version, status, battery_level, connection_type, ip_address, mac_address,
		remote_unlock_enabled, auto_lock_enabled, auto_lock_delay_seconds, settings, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`, lock.TenantID, lock.PropertyID, lock.RoomID, lock.DeviceName, lock.DeviceSerial, lock.Manufacturer, lock.Model,
		lock.FirmwareVersion, lock.Status, lock.BatteryLevel, lock.ConnectionType, lock.IPAddress, lock.MACAddress,
		lock.RemoteUnlockEnabled, lock.AutoLockEnabled, lock.AutoLockDelaySeconds, lock.Settings, lock.Metadata)
	if err != nil {
		r.metrics.IncError("smart_lock", "create", "query")
		return fmt.Errorf("create lock: %w", err)
	}
	return nil
}

func (r *SmartLockRepository) ListLocks(ctx context.Context, tenantID, propertyID string) ([]domain.SmartLock, error) {
	start := time.Now()
	r.metrics.IncQuery("smart_lock", "list")
	defer r.metrics.ObserveDuration("smart_lock", "list", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("smart_lock", "list", "tenant_bind")
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, room_id, device_name, device_serial, manufacturer, model,
		firmware_version, status, battery_level, connection_type, ip_address, mac_address,
		last_communication_at, last_unlock_at, last_lock_at,
		remote_unlock_enabled, auto_lock_enabled, auto_lock_delay_seconds, settings, metadata, created_at, updated_at
		FROM smart_locks WHERE tenant_id = $1 AND property_id = $2 ORDER BY device_name
	`, tenantID, propertyID)
	if err != nil {
		r.metrics.IncError("smart_lock", "list", "query")
		return nil, fmt.Errorf("list locks: %w", err)
	}
	defer rows.Close()

	return scanSmartLocks(rows)
}

func (r *SmartLockRepository) GetLock(ctx context.Context, tenantID, id string) (*domain.SmartLock, error) {
	start := time.Now()
	r.metrics.IncQuery("smart_lock", "get")
	defer r.metrics.ObserveDuration("smart_lock", "get", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("smart_lock", "get", "tenant_bind")
		return nil, err
	}

	var l domain.SmartLock
	var settings, metadata map[string]interface{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, room_id, device_name, device_serial, manufacturer, model,
		firmware_version, status, battery_level, connection_type, ip_address, mac_address,
		last_communication_at, last_unlock_at, last_lock_at,
		remote_unlock_enabled, auto_lock_enabled, auto_lock_delay_seconds, settings, metadata, created_at, updated_at
		FROM smart_locks WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(
		&l.ID, &l.TenantID, &l.PropertyID, &l.RoomID, &l.DeviceName, &l.DeviceSerial, &l.Manufacturer, &l.Model,
		&l.FirmwareVersion, &l.Status, &l.BatteryLevel, &l.ConnectionType, &l.IPAddress, &l.MACAddress,
		&l.LastCommunicationAt, &l.LastUnlockAt, &l.LastLockAt,
		&l.RemoteUnlockEnabled, &l.AutoLockEnabled, &l.AutoLockDelaySeconds, &settings, &metadata, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		r.metrics.IncError("smart_lock", "get", "query")
		return nil, fmt.Errorf("get lock: %w", err)
	}
	l.Settings = settings
	l.Metadata = metadata
	return &l, nil
}

func (r *SmartLockRepository) UpdateLock(ctx context.Context, lock *domain.SmartLock) error {
	start := time.Now()
	r.metrics.IncQuery("smart_lock", "update")
	defer r.metrics.ObserveDuration("smart_lock", "update", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, lock.TenantID); err != nil {
		r.metrics.IncError("smart_lock", "update", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE smart_locks SET room_id = $1, device_name = $2, firmware_version = $3, status = $4,
		battery_level = $5, ip_address = $6, mac_address = $7, last_communication_at = $8,
		last_unlock_at = $9, last_lock_at = $10, remote_unlock_enabled = $11, auto_lock_enabled = $12,
		auto_lock_delay_seconds = $13, settings = $14, metadata = $15
		WHERE id = $16 AND tenant_id = $17
	`, lock.RoomID, lock.DeviceName, lock.FirmwareVersion, lock.Status,
		lock.BatteryLevel, lock.IPAddress, lock.MACAddress, lock.LastCommunicationAt,
		lock.LastUnlockAt, lock.LastLockAt, lock.RemoteUnlockEnabled, lock.AutoLockEnabled,
		lock.AutoLockDelaySeconds, lock.Settings, lock.Metadata,
		lock.ID, lock.TenantID)
	if err != nil {
		r.metrics.IncError("smart_lock", "update", "query")
		return fmt.Errorf("update lock: %w", err)
	}
	return nil
}

func (r *SmartLockRepository) DeleteLock(ctx context.Context, id, tenantID string) error {
	start := time.Now()
	r.metrics.IncQuery("smart_lock", "delete")
	defer r.metrics.ObserveDuration("smart_lock", "delete", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("smart_lock", "delete", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `DELETE FROM smart_locks WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		r.metrics.IncError("smart_lock", "delete", "query")
		return fmt.Errorf("delete lock: %w", err)
	}
	return nil
}

func (r *SmartLockRepository) GetLockByRoom(ctx context.Context, tenantID, roomID string) (*domain.SmartLock, error) {
	start := time.Now()
	r.metrics.IncQuery("smart_lock", "get_by_room")
	defer r.metrics.ObserveDuration("smart_lock", "get_by_room", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("smart_lock", "get_by_room", "tenant_bind")
		return nil, err
	}

	var l domain.SmartLock
	var settings, metadata map[string]interface{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, room_id, device_name, device_serial, manufacturer, model,
		firmware_version, status, battery_level, connection_type, ip_address, mac_address,
		last_communication_at, last_unlock_at, last_lock_at,
		remote_unlock_enabled, auto_lock_enabled, auto_lock_delay_seconds, settings, metadata, created_at, updated_at
		FROM smart_locks WHERE tenant_id = $1 AND room_id = $2 LIMIT 1
	`, tenantID, roomID).Scan(
		&l.ID, &l.TenantID, &l.PropertyID, &l.RoomID, &l.DeviceName, &l.DeviceSerial, &l.Manufacturer, &l.Model,
		&l.FirmwareVersion, &l.Status, &l.BatteryLevel, &l.ConnectionType, &l.IPAddress, &l.MACAddress,
		&l.LastCommunicationAt, &l.LastUnlockAt, &l.LastLockAt,
		&l.RemoteUnlockEnabled, &l.AutoLockEnabled, &l.AutoLockDelaySeconds, &settings, &metadata, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		r.metrics.IncError("smart_lock", "get_by_room", "query")
		return nil, fmt.Errorf("get lock by room: %w", err)
	}
	l.Settings = settings
	l.Metadata = metadata
	return &l, nil
}

// ==================== ACCESS CODES ====================

func (r *SmartLockRepository) CreateAccessCode(ctx context.Context, code *domain.LockAccessCode) error {
	start := time.Now()
	r.metrics.IncQuery("lock_access_code", "create")
	defer r.metrics.ObserveDuration("lock_access_code", "create", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, code.TenantID); err != nil {
		r.metrics.IncError("lock_access_code", "create", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO lock_access_codes (tenant_id, property_id, smart_lock_id, reservation_id, guest_id, code,
		code_type, label, valid_from, valid_until, max_uses, is_active, is_master, created_by, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, code.TenantID, code.PropertyID, code.SmartLockID, code.ReservationID, code.GuestID, code.Code,
		code.CodeType, code.Label, code.ValidFrom, code.ValidUntil, code.MaxUses, code.IsActive, code.IsMaster, code.CreatedBy, code.Metadata)
	if err != nil {
		r.metrics.IncError("lock_access_code", "create", "query")
		return fmt.Errorf("create access code: %w", err)
	}
	return nil
}

func (r *SmartLockRepository) ListAccessCodes(ctx context.Context, lockID string) ([]domain.LockAccessCode, error) {
	start := time.Now()
	r.metrics.IncQuery("lock_access_code", "list")
	defer r.metrics.ObserveDuration("lock_access_code", "list", time.Since(start).Seconds())

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, smart_lock_id, reservation_id, guest_id, code, code_type, label,
		valid_from, valid_until, max_uses, uses_count, is_active, is_master, created_by,
		created_at, updated_at, revoked_at, revoked_reason, metadata
		FROM lock_access_codes WHERE smart_lock_id = $1 ORDER BY created_at DESC
	`, lockID)
	if err != nil {
		r.metrics.IncError("lock_access_code", "list", "query")
		return nil, fmt.Errorf("list access codes: %w", err)
	}
	defer rows.Close()

	return scanAccessCodes(rows)
}

func (r *SmartLockRepository) GetAccessCode(ctx context.Context, id string) (*domain.LockAccessCode, error) {
	start := time.Now()
	r.metrics.IncQuery("lock_access_code", "get")
	defer r.metrics.ObserveDuration("lock_access_code", "get", time.Since(start).Seconds())

	var c domain.LockAccessCode
	var metadata map[string]interface{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, smart_lock_id, reservation_id, guest_id, code, code_type, label,
		valid_from, valid_until, max_uses, uses_count, is_active, is_master, created_by,
		created_at, updated_at, revoked_at, revoked_reason, metadata
		FROM lock_access_codes WHERE id = $1
	`, id).Scan(
		&c.ID, &c.TenantID, &c.PropertyID, &c.SmartLockID, &c.ReservationID, &c.GuestID, &c.Code, &c.CodeType, &c.Label,
		&c.ValidFrom, &c.ValidUntil, &c.MaxUses, &c.UsesCount, &c.IsActive, &c.IsMaster, &c.CreatedBy,
		&c.CreatedAt, &c.UpdatedAt, &c.RevokedAt, &c.RevokedReason, &metadata,
	)
	if err != nil {
		r.metrics.IncError("lock_access_code", "get", "query")
		return nil, fmt.Errorf("get access code: %w", err)
	}
	c.Metadata = metadata
	return &c, nil
}

func (r *SmartLockRepository) RevokeAccessCode(ctx context.Context, id, reason string) error {
	start := time.Now()
	r.metrics.IncQuery("lock_access_code", "revoke")
	defer r.metrics.ObserveDuration("lock_access_code", "revoke", time.Since(start).Seconds())

	_, err := r.pool.Exec(ctx,
		`UPDATE lock_access_codes SET is_active = false, revoked_at = $1, revoked_reason = $2 WHERE id = $3`,
		time.Now(), reason, id,
	)
	if err != nil {
		r.metrics.IncError("lock_access_code", "revoke", "query")
		return fmt.Errorf("revoke access code: %w", err)
	}
	return nil
}

func (r *SmartLockRepository) IncrementCodeUse(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE lock_access_codes SET uses_count = uses_count + 1 WHERE id = $1`, id,
	)
	return err
}

// ==================== EVENTS ====================

func (r *SmartLockRepository) CreateEvent(ctx context.Context, event *domain.LockEvent) error {
	start := time.Now()
	r.metrics.IncQuery("lock_event", "create")
	defer r.metrics.ObserveDuration("lock_event", "create", time.Since(start).Seconds())

	_, err := r.pool.Exec(ctx, `
		INSERT INTO lock_events (tenant_id, property_id, smart_lock_id, access_code_id, event_type, event_source,
		user_id, guest_id, details, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, event.TenantID, event.PropertyID, event.SmartLockID, event.AccessCodeID, event.EventType,
		event.EventSource, event.UserID, event.GuestID, event.Details, event.OccurredAt)
	if err != nil {
		r.metrics.IncError("lock_event", "create", "query")
		return fmt.Errorf("create event: %w", err)
	}
	return nil
}

func (r *SmartLockRepository) ListEvents(ctx context.Context, lockID string, limit int) ([]domain.LockEvent, error) {
	start := time.Now()
	r.metrics.IncQuery("lock_event", "list")
	defer r.metrics.ObserveDuration("lock_event", "list", time.Since(start).Seconds())

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, smart_lock_id, access_code_id, event_type, event_source,
		user_id, guest_id, details, occurred_at
		FROM lock_events WHERE smart_lock_id = $1 ORDER BY occurred_at DESC LIMIT $2
	`, lockID, limit)
	if err != nil {
		r.metrics.IncError("lock_event", "list", "query")
		return nil, fmt.Errorf("list events: %w", err)
	}
	defer rows.Close()

	return scanLockEvents(rows)
}

func (r *SmartLockRepository) GetEventsByType(ctx context.Context, tenantID, eventType string, from, to time.Time) ([]domain.LockEvent, error) {
	start := time.Now()
	r.metrics.IncQuery("lock_event", "get_by_type")
	defer r.metrics.ObserveDuration("lock_event", "get_by_type", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("lock_event", "get_by_type", "tenant_bind")
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, smart_lock_id, access_code_id, event_type, event_source,
		user_id, guest_id, details, occurred_at
		FROM lock_events WHERE tenant_id = $1 AND event_type = $2 AND occurred_at BETWEEN $3 AND $4
		ORDER BY occurred_at DESC
	`, tenantID, eventType, from, to)
	if err != nil {
		r.metrics.IncError("lock_event", "get_by_type", "query")
		return nil, fmt.Errorf("get events by type: %w", err)
	}
	defer rows.Close()

	return scanLockEvents(rows)
}

// ==================== OVERVIEW ====================

func (r *SmartLockRepository) GetOverview(ctx context.Context, tenantID, propertyID string) (*domain.LockOverview, error) {
	start := time.Now()
	r.metrics.IncQuery("smart_lock", "overview")
	defer r.metrics.ObserveDuration("smart_lock", "overview", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("smart_lock", "overview", "tenant_bind")
		return nil, err
	}

	var o domain.LockOverview
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM smart_locks WHERE tenant_id = $1 AND property_id = $2`,
		tenantID, propertyID,
	).Scan(&o.TotalLocks)

	_ = r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FILTER (WHERE status = 'online'), COUNT(*) FILTER (WHERE status = 'offline'),
		COUNT(*) FILTER (WHERE battery_level < 20)
		FROM smart_locks WHERE tenant_id = $1 AND property_id = $2`,
		tenantID, propertyID,
	).Scan(&o.OnlineLocks, &o.OfflineLocks, &o.LowBatteryLocks)

	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM lock_access_codes WHERE tenant_id = $1 AND property_id = $2 AND is_active = true`,
		tenantID, propertyID,
	).Scan(&o.ActiveAccessCodes)

	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM lock_events WHERE tenant_id = $1 AND property_id = $2 AND occurred_at >= CURRENT_DATE`,
		tenantID, propertyID,
	).Scan(&o.TodayEvents)

	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM lock_events WHERE tenant_id = $1 AND property_id = $2 AND event_type = 'unlock' AND occurred_at >= CURRENT_DATE`,
		tenantID, propertyID,
	).Scan(&o.RecentUnlocks)

	return &o, nil
}

// ==================== SCAN HELPERS ====================

func scanSmartLocks(rows pgx.Rows) ([]domain.SmartLock, error) {
	var locks []domain.SmartLock
	for rows.Next() {
		var l domain.SmartLock
		var settings, metadata map[string]interface{}
		if err := rows.Scan(
			&l.ID, &l.TenantID, &l.PropertyID, &l.RoomID, &l.DeviceName, &l.DeviceSerial, &l.Manufacturer, &l.Model,
			&l.FirmwareVersion, &l.Status, &l.BatteryLevel, &l.ConnectionType, &l.IPAddress, &l.MACAddress,
			&l.LastCommunicationAt, &l.LastUnlockAt, &l.LastLockAt,
			&l.RemoteUnlockEnabled, &l.AutoLockEnabled, &l.AutoLockDelaySeconds, &settings, &metadata, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan lock: %w", err)
		}
		l.Settings = settings
		l.Metadata = metadata
		locks = append(locks, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return locks, nil
}

func scanAccessCodes(rows pgx.Rows) ([]domain.LockAccessCode, error) {
	var codes []domain.LockAccessCode
	for rows.Next() {
		var c domain.LockAccessCode
		var metadata map[string]interface{}
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.PropertyID, &c.SmartLockID, &c.ReservationID, &c.GuestID, &c.Code, &c.CodeType, &c.Label,
			&c.ValidFrom, &c.ValidUntil, &c.MaxUses, &c.UsesCount, &c.IsActive, &c.IsMaster, &c.CreatedBy,
			&c.CreatedAt, &c.UpdatedAt, &c.RevokedAt, &c.RevokedReason, &metadata,
		); err != nil {
			return nil, fmt.Errorf("scan access code: %w", err)
		}
		c.Metadata = metadata
		codes = append(codes, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return codes, nil
}

func scanLockEvents(rows pgx.Rows) ([]domain.LockEvent, error) {
	var events []domain.LockEvent
	for rows.Next() {
		var e domain.LockEvent
		var details map[string]interface{}
		if err := rows.Scan(
			&e.ID, &e.TenantID, &e.PropertyID, &e.SmartLockID, &e.AccessCodeID, &e.EventType, &e.EventSource,
			&e.UserID, &e.GuestID, &details, &e.OccurredAt,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		e.Details = details
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return events, nil
}
