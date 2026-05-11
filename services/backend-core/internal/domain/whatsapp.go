package domain

import (
	"time"

	"github.com/google/uuid"
)

// WhatsAppConversation represents an active or completed guest conversation.
type WhatsAppConversation struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	GuestPhone    string    `json:"guest_phone"`
	GuestName     *string   `json:"guest_name,omitempty"`
	CurrentState  string    `json:"current_state"` // greeting, ask_dates, ask_guests, ask_room, confirm, complete, support
	BookingCreated bool     `json:"booking_created"`
	Messages      int       `json:"messages"`
	LastMessageAt time.Time `json:"last_message_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// WhatsAppMessage represents a single message in a conversation.
type WhatsAppMessage struct {
	ID              uuid.UUID `json:"id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	ConversationID  uuid.UUID `json:"conversation_id"`
	Direction       string    `json:"direction"` // inbound, outbound
	Body            string    `json:"body"`
	Status          string    `json:"status"`      // sent, delivered, read, failed
	SentAt          time.Time `json:"sent_at"`
}

// WhatsAppBotConfig holds the tenant's WhatsApp bot settings.
type WhatsAppBotConfig struct {
	TenantID        uuid.UUID `json:"tenant_id"`
	BotEnabled      bool      `json:"bot_enabled"`
	BookingEnabled  bool      `json:"booking_enabled"`
	AutoReplyEnabled bool     `json:"auto_reply_enabled"`
	WelcomeMessage  string    `json:"welcome_message"`
	PhoneNumberID   string    `json:"phone_number_id"`
	AccessToken     string    `json:"access_token,omitempty"`
	WebhookURL      string    `json:"webhook_url"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// WhatsAppTemplate represents a quick-response template.
type WhatsAppTemplate struct {
	ID        uuid.UUID `json:"id"`
	TenantID  uuid.UUID `json:"tenant_id"`
	Trigger   string    `json:"trigger"`
	Response  string    `json:"response"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateConversationRequest starts a new conversation.
type CreateConversationRequest struct {
	GuestPhone string `json:"guest_phone"`
	GuestName  string `json:"guest_name,omitempty"`
}

// SendMessageRequest sends a message.
type SendMessageRequest struct {
	Body string `json:"body"`
}

// UpdateBotConfigRequest updates bot settings.
type UpdateBotConfigRequest struct {
	BotEnabled       *bool  `json:"bot_enabled,omitempty"`
	BookingEnabled   *bool  `json:"booking_enabled,omitempty"`
	AutoReplyEnabled *bool  `json:"auto_reply_enabled,omitempty"`
	WelcomeMessage   string `json:"welcome_message,omitempty"`
	PhoneNumberID    string `json:"phone_number_id,omitempty"`
}
