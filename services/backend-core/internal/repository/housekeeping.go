package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// HousekeepingRepository provides housekeeping data access.
type HousekeepingRepository struct {
	pool *db.Pool
}

// NewHousekeepingRepository creates a housekeeping repository.
func NewHousekeepingRepository(pool *db.Pool) *HousekeepingRepository {
	return &HousekeepingRepository{pool: pool}
}

func (r *HousekeepingRepository) execer(tx pgx.Tx) interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error)
} {
	if tx != nil {
		return tx
	}
	return r.pool
}

// ListStaff returns all active housekeeping staff.
func (r *HousekeepingRepository) ListStaff(ctx context.Context, tenantID string) ([]domain.HousekeepingStaff, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, role, active_shift, max_rooms_per_day, phone, email, active,
			created_at, updated_at, deleted_at, version
		FROM housekeeping_staff
		WHERE tenant_id = $1 AND deleted_at IS NULL AND active = TRUE
		ORDER BY name
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list staff: %w", err)
	}
	defer rows.Close()

	var staff []domain.HousekeepingStaff
	for rows.Next() {
		var s domain.HousekeepingStaff
		var phone, email *string
		if err := rows.Scan(
			&s.ID, &s.TenantID, &s.Name, &s.Role, &s.ActiveShift, &s.MaxRoomsPerDay,
			&phone, &email, &s.Active,
			&s.CreatedAt, &s.UpdatedAt, &s.DeletedAt, &s.Version,
		); err != nil {
			return nil, fmt.Errorf("scan staff: %w", err)
		}
		s.Phone = phone
		s.Email = email
		staff = append(staff, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("staff rows: %w", err)
	}
	return staff, nil
}

// ListTasks returns all active housekeeping tasks.
func (r *HousekeepingRepository) ListTasks(ctx context.Context, tenantID string) ([]domain.HousekeepingTask, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, room_number, task_type, status, assigned_to, priority, notes,
			started_at, completed_at, created_at, updated_at, deleted_at, version
		FROM housekeeping_tasks
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY 
			CASE priority
				WHEN 'urgent' THEN 1
				WHEN 'high' THEN 2
				WHEN 'normal' THEN 3
				WHEN 'low' THEN 4
			END,
			created_at
	`, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []domain.HousekeepingTask
	for rows.Next() {
		var t domain.HousekeepingTask
		var assignedTo *uuid.UUID
		var notes *string
		var startedAt, completedAt *time.Time
		if err := rows.Scan(
			&t.ID, &t.TenantID, &t.RoomNumber, &t.TaskType, &t.Status, &assignedTo,
			&t.Priority, &notes, &startedAt, &completedAt,
			&t.CreatedAt, &t.UpdatedAt, &t.DeletedAt, &t.Version,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		t.AssignedTo = assignedTo
		t.Notes = notes
		t.StartedAt = startedAt
		t.CompletedAt = completedAt
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("task rows: %w", err)
	}
	return tasks, nil
}

// CreateTask inserts a new housekeeping task.
func (r *HousekeepingRepository) CreateTask(ctx context.Context, tenantID string, req *domain.HousekeepingTaskCreateRequest) (*domain.HousekeepingTask, error) {
	if req.RoomNumber == "" || req.TaskType == "" {
		return nil, fmt.Errorf("room_number and task_type required")
	}

	task := &domain.HousekeepingTask{
		ID:         uuid.Must(uuid.NewRandom()),
		TenantID:   uuid.MustParse(tenantID),
		RoomNumber: req.RoomNumber,
		TaskType:   req.TaskType,
		Status:     "pending",
		Priority:   req.Priority,
	}
	if req.Priority == "" {
		task.Priority = "normal"
	}
	if req.AssignedTo != nil && *req.AssignedTo != uuid.Nil {
		task.AssignedTo = req.AssignedTo
		task.Status = "in_progress"
		task.StartedAt = &task.CreatedAt
	}
	if req.Notes != "" {
		task.Notes = &req.Notes
	}

	execer := r.execer(nil)
	_, err := execer.Exec(ctx, `
		INSERT INTO housekeeping_tasks (
			id, tenant_id, room_number, task_type, status, assigned_to, priority, notes,
			started_at, completed_at, created_at, updated_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW(), 1)
	`, task.ID, task.TenantID, task.RoomNumber, task.TaskType, task.Status,
		task.AssignedTo, task.Priority, task.Notes,
		task.StartedAt, task.CompletedAt)
	if err != nil {
		return nil, fmt.Errorf("insert task: %w", err)
	}
	return task, nil
}

// UpdateTask updates a housekeeping task.
func (r *HousekeepingRepository) UpdateTask(ctx context.Context, tenantID string, id uuid.UUID, req *domain.HousekeepingTaskUpdateRequest) error {
	setClauses := []string{"updated_at = NOW()", "version = version + 1"}
	args := []interface{}{id, tenantID}
	argIdx := 3

	if req.Status != nil && *req.Status != "" {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *req.Status)
		argIdx++
		if *req.Status == "completed" {
			setClauses = append(setClauses, fmt.Sprintf("completed_at = $%d", argIdx))
			args = append(args, time.Now().UTC())
			argIdx++
		}
		if *req.Status == "in_progress" {
			setClauses = append(setClauses, fmt.Sprintf("started_at = $%d", argIdx))
			args = append(args, time.Now().UTC())
			argIdx++
		}
	}
	if req.AssignedTo != nil {
		setClauses = append(setClauses, fmt.Sprintf("assigned_to = $%d", argIdx))
		args = append(args, *req.AssignedTo)
		argIdx++
	}
	if req.Notes != nil {
		setClauses = append(setClauses, fmt.Sprintf("notes = $%d", argIdx))
		args = append(args, *req.Notes)
		argIdx++
	}

	query := fmt.Sprintf("UPDATE housekeeping_tasks SET %s WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL", 
		joinStrings(setClauses, ", "))
	
	cmdTag, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("task %s: %w", id, ErrNotFound)
	}
	return nil
}
