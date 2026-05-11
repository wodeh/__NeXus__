package repository

import (
	"context"
	"time"

	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ReviewsRepository handles review data.
type ReviewsRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewReviewsRepository creates a new repository.
func NewReviewsRepository(pool *db.Pool, metrics *RepositoryMetrics) *ReviewsRepository {
	return &ReviewsRepository{pool: pool, metrics: metrics}
}

func (r *ReviewsRepository) CreateReview(ctx context.Context, tenantID string, rev *domain.GuestReview) (*domain.GuestReview, error) {
	r.pool.SetTenant(ctx, tenantID)
	query := `
		INSERT INTO guest_reviews (tenant_id, guest_id, guest_name, reservation_id, room_number, channel, external_ref, overall_rating, cleanliness, service, location, value, amenities, comment, is_published, sent_at, completed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	err := r.pool.QueryRow(ctx, query, tenantID, rev.GuestID, rev.GuestName, rev.ReservationID, rev.RoomNumber, rev.Channel, rev.ExternalRef, rev.OverallRating, rev.Cleanliness, rev.Service, rev.Location, rev.Value, rev.Amenities, rev.Comment, rev.IsPublished, rev.SentAt, rev.CompletedAt).Scan(
		&rev.ID, &rev.CreatedAt, &rev.UpdatedAt)
	rev.TenantID = tenantID
	return rev, err
}

func (r *ReviewsRepository) ListReviews(ctx context.Context, tenantID string, channel *domain.ReviewChannel, limit int) ([]*domain.GuestReview, error) {
	r.pool.SetTenant(ctx, tenantID)
	query := `SELECT id, tenant_id, guest_id, guest_name, reservation_id, room_number, channel, external_ref, overall_rating, cleanliness, service, location, value, amenities, comment, is_published, staff_response, responded_by, responded_at, sent_at, completed_at, created_at, updated_at FROM guest_reviews WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	if channel != nil {
		query += ` AND channel = $2`
		args = append(args, *channel)
	}
	query += ` ORDER BY created_at DESC LIMIT $` + string(rune('0'+len(args)+1))
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.GuestReview
	for rows.Next() {
		var rev domain.GuestReview
		rows.Scan(
			&rev.ID, &rev.TenantID, &rev.GuestID, &rev.GuestName, &rev.ReservationID, &rev.RoomNumber, &rev.Channel, &rev.ExternalRef, &rev.OverallRating, &rev.Cleanliness, &rev.Service, &rev.Location, &rev.Value, &rev.Amenities, &rev.Comment, &rev.IsPublished, &rev.StaffResponse, &rev.RespondedBy, &rev.RespondedAt, &rev.SentAt, &rev.CompletedAt, &rev.CreatedAt, &rev.UpdatedAt)
		out = append(out, &rev)
	}
	return out, nil
}

func (r *ReviewsRepository) RespondToReview(ctx context.Context, tenantID, id, response, respondedBy string) error {
	r.pool.SetTenant(ctx, tenantID)
	_, err := r.pool.Exec(ctx, `UPDATE guest_reviews SET staff_response = $1, responded_by = $2, responded_at = NOW(), updated_at = NOW() WHERE id = $3 AND tenant_id = $4`, response, respondedBy, id, tenantID)
	return err
}

func (r *ReviewsRepository) CreateReviewRequest(ctx context.Context, tenantID string, req *domain.ReviewRequest) (*domain.ReviewRequest, error) {
	r.pool.SetTenant(ctx, tenantID)
	err := r.pool.QueryRow(ctx, `
		INSERT INTO review_requests (tenant_id, guest_id, reservation_id, channel, status, scheduled_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW()) RETURNING id, created_at`,
		tenantID, req.GuestID, req.ReservationID, req.Channel, req.Status, req.ScheduledAt).Scan(
		&req.ID, &req.CreatedAt)
	req.TenantID = tenantID
	return req, err
}

func (r *ReviewsRepository) ListReviewRequests(ctx context.Context, tenantID string, status *string, limit int) ([]*domain.ReviewRequest, error) {
	r.pool.SetTenant(ctx, tenantID)
	query := `SELECT id, tenant_id, guest_id, reservation_id, channel, status, scheduled_at, sent_at, opened_at, completed_at, bounce_reason, created_at FROM review_requests WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	if status != nil {
		query += ` AND status = $2`
		args = append(args, *status)
	}
	query += ` ORDER BY scheduled_at ASC`
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.ReviewRequest
	for rows.Next() {
		var req domain.ReviewRequest
		rows.Scan(
			&req.ID, &req.TenantID, &req.GuestID, &req.ReservationID, &req.Channel, &req.Status, &req.ScheduledAt, &req.SentAt, &req.OpenedAt, &req.CompletedAt, &req.BounceReason, &req.CreatedAt)
		out = append(out, &req)
	}
	return out, nil
}

func (r *ReviewsRepository) GetRatingSnapshot(ctx context.Context, tenantID, period string) (*domain.RatingSnapshot, error) {
	r.pool.SetTenant(ctx, tenantID)
	var days int
	switch period {
	case "7d":
		days = 7
	case "30d":
		days = 30
	case "90d":
		days = 90
	default:
		days = 365
	}

	var s domain.RatingSnapshot
	s.TenantID = tenantID
	s.Period = period
	err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(AVG(overall_rating), 0), COALESCE(AVG(cleanliness), 0), COALESCE(AVG(service), 0), COALESCE(AVG(location), 0), COALESCE(AVG(value), 0), COALESCE(AVG(amenities), 0)
		FROM guest_reviews WHERE tenant_id = $1 AND created_at >= NOW() - INTERVAL '1 day' * $2`,
		tenantID, days).Scan(
		&s.TotalReviews, &s.AverageOverall, &s.AverageCleanliness, &s.AverageService, &s.AverageLocation, &s.AverageValue, &s.AverageAmenities)
	if err != nil {
		return nil, err
	}

	// NPS: promoters (4-5) - detractors (1-2) / total * 100
	var promoters, detractors int
	_ = r.pool.QueryRow(ctx, `
		SELECT COUNT(CASE WHEN overall_rating >= 4 THEN 1 END), COUNT(CASE WHEN overall_rating <= 2 THEN 1 END)
		FROM guest_reviews WHERE tenant_id = $1 AND created_at >= NOW() - INTERVAL '1 day' * $2`,
		tenantID, days).Scan(&promoters, &detractors)
	if s.TotalReviews > 0 {
		s.NPS = float64(promoters-detractors) / float64(s.TotalReviews) * 100
	}

	// Response rate
	var responded int
	_ = r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM guest_reviews WHERE tenant_id = $1 AND created_at >= NOW() - INTERVAL '1 day' * $2 AND staff_response IS NOT NULL`,
		tenantID, days).Scan(&responded)
	if s.TotalReviews > 0 {
		s.ResponseRate = float64(responded) / float64(s.TotalReviews) * 100
	}

	return &s, nil
}
