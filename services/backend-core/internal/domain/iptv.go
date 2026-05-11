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
