package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// HousekeepingRepository manages housekeeping tasks, staff, and supplies.
type HousekeepingRepository struct {
	pool *pgxpool.Pool
}

// NewHousekeepingRepository creates a new housekeeping repository.
func NewHousekeepingRepository(pool *pgxpool.Pool) *HousekeepingRepository {
	return &HousekeepingRepository{pool: pool}
}

// ===================== TASKS =====================

// ListTasks returns tasks for a tenant with optional filters.
func (r *HousekeepingRepository) ListTasks(ctx context.Context, tenantID string, status, floor, staffID, shift string, limit, offset int) ([]domain.HousekeepingTask, error) {
	query := `
		SELECT id, tenant_id, room_id, room_number, floor, task_type, priority, status,
			assigned_to, assigned_by, shift, estimated_minutes, actual_minutes,
			started_at, completed_at, checklist_done, supplies_used, notes, damage_report, card_mode, created_at, updated_at
		FROM housekeeping_tasks
		WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argCount := 1

	if status != "" {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, status)
	}
	if floor != "" {
		argCount++
		query += fmt.Sprintf(" AND floor = $%d", argCount)
		args = append(args, floor)
	}
	if staffID != "" {
		argCount++
		query += fmt.Sprintf(" AND assigned_to = $%d", argCount)
		args = append(args, staffID)
	}
	if shift != "" {
		argCount++
		query += fmt.Sprintf(" AND shift = $%d", argCount)
		args = append(args, shift)
	}

	query += " ORDER BY priority DESC, created_at ASC"
	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
		if offset > 0 {
			argCount++
			query += fmt.Sprintf(" OFFSET $%d", argCount)
			args = append(args, offset)
		}
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.HousekeepingTask])
}

// GetTask fetches a single task by ID.
func (r *HousekeepingRepository) GetTask(ctx context.Context, tenantID, taskID string) (*domain.HousekeepingTask, error) {
	query := `
		SELECT id, tenant_id, room_id, room_number, floor, task_type, priority, status,
			assigned_to, assigned_by, shift, estimated_minutes, actual_minutes,
			started_at, completed_at, checklist_done, supplies_used, notes, damage_report, card_mode, created_at, updated_at
		FROM housekeeping_tasks
		WHERE tenant_id = $1 AND id = $2`
	row, err := r.pool.Query(ctx, query, tenantID, taskID)
	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	defer row.Close()
	task, err := pgx.CollectOneRow(row, pgx.RowToStructByName[domain.HousekeepingTask])
	if err != nil {
		return nil, fmt.Errorf("get task collect: %w", err)
	}
	return &task, nil
}

// CreateTask inserts a new housekeeping task.
func (r *HousekeepingRepository) CreateTask(ctx context.Context, task *domain.HousekeepingTask) error {
	query := `
		INSERT INTO housekeeping_tasks (tenant_id, room_id, room_number, floor, task_type, priority, status,
			assigned_to, assigned_by, shift, estimated_minutes, actual_minutes, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query,
		task.TenantID, task.RoomID, task.RoomNumber, task.Floor, task.TaskType, task.Priority, task.Status,
		task.AssignedTo, task.AssignedBy, task.Shift, task.EstimatedMin, task.ActualMin, task.Notes,
	).Scan(&task.ID, &task.CreatedAt, &task.UpdatedAt)
}

// UpdateTask updates task fields.
func (r *HousekeepingRepository) UpdateTask(ctx context.Context, tenantID, taskID string, updates map[string]interface{}) error {
	// Build dynamic update
	query := "UPDATE housekeeping_tasks SET updated_at = NOW()"
	args := []interface{}{}
	argCount := 0

	for col, val := range updates {
		argCount++
		query += fmt.Sprintf(", %s = $%d", col, argCount)
		args = append(args, val)
	}
	argCount++
	query += fmt.Sprintf(" WHERE tenant_id = $%d AND id = $%d", argCount, argCount+1)
	args = append(args, tenantID, taskID)

	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// DeleteTask removes a task.
func (r *HousekeepingRepository) DeleteTask(ctx context.Context, tenantID, taskID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM housekeeping_tasks WHERE tenant_id = $1 AND id = $2`, tenantID, taskID)
	return err
}

// AssignTask assigns a task to a staff member.
func (r *HousekeepingRepository) AssignTask(ctx context.Context, tenantID, taskID, staffID, assignedBy string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE housekeeping_tasks SET assigned_to = $1, assigned_by = $2, updated_at = NOW() WHERE tenant_id = $3 AND id = $4`,
		staffID, assignedBy, tenantID, taskID,
	)
	return err
}

