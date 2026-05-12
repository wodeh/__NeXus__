package domain

import (
	"time"

	"github.com/google/uuid"
)

// Property represents a hotel property managed by a tenant.
type Property struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	Name        string                 `json:"name"`
	Address     string                 `json:"address"`
	City        string                 `json:"city"`
	Country     string                 `json:"country"`
	Phone       string                 `json:"phone"`
	Email       string                 `json:"email"`
	Timezone    string                 `json:"timezone"`
	Currency    string                 `json:"currency"`
	StarRating  int                    `json:"star_rating"`
	IsActive    bool                   `json:"is_active"`
	Config      map[string]interface{} `json:"config,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// PropertySummary is a lightweight view for listing.
type PropertySummary struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	City      string    `json:"city"`
	Country   string    `json:"country"`
	IsActive  bool      `json:"is_active"`
	RoomCount int       `json:"room_count"`
}
