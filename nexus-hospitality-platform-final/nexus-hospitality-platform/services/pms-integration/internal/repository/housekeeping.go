// Package repository provides tenant-aware data access for housekeeping and maintenance.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// HousekeepingTaskRepository persists and retrieves HousekeepingTask aggregates.
type HousekeepingTaskRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.HousekeepingTask, error)
	Create(ctx context.Context, t *domain.HousekeepingTask) error
	Update(ctx context.Context, t *domain.HousekeepingTask) error
	Delete(ctx context.Context, tenantID, id string) error
	ListByProperty(ctx context.Context, tenantID, propertyID string, status string, limit, offset int) ([]*domain.HousekeepingTask, error)
	ListByAssignedTo(ctx context.Context, tenantID, assignedTo string, status string, limit, offset int) ([]*domain.HousekeepingTask, error)
	ListBoard(ctx context.Context, tenantID, propertyID string) ([]*domain.HousekeepingTask, error)
	UpdateStatus(ctx context.Context, tenantID, id string, status domain.TaskStatus) error
}

// PostgresHousekeepingTaskRepository is the PostgreSQL implementation.
type PostgresHousekeepingTaskRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewHousekeepingTaskRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresHousekeepingTaskRepository {
	return &PostgresHousekeepingTaskRepository{pool: pool, metrics: metrics}
}

func (r *PostgresHousekeepingTaskRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresHousekeepingTaskRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.HousekeepingTask, error) {
	start := time.Now()
	r.metrics.IncQuery("housekeeping_task", "get_by_id")
	defer r.metrics.ObserveDuration("housekeeping_task", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, room_id, task_type, priority, status,
		       assigned_to, scheduled_at, started_at, completed_at,
		       estimated_duration_minutes, actual_duration_minutes,
		       notes, inspection_score, inspection_notes, inspected_by,
		       created_by, created_at, updated_at, version
		FROM housekeeping_tasks
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID)

	var t domain.HousekeepingTask
	var scheduledAt, startedAt, completedAt *time.Time
	err := row.Scan(
		&t.ID, &t.TenantID, &t.PropertyID, &t.RoomID, &t.TaskType, &t.Priority, &t.Status,
		&t.AssignedTo, &scheduledAt, &startedAt, &completedAt,
		&t.EstimatedDuration, &t.ActualDuration,
		&t.Notes, &t.InspectionScore, &t.InspectionNotes, &t.InspectedBy,
		&t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &t.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("housekeeping task not found: %w", err)
	}
	t.ScheduledAt = scheduledAt
	t.StartedAt = startedAt
	t.CompletedAt = completedAt
	return &t, nil
}

