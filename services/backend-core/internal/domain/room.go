package domain

import (
	"time"

	"github.com/google/uuid"
)

// Room represents a physical hotel room.
type Room struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	PropertyID string    `json:"property_id"`
	Number     string    `json:"number"`
	Type       string    `json:"type"`
	Floor      string    `json:"floor"`
	BedType    *string   `json:"bed_type,omitempty"`
	Status     string    `json:"status"` // vacant_clean, vacant_dirty, occupied, blocked, maintenance
	RateNight  int       `json:"rate_night"`
	Config     map[string]interface{} `json:"config,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	DeletedAt  *time.Time `json:"deleted_at,omitempty"`
	Version    int       `json:"version"`
}

// RoomStatusUpdateRequest updates a room's status.
type RoomStatusUpdateRequest struct {
	Status string `json:"status"`
}
