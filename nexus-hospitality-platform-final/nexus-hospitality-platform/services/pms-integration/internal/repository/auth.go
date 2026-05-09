// Package repository provides tenant-aware data access for authentication.
package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// UserRepository persists and retrieves User aggregates.
type UserRepository interface {
	GetByID(ctx context.Context, tenantID, id string) (*domain.User, error)
	GetByEmail(ctx context.Context, tenantID, email string) (*domain.User, error)
	Create(ctx context.Context, u *domain.User) error
	Update(ctx context.Context, u *domain.User) error
	ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.User, error)
}

// PostgresUserRepository is the PostgreSQL implementation.
type PostgresUserRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewUserRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool, metrics: metrics}
}

func (r *PostgresUserRepository) withTenant(ctx context.Context, tenantID string) (context.Context, error) {
	if err := r.pool.SetTenant(ctx, tenantID); err != nil {
		return ctx, err
	}
	return ctx, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.User, error) {
	start := time.Now()
	r.metrics.IncQuery("user", "get_by_id")
	defer r.metrics.ObserveDuration("user", "get_by_id", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, password_hash, first_name, last_name, phone,
		       role, status, email_verified, last_login_at, password_changed_at,
		       failed_login_attempts, locked_until, created_at, updated_at, version
		FROM users
		WHERE id = $1 AND tenant_id = $2
	`, id, tenantID)

	var u domain.User
	var lastLogin, lockedUntil *time.Time
	err := row.Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone,
		&u.Role, &u.Status, &u.EmailVerified, &lastLogin, &u.PasswordChangedAt,
		&u.FailedLoginAttempts, &lockedUntil, &u.CreatedAt, &u.UpdatedAt, &u.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	u.LastLoginAt = lastLogin
	u.LockedUntil = lockedUntil
	return &u, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, tenantID, email string) (*domain.User, error) {
	start := time.Now()
	r.metrics.IncQuery("user", "get_by_email")
	defer r.metrics.ObserveDuration("user", "get_by_email", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, password_hash, first_name, last_name, phone,
		       role, status, email_verified, last_login_at, password_changed_at,
		       failed_login_attempts, locked_until, created_at, updated_at, version
		FROM users
		WHERE email = $1 AND tenant_id = $2
	`, email, tenantID)

	var u domain.User
	var lastLogin, lockedUntil *time.Time
	err := row.Scan(
		&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone,
		&u.Role, &u.Status, &u.EmailVerified, &lastLogin, &u.PasswordChangedAt,
		&u.FailedLoginAttempts, &lockedUntil, &u.CreatedAt, &u.UpdatedAt, &u.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	u.LastLoginAt = lastLogin
	u.LockedUntil = lockedUntil
	return &u, nil
}

func (r *PostgresUserRepository) Create(ctx context.Context, u *domain.User) error {
	start := time.Now()
	r.metrics.IncQuery("user", "create")
	defer r.metrics.ObserveDuration("user", "create", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, u.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, phone,
		                   role, status, email_verified, last_login_at, password_changed_at,
		                   failed_login_attempts, locked_until, version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	`, u.ID, u.TenantID, u.Email, u.PasswordHash, u.FirstName, u.LastName, u.Phone,
		u.Role, u.Status, u.EmailVerified, u.LastLoginAt, u.PasswordChangedAt,
		u.FailedLoginAttempts, u.LockedUntil, u.Version, u.CreatedAt, u.UpdatedAt)
	return err
}

func (r *PostgresUserRepository) Update(ctx context.Context, u *domain.User) error {
	start := time.Now()
	r.metrics.IncQuery("user", "update")
	defer r.metrics.ObserveDuration("user", "update", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, u.TenantID); err != nil {
		return err
	}

	_, err := r.pool.Exec(ctx, `
		UPDATE users
		SET email = $1, password_hash = $2, first_name = $3, last_name = $4, phone = $5,
		    role = $6, status = $7, email_verified = $8, last_login_at = $9,
		    password_changed_at = $10, failed_login_attempts = $11, locked_until = $12,
		    version = $13, updated_at = $14
		WHERE id = $15 AND tenant_id = $16 AND version = $17
	`, u.Email, u.PasswordHash, u.FirstName, u.LastName, u.Phone,
		u.Role, u.Status, u.EmailVerified, u.LastLoginAt,
		u.PasswordChangedAt, u.FailedLoginAttempts, u.LockedUntil,
		u.Version, u.UpdatedAt,
		u.ID, u.TenantID, u.Version-1)
	return err
}

func (r *PostgresUserRepository) ListByTenant(ctx context.Context, tenantID string, limit, offset int) ([]*domain.User, error) {
	start := time.Now()
	r.metrics.IncQuery("user", "list")
	defer r.metrics.ObserveDuration("user", "list", time.Since(start).Seconds())

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := r.withTenant(ctx, tenantID); err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, email, password_hash, first_name, last_name, phone,
		       role, status, email_verified, last_login_at, password_changed_at,
		       failed_login_attempts, locked_until, created_at, updated_at, version
		FROM users
		WHERE tenant_id = $1
		ORDER BY created_at DESC LIMIT $2 OFFSET $3
	`, tenantID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanUsers(rows)
}

func scanUsers(rows pgx.Rows) ([]*domain.User, error) {
	var users []*domain.User
	for rows.Next() {
		var u domain.User
		var lastLogin, lockedUntil *time.Time
		err := rows.Scan(
			&u.ID, &u.TenantID, &u.Email, &u.PasswordHash, &u.FirstName, &u.LastName, &u.Phone,
			&u.Role, &u.Status, &u.EmailVerified, &lastLogin, &u.PasswordChangedAt,
			&u.FailedLoginAttempts, &lockedUntil, &u.CreatedAt, &u.UpdatedAt, &u.Version,
		)
		if err != nil {
			return nil, err
		}
		u.LastLoginAt = lastLogin
		u.LockedUntil = lockedUntil
		users = append(users, &u)
	}
	return users, rows.Err()
}

// ==================== TENANT REPOSITORY ====================

type TenantRepository interface {
	GetByID(ctx context.Context, id string) (*domain.Tenant, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error)
	Create(ctx context.Context, t *domain.Tenant) error
	Update(ctx context.Context, t *domain.Tenant) error
	ListAll(ctx context.Context, limit, offset int) ([]*domain.Tenant, error)
}

type PostgresTenantRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewTenantRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresTenantRepository {
	return &PostgresTenantRepository{pool: pool, metrics: metrics}
}

func (r *PostgresTenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	start := time.Now()
	r.metrics.IncQuery("tenant", "get_by_id")
	defer r.metrics.ObserveDuration("tenant", "get_by_id", time.Since(start).Seconds())

	row := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, email, phone, address, country, currency_code, timezone,
		       status, plan, max_properties, max_rooms, max_users,
		       billing_address, tax_id, stripe_customer_id, trial_ends_at,
		       created_at, updated_at, version
		FROM tenants
		WHERE id = $1
	`, id)

	var t domain.Tenant
	var trialEnds *time.Time
	err := row.Scan(
		&t.ID, &t.Slug, &t.Name, &t.Email, &t.Phone, &t.Address, &t.Country, &t.CurrencyCode, &t.Timezone,
		&t.Status, &t.Plan, &t.MaxProperties, &t.MaxRooms, &t.MaxUsers,
		&t.BillingAddress, &t.TaxID, &t.StripeCustomerID, &trialEnds,
		&t.CreatedAt, &t.UpdatedAt, &t.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	t.TrialEndsAt = trialEnds
	return &t, nil
}

func (r *PostgresTenantRepository) GetBySlug(ctx context.Context, slug string) (*domain.Tenant, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, slug, name, email, phone, address, country, currency_code, timezone,
		       status, plan, max_properties, max_rooms, max_users,
		       billing_address, tax_id, stripe_customer_id, trial_ends_at,
		       created_at, updated_at, version
		FROM tenants
		WHERE slug = $1
	`, slug)

	var t domain.Tenant
	var trialEnds *time.Time
	err := row.Scan(
		&t.ID, &t.Slug, &t.Name, &t.Email, &t.Phone, &t.Address, &t.Country, &t.CurrencyCode, &t.Timezone,
		&t.Status, &t.Plan, &t.MaxProperties, &t.MaxRooms, &t.MaxUsers,
		&t.BillingAddress, &t.TaxID, &t.StripeCustomerID, &trialEnds,
		&t.CreatedAt, &t.UpdatedAt, &t.Version,
	)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	t.TrialEndsAt = trialEnds
	return &t, nil
}

func (r *PostgresTenantRepository) Create(ctx context.Context, t *domain.Tenant) error {
	start := time.Now()
	r.metrics.IncQuery("tenant", "create")
	defer r.metrics.ObserveDuration("tenant", "create", time.Since(start).Seconds())

	_, err := r.pool.Exec(ctx, `
		INSERT INTO tenants (id, slug, name, email, phone, address, country, currency_code, timezone,
		                   status, plan, max_properties, max_rooms, max_users,
		                   billing_address, tax_id, stripe_customer_id, trial_ends_at,
		                   version, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
	`, t.ID, t.Slug, t.Name, t.Email, t.Phone, t.Address, t.Country, t.CurrencyCode, t.Timezone,
		t.Status, t.Plan, t.MaxProperties, t.MaxRooms, t.MaxUsers,
		t.BillingAddress, t.TaxID, t.StripeCustomerID, t.TrialEndsAt,
		t.Version, t.CreatedAt, t.UpdatedAt)
	return err
}

func (r *PostgresTenantRepository) Update(ctx context.Context, t *domain.Tenant) error {
	start := time.Now()
	r.metrics.IncQuery("tenant", "update")
	defer r.metrics.ObserveDuration("tenant", "update", time.Since(start).Seconds())

	_, err := r.pool.Exec(ctx, `
		UPDATE tenants
		SET slug = $1, name = $2, email = $3, phone = $4, address = $5, country = $6,
		    currency_code = $7, timezone = $8, status = $9, plan = $10,
		    max_properties = $11, max_rooms = $12, max_users = $13,
		    billing_address = $14, tax_id = $15, stripe_customer_id = $16,
		    trial_ends_at = $17, version = $18, updated_at = $19
		WHERE id = $20 AND version = $21
	`, t.Slug, t.Name, t.Email, t.Phone, t.Address, t.Country,
		t.CurrencyCode, t.Timezone, t.Status, t.Plan,
		t.MaxProperties, t.MaxRooms, t.MaxUsers,
		t.BillingAddress, t.TaxID, t.StripeCustomerID,
		t.TrialEndsAt, t.Version, t.UpdatedAt,
		t.ID, t.Version-1)
	return err
}

func (r *PostgresTenantRepository) ListAll(ctx context.Context, limit, offset int) ([]*domain.Tenant, error) {
	start := time.Now()
	r.metrics.IncQuery("tenant", "list")
	defer r.metrics.ObserveDuration("tenant", "list", time.Since(start).Seconds())

	rows, err := r.pool.Query(ctx, `
		SELECT id, slug, name, email, phone, address, country, currency_code, timezone,
		       status, plan, max_properties, max_rooms, max_users,
		       billing_address, tax_id, stripe_customer_id, trial_ends_at,
		       created_at, updated_at, version
		FROM tenants
		ORDER BY created_at DESC LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []*domain.Tenant
	for rows.Next() {
		var t domain.Tenant
		var trialEnds *time.Time
		err := rows.Scan(
			&t.ID, &t.Slug, &t.Name, &t.Email, &t.Phone, &t.Address, &t.Country, &t.CurrencyCode, &t.Timezone,
			&t.Status, &t.Plan, &t.MaxProperties, &t.MaxRooms, &t.MaxUsers,
			&t.BillingAddress, &t.TaxID, &t.StripeCustomerID, &trialEnds,
			&t.CreatedAt, &t.UpdatedAt, &t.Version,
		)
		if err != nil {
			return nil, err
		}
		t.TrialEndsAt = trialEnds
		tenants = append(tenants, &t)
	}
	return tenants, rows.Err()
}

// ==================== REFRESH TOKEN REPOSITORY ====================

type RefreshTokenRepository interface {
	Create(ctx context.Context, rt *domain.RefreshToken) error
	GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
	RevokeAllForUser(ctx context.Context, userID string) error
}

type PostgresRefreshTokenRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewRefreshTokenRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresRefreshTokenRepository {
	return &PostgresRefreshTokenRepository{pool: pool, metrics: metrics}
}

func (r *PostgresRefreshTokenRepository) Create(ctx context.Context, rt *domain.RefreshToken) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO refresh_tokens (id, user_id, tenant_id, token_hash, expires_at,
		                            created_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, rt.ID, rt.UserID, rt.TenantID, rt.TokenHash, rt.ExpiresAt,
		rt.CreatedAt, rt.IPAddress, rt.UserAgent)
	return err
}

func (r *PostgresRefreshTokenRepository) GetByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	// This is a simplified lookup; in production, you'd hash and compare
	row := r.pool.QueryRow(ctx, `
		SELECT id, user_id, tenant_id, token_hash, expires_at, created_at,
		       revoked_at, ip_address, user_agent
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()
	`, token)

	var rt domain.RefreshToken
	var revokedAt *time.Time
	err := row.Scan(
		&rt.ID, &rt.UserID, &rt.TenantID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt,
		&revokedAt, &rt.IPAddress, &rt.UserAgent,
	)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found: %w", err)
	}
	rt.RevokedAt = revokedAt
	return &rt, nil
}

func (r *PostgresRefreshTokenRepository) Revoke(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1
	`, id)
	return err
}

func (r *PostgresRefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	return err
}
