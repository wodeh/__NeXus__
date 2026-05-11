package domain

import (
	"time"

	"github.com/google/uuid"
)

// RevenueDashboard provides daily revenue and occupancy statistics.
type RevenueDashboard struct {
	TenantID         uuid.UUID         `json:"tenant_id"`
	Date             time.Time         `json:"date"`
	TotalRevenue     float64           `json:"total_revenue"`
	RoomRevenue      float64           `json:"room_revenue"`
	ExtraRevenue     float64           `json:"extra_revenue"`
	TaxRevenue       float64           `json:"tax_revenue"`
	TotalRooms       int               `json:"total_rooms"`
	OccupiedRooms    int               `json:"occupied_rooms"`
	AvailableRooms   int               `json:"available_rooms"`
	OccupancyRate    float64           `json:"occupancy_rate"`
	ADR              float64           `json:"adr"` // Average Daily Rate
	RevPAR           float64           `json:"revpar"` // Revenue Per Available Room
	Arrivals         int               `json:"arrivals"`
	Departures       int               `json:"departures"`
	Stayovers        int               `json:"stayovers"`
	WalkIns          int               `json:"walk_ins"`
	NoShows          int               `json:"no_shows"`
	Cancellations    int               `json:"cancellations"`
	CreatedAt        time.Time         `json:"created_at"`
}

// RevenueForecast provides projected revenue for future dates.
type RevenueForecast struct {
	Date            time.Time `json:"date"`
	ProjectedRevenue float64  `json:"projected_revenue"`
	ProjectedOccupancy float64 `json:"projected_occupancy"`
	Confidence      float64   `json:"confidence"` // 0.0 - 1.0
	BookedRooms     int       `json:"booked_rooms"`
	TotalRooms      int       `json:"total_rooms"`
}

// PricingRule represents a dynamic pricing rule.
type PricingRule struct {
	ID             uuid.UUID              `json:"id"`
	TenantID       uuid.UUID              `json:"tenant_id"`
	Name           string                 `json:"name"`
	RoomType       string                 `json:"room_type"`
	Condition      string                 `json:"condition"` // occupancy_based, advance_booking, length_of_stay, seasonal
	TriggerValue   float64                `json:"trigger_value"`
	AdjustmentType string                 `json:"adjustment_type"` // percentage, fixed_amount
	AdjustmentValue float64               `json:"adjustment_value"`
	MinRate        float64                `json:"min_rate"`
	MaxRate        float64                `json:"max_rate"`
	IsActive       bool                   `json:"is_active"`
	Priority       int                    `json:"priority"`
	Config         map[string]interface{} `json:"config,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// ChannelRevenueBreakdown shows revenue by booking source.
type ChannelRevenueBreakdown struct {
	Channel       string  `json:"channel"`
	Bookings      int     `json:"bookings"`
	Revenue       float64 `json:"revenue"`
	Commission    float64 `json:"commission"`
	NetRevenue    float64 `json:"net_revenue"`
	AvgRate       float64 `json:"avg_rate"`
}

// RoomTypeRevenue shows revenue by room type.
type RoomTypeRevenue struct {
	RoomType      string  `json:"room_type"`
	NightsSold    int     `json:"nights_sold"`
	Revenue       float64 `json:"revenue"`
	AvgRate       float64 `json:"avg_rate"`
	OccupancyPct  float64 `json:"occupancy_pct"`
}

// RevenueComparison compares current period to previous.
type RevenueComparison struct {
	Period          string  `json:"period"` // day, week, month, year
	CurrentRevenue  float64 `json:"current_revenue"`
	PreviousRevenue float64 `json:"previous_revenue"`
	ChangePct       float64 `json:"change_pct"`
	CurrentOccupancy float64 `json:"current_occupancy"`
	PreviousOccupancy float64 `json:"previous_occupancy"`
}
