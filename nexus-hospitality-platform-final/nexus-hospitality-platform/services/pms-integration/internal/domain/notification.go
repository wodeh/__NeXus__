// Package domain defines notification aggregates.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// ==================== NOTIFICATION TEMPLATE ====================

type NotificationTemplateID string

type NotificationChannel string

const (
	NotificationEmail NotificationChannel = "email"
	NotificationSMS   NotificationChannel = "sms"
	NotificationPush  NotificationChannel = "push"
	NotificationInApp NotificationChannel = "in_app"
)

type NotificationTemplate struct {
	ID        NotificationTemplateID
	TenantID  string
	Name      string
	Channel   NotificationChannel
	Subject   string
	Body      string
	Variables []string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewNotificationTemplate(tenantID, name string, channel NotificationChannel, subject, body string) *NotificationTemplate {
	now := time.Now().UTC()
	return &NotificationTemplate{
		ID:        NotificationTemplateID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:  tenantID,
		Name:      name,
		Channel:   channel,
		Subject:   subject,
		Body:      body,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ==================== NOTIFICATION LOG ====================

type NotificationLogID string

type NotificationStatus string

const (
	NotificationStatusQueued     NotificationStatus = "queued"
	NotificationStatusSent       NotificationStatus = "sent"
	NotificationStatusDelivered  NotificationStatus = "delivered"
	NotificationStatusFailed     NotificationStatus = "failed"
	NotificationStatusBounced    NotificationStatus = "bounced"
)

type NotificationLog struct {
	ID           NotificationLogID
	TenantID     string
	TemplateID   string
	Channel      NotificationChannel
	Recipient    string
	Subject      string
	Body         string
	Status       NotificationStatus
	ErrorMessage string
	SentAt       *time.Time
	DeliveredAt  *time.Time
	CreatedAt    time.Time
}

func NewNotificationLog(tenantID, templateID string, channel NotificationChannel, recipient, subject, body string) *NotificationLog {
	return &NotificationLog{
		ID:         NotificationLogID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:   tenantID,
		TemplateID: templateID,
		Channel:    channel,
		Recipient:  recipient,
		Subject:    subject,
		Body:       body,
		Status:     NotificationStatusQueued,
		CreatedAt:  time.Now().UTC(),
	}
}

// ==================== GUEST MESSAGE ====================

type GuestMessageID string

type GuestMessageDirection string

const (
	GuestMessageInbound  GuestMessageDirection = "inbound"
	GuestMessageOutbound GuestMessageDirection = "outbound"
)

type GuestMessage struct {
	ID          GuestMessageID
	TenantID    string
	GuestID     string
	ReservationID *string
	Direction   GuestMessageDirection
	Channel     NotificationChannel
	Content     string
	SentBy      *string // user ID for outbound
	IsRead      bool
	ReadAt      *time.Time
	CreatedAt   time.Time
}

func NewGuestMessage(tenantID, guestID string, direction GuestMessageDirection, channel NotificationChannel, content string) *GuestMessage {
	return &GuestMessage{
		ID:        GuestMessageID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:  tenantID,
		GuestID:   guestID,
		Direction: direction,
		Channel:   channel,
		Content:   content,
		CreatedAt: time.Now().UTC(),
	}
}
