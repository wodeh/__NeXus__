package domain

import (
	"time"

	"github.com/google/uuid"
)

// ChannelReservation represents a reservation originating from an external channel.
type ChannelReservation struct {
	ID             uuid.UUID              `json:"id"`
	TenantID       uuid.UUID              `json:"tenant_id"`
	ChannelSource  string                 `json:"channel_source"` // booking_com, expedia, airbnb, email, whatsapp, walk_in
	ExternalRef    string                 `json:"external_ref"`   // OTA confirmation number, email thread id, etc.
	GuestName      string                 `json:"guest_name"`
	GuestEmail     string                 `json:"guest_email,omitempty"`
	GuestPhone     string                 `json:"guest_phone,omitempty"`
	RoomType       string                 `json:"room_type"`
	RoomNumber     string                 `json:"room_number,omitempty"`
	CheckIn        time.Time              `json:"check_in"`
	CheckOut       time.Time              `json:"check_out"`
	Adults         int                    `json:"adults"`
	Children       int                    `json:"children"`
	Total          float64                `json:"total"`
	Currency       string                 `json:"currency"`
	Status         string                 `json:"status"` // pending, confirmed, cancelled, no_show
	SpecialRequests string                `json:"special_requests,omitempty"`
	RawPayload     map[string]interface{} `json:"raw_payload,omitempty"` // Original channel data
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// Channel represents an integrated OTA or booking source.
type Channel struct {
	ID             uuid.UUID              `json:"id"`
	TenantID       uuid.UUID              `json:"tenant_id"`
	Source         string                 `json:"source"`          // booking_com, expedia, airbnb
	DisplayName    string                 `json:"display_name"`
	IsActive       bool                   `json:"is_active"`
	CommissionPct  float64                `json:"commission_pct"`
	APIKey         string                 `json:"-"`             // excluded from JSON
	APISecret      string                 `json:"-"`             // excluded from JSON
	WebhookURL     string                 `json:"webhook_url,omitempty"`
	LastSyncAt     *time.Time             `json:"last_sync_at,omitempty"`
	LastSyncStatus string                 `json:"last_sync_status"` // success, warning, error, n/a
	Config         map[string]interface{} `json:"config,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// EmailReservationRequest represents a reservation parsed from an incoming email.
type EmailReservationRequest struct {
	From        string    `json:"from"`
	Subject     string    `json:"subject"`
	Body        string    `json:"body"`
	ReceivedAt  time.Time `json:"received_at"`
	Attachments []string  `json:"attachments,omitempty"`
}

// WhatsAppBotReservation represents a reservation initiated via WhatsApp bot.
type WhatsAppBotReservation struct {
	GuestPhone     string    `json:"guest_phone"`
	GuestName      string    `json:"guest_name"`
	RoomType       string    `json:"room_type"`
	CheckIn        string    `json:"check_in"` // YYYY-MM-DD
	CheckOut       string    `json:"check_out"` // YYYY-MM-DD
	Adults         int       `json:"adults"`
	Children       int       `json:"children"`
	SpecialRequests string   `json:"special_requests,omitempty"`
	ConversationID string   `json:"conversation_id"`
}

// FrontDeskWalkIn represents a walk-in reservation created at the front desk.
type FrontDeskWalkIn struct {
	GuestName       string  `json:"guest_name"`
	GuestEmail      string  `json:"guest_email,omitempty"`
	GuestPhone      string  `json:"guest_phone,omitempty"`
	RoomType        string  `json:"room_type"`
	RoomNumber      string  `json:"room_number,omitempty"`
	CheckIn         string  `json:"check_in"` // YYYY-MM-DD
	CheckOut        string  `json:"check_out"` // YYYY-MM-DD
	Adults          int     `json:"adults"`
	Children        int     `json:"children"`
	Total           float64 `json:"total"`
	Deposit         float64 `json:"deposit,omitempty"`
	PaymentMethod   string  `json:"payment_method"` // cash, card, transfer
	SpecialRequests string  `json:"special_requests,omitempty"`
	IDRequired      bool    `json:"id_required"`
	IDType          string  `json:"id_type,omitempty"`
	IDNumber        string  `json:"id_number,omitempty"`
}

// ChannelAvailability tracks real-time room availability per channel.
type ChannelAvailability struct {
	TenantID     uuid.UUID `json:"tenant_id"`
	RoomType     string    `json:"room_type"`
	Date         time.Time `json:"date"`
	TotalRooms   int       `json:"total_rooms"`
	BookedRooms  int       `json:"booked_rooms"`
	BlockedRooms int       `json:"blocked_rooms"`
	Available    int       `json:"available"`
	Rate         float64   `json:"rate"`
	Currency     string    `json:"currency"`
}
