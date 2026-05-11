package domain

import "time"

// PromoCode for direct bookings.
type PromoCode struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	Code           string    `json:"code"`
	DiscountType   string    `json:"discount_type"` // percentage, fixed
	DiscountValue  float64   `json:"discount_value"`
	MaxUses        int       `json:"max_uses"`
	UsesCount      int       `json:"uses_count"`
	MinNights      int       `json:"min_nights"`
	ValidFrom      time.Time `json:"valid_from"`
	ValidUntil     time.Time `json:"valid_until"`
	ApplicableRoomTypes []string `json:"applicable_room_types,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// DirectBookingWidgetConfig for embeddable booking widget.
type DirectBookingWidgetConfig struct {
	TenantID           string    `json:"tenant_id"`
	Enabled            bool      `json:"enabled"`
	PropertyID         string    `json:"property_id"`
	ThemeColor         string    `json:"theme_color"`
	LogoURL            string    `json:"logo_url,omitempty"`
	Title              string    `json:"title"`
	Subtitle           string    `json:"subtitle"`
	ShowPromoCode      bool      `json:"show_promo_code"`
	ShowExtras         bool      `json:"show_extras"`
	UpsellEnabled      bool      `json:"upsell_enabled"`
	RequireDeposit     bool      `json:"require_deposit"`
	DepositPct         float64   `json:"deposit_pct"`
	SuccessRedirectURL string    `json:"success_redirect_url,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// BookingUpsell is an optional add-on during checkout.
type BookingUpsell struct {
	ID          string  `json:"id"`
	TenantID    string  `json:"tenant_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	PerNight    bool    `json:"per_night"`
	IsActive    bool    `json:"is_active"`
}

// DirectBookingSession tracks an in-progress booking.
type DirectBookingSession struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	PropertyID   string    `json:"property_id"`
	WidgetConfigID string  `json:"widget_config_id,omitempty"`
	GuestName    string    `json:"guest_name"`
	GuestEmail   string    `json:"guest_email"`
	GuestPhone   string    `json:"guest_phone"`
	RoomTypeID   string    `json:"room_type_id"`
	CheckIn      time.Time `json:"check_in"`
	CheckOut     time.Time `json:"check_out"`
	Nights       int       `json:"nights"`
	Adults       int       `json:"adults"`
	Children     int       `json:"children"`
	BaseRate     float64   `json:"base_rate"`
	UpsellsJSON  string    `json:"upsells_json,omitempty"`
	PromoCodeID  string    `json:"promo_code_id,omitempty"`
	Discount     float64   `json:"discount"`
	Total        float64   `json:"total"`
	DepositAmount float64  `json:"deposit_amount"`
	Currency     string    `json:"currency"`
	Status       string    `json:"status"` // pending, confirmed, abandoned, expired
	ExpiresAt    time.Time `json:"expires_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
