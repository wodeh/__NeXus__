package domain

import "time"

// ReviewChannel where the review was collected.
type ReviewChannel string

const (
	ReviewChannelInternal  ReviewChannel = "internal"
	ReviewChannelGoogle    ReviewChannel = "google"
	ReviewChannelBooking   ReviewChannel = "booking_com"
	ReviewChannelTripAdvisor ReviewChannel = "tripadvisor"
	ReviewChannelAirbnb    ReviewChannel = "airbnb"
	ReviewChannelExpedia   ReviewChannel = "expedia"
)

// GuestReview is a single review from a guest.
type GuestReview struct {
	ID             string        `json:"id"`
	TenantID       string        `json:"tenant_id"`
	GuestID        string        `json:"guest_id"`
	GuestName      string        `json:"guest_name"`
	ReservationID  string        `json:"reservation_id"`
	RoomNumber     string        `json:"room_number,omitempty"`
	Channel        ReviewChannel `json:"channel"`
	ExternalRef    string        `json:"external_ref,omitempty"`
	OverallRating  int           `json:"overall_rating"` // 1-5
	Cleanliness    int           `json:"cleanliness"`
	Service        int           `json:"service"`
	Location       int           `json:"location"`
	Value          int           `json:"value"`
	Amenities      int           `json:"amenities"`
	Comment        string        `json:"comment,omitempty"`
	IsPublished    bool          `json:"is_published"`
	StaffResponse  string        `json:"staff_response,omitempty"`
	RespondedBy    string        `json:"responded_by,omitempty"`
	RespondedAt    *time.Time    `json:"responded_at,omitempty"`
	SentAt         *time.Time    `json:"sent_at,omitempty"`
	CompletedAt    *time.Time    `json:"completed_at,omitempty"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// ReviewRequest is a scheduled or sent review solicitation.
type ReviewRequest struct {
	ID            string        `json:"id"`
	TenantID      string        `json:"tenant_id"`
	GuestID       string        `json:"guest_id"`
	ReservationID string        `json:"reservation_id"`
	Channel       ReviewChannel `json:"channel"`
	Status        string        `json:"status"` // scheduled, sent, opened, completed, bounced
	ScheduledAt   time.Time     `json:"scheduled_at"`
	SentAt        *time.Time    `json:"sent_at,omitempty"`
	OpenedAt      *time.Time    `json:"opened_at,omitempty"`
	CompletedAt   *time.Time    `json:"completed_at,omitempty"`
	BounceReason  string        `json:"bounce_reason,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
}

// RatingSnapshot holds aggregated scores.
type RatingSnapshot struct {
	TenantID         string  `json:"tenant_id"`
	Period           string  `json:"period"` // 7d, 30d, 90d, ytd
	TotalReviews     int     `json:"total_reviews"`
	AverageOverall   float64 `json:"average_overall"`
	AverageCleanliness float64 `json:"average_cleanliness"`
	AverageService   float64 `json:"average_service"`
	AverageLocation  float64 `json:"average_location"`
	AverageValue     float64 `json:"average_value"`
	AverageAmenities float64 `json:"average_amenities"`
	NPS              float64 `json:"nps"` // Net Promoter Score
	ResponseRate     float64 `json:"response_rate"`
	AvgResponseTime  float64 `json:"avg_response_time_hours"`
}
