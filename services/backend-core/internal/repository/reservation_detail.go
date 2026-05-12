package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// ReservationDetailRepository handles detailed reservation queries.
type ReservationDetailRepository struct{ pool *db.Pool }

func NewReservationDetailRepository(pool *db.Pool) *ReservationDetailRepository { return &ReservationDetailRepository{pool: pool} }

func (r *ReservationDetailRepository) GetDetail(ctx context.Context, tenantID, id uuid.UUID) (*domain.ReservationDetail, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT r.id, r.tenant_id, r.guest_name, r.email, r.phone, r.room_number, r.room_type, 
		       r.check_in, r.check_out, r.adults, r.children, r.status, r.source, 
		       r.total, r.balance, r.special_requests, r.vip, r.created_at, r.updated_at,
		       rm.floor, rm.bed_type, rm.rate_night
		FROM reservations r
		LEFT JOIN rooms rm ON rm.tenant_id = r.tenant_id AND rm.number = r.room_number
		WHERE r.tenant_id = $1 AND r.id = $2`, tenantID, id)

	var d domain.ReservationDetail
	var specReqs, floor, bedType *string
	var rateNight float64
	err := row.Scan(&d.ID, &d.TenantID, &d.Guest.Name, &d.Guest.Email, &d.Guest.Phone, &d.Room.RoomNumber, &d.Room.RoomType,
		&d.Dates.CheckIn, &d.Dates.CheckOut, &d.Party.Adults, &d.Party.Children, &d.Status, &d.Source,
		&d.Financials.Total, &d.Financials.Balance, &specReqs, &d.Guest.VIP, &d.CreatedAt, &d.UpdatedAt,
		&floor, &bedType, &rateNight)
	if err != nil { return nil, fmt.Errorf("get reservation detail: %w", err) }

	if specReqs != nil { d.SpecialRequests = *specReqs }
	if floor != nil { d.Room.Floor = *floor }
	if bedType != nil { d.Room.BedType = *bedType }
	d.Room.RateNight = rateNight
	d.Financials.Currency = "USD"
	d.Financials.Balance = d.Financials.Total - d.Financials.Paid

	// Calculate nights
	d.Room.TotalNights = int(d.Dates.CheckOut.Sub(d.Dates.CheckIn).Hours() / 24)
	if d.Room.TotalNights <= 0 { d.Room.TotalNights = 1 }

	return &d, nil
}

func (r *ReservationDetailRepository) ListByDateRange(ctx context.Context, tenantID uuid.UUID, from, to time.Time) ([]domain.ReservationDetail, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT r.id, r.tenant_id, r.guest_name, r.email, r.phone, r.room_number, r.room_type,
		       r.check_in, r.check_out, r.adults, r.children, r.status, r.source,
		       r.total, r.balance, r.vip, r.created_at, r.updated_at
		FROM reservations r
		WHERE r.tenant_id = $1 AND r.deleted_at IS NULL
		  AND (r.check_in BETWEEN $2 AND $3 OR r.check_out BETWEEN $2 AND $3 OR (r.check_in <= $2 AND r.check_out >= $3))
		ORDER BY r.check_in`, tenantID, from, to)
	if err != nil { return nil, fmt.Errorf("list reservations by date: %w", err) }
	defer rows.Close()

	var out []domain.ReservationDetail
	for rows.Next() {
		var d domain.ReservationDetail
		var roomNum *string
		err := rows.Scan(&d.ID, &d.TenantID, &d.Guest.Name, &d.Guest.Email, &d.Guest.Phone, &roomNum, &d.Room.RoomType,
			&d.Dates.CheckIn, &d.Dates.CheckOut, &d.Party.Adults, &d.Party.Children, &d.Status, &d.Source,
			&d.Financials.Total, &d.Financials.Balance, &d.Guest.VIP, &d.CreatedAt, &d.UpdatedAt)
		if err != nil { continue }
		if roomNum != nil { d.Room.RoomNumber = *roomNum }
		d.Room.TotalNights = int(d.Dates.CheckOut.Sub(d.Dates.CheckIn).Hours() / 24)
		if d.Room.TotalNights <= 0 { d.Room.TotalNights = 1 }
		out = append(out, d)
	}
	return out, nil
}

func (r *ReservationDetailRepository) GetRoomStatusView(ctx context.Context, tenantID uuid.UUID, date time.Time) ([]domain.RoomDailyStatus, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT rm.number, rm.type, rm.floor, rm.status, rm.bed_type, rm.rate_night,
		       r.id as res_id, r.guest_name, r.check_in, r.check_out
		FROM rooms rm
		LEFT JOIN reservations r ON r.tenant_id = rm.tenant_id AND r.room_number = rm.number
			AND r.deleted_at IS NULL AND r.status IN ('confirmed', 'checked_in')
			AND $1 BETWEEN r.check_in AND r.check_out
		WHERE rm.tenant_id = $2 AND rm.deleted_at IS NULL
		ORDER BY rm.number`, date, tenantID)
	if err != nil { return nil, fmt.Errorf("get room status view: %w", err) }
	defer rows.Close()

	var out []domain.RoomDailyStatus
	for rows.Next() {
		var s domain.RoomDailyStatus
		var resID, guestName *string
		var checkIn, checkOut *time.Time
		err := rows.Scan(&s.RoomNumber, &s.RoomType, &s.Status, &s.Status, &s.Status, &s.Rate,
			&resID, &guestName, &checkIn, &checkOut)
		if err != nil { continue }
		if resID != nil {
			s.ReservationID = *resID
			s.Status = "occupied"
		}
		if guestName != nil { s.GuestName = *guestName }
		if checkIn != nil && checkOut != nil {
			if date.Equal(*checkIn) { s.CheckIn = true }
			if date.Equal(*checkOut) { s.CheckOut = true }
		}
		out = append(out, s)
	}
	return out, nil
}
