package domain

import (
	"time"

	"github.com/google/uuid"
)

// ReservationDetail provides full reservation information for the detail page.
type ReservationDetail struct {
	ID               uuid.UUID              `json:"id"`
	TenantID         uuid.UUID              `json:"tenant_id"`
	Guest            GuestInfo              `json:"guest"`
	Room             RoomAssignment         `json:"room"`
	Dates            StayDates              `json:"dates"`
	Party            PartyInfo              `json:"party"`
	Financials       Financials             `json:"financials"`
	Status           string                 `json:"status"` // confirmed, checked_in, checked_out, cancelled, no_show
	Source           string                 `json:"source"` // walk_in, ota, direct, agent, email, whatsapp
	SpecialRequests  string                 `json:"special_requests,omitempty"`
	InternalNotes    string                 `json:"internal_notes,omitempty"`
	CommunicationLog []CommunicationEntry   `json:"communication_log,omitempty"`
	ActivityLog      []ActivityEntry        `json:"activity_log,omitempty"`
	Documents        []Document             `json:"documents,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
}

// GuestInfo contains guest contact and identification details.
type GuestInfo struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Address    string `json:"address,omitempty"`
	City       string `json:"city,omitempty"`
	Country    string `json:"country,omitempty"`
	IDType     string `json:"id_type,omitempty"` // passport, national_id, drivers_license
	IDNumber   string `json:"id_number,omitempty"`
	BirthDate  string `json:"birth_date,omitempty"`
	Nationality string `json:"nationality,omitempty"`
	VIP        bool   `json:"vip"`
}

// RoomAssignment contains room details for a reservation.
type RoomAssignment struct {
	RoomNumber   string  `json:"room_number"`
	RoomType     string  `json:"room_type"`
	Floor        string  `json:"floor,omitempty"`
	BedType      string  `json:"bed_type,omitempty"`
	RateNight    float64 `json:"rate_night"`
	TotalNights  int     `json:"total_nights"`
}

// StayDates contains check-in and check-out information.
type StayDates struct {
	CheckIn          time.Time `json:"check_in"`
	CheckOut         time.Time `json:"check_out"`
	ArrivalTime      string    `json:"arrival_time,omitempty"`
	DepartureTime    string    `json:"departure_time,omitempty"`
	LateCheckout     bool      `json:"late_checkout,omitempty"`
	EarlyCheckin     bool      `json:"early_checkin,omitempty"`
}

// PartyInfo contains guest party composition.
type PartyInfo struct {
	Adults   int `json:"adults"`
	Children int `json:"children"`
	Infants  int `json:"infants,omitempty"`
}

// Financials contains pricing and payment information.
type Financials struct {
	RoomTotal     float64 `json:"room_total"`
	ExtrasTotal   float64 `json:"extras_total,omitempty"`
	TaxTotal      float64 `json:"tax_total,omitempty"`
	Discount      float64 `json:"discount,omitempty"`
	Total         float64 `json:"total"`
	Paid          float64 `json:"paid,omitempty"`
	Balance       float64 `json:"balance"`
	Currency      string  `json:"currency"`
	DepositRequired float64 `json:"deposit_required,omitempty"`
	DepositPaid   float64 `json:"deposit_paid,omitempty"`
}

// CommunicationEntry represents a guest communication record.
type CommunicationEntry struct {
	ID        string    `json:"id"`
	Channel   string    `json:"channel"` // email, sms, whatsapp, phone
	Direction string    `json:"direction"` // inbound, outbound
	Content   string    `json:"content"`
	SentAt    time.Time `json:"sent_at"`
	Status    string    `json:"status"` // sent, delivered, read, failed
}

// ActivityEntry represents a reservation activity log entry.
type ActivityEntry struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	UserName  string    `json:"user_name"`
	Action    string    `json:"action"` // created, modified, checked_in, checked_out, cancelled, assigned_room
	Details   string    `json:"details,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Document represents an attached document.
type Document struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"` // id_scan, contract, invoice, receipt
	URL      string `json:"url"`
	UploadedAt time.Time `json:"uploaded_at"`
}

// BulkReservationRequest represents a request to create multiple reservations at once.
type BulkReservationRequest struct {
	Reservations []ReservationCreateRequest `json:"reservations"`
}

// RoomStatusView represents room status for a specific date range view.
type RoomStatusView struct {
	Date        time.Time          `json:"date"`
	Rooms       []RoomDailyStatus  `json:"rooms"`
	Occupancy   float64            `json:"occupancy"`
	Revenue     float64            `json:"revenue"`
	Arrivals    int                `json:"arrivals"`
	Departures  int                `json:"departures"`
	Stayovers   int                `json:"stayovers"`
}

// RoomDailyStatus represents a single room's status on a specific date.
type RoomDailyStatus struct {
	RoomNumber   string    `json:"room_number"`
	RoomType     string    `json:"room_type"`
	Status       string    `json:"status"` // vacant_clean, vacant_dirty, occupied, maintenance, blocked
	GuestName    string    `json:"guest_name,omitempty"`
	ReservationID string   `json:"reservation_id,omitempty"`
	CheckIn      bool      `json:"check_in,omitempty"`
	CheckOut     bool      `json:"check_out,omitempty"`
	Housekeeping string    `json:"housekeeping,omitempty"` // clean, dirty, inspected, in_progress
	Rate         float64   `json:"rate,omitempty"`
}
