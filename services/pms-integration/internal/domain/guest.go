package domain

import (
	"time"

	"github.com/google/uuid"
)

// GuestID is a typed identifier for guests.
type GuestID string

// Guest is the aggregate root for hospitality guests.
type Guest struct {
	ID          GuestID
	TenantID    string
	PropertyID  string
	FirstName   string
	LastName    string
	Email       string
	Phone       string
	LoyaltyID   string
	Preferences map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Version     int
}

// NewGuest creates a new guest aggregate.
func NewGuest(tenantID, propertyID, firstName, lastName, email string) *Guest {
	now := time.Now().UTC()
	return &Guest{
		ID:          GuestID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:    tenantID,
		PropertyID:  propertyID,
		FirstName:   firstName,
		LastName:    lastName,
		Email:       email,
		Preferences: make(map[string]string),
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	}
}

// UpdatePreferences replaces the guest's preferences.
func (g *Guest) UpdatePreferences(prefs map[string]string) {
	g.Preferences = prefs
	g.bumpVersion()
}

func (g *Guest) bumpVersion() {
	g.UpdatedAt = time.Now().UTC()
	g.Version++
}