func (r *PostgresHousekeepingTaskRepository) Create(ctx context.Context, t *domain.HousekeepingTask) error {
	start := time.Now()
	r.metrics.IncQuery("housekeeping_task", "create")
	defer r.metrics.ObserveDuration("housekeeping_task", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, t.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO housekeeping_tasks (id, tenant_id, property_id, room_id, task_type, priority, status,
		                               assigned_to, scheduled_at, estimated_duration_minutes, notes,
		                               created_by, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, t.ID, t.TenantID, t.PropertyID, t.RoomID, t.TaskType, t.Priority, t.Status,
		t.AssignedTo, t.ScheduledAt, t.EstimatedDuration, t.Notes,
		t.CreatedBy, t.Version, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *PostgresHousekeepingTaskRepository) Update(ctx context.Context, t *domain.HousekeepingTask) error {
	start := time.Now()
	r.metrics.IncQuery("housekeeping_task", "update")
	defer r.metrics.ObserveDuration("housekeeping_task", "update", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, t.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE housekeeping_tasks
		SET status = $1, assigned_to = $2, scheduled_at = $3, started_at = $4, completed_at = $5,
		    estimated_duration_minutes = $6, actual_duration_minutes = $7,
		    notes = $8, inspection_score = $9, inspection_notes = $10,
		    inspected_by = $11, version = $12, updated_at = $13
		WHERE id = $14 AND tenant_id = $15 AND version = $16 AND deleted_at IS NULL
	`, t.Status, t.AssignedTo, t.ScheduledAt, t.StartedAt, t.CompletedAt,
		t.EstimatedDuration, t.ActualDuration,
		t.Notes, t.InspectionScore, t.InspectionNotes,
		t.InspectedBy, t.Version, t.UpdatedAt,
		t.ID, t.TenantID, t.Version-1)
	return err
}

func (r *PostgresHousekeepingTaskRepository) Delete(ctx context.Context, tenantID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE housekeeping_tasks
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID)
	return err
}

func (r *PostgresHousekeepingTaskRepository) ListByProperty(ctx context.Context, tenantID, propertyID string, status string, limit, offset int) ([]*domain.HousekeepingTask, error) {
	start := time.Now()
	r.metrics.IncQuery("housekeeping_task", "list_by_property")
	defer r.metrics.ObserveDuration("housekeeping_task", "list_by_property", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	var rows pgx.Rows
	var err error
	if status != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, property_id, room_id, task_type, priority, status,
			       assigned_to, scheduled_at, started_at, completed_at,
			       estimated_duration_minutes, actual_duration_minutes,
			       notes, inspection_score, inspection_notes, inspected_by,
			       created_by, created_at, updated_at, version
			FROM housekeeping_tasks
			WHERE tenant_id = $1 AND property_id = $2 AND status = $3 AND deleted_at IS NULL
			ORDER BY scheduled_at ASC, priority DESC
			LIMIT $4 OFFSET $5
		`, tenantID, propertyID, status, limit, offset)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, property_id, room_id, task_type, priority, status,
			       assigned_to, scheduled_at, started_at, completed_at,
			       estimated_duration_minutes, actual_duration_minutes,
			       notes, inspection_score, inspection_notes, inspected_by,
			       created_by, created_at, updated_at, version
			FROM housekeeping_tasks
			WHERE tenant_id = $1 AND property_id = $2 AND deleted_at IS NULL
			ORDER BY scheduled_at ASC, priority DESC
			LIMIT $3 OFFSET $4
		`, tenantID, propertyID, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHousekeepingTasks(rows)
}

func (r *PostgresHousekeepingTaskRepository) ListByAssignedTo(ctx context.Context, tenantID, assignedTo string, status string, limit, offset int) ([]*domain.HousekeepingTask, error) {
	start := time.Now()
	r.metrics.IncQuery("housekeeping_task", "list_by_assigned")
	defer r.metrics.ObserveDuration("housekeeping_task", "list_by_assigned", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	var rows pgx.Rows
	var err error
	if status != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, property_id, room_id, task_type, priority, status,
			       assigned_to, scheduled_at, started_at, completed_at,
			       estimated_duration_minutes, actual_duration_minutes,
			       notes, inspection_score, inspection_notes, inspected_by,
			       created_by, created_at, updated_at, version
			FROM housekeeping_tasks
			WHERE tenant_id = $1 AND assigned_to = $2 AND status = $3 AND deleted_at IS NULL
			ORDER BY scheduled_at ASC, priority DESC
			LIMIT $4 OFFSET $5
		`, tenantID, assignedTo, status, limit, offset)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, property_id, room_id, task_type, priority, status,
			       assigned_to, scheduled_at, started_at, completed_at,
			       estimated_duration_minutes, actual_duration_minutes,
			       notes, inspection_score, inspection_notes, inspected_by,
			       created_by, created_at, updated_at, version
			FROM housekeeping_tasks
			WHERE tenant_id = $1 AND assigned_to = $2 AND deleted_at IS NULL
			ORDER BY scheduled_at ASC, priority DESC
			LIMIT $3 OFFSET $4
		`, tenantID, assignedTo, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHousekeepingTasks(rows)
}

func (r *PostgresHousekeepingTaskRepository) ListBoard(ctx context.Context, tenantID, propertyID string) ([]*domain.HousekeepingTask, error) {
	start := time.Now()
	r.metrics.IncQuery("housekeeping_task", "list_board")
	defer r.metrics.ObserveDuration("housekeeping_task", "list_board", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, room_id, task_type, priority, status,
		       assigned_to, scheduled_at, started_at, completed_at,
		       estimated_duration_minutes, actual_duration_minutes,
		       notes, inspection_score, inspection_notes, inspected_by,
		       created_by, created_at, updated_at, version
		FROM housekeeping_tasks
		WHERE tenant_id = $1 AND property_id = $2 AND status NOT IN ('completed', 'inspected') AND deleted_at IS NULL
		ORDER BY 
			CASE priority
				WHEN 'urgent' THEN 1
				WHEN 'high' THEN 2
				WHEN 'normal' THEN 3
				WHEN 'low' THEN 4
			END,
			scheduled_at ASC
	`, tenantID, propertyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanHousekeepingTasks(rows)
}

func (r *PostgresHousekeepingTaskRepository) UpdateStatus(ctx context.Context, tenantID, id string, status domain.TaskStatus) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}

	var startedAt, completedAt interface{}
	if status == domain.TaskStatusInProgress {
		startedAt = time.Now().UTC()
	}
	if status == domain.TaskStatusCompleted {
		completedAt = time.Now().UTC()
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE housekeeping_tasks
		SET status = $1, started_at = COALESCE($2, started_at), completed_at = COALESCE($3, completed_at),
		    updated_at = NOW(), version = version + 1
		WHERE id = $4 AND tenant_id = $5 AND deleted_at IS NULL
	`, status, startedAt, completedAt, id, tenantID)
	return err
}

func scanHousekeepingTasks(rows pgx.Rows) ([]*domain.HousekeepingTask, error) {
	var tasks []*domain.HousekeepingTask
	for rows.Next() {
		var t domain.HousekeepingTask
		var scheduledAt, startedAt, completedAt *time.Time
		err := rows.Scan(
			&t.ID, &t.TenantID, &t.PropertyID, &t.RoomID, &t.TaskType, &t.Priority, &t.Status,
			&t.AssignedTo, &scheduledAt, &startedAt, &completedAt,
			&t.EstimatedDuration, &t.ActualDuration,
			&t.Notes, &t.InspectionScore, &t.InspectionNotes, &t.InspectedBy,
			&t.CreatedBy, &t.CreatedAt, &t.UpdatedAt, &t.Version,
		)
		if err != nil {
			return nil, err
		}
		t.ScheduledAt = scheduledAt
		t.StartedAt = startedAt
		t.CompletedAt = completedAt
		tasks = append(tasks, &t)
	}
	return tasks, rows.Err()
}

// ==================== WORK ORDER REPOSITORY ====================

// WorkOrderRepository persists and retrieves WorkOrder aggregates.
type WorkOrderRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.WorkOrder, error)
	Create(ctx context.Context, wo *domain.WorkOrder) error
	Update(ctx context.Context, wo *domain.WorkOrder) error
	Delete(ctx context.Context, tenantID, id string) error
	ListByProperty(ctx context.Context, tenantID, propertyID string, status string, limit, offset int) ([]*domain.WorkOrder, error)
	ListByStatus(ctx context.Context, tenantID string, status string, limit, offset int) ([]*domain.WorkOrder, error)
}

// PostgresWorkOrderRepository is the PostgreSQL implementation.
type PostgresWorkOrderRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewWorkOrderRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresWorkOrderRepository {
	return &PostgresWorkOrderRepository{pool: pool, metrics: metrics}
}

func (r *PostgresWorkOrderRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresWorkOrderRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.WorkOrder, error) {
	start := time.Now()
	r.metrics.IncQuery("work_order", "get_by_id")
	defer r.metrics.ObserveDuration("work_order", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, room_id, asset_id, work_order_number, category,
		       priority, status, title, description, reported_by, assigned_to,
		       estimated_cost, actual_cost, scheduled_date, started_at, completed_at,
		       resolution_notes, parts_used, created_at, updated_at, version
		FROM work_orders
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID)

	var wo domain.WorkOrder
	var roomID, assetID, scheduledDate, startedAt, completedAt *string
	err := row.Scan(
		&wo.ID, &wo.TenantID, &wo.PropertyID, &roomID, &assetID, &wo.WorkOrderNumber, &wo.Category,
		&wo.Priority, &wo.Status, &wo.Title, &wo.Description, &wo.ReportedBy, &wo.AssignedTo,
		&wo.EstimatedCost, &wo.ActualCost, &scheduledDate, &startedAt, &completedAt,
		&wo.ResolutionNotes, &wo.PartsUsed, &wo.CreatedAt, &wo.UpdatedAt, &wo.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("work order not found: %w", err)
	}
	if roomID != nil { wo.RoomID = roomID }
	if assetID != nil { wo.AssetID = assetID }
	if scheduledDate != nil { t, _ := time.Parse("2006-01-02", *scheduledDate); wo.ScheduledDate = &t }
	if startedAt != nil { t, _ := time.Parse(time.RFC3339, *startedAt); wo.StartedAt = &t }
	if completedAt != nil { t, _ := time.Parse(time.RFC3339, *completedAt); wo.CompletedAt = &t }
	return &wo, nil
}

func (r *PostgresWorkOrderRepository) Create(ctx context.Context, wo *domain.WorkOrder) error {
	start := time.Now()
	r.metrics.IncQuery("work_order", "create")
	defer r.metrics.ObserveDuration("work_order", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, wo.TenantID); err != nil {
		return err
	}

	var roomID, assetID interface{}
	if wo.RoomID != nil { roomID = *wo.RoomID }
	if wo.AssetID != nil { assetID = *wo.AssetID }

	_, err := r.pool.Exec(ctx, `
		INSERT INTO work_orders (id, tenant_id, property_id, room_id, asset_id, work_order_number,
		                        category, priority, status, title, description,
		                        reported_by, estimated_cost, scheduled_date, parts_used,
		                        version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`, wo.ID, wo.TenantID, wo.PropertyID, roomID, assetID, wo.WorkOrderNumber,
		wo.Category, wo.Priority, wo.Status, wo.Title, wo.Description,
		wo.ReportedBy, wo.EstimatedCost, wo.ScheduledDate, wo.PartsUsed,
		wo.Version, wo.CreatedAt, wo.UpdatedAt)
	return err
}

func (r *PostgresWorkOrderRepository) Update(ctx context.Context, wo *domain.WorkOrder) error {
	start := time.Now()
	r.metrics.IncQuery("work_order", "update")
	defer r.metrics.ObserveDuration("work_order", "update", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, wo.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE work_orders
		SET priority = $1, status = $2, title = $3, description = $4,
		    assigned_to = $5, estimated_cost = $6, actual_cost = $7,
		    scheduled_date = $8, started_at = $9, completed_at = $10,
		    resolution_notes = $11, parts_used = $12,
		    version = $13, updated_at = $14
		WHERE id = $15 AND tenant_id = $16 AND version = $17 AND deleted_at IS NULL
	`, wo.Priority, wo.Status, wo.Title, wo.Description,
		wo.AssignedTo, wo.EstimatedCost, wo.ActualCost,
		wo.ScheduledDate, wo.StartedAt, wo.CompletedAt,
		wo.ResolutionNotes, wo.PartsUsed,
		wo.Version, wo.UpdatedAt,
		wo.ID, wo.TenantID, wo.Version-1)
	return err
}

