package domain

import (
	"time"
)

// GuestJourney defines an automated sequence of touchpoints
// triggered by reservation lifecycle events.
type GuestJourney struct {
	ID          string     `json:"id"`
	TenantID    string     `json:"tenant_id"`
	Name        string     `json:"name"`        // e.g. "Standard Check-in Journey"
	Trigger     string     `json:"trigger"`     // booking_confirmed, check_in, check_out, no_show
	IsActive    bool       `json:"is_active"`
	Steps       []JourneyStep `json:"steps"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type JourneyStep struct {
	ID            string    `json:"id"`
	DelayHours    int       `json:"delay_hours"`     // hours after trigger to send
	Channel       string    `json:"channel"`         // whatsapp, email, sms
	TemplateID    string    `json:"template_id"`     // references template in comms
	UpsellOfferID *string   `json:"upsell_offer_id,omitempty"`
	Condition     string    `json:"condition"`       // always, vip_only, first_time, returning
	IsActive      bool      `json:"is_active"`
}

// GuestJourneyExecution tracks a single reservation's progress through a journey
type GuestJourneyExecution struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	JourneyID       string    `json:"journey_id"`
	ReservationID   string    `json:"reservation_id"`
	GuestPhone      string    `json:"guest_phone"`
	CurrentStep     int       `json:"current_step"`
	TotalSteps      int       `json:"total_steps"`
	Status          string    `json:"status"`        // running, completed, cancelled, failed
	StartedAt       time.Time `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	NextTriggerAt   *time.Time `json:"next_trigger_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UpsellOffer is a sellable add-on presented to guests
type UpsellOffer struct {
	ID            string  `json:"id"`
	TenantID      string  `json:"tenant_id"`
	Name          string  `json:"name"`          // "Late Checkout", "Room Upgrade"
	Description   string  `json:"description"`
	Category      string  `json:"category"`      // room_upgrade, late_checkout, early_checkin, breakfast, spa, parking
	Price         float64 `json:"price"`
	Currency      string  `json:"currency"`
	ImageURL      string  `json:"image_url,omitempty"`
	IsActive      bool    `json:"is_active"`
	AutoOffer     bool    `json:"auto_offer"`    // include in AI journey sequences
	Conditions    map[string]interface{} `json:"conditions,omitempty"`
	// conditions: min_nights, max_nights, room_types[], vip_only, first_time_only
	DisplayOrder  int     `json:"display_order"`
	ConversionRate float64 `json:"conversion_rate"` // computed field
	TotalSold     int     `json:"total_sold"`      // computed field
	RevenueGenerated float64 `json:"revenue_generated"` // computed field
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UpsellPurchase records a guest buying an upsell
type UpsellPurchase struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	ReservationID   string    `json:"reservation_id"`
	GuestPhone      string    `json:"guest_phone"`
	OfferID         string    `json:"offer_id"`
	OfferName       string    `json:"offer_name"`
	Price           float64   `json:"price"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`        // pending, confirmed, cancelled, refunded
	PaymentMethod   string    `json:"payment_method"` // on_bill, credit_card, whatsapp_pay
	FolioPosted     bool      `json:"folio_posted"`
	JourneyStepID   *string   `json:"journey_step_id,omitempty"` // attribution
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CompetitorRate tracks a competitor hotel's pricing
type CompetitorRate struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	CompetitorID  string    `json:"competitor_id"`
	CompetitorName string   `json:"competitor_name"`
	RoomType      string    `json:"room_type"`
	Date          string    `json:"date"`          // YYYY-MM-DD
	Rate          float64   `json:"rate"`
	Currency      string    `json:"currency"`
	Availability  int       `json:"availability"`  // rooms available
	MinStay       int       `json:"min_stay"`
	IsPromo       bool      `json:"is_promo"`
	Source        string    `json:"source"`        // booking_com, expedia, scraped
	ScrapedAt     time.Time `json:"scraped_at"`
	CreatedAt     time.Time `json:"created_at"`
}

// CompetitorHotel is a hotel we track
type CompetitorHotel struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Name        string    `json:"name"`
	Address     string    `json:"address,omitempty"`
	City        string    `json:"city,omitempty"`
	Country     string    `json:"country,omitempty"`
	StarRating  int       `json:"star_rating,omitempty"`
	RoomCount   int       `json:"room_count,omitempty"`
	Website     string    `json:"website,omitempty"`
	BookingURL  string    `json:"booking_url,omitempty"`
	IsActive    bool      `json:"is_active"`
	LastScraped *time.Time `json:"last_scraped,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RateRecommendation is an AI-generated pricing suggestion
type RateRecommendation struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	RoomType        string    `json:"room_type"`
	Date            string    `json:"date"`
	CurrentRate     float64   `json:"current_rate"`
	RecommendedRate float64   `json:"recommended_rate"`
	Confidence      float64   `json:"confidence"`      // 0-1
	Reason          string    `json:"reason"`          // "Competitor dropped rate by 15%"
	Factors         []string  `json:"factors"`         // ["competitor_rate_drop", "low_occupancy", "local_event"]
	Applied         bool      `json:"applied"`
	AppliedAt       *time.Time `json:"applied_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// RateShopConfig controls how often we scrape competitors
type RateShopConfig struct {
	TenantID        string `json:"tenant_id"`
	Enabled         bool   `json:"enabled"`
	FrequencyHours  int    `json:"frequency_hours"` // e.g. 6 = check every 6 hours
	LookaheadDays   int    `json:"lookahead_days"`  // how many days forward to shop
	AutoAdjust      bool   `json:"auto_adjust"`     // automatically apply recommendations?
	MaxAdjustmentPct float64 `json:"max_adjustment_pct"` // e.g. 0.30 = max 30% change
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
