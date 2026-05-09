package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// FolioID is a typed identifier for folios.
type FolioID string

// Folio is the aggregate root for billing statements.
type Folio struct {
	ID            FolioID
	TenantID      string
	PropertyID    string
	ReservationID string
	GuestID       string
	Charges       []Charge
	Payments      []Payment
	Balance       float64
	Currency      string
	Status        FolioStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Version       int
}

// FolioStatus represents the billing lifecycle.
type FolioStatus string

const (
	FolioStatusOpen     FolioStatus = "open"
	FolioStatusClosed   FolioStatus = "closed"
	FolioStatusDisputed FolioStatus = "disputed"
)

// NewFolio creates a new open folio.
func NewFolio(tenantID, propertyID, reservationID, guestID, currency string) *Folio {
	now := time.Now().UTC()
	return &Folio{
		ID:            FolioID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:      tenantID,
		PropertyID:    propertyID,
		ReservationID: reservationID,
		GuestID:       guestID,
		Currency:      currency,
		Status:        FolioStatusOpen,
		Charges:       []Charge{},
		Payments:      []Payment{},
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	}
}

// AddCharge appends a charge and recalculates balance.
func (f *Folio) AddCharge(c Charge) error {
	if f.Status != FolioStatusOpen {
		return fmt.Errorf("cannot add charge to folio in status %s", f.Status)
	}
	f.Charges = append(f.Charges, c)
	f.Balance += c.Amount
	f.bumpVersion()
	return nil
}

// AddPayment appends a payment and recalculates balance.
func (f *Folio) AddPayment(p Payment) error {
	if f.Status != FolioStatusOpen {
		return fmt.Errorf("cannot add payment to folio in status %s", f.Status)
	}
	f.Payments = append(f.Payments, p)
	f.Balance -= p.Amount
	f.bumpVersion()
	return nil
}

// Close transitions the folio to closed.
func (f *Folio) Close() error {
	if f.Status != FolioStatusOpen {
		return fmt.Errorf("cannot close folio in status %s", f.Status)
	}
	f.Status = FolioStatusClosed
	f.bumpVersion()
	return nil
}

func (f *Folio) bumpVersion() {
	f.UpdatedAt = time.Now().UTC()
	f.Version++
}

// Charge represents a posted charge line item.
type Charge struct {
	ID          string
	Description string
	Amount      float64
	PostedAt    time.Time
}

// Payment represents a payment line item.
type Payment struct {
	ID       string
	Method   string
	Amount   float64
	PostedAt time.Time
}
