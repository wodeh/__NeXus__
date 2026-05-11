package repository

import (
	"context"

	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

type RoleRepository struct {
	pool *db.Pool
}

func NewRoleRepository(pool *db.Pool) *RoleRepository {
	return &RoleRepository{pool: pool}
}

func (r *RoleRepository) List(ctx context.Context, tenantID string) ([]domain.Role, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, permissions, description, is_system, created_at, updated_at
		FROM roles
		WHERE tenant_id = $1
		ORDER BY name
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []domain.Role
	for rows.Next() {
		var rl domain.Role
		var perms []byte
		if err := rows.Scan(&rl.ID, &rl.TenantID, &rl.Name, &perms, &rl.Description, &rl.IsSystem, &rl.CreatedAt, &rl.UpdatedAt); err != nil {
			return nil, err
		}
		rl.Permissions = mustUnmarshalStringSlice(perms)
		roles = append(roles, rl)
	}
	return roles, rows.Err()
}

func (r *RoleRepository) Create(ctx context.Context, tenantID string, rl *domain.Role) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO roles (tenant_id, name, permissions, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, is_system, created_at, updated_at
	`, tenantID, rl.Name, mustMarshalStringSlice(rl.Permissions), rl.Description).Scan(&rl.ID, &rl.IsSystem, &rl.CreatedAt, &rl.UpdatedAt)
}

func (r *RoleRepository) Update(ctx context.Context, tenantID, id string, rl *domain.Role) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE roles SET name=$1, permissions=$2, description=$3
		WHERE tenant_id=$4 AND id=$5 AND is_system = FALSE
	`, rl.Name, mustMarshalStringSlice(rl.Permissions), rl.Description, tenantID, id)
	return err
}

func (r *RoleRepository) Delete(ctx context.Context, tenantID, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM roles WHERE tenant_id=$1 AND id=$2 AND is_system = FALSE`, tenantID, id)
	return err
}
