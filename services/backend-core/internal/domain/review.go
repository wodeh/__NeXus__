package domain

import (
	"time"

	"github.com/google/uuid"
)

// Review represents a guest review/rating.
type Review struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	ReservationID uuid.UUID `json:"reservation_id"`
	GuestName     string    `json:"guest_name"`
	RoomNumber    string    `json:"room_number"`
	Rating        int       `json:"rating"`        // 1-5
	Cleanliness   int       `json:"cleanliness"`   // 1-5
	Service       int       `json:"service"`       // 1-5
	Location      int       `json:"location"`      // 1-5
	Value         int       `json:"value"`         // 1-5
	Comment       string    `json:"comment"`
	StaffReply    *string   `json:"staff_reply,omitempty"`
	RepliedAt     *time.Time `json:"replied_at,omitempty"`
	IsPublished   bool      `json:"is_published"`
	Source        string    `json:"source"` // direct, ota, email
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ReviewStats holds aggregated review metrics.
type ReviewStats struct {
	TenantID        string  `json:"tenant_id"`
	TotalReviews    int     `json:"total_reviews"`
	AverageRating   float64 `json:"average_rating"`
	AverageCleanliness float64 `json:"average_cleanliness"`
	AverageService  float64 `json:"average_service"`
	AverageLocation float64 `json:"average_location"`
	AverageValue    float64 `json:"average_value"`
	FiveStarCount   int     `json:"five_star_count"`
	FourStarCount   int     `json:"four_star_count"`
	ThreeStarCount  int     `json:"three_star_count"`
	TwoStarCount    int     `json:"two_star_count"`
	OneStarCount    int     `json:"one_star_count"`
}

// CreateReviewRequest creates a review.
type CreateReviewRequest struct {
	ReservationID string `json:"reservation_id"`
	Rating        int    `json:"rating"`
	Cleanliness   int    `json:"cleanliness,omitempty"`
	Service       int    `json:"service,omitempty"`
	Location      int    `json:"location,omitempty"`
	Value         int    `json:"value,omitempty"`
	Comment       string `json:"comment,omitempty"`
	Source        string `json:"source"`
}

// ReplyReviewRequest adds a staff reply.
type ReplyReviewRequest struct {
	StaffReply string `json:"staff_reply"`
}
