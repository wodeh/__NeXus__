package repository

import (
	"context"

	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// BookingEngineRepository handles direct booking data.
type BookingEngineRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewBookingEngineRepository creates a new repository.
func NewBookingEngineRepository(pool *db.Pool, metrics *RepositoryMetrics) *BookingEngineRepository {
	return &BookingEngineRepository{pool: pool, metrics: metrics}
}

// ========== PROMO CODES ==========

func (r *BookingEngineRepository) CreatePromoCode(ctx context.Context, tenantID string, p *domain.PromoCode) (*domain.PromoCode, error) {
	r.pool.SetTenant(ctx, tenantID)
	err := r.pool.QueryRow(ctx, `
		INSERT INTO promo_codes (tenant_id, code, discount_type, discount_value, max_uses, uses_count, min_nights, valid_from, valid_until, applicable_room_types, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, 0, $6, $7, $8, $9, $10, NOW(), NOW()) RETURNING id, created_at, updated_at`,
		tenantID, p.Code, p.DiscountType, p.DiscountValue, p.MaxUses, p.MinNights, p.ValidFrom, p.ValidUntil, p.ApplicableRoomTypes, p.IsActive).Scan(
		&p.ID, &p.CreatedAt, &p.UpdatedAt)
	p.TenantID = tenantID
	return p, err
}

func (r *BookingEngineRepository) ListPromoCodes(ctx context.Context, tenantID string) ([]*domain.PromoCode, error) {
	r.pool.SetTenant(ctx, tenantID)
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, code, discount_type, discount_value, max_uses, uses_count, min_nights, valid_from, valid_until, applicable_room_types, is_active, created_at, updated_at
		FROM promo_codes WHERE tenant_id = $1 ORDER BY code`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.PromoCode
	for rows.Next() {
		var p domain.PromoCode
		rows.Scan(
			&p.ID, &p.TenantID, &p.Code, &p.DiscountType, &p.DiscountValue, &p.MaxUses, &p.UsesCount, &p.MinNights, &p.ValidFrom, &p.ValidUntil, &p.ApplicableRoomTypes, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
		out = append(out, &p)
	}
	return out, nil
}

func (r *BookingEngineRepository) GetPromoCodeByCode(ctx context.Context, tenantID, code string) (*domain.PromoCode, error) {
	r.pool.SetTenant(ctx, tenantID)
	var p domain.PromoCode
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, code, discount_type, discount_value, max_uses, uses_count, min_nights, valid_from, valid_until, applicable_room_types, is_active
		FROM promo_codes WHERE tenant_id = $1 AND code = $2 AND is_active = true`, tenantID, code)
	err := row.Scan(
		&p.ID, &p.TenantID, &p.Code, &p.DiscountType, &p.DiscountValue, &p.MaxUses, &p.UsesCount, &p.MinNights, &p.ValidFrom, &p.ValidUntil, &p.ApplicableRoomTypes, &p.IsActive)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *BookingEngineRepository) IncrementPromoUse(ctx context.Context, tenantID, id string) error {
	r.pool.SetTenant(ctx, tenantID)
	_, err := r.pool.Exec(ctx, `UPDATE promo_codes SET uses_count = uses_count + 1 WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

func (r *BookingEngineRepository) DeletePromoCode(ctx context.Context, tenantID, id string) error {
	r.pool.SetTenant(ctx, tenantID)
	_, err := r.pool.Exec(ctx, `DELETE FROM promo_codes WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

// ========== WIDGET CONFIG ==========

func (r *BookingEngineRepository) GetWidgetConfig(ctx context.Context, tenantID string) (*domain.DirectBookingWidgetConfig, error) {
	r.pool.SetTenant(ctx, tenantID)
	var c domain.DirectBookingWidgetConfig
	row := r.pool.QueryRow(ctx, `
		SELECT tenant_id, enabled, property_id, theme_color, logo_url, title, subtitle, show_promo_code, show_extras, upsell_enabled, require_deposit, deposit_pct, success_redirect_url, created_at, updated_at
		FROM booking_widget_configs WHERE tenant_id = $1`, tenantID)
	err := row.Scan(
		&c.TenantID, &c.Enabled, &c.PropertyID, &c.ThemeColor, &c.LogoURL, &c.Title, &c.Subtitle, &c.ShowPromoCode, &c.ShowExtras, &c.UpsellEnabled, &c.RequireDeposit, &c.DepositPct, &c.SuccessRedirectURL, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *BookingEngineRepository) SaveWidgetConfig(ctx context.Context, tenantID string, c *domain.DirectBookingWidgetConfig) error {
	r.pool.SetTenant(ctx, tenantID)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO booking_widget_configs (tenant_id, enabled, property_id, theme_color, logo_url, title, subtitle, show_promo_code, show_extras, upsell_enabled, require_deposit, deposit_pct, success_redirect_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		ON CONFLICT (tenant_id) DO UPDATE SET
			enabled = EXCLUDED.enabled, property_id = EXCLUDED.property_id, theme_color = EXCLUDED.theme_color, logo_url = EXCLUDED.logo_url,
			title = EXCLUDED.title, subtitle = EXCLUDED.subtitle, show_promo_code = EXCLUDED.show_promo_code, show_extras = EXCLUDED.show_extras,
			upsell_enabled = EXCLUDED.upsell_enabled, require_deposit = EXCLUDED.require_deposit, deposit_pct = EXCLUDED.deposit_pct,
			success_redirect_url = EXCLUDED.success_redirect_url, updated_at = NOW()`,
		tenantID, c.Enabled, c.PropertyID, c.ThemeColor, c.LogoURL, c.Title, c.Subtitle, c.ShowPromoCode, c.ShowExtras, c.UpsellEnabled, c.RequireDeposit, c.DepositPct, c.SuccessRedirectURL)
	return err
}

// ========== UPSELLS ==========

func (r *BookingEngineRepository) CreateUpsell(ctx context.Context, tenantID string, u *domain.BookingUpsell) (*domain.BookingUpsell, error) {
	r.pool.SetTenant(ctx, tenantID)
	err := r.pool.QueryRow(ctx, `
		INSERT INTO booking_upsells (tenant_id, name, description, price, per_night, is_active, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW()) RETURNING id`,
		tenantID, u.Name, u.Description, u.Price, u.PerNight, u.IsActive).Scan(&u.ID)
	u.TenantID = tenantID
	return u, err
}

func (r *BookingEngineRepository) ListUpsells(ctx context.Context, tenantID string) ([]*domain.BookingUpsell, error) {
	r.pool.SetTenant(ctx, tenantID)
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, description, price, per_night, is_active
		FROM booking_upsells WHERE tenant_id = $1 ORDER BY name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.BookingUpsell
	for rows.Next() {
		var u domain.BookingUpsell
		rows.Scan(&u.ID, &u.TenantID, &u.Name, &u.Description, &u.Price, &u.PerNight, &u.IsActive)
		out = append(out, &u)
	}
	return out, nil
}

// ========== BOOKING SESSIONS ==========

func (r *BookingEngineRepository) CreateBookingSession(ctx context.Context, tenantID string, s *domain.DirectBookingSession) (*domain.DirectBookingSession, error) {
	r.pool.SetTenant(ctx, tenantID)
	query := `
		INSERT INTO direct_booking_sessions (tenant_id, property_id, widget_config_id, guest_name, guest_email, guest_phone, room_type_id, check_in, check_out, nights, adults, children, base_rate, upsells_json, promo_code_id, discount, total, deposit_amount, currency, status, expires_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, NOW() + INTERVAL '1 hour', NOW(), NOW())
		RETURNING id, expires_at, created_at, updated_at`
	err := r.pool.QueryRow(ctx, query,
		tenantID, s.PropertyID, s.WidgetConfigID, s.GuestName, s.GuestEmail, s.GuestPhone, s.RoomTypeID, s.CheckIn, s.CheckOut, s.Nights, s.Adults, s.Children, s.BaseRate, s.UpsellsJSON, s.PromoCodeID, s.Discount, s.Total, s.DepositAmount, s.Currency, s.Status).Scan(
		&s.ID, &s.ExpiresAt, &s.CreatedAt, &s.UpdatedAt)
	s.TenantID = tenantID
	return s, err
}

func (r *BookingEngineRepository) GetBookingSession(ctx context.Context, tenantID, id string) (*domain.DirectBookingSession, error) {
	r.pool.SetTenant(ctx, tenantID)
	var s domain.DirectBookingSession
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, property_id, widget_config_id, guest_name, guest_email, guest_phone, room_type_id, check_in, check_out, nights, adults, children, base_rate, upsells_json, promo_code_id, discount, total, deposit_amount, currency, status, expires_at, created_at, updated_at
		FROM direct_booking_sessions WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	err := row.Scan(
		&s.ID, &s.TenantID, &s.PropertyID, &s.WidgetConfigID, &s.GuestName, &s.GuestEmail, &s.GuestPhone, &s.RoomTypeID, &s.CheckIn, &s.CheckOut, &s.Nights, &s.Adults, &s.Children, &s.BaseRate, &s.UpsellsJSON, &s.PromoCodeID, &s.Discount, &s.Total, &s.DepositAmount, &s.Currency, &s.Status, &s.ExpiresAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *BookingEngineRepository) UpdateBookingSessionStatus(ctx context.Context, tenantID, id, status string) error {
	r.pool.SetTenant(ctx, tenantID)
	_, err := r.pool.Exec(ctx, `UPDATE direct_booking_sessions SET status = $1, updated_at = NOW() WHERE id = $2 AND tenant_id = $3`, status, id, tenantID)
	return err
}