// StartTask marks a task as in_progress with timestamp.
func (r *HousekeepingRepository) StartTask(ctx context.Context, tenantID, taskID, cardMode string) error {
	query := `UPDATE housekeeping_tasks SET status = 'in_progress', started_at = NOW(), card_mode = $1, updated_at = NOW() WHERE tenant_id = $2 AND id = $3`
	_, err := r.pool.Exec(ctx, query, cardMode, tenantID, taskID)
	return err
}

// CompleteTask marks a task done with actual duration.
func (r *HousekeepingRepository) CompleteTask(ctx context.Context, tenantID, taskID string, actualMin int, checklist []string, supplies []domain.SupplyUsage, notes string) error {
	query := `
		UPDATE housekeeping_tasks
		SET status = 'completed', completed_at = NOW(), actual_minutes = $1,
			checklist_done = $2, supplies_used = $3, notes = $4, updated_at = NOW()
		WHERE tenant_id = $5 AND id = $6`
	_, err := r.pool.Exec(ctx, query, actualMin, checklist, supplies, notes, tenantID, taskID)
	return err
}

// ===================== STAFF =====================

// ListStaff returns housekeeping staff for a tenant.
func (r *HousekeepingRepository) ListStaff(ctx context.Context, tenantID string, activeOnly bool, shift string) ([]domain.HousekeepingStaff, error) {
	query := `SELECT id, tenant_id, user_id, name, role, active, shift, floors, max_rooms_per_day, current_load, rating, completed_today, created_at
		FROM housekeeping_staff WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	if activeOnly {
		query += " AND active = true"
	}
	if shift != "" {
		query += " AND shift = $2"
		args = append(args, shift)
	}
	query += " ORDER BY name ASC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.HousekeepingStaff])
}

// CreateStaff adds a new staff member.
func (r *HousekeepingRepository) CreateStaff(ctx context.Context, staff *domain.HousekeepingStaff) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO housekeeping_staff (tenant_id, user_id, name, role, active, shift, floors, max_rooms_per_day)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`,
		staff.TenantID, staff.UserID, staff.Name, staff.Role, staff.Active, staff.Shift, staff.Floors, staff.MaxRoomsPerDay,
	).Scan(&staff.ID, &staff.CreatedAt)
}

// UpdateStaff updates staff fields.
func (r *HousekeepingRepository) UpdateStaff(ctx context.Context, tenantID, staffID string, updates map[string]interface{}) error {
	query := "UPDATE housekeeping_staff SET"
	args := []interface{}{}
	argCount := 0
	for col, val := range updates {
		argCount++
		if argCount > 1 {
			query += ","
		}
		query += fmt.Sprintf(" %s = $%d", col, argCount)
		args = append(args, val)
	}
	argCount++
	query += fmt.Sprintf(" WHERE tenant_id = $%d AND id = $%d", argCount, argCount+1)
	args = append(args, tenantID, staffID)
	_, err := r.pool.Exec(ctx, query, args...)
	return err
}

// DeleteStaff removes a staff member.
func (r *HousekeepingRepository) DeleteStaff(ctx context.Context, tenantID, staffID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM housekeeping_staff WHERE tenant_id = $1 AND id = $2`, tenantID, staffID)
	return err
}

// ===================== CHECKLISTS =====================

// GetChecklistTemplate returns checklist for a room type + task type.
func (r *HousekeepingRepository) GetChecklistTemplate(ctx context.Context, tenantID, roomType, taskType string) (*domain.CleaningChecklistTemplate, error) {
	query := `SELECT id, tenant_id, room_type, task_type, items, created_at
		FROM cleaning_checklist_templates WHERE tenant_id = $1 AND room_type = $2 AND task_type = $3`
	row, err := r.pool.Query(ctx, query, tenantID, roomType, taskType)
	if err != nil {
		return nil, err
	}
	defer row.Close()
	tmpl, err := pgx.CollectOneRow(row, pgx.RowToStructByName[domain.CleaningChecklistTemplate])
	if err != nil {
		return nil, err
	}
	return &tmpl, nil
}

// SaveChecklistTemplate upserts a checklist template.
func (r *HousekeepingRepository) SaveChecklistTemplate(ctx context.Context, tmpl *domain.CleaningChecklistTemplate) error {
	query := `
		INSERT INTO cleaning_checklist_templates (tenant_id, room_type, task_type, items)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (tenant_id, room_type, task_type) DO UPDATE SET items = EXCLUDED.items
		RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, tmpl.TenantID, tmpl.RoomType, tmpl.TaskType, tmpl.Items).Scan(&tmpl.ID, &tmpl.CreatedAt)
}

