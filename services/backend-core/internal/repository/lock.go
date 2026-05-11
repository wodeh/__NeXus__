package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// LockRepository provides smart lock data access.
type LockRepository struct {
	pool *db.Pool
}

// NewLockRepository creates a lock repository.
func NewLockRepository(pool *db.Pool) *LockRepository {
	return &LockRepository{pool: pool}
}

func (r *LockRepository) execer(tx pgx.Tx) interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
} {
	if tx != nil {
		return tx
	}
	return r.pool
}

// List returns all locks for a tenant.
func (r *LockRepository) List(ctx context.Context, tenantID string) ([]domain.SmartLock, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, room_id, room_number, serial_number, model, manufacturer,
			status, battery_level, last_communication_at, last_unlock_at, last_lock_at,
			firmware_version, remote_unlock_enabled, auto_lock_enabled, created_at, updated_at
		FROM smart_locks
		WHERE tenant_id = $1
		ORDER BY room_number
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list locks: %w", err)
	}
	defer rows.Close()

	var locks []domain.SmartLock
	for rows.Next() {
		var l domain.SmartLock
		var lastComm, lastUnlock, lastLock *time.Time
		if err := rows.Scan(
			&l.ID, &l.TenantID, &l.RoomID, &l.RoomNumber, &l.SerialNumber, &l.Model, &l.Manufacturer,
			&l.Status, &l.BatteryLevel, &lastComm, &lastUnlock, &lastLock,
			&l.FirmwareVersion, &l.RemoteUnlockEnabled, &l.AutoLockEnabled, &l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan lock: %w", err)
		}
		l.LastCommunicationAt = lastComm
		l.LastUnlockAt = lastUnlock
		l.LastLockAt = lastLock
		locks = append(locks, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("lock rows: %w", err)
	}
	return locks, nil
}

// Get returns a single lock by ID.
func (r *LockRepository) Get(ctx context.Context, tenantID string, id uuid.UUID) (*domain.SmartLock, error) {
	var l domain.SmartLock
	var lastComm, lastUnlock, lastLock *time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, room_id, room_number, serial_number, model, manufacturer,
			status, battery_level, last_communication_at, last_unlock_at, last_lock_at,
			firmware_version, remote_unlock_enabled, auto_lock_enabled, created_at, updated_at
		FROM smart_locks
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, id).Scan(
		&l.ID, &l.TenantID, &l.RoomID, &l.RoomNumber, &l.SerialNumber, &l.Model, &l.Manufacturer,
		&l.Status, &l.BatteryLevel, &lastComm, &lastUnlock, &lastLock,
		&l.FirmwareVersion, &l.RemoteUnlockEnabled, &l.AutoLockEnabled, &l.CreatedAt, &l.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get lock: %w", err)
	}
	l.LastCommunicationAt = lastComm
	l.LastUnlockAt = lastUnlock
	l.LastLockAt = lastLock
	return &l, nil
}

// Create inserts a new lock.
func (r *LockRepository) Create(ctx context.Context, tenantID string, l *domain.SmartLock) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO smart_locks (tenant_id, room_id, room_number, serial_number, model, manufacturer,
			status, battery_level, firmware_version, remote_unlock_enabled, auto_lock_enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, created_at, updated_at
	`, tenantID, l.RoomID, l.RoomNumber, l.SerialNumber, l.Model, l.Manufacturer,
		l.Status, l.BatteryLevel, l.FirmwareVersion, l.RemoteUnlockEnabled, l.AutoLockEnabled,
	).Scan(&l.ID, &l.CreatedAt, &l.UpdatedAt)
}

// UpdateStatus updates lock status and settings.
func (r *LockRepository) UpdateStatus(ctx context.Context, tenantID string, id uuid.UUID, status string, battery int, remoteUnlock, autoLock *bool) error {
	query := `UPDATE smart_locks SET status = $3, battery_level = $4, updated_at = NOW()`
	args := []interface{}{tenantID, id, status, battery}
	argIdx := 5
	if remoteUnlock != nil {
		query += fmt.Sprintf(", remote_unlock_enabled = $%d", argIdx)
		args = append(args, *remoteUnlock)
		argIdx++
	}
	if autoLock != nil {
		query += fmt.Sprintf(", auto_lock_enabled = $%d", argIdx)
		args = append(args, *autoLock)
		argIdx++
	}
	query += fmt.Sprintf(" WHERE tenant_id = $1 AND id = $2")
	
	cmdTag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update lock: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("lock %s: %w", id, ErrNotFound)
	}
	return nil
}

// RecordEvent logs a lock event.
func (r *LockRepository) RecordEvent(ctx context.Context, tenantID string, e *domain.LockEvent) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO lock_events (tenant_id, lock_id, room_number, event_type, event_source, details, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, tenantID, e.LockID, e.RoomNumber, e.EventType, e.EventSource, e.Details, e.OccurredAt,
	).Scan(&e.ID)
}

// ListEvents returns events for a lock.
func (r *LockRepository) ListEvents(ctx context.Context, tenantID string, lockID uuid.UUID) ([]domain.LockEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, lock_id, room_number, event_type, event_source, details, occurred_at
		FROM lock_events
		WHERE tenant_id = $1 AND lock_id = $2
		ORDER BY occurred_at DESC
		LIMIT 100
	`, tenantID, lockID)
	if err != nil {
		return nil, fmt.Errorf("list lock events: %w", err)
	}
	defer rows.Close()

	var events []domain.LockEvent
	for rows.Next() {
		var e domain.LockEvent
		if err := rows.Scan(&e.ID, &e.TenantID, &e.LockID, &e.RoomNumber, &e.EventType, &e.EventSource, &e.Details, &e.OccurredAt); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

// CreateAccessCode inserts a new access code.
func (r *LockRepository) CreateAccessCode(ctx context.Context, tenantID string, c *domain.AccessCode) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO lock_access_codes (tenant_id, lock_id, code, label, is_active, valid_from, valid_until, max_uses, use_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0)
		RETURNING id, created_at, updated_at
	`, tenantID, c.LockID, c.Code, c.Label, c.IsActive, c.ValidFrom, c.ValidUntil, c.MaxUses,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

// ListAccessCodes returns codes for a lock.
func (r *LockRepository) ListAccessCodes(ctx context.Context, tenantID string, lockID uuid.UUID) ([]domain.AccessCode, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, lock_id, code, label, is_active, valid_from, valid_until, max_uses, use_count, created_at, updated_at
		FROM lock_access_codes
		WHERE tenant_id = $1 AND lock_id = $2
		ORDER BY created_at DESC
	`, tenantID, lockID)
	if err != nil {
		return nil, fmt.Errorf("list access codes: %w", err)
	}
	defer rows.Close()

	var codes []domain.AccessCode
	for rows.Next() {
		var c domain.AccessCode
		var validUntil *time.Time
		var maxUses *int
		if err := rows.Scan(&c.ID, &c.TenantID, &c.LockID, &c.Code, &c.Label, &c.IsActive, &c.ValidFrom, &validUntil, &maxUses, &c.UseCount, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan access code: %w", err)
		}
		c.ValidUntil = validUntil
		c.MaxUses = maxUses
		codes = append(codes, c)
	}
	return codes, rows.Err()
}

// DeleteAccessCode removes an access code.
func (r *LockRepository) DeleteAccessCode(ctx context.Context, tenantID string, codeID uuid.UUID) error {
	cmdTag, err := r.pool.Exec(ctx, `
		DELETE FROM lock_access_codes WHERE tenant_id = $1 AND id = $2
	`, tenantID, codeID)
	if err != nil {
		return fmt.Errorf("delete access code: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
