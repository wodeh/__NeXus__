// Package repository provides tenant-aware data access for property inventory.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// PropertyRepository persists and retrieves Property aggregates.
type PropertyRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.Property, error)
	GetBySlug(ctx context.Context, tenantID, slug string) (*domain.Property, error)
	Create(ctx context.Context, p *domain.Property) error
	Update(ctx context.Context, p *domain.Property) error
	SoftDelete(ctx context.Context, tenantID, id string) error
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Property, error)
}

// PostgresPropertyRepository is the PostgreSQL implementation.
type PostgresPropertyRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewPropertyRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresPropertyRepository {
	return &PostgresPropertyRepository{pool: pool, metrics: metrics}
}

func (r *PostgresPropertyRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresPropertyRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Property, error) {
	start := time.Now()
	r.metrics.IncQuery("property", "get_by_id")
	defer r.metrics.ObserveDuration("property", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	var p domain.Property
	var address map[string]string
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, slug, address, timezone, currency_code, contact_email, contact_phone, is_active, created_at, updated_at, version
		FROM properties WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID).Scan(
		&p.ID, &p.TenantID, &p.Name, &p.Slug, &address, &p.Timezone, &p.CurrencyCode, &p.ContactEmail, &p.ContactPhone, &p.IsActive, &p.CreatedAt, &p.UpdatedAt, &p.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("get property: %w", err)
	}
	p.Address = address
	return &p, nil
}

func (r *PostgresPropertyRepository) GetBySlug(ctx context.Context, tenantID, slug string) (*domain.Property, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	var p domain.Property
	var address map[string]string
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, slug, address, timezone, currency_code, contact_email, contact_phone, is_active, created_at, updated_at, version
		FROM properties WHERE tenant_id = $1 AND slug = $2 AND deleted_at IS NULL
	`, tenantID, slug).Scan(
		&p.ID, &p.TenantID, &p.Name, &p.Slug, &address, &p.Timezone, &p.CurrencyCode, &p.ContactEmail, &p.ContactPhone, &p.IsActive, &p.CreatedAt, &p.UpdatedAt, &p.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("get property by slug: %w", err)
	}
	p.Address = address
	return &p, nil
}

func (r *PostgresPropertyRepository) Create(ctx context.Context, p *domain.Property) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, p.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO properties (id, tenant_id, name, slug, address, timezone, currency_code, contact_email, contact_phone, is_active, created_at, updated_at, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, p.ID, p.TenantID, p.Name, p.Slug, p.Address, p.Timezone, p.CurrencyCode, p.ContactEmail, p.ContactPhone, p.IsActive, p.CreatedAt, p.UpdatedAt, p.Version)
	if err != nil {
		return fmt.Errorf("create property: %w", err)
	}
	return nil
}

func (r *PostgresPropertyRepository) Update(ctx context.Context, p *domain.Property) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, p.TenantID); err != nil {
		return err
	}

	p.UpdatedAt = time.Now().UTC()
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE properties SET name=$1, slug=$2, address=$3, timezone=$4, currency_code=$5, contact_email=$6, contact_phone=$7, is_active=$8, updated_at=$9, version=version+1
		WHERE id=$10 AND tenant_id=$11 AND version=$12 AND deleted_at IS NULL
	`, p.Name, p.Slug, p.Address, p.Timezone, p.CurrencyCode, p.ContactEmail, p.ContactPhone, p.IsActive, p.UpdatedAt, p.ID, p.TenantID, p.Version)
	if err != nil {
		return fmt.Errorf("update property: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("property update conflict: id=%s version=%d", p.ID, p.Version)
	}
	p.Version++
	return nil
}

func (r *PostgresPropertyRepository) SoftDelete(ctx context.Context, tenantID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `UPDATE properties SET deleted_at = NOW() WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`, id, tenantID)
	if err != nil {
		return fmt.Errorf("soft delete property: %w", err)
	}
	return nil
}

