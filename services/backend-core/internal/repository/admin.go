package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/backend-core/internal/db"
	"nexus-hospitality-platform/services/backend-core/internal/domain"
)

// AdminRepository handles admin CRUD for users, properties, and system config.
type AdminRepository struct{ pool *db.Pool }

func NewAdminRepository(pool *db.Pool) *AdminRepository { return &AdminRepository{pool: pool} }

/* ─── Properties ─── */

func (r *AdminRepository) ListProperties(ctx context.Context, tenantID uuid.UUID) ([]domain.Property, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, address, city, country, phone, email, timezone, currency, star_rating, is_active, config, created_at, updated_at
		FROM properties WHERE tenant_id=$1 AND deleted_at IS NULL`, tenantID)
	if err != nil { return nil, fmt.Errorf("list properties: %w", err) }
	defer rows.Close()
	var out []domain.Property
	for rows.Next() {
		var p domain.Property
		var configJSON []byte
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.Address, &p.City, &p.Country, &p.Phone, &p.Email, &p.Timezone, &p.Currency, &p.StarRating, &p.IsActive, &configJSON, &p.CreatedAt, &p.UpdatedAt); err != nil { continue }
		_ = unmarshalJSON(configJSON, &p.Config)
		out = append(out, p)
	}
	return out, nil
}

func (r *AdminRepository) GetProperty(ctx context.Context, tenantID, id uuid.UUID) (*domain.Property, error) {
	var p domain.Property
	var configJSON []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, address, city, country, phone, email, timezone, currency, star_rating, is_active, config, created_at, updated_at
		FROM properties WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`, tenantID, id).
		Scan(&p.ID, &p.TenantID, &p.Name, &p.Address, &p.City, &p.Country, &p.Phone, &p.Email, &p.Timezone, &p.Currency, &p.StarRating, &p.IsActive, &configJSON, &p.CreatedAt, &p.UpdatedAt)
	if err != nil { return nil, fmt.Errorf("get property: %w", err) }
	_ = unmarshalJSON(configJSON, &p.Config)
	return &p, nil
}

func (r *AdminRepository) CreateProperty(ctx context.Context, p *domain.Property) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO properties (id, tenant_id, name, address, city, country, phone, email, timezone, currency, star_rating, is_active, config)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (tenant_id, name) DO UPDATE SET
		address=$4, city=$5, country=$6, phone=$7, email=$8, timezone=$9, currency=$10, star_rating=$11, is_active=$12, config=$13, updated_at=NOW()`,
		p.ID, p.TenantID, p.Name, p.Address, p.City, p.Country, p.Phone, p.Email, p.Timezone, p.Currency, p.StarRating, p.IsActive, marshalJSON(p.Config))
	if err != nil { return fmt.Errorf("create property: %w", err) }
	return nil
}

/* ─── Users ─── */

func (r *AdminRepository) ListUsers(ctx context.Context, tenantID uuid.UUID) ([]domain.UserWithRole, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, email, name, role, is_active, last_login_at, created_at
		FROM users WHERE tenant_id=$1 AND deleted_at IS NULL`, tenantID)
	if err != nil { return nil, fmt.Errorf("list users: %w", err) }
	defer rows.Close()
	var out []domain.UserWithRole
	for rows.Next() {
		var u domain.UserWithRole
		var lastLogin pgx.NullTime
		if err := rows.Scan(&u.ID, &u.TenantID, &u.Email, &u.Name, &u.Role, &u.IsActive, &lastLogin, &u.CreatedAt); err != nil { continue }
		if lastLogin.Valid { t := lastLogin.Time; u.LastLoginAt = &t }
		out = append(out, u)
	}
	return out, nil
}

func (r *AdminRepository) GetUser(ctx context.Context, tenantID, id uuid.UUID) (*domain.UserWithRole, error) {
	var u domain.UserWithRole
	var lastLogin pgx.NullTime
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, email, name, role, is_active, last_login_at, created_at
		FROM users WHERE tenant_id=$1 AND id=$2 AND deleted_at IS NULL`, tenantID, id).
		Scan(&u.ID, &u.TenantID, &u.Email, &u.Name, &u.Role, &u.IsActive, &lastLogin, &u.CreatedAt)
	if err != nil { return nil, fmt.Errorf("get user: %w", err) }
	if lastLogin.Valid { t := lastLogin.Time; u.LastLoginAt = &t }
	return &u, nil
}

func (r *AdminRepository) CreateUser(ctx context.Context, u *domain.UserWithRole, passwordHash string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO users (id, tenant_id, email, password_hash, name, role, is_active)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (tenant_id, email) DO UPDATE SET
		name=$5, role=$6, is_active=$7, updated_at=NOW()`,
		u.ID, u.TenantID, u.Email, passwordHash, u.Name, u.Role, u.IsActive)
	if err != nil { return fmt.Errorf("create user: %w", err) }
	return nil
}

func (r *AdminRepository) UpdateUser(ctx context.Context, tenantID, id uuid.UUID, name, role string, isActive bool) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET name=$1, role=$2, is_active=$3, updated_at=NOW() WHERE tenant_id=$4 AND id=$5`, name, role, isActive, tenantID, id)
	if err != nil { return fmt.Errorf("update user: %w", err) }
	return nil
}

func (r *AdminRepository) DeleteUser(ctx context.Context, tenantID, id uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `UPDATE users SET deleted_at=NOW() WHERE tenant_id=$1 AND id=$2`, tenantID, id)
	if err != nil { return fmt.Errorf("delete user: %w", err) }
	return nil
}

/* ─── System Config ─── */

func (r *AdminRepository) GetSystemConfig(ctx context.Context, tenantID uuid.UUID) (*domain.SystemConfig, error) {
	var c domain.SystemConfig
	var jsonData []byte
	row := r.pool.QueryRow(ctx, `SELECT config FROM tenant_properties WHERE tenant_id=$1`, tenantID)
	if err := row.Scan(&jsonData); err != nil {
		return &domain.SystemConfig{TenantID: tenantID, DefaultCheckInTime: "15:00", DefaultCheckOutTime: "11:00", AutoConfirm: true, DepositPercent: 20, AllowWalkIn: true}, nil
	}
	var raw map[string]interface{}
	_ = unmarshalJSON(jsonData, &raw)
	c.TenantID = tenantID
	if v, ok := raw["default_check_in_time"].(string); ok { c.DefaultCheckInTime = v }
	if v, ok := raw["default_check_out_time"].(string); ok { c.DefaultCheckOutTime = v }
	if v, ok := raw["auto_confirm"].(bool); ok { c.AutoConfirm = v }
	if v, ok := raw["require_deposit"].(bool); ok { c.RequireDeposit = v }
	if v, ok := raw["deposit_percent"].(float64); ok { c.DepositPercent = int(v) }
	if v, ok := raw["allow_walk_in"].(bool); ok { c.AllowWalkIn = v }
	return &c, nil
}

func (r *AdminRepository) UpdateSystemConfig(ctx context.Context, tenantID uuid.UUID, config map[string]interface{}) error {
	_, err := r.pool.Exec(ctx, `UPDATE tenant_properties SET config=$1, updated_at=NOW() WHERE tenant_id=$2`, marshalJSON(config), tenantID)
	if err != nil { return fmt.Errorf("update config: %w", err) }
	return nil
}
