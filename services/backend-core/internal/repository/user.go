package repository

import (
	"context"
	"encoding/json"

	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

type UserRepository struct {
	pool *db.Pool
}

func NewUserRepository(pool *db.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) List(ctx context.Context, tenantID string) ([]domain.User, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, email, name, phone, role, status, last_login, permissions, created_at, updated_at
		FROM users
		WHERE tenant_id = $1
		ORDER BY name
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		var perms []byte
		if err := rows.Scan(&u.ID, &u.TenantID, &u.Email, &u.Name, &u.Phone, &u.Role, &u.Status, &u.LastLogin, &perms, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		u.Permissions = mustUnmarshalStringSlice(perms)
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *UserRepository) Get(ctx context.Context, tenantID, id string) (*domain.User, error) {
	var u domain.User
	var perms []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, name, phone, role, status, last_login, permissions, created_at, updated_at
		FROM users WHERE tenant_id = $1 AND id = $2
	`, tenantID, id).Scan(&u.ID, &u.TenantID, &u.Email, &u.Name, &u.Phone, &u.Role, &u.Status, &u.LastLogin, &perms, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	u.Permissions = mustUnmarshalStringSlice(perms)
	return &u, nil
}

func (r *UserRepository) Create(ctx context.Context, tenantID string, u *domain.User) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO users (tenant_id, email, name, phone, role, status, permissions)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`, tenantID, u.Email, u.Name, u.Phone, u.Role, u.Status, mustMarshalStringSlice(u.Permissions)).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepository) Update(ctx context.Context, tenantID, id string, u *domain.User) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET email=$1, name=$2, phone=$3, role=$4, status=$5, permissions=$6
		WHERE tenant_id=$7 AND id=$8
	`, u.Email, u.Name, u.Phone, u.Role, u.Status, mustMarshalStringSlice(u.Permissions), tenantID, id)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, tenantID, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM users WHERE tenant_id=$1 AND id=$2`, tenantID, id)
	return err
}

func mustMarshalStringSlice(v []string) []byte {
	if v == nil {
		return []byte("[]")
	}
	b, _ := json.Marshal(v)
	return b
}

func mustUnmarshalStringSlice(data []byte) []string {
	var v []string
	_ = json.Unmarshal(data, &v)
	return v
}
