package domain

import (
	"errors"
	"time"
)

// RoomBlockType defines why a room is blocked.
type RoomBlockType string

const (
	RoomBlockGroup       RoomBlockType = "group"
	RoomBlockVIP         RoomBlockType = "vip"
	RoomBlockMaintenance RoomBlockType = "maintenance"
	RoomBlockHold        RoomBlockType = "hold"
)

// RoomBlockStatus tracks whether a block is active or released.
type RoomBlockStatus string

const (
	RoomBlockActive   RoomBlockStatus = "active"
	RoomBlockReleased RoomBlockStatus = "released"
	RoomBlockExpired  RoomBlockStatus = "expired"
)

// RoomBlock holds a set of rooms out of inventory for a date range.
type RoomBlock struct {
	ID          string
	TenantID    string
	PropertyID  string
	RoomIDs     []string
	StartDate   time.Time
	EndDate     time.Time
	Reason      string
	Type        RoomBlockType
	Status      RoomBlockStatus
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Validate checks a room block before creation.
func (rb *RoomBlock) Validate() error {
	if rb.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if rb.PropertyID == "" {
		return errors.New("property_id is required")
	}
	if len(rb.RoomIDs) == 0 {
		return errors.New("at least one room must be blocked")
	}
	if rb.StartDate.IsZero() || rb.EndDate.IsZero() {
		return errors.New("start_date and end_date are required")
	}
	if rb.EndDate.Before(rb.StartDate) {
		return errors.New("end_date must be after start_date")
	}
	if rb.Reason == "" {
		return errors.New("reason is required")
	}
	if rb.Type == "" {
		return errors.New("block type is required")
	}
	return nil
}

// Overlaps checks if the block covers a given date.
func (rb *RoomBlock) Overlaps(date time.Time) bool {
	d := date.Truncate(24 * time.Hour)
	start := rb.StartDate.Truncate(24 * time.Hour)
	end := rb.EndDate.Truncate(24 * time.Hour)
	return !d.Before(start) && !d.After(end)
}
