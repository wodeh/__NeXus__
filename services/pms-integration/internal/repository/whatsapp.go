package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// WhatsAppRepository handles WhatsApp bot data.
type WhatsAppRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewWhatsAppRepository creates a new repository.
func NewWhatsAppRepository(pool *db.Pool, metrics *RepositoryMetrics) *WhatsAppRepository {
	return &WhatsAppRepository{pool: pool, metrics: metrics}
}

func (r *WhatsAppRepository) GetConfig(ctx context.Context, tenantID string) (*domain.WhatsAppBotConfig, error) {
	r.pool.SetTenant(ctx, tenantID)
	var c domain.WhatsAppBotConfig
	row := r.pool.QueryRow(ctx, `
		SELECT tenant_id, enabled, phone_number_id, access_token, webhook_secret, welcome_message, auto_reply_enabled, booking_enabled, checkin_enabled, checkout_enabled, support_enabled, created_at, updated_at
		FROM whatsapp_bot_configs WHERE tenant_id = $1`, tenantID)
	err := row.Scan(&c.TenantID, &c.Enabled, &c.PhoneNumberID, &c.AccessToken, &c.WebhookSecret, &c.WelcomeMessage, &c.AutoReplyEnabled, &c.BookingEnabled, &c.CheckinEnabled, &c.CheckoutEnabled, &c.SupportEnabled, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *WhatsAppRepository) SaveConfig(ctx context.Context, tenantID string, c *domain.WhatsAppBotConfig) error {
	r.pool.SetTenant(ctx, tenantID)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO whatsapp_bot_configs (tenant_id, enabled, phone_number_id, access_token, webhook_secret, welcome_message, auto_reply_enabled, booking_enabled, checkin_enabled, checkout_enabled, support_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		ON CONFLICT (tenant_id) DO UPDATE SET
			enabled = EXCLUDED.enabled, phone_number_id = EXCLUDED.phone_number_id, access_token = EXCLUDED.access_token,
			webhook_secret = EXCLUDED.webhook_secret, welcome_message = EXCLUDED.welcome_message,
			auto_reply_enabled = EXCLUDED.auto_reply_enabled, booking_enabled = EXCLUDED.booking_enabled,
			checkin_enabled = EXCLUDED.checkin_enabled, checkout_enabled = EXCLUDED.checkout_enabled,
			support_enabled = EXCLUDED.support_enabled, updated_at = NOW()`,
		tenantID, c.Enabled, c.PhoneNumberID, c.AccessToken, c.WebhookSecret, c.WelcomeMessage, c.AutoReplyEnabled, c.BookingEnabled, c.CheckinEnabled, c.CheckoutEnabled, c.SupportEnabled)
	return err
}

func (r *WhatsAppRepository) GetConversationByPhone(ctx context.Context, tenantID, phone string) (*domain.WhatsAppConversation, error) {
	r.pool.SetTenant(ctx, tenantID)
	var c domain.WhatsAppConversation
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, guest_phone, guest_name, current_state, context_json, last_message_at, created_at, updated_at
		FROM whatsapp_conversations WHERE tenant_id = $1 AND guest_phone = $2 ORDER BY last_message_at DESC LIMIT 1`, tenantID, phone)
	err := row.Scan(&c.ID, &c.TenantID, &c.GuestPhone, &c.GuestName, &c.CurrentState, &c.ContextJSON, &c.LastMessageAt, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *WhatsAppRepository) CreateConversation(ctx context.Context, tenantID string, c *domain.WhatsAppConversation) (*domain.WhatsAppConversation, error) {
	r.pool.SetTenant(ctx, tenantID)
	err := r.pool.QueryRow(ctx, `
		INSERT INTO whatsapp_conversations (tenant_id, guest_phone, guest_name, current_state, context_json, last_message_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW(), NOW()) RETURNING id, created_at, updated_at`,
		tenantID, c.GuestPhone, c.GuestName, c.CurrentState, c.ContextJSON).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
	c.TenantID = tenantID
	return c, err
}

func (r *WhatsAppRepository) UpdateConversation(ctx context.Context, tenantID, id string, state string, ctxJSON string) error {
	r.pool.SetTenant(ctx, tenantID)
	_, err := r.pool.Exec(ctx, `
		UPDATE whatsapp_conversations SET current_state = $1, context_json = $2, last_message_at = NOW(), updated_at = NOW()
		WHERE id = $3 AND tenant_id = $4`, state, ctxJSON, id, tenantID)
	return err
}

func (r *WhatsAppRepository) ListConversations(ctx context.Context, tenantID string, limit int) ([]*domain.WhatsAppConversation, error) {
	r.pool.SetTenant(ctx, tenantID)
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, guest_phone, guest_name, current_state, context_json, last_message_at, created_at, updated_at
		FROM whatsapp_conversations WHERE tenant_id = $1 ORDER BY last_message_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.WhatsAppConversation
	for rows.Next() {
		var c domain.WhatsAppConversation
		rows.Scan(&c.ID, &c.TenantID, &c.GuestPhone, &c.GuestName, &c.CurrentState, &c.ContextJSON, &c.LastMessageAt, &c.CreatedAt, &c.UpdatedAt)
		out = append(out, &c)
	}
	return out, nil
}

func (r *WhatsAppRepository) AddMessage(ctx context.Context, tenantID string, m *domain.WhatsAppMessage) (*domain.WhatsAppMessage, error) {
	r.pool.SetTenant(ctx, tenantID)
	err := r.pool.QueryRow(ctx, `
		INSERT INTO whatsapp_messages (tenant_id, conversation_id, direction, body, message_type, button_payload, sent_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), $7) RETURNING id, sent_at`,
		tenantID, m.ConversationID, m.Direction, m.Body, m.MessageType, m.ButtonPayload, m.Status).Scan(&m.ID, &m.SentAt)
	m.TenantID = tenantID
	return m, err
}

func (r *WhatsAppRepository) ListMessages(ctx context.Context, tenantID, conversationID string) ([]*domain.WhatsAppMessage, error) {
	r.pool.SetTenant(ctx, tenantID)
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, conversation_id, direction, body, message_type, button_payload, sent_at, delivered_at, read_at, status
		FROM whatsapp_messages WHERE tenant_id = $1 AND conversation_id = $2 ORDER BY sent_at ASC`, tenantID, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.WhatsAppMessage
	for rows.Next() {
		var m domain.WhatsAppMessage
		rows.Scan(&m.ID, &m.TenantID, &m.ConversationID, &m.Direction, &m.Body, &m.MessageType, &m.ButtonPayload, &m.SentAt, &m.DeliveredAt, &m.ReadAt, &m.Status)
		out = append(out, &m)
	}
	return out, nil
}
