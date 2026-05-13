package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// ReservationRepository provides reservation data access.
type ReservationRepository struct {
	pool *db.Pool
}

// NewReservationRepository creates a reservation repository.
func NewReservationRepository(pool *db.Pool) *ReservationRepository {
	return &ReservationRepository{pool: pool}
}

func (r *ReservationRepository) execer(tx pgx.Tx) interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
} {
	if tx != nil {
		return tx
	}
	return r.pool
}

// List returns all active reservations for a tenant.
func (r *ReservationRepository) List(ctx context.Context, tenantID string) ([]domain.Reservation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, guest_name, email, phone, room_number, room_type,
			check_in::text, check_out::text, adults, children, status, source, total, balance,
			special_requests, vip, color, config, created_at, updated_at, deleted_at, version
		FROM reservations
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY check_in DESC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list reservations: %w", err)
	}
	defer rows.Close()

	var reservations []domain.Reservation
	for rows.Next() {
		var res domain.Reservation
		var email, phone, roomNumber, specialRequests, color *string
		var configJSON []byte
		if err := rows.Scan(
			&res.ID, &res.TenantID, &res.PropertyID, &res.GuestName, &email, &phone, &roomNumber,
			&res.RoomType, &res.CheckIn, &res.CheckOut, &res.Adults, &res.Children,
			&res.Status, &res.Source, &res.Total, &res.Balance,
			&specialRequests, &res.VIP, &color, &configJSON,
			&res.CreatedAt, &res.UpdatedAt, &res.DeletedAt, &res.Version,
		); err != nil {
			return nil, fmt.Errorf("scan reservation: %w", err)
		}
		res.Email = email
		res.Phone = phone
		res.RoomNumber = roomNumber
		res.SpecialRequests = specialRequests
		res.Color = color
		reservations = append(reservations, res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reservation rows: %w", err)
	}
	return reservations, nil
}

// Get returns a single reservation by ID.
func (r *ReservationRepository) Get(ctx context.Context, tenantID string, id uuid.UUID) (*domain.Reservation, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, guest_name, email, phone, room_number, room_type,
			check_in::text, check_out::text, adults, children, status, source, total, balance,
			special_requests, vip, color, config, created_at, updated_at, deleted_at, version
		FROM reservations
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID)

	var res domain.Reservation
	var email, phone, roomNumber, specialRequests, color *string
	var configJSON []byte
	if err := row.Scan(
		&res.ID, &res.TenantID, &res.PropertyID, &res.GuestName, &email, &phone, &roomNumber,
		&res.RoomType, &res.CheckIn, &res.CheckOut, &res.Adults, &res.Children,
		&res.Status, &res.Source, &res.Total, &res.Balance,
		&specialRequests, &res.VIP, &color, &configJSON,
		&res.CreatedAt, &res.UpdatedAt, &res.DeletedAt, &res.Version,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("reservation %s: %w", id, ErrNotFound)
		}
		return nil, fmt.Errorf("get reservation: %w", err)
	}
	res.Email = email
	res.Phone = phone
	res.RoomNumber = roomNumber
	res.SpecialRequests = specialRequests
	res.Color = color
	return &res, nil
}