func (r *PostgresPropertyRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.Property, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 { limit = 100 }

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, slug, address, timezone, currency_code, contact_email, contact_phone, is_active, created_at, updated_at, version
		FROM properties WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list properties: %w", err)
	}
	defer rows.Close()

	return scanProperties(rows)
}

func scanProperties(rows pgx.Rows) ([]*domain.Property, error) {
	var props []*domain.Property
	for rows.Next() {
		var p domain.Property
		var address map[string]string
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.Slug, &address, &p.Timezone, &p.CurrencyCode, &p.ContactEmail, &p.ContactPhone, &p.IsActive, &p.CreatedAt, &p.UpdatedAt, &p.Version); err != nil {
			return nil, fmt.Errorf("scan property: %w", err)
		}
		p.Address = address
		props = append(props, &p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return props, nil
}

// ==================== ROOM TYPE ====================

type RoomTypeRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.RoomType, error)
	Create(ctx context.Context, rt *domain.RoomType) error
	Update(ctx context.Context, rt *domain.RoomType) error
	SoftDelete(ctx context.Context, tenantID, id string) error
	ListByProperty(ctx context.Context, tenantID, propertyID string, limit, offset int) ([]*domain.RoomType, error)
}

type PostgresRoomTypeRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewRoomTypeRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresRoomTypeRepository {
	return &PostgresRoomTypeRepository{pool: pool, metrics: metrics}
}

func (r *PostgresRoomTypeRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresRoomTypeRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.RoomType, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil { return nil, err }

	var rt domain.RoomType
	var amenities, images []string
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, code, name, description, base_occupancy, max_occupancy, amenities, images, is_active, created_at, updated_at, version
		FROM room_types WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID).Scan(
		&rt.ID, &rt.TenantID, &rt.PropertyID, &rt.Code, &rt.Name, &rt.Description, &rt.BaseOccupancy, &rt.MaxOccupancy, &amenities, &images, &rt.IsActive, &rt.CreatedAt, &rt.UpdatedAt, &rt.Version,
	)
	if err != nil { return nil, fmt.Errorf("get room type: %w", err) }
	rt.Amenities = amenities; rt.Images = images
	return &rt, nil
}

func (r *PostgresRoomTypeRepository) Create(ctx context.Context, rt *domain.RoomType) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, rt.TenantID); err != nil { return err }
	_, err := r.pool.Exec(ctx, `
		INSERT INTO room_types (id, tenant_id, property_id, code, name, description, base_occupancy, max_occupancy, amenities, images, is_active, created_at, updated_at, version)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, rt.ID, rt.TenantID, rt.PropertyID, rt.Code, rt.Name, rt.Description, rt.BaseOccupancy, rt.MaxOccupancy, rt.Amenities, rt.Images, rt.IsActive, rt.CreatedAt, rt.UpdatedAt, rt.Version)
	if err != nil { return fmt.Errorf("create room type: %w", err) }
	return nil
}

func (r *PostgresRoomTypeRepository) Update(ctx context.Context, rt *domain.RoomType) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, rt.TenantID); err != nil { return err }
	rt.UpdatedAt = time.Now().UTC()
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE room_types SET property_id=$1, code=$2, name=$3, description=$4, base_occupancy=$5, max_occupancy=$6, amenities=$7, images=$8, is_active=$9, updated_at=$10, version=version+1
		WHERE id=$11 AND tenant_id=$12 AND version=$13 AND deleted_at IS NULL
	`, rt.PropertyID, rt.Code, rt.Name, rt.Description, rt.BaseOccupancy, rt.MaxOccupancy, rt.Amenities, rt.Images, rt.IsActive, rt.UpdatedAt, rt.ID, rt.TenantID, rt.Version)
	if err != nil { return fmt.Errorf("update room type: %w", err) }
	if cmdTag.RowsAffected() == 0 { return fmt.Errorf("room type update conflict: id=%s version=%d", rt.ID, rt.Version) }
	rt.Version++
	return nil
}

