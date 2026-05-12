package domain

import (
	"time"
)

// VillaProperty represents a bookable villa/chalet unit.
type VillaProperty struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Address         string    `json:"address"`
	City            string    `json:"city"`
	Country         string    `json:"country"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	Elevation       float64   `json:"elevation"` // meters above sea level (baseline for floor detection)
	Bedrooms        int       `json:"bedrooms"`
	Bathrooms       int       `json:"bathrooms"`
	MaxGuests       int       `json:"max_guests"`
	Amenities       []string  `json:"amenities"`
	Images          []string  `json:"images"`
	PricePerNight   float64   `json:"price_per_night"`
	Currency        string    `json:"currency"`
	CleaningFee     float64   `json:"cleaning_fee"`
	SecurityDeposit float64   `json:"security_deposit"`
	IsActive        bool      `json:"is_active"`
	Status          string    `json:"status"` // available, maintenance, blocked
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// VillaReservation is a day-based booking for a single villa.
type VillaReservation struct {
	ID              string           `json:"id"`
	TenantID        string           `json:"tenant_id"`
	VillaID         string           `json:"villa_id"`
	VillaName       string           `json:"villa_name"`
	GuestName       string           `json:"guest_name"`
	GuestPhone      string           `json:"guest_phone"`
	GuestEmail      string           `json:"guest_email"`
	GuestCount      int              `json:"guest_count"`
	CheckInDate     string           `json:"check_in_date"`  // YYYY-MM-DD
	CheckOutDate    string           `json:"check_out_date"` // YYYY-MM-DD
	Nights          int              `json:"nights"`
	TotalAmount     float64          `json:"total_amount"`
	Currency        string           `json:"currency"`
	Status          string           `json:"status"` // pending, reserved, cancelled, completed
	Source          string           `json:"source"` // phone, whatsapp, walkin
	InternalNotes   string           `json:"internal_notes"`
	DownPayment     *DownPayment     `json:"down_payment,omitempty"`
	BalanceDue      float64          `json:"balance_due"`
	BalancePaid     bool             `json:"balance_paid"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
	CompletedAt     *time.Time       `json:"completed_at,omitempty"`
}

// DownPayment tracks the deposit that confirms a reservation.
type DownPayment struct {
	Amount      float64   `json:"amount"`
	Method      string    `json:"method"` // visa, cash, bank_transfer, third_party
	Status      string    `json:"status"` // pending, received, refunded
	ReceivedAt  time.Time `json:"received_at"`
	Reference   string    `json:"reference"` // transaction ID, receipt number
	Notes       string    `json:"notes"`
}

// CleanerSensorLog captures phone sensor data from cleaning staff.
type CleanerSensorLog struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	VillaID      string    `json:"villa_id"`
	VillaName    string    `json:"villa_name"`
	CleanerID    string    `json:"cleaner_id"`
	CleanerName  string    `json:"cleaner_name"`
	Temperature  float64   `json:"temperature"`  // °C
	Latitude     float64   `json:"latitude"`
	Longitude    float64   `json:"longitude"`
	Altitude     float64   `json:"altitude"`     // meters above sea level
	Floor        int       `json:"floor"`        // derived from altitude - property elevation
	LocationType string    `json:"location_type"` // inside, outside, balcony, rooftop
	BatteryLevel float64   `json:"battery_level"` // 0-100
	RecordedAt   time.Time `json:"recorded_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// VillaRevenueStats aggregates financial data for villa owners.
type VillaRevenueStats struct {
	TenantID           string  `json:"tenant_id"`
	PeriodStart        string  `json:"period_start"`
	PeriodEnd          string  `json:"period_end"`
	TotalRevenue       float64 `json:"total_revenue"`
	TotalReservations  int     `json:"total_reservations"`
	OccupancyRate      float64 `json:"occupancy_rate"`
	AvgBookingValue    float64 `json:"avg_booking_value"`
	DownPaymentTotal   float64 `json:"down_payment_total"`
	PendingBalance     float64 `json:"pending_balance"`
	VillaBreakdown     []VillaRevenueBreakdown `json:"villa_breakdown"`
}

// VillaRevenueBreakdown per villa.
type VillaRevenueBreakdown struct {
	VillaID       string  `json:"villa_id"`
	VillaName     string  `json:"villa_name"`
	Revenue       float64 `json:"revenue"`
	NightsBooked  int     `json:"nights_booked"`
	OccupancyPct  float64 `json:"occupancy_pct"`
}

// VillaAvailability shows which dates are blocked/available for a villa.
type VillaAvailability struct {
	VillaID  string `json:"villa_id"`
	Date     string `json:"date"`     // YYYY-MM-DD
	Status   string `json:"status"`   // available, reserved, blocked, maintenance
	ReservationID *string `json:"reservation_id,omitempty"`
}
