package domain

import (
	"time"

	"github.com/google/uuid"
)

// Channel represents an OTA or booking channel connection.
type Channel struct {
	ID             uuid.UUID `json:"id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	Source         string    `json:"source"`          // booking_com, expedia, airbnb, direct, whatsapp
	DisplayName    string    `json:"display_name"`
	IsActive       bool      `json:"is_active"`
	CommissionPct  int       `json:"commission_pct"`
	LastSyncAt     *time.Time `json:"last_sync_at,omitempty"`
	LastSyncStatus string    `json:"last_sync_status"` // success, warning, error, n/a
	Config         map[string]interface{} `json:"config,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// ChannelCreateRequest is the payload for creating a channel.
type ChannelCreateRequest struct {
	Source        string `json:"source"`
	DisplayName   string `json:"display_name"`
	CommissionPct int    `json:"commission_pct"`
}

// ChannelUpdateRequest is the payload for updating a channel.
type ChannelUpdateRequest struct {
	DisplayName   string `json:"display_name,omitempty"`
	IsActive      *bool  `json:"is_active,omitempty"`
	CommissionPct *int   `json:"commission_pct,omitempty"`
}

// ChannelSyncLog represents a sync operation log.
type ChannelSyncLog struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	ChannelID uuid.UUID `json:"channel_id"`
	ChannelSource string `json:"channel_source"`
	Direction string    `json:"direction"` // pull, push
	Status    string    `json:"status"`    // success, partial, error
	Records   int       `json:"records"`
	Duration  string    `json:"duration"`
	Error     *string   `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// ChannelReservation represents a reservation that came through a channel.
type ChannelReservation struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	ChannelID     uuid.UUID `json:"channel_id"`
	ExternalRef   string    `json:"external_ref"`
	GuestName     string    `json:"guest_name"`
	RoomType      string    `json:"room_type"`
	CheckIn       string    `json:"check_in"`
	CheckOut      string    `json:"check_out"`
	Nights        int       `json:"nights"`
	Total         int       `json:"total"`
	Commission    int       `json:"commission"`
	NetAmount     int       `json:"net_amount"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
