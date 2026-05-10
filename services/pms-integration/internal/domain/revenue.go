package domain

import (
	"time"
)

// RevenueForecast predicts occupancy and ADR for future dates.
type RevenueForecast struct {
	ID           string
	TenantID     string
	PropertyID   string
	Date         time.Time
	PredictedOccupancy float64
	PredictedADR float64
	PredictedRevPAR float64
	Confidence   float64
	ModelVersion string
	CreatedAt    time.Time
}

// DynamicPricingRule adjusts rates based on demand signals.
type DynamicPricingRule struct {
	ID              string
	TenantID        string
	PropertyID      string
	RoomTypeID      string
	Name            string
	MinAdvanceDays  int      // Apply when booking X+ days in advance
	MaxAdvanceDays  int
	MinLOS          int      // Minimum length of stay
	LeadTimeDiscount float64 // % discount for early bookings
	LastMinutePremium float64 // % premium for bookings within X days
	OccupancyThreshold float64 // Trigger at this occupancy level
	OccupancyPremium   float64 // Premium when above threshold
	IsActive        bool
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// DemandSignal captures external demand indicators.
type DemandSignal struct {
	ID           string
	TenantID     string
	PropertyID   string
	Date         time.Time
	LocalEvents  []string // concerts, conferences, holidays
	CompetitorRates map[string]float64 // competitor name -> rate
	FlightSearchVolume int
	WeatherForecast string
	DemandScore  float64 // 0-1 aggregated score
	CreatedAt    time.Time
}

// PriceRecommendation is a suggested rate adjustment.
type PriceRecommendation struct {
	ID           string
	TenantID     string
	PropertyID   string
	RoomTypeID   string
	Date         time.Time
	CurrentRate  float64
	SuggestedRate float64
	ChangePercent float64
	Reason       string
	Confidence   float64
	Applied      bool
	CreatedAt    time.Time
}
