package domain

import (
	"time"
)

// GuestProfile extends the base Guest with CRM fields.
type GuestProfile struct {
	ID                string
	TenantID          string
	FirstName         string
	LastName          string
	Email             string
	Phone             string
	VIPStatus         string
	LoyaltyTier       string
	TotalStays        int
	TotalNights       int
	TotalRevenue      float64
	AverageDailyRate  float64
	LastStayDate      *time.Time
	Preferences       GuestPreferences
	CommunicationLog  []CommunicationEntry
	Notes             string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// GuestPreferences tracks recurring requests.
type GuestPreferences struct {
	RoomType     string   `json:"room_type,omitempty"`
	Floor        string   `json:"floor,omitempty"`
	Smoking      bool     `json:"smoking,omitempty"`
	BedType      string   `json:"bed_type,omitempty"`
	Amenities    []string `json:"amenities,omitempty"`
	Dietary      string   `json:"dietary,omitempty"`
	SpecialRequests string `json:"special_requests,omitempty"`
}

// CommunicationEntry logs outreach to the guest.
type CommunicationEntry struct {
	ID        string
	Channel   string // email, sms, call
	Direction string // inbound, outbound
	Subject   string
	Body      string
	SentAt    time.Time
}

// VIP status tiers.
const (
	VIPNone   = "none"
	VIPSilver = "silver"
	VIPGold   = "gold"
	VIPPlatinum = "platinum"
)
