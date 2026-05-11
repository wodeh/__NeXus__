package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// CommunicationsRepository handles communication templates and scheduling.
type CommunicationsRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewCommunicationsRepository creates a new repository.
func NewCommunicationsRepository(pool *db.Pool, metrics *RepositoryMetrics) *CommunicationsRepository {
	return &CommunicationsRepository{pool: pool, metrics: metrics}
}

// ========== TEMPLATES ==========

func (r *CommunicationsRepository) CreateTemplate(ctx context.Context, tenantID string, t *domain.CommunicationTemplate) (*domain.CommunicationTemplate, error) {
	r.pool.SetTenant(ctx, tenantID)
	err := r.pool.QueryRow(ctx, `
		INSERT INTO communication_templates (tenant_id, name, subject, body, channel, category, variables, is_active, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW()) RETURNING id, created_at, updated_at`,
		tenantID, t.Name, t.Subject, t.Body, t.Channel, t.Category, t.Variables, t.IsActive, t.IsDefault).Scan(
		&t.ID, &t.CreatedAt, &t.UpdatedAt)
	t.TenantID = tenantID
	return t, err
}

func (r *CommunicationsRepository) ListTemplates(ctx context.Context, tenantID string, category *string) ([]*domain.CommunicationTemplate, error) {
	r.pool.SetTenant(ctx, tenantID)
	query := `SELECT id, tenant_id, name, subject, body, channel, category, variables, is_active, is_default, created_at, updated_at FROM communication_templates WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	if category != nil {
		query += ` AND category = $2`
		args = append(args, *category)
	}
	query += ` ORDER BY category, name`

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.CommunicationTemplate
	for rows.Next() {
		var t domain.CommunicationTemplate
		rows.Scan(
			&t.ID, &t.TenantID, &t.Name, &t.Subject, &t.Body, &t.Channel, &t.Category, &t.Variables, &t.IsActive, &t.IsDefault, &t.CreatedAt, &t.UpdatedAt)
		out = append(out, &t)
	}
	return out, nil
}

func (r *CommunicationsRepository) UpdateTemplate(ctx context.Context, tenantID, id string, updates map[string]interface{}) error {
	r.pool.SetTenant(ctx, tenantID)
	query := "UPDATE communication_templates SET updated_at = NOW()"
	args := []interface{}{}
	argCount := 0
	for col, val := range updates {
		argCount++
		query += fmt.Sprintf(", %s = $%d", col, argCount)
		args = append(args, val)
	}
	argCount++
	query += fmt.Sprintf(" WHERE tenant_id = $%d AND id = $%d", argCount, argCount+1)
	args = append(args, tenantID, id)
	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

func (r *CommunicationsRepository) DeleteTemplate(ctx context.Context, tenantID, id string) error {
	r.pool.SetTenant(ctx, tenantID)
	_, err := r.pool.Exec(ctx, `DELETE FROM communication_templates WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

// ========== SEQUENCES ==========

func (r *CommunicationsRepository) CreateSequence(ctx context.Context, tenantID string, s *domain.CommunicationSequence) (*domain.CommunicationSequence, error) {
	r.pool.SetTenant(ctx, tenantID)
	err := r.pool.QueryRow(ctx, `
		INSERT INTO communication_sequences (tenant_id, name, description, trigger, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW()) RETURNING id, created_at, updated_at`,
		tenantID, s.Name, s.Description, s.Trigger, s.IsActive).Scan(
		&s.ID, &s.CreatedAt, &s.UpdatedAt)
	s.TenantID = tenantID
	return s, err
}

func (r *CommunicationsRepository) ListSequences(ctx context.Context, tenantID string) ([]*domain.CommunicationSequence, error) {
	r.pool.SetTenant(ctx, tenantID)
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, description, trigger, is_active, created_at, updated_at
		FROM communication_sequences WHERE tenant_id = $1 ORDER BY name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.CommunicationSequence
	for rows.Next() {
		var s domain.CommunicationSequence
		rows.Scan(&s.ID, &s.TenantID, &s.Name, &s.Description, &s.Trigger, &s.IsActive, &s.CreatedAt, &s.UpdatedAt)
		out = append(out, &s)
	}
	return out, nil
}

func (r *CommunicationsRepository) AddSequenceStep(ctx context.Context, tenantID string, step *domain.CommunicationSequenceStep) (*domain.CommunicationSequenceStep, error) {
	r.pool.SetTenant(ctx, tenantID)
	err := r.pool.QueryRow(ctx, `
		INSERT INTO communication_sequence_steps (sequence_id, template_id, step_order, delay_hours, channel, conditions_json, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW()) RETURNING id, created_at`,
		step.SequenceID, step.TemplateID, step.StepOrder, step.DelayHours, step.Channel, step.ConditionsJSON).Scan(
		&step.ID, &step.CreatedAt)
	return step, err
}

func (r *CommunicationsRepository) GetSequenceSteps(ctx context.Context, tenantID, sequenceID string) ([]*domain.CommunicationSequenceStep, error) {
	r.pool.SetTenant(ctx, tenantID)
	rows, err := r.pool.Query(ctx, `
		SELECT id, sequence_id, template_id, step_order, delay_hours, channel, conditions_json, created_at
		FROM communication_sequence_steps WHERE sequence_id = $1 ORDER BY step_order`, sequenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.CommunicationSequenceStep
	for rows.Next() {
		var s domain.CommunicationSequenceStep
		rows.Scan(&s.ID, &s.SequenceID, &s.TemplateID, &s.StepOrder, &s.DelayHours, &s.Channel, &s.ConditionsJSON, &s.CreatedAt)
		out = append(out, &s)
	}
	return out, nil
}

// ========== SCHEDULED COMMUNICATIONS ==========

func (r *CommunicationsRepository) ScheduleCommunication(ctx context.Context, tenantID string, c *domain.ScheduledCommunication) (*domain.ScheduledCommunication, error) {
	r.pool.SetTenant(ctx, tenantID)
	err := r.pool.QueryRow(ctx, `
		INSERT INTO scheduled_communications (tenant_id, reservation_id, guest_id, guest_phone, guest_email, template_id, sequence_id, sequence_step_id, channel, subject, body_rendered, status, scheduled_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW()) RETURNING id, created_at, updated_at`,
		tenantID, c.ReservationID, c.GuestID, c.GuestPhone, c.GuestEmail, c.TemplateID, c.SequenceID, c.SequenceStepID, c.Channel, c.Subject, c.BodyRendered, c.Status, c.ScheduledAt).Scan(
		&c.ID, &c.CreatedAt, &c.UpdatedAt)
	c.TenantID = tenantID
	return c, err
}

func (r *CommunicationsRepository) ListScheduledCommunications(ctx context.Context, tenantID string, status *string, limit int) ([]*domain.ScheduledCommunication, error) {
	r.pool.SetTenant(ctx, tenantID)
	query := `SELECT id, tenant_id, reservation_id, guest_id, guest_phone, guest_email, template_id, sequence_id, sequence_step_id, channel, subject, body_rendered, status, scheduled_at, sent_at, delivered_at, failed_at, error_message, created_at, updated_at FROM scheduled_communications WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	if status != nil {
		query += ` AND status = $2`
		args = append(args, *status)
	}
	query += ` ORDER BY scheduled_at ASC LIMIT $` + string(rune('0'+len(args)+1))
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.ScheduledCommunication
	for rows.Next() {
		var c domain.ScheduledCommunication
		rows.Scan(
			&c.ID, &c.TenantID, &c.ReservationID, &c.GuestID, &c.GuestPhone, &c.GuestEmail, &c.TemplateID, &c.SequenceID, &c.SequenceStepID, &c.Channel, &c.Subject, &c.BodyRendered, &c.Status, &c.ScheduledAt, &c.SentAt, &c.DeliveredAt, &c.FailedAt, &c.ErrorMessage, &c.CreatedAt, &c.UpdatedAt)
		out = append(out, &c)
	}
	return out, nil
}

func (r *CommunicationsRepository) UpdateScheduledStatus(ctx context.Context, tenantID, id, status string, sentAt *time.Time) error {
	r.pool.SetTenant(ctx, tenantID)
	_, err := r.pool.Exec(ctx, `
		UPDATE scheduled_communications SET status = $1, sent_at = $2, updated_at = NOW() WHERE id = $3 AND tenant_id = $4`,
		status, sentAt, id, tenantID)
	return err
}

// ========== LOGS ==========

func (r *CommunicationsRepository) LogCommunication(ctx context.Context, tenantID string, log *domain.CommunicationLog) (*domain.CommunicationLog, error) {
	r.pool.SetTenant(ctx, tenantID)
	err := r.pool.QueryRow(ctx, `
		INSERT INTO communication_logs (tenant_id, guest_id, reservation_id, channel, direction, subject, body, status, external_id, sent_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW()) RETURNING id, created_at`,
		tenantID, log.GuestID, log.ReservationID, log.Channel, log.Direction, log.Subject, log.Body, log.Status, log.ExternalID, log.SentAt).Scan(
		&log.ID, &log.CreatedAt)
	log.TenantID = tenantID
	return log, err
}

func (r *CommunicationsRepository) ListLogs(ctx context.Context, tenantID string, limit int) ([]*domain.CommunicationLog, error) {
	r.pool.SetTenant(ctx, tenantID)
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, guest_id, reservation_id, channel, direction, subject, body, status, external_id, sent_at, created_at
		FROM communication_logs WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.CommunicationLog
	for rows.Next() {
		var l domain.CommunicationLog
		rows.Scan(&l.ID, &l.TenantID, &l.GuestID, &l.ReservationID, &l.Channel, &l.Direction, &l.Subject, &l.Body, &l.Status, &l.ExternalID, &l.SentAt, &l.CreatedAt)
		out = append(out, &l)
	}
	return out, nil
}
