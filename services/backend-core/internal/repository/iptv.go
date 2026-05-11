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

// IPTVRepository provides IPTV data access.
type IPTVRepository struct {
	pool *db.Pool
}

// NewIPTVRepository creates an IPTV repository.
func NewIPTVRepository(pool *db.Pool) *IPTVRepository {
	return &IPTVRepository{pool: pool}
}

func (r *IPTVRepository) execer(tx pgx.Tx) interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
} {
	if tx != nil {
		return tx
	}
	return r.pool
}

// ListChannels returns all channels for a tenant.
func (r *IPTVRepository) ListChannels(ctx context.Context, tenantID string) ([]domain.IPTVChannel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, number, stream_url, logo_url, category, language, is_active, is_premium, created_at, updated_at
		FROM iptv_channels
		WHERE tenant_id = $1
		ORDER BY number
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	defer rows.Close()

	var channels []domain.IPTVChannel
	for rows.Next() {
		var c domain.IPTVChannel
		var logoURL *string
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.Name, &c.Number, &c.StreamURL, &logoURL,
			&c.Category, &c.Language, &c.IsActive, &c.IsPremium, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		c.LogoURL = logoURL
		channels = append(channels, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("channel rows: %w", err)
	}
	return channels, nil
}

// ListContent returns all on-demand content for a tenant.
func (r *IPTVRepository) ListContent(ctx context.Context, tenantID string) ([]domain.IPTVContent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, title, type, description, duration, thumbnail_url, category, is_active, created_at, updated_at
		FROM iptv_content
		WHERE tenant_id = $1
		ORDER BY title
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list content: %w", err)
	}
	defer rows.Close()

	var items []domain.IPTVContent
	for rows.Next() {
		var i domain.IPTVContent
		var duration *int
		var thumb *string
		if err := rows.Scan(
			&i.ID, &i.TenantID, &i.Title, &i.Type, &i.Description, &duration, &thumb,
			&i.Category, &i.IsActive, &i.CreatedAt, &i.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan content: %w", err)
		}
		i.Duration = duration
		i.ThumbnailURL = thumb
		items = append(items, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("content rows: %w", err)
	}
	return items, nil
}

// ListRoomStatus returns IPTV status for all rooms.
func (r *IPTVRepository) ListRoomStatus(ctx context.Context, tenantID string) ([]domain.IPTVRoomStatus, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, room_id, room_number, is_online, current_channel, last_activity_at, created_at, updated_at
		FROM iptv_room_status
		WHERE tenant_id = $1
		ORDER BY room_number
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list room status: %w", err)
	}
	defer rows.Close()

	var statuses []domain.IPTVRoomStatus
	for rows.Next() {
		var s domain.IPTVRoomStatus
		var ch *int
		var last *time.Time
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.RoomID, &s.RoomNumber, &s.IsOnline, &ch, &last,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room status: %w", err)
		}
		s.CurrentChannel = ch
		s.LastActivityAt = last
		statuses = append(statuses, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("room status rows: %w", err)
	}
	return statuses, nil
}
