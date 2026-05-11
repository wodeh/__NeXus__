// Package repository provides tenant-aware data access for PMS integration.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// TapechartEntry represents a room + its reservations for the tapechart view.
type TapechartEntry struct {
	Room         *domain.Room         `json:"room"`
	RoomTypeName string               `json:"room_type_name"`
	Reservations []TapechartReservation `json:"reservations"`
}

// TapechartReservation is a reservation flattened for tapechart display.
type TapechartReservation struct {
	ID              string    `json:"id"`
	GuestID         string    `json:"guest_id"`
	GuestName       string    `json:"guest_name"`
	CheckInDate     time.Time `json:"check_in_date"`
	CheckOutDate    time.Time `json:"check_out_date"`
	Status          string    `json:"status"`
	BalanceDue      float64   `json:"balance_due"`
	NumGuests       int       `json:"num_guests"`
	SpecialRequests []string  `json:"special_requests"`
}

// TapechartRepository provides tapechart data queries.
type TapechartRepository interface {
	GetTapechart(ctx context.Context, tenantID, propertyID string, startDate, endDate time.Time) ([]TapechartEntry, error)
}

// PostgresTapechartRepository is the PostgreSQL implementation.
type PostgresTapechartRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewTapechartRepository creates a new tapechart repository.
func NewTapechartRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresTapechartRepository {
	return &PostgresTapechartRepository{pool: pool, metrics: metrics}
}

func (r *PostgresTapechartRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// GetTapechart returns all rooms for a property with their reservations overlapping the date range.
func (r *PostgresTapechartRepository) GetTapechart(ctx context.Context, tenantID, propertyID string, startDate, endDate time.Time) ([]TapechartEntry, error) {
	start := time.Now()
	r.metrics.IncQuery("tapechart", "get")
	defer r.metrics.ObserveDuration("tapechart", "get", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("tapechart", "get", "tenant_bind")
		return nil, err
	}

	// 1. Get all rooms for the property with their room type names
	roomRows, err := r.pool.Query(ctx, `
		SELECT r.id, r.tenant_id, r.property_id, r.room_type_id, r.room_number, r.floor, r.status, r.is_smoking, r.has_ac, r.housekeeping_status, r.attributes, r.smart_lock_device_id, r.created_at, r.updated_at, r.version, COALESCE(rt.name, '') as room_type_name
		FROM rooms r
		LEFT JOIN room_types rt ON r.room_type_id = rt.id
		WHERE r.tenant_id = $1 AND r.property_id = $2 AND r.deleted_at IS NULL
		ORDER BY r.floor, r.room_number
	`, tenantID, propertyID)
	if err != nil {
		r.metrics.IncError("tapechart", "get", "rooms_query")
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer roomRows.Close()

	type roomWithType struct {
		room         *domain.Room
		roomTypeName string
	}
	var rooms []roomWithType
	for roomRows.Next() {
		var room domain.Room
		var attributes []string
		var roomTypeName string
		err := roomRows.Scan(
			&room.ID, &room.TenantID, &room.PropertyID, &room.RoomTypeID, &room.RoomNumber, &room.Floor,
			&room.Status, &room.IsSmoking, &room.HasAC, &room.HousekeepingStatus, &attributes,
			&room.SmartLockDeviceID, &room.CreatedAt, &room.UpdatedAt, &room.Version, &roomTypeName,
		)
		if err != nil {
			continue
		}
		room.Attributes = attributes
		rooms = append(rooms, roomWithType{room: &room, roomTypeName: roomTypeName})
	}
	roomRows.Close()

	// 2. Get reservations overlapping the date range
	resRows, err := r.pool.Query(ctx, `
		SELECT 
			res.id, res.guest_id, res.room_id, res.check_in_date, res.check_out_date, res.status, res.special_requests,
			COALESCE(g.first_name, '') || ' ' || COALESCE(g.last_name, '') as guest_name,
			COALESCE(g.num_adults, 1) + COALESCE(g.num_children, 0) as num_guests
		FROM reservations res
		LEFT JOIN guests g ON res.guest_id = g.id
		WHERE res.tenant_id = $1 AND res.property_id = $2 
			AND res.status NOT IN ('cancelled', 'checked_out')
			AND res.check_in_date <= $4 AND res.check_out_date >= $3
			AND res.deleted_at IS NULL
		ORDER BY res.check_in_date
	`, tenantID, propertyID, startDate, endDate)
	if err != nil {
		r.metrics.IncError("tapechart", "get", "res_query")
		return nil, fmt.Errorf("list reservations: %w", err)
	}
	defer resRows.Close()

	resByRoom := make(map[string][]TapechartReservation)
	for resRows.Next() {
		var tr TapechartReservation
		var roomID string
		var specialReq []string
		err := resRows.Scan(
			&tr.ID, &tr.GuestID, &roomID, &tr.CheckInDate, &tr.CheckOutDate,
			&tr.Status, &specialReq, &tr.GuestName, &tr.NumGuests,
		)
		if err != nil {
			continue
		}
		tr.SpecialRequests = specialReq
		resByRoom[roomID] = append(resByRoom[roomID], tr)
	}

	// 3. Build entries
	entries := make([]TapechartEntry, 0, len(rooms))
	for _, rwt := range rooms {
		entries = append(entries, TapechartEntry{
			Room:         rwt.room,
			RoomTypeName: rwt.roomTypeName,
			Reservations: resByRoom[string(rwt.room.ID)],
		})
	}

	return entries, nil
}
