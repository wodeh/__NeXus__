package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// ChannelRepository handles external channel reservations and integrations.
type ChannelRepository struct{ pool *db.Pool }

func NewChannelRepository(pool *db.Pool) *ChannelRepository { return &ChannelRepository{pool: pool} }

/* ─── Channel Reservations ─── */

func (r *ChannelRepository) ListChannelReservations(ctx context.Context, tenantID uuid.UUID, source string) ([]domain.ChannelReservation, error) {
	query := `SELECT id, tenant_id, channel_source, external_ref, guest_name, guest_email, guest_phone, room_type, room_number, check_in::text, check_out::text, adults, children, total, currency, status, special_requests, raw_payload, created_at, updated_at FROM channel_reservations WHERE tenant_id=$1`
	args := []interface{}{tenantID}
	if source != "" {
		query += ` AND channel_source=$2`
		args = append(args, source)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil { return nil, fmt.Errorf("list channel reservations: %w", err) }
	defer rows.Close()

	var out []domain.ChannelReservation
	for rows.Next() {
		var res domain.ChannelReservation
		var rawJSON []byte
		var roomNum, specReqs *string
		if err := rows.Scan(&res.ID, &res.TenantID, &res.ChannelSource, &res.ExternalRef, &res.GuestName, &res.GuestEmail, &res.GuestPhone, &res.RoomType, &roomNum, &res.CheckIn, &res.CheckOut, &res.Adults, &res.Children, &res.Total, &res.Currency, &res.Status, &specReqs, &rawJSON, &res.CreatedAt, &res.UpdatedAt); err != nil { continue }
		if roomNum != nil { res.RoomNumber = *roomNum }
		if specReqs != nil { res.SpecialRequests = *specReqs }
		_ = unmarshalJSON(rawJSON, &res.RawPayload)
		out = append(out, res)
	}
	return out, nil
}

func (r *ChannelRepository) CreateChannelReservation(ctx context.Context, res *domain.ChannelReservation) error {
	if res.ID == uuid.Nil { res.ID = uuid.New() }
	now := time.Now().UTC()
	res.CreatedAt = now
	res.UpdatedAt = now
	if res.Status == "" { res.Status = "pending" }
	if res.Currency == "" { res.Currency = "USD" }

	_, err := r.pool.Exec(ctx, `
		INSERT INTO channel_reservations (id, tenant_id, channel_source, external_ref, guest_name, guest_email, guest_phone, room_type, room_number, check_in, check_out, adults, children, total, currency, status, special_requests, raw_payload, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
		res.ID, res.TenantID, res.ChannelSource, res.ExternalRef, res.GuestName, res.GuestEmail, res.GuestPhone, res.RoomType, res.RoomNumber, res.CheckIn, res.CheckOut, res.Adults, res.Children, res.Total, res.Currency, res.Status, res.SpecialRequests, marshalJSON(res.RawPayload), res.CreatedAt, res.UpdatedAt)
	if err != nil { return fmt.Errorf("create channel reservation: %w", err) }
	return nil
}

func (r *ChannelRepository) UpdateChannelReservationStatus(ctx context.Context, tenantID, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `UPDATE channel_reservations SET status=$1, updated_at=NOW() WHERE tenant_id=$2 AND id=$3`, status, tenantID, id)
	if err != nil { return fmt.Errorf("update channel reservation: %w", err) }
	return nil
}

/* ─── Channels (OTA Sources) ─── */

func (r *ChannelRepository) ListChannels(ctx context.Context, tenantID uuid.UUID) ([]domain.Channel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, source, display_name, is_active, commission_pct, webhook_url, last_sync_at, last_sync_status, config, created_at, updated_at
		FROM channels WHERE tenant_id=$1 AND deleted_at IS NULL`, tenantID)
	if err != nil { return nil, fmt.Errorf("list channels: %w", err) }
	defer rows.Close()

	var out []domain.Channel
	for rows.Next() {
		var c domain.Channel
		var configJSON []byte
		var lastSync *time.Time
		if err := rows.Scan(&c.ID, &c.TenantID, &c.Source, &c.DisplayName, &c.IsActive, &c.CommissionPct, &c.WebhookURL, &lastSync, &c.LastSyncStatus, &configJSON, &c.CreatedAt, &c.UpdatedAt); err != nil { continue }
		if lastSync != nil { c.LastSyncAt = lastSync }
		_ = unmarshalJSON(configJSON, &c.Config)
		out = append(out, c)
	}
	return out, nil
}

func (r *ChannelRepository) CreateChannel(ctx context.Context, c *domain.Channel) error {
	if c.ID == uuid.Nil { c.ID = uuid.New() }
	now := time.Now().UTC()
	c.CreatedAt = now
	c.UpdatedAt = now
	if c.LastSyncStatus == "" { c.LastSyncStatus = "n/a" }

	_, err := r.pool.Exec(ctx, `
		INSERT INTO channels (id, tenant_id, source, display_name, is_active, commission_pct, api_key, api_secret, webhook_url, last_sync_status, config, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		c.ID, c.TenantID, c.Source, c.DisplayName, c.IsActive, c.CommissionPct, c.APIKey, c.APISecret, c.WebhookURL, c.LastSyncStatus, marshalJSON(c.Config), c.CreatedAt, c.UpdatedAt)
	if err != nil { return fmt.Errorf("create channel: %w", err) }
	return nil
}

func (r *ChannelRepository) UpdateChannel(ctx context.Context, tenantID, id uuid.UUID, updates map[string]interface{}) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE channels SET display_name=$1, is_active=$2, commission_pct=$3, api_key=$4, api_secret=$5, webhook_url=$6, config=$7, updated_at=NOW()
		WHERE tenant_id=$8 AND id=$9`,
		updates["display_name"], updates["is_active"], updates["commission_pct"], updates["api_key"], updates["api_secret"], updates["webhook_url"], marshalJSON(updates["config"]), tenantID, id)
	if err != nil { return fmt.Errorf("update channel: %w", err) }
	return nil
}

func (r *ChannelRepository) DeleteChannel(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE channels SET deleted_at=NOW() WHERE tenant_id=$1 AND id=$2`, tenantID, id)
	if err != nil { return fmt.Errorf("delete channel: %w", err) }
	return nil
}

/* ─── Availability ─── */

func (r *ChannelRepository) GetAvailability(ctx context.Context, tenantID uuid.UUID, roomType string, from, to time.Time) ([]domain.ChannelAvailability, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT tenant_id, room_type, date, total_rooms, booked_rooms, blocked_rooms, available, rate, currency
		FROM channel_availability WHERE tenant_id=$1 AND room_type=$2 AND date BETWEEN $3 AND $4 ORDER BY date`,
		tenantID, roomType, from, to)
	if err != nil { return nil, fmt.Errorf("get availability: %w", err) }
	defer rows.Close()

	var out []domain.ChannelAvailability
	for rows.Next() {
		var a domain.ChannelAvailability
		if err := rows.Scan(&a.TenantID, &a.RoomType, &a.Date, &a.TotalRooms, &a.BookedRooms, &a.BlockedRooms, &a.Available, &a.Rate, &a.Currency); err != nil { continue }
		out = append(out, a)
	}
	return out, nil
}

func (r *ChannelRepository) UpdateAvailability(ctx context.Context, a *domain.ChannelAvailability) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO channel_availability (tenant_id, room_type, date, total_rooms, booked_rooms, blocked_rooms, available, rate, currency)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		ON CONFLICT (tenant_id, room_type, date) DO UPDATE SET
		total_rooms=$4, booked_rooms=$5, blocked_rooms=$6, available=$7, rate=$8, currency=$9`,
		a.TenantID, a.RoomType, a.Date, a.TotalRooms, a.BookedRooms, a.BlockedRooms, a.Available, a.Rate, a.Currency)
	if err != nil { return fmt.Errorf("update availability: %w", err) }
	return nil
}
