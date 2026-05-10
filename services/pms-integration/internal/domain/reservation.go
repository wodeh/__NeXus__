package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ReservationID is a typed identifier for reservations.
type ReservationID string

// Reservation is the aggregate root for guest room bookings.
type Reservation struct {
	ID              ReservationID
	TenantID        string
	PropertyID      string
	GuestID         string
	RoomID          string
// CheckInDate is the arrival date.
	CheckInDate  time.Time
	// CheckOutDate is the departure date.
	CheckOutDate time.Time
	Status          ReservationStatus
	SpecialRequests []string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	Version         int
}

// ReservationStatus represents the lifecycle state of a reservation.
type ReservationStatus string

const (
	ReservationStatusPending    ReservationStatus = "pending"
	ReservationStatusConfirmed  ReservationStatus = "confirmed"
	ReservationStatusCheckedIn  ReservationStatus = "checked_in"
	ReservationStatusCheckedOut ReservationStatus = "checked_out"
	ReservationStatusCancelled  ReservationStatus = "cancelled"
)

// NewReservation creates a new pending reservation.
func NewReservation(tenantID, propertyID, guestID, roomID string, checkIn, checkOut time.Time) *Reservation {
	now := time.Now().UTC()
	return &Reservation{
		ID:         ReservationID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:   tenantID,
		PropertyID: propertyID,
		GuestID:    guestID,
		RoomID:     roomID,
		CheckInDate:    checkIn,
		CheckOutDate:   checkOut,
		Status:     ReservationStatusPending,
		CreatedAt:  now,
		UpdatedAt:  now,
		Version:    1,
	}
}

// Confirm transitions a pending reservation to confirmed.
func (r *Reservation) Confirm() error {
	if r.Status != ReservationStatusPending {
		return fmt.Errorf("cannot confirm reservation in status %s", r.Status)
	}
	r.Status = ReservationStatusConfirmed
	r.bumpVersion()
	return nil
}

// DoCheckIn transitions a confirmed reservation to checked-in.
func (r *Reservation) DoCheckIn() error {
	if r.Status != ReservationStatusConfirmed {
		return fmt.Errorf("cannot check in reservation in status %s", r.Status)
	}
	r.Status = ReservationStatusCheckedIn
	r.bumpVersion()
	return nil
}

// DoCheckOut transitions a checked-in reservation to checked-out.
func (r *Reservation) DoCheckOut() error {
	if r.Status != ReservationStatusCheckedIn {
		return fmt.Errorf("cannot check out reservation in status %s", r.Status)
	}
	r.Status = ReservationStatusCheckedOut
	r.bumpVersion()
	return nil
}

// Cancel transitions a reservation to cancelled (idempotent if already cancelled).
func (r *Reservation) Cancel() error {
	if r.Status == ReservationStatusCancelled {
		return nil
	}
	if r.Status == ReservationStatusCheckedOut {
		return fmt.Errorf("cannot cancel checked-out reservation")
	}
	r.Status = ReservationStatusCancelled
	r.bumpVersion()
	return nil
}

func (r *Reservation) bumpVersion() {
	r.UpdatedAt = time.Now().UTC()
	r.Version++
}
