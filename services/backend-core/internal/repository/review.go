package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// ReviewRepository provides review data access.
type ReviewRepository struct {
	pool *db.Pool
}

// NewReviewRepository creates a review repository.
func NewReviewRepository(pool *db.Pool) *ReviewRepository {
	return &ReviewRepository{pool: pool}
}

func (r *ReviewRepository) execer(tx pgx.Tx) interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
} {
	if tx != nil {
		return tx
	}
	return r.pool
}

// List returns all reviews for a tenant.
func (r *ReviewRepository) List(ctx context.Context, tenantID string) ([]domain.Review, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, reservation_id, guest_name, room_number, rating,
			cleanliness, service, location, value, comment, staff_reply, replied_at,
			is_published, source, created_at, updated_at
		FROM reviews
		WHERE tenant_id = $1
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list reviews: %w", err)
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		var rv domain.Review
		var staffReply *string
		var repliedAt *time.Time
		if err := rows.Scan(
			&rv.ID, &rv.TenantID, &rv.ReservationID, &rv.GuestName, &rv.RoomNumber, &rv.Rating,
			&rv.Cleanliness, &rv.Service, &rv.Location, &rv.Value, &rv.Comment, &staffReply, &repliedAt,
			&rv.IsPublished, &rv.Source, &rv.CreatedAt, &rv.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan review: %w", err)
		}
		rv.StaffReply = staffReply
		rv.RepliedAt = repliedAt
		reviews = append(reviews, rv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("review rows: %w", err)
	}
	return reviews, nil
}

// Create inserts a new review.
func (r *ReviewRepository) Create(ctx context.Context, tenantID string, rv *domain.Review) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO reviews (tenant_id, reservation_id, guest_name, room_number, rating,
			cleanliness, service, location, value, comment, source, is_published)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, false)
		RETURNING id, created_at, updated_at
	`, tenantID, rv.ReservationID, rv.GuestName, rv.RoomNumber, rv.Rating,
		rv.Cleanliness, rv.Service, rv.Location, rv.Value, rv.Comment, rv.Source,
	).Scan(&rv.ID, &rv.CreatedAt, &rv.UpdatedAt)
}

// AddReply adds a staff reply to a review.
func (r *ReviewRepository) AddReply(ctx context.Context, tenantID string, id uuid.UUID, reply string) error {
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE reviews
		SET staff_reply = $3, replied_at = NOW(), updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, id, reply)
	if err != nil {
		return fmt.Errorf("add reply: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Publish toggles review publication status.
func (r *ReviewRepository) Publish(ctx context.Context, tenantID string, id uuid.UUID, published bool) error {
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE reviews
		SET is_published = $3, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, id, published)
	if err != nil {
		return fmt.Errorf("publish review: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// GetStats returns aggregated review statistics.
func (r *ReviewRepository) GetStats(ctx context.Context, tenantID string) (*domain.ReviewStats, error) {
	var s domain.ReviewStats
	s.TenantID = tenantID
	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COALESCE(AVG(rating), 0),
			COALESCE(AVG(cleanliness), 0),
			COALESCE(AVG(service), 0),
			COALESCE(AVG(location), 0),
			COALESCE(AVG(value), 0),
			COUNT(*) FILTER (WHERE rating = 5),
			COUNT(*) FILTER (WHERE rating = 4),
			COUNT(*) FILTER (WHERE rating = 3),
			COUNT(*) FILTER (WHERE rating = 2),
			COUNT(*) FILTER (WHERE rating = 1)
		FROM reviews
		WHERE tenant_id = $1
	`, tenantID).Scan(
		&s.TotalReviews, &s.AverageRating, &s.AverageCleanliness, &s.AverageService,
		&s.AverageLocation, &s.AverageValue, &s.FiveStarCount, &s.FourStarCount,
		&s.ThreeStarCount, &s.TwoStarCount, &s.OneStarCount,
	)
	if err != nil {
		return nil, fmt.Errorf("review stats: %w", err)
	}
	return &s, nil
}