// Create inserts a new reservation.
func (r *ReservationRepository) Create(ctx context.Context, tx pgx.Tx, tenantID string, req *domain.ReservationCreateRequest) (*domain.Reservation, error) {
	if req.GuestName == "" || req.CheckIn == "" || req.CheckOut == "" {
		return nil, fmt.Errorf("guest_name, check_in, check_out required")
	}

	res := &domain.Reservation{
		ID:         uuid.Must(uuid.NewRandom()),
		TenantID:   uuid.MustParse(tenantID),
		PropertyID: req.RoomNumber,
		GuestName:  req.GuestName,
		RoomType:   req.RoomType,
		CheckIn:    req.CheckIn,
		CheckOut:   req.CheckOut,
		Adults:     req.Adults,
		Children:   req.Children,
		Status:     "confirmed",
		Source:     req.Source,
		VIP:        req.VIP,
	}
	if req.Email != "" {
		res.Email = &req.Email
	}
	if req.Phone != "" {
		res.Phone = &req.Phone
	}
	if req.RoomNumber != "" {
		res.RoomNumber = &req.RoomNumber
	}
	if req.SpecialRequests != "" {
		res.SpecialRequests = &req.SpecialRequests
	}

	// Calculate total based on room type rate
	rate := 89
	switch req.RoomType {
	case "Deluxe", "Deluxe King":
		rate = 129
	case "Suite":
		rate = 229
	}
	days := diffDays(req.CheckIn, req.CheckOut)
	if days < 1 {
		days = 1
	}
	res.Total = days * rate
	res.Balance = res.Total

	configJSON := []byte("{}")
	if res.Config != nil {
		var err error
		configJSON, err = json.Marshal(res.Config)
		if err != nil {
			return nil, fmt.Errorf("marshal config: %w", err)
		}
	}

	execer := r.execer(tx)
	_, err := execer.Exec(ctx, `
		INSERT INTO reservations (
			id, tenant_id, property_id, guest_name, email, phone, room_number, room_type,
			check_in, check_out, adults, children, status, source, total, balance,
			special_requests, vip, color, config, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, NOW(), NOW(), 1)
	`, res.ID, res.TenantID, res.PropertyID, res.GuestName, res.Email, res.Phone, res.RoomNumber,
		res.RoomType, res.CheckIn, res.CheckOut, res.Adults, res.Children, res.Status, res.Source,
		res.Total, res.Balance, res.SpecialRequests, res.VIP, res.Color, configJSON)
	if err != nil {
		return nil, fmt.Errorf("insert reservation: %w", err)
	}
	return res, nil
}

// CheckIn updates reservation status to checked_in.
func (r *ReservationRepository) CheckIn(ctx context.Context, tenantID string, id uuid.UUID) error {
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE reservations
		SET status = 'checked_in', updated_at = NOW(), version = version + 1
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL AND status = 'confirmed'
	`, id, tenantID)
	if err != nil {
		return fmt.Errorf("checkin reservation: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("reservation %s: %w", id, ErrNotFound)
	}
	return nil
}

// CheckOut updates reservation status to checked_out.
func (r *ReservationRepository) CheckOut(ctx context.Context, tenantID string, id uuid.UUID) error {
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE reservations
		SET status = 'checked_out', updated_at = NOW(), version = version + 1
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL AND status = 'checked_in'
	`, id, tenantID)
	if err != nil {
		return fmt.Errorf("checkout reservation: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("reservation %s: %w", id, ErrNotFound)
	}
	return nil
}

// Cancel updates reservation status to cancelled.
func (r *ReservationRepository) Cancel(ctx context.Context, tenantID string, id uuid.UUID) error {
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE reservations
		SET status = 'cancelled', updated_at = NOW(), version = version + 1
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID)
	if err != nil {
		return fmt.Errorf("cancel reservation: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("reservation %s: %w", id, ErrNotFound)
	}
	return nil
}

