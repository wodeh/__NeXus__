package domain

import (
	"time"
)

// Capability constants for IPTV
type IPTVCapability string

const (
	CapIPTVBasic    IPTVCapability = "iptv_basic"
	CapIPTVPremium  IPTVCapability = "iptv_premium"
	CapIPTVWelcome  IPTVCapability = "iptv_welcome"
	CapIPTVContent  IPTVCapability = "iptv_content"
)

// IPTVChannel represents a TV channel available on the system
type IPTVChannel struct {
	ID          string    `json:"id" db:"id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	PropertyID  string    `json:"property_id" db:"property_id"`
	Name        string    `json:"name" db:"name"`
	ChannelNum  int       `json:"channel_number" db:"channel_number"`
	Category    string    `json:"category" db:"category"`
	StreamURL   string    `json:"stream_url,omitempty" db:"stream_url"`
	IconURL     string    `json:"icon_url,omitempty" db:"icon_url"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	IsPremium   bool      `json:"is_premium" db:"is_premium"`
	Language    string    `json:"language" db:"language"`
	Metadata    JSONMap   `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// IPTVContent represents on-demand content (movies, info, etc)
type IPTVContent struct {
	ID          string    `json:"id" db:"id"`
	TenantID    string    `json:"tenant_id" db:"tenant_id"`
	PropertyID  string    `json:"property_id" db:"property_id"`
	Title       string    `json:"title" db:"title"`
	ContentType string    `json:"content_type" db:"content_type"`
	Category    string    `json:"category" db:"category"`
	Description string    `json:"description,omitempty" db:"description"`
	PosterURL   string    `json:"poster_url,omitempty" db:"poster_url"`
	StreamURL   string    `json:"stream_url,omitempty" db:"stream_url"`
	Duration    int       `json:"duration_minutes,omitempty" db:"duration_minutes"`
	IsPremium   bool      `json:"is_premium" db:"is_premium"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	Metadata    JSONMap   `json:"metadata,omitempty" db:"metadata"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// IPTVRoomBinding links a TV device to a room
type IPTVRoomBinding struct {
	ID                    string    `json:"id" db:"id"`
	TenantID              string    `json:"tenant_id" db:"tenant_id"`
	PropertyID            string    `json:"property_id" db:"property_id"`
	RoomID                string    `json:"room_id" db:"room_id"`
	DeviceID              string    `json:"device_id" db:"device_id"`
	DeviceType            string    `json:"device_type" db:"device_type"`
	WelcomeScreenEnabled  bool      `json:"welcome_screen_enabled" db:"welcome_screen_enabled"`
	WelcomeMessage        string    `json:"welcome_message,omitempty" db:"welcome_message"`
	GuestNameDisplay      bool      `json:"guest_name_display" db:"guest_name_display"`
	CheckoutReminderEnabled bool    `json:"checkout_reminder_enabled" db:"checkout_reminder_enabled"`
	LanguageOverride      string    `json:"language_override,omitempty" db:"language_override"`
	ChannelsEnabled       JSONMap   `json:"channels_enabled,omitempty" db:"channels_enabled"`
	ContentEnabled        JSONMap   `json:"content_enabled,omitempty" db:"content_enabled"`
	LastSyncAt            *time.Time `json:"last_sync_at,omitempty" db:"last_sync_at"`
	Status                string    `json:"status" db:"status"`
	Metadata              JSONMap   `json:"metadata,omitempty" db:"metadata"`
	CreatedAt             time.Time `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" db:"updated_at"`
}

// IPTVGuestSession tracks what a guest watched
type IPTVGuestSession struct {
	ID                    string     `json:"id" db:"id"`
	TenantID              string     `json:"tenant_id" db:"tenant_id"`
	PropertyID            string     `json:"property_id" db:"property_id"`
	RoomBindingID         string     `json:"room_binding_id" db:"room_binding_id"`
	ReservationID         *string    `json:"reservation_id,omitempty" db:"reservation_id"`
	GuestID               *string    `json:"guest_id,omitempty" db:"guest_id"`
	SessionStartedAt      time.Time  `json:"session_started_at" db:"session_started_at"`
	SessionEndedAt        *time.Time `json:"session_ended_at,omitempty" db:"session_ended_at"`
	ChannelWatchHistory   JSONMap    `json:"channel_watch_history,omitempty" db:"channel_watch_history"`
	ContentWatchHistory   JSONMap    `json:"content_watch_history,omitempty" db:"content_watch_history"`
	WelcomeShownAt        *time.Time `json:"welcome_shown_at,omitempty" db:"welcome_shown_at"`
	CheckoutReminderShownAt *time.Time `json:"checkout_reminder_shown_at,omitempty" db:"checkout_reminder_shown_at"`
	TotalWatchMinutes     int        `json:"total_watch_minutes" db:"total_watch_minutes"`
	Metadata              JSONMap    `json:"metadata,omitempty" db:"metadata"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

// IPTVWelcomeScreen represents the data needed to render a welcome screen
type IPTVWelcomeScreen struct {
	GuestName       string           `json:"guest_name,omitempty"`
	RoomNumber      string           `json:"room_number"`
	CheckInDate     string           `json:"check_in_date,omitempty"`
	CheckOutDate    string           `json:"check_out_date,omitempty"`
	Nights          int              `json:"nights,omitempty"`
	WelcomeMessage  string           `json:"welcome_message"`
	HotelName       string           `json:"hotel_name"`
	Channels        []IPTVChannel    `json:"channels"`
	Content         []IPTVContent    `json:"content"`
	Language        string           `json:"language"`
	Weather         *IPTVWeatherInfo `json:"weather,omitempty"`
	LocalTime       string           `json:"local_time"`
}

// IPTVWeatherInfo embedded in welcome screen
type IPTVWeatherInfo struct {
	Temp        int    `json:"temp"`
	Condition   string `json:"condition"`
	Icon        string `json:"icon"`
	High        int    `json:"high"`
	Low         int    `json:"low"`
}

// IPTVAnalytics aggregates viewing data
type IPTVAnalytics struct {
	TotalSessions        int            `json:"total_sessions"`
	TotalWatchMinutes    int            `json:"total_watch_minutes"`
	MostWatchedChannel   *IPTVChannel   `json:"most_watched_channel,omitempty"`
	MostWatchedContent   *IPTVContent   `json:"most_watched_content,omitempty"`
	TopCategories        []CategoryStat `json:"top_categories"`
	ActiveDevices        int            `json:"active_devices"`
	OfflineDevices       int            `json:"offline_devices"`
}

// CategoryStat for analytics
type CategoryStat struct {
	Category string `json:"category"`
	Views    int    `json:"views"`
	Minutes  int    `json:"minutes"`
}
