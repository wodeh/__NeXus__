package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// WhatsAppRepository provides WhatsApp conversation data access.
type WhatsAppRepository struct {
	pool *db.Pool
}

// NewWhatsAppRepository creates a WhatsApp repository.
func NewWhatsAppRepository(pool *db.Pool) *WhatsAppRepository {
	return &WhatsAppRepository{pool: pool}
}

func (r *WhatsAppRepository) execer(tx pgx.Tx) interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
} {
	if tx != nil {
		return tx
	}
	return r.pool
}

// ListConversations returns all conversations for a tenant.
func (r *WhatsAppRepository) ListConversations(ctx context.Context, tenantID string) ([]domain.WhatsAppConversation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, guest_phone, guest_name, current_state, booking_created,
			messages, last_message_at, created_at, updated_at
		FROM whatsapp_conversations
		WHERE tenant_id = $1
		ORDER BY last_message_at DESC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	var convs []domain.WhatsAppConversation
	for rows.Next() {
		var c domain.WhatsAppConversation
		var guestName *string
		if err := rows.Scan(
			&c.ID, &c.TenantID, &c.GuestPhone, &guestName, &c.CurrentState, &c.BookingCreated,
			&c.Messages, &c.LastMessageAt, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan conversation: %w", err)
		}
		c.GuestName = guestName
		convs = append(convs, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("conversation rows: %w", err)
	}
	return convs, nil
}

// GetConversation returns a single conversation.
func (r *WhatsAppRepository) GetConversation(ctx context.Context, tenantID string, id uuid.UUID) (*domain.WhatsAppConversation, error) {
	var c domain.WhatsAppConversation
	var guestName *string
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, guest_phone, guest_name, current_state, booking_created,
			messages, last_message_at, created_at, updated_at
		FROM whatsapp_conversations
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, id).Scan(
		&c.ID, &c.TenantID, &c.GuestPhone, &guestName, &c.CurrentState, &c.BookingCreated,
		&c.Messages, &c.LastMessageAt, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get conversation: %w", err)
	}
	c.GuestName = guestName
	return &c, nil
}

// CreateConversation inserts a new conversation.
func (r *WhatsAppRepository) CreateConversation(ctx context.Context, tenantID string, c *domain.WhatsAppConversation) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO whatsapp_conversations (tenant_id, guest_phone, guest_name, current_state, booking_created, messages, last_message_at)
		VALUES ($1, $2, $3, $4, $5, 0, NOW())
		RETURNING id, created_at, updated_at
	`, tenantID, c.GuestPhone, c.GuestName, c.CurrentState, c.BookingCreated,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

// UpdateConversationState updates conversation state.
func (r *WhatsAppRepository) UpdateConversationState(ctx context.Context, tenantID string, id uuid.UUID, state string, bookingCreated bool) error {
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE whatsapp_conversations
		SET current_state = $3, booking_created = $4, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, id, state, bookingCreated)
	if err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListMessages returns messages for a conversation.
func (r *WhatsAppRepository) ListMessages(ctx context.Context, tenantID string, convID uuid.UUID) ([]domain.WhatsAppMessage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, conversation_id, direction, body, status, sent_at
		FROM whatsapp_messages
		WHERE tenant_id = $1 AND conversation_id = $2
		ORDER BY sent_at ASC
	`, tenantID, convID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	defer rows.Close()

	var msgs []domain.WhatsAppMessage
	for rows.Next() {
		var m domain.WhatsAppMessage
		if err := rows.Scan(&m.ID, &m.TenantID, &m.ConversationID, &m.Direction, &m.Body, &m.Status, &m.SentAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msgs = append(msgs, m)
	}
	return msgs, rows.Err()
}

// SendMessage inserts a message and increments conversation count.
func (r *WhatsAppRepository) SendMessage(ctx context.Context, tenantID string, m *domain.WhatsAppMessage) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := tx.QueryRow(ctx, `
		INSERT INTO whatsapp_messages (tenant_id, conversation_id, direction, body, status, sent_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, tenantID, m.ConversationID, m.Direction, m.Body, m.Status, m.SentAt).Scan(&m.ID); err != nil {
		return fmt.Errorf("insert message: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE whatsapp_conversations
		SET messages = messages + 1, last_message_at = $3, updated_at = NOW()
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, m.ConversationID, m.SentAt); err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}

	return tx.Commit(ctx)
}

// GetBotConfig returns bot configuration for a tenant.
func (r *WhatsAppRepository) GetBotConfig(ctx context.Context, tenantID string) (*domain.WhatsAppBotConfig, error) {
	var c domain.WhatsAppBotConfig
	err := r.pool.QueryRow(ctx, `
		SELECT tenant_id, bot_enabled, booking_enabled, auto_reply_enabled,
			welcome_message, phone_number_id, access_token, webhook_url, updated_at
		FROM whatsapp_bot_configs
		WHERE tenant_id = $1
	`, tenantID).Scan(
		&c.TenantID, &c.BotEnabled, &c.BookingEnabled, &c.AutoReplyEnabled,
		&c.WelcomeMessage, &c.PhoneNumberID, &c.AccessToken, &c.WebhookURL, &c.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get bot config: %w", err)
	}
	return &c, nil
}

// UpdateBotConfig updates bot configuration.
func (r *WhatsAppRepository) UpdateBotConfig(ctx context.Context, tenantID string, c *domain.WhatsAppBotConfig) error {
	cmdTag, err := r.pool.Exec(ctx, `
		INSERT INTO whatsapp_bot_configs (tenant_id, bot_enabled, booking_enabled, auto_reply_enabled,
			welcome_message, phone_number_id, access_token, webhook_url, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
		ON CONFLICT (tenant_id) DO UPDATE SET
			bot_enabled = EXCLUDED.bot_enabled,
			booking_enabled = EXCLUDED.booking_enabled,
			auto_reply_enabled = EXCLUDED.auto_reply_enabled,
			welcome_message = EXCLUDED.welcome_message,
			phone_number_id = EXCLUDED.phone_number_id,
			access_token = EXCLUDED.access_token,
			webhook_url = EXCLUDED.webhook_url,
			updated_at = NOW()
	`, tenantID, c.BotEnabled, c.BookingEnabled, c.AutoReplyEnabled,
		c.WelcomeMessage, c.PhoneNumberID, c.AccessToken, c.WebhookURL)
	if err != nil {
		return fmt.Errorf("update bot config: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ListTemplates returns quick-response templates.
func (r *WhatsAppRepository) ListTemplates(ctx context.Context, tenantID string) ([]domain.WhatsAppTemplate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, trigger, response, is_active, created_at
		FROM whatsapp_templates
		WHERE tenant_id = $1 AND is_active = true
		ORDER BY trigger
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	var tmpls []domain.WhatsAppTemplate
	for rows.Next() {
		var t domain.WhatsAppTemplate
		if err := rows.Scan(&t.ID, &t.TenantID, &t.Trigger, &t.Response, &t.IsActive, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		tmpls = append(tmpls, t)
	}
	return tmpls, rows.Err()
}
