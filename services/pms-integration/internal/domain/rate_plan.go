package domain

import (
	"errors"
	"time"
)

// SeasonalRate defines pricing for a date range per room type.
type SeasonalRate struct {
	RoomTypeID string    `json:"room_type_id"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	BasePrice  float64   `json:"base_price"`
}

// RateRestriction defines booking rules.
type RateRestriction struct {
	MinStay           int  `json:"min_stay"`
	MaxStay           int  `json:"max_stay"`
	ClosedToArrival   bool `json:"closed_to_arrival"`
	ClosedToDeparture bool `json:"closed_to_departure"`
}

// RatePlan holds pricing rules for a property.
type RatePlanSeasonal struct {
	ID            string
	TenantID      string
	PropertyID    string
	Name          string
	Code          string
	Description   string
	SeasonalRates []SeasonalRate
	Restrictions  RateRestriction
	IsActive      bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Validate checks a rate plan before creation.
func (rp *RatePlanSeasonal) Validate() error {
	if rp.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if rp.PropertyID == "" {
		return errors.New("property_id is required")
	}
	if rp.Name == "" {
		return errors.New("name is required")
	}
	if rp.Code == "" {
		return errors.New("code is required")
	}
	return nil
}

// GetRateForDate returns the applicable rate for a room type on a date.
func (rp *RatePlanSeasonal) GetRateForDate(roomTypeID string, date time.Time) float64 {
	for _, sr := range rp.SeasonalRates {
		if sr.RoomTypeID != roomTypeID {
			continue
		}
		d := date.Truncate(24 * time.Hour)
		start := sr.StartDate.Truncate(24 * time.Hour)
		end := sr.EndDate.Truncate(24 * time.Hour)
		if !d.Before(start) && !d.After(end) {
			return sr.BasePrice
		}
	}
	return 0
}
