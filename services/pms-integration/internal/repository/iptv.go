package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// IPTVRepository provides CRUD for IPTV entities.
type IPTVRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewIPTVRepository creates a new IPTV repository.
func NewIPTVRepository(pool *db.Pool, metrics *RepositoryMetrics) *IPTVRepository {
	return &IPTVRepository{pool: pool, metrics: metrics}
}

func (r *IPTVRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

// ==================== CHANNELS ====================

func (r *IPTVRepository) CreateChannel(ctx context.Context, ch *domain.IPTVChannel) error {
	start := time.Now()
	r.metrics.IncQuery("iptv_channel", "create")
	defer r.metrics.ObserveDuration("iptv_channel", "create", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, ch.TenantID); err != nil {
		r.metrics.IncError("iptv_channel", "create", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO iptv_channels (tenant_id, property_id, name, channel_number, category, stream_url, icon_url, is_active, is_premium, language, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, ch.TenantID, ch.PropertyID, ch.Name, ch.ChannelNum, ch.Category, ch.StreamURL, ch.IconURL, ch.IsActive, ch.IsPremium, ch.Language, ch.Metadata)
	if err != nil {
		r.metrics.IncError("iptv_channel", "create", "query")
		return fmt.Errorf("create channel: %w", err)
	}
	return nil
}

func (r *IPTVRepository) ListChannels(ctx context.Context, tenantID, propertyID string) ([]domain.IPTVChannel, error) {
	start := time.Now()
	r.metrics.IncQuery("iptv_channel", "list")
	defer r.metrics.ObserveDuration("iptv_channel", "list", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("iptv_channel", "list", "tenant_bind")
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, name, channel_number, category, stream_url, icon_url, is_active, is_premium, language, metadata, created_at, updated_at
		FROM iptv_channels WHERE tenant_id = $1 AND property_id = $2 AND is_active = true ORDER BY channel_number
	`, tenantID, propertyID)
	if err != nil {
		r.metrics.IncError("iptv_channel", "list", "query")
		return nil, fmt.Errorf("list channels: %w", err)
	}
	defer rows.Close()

	return scanIPTVChannels(rows)
}

func (r *IPTVRepository) GetChannel(ctx context.Context, tenantID, id string) (*domain.IPTVChannel, error) {
	start := time.Now()
	r.metrics.IncQuery("iptv_channel", "get")
	defer r.metrics.ObserveDuration("iptv_channel", "get", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("iptv_channel", "get", "tenant_bind")
		return nil, err
	}

	var ch domain.IPTVChannel
	var metadata map[string]interface{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, name, channel_number, category, stream_url, icon_url, is_active, is_premium, language, metadata, created_at, updated_at
		FROM iptv_channels WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(
		&ch.ID, &ch.TenantID, &ch.PropertyID, &ch.Name, &ch.ChannelNum, &ch.Category,
		&ch.StreamURL, &ch.IconURL, &ch.IsActive, &ch.IsPremium, &ch.Language, &metadata, &ch.CreatedAt, &ch.UpdatedAt,
	)
	if err != nil {
		r.metrics.IncError("iptv_channel", "get", "query")
		return nil, fmt.Errorf("get channel: %w", err)
	}
	ch.Metadata = metadata
	return &ch, nil
}

func (r *IPTVRepository) UpdateChannel(ctx context.Context, ch *domain.IPTVChannel) error {
	start := time.Now()
	r.metrics.IncQuery("iptv_channel", "update")
	defer r.metrics.ObserveDuration("iptv_channel", "update", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, ch.TenantID); err != nil {
		r.metrics.IncError("iptv_channel", "update", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE iptv_channels SET name = $1, channel_number = $2, category = $3, stream_url = $4, icon_url = $5,
		is_active = $6, is_premium = $7, language = $8, metadata = $9
		WHERE id = $10 AND tenant_id = $11
	`, ch.Name, ch.ChannelNum, ch.Category, ch.StreamURL, ch.IconURL, ch.IsActive, ch.IsPremium, ch.Language, ch.Metadata, ch.ID, ch.TenantID)
	if err != nil {
		r.metrics.IncError("iptv_channel", "update", "query")
		return fmt.Errorf("update channel: %w", err)
	}
	return nil
}

func (r *IPTVRepository) DeleteChannel(ctx context.Context, id, tenantID string) error {
	start := time.Now()
	r.metrics.IncQuery("iptv_channel", "delete")
	defer r.metrics.ObserveDuration("iptv_channel", "delete", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("iptv_channel", "delete", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `UPDATE iptv_channels SET is_active = false WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		r.metrics.IncError("iptv_channel", "delete", "query")
		return fmt.Errorf("delete channel: %w", err)
	}
	return nil
}

// ==================== CONTENT ====================

func (r *IPTVRepository) CreateContent(ctx context.Context, c *domain.IPTVContent) error {
	start := time.Now()
	r.metrics.IncQuery("iptv_content", "create")
	defer r.metrics.ObserveDuration("iptv_content", "create", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, c.TenantID); err != nil {
		r.metrics.IncError("iptv_content", "create", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO iptv_content (tenant_id, property_id, title, content_type, category, description, poster_url, stream_url, duration_minutes, is_premium, is_active, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`, c.TenantID, c.PropertyID, c.Title, c.ContentType, c.Category, c.Description, c.PosterURL, c.StreamURL, c.Duration, c.IsPremium, c.IsActive, c.Metadata)
	if err != nil {
		r.metrics.IncError("iptv_content", "create", "query")
		return fmt.Errorf("create content: %w", err)
	}
	return nil
}

func (r *IPTVRepository) ListContent(ctx context.Context, tenantID, propertyID string) ([]domain.IPTVContent, error) {
	start := time.Now()
	r.metrics.IncQuery("iptv_content", "list")
	defer r.metrics.ObserveDuration("iptv_content", "list", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("iptv_content", "list", "tenant_bind")
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, title, content_type, category, description, poster_url, stream_url, duration_minutes, is_premium, is_active, metadata, created_at, updated_at
		FROM iptv_content WHERE tenant_id = $1 AND property_id = $2 AND is_active = true ORDER BY title
	`, tenantID, propertyID)
	if err != nil {
		r.metrics.IncError("iptv_content", "list", "query")
		return nil, fmt.Errorf("list content: %w", err)
	}
	defer rows.Close()

	return scanIPTVContent(rows)
}

func (r *IPTVRepository) GetContent(ctx context.Context, tenantID, id string) (*domain.IPTVContent, error) {
	start := time.Now()
	r.metrics.IncQuery("iptv_content", "get")
	defer r.metrics.ObserveDuration("iptv_content", "get", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("iptv_content", "get", "tenant_bind")
		return nil, err
	}

	var c domain.IPTVContent
	var metadata map[string]interface{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, title, content_type, category, description, poster_url, stream_url, duration_minutes, is_premium, is_active, metadata, created_at, updated_at
		FROM iptv_content WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(
		&c.ID, &c.TenantID, &c.PropertyID, &c.Title, &c.ContentType, &c.Category,
		&c.Description, &c.PosterURL, &c.StreamURL, &c.Duration, &c.IsPremium, &c.IsActive, &metadata, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		r.metrics.IncError("iptv_content", "get", "query")
		return nil, fmt.Errorf("get content: %w", err)
	}
	c.Metadata = metadata
	return &c, nil
}

func (r *IPTVRepository) UpdateContent(ctx context.Context, c *domain.IPTVContent) error {
	start := time.Now()
	r.metrics.IncQuery("iptv_content", "update")
	defer r.metrics.ObserveDuration("iptv_content", "update", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, c.TenantID); err != nil {
		r.metrics.IncError("iptv_content", "update", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE iptv_content SET title = $1, content_type = $2, category = $3, description = $4,
		poster_url = $5, stream_url = $6, duration_minutes = $7, is_premium = $8, is_active = $9, metadata = $10
		WHERE id = $11 AND tenant_id = $12
	`, c.Title, c.ContentType, c.Category, c.Description, c.PosterURL, c.StreamURL, c.Duration, c.IsPremium, c.IsActive, c.Metadata, c.ID, c.TenantID)
	if err != nil {
		r.metrics.IncError("iptv_content", "update", "query")
		return fmt.Errorf("update content: %w", err)
	}
	return nil
}

func (r *IPTVRepository) DeleteContent(ctx context.Context, id, tenantID string) error {
	start := time.Now()
	r.metrics.IncQuery("iptv_content", "delete")
	defer r.metrics.ObserveDuration("iptv_content", "delete", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("iptv_content", "delete", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `UPDATE iptv_content SET is_active = false WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		r.metrics.IncError("iptv_content", "delete", "query")
		return fmt.Errorf("delete content: %w", err)
	}
	return nil
}

// ==================== ROOM BINDINGS ====================

func (r *IPTVRepository) CreateRoomBinding(ctx context.Context, b *domain.IPTVRoomBinding) error {
	start := time.Now()
	r.metrics.IncQuery("iptv_room_binding", "create")
	defer r.metrics.ObserveDuration("iptv_room_binding", "create", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, b.TenantID); err != nil {
		r.metrics.IncError("iptv_room_binding", "create", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO iptv_room_bindings (tenant_id, property_id, room_id, device_id, device_type, welcome_screen_enabled,
		welcome_message, guest_name_display, checkout_reminder_enabled, language_override, channels_enabled, content_enabled, status, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`, b.TenantID, b.PropertyID, b.RoomID, b.DeviceID, b.DeviceType, b.WelcomeScreenEnabled,
		b.WelcomeMessage, b.GuestNameDisplay, b.CheckoutReminderEnabled, b.LanguageOverride,
		b.ChannelsEnabled, b.ContentEnabled, b.Status, b.Metadata)
	if err != nil {
		r.metrics.IncError("iptv_room_binding", "create", "query")
		return fmt.Errorf("create room binding: %w", err)
	}
	return nil
}

func (r *IPTVRepository) ListRoomBindings(ctx context.Context, tenantID, propertyID string) ([]domain.IPTVRoomBinding, error) {
	start := time.Now()
	r.metrics.IncQuery("iptv_room_binding", "list")
	defer r.metrics.ObserveDuration("iptv_room_binding", "list", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("iptv_room_binding", "list", "tenant_bind")
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, room_id, device_id, device_type, welcome_screen_enabled,
		welcome_message, guest_name_display, checkout_reminder_enabled, language_override,
		channels_enabled, content_enabled, last_sync_at, status, metadata, created_at, updated_at
		FROM iptv_room_bindings WHERE tenant_id = $1 AND property_id = $2 ORDER BY device_id
	`, tenantID, propertyID)
	if err != nil {
		r.metrics.IncError("iptv_room_binding", "list", "query")
		return nil, fmt.Errorf("list room bindings: %w", err)
	}
	defer rows.Close()

	return scanIPTVRoomBindings(rows)
}

func (r *IPTVRepository) GetRoomBinding(ctx context.Context, tenantID, id string) (*domain.IPTVRoomBinding, error) {
	start := time.Now()
	r.metrics.IncQuery("iptv_room_binding", "get")
	defer r.metrics.ObserveDuration("iptv_room_binding", "get", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("iptv_room_binding", "get", "tenant_bind")
		return nil, err
	}

	var b domain.IPTVRoomBinding
	var metadata map[string]interface{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, room_id, device_id, device_type, welcome_screen_enabled,
		welcome_message, guest_name_display, checkout_reminder_enabled, language_override,
		channels_enabled, content_enabled, last_sync_at, status, metadata, created_at, updated_at
		FROM iptv_room_bindings WHERE id = $1 AND tenant_id = $2
	`, id, tenantID).Scan(
		&b.ID, &b.TenantID, &b.PropertyID, &b.RoomID, &b.DeviceID, &b.DeviceType,
		&b.WelcomeScreenEnabled, &b.WelcomeMessage, &b.GuestNameDisplay, &b.CheckoutReminderEnabled,
		&b.LanguageOverride, &b.ChannelsEnabled, &b.ContentEnabled, &b.LastSyncAt, &b.Status,
		&metadata, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		r.metrics.IncError("iptv_room_binding", "get", "query")
		return nil, fmt.Errorf("get room binding: %w", err)
	}
	b.Metadata = metadata
	return &b, nil
}

func (r *IPTVRepository) UpdateRoomBinding(ctx context.Context, b *domain.IPTVRoomBinding) error {
	start := time.Now()
	r.metrics.IncQuery("iptv_room_binding", "update")
	defer r.metrics.ObserveDuration("iptv_room_binding", "update", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, b.TenantID); err != nil {
		r.metrics.IncError("iptv_room_binding", "update", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE iptv_room_bindings SET device_id = $1, device_type = $2, welcome_screen_enabled = $3,
		welcome_message = $4, guest_name_display = $5, checkout_reminder_enabled = $6,
		language_override = $7, channels_enabled = $8, content_enabled = $9, status = $10, metadata = $11
		WHERE id = $12 AND tenant_id = $13
	`, b.DeviceID, b.DeviceType, b.WelcomeScreenEnabled,
		b.WelcomeMessage, b.GuestNameDisplay, b.CheckoutReminderEnabled,
		b.LanguageOverride, b.ChannelsEnabled, b.ContentEnabled, b.Status, b.Metadata,
		b.ID, b.TenantID)
	if err != nil {
		r.metrics.IncError("iptv_room_binding", "update", "query")
		return fmt.Errorf("update room binding: %w", err)
	}
	return nil
}

func (r *IPTVRepository) DeleteRoomBinding(ctx context.Context, id, tenantID string) error {
	start := time.Now()
	r.metrics.IncQuery("iptv_room_binding", "delete")
	defer r.metrics.ObserveDuration("iptv_room_binding", "delete", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("iptv_room_binding", "delete", "tenant_bind")
		return err
	}

	_, err := r.pool.Exec(ctx, `DELETE FROM iptv_room_bindings WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	if err != nil {
		r.metrics.IncError("iptv_room_binding", "delete", "query")
		return fmt.Errorf("delete room binding: %w", err)
	}
	return nil
}

func (r *IPTVRepository) GetRoomBindingByRoom(ctx context.Context, tenantID, roomID string) (*domain.IPTVRoomBinding, error) {
	start := time.Now()
	r.metrics.IncQuery("iptv_room_binding", "get_by_room")
	defer r.metrics.ObserveDuration("iptv_room_binding", "get_by_room", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("iptv_room_binding", "get_by_room", "tenant_bind")
		return nil, err
	}

	var b domain.IPTVRoomBinding
	var metadata map[string]interface{}
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, room_id, device_id, device_type, welcome_screen_enabled,
		welcome_message, guest_name_display, checkout_reminder_enabled, language_override,
		channels_enabled, content_enabled, last_sync_at, status, metadata, created_at, updated_at
		FROM iptv_room_bindings WHERE tenant_id = $1 AND room_id = $2 LIMIT 1
	`, tenantID, roomID).Scan(
		&b.ID, &b.TenantID, &b.PropertyID, &b.RoomID, &b.DeviceID, &b.DeviceType,
		&b.WelcomeScreenEnabled, &b.WelcomeMessage, &b.GuestNameDisplay, &b.CheckoutReminderEnabled,
		&b.LanguageOverride, &b.ChannelsEnabled, &b.ContentEnabled, &b.LastSyncAt, &b.Status,
		&metadata, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		r.metrics.IncError("iptv_room_binding", "get_by_room", "query")
		return nil, fmt.Errorf("get room binding by room: %w", err)
	}
	b.Metadata = metadata
	return &b, nil
}

// ==================== ANALYTICS ====================

func (r *IPTVRepository) GetAnalytics(ctx context.Context, tenantID, propertyID string, from, to time.Time) (*domain.IPTVAnalytics, error) {
	start := time.Now()
	r.metrics.IncQuery("iptv_analytics", "get")
	defer r.metrics.ObserveDuration("iptv_analytics", "get", time.Since(start).Seconds())

	if _, err := r.withTenant(ctx, tenantID); err != nil {
		r.metrics.IncError("iptv_analytics", "get", "tenant_bind")
		return nil, err
	}

	var analytics domain.IPTVAnalytics
	_ = r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FILTER (WHERE status = 'online'), COUNT(*) FILTER (WHERE status = 'offline')
		FROM iptv_room_bindings WHERE tenant_id = $1 AND property_id = $2
	`, tenantID, propertyID).Scan(&analytics.ActiveDevices, &analytics.OfflineDevices)

	_ = r.pool.QueryRow(ctx, `
		SELECT COUNT(*), COALESCE(SUM(total_watch_minutes), 0)
		FROM iptv_guest_sessions WHERE tenant_id = $1 AND property_id = $2 AND session_started_at BETWEEN $3 AND $4
	`, tenantID, propertyID, from, to).Scan(&analytics.TotalSessions, &analytics.TotalWatchMinutes)

	return &analytics, nil
}

// ==================== SCAN HELPERS ====================

func scanIPTVChannels(rows pgx.Rows) ([]domain.IPTVChannel, error) {
	var channels []domain.IPTVChannel
	for rows.Next() {
		var ch domain.IPTVChannel
		var metadata map[string]interface{}
		if err := rows.Scan(
			&ch.ID, &ch.TenantID, &ch.PropertyID, &ch.Name, &ch.ChannelNum, &ch.Category,
			&ch.StreamURL, &ch.IconURL, &ch.IsActive, &ch.IsPremium, &ch.Language, &metadata, &ch.CreatedAt, &ch.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan channel: %w", err)
		}
		ch.Metadata = metadata
		channels = append(channels, ch)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return channels, nil
}

func scanIPTVContent(rows pgx.Rows) ([]domain.IPTVContent, error) {
	var items []domain.IPTVContent
	for rows.Next() {
		var c domain.IPTVContent
		var metadata map[string]interface{}
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.PropertyID, &c.Title, &c.ContentType, &c.Category,
			&c.Description, &c.PosterURL, &c.StreamURL, &c.Duration, &c.IsPremium, &c.IsActive, &metadata, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan content: %w", err)
		}
		c.Metadata = metadata
		items = append(items, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return items, nil
}

func scanIPTVRoomBindings(rows pgx.Rows) ([]domain.IPTVRoomBinding, error) {
	var bindings []domain.IPTVRoomBinding
	for rows.Next() {
		var b domain.IPTVRoomBinding
		var metadata map[string]interface{}
		if err := rows.Scan(
			&b.ID, &b.TenantID, &b.PropertyID, &b.RoomID, &b.DeviceID, &b.DeviceType,
			&b.WelcomeScreenEnabled, &b.WelcomeMessage, &b.GuestNameDisplay, &b.CheckoutReminderEnabled,
			&b.LanguageOverride, &b.ChannelsEnabled, &b.ContentEnabled, &b.LastSyncAt, &b.Status,
			&metadata, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan room binding: %w", err)
		}
		b.Metadata = metadata
		bindings = append(bindings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return bindings, nil
}
