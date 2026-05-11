package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// ChannelRepository provides channel data access.
type ChannelRepository struct {
	pool *db.Pool
}

// NewChannelRepository creates a channel repository.
func NewChannelRepository(pool *db.Pool) *ChannelRepository {
	return &ChannelRepository{pool: pool}
}

// List returns all channels for a tenant.
func (r *ChannelRepository) List(ctx context.Context, tenantID string) ([]domain.Channel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, source, display_name, is_active, commission_pct, last_sync_at, last_sync_status, config, created_at, updated_at
		FROM channels
		WHERE tenant_id = $1
		ORDER BY display_name
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list channels: %w", err)
	}
	defer rows.Close()

	var channels []domain.Channel
	for rows.Next() {
		var c domain.Channel
		var configJSON []byte
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.Source, &c.DisplayName, &c.IsActive, &c.CommissionPct, &c.LastSyncAt, &c.LastSyncStatus, &configJSON, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		if len(configJSON) > 0 {
			_ = json.Unmarshal(configJSON, &c.Config)
		}
		channels = append(channels, c)
	}
	return channels, rows.Err()
}

// GetByID returns a single channel.
func (r *ChannelRepository) GetByID(ctx context.Context, tenantID string, id uuid.UUID) (*domain.Channel, error) {
	var c domain.Channel
	var configJSON []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, source, display_name, is_active, commission_pct, last_sync_at, last_sync_status, config, created_at, updated_at
		FROM channels
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, id).Scan(
		&c.ID, &c.TenantID, &c.Source, &c.DisplayName, &c.IsActive, &c.CommissionPct, &c.LastSyncAt, &c.LastSyncStatus, &configJSON, &c.CreatedAt, &c.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("channel not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get channel: %w", err)
	}
	if len(configJSON) > 0 {
		_ = json.Unmarshal(configJSON, &c.Config)
	}
	return &c, nil
}

// Create creates a new channel.
func (r *ChannelRepository) Create(ctx context.Context, tenantID string, req *domain.ChannelCreateRequest) (*domain.Channel, error) {
	id := uuid.New()
	now := time.Now().UTC()
	var configJSON []byte

	_, err := r.pool.Exec(ctx, `
		INSERT INTO channels (id, tenant_id, source, display_name, is_active, commission_pct, last_sync_status, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
	`, id, tenantID, req.Source, req.DisplayName, true, req.CommissionPct, "n/a", configJSON, now)
	if err != nil {
		return nil, fmt.Errorf("create channel: %w", err)
	}

	return &domain.Channel{
		ID: id, TenantID: uuid.MustParse(tenantID), Source: req.Source, DisplayName: req.DisplayName,
		IsActive: true, CommissionPct: req.CommissionPct, LastSyncStatus: "n/a",
		CreatedAt: now, UpdatedAt: now,
	}, nil
}

// Update updates a channel.
func (r *ChannelRepository) Update(ctx context.Context, tenantID string, id uuid.UUID, req *domain.ChannelUpdateRequest) (*domain.Channel, error) {
	now := time.Now().UTC()
	_, err := r.pool.Exec(ctx, `
		UPDATE channels SET
			display_name = COALESCE(NULLIF($3, ''), display_name),
			is_active = COALESCE($4, is_active),
			commission_pct = COALESCE($5, commission_pct),
			updated_at = $6
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, id, req.DisplayName, req.IsActive, req.CommissionPct, now)
	if err != nil {
		return nil, fmt.Errorf("update channel: %w", err)
	}
	return r.GetByID(ctx, tenantID, id)
}

// Delete removes a channel.
func (r *ChannelRepository) Delete(ctx context.Context, tenantID string, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM channels WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	if err != nil {
		return fmt.Errorf("delete channel: %w", err)
	}
	return nil
}

// ListSyncLogs returns sync logs for a tenant/channel.
func (r *ChannelRepository) ListSyncLogs(ctx context.Context, tenantID string, channelID *uuid.UUID, limit int) ([]domain.ChannelSyncLog, error) {
	query := `
		SELECT id, tenant_id, channel_id, channel_source, direction, status, records, duration, error, created_at
		FROM channel_sync_logs
		WHERE tenant_id = $1
	`
	args := []interface{}{tenantID}
	if channelID != nil {
		query += ` AND channel_id = $2`
		args = append(args, *channelID)
	}
	query += ` ORDER BY created_at DESC`
	if limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, len(args)+1)
		args = append(args, limit)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list sync logs: %w", err)
	}
	defer rows.Close()

	var logs []domain.ChannelSyncLog
	for rows.Next() {
		var l domain.ChannelSyncLog
		if err := rows.Scan(&l.ID, &l.TenantID, &l.ChannelID, &l.ChannelSource, &l.Direction, &l.Status, &l.Records, &l.Duration, &l.Error, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan sync log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// CreateSyncLog creates a sync log entry.
func (r *ChannelRepository) CreateSyncLog(ctx context.Context, tenantID string, log *domain.ChannelSyncLog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO channel_sync_logs (id, tenant_id, channel_id, channel_source, direction, status, records, duration, error, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, log.ID, tenantID, log.ChannelID, log.ChannelSource, log.Direction, log.Status, log.Records, log.Duration, log.Error, log.CreatedAt)
	if err != nil {
		return fmt.Errorf("create sync log: %w", err)
	}
	return nil
}

// UpdateLastSync updates a channel's last sync status.
func (r *ChannelRepository) UpdateLastSync(ctx context.Context, tenantID string, id uuid.UUID, status string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE channels SET last_sync_at = $3, last_sync_status = $4, updated_at = $3
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, id, time.Now().UTC(), status)
	if err != nil {
		return fmt.Errorf("update last sync: %w", err)
	}
	return nil
}
