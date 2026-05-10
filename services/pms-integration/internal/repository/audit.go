package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// AuditLogRepository defines audit storage.
type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	List(ctx context.Context, tenantID string, resource, action string, limit, offset int) ([]*domain.AuditLog, error)
	GetByResourceID(ctx context.Context, tenantID, resource, resourceID string) ([]*domain.AuditLog, error)
}

// PostgresAuditLogRepository is a PostgreSQL implementation.
type PostgresAuditLogRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewAuditLogRepository creates a new repository.
func NewAuditLogRepository(pool *db.Pool, metrics *RepositoryMetrics) AuditLogRepository {
	return &PostgresAuditLogRepository{pool: pool, metrics: metrics}
}

func (r *PostgresAuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	defer r.metrics.ObserveQuery("audit_log_create")()
	oldJSON, _ := json.Marshal(log.OldValues)
	newJSON, _ := json.Marshal(log.NewValues)
	query := `INSERT INTO audit_logs (tenant_id, user_id, guest_id, action, resource, resource_id, old_values, new_values, ip_address, user_agent) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, log.TenantID, log.UserID, log.GuestID, log.Action, log.Resource, log.ResourceID, oldJSON, newJSON, log.IPAddress, log.UserAgent).Scan(&log.ID, &log.CreatedAt)
}

func (r *PostgresAuditLogRepository) List(ctx context.Context, tenantID string, resource, action string, limit, offset int) ([]*domain.AuditLog, error) {
	defer r.metrics.ObserveQuery("audit_log_list")()
	query := `SELECT id, tenant_id, user_id, guest_id, action, resource, resource_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argCount := 1
	if resource != "" {
		argCount++
		query += fmt.Sprintf(" AND resource = $%d", argCount)
		args = append(args, resource)
	}
	if action != "" {
		argCount++
		query += fmt.Sprintf(" AND action = $%d", argCount)
		args = append(args, action)
	}
	argCount++
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		var userID, guestID sql.NullString
		var oldJSON, newJSON []byte
		if err := rows.Scan(&l.ID, &l.TenantID, &userID, &guestID, &l.Action, &l.Resource, &l.ResourceID, &oldJSON, &newJSON, &l.IPAddress, &l.UserAgent, &l.CreatedAt); err != nil {
			return nil, err
		}
		if userID.Valid {
			l.UserID = &userID.String
		}
		if guestID.Valid {
			l.GuestID = &guestID.String
		}
		json.Unmarshal(oldJSON, &l.OldValues)
		json.Unmarshal(newJSON, &l.NewValues)
		logs = append(logs, &l)
	}
	return logs, rows.Err()
}

func (r *PostgresAuditLogRepository) GetByResourceID(ctx context.Context, tenantID, resource, resourceID string) ([]*domain.AuditLog, error) {
	defer r.metrics.ObserveQuery("audit_log_by_resource")()
	query := `SELECT id, tenant_id, user_id, guest_id, action, resource, resource_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs WHERE tenant_id = $1 AND resource = $2 AND resource_id = $3 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, tenantID, resource, resourceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*domain.AuditLog
	for rows.Next() {
		var l domain.AuditLog
		var userID, guestID sql.NullString
		var oldJSON, newJSON []byte
		if err := rows.Scan(&l.ID, &l.TenantID, &userID, &guestID, &l.Action, &l.Resource, &l.ResourceID, &oldJSON, &newJSON, &l.IPAddress, &l.UserAgent, &l.CreatedAt); err != nil {
			return nil, err
		}
		if userID.Valid {
			l.UserID = &userID.String
		}
		if guestID.Valid {
			l.GuestID = &guestID.String
		}
		json.Unmarshal(oldJSON, &l.OldValues)
		json.Unmarshal(newJSON, &l.NewValues)
		logs = append(logs, &l)
	}
	return logs, rows.Err()
}