func (r *PostgresRoomTypeRepository) SoftDelete(ctx context.Context, tenantID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil { return err }
	_, err := r.pool.Exec(ctx, `UPDATE room_types SET deleted_at = NOW() WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`, id, tenantID)
	if err != nil { return fmt.Errorf("soft delete room type: %w", err) }
	return nil
}

func (r *PostgresRoomTypeRepository) ListByProperty(ctx context.Context, tenantID, propertyID string, limit, offset int) ([]*domain.RoomType, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil { return nil, err }
	if limit <= 0 || limit > 100 { limit = 100 }

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, code, name, description, base_occupancy, max_occupancy, amenities, images, is_active, created_at, updated_at, version
		FROM room_types WHERE tenant_id = $1 AND property_id = $2 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $3 OFFSET $4
	`, tenantID, propertyID, limit, offset)
	if err != nil { return nil, fmt.Errorf("list room types: %w", err) }
	defer rows.Close()

	var rts []*domain.RoomType
	for rows.Next() {
		var rt domain.RoomType
		var amenities, images []string
		if err := rows.Scan(&rt.ID, &rt.TenantID, &rt.PropertyID, &rt.Code, &rt.Name, &rt.Description, &rt.BaseOccupancy, &rt.MaxOccupancy, &amenities, &images, &rt.IsActive, &rt.CreatedAt, &rt.UpdatedAt, &rt.Version); err != nil {
			return nil, fmt.Errorf("scan room type: %w", err)
		}
		rt.Amenities = amenities; rt.Images = images
		rts = append(rts, &rt)
	}
	if err := rows.Err(); err != nil { return nil, err }
	return rts, nil
}

// ==================== ROOM ====================

type RoomRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.Room, error)
	GetByNumber(ctx context.Context, tenantID, propertyID, roomNumber string) (*domain.Room, error)
	Create(ctx context.Context, room *domain.Room) error
	Update(ctx context.Context, room *domain.Room) error
	UpdateStatus(ctx context.Context, tenantID, id string, status domain.RoomStatus) error
	UpdateHousekeeping(ctx context.Context, tenantID, id string, status domain.HousekeepingStatus) error
	SoftDelete(ctx context.Context, tenantID, id string) error
	ListByProperty(ctx context.Context, tenantID, propertyID string, limit, offset int) ([]*domain.Room, error)
	ListByRoomType(ctx context.Context, tenantID, roomTypeID string, limit, offset int) ([]*domain.Room, error)
	ListAvailable(ctx context.Context, tenantID, propertyID string, checkIn, checkOut time.Time, limit, offset int) ([]*domain.Room, error)
}

type PostgresRoomRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewRoomRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresRoomRepository {
	return &PostgresRoomRepository{pool: pool, metrics: metrics}
}

func (r *PostgresRoomRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresRoomRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Room, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil { return nil, err }

	var room domain.Room
	var attributes []string
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, room_type_id, room_number, floor, status, is_smoking, has_ac, housekeeping_status, attributes, smart_lock_device_id, created_at, updated_at, version
		FROM rooms WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID).Scan(
		&room.ID, &room.TenantID, &room.PropertyID, &room.RoomTypeID, &room.RoomNumber, &room.Floor, &room.Status, &room.IsSmoking, &room.HasAC, &room.HousekeepingStatus, &attributes, &room.SmartLockDeviceID, &room.CreatedAt, &room.UpdatedAt, &room.Version,
	)
	if err != nil { return nil, fmt.Errorf("get room: %w", err) }
	room.Attributes = attributes
	return &room, nil
}

func (r *PostgresRoomRepository) GetByNumber(ctx context.Context, tenantID, propertyID, roomNumber string) (*domain.Room, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil { return nil, err }

	var room domain.Room
	var attributes []string
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, room_type_id, room_number, floor, status, is_smoking, has_ac, housekeeping_status, attributes, smart_lock_device_id, created_at, updated_at, version
		FROM rooms WHERE tenant_id = $1 AND property_id = $2 AND room_number = $3 AND deleted_at IS NULL
	`, tenantID, propertyID, roomNumber).Scan(
		&room.ID, &room.TenantID, &room.PropertyID, &room.RoomTypeID, &room.RoomNumber, &room.Floor, &room.Status, &room.IsSmoking, &room.HasAC, &room.HousekeepingStatus, &attributes, &room.SmartLockDeviceID, &room.CreatedAt, &room.UpdatedAt, &room.Version,
	)
	if err != nil { return nil, fmt.Errorf("get room by number: %w", err) }
	room.Attributes = attributes
	return &room, nil
}

