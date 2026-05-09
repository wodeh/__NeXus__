// Package repository provides tenant-aware data access for notifications.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// NotificationLogRepository persists notification delivery records.
type NotificationLogRepository interface {
	Create(ctx context.Context, log *domain.NotificationLog) error
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.NotificationLog, error)
	UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus, errorMsg string) error
}

type PostgresNotificationLogRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewNotificationLogRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresNotificationLogRepository {
	return &PostgresNotificationLogRepository{pool: pool, metrics: metrics}
}

func (r *PostgresNotificationLogRepository) Create(ctx context.Context, log *domain.NotificationLog) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO notification_logs (id, tenant_id, template_id, channel, recipient, subject, body, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, log.ID, log.TenantID, log.TemplateID, log.Channel, log.Recipient, log.Subject, log.Body, log.Status, log.CreatedAt)
	return err
}

func (r *PostgresNotificationLogRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.NotificationLog, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, template_id, channel, recipient, subject, body, status, error_message, sent_at, delivered_at, created_at
		FROM notification_logs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanNotificationLogs(rows)
}

func (r *PostgresNotificationLogRepository) UpdateStatus(ctx context.Context, id string, status domain.NotificationStatus, errorMsg string) error {
	var sentAt, deliveredAt *time.Time
	now := time.Now().UTC()
	if status == domain.NotificationStatusSent {
		sentAt = &now
	}
	if status == domain.NotificationStatusDelivered {
		deliveredAt = &now
		if sentAt == nil {
			sentAt = &now
		}
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE notification_logs
		SET status = $1, error_message = $2, sent_at = $3, delivered_at = $4
		WHERE id = $5
	`, status, errorMsg, sentAt, deliveredAt, id)
	return err
}

func scanNotificationLogs(rows pgx.Rows) ([]*domain.NotificationLog, error) {
	var logs []*domain.NotificationLog
	for rows.Next() {
		var l domain.NotificationLog
		var sentAt, deliveredAt *time.Time
		err := rows.Scan(
			&l.ID, &l.TenantID, &l.TemplateID, &l.Channel, &l.Recipient, &l.Subject, &l.Body,
			&l.Status, &l.ErrorMessage, &sentAt, &deliveredAt, &l.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		l.SentAt = sentAt
		l.DeliveredAt = deliveredAt
		logs = append(logs, &l)
	}
	return logs, rows.Err()
}

// GuestMessageRepository persists guest messaging.
type GuestMessageRepository interface {
	Create(ctx context.Context, msg *domain.GuestMessage) error
	ListByGuest(ctx context.Context, tenantID, guestID string, limit, offset int) ([]*domain.GuestMessage, error)
	MarkAsRead(ctx context.Context, id string) error
	ListUnread(ctx context.Context, tenantID string) ([]*domain.GuestMessage, error)
}

type PostgresGuestMessageRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewGuestMessageRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresGuestMessageRepository {
	return &PostgresGuestMessageRepository{pool: pool, metrics: metrics}
}

func (r *PostgresGuestMessageRepository) Create(ctx context.Context, msg *domain.GuestMessage) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO guest_messages (id, tenant_id, guest_id, reservation_id, direction, channel, content, sent_by, is_read, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`, msg.ID, msg.TenantID, msg.GuestID, msg.ReservationID, msg.Direction, msg.Channel, msg.Content, msg.SentBy, msg.IsRead, msg.CreatedAt)
	return err
}

func (r *PostgresGuestMessageRepository) ListByGuest(ctx context.Context, tenantID, guestID string, limit, offset int) ([]*domain.GuestMessage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, guest_id, reservation_id, direction, channel, content, sent_by, is_read, read_at, created_at
		FROM guest_messages
		WHERE tenant_id = $1 AND guest_id = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, tenantID, guestID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGuestMessages(rows)
}

func (r *PostgresGuestMessageRepository) MarkAsRead(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE guest_messages
		SET is_read = TRUE, read_at = NOW()
		WHERE id = $1
	`, id)
	return err
}

func (r *PostgresGuestMessageRepository) ListUnread(ctx context.Context, tenantID string) ([]*domain.GuestMessage, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, guest_id, reservation_id, direction, channel, content, sent_by, is_read, read_at, created_at
		FROM guest_messages
		WHERE tenant_id = $1 AND is_read = FALSE AND direction = 'inbound'
		ORDER BY created_at DESC
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGuestMessages(rows)
}

func scanGuestMessages(rows pgx.Rows) ([]*domain.GuestMessage, error) {
	var msgs []*domain.GuestMessage
	for rows.Next() {
		var m domain.GuestMessage
		var resID, sentBy, readAt *string
		err := rows.Scan(
			&m.ID, &m.TenantID, &m.GuestID, &resID, &m.Direction, &m.Channel, &m.Content,
			&sentBy, &m.IsRead, &readAt, &m.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		if resID != nil { m.ReservationID = resID }
		if sentBy != nil { m.SentBy = sentBy }
		if readAt != nil {
			t, _ := time.Parse(time.RFC3339, *readAt)
			m.ReadAt = &t
		}
		msgs = append(msgs, &m)
	}
	return msgs, rows.Err()
}
