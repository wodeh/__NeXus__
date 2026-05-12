package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// CommRepository provides communication data access.
type CommRepository struct {
	pool *db.Pool
}

// NewCommRepository creates a communication repository.
func NewCommRepository(pool *db.Pool) *CommRepository {
	return &CommRepository{pool: pool}
}

func (r *CommRepository) execer(tx pgx.Tx) interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
} {
	if tx != nil {
		return tx
	}
	return r.pool
}

// ListTemplates returns all communication templates.
func (r *CommRepository) ListTemplates(ctx context.Context, tenantID string) ([]domain.CommTemplate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, subject, body, channel, category, is_active, created_at, updated_at
		FROM comm_templates
		WHERE tenant_id = $1
		ORDER BY name
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list templates: %w", err)
	}
	defer rows.Close()

	var items []domain.CommTemplate
	for rows.Next() {
		var t domain.CommTemplate
		var subj *string
		if err := rows.Scan(
			&t.ID, &t.TenantID, &t.Name, &subj, &t.Body, &t.Channel, &t.Category, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan template: %w", err)
		}
		t.Subject = subj
		items = append(items, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("template rows: %w", err)
	}
	return items, nil
}

// ListSequences returns all communication sequences.
func (r *CommRepository) ListSequences(ctx context.Context, tenantID string) ([]domain.CommSequence, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, description, trigger, is_active, steps, created_at, updated_at
		FROM comm_sequences
		WHERE tenant_id = $1
		ORDER BY name
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list sequences: %w", err)
	}
	defer rows.Close()

	var items []domain.CommSequence
	for rows.Next() {
		var s domain.CommSequence
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.Name, &s.Description, &s.Trigger, &s.IsActive, &s.Steps, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan sequence: %w", err)
		}
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sequence rows: %w", err)
	}
	return items, nil
}

// ListScheduled returns all scheduled communications.
func (r *CommRepository) ListScheduled(ctx context.Context, tenantID string) ([]domain.CommScheduled, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, guest_name, channel, subject, body, status, scheduled_at, sent_at, error, created_at
		FROM comm_scheduled
		WHERE tenant_id = $1
		ORDER BY scheduled_at DESC
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list scheduled: %w", err)
	}
	defer rows.Close()

	var items []domain.CommScheduled
	for rows.Next() {
		var s domain.CommScheduled
		var subj, body, errStr *string
		var sentAt *time.Time
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.GuestName, &s.Channel, &subj, &body, &s.Status, &s.ScheduledAt, &sentAt, &errStr, &s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan scheduled: %w", err)
		}
		s.Subject = subj
		s.Body = body
		s.SentAt = sentAt
		s.Error = errStr
		items = append(items, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("scheduled rows: %w", err)
	}
	return items, nil
}