// ===================== CLEANER CARDS =====================

// ListCleanerCards returns all active cleaner cards for a tenant.
func (r *HousekeepingRepository) ListCleanerCards(ctx context.Context, tenantID string) ([]domain.LockAccessCodeWithMode, error) {
	query := `SELECT id, card_code, cleaner_mode, valid_from, valid_until, floor, room_ids, staff_id
		FROM cleaner_card_access WHERE tenant_id = $1 AND is_active = true ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []domain.LockAccessCodeWithMode
	for rows.Next() {
		var c domain.LockAccessCodeWithMode
		var id string
		var staffID string
		err := rows.Scan(&id, &c.Code, &c.CleanerMode, &c.ValidFrom, &c.ValidUntil, &c.Floor, &c.RoomIDs, &staffID)
		if err != nil {
			continue
		}
		c.AccessCodeID = id
		c.AssignedTo = staffID
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

// CreateCleanerCard adds a new cleaner card.
func (r *HousekeepingRepository) CreateCleanerCard(ctx context.Context, tenantID string, card *domain.LockAccessCodeWithMode) error {
	return r.pool.QueryRow(ctx,
		`INSERT INTO cleaner_card_access (tenant_id, staff_id, card_code, cleaner_mode, valid_from, valid_until, floor, room_ids)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`,
		tenantID, card.AssignedTo, card.Code, card.CleanerMode, card.ValidFrom, card.ValidUntil, card.Floor, card.RoomIDs,
	).Scan(&card.AccessCodeID, &card.ValidFrom)
}

// RevokeCleanerCard deactivates a card.
func (r *HousekeepingRepository) RevokeCleanerCard(ctx context.Context, tenantID, cardID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE cleaner_card_access SET is_active = false WHERE tenant_id = $1 AND id = $2`,
		tenantID, cardID,
	)
	return err
}

// RecordCardUse updates last_used_at for a card.
func (r *HousekeepingRepository) RecordCardUse(ctx context.Context, tenantID, cardID string) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE cleaner_card_access SET last_used_at = NOW() WHERE tenant_id = $1 AND id = $2`,
		tenantID, cardID,
	)
	return err
}

// ===================== SUPPLIES =====================

// ListSupplies returns inventory for a tenant.
func (r *HousekeepingRepository) ListSupplies(ctx context.Context, tenantID, category string) ([]domain.SuppliesInventoryItem, error) {
	query := `SELECT id, tenant_id, item_name, category, unit, current_stock, reorder_level, cost_per_unit, supplier, last_restocked, created_at
		FROM supplies_inventory WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	if category != "" {
		query += " AND category = $2"
		args = append(args, category)
	}
	query += " ORDER BY item_name ASC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.SuppliesInventoryItem])
}

