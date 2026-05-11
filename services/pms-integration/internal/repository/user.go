package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// UserRepository persists and retrieves users.
type UserRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, tenantID, email string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error
	UpdateLastLogin(ctx context.Context, tenantID, id string) error
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.User, error)
	Delete(ctx context.Context, tenantID, id string) error
}

// PostgresUserRepository is the PostgreSQL implementation.
type PostgresUserRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewUserRepository creates a new user repository.
func NewUserRepository(pool *db.Pool, metrics *RepositoryMetrics) UserRepository {
	return &PostgresUserRepository{pool: pool, metrics: metrics}
}

func (r *PostgresUserRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	var u domain.User
	var lastLogin *time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, email, password_hash, first_name, last_name, role, is_active, last_login_at, created_at, updated_at, version
		FROM users WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID).Scan(
		&u.ID, &u.TenantID, &u.PropertyID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &lastLogin, &u.CreatedAt, &u.UpdatedAt, &u.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	u.LastLoginAt = lastLogin
	return &u, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, tenantID, email string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	var u domain.User
	var lastLogin *time.Time
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, email, password_hash, first_name, last_name, role, is_active, last_login_at, created_at, updated_at, version
		FROM users WHERE tenant_id = $1 AND email = $2 AND deleted_at IS NULL
	`, tenantID, email).Scan(
		&u.ID, &u.TenantID, &u.PropertyID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &lastLogin, &u.CreatedAt, &u.UpdatedAt, &u.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	u.LastLoginAt = lastLogin
	return &u, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, u.TenantID); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, tenant_id, property_id, email, password_hash, first_name, last_name, role, is_active, created_at, updated_at, version)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
	`, u.ID, u.TenantID, u.PropertyID, u.Email, u.PasswordHash, u.FirstName, u.LastName, u.Role, u.IsActive, u.CreatedAt, u.UpdatedAt, u.Version)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) Update(ctx context.Context, u *domain.User) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, u.TenantID); err != nil {
		return err
	}
	u.UpdatedAt = time.Now().UTC()
	cmdTag, err := r.pool.Exec(ctx, `
		UPDATE users SET property_id=$1, email=$2, first_name=$3, last_name=$4, role=$5, is_active=$6, updated_at=$7, version=version+1
		WHERE id=$8 AND tenant_id=$9 AND version=$10 AND deleted_at IS NULL
	`, u.PropertyID, u.Email, u.FirstName, u.LastName, u.Role, u.IsActive, u.UpdatedAt, u.ID, u.TenantID, u.Version)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("user update conflict: id=%s version=%d", u.ID, u.Version)
	}
	u.Version++
	return nil
}

func (r *PostgresUserRepository) UpdateLastLogin(ctx context.Context, tenantID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE users SET last_login_at = NOW(), updated_at = NOW(), version = version + 1
		WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL
	`, id, tenantID)
	if err != nil {
		return fmt.Errorf("update last login: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 100
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, property_id, email, password_hash, first_name, last_name, role, is_active, last_login_at, created_at, updated_at, version
		FROM users WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var u domain.User
		var lastLogin *time.Time
		if err := rows.Scan(&u.ID, &u.TenantID, &u.PropertyID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Role, &u.IsActive, &lastLogin, &u.CreatedAt, &u.UpdatedAt, &u.Version); err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		u.LastLoginAt = lastLogin
		users = append(users, &u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *PostgresUserRepository) Delete(ctx context.Context, tenantID, id string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return err
	}
	_, err := r.pool.Exec(ctx, `UPDATE users SET deleted_at = NOW() WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL`, id, tenantID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
