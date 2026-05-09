package domain

import (
	"testing"
	"time"
)

func TestReservationLifecycle(t *testing.T) {
	r := NewReservation("tenant-1", "prop-1", "guest-1", "room-101",
		time.Now().Add(24*time.Hour),
		time.Now().Add(48*time.Hour),
	)

	if r.Status != ReservationStatusPending {
		t.Fatalf("expected pending, got %s", r.Status)
	}

	if err := r.Confirm(); err != nil {
		t.Fatalf("confirm failed: %v", err)
	}
	if r.Status != ReservationStatusConfirmed {
		t.Fatalf("expected confirmed, got %s", r.Status)
	}

	if err := r.CheckIn(); err != nil {
		t.Fatalf("check in failed: %v", err)
	}
	if r.Status != ReservationStatusCheckedIn {
		t.Fatalf("expected checked_in, got %s", r.Status)
	}

	if err := r.CheckOut(); err != nil {
		t.Fatalf("check out failed: %v", err)
	}
	if r.Status != ReservationStatusCheckedOut {
		t.Fatalf("expected checked_out, got %s", r.Status)
	}
}

func TestReservation_InvalidTransitions(t *testing.T) {
	r := NewReservation("tenant-1", "prop-1", "guest-1", "room-101",
		time.Now().Add(24*time.Hour),
		time.Now().Add(48*time.Hour),
	)

	if err := r.CheckIn(); err == nil {
		t.Fatal("expected error checking in pending reservation")
	}

	r.Status = ReservationStatusCheckedOut
	if err := r.Cancel(); err == nil {
		t.Fatal("expected error cancelling checked-out reservation")
	}
}
