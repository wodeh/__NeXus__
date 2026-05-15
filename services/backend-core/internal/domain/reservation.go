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
	GroupID         *uuid.UUID `json:"group_id,omitempty"`
	GroupName       *string   `json:"group_name,omitempty"`
	PreArrivalReady bool       `json:"pre_arrival_ready"`
	DepositPaid     int        `json:"deposit_paid"`
	SpecialRequestsAcknowledged bool `json:"special_requests_acknowledged"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
	Version         int        `json:"version"`
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

// GroupCreateRequest creates multiple reservations under one group.
type GroupCreateRequest struct {
	GroupName string                    `json:"group_name"`
	Members   []ReservationCreateRequest `json:"members"`
}

// WaitlistEntry represents a guest waiting for a room assignment.
type WaitlistEntry struct {
	ID                uuid.UUID  `json:"id"`
	TenantID          uuid.UUID  `json:"tenant_id"`
	PropertyID        *uuid.UUID `json:"property_id,omitempty"`
	GuestName         string     `json:"guest_name"`
	Email             *string    `json:"email,omitempty"`
	Phone             *string    `json:"phone,omitempty"`
	Adults            int        `json:"adults"`
	Children          int        `json:"children"`
	RoomType          *string    `json:"room_type,omitempty"`
	RequestedCheckIn  string     `json:"requested_check_in"`
	RequestedCheckOut string     `json:"requested_check_out"`
	Priority          int        `json:"priority"`
	Notes             *string    `json:"notes,omitempty"`
	Status            string     `json:"status"`
	AssignedRoomNumber *string   `json:"assigned_room_number,omitempty"`
	AssignedReservationID *uuid.UUID `json:"assigned_reservation_id,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// WaitlistCreateRequest adds a guest to the waitlist.
type WaitlistCreateRequest struct {
	GuestName         string `json:"guest_name"`
	Email             string `json:"email,omitempty"`
	Phone             string `json:"phone,omitempty"`
	Adults            int    `json:"adults"`
	Children          int    `json:"children"`
	RoomType          string `json:"room_type,omitempty"`
	RequestedCheckIn  string `json:"requested_check_in"`
	RequestedCheckOut string `json:"requested_check_out"`
	Priority          int    `json:"priority"`
	Notes             string `json:"notes,omitempty"`
}

// WaitlistAssignRequest assigns a room to a waitlisted guest.
type WaitlistAssignRequest struct {
	RoomNumber string `json:"room_number"`
}

// ReservationMoveRequest moves a reservation to a new room and/or dates.
type ReservationMoveRequest struct {
	RoomNumber string `json:"room_number,omitempty"`
	CheckIn    string `json:"check_in,omitempty"`
	CheckOut   string `json:"check_out,omitempty"`
}

// RateRule represents a dynamic pricing rule.
type RateRule struct {
	ID                 uuid.UUID  `json:"id"`
	TenantID           uuid.UUID  `json:"tenant_id"`
	PropertyID         *uuid.UUID `json:"property_id,omitempty"`
	Name               string     `json:"name"`
	RoomType           *string    `json:"room_type,omitempty"`
	ConditionType      string     `json:"condition_type"`
	ConditionValue     string     `json:"condition_value"`
	RateAdjustmentType string     `json:"rate_adjustment_type"`
	RateAdjustmentValue int       `json:"rate_adjustment_value"`
	MinNights          int        `json:"min_nights"`
	MaxNights          *int       `json:"max_nights,omitempty"`
	StartDate          *string    `json:"start_date,omitempty"`
	EndDate            *string    `json:"end_date,omitempty"`
	Priority           int        `json:"priority"`
	Active             bool       `json:"active"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// RateRuleCreateRequest creates a rate rule.
type RateRuleCreateRequest struct {
	Name               string `json:"name"`
	RoomType           string `json:"room_type,omitempty"`
	ConditionType      string `json:"condition_type"`
	ConditionValue     string `json:"condition_value"`
	RateAdjustmentType string `json:"rate_adjustment_type"`
	RateAdjustmentValue int   `json:"rate_adjustment_value"`
	MinNights          int    `json:"min_nights"`
	MaxNights          int    `json:"max_nights,omitempty"`
	StartDate          string `json:"start_date,omitempty"`
	EndDate            string `json:"end_date,omitempty"`
	Priority           int    `json:"priority"`
}



// RateCalculateRequest computes a rate for a stay.
type RateCalculateRequest struct {
	RoomType string `json:"room_type"`
	CheckIn  string `json:"check_in"`
	CheckOut string `json:"check_out"`
}

// RateAdjustment shows a single rule's impact.
type RateAdjustment struct {
	RuleName string `json:"rule_name"`
	Type     string `json:"type"`
	Value    int    `json:"value"`
	Amount   int    `json:"amount"`
}

// RateCalculateResponse returns the computed rate.
type RateCalculateResponse struct {
	BaseRate    int              `json:"base_rate"`
	Nights      int              `json:"nights"`
	Adjustments []RateAdjustment `json:"adjustments"`
	FinalRate   int              `json:"final_rate"`
}

// GuestJourneyEvent represents a single touchpoint in the guest lifecycle.
type GuestJourneyEvent struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	ReservationID uuid.UUID `json:"reservation_id"`
	EventType     string    `json:"event_type"`
	EventData     map[string]interface{} `json:"event_data"`
	OccurredAt    time.Time `json:"occurred_at"`
	CreatedBy     string    `json:"created_by"`
	CreatedAt     time.Time `json:"created_at"`
}

// GuestJourneyTimeline returns all events for a reservation in chronological order.
type GuestJourneyTimeline struct {
	ReservationID uuid.UUID           `json:"reservation_id"`
	GuestName     string              `json:"guest_name"`
	Events        []GuestJourneyEvent `json:"events"`
}

// ReservationAssignRequest assigns a room to a reservation.
type ReservationAssignRequest struct {
	RoomNumber string `json:"room_number"`
}