func (r *PostgresWorkOrderRepository) Delete(ctx context.Context, tenantID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE work_orders
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID)
	return err
}

func (r *PostgresWorkOrderRepository) ListByProperty(ctx context.Context, tenantID, propertyID string, status string, limit, offset int) ([]*domain.WorkOrder, error) {
	start := time.Now()
	r.metrics.IncQuery("work_order", "list_by_property")
	defer r.metrics.ObserveDuration("work_order", "list_by_property", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	var rows pgx.Rows
	var err error
	if status != "" {
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, property_id, room_id, asset_id, work_order_number, category,
			       priority, status, title, description, reported_by, assigned_to,
			       estimated_cost, actual_cost, scheduled_date, started_at, completed_at,
			       resolution_notes, parts_used, created_at, updated_at, version
			FROM work_orders
			WHERE tenant_id = $1 AND property_id = $2 AND status = $3 AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $4 OFFSET $5
		`, tenantID, propertyID, status, limit, offset)
	} else {
		rows, err = r.pool.Query(ctx, `
			SELECT id, tenant_id, property_id, room_id, asset_id, work_order_number, category,
			       priority, status, title, description, reported_by, assigned_to,
			       estimated_cost, actual_cost, scheduled_date, started_at, completed_at,
			       resolution_notes, parts_used, created_at, updated_at, version
			FROM work_orders
			WHERE tenant_id = $1 AND property_id = $2 AND deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $3 OFFSET $4
		`, tenantID, propertyID, limit, offset)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanWorkOrders(rows)
}

func (r *PostgresWorkOrderRepository) ListByStatus(ctx context.Context, tenantID string, status string, limit, offset int) ([]*domain.WorkOrder, error) {
	start := time.Now()
	r.metrics.IncQuery("work_order", "list_by_status")
	defer r.metrics.ObserveDuration("work_order", "list_by_status", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, room_id, asset_id, work_order_number, category,
		       priority, status, title, description, reported_by, assigned_to,
		       estimated_cost, actual_cost, scheduled_date, started_at, completed_at,
		       resolution_notes, parts_used, created_at, updated_at, version
		FROM work_orders
		WHERE tenant_id = $1 AND status = $2 AND deleted_at IS NULL
		ORDER BY 
			CASE priority
				WHEN 'critical' THEN 1
				WHEN 'high' THEN 2
				WHEN 'medium' THEN 3
				WHEN 'low' THEN 4
			END,
			created_at DESC
		LIMIT $3 OFFSET $4
	`, tenantID, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanWorkOrders(rows)
}

func scanWorkOrders(rows pgx.Rows) ([]*domain.WorkOrder, error) {
	var orders []*domain.WorkOrder
	for rows.Next() {
		var wo domain.WorkOrder
		var roomID, assetID, scheduledDate, startedAt, completedAt *string
		err := rows.Scan(
			&wo.ID, &wo.TenantID, &wo.PropertyID, &roomID, &assetID, &wo.WorkOrderNumber, &wo.Category,
			&wo.Priority, &wo.Status, &wo.Title, &wo.Description, &wo.ReportedBy, &wo.AssignedTo,
			&wo.EstimatedCost, &wo.ActualCost, &scheduledDate, &startedAt, &completedAt,
			&wo.ResolutionNotes, &wo.PartsUsed, &wo.CreatedAt, &wo.UpdatedAt, &wo.Version,
		)
		if err != nil {
			return nil, err
		}
		if roomID != nil { wo.RoomID = roomID }
		if assetID != nil { wo.AssetID = assetID }
		if scheduledDate != nil { t, _ := time.Parse("2006-01-02", *scheduledDate); wo.ScheduledDate = &t }
		if startedAt != nil { t, _ := time.Parse(time.RFC3339, *startedAt); wo.StartedAt = &t }
		if completedAt != nil { t, _ := time.Parse(time.RFC3339, *completedAt); wo.CompletedAt = &t }
		orders = append(orders, &wo)
	}
	return orders, rows.Err()
}