func (r *PostgresRoomRepository) Create(ctx context.Context, room *domain.Room) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, room.TenantID); err != nil { return err }
	_, err := r.pool.Exec(ctx, `
		INSERT INTO rooms (id, tenant_id, property_id, room_type_id, room_number, floor, status, is_smoking, has_ac, housekeeping_status, attributes, smart_lock_device_id, created_at, updated_at, version)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
	`, room.ID, room.TenantID, room.PropertyID, room.RoomTypeID, room.RoomNumber, room.Floor, room.Status, room.IsSmoking, room.HasAC, room.HousekeepingStatus, room.Attributes, room.SmartLockDeviceID, room.CreatedAt, room.UpdatedAt, room.Version)
	if err != nil { return fmt.Errorf("create room: %w", err) }
	return nil
}

func (r *PostgresRoomRepository) Update(ctx context.Context, room *domain.Room) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, room.TenantID); err != nil { return err }
	room.UpdatedAt = time.Now().UTC()
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE rooms SET property_id=$1, room_type_id=$2, room_number=$3, floor=$4, status=$5, is_smoking=$6, has_ac=$7, housekeeping_status=$8, attributes=$9, smart_lock_device_id=$10, updated_at=$11, version=version+1
		WHERE id=$12 AND tenant_id=$13 AND version=$14 AND deleted_at IS NULL
	`, room.PropertyID, room.RoomTypeID, room.RoomNumber, room.Floor, room.Status, room.IsSmoking, room.HasAC, room.HousekeepingStatus, room.Attributes, room.SmartLockDeviceID, room.UpdatedAt, room.ID, room.TenantID, room.Version)
	if err != nil { return fmt.Errorf("update room: %w", err) }
	if cmdTag.RowsAffected() == 0 { return fmt.Errorf("room update conflict: id=%s version=%d", room.ID, room.Version) }
	room.Version++
	return nil
}

func (r *PostgresRoomRepository) UpdateStatus(ctx context.Context, tenantID, id string, status domain.RoomStatus) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil { return err }
	_, err := r.pool.Exec(ctx, `UPDATE rooms SET status = $1, updated_at = NOW(), version = version + 1 WHERE id = $2 AND tenant_id = $3 AND deleted_at IS NULL`, status, id, tenantID)
	if err != nil { return fmt.Errorf("update room status: %w", err) }
	return nil
}

func (r *PostgresRoomRepository) UpdateHousekeeping(ctx context.Context, tenantID, id string, status domain.HousekeepingStatus) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil { return err }
	_, err := r.pool.Exec(ctx, `UPDATE rooms SET housekeeping_status = $1, updated_at = NOW(), version = version + 1 WHERE id = $2 AND tenant_id = $3 AND deleted_at IS NULL`, status, id, tenantID)
	if err != nil { return fmt.Errorf("update housekeeping status: %w", err) }
	return nil
}

func (r *PostgresRoomRepository) SoftDelete(ctx context.Context, tenantID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil { return err }
	_, err := r.pool.Exec(ctx, `UPDATE rooms SET deleted_at = NOW() WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`, id, tenantID)
	if err != nil { return fmt.Errorf("soft delete room: %w", err) }
	return nil
}

func (r *PostgresRoomRepository) ListByProperty(ctx context.Context, tenantID, propertyID string, limit, offset int) ([]*domain.Room, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil { return nil, err }
	if limit <= 0 || limit > 100 { limit = 100 }
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, room_type_id, room_number, floor, status, is_smoking, has_ac, housekeeping_status, attributes, smart_lock_device_id, created_at, updated_at, version
		FROM rooms WHERE tenant_id = $1 AND property_id = $2 AND deleted_at IS NULL ORDER BY room_number LIMIT $3 OFFSET $4
	`, tenantID, propertyID, limit, offset)
	if err != nil { return nil, fmt.Errorf("list rooms: %w", err) }
	defer rows.Close()
	return scanRooms(rows)
}

