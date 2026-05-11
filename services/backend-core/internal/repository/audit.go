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

// AuditRepository provides audit log data access.
type AuditRepository struct {
	pool *db.Pool
}

// NewAuditRepository creates an audit repository.
func NewAuditRepository(pool *db.Pool) *AuditRepository {
	return &AuditRepository{pool: pool}
}

func (r *AuditRepository) execer(tx pgx.Tx) interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
} {
	if tx != nil {
		return tx
	}
	return r.pool
}

// List returns audit logs for a tenant.
func (r *AuditRepository) List(ctx context.Context, tenantID string, limit int) ([]domain.AuditLog, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, user_id, user_email, action, resource, resource_id,
			details, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, tenantID, limit)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		var userID *uuid.UUID
		var userEmail *string
		var detailsJSON []byte
		if err := rows.Scan(
			&l.ID, &l.TenantID, &userID, &userEmail, &l.Action, &l.Resource, &l.ResourceID,
			&detailsJSON, &l.IPAddress, &l.UserAgent, &l.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit log: %w", err)
		}
		l.UserID = userID
		l.UserEmail = userEmail
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("audit rows: %w", err)
	}
	return logs, nil
}

// Create inserts a new audit log.
func (r *AuditRepository) Create(ctx context.Context, tenantID string, l *domain.AuditLog) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO audit_logs (tenant_id, user_id, user_email, action, resource, resource_id,
			details, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`, tenantID, l.UserID, l.UserEmail, l.Action, l.Resource, l.ResourceID,
		l.Details, l.IPAddress, l.UserAgent, l.CreatedAt,
	).Scan(&l.ID)
}

// GetStats returns audit statistics.
func (r *AuditRepository) GetStats(ctx context.Context, tenantID string) (*domain.AuditStats, error) {
	var s domain.AuditStats
	s.TenantID = tenantID
	err := r.pool.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE created_at >= CURRENT_DATE),
			COUNT(*) FILTER (WHERE user_id IS NOT NULL),
			COUNT(*) FILTER (WHERE user_id IS NULL),
			COUNT(*) FILTER (WHERE action LIKE '%failed%')
		FROM audit_logs
		WHERE tenant_id = $1
	`, tenantID).Scan(
		&s.TotalEvents, &s.TodayEvents, &s.UserActions, &s.SystemActions, &s.FailedActions,
	)
	if err != nil {
		return nil, fmt.Errorf("audit stats: %w", err)
	}
	return &s, nil
}