// UpdateSupplyStock adjusts inventory quantity.
func (r *HousekeepingRepository) UpdateSupplyStock(ctx context.Context, tenantID, itemID string, delta int) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE supplies_inventory SET current_stock = current_stock + $1, last_restocked = NOW(), updated_at = NOW()
		WHERE tenant_id = $2 AND id = $3`,
		delta, tenantID, itemID,
	)
	return err
}

// ===================== DASHBOARD STATS =====================

// GetDashboardStats returns aggregated housekeeping metrics.
func (r *HousekeepingRepository) GetDashboardStats(ctx context.Context, tenantID string) (*domain.HousekeepingDashboardStats, error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE status != 'cancelled') as total_rooms,
			COUNT(*) FILTER (WHERE status = 'pending') as dirty_rooms,
			COUNT(*) FILTER (WHERE status = 'in_progress') as in_progress_rooms,
			COUNT(*) FILTER (WHERE status = 'completed') as ready_rooms,
			COUNT(*) FILTER (WHERE status = 'inspected') as inspected_rooms,
			COUNT(*) FILTER (WHERE status = 'blocked') as blocked_rooms,
			COUNT(*) FILTER (WHERE status = 'maintenance') as maintenance_rooms,
			COUNT(*) FILTER (WHERE status = 'pending' AND assigned_to IS NULL) as unassigned_tasks,
			COALESCE(AVG(actual_minutes) FILTER (WHERE status = 'completed' AND actual_minutes > 0), 0) as avg_clean_time
		FROM housekeeping_tasks
		WHERE tenant_id = $1 AND created_at >= CURRENT_DATE`

	var stats domain.HousekeepingDashboardStats
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(
		&stats.TotalRooms, &stats.DirtyRooms, &stats.InProgressRooms, &stats.ReadyRooms,
		&stats.InspectedRooms, &stats.BlockedRooms, &stats.MaintenanceRooms,
		&stats.UnassignedTasks, &stats.AvgCleanTime,
	)
	if err != nil {
		return nil, err
	}

	// Staff on duty
	_ = r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM housekeeping_staff WHERE tenant_id = $1 AND active = true`,
		tenantID,
	).Scan(&stats.StaffOnDuty)

	// Staff load
	loadQuery := `
		SELECT s.id, s.name,
			COUNT(t.id) FILTER (WHERE t.status IN ('pending', 'in_progress')) as assigned,
			COUNT(t.id) FILTER (WHERE t.status = 'in_progress') as in_progress,
			COUNT(t.id) FILTER (WHERE t.status = 'completed') as completed,
			COALESCE(SUM(t.estimated_minutes) FILTER (WHERE t.status IN ('pending', 'in_progress')), 0) as est_min
		FROM housekeeping_staff s
		LEFT JOIN housekeeping_tasks t ON t.assigned_to = s.id AND t.tenant_id = s.tenant_id AND t.created_at >= CURRENT_DATE
		WHERE s.tenant_id = $1 AND s.active = true
		GROUP BY s.id, s.name
		ORDER BY assigned DESC`

	loadRows, err := r.pool.Query(ctx, loadQuery, tenantID)
	if err == nil {
		defer loadRows.Close()
		for loadRows.Next() {
			var item domain.StaffLoadItem
			_ = loadRows.Scan(&item.StaffID, &item.StaffName, &item.Assigned, &item.InProgress, &item.Completed, &item.EstimatedMin)
			stats.StaffLoad = append(stats.StaffLoad, item)
		}
	}

	// Floor progress
	floorQuery := `
		SELECT floor,
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'pending') as dirty,
			COUNT(*) FILTER (WHERE status = 'in_progress') as in_progress,
			COUNT(*) FILTER (WHERE status = 'completed') as ready
		FROM housekeeping_tasks
		WHERE tenant_id = $1 AND created_at >= CURRENT_DATE
		GROUP BY floor
		ORDER BY floor`

	floorRows, err := r.pool.Query(ctx, floorQuery, tenantID)
	if err == nil {
		defer floorRows.Close()
		for floorRows.Next() {
			var item domain.FloorProgressItem
			_ = floorRows.Scan(&item.Floor, &item.Total, &item.Dirty, &item.InProgress, &item.Ready)
			if item.Total > 0 {
				item.PercentDone = ((item.Ready + item.InProgress) * 100) / item.Total
			}
			stats.FloorProgress = append(stats.FloorProgress, item)
		}
	}

	return &stats, nil
}

// GetStaffPerformance returns performance for a date range.
func (r *HousekeepingRepository) GetStaffPerformance(ctx context.Context, tenantID, staffID string, from, to time.Time) ([]domain.StaffPerformance, error) {
	query := `SELECT staff_id, staff_name, performance_date, tasks_completed, avg_minutes_per_room, quality_score, complaints, rooms_cleaned, rooms_inspected
		FROM housekeeping_staff_performance
		WHERE tenant_id = $1 AND staff_id = $2 AND performance_date BETWEEN $3 AND $4
		ORDER BY performance_date DESC`
	rows, err := r.pool.Query(ctx, query, tenantID, staffID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, pgx.RowToStructByName[domain.StaffPerformance])
}
