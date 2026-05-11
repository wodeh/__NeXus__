package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// RoomRepository provides room data access.
type RoomRepository struct {
	pool *db.Pool
}

// NewRoomRepository creates a room repository.
func NewRoomRepository(pool *db.Pool) *RoomRepository {
	return &RoomRepository{pool: pool}
}

func (r *RoomRepository) execer(tx pgx.Tx) interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
} {
	if tx != nil {
		return tx
	}
	return r.pool
}

// List returns all active rooms for a tenant.
func (r *RoomRepository) List(ctx context.Context, tenantID string) ([]domain.Room, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, number, type, floor, bed_type, status, rate_night,
			config, created_at, updated_at, deleted_at, version
		FROM rooms
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY floor, number
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	var rooms []domain.Room
	for rows.Next() {
		var rm domain.Room
		var bedType *string
		var configJSON []byte
		if err := rows.Scan(
			&rm.ID, &rm.TenantID, &rm.PropertyID, &rm.Number, &rm.Type, &rm.Floor,
			&bedType, &rm.Status, &rm.RateNight, &configJSON,
			&rm.CreatedAt, &rm.UpdatedAt, &rm.DeletedAt, &rm.Version,
		); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		rm.BedType = bedType
		rooms = append(rooms, rm)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("room rows: %w", err)
	}
	return rooms, nil
}

// Create inserts a new room.
func (r *RoomRepository) Create(ctx context.Context, tenantID string, rm *domain.Room) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO rooms (tenant_id, property_id, number, type, floor, bed_type, status, rate_night, config)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at, version
	`, tenantID, rm.PropertyID, rm.Number, rm.Type, rm.Floor, rm.BedType, rm.Status, rm.RateNight, `{"wifi":true,"tv":true,"ac":true}`).Scan(
		&rm.ID, &rm.CreatedAt, &rm.UpdatedAt, &rm.Version,
	)
}

// UpdateStatus updates a room's status.
func (r *RoomRepository) UpdateStatus(ctx context.Context, tenantID string, number string, status string) error {
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE rooms
		SET status = $3, updated_at = NOW(), version = version + 1
		WHERE number = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, number, tenantID, status)
	if err != nil {
		return fmt.Errorf("update room status: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("room %s: %w", number, ErrNotFound)
	}
	return nil
}
