package domain

import (
	"time"
)

// MobileDevice tracks guest mobile app registrations.
type MobileDevice struct {
	ID           string
	TenantID     string
	GuestID      string
	DeviceToken  string // FCM/APNs token
	Platform     string // ios, android
	AppVersion   string
	OSVersion    string
	DeviceModel  string
	LastActiveAt time.Time
	CreatedAt    time.Time
}

// PushNotification is a message queued for delivery.
type PushNotification struct {
	ID        string
	TenantID  string
	GuestID   string
	DeviceID  string
	Title     string
	Body      string
	Data      map[string]interface{}
	SentAt    *time.Time
	ReadAt    *time.Time
	CreatedAt time.Time
}

// GuestSelfServiceRequest allows guests to manage their stay.
type GuestSelfServiceRequest struct {
	ID            string
	TenantID      string
	GuestID       string
	ReservationID string
	Type          string // late_checkout, room_upgrade, extra_towels, maintenance
	Status        string // pending, approved, rejected, completed
	Details       map[string]interface{}
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