func (r *PostgresRoomRepository) ListByRoomType(ctx context.Context, tenantID, roomTypeID string, limit, offset int) ([]*domain.Room, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil { return nil, err }
	if limit <= 0 || limit > 100 { limit = 100 }
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, room_type_id, room_number, floor, status, is_smoking, has_ac, housekeeping_status, attributes, smart_lock_device_id, created_at, updated_at, version
		FROM rooms WHERE tenant_id = $1 AND room_type_id = $2 AND deleted_at IS NULL ORDER BY room_number LIMIT $3 OFFSET $4
	`, tenantID, roomTypeID, limit, offset)
	if err != nil { return nil, fmt.Errorf("list rooms by type: %w", err) }
	defer rows.Close()
	return scanRooms(rows)
}

// ListAvailable returns rooms that are not occupied during the given date range.
// This is a simplified MVP version — production would use an occupancy cache or TimescaleDB continuous aggregate.
func (r *PostgresRoomRepository) ListAvailable(ctx context.Context, tenantID, propertyID string, checkIn, checkOut time.Time, limit, offset int) ([]*domain.Room, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil { return nil, err }
	if limit <= 0 || limit > 100 { limit = 100 }

	rows, err := r.pool.Query(ctx, `
		SELECT r.id, r.tenant_id, r.property_id, r.room_type_id, r.room_number, r.floor, r.status, r.is_smoking, r.has_ac, r.housekeeping_status, r.attributes, r.smart_lock_device_id, r.created_at, r.updated_at, r.version
		FROM rooms r
		WHERE r.tenant_id = $1 AND r.property_id = $2 AND r.deleted_at IS NULL
		  AND r.status = 'available'
		  AND r.id NOT IN (
			  SELECT room_id FROM reservations
			  WHERE tenant_id = $1 AND room_id IS NOT NULL AND deleted_at IS NULL
			    AND status IN ('confirmed', 'checked_in')
			    AND check_in_date < $4 AND check_out_date > $3
		  )
		ORDER BY r.room_number
		LIMIT $5 OFFSET $6
	`, tenantID, propertyID, checkIn, checkOut, limit, offset)
	if err != nil { return nil, fmt.Errorf("list available rooms: %w", err) }
	defer rows.Close()
	return scanRooms(rows)
}

func scanRooms(rows pgx.Rows) ([]*domain.Room, error) {
	var rooms []*domain.Room
	for rows.Next() {
		var room domain.Room
		var attributes []string
		if err := rows.Scan(&room.ID, &room.TenantID, &room.PropertyID, &room.RoomTypeID, &room.RoomNumber, &room.Floor, &room.Status, &room.IsSmoking, &room.HasAC, &room.HousekeepingStatus, &attributes, &room.SmartLockDeviceID, &room.CreatedAt, &room.UpdatedAt, &room.Version); err != nil {
			return nil, fmt.Errorf("scan room: %w", err)
		}
		room.Attributes = attributes
		rooms = append(rooms, &room)
	}
	if err := rows.Err(); err != nil { return nil, err }
	return rooms, nil
}
