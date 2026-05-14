package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// AuthRepository handles authentication queries.
type AuthRepository struct{ pool *db.Pool }

func NewAuthRepository(pool *db.Pool) *AuthRepository { return &AuthRepository{pool: pool} }

// UserCredentials holds login info for a user.
type UserCredentials struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	Email        string
	PasswordHash string
	Name         string
	Role         string
	IsActive     bool
}

// GetUserByEmail looks up a user by email across all tenants (bypasses RLS).
func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*UserCredentials, error) {
	var u UserCredentials
	// Use a direct query that doesn't rely on RLS tenant setting
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, password_hash, name, role, is_active
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`, email).Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive)
	if err != nil {
		return nil, fmt.Errorf("user lookup: %w", err)
	}
	return &u, nil
}

// GetUserByID looks up a user by ID (bypasses RLS).
func (r *AuthRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*UserCredentials, error) {
	var u UserCredentials
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, password_hash, name, role, is_active
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, userID).Scan(&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.Name, &u.Role, &u.IsActive)
	if err != nil {
		return nil, fmt.Errorf("user lookup by id: %w", err)
	}
	return &u, nil
}

// UpdateLastLogin sets the last_login_at timestamp.
func (r *AuthRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET last_login_at = NOW() WHERE id = $1`, userID)
	return err
}

// GetTenantByID fetches tenant info for login response.
func (r *AuthRepository) GetTenantByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	var t domain.Tenant
	err := r.pool.QueryRow(ctx, `
		SELECT id, external_id, name, region, tier, config, created_at, updated_at
		FROM tenants WHERE id = $1
	`, tenantID).Scan(&t.ID, &t.ExternalID, &t.Name, &t.Region, &t.Tier, &t.Config, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
