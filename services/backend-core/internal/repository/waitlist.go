package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// WaitlistRepository provides waitlist data access.
type WaitlistRepository struct {
	pool *db.Pool
}

// NewWaitlistRepository creates a waitlist repository.
func NewWaitlistRepository(pool *db.Pool) *WaitlistRepository {
	return &WaitlistRepository{pool: pool}
}

// Create inserts a new waitlist entry.
func (r *WaitlistRepository) Create(ctx context.Context, tenantID string, req *domain.WaitlistCreateRequest) (*domain.WaitlistEntry, error) {
	var e domain.WaitlistEntry
	err := r.pool.QueryRow(ctx, `
		INSERT INTO waitlist (tenant_id, guest_name, email, phone, adults, children, room_type, requested_check_in, requested_check_out, priority, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, tenant_id, guest_name, email, phone, adults, children, room_type, requested_check_in::text, requested_check_out::text, priority, notes, status, created_at, updated_at
	`, tenantID, req.GuestName, strPtr(req.Email), strPtr(req.Phone), req.Adults, req.Children, strPtr(req.RoomType), req.RequestedCheckIn, req.RequestedCheckOut, req.Priority, strPtr(req.Notes)).Scan(
		&e.ID, &e.TenantID, &e.GuestName, &e.Email, &e.Phone, &e.Adults, &e.Children, &e.RoomType, &e.RequestedCheckIn, &e.RequestedCheckOut, &e.Priority, &e.Notes, &e.Status, &e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create waitlist: %w", err)
	}
	return &e, nil
}

// List returns waitlist entries for a tenant.
func (r *WaitlistRepository) List(ctx context.Context, tenantID string) ([]domain.WaitlistEntry, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, guest_name, email, phone, adults, children, room_type, requested_check_in::text, requested_check_out::text, priority, notes, status, assigned_room_number, assigned_reservation_id, created_at, updated_at
		FROM waitlist
		WHERE tenant_id = $1 AND status = 'waiting' AND deleted_at IS NULL
		ORDER BY priority DESC, created_at ASC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list waitlist: %w", err)
	}
	defer rows.Close()

	var entries []domain.WaitlistEntry
	for rows.Next() {
		var e domain.WaitlistEntry
		if err := rows.Scan(&e.ID, &e.TenantID, &e.GuestName, &e.Email, &e.Phone, &e.Adults, &e.Children, &e.RoomType, &e.RequestedCheckIn, &e.RequestedCheckOut, &e.Priority, &e.Notes, &e.Status, &e.AssignedRoomNumber, &e.AssignedReservationID, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan waitlist: %w", err)
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("waitlist rows: %w", err)
	}
	return entries, nil
}

// SuggestRooms adds available room suggestions to waitlist entries.
func (r *WaitlistRepository) SuggestRooms(ctx context.Context, tenantID string, entries []domain.WaitlistEntry) ([]domain.WaitlistEntry, error) {
	for i, e := range entries {
		var roomNumber string
		err := r.pool.QueryRow(ctx, `
			SELECT r.number
			FROM rooms r
			WHERE r.tenant_id = $1
			  AND ($2 IS NULL OR r.type = $2)
			  AND r.status = 'vacant_clean'
			  AND NOT EXISTS (
				SELECT 1 FROM reservations res
				WHERE res.tenant_id = r.tenant_id
				  AND res.room_number = r.number
				  AND res.deleted_at IS NULL
				  AND res.status NOT IN ('cancelled', 'checked_out')
				  AND res.check_in <= $3
				  AND res.check_out > $3
			  )
			ORDER BY r.floor, r.number
			LIMIT 1
		`, tenantID, e.RoomType, e.RequestedCheckIn).Scan(&roomNumber)
		if err == nil {
			entries[i].AssignedRoomNumber = &roomNumber
		}
	}
	return entries, nil
}

// AssignRoom assigns a room to a waitlist entry and creates a reservation.
func (r *WaitlistRepository) AssignRoom(ctx context.Context, tenantID string, waitlistID uuid.UUID, roomNumber string) (*domain.Reservation, error) {
	// Get waitlist entry
	var w domain.WaitlistEntry
	err := r.pool.QueryRow(ctx, `
		SELECT id, guest_name, email, phone, adults, children, room_type, requested_check_in::text, requested_check_out::text
		FROM waitlist
		WHERE id = $1 AND tenant_id = $2 AND status = 'waiting' AND deleted_at IS NULL
	`, waitlistID, tenantID).Scan(&w.ID, &w.GuestName, &w.Email, &w.Phone, &w.Adults, &w.Children, &w.RoomType, &w.RequestedCheckIn, &w.RequestedCheckOut)
	if err != nil {
		return nil, fmt.Errorf("get waitlist entry: %w", err)
	}

	// Create reservation
	resRepo := NewReservationRepository(r.pool)
	res, err := resRepo.Create(ctx, nil, tenantID, &domain.ReservationCreateRequest{
		GuestName:  w.GuestName,
		Email:      strDeref(w.Email),
		Phone:      strDeref(w.Phone),
		RoomNumber: roomNumber,
		RoomType:   strDeref(w.RoomType),
		CheckIn:    w.RequestedCheckIn,
		CheckOut:   w.RequestedCheckOut,
		Adults:     w.Adults,
		Children:   w.Children,
		Source:     "waitlist",
	})
	if err != nil {
		return nil, fmt.Errorf("create reservation from waitlist: %w", err)
	}

	// Update waitlist entry
	_, err = r.pool.Exec(ctx, `
		UPDATE waitlist SET status = 'assigned', assigned_room_number = $3, assigned_reservation_id = $4, updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2
	`, waitlistID, tenantID, roomNumber, res.ID)
	if err != nil {
		return nil, fmt.Errorf("update waitlist entry: %w", err)
	}

	return res, nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func strDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
