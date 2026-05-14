package domain

import (
	"time"

	"github.com/google/uuid"
)

// IPTVChannel represents a TV channel in the hospitality IPTV system.
type IPTVChannel struct {
	ID         uuid.UUID `json:"id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	Name       string    `json:"name"`
	Number     int       `json:"number"`
	StreamURL  string    `json:"stream_url"`
	LogoURL    *string   `json:"logo_url,omitempty"`
	Category   string    `json:"category"`
	Language   string    `json:"language"`
	IsActive   bool      `json:"is_active"`
	IsPremium  bool      `json:"is_premium"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// IPTVContent represents on-demand content (movies, info, music).
type IPTVContent struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	Title         string    `json:"title"`
	Type          string    `json:"type"` // movie, series, music, info
	Description   string    `json:"description"`
	Duration      *int      `json:"duration,omitempty"`
	ThumbnailURL  *string   `json:"thumbnail_url,omitempty"`
	Category      string    `json:"category"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// IPTVRoomStatus represents the IPTV status of a room.
type IPTVRoomStatus struct {
	ID              uuid.UUID `json:"id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	RoomID          uuid.UUID `json:"room_id"`
	RoomNumber      string    `json:"room_number"`
	IsOnline        bool      `json:"is_online"`
	CurrentChannel  *int      `json:"current_channel,omitempty"`
	LastActivityAt  *time.Time `json:"last_activity_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// IPTVWelcomeScreen holds data for the smart TV welcome display.
type IPTVWelcomeScreen struct {
	VillaID        string `json:"villa_id"`
	VillaName      string `json:"villa_name"`
	GuestName      string `json:"guest_name"`
	CheckInDate    string `json:"check_in_date"`
	CheckOutDate   string `json:"check_out_date"`
	Nights         int    `json:"nights"`
	WelcomeMessage string `json:"welcome_message"`
	HasReservation bool   `json:"has_reservation"`
}

// RoomServiceItem represents an item in a room service order.
type RoomServiceItem struct {
	MenuItemID uuid.UUID `json:"menu_item_id"`
	Name       string    `json:"name"`
	Quantity   int       `json:"quantity"`
	Price      float64   `json:"price"`
	Notes      string    `json:"notes,omitempty"`
}

// RoomServiceOrder represents a room service order placed from the TV.
type RoomServiceOrder struct {
	ID              uuid.UUID         `json:"id"`
	TenantID        uuid.UUID         `json:"tenant_id"`
	RoomID          uuid.UUID         `json:"room_id"`
	Items           []RoomServiceItem `json:"items"`
	Status          string            `json:"status"` // pending, preparing, delivered, cancelled
	SpecialRequests string            `json:"special_requests,omitempty"`
	BillToRoom      bool              `json:"bill_to_room"`
	TotalAmount     float64           `json:"total_amount"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
}

// GuestNotification represents a message sent to a guest's TV.
type GuestNotification struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	RoomID    uuid.UUID `json:"room_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"` // info, warning, urgent
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
