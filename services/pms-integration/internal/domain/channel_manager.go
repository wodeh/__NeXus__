package domain

import "time"

// ChannelManagerSource represents an OTA or booking channel.
type ChannelManagerSource string

const (
	ChannelBookingCom   ChannelManagerSource = "booking.com"
	ChannelExpedia      ChannelManagerSource = "expedia"
	ChannelAirbnb       ChannelManagerSource = "airbnb"
	ChannelAgoda        ChannelManagerSource = "agoda"
	ChannelTripCom      ChannelManagerSource = "trip.com"
	ChannelDirect       ChannelManagerSource = "direct"
	ChannelPhone        ChannelManagerSource = "phone"
	ChannelEmail        ChannelManagerSource = "email"
	ChannelWalkIn       ChannelManagerSource = "walk_in"
	ChannelWhatsApp     ChannelManagerSource = "whatsapp"
)

// ChannelConnection represents a connected OTA channel for a tenant.
type ChannelConnection struct {
	ID           string               `json:"id"`
	TenantID     string               `json:"tenant_id"`
	Source       ChannelManagerSource `json:"source"`
	DisplayName  string               `json:"display_name"`
	APIKey       string               `json:"-"` // sensitive
	APISecret    string               `json:"-"` // sensitive
	PropertyID   string               `json:"property_id"` // external property ID on the channel
	IsActive     bool                 `json:"is_active"`
	CommissionPct float64             `json:"commission_pct"`
	LastSyncAt   *time.Time           `json:"last_sync_at"`
	LastSyncStatus string             `json:"last_sync_status"`
	LastSyncError  string             `json:"last_sync_error,omitempty"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

// ChannelReservation is a reservation that came from an OTA channel.
type ChannelReservation struct {
	ID             string               `json:"id"`
	TenantID       string               `json:"tenant_id"`
	ChannelID      string               `json:"channel_id"`
	Source         ChannelManagerSource `json:"source"`
	ExternalRef    string               `json:"external_ref"` // OTA confirmation number
	GuestName      string               `json:"guest_name"`
	GuestEmail     string               `json:"guest_email"`
	GuestPhone     string               `json:"guest_phone"`
	RoomTypeID     string               `json:"room_type_id"`
	RoomID         string               `json:"room_id,omitempty"`
	CheckIn        time.Time            `json:"check_in"`
	CheckOut       time.Time            `json:"check_out"`
	Nights         int                  `json:"nights"`
	Adults         int                  `json:"adults"`
	Children       int                  `json:"children"`
	TotalAmount    float64              `json:"total_amount"`
	Commission     float64              `json:"commission"`
	NetAmount      float64              `json:"net_amount"`
	Currency       string               `json:"currency"`
	Status         string               `json:"status"` // confirmed, cancelled, no_show
	SpecialRequests string              `json:"special_requests,omitempty"`
	RawPayload     string               `json:"-"` // original OTA payload
	MappedToReservationID string        `json:"mapped_to_reservation_id,omitempty"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

// ChannelHealth holds aggregated performance data per channel.
type ChannelHealth struct {
	ChannelID       string  `json:"channel_id"`
	Source          ChannelManagerSource `json:"source"`
	DisplayName     string  `json:"display_name"`
	Bookings30d     int     `json:"bookings_30d"`
	Revenue30d      float64 `json:"revenue_30d"`
	Commission30d   float64 `json:"commission_30d"`
	NetRevenue30d   float64 `json:"net_revenue_30d"`
	ADR             float64 `json:"adr"`
	ConversionRate  float64 `json:"conversion_rate"`
	CancellationRate float64 `json:"cancellation_rate"`
	IsHealthy       bool    `json:"is_healthy"`
}

// ChannelSyncLog tracks every sync attempt.
type ChannelSyncLog struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	ChannelID   string    `json:"channel_id"`
	Direction   string    `json:"direction"` // push, pull
	Status      string    `json:"status"`    // success, error, partial
	Records     int       `json:"records"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
	StartedAt   time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}
