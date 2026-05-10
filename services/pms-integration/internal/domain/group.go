package domain

import (
	"errors"
	"time"
)

// GroupReservation links multiple guest reservations under a master folio.
type GroupReservation struct {
	ID              string
	TenantID        string
	PropertyID      string
	MasterFolioID   string
	GroupName       string
	ContactName     string
	ContactEmail    string
	ContactPhone    string
	NumRooms        int
	NumGuests       int
	CheckIn         time.Time
	CheckOut        time.Time
	Status          string
	RateCode        string
	SpecialRequests string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// GroupReservationStatus values.
const (
	GroupStatusTentative  = "tentative"
	GroupStatusDefinite   = "definite"
	GroupStatusInHouse    = "in_house"
	GroupStatusCheckedOut = "checked_out"
	GroupStatusCancelled  = "cancelled"
)

// Validate checks a group reservation before creation.
func (gr *GroupReservation) Validate() error {
	if gr.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if gr.PropertyID == "" {
		return errors.New("property_id is required")
	}
	if gr.GroupName == "" {
		return errors.New("group_name is required")
	}
	if gr.CheckIn.IsZero() || gr.CheckOut.IsZero() {
		return errors.New("check_in and check_out are required")
	}
	if gr.CheckOut.Before(gr.CheckIn) {
		return errors.New("check_out must be after check_in")
	}
	if gr.NumRooms <= 0 {
		return errors.New("num_rooms must be greater than 0")
	}
	return nil
}
