package domain

import (
	"errors"
	"time"
)

// CheckInStatus tracks the progress of guest arrival.
type CheckInStatus string

const (
	CheckInPending    CheckInStatus = "pending"
	CheckInProgress   CheckInStatus = "in_progress"
	CheckInComplete   CheckInStatus = "complete"
)

// CheckInWorkflow tracks each step of the check-in process.
type CheckInWorkflow struct {
	ID               string
	TenantID         string
	ReservationID    string
	GuestID          string
	RoomID           string
	Status           CheckInStatus
	IDVerified       bool
	PaymentCollected bool
	KeyIssued        bool
	RoomInspected    bool
	WelcomeSent      bool
	StartedAt        *time.Time
	CompletedAt      *time.Time
	Notes            string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Validate checks a check-in workflow.
func (ci *CheckInWorkflow) Validate() error {
	if ci.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if ci.ReservationID == "" {
		return errors.New("reservation_id is required")
	}
	return nil
}

// IsComplete returns true if all required steps are done.
func (ci *CheckInWorkflow) IsComplete() bool {
	return ci.IDVerified && ci.PaymentCollected && ci.KeyIssued && ci.RoomInspected
}

// CheckOutWorkflow tracks departure.
type CheckOutWorkflow struct {
	ID              string
	TenantID        string
	ReservationID   string
	GuestID         string
	RoomID          string
	Status          CheckInStatus
	BalanceSettled  bool
	KeyReturned     bool
	RoomInspected   bool
	FeedbackSent    bool
	CompletedAt     *time.Time
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
