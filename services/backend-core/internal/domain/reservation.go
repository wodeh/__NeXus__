package domain

import (
	"time"

	"github.com/google/uuid"
)

// Reservation represents a guest booking.
type Reservation struct {
	ID              uuid.UUID `json:"id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	PropertyID      string    `json:"property_id"`
	GuestName       string    `json:"guest_name"`
	Email           *string   `json:"email,omitempty"`
	Phone           *string   `json:"phone,omitempty"`
	RoomNumber      *string   `json:"room_number,omitempty"`
	RoomType        string    `json:"room_type"`
	CheckIn         string    `json:"check_in"`  // YYYY-MM-DD
	CheckOut        string    `json:"check_out"` // YYYY-MM-DD
	Adults          int       `json:"adults"`
	Children        int       `json:"children"`
	Status          string    `json:"status"` // confirmed, checked_in, checked_out, cancelled, no_show
	Source          string    `json:"source"` // walk_in, ota, direct, agent
	Total           int       `json:"total"`
	Balance         int       `json:"balance"`
	SpecialRequests *string   `json:"special_requests,omitempty"`
	VIP             bool      `json:"vip"`
	Color           *string   `json:"color,omitempty"`
	Config          map[string]interface{} `json:"config,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	Version         int       `json:"version"`
}

// ReservationCreateRequest is the payload for creating a reservation.
type ReservationCreateRequest struct {
	GuestName       string `json:"guest_name"`
	Email           string `json:"email,omitempty"`
	Phone           string `json:"phone,omitempty"`
	RoomType        string `json:"room_type"`
	RoomNumber      string `json:"room_number,omitempty"`
	CheckIn         string `json:"check_in"`
	CheckOut        string `json:"check_out"`
	Adults          int    `json:"adults"`
	Children        int    `json:"children"`
	Source          string `json:"source"`
	SpecialRequests string `json:"special_requests,omitempty"`
	VIP             bool   `json:"vip"`
}

// ReservationAssignRequest assigns a room to a reservation.
type ReservationAssignRequest struct {
	RoomNumber string `json:"room_number"`
}
