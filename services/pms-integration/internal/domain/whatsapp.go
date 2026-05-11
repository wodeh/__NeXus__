package domain

import "time"

// WhatsAppConversation tracks a guest's conversation with the bot.
type WhatsAppConversation struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	GuestPhone   string    `json:"guest_phone"`
	GuestName    string    `json:"guest_name,omitempty"`
	CurrentState string    `json:"current_state"` // greeting, ask_dates, ask_guests, confirm, payment, complete
	ContextJSON  string    `json:"context_json"`  // serialized booking context
	LastMessageAt time.Time `json:"last_message_at"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// WhatsAppMessage is a single inbound or outbound message.
type WhatsAppMessage struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	ConversationID string   `json:"conversation_id"`
	Direction     string    `json:"direction"` // inbound, outbound
	Body          string    `json:"body"`
	MessageType   string    `json:"message_type"` // text, button, template, media
	ButtonPayload string    `json:"button_payload,omitempty"`
	SentAt        time.Time `json:"sent_at"`
	DeliveredAt   *time.Time `json:"delivered_at,omitempty"`
	ReadAt        *time.Time `json:"read_at,omitempty"`
	Status        string    `json:"status"` // sent, delivered, read, failed
}

// WhatsAppBotConfig holds tenant-specific bot settings.
type WhatsAppBotConfig struct {
	TenantID          string    `json:"tenant_id"`
	Enabled           bool      `json:"enabled"`
	PhoneNumberID     string    `json:"phone_number_id"`
	AccessToken       string    `json:"-"`
	WebhookSecret     string    `json:"-"`
	WelcomeMessage    string    `json:"welcome_message"`
	AutoReplyEnabled  bool      `json:"auto_reply_enabled"`
	BookingEnabled    bool      `json:"booking_enabled"`
	CheckinEnabled    bool      `json:"checkin_enabled"`
	CheckoutEnabled   bool      `json:"checkout_enabled"`
	SupportEnabled    bool      `json:"support_enabled"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// WhatsAppBookingContext is the transient state during a booking conversation.
type WhatsAppBookingContext struct {
	CheckIn     string `json:"check_in,omitempty"`
	CheckOut    string `json:"check_out,omitempty"`
	Adults      int    `json:"adults,omitempty"`
	Children    int    `json:"children,omitempty"`
	RoomTypeID  string `json:"room_type_id,omitempty"`
	GuestName   string `json:"guest_name,omitempty"`
	GuestEmail  string `json:"guest_email,omitempty"`
	TotalAmount float64 `json:"total_amount,omitempty"`
}