// Restore updates reservation status back to confirmed.
func (r *ReservationRepository) Restore(ctx context.Context, tenantID string, id uuid.UUID) error {
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE reservations
		SET status = 'confirmed', updated_at = NOW(), version = version + 1
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL AND status = 'cancelled'
	`, id, tenantID)
	if err != nil {
		return fmt.Errorf("restore reservation: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("reservation %s: not found or not cancelled", id)
	}
	return nil
}

// AssignRoom assigns a room number to a reservation.
func (r *ReservationRepository) AssignRoom(ctx context.Context, tenantID string, id uuid.UUID, roomNumber string) error {
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE reservations
		SET room_number = $3, updated_at = NOW(), version = version + 1
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID, roomNumber)
	if err != nil {
		return fmt.Errorf("assign room: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("reservation %s: %w", id, ErrNotFound)
	}
	return nil
}

func diffDays(a, b string) int {
	// Simple day diff for YYYY-MM-DD strings
	// Parse manually to avoid timezone issues
	return 0 // Simplified; actual implementation would parse dates
}

// Move updates room number and/or dates for a reservation, and syncs room status.
func (r *ReservationRepository) Move(ctx context.Context, tenantID string, id uuid.UUID, req *domain.ReservationMoveRequest) (*domain.Reservation, error) {
	fmt.Printf("[BACKEND] Move called: tenant=%s id=%s req=%+v\n", tenantID, id, req)
	// Get current reservation to know old room
	var oldRoom string
	err := r.pool.QueryRow(ctx, `
		SELECT room_number FROM reservations
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`,
		id, tenantID).Scan(&oldRoom)
	if err != nil {
		fmt.Printf("[BACKEND] Move: fetch old room failed: %v\n", err)
		return nil, fmt.Errorf("move: fetch old room: %w", err)
	}
	fmt.Printf("[BACKEND] Move: oldRoom=%q\n", oldRoom)

	updates := []string{}
	args := []interface{}{id, tenantID}
	argIdx := 3

	if req.RoomNumber != "" {
		updates = append(updates, fmt.Sprintf("room_number = $%d", argIdx))
		args = append(args, req.RoomNumber)
		argIdx++
	}
	if req.CheckIn != "" {
		updates = append(updates, fmt.Sprintf("check_in = $%d", argIdx))
		args = append(args, req.CheckIn)
		argIdx++
	}
	if req.CheckOut != "" {
		updates = append(updates, fmt.Sprintf("check_out = $%d", argIdx))
		args = append(args, req.CheckOut)
		argIdx++
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	updates = append(updates, "updated_at = NOW(), version = version + 1")
	sql := fmt.Sprintf("UPDATE reservations SET %s WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL", joinUpdates(updates))
	fmt.Printf("[BACKEND] Move: SQL=%s args=%v\n", sql, args)

	cmdTag, err := r.pool.Exec(ctx, sql, args...)
	if err != nil {
		fmt.Printf("[BACKEND] Move: exec failed: %v\n", err)
		return nil, fmt.Errorf("move reservation: %w", err)
	}
	fmt.Printf("[BACKEND] Move: rows affected=%d\n", cmdTag.RowsAffected())
	if cmdTag.RowsAffected() == 0 {
		return nil, fmt.Errorf("reservation %s: %w", id, ErrNotFound)
	}

	// Sync room statuses: new room → occupied, old room → check if still occupied
	if req.RoomNumber != "" && req.RoomNumber != oldRoom {
		// New room becomes occupied
		_, _ = r.pool.Exec(ctx, `
			UPDATE rooms SET status = 'occupied', updated_at = NOW()
			WHERE tenant_id = $1 AND number = $2`,
			tenantID, req.RoomNumber)

		// Check if old room still has active reservations
		var count int
		_ = r.pool.QueryRow(ctx, `
			SELECT COUNT(*) FROM reservations
			WHERE tenant_id = $1 AND room_number = $2 AND deleted_at IS NULL
			  AND status NOT IN ('cancelled', 'checked_out')
			  AND check_out > CURRENT_DATE`,
			tenantID, oldRoom).Scan(&count)

		if count == 0 {
			_, _ = r.pool.Exec(ctx, `
				UPDATE rooms SET status = 'vacant_clean', updated_at = NOW()
				WHERE tenant_id = $1 AND number = $2`,
				tenantID, oldRoom)
		}
	}

	// Fetch updated reservation
	updated, err := r.Get(ctx, tenantID, id)
	if err != nil {
		fmt.Printf("[BACKEND] Move: fetch updated reservation failed: %v\n", err)
		return nil, fmt.Errorf("move: fetch updated reservation: %w", err)
	}
	fmt.Printf("[BACKEND] Move: returning updated reservation check_in=%s check_out=%s room=%s\n", updated.CheckIn, updated.CheckOut, *updated.RoomNumber)
	return updated, nil
}

// ErrNotFound is returned when a record is not found.
var ErrNotFound = fmt.Errorf("not found")
