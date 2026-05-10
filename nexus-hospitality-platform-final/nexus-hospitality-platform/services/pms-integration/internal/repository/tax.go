package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// TaxRateRepository defines the interface for tax rate storage.
type TaxRateRepository interface {
	Create(ctx context.Context, tr *domain.TaxRate) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.TaxRate, error)
	ListByJurisdiction(ctx context.Context, tenantID, countryCode, stateCode string) ([]*domain.TaxRate, error)
	ListActive(ctx context.Context, tenantID string) ([]*domain.TaxRate, error)
	Update(ctx context.Context, tenantID string, tr *domain.TaxRate) error
	Delete(ctx context.Context, tenantID, id string) error
}

// TaxExemptionRepository defines the interface for tax exemption storage.
type TaxExemptionRepository interface {
	Create(ctx context.Context, te *domain.TaxExemption) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.TaxExemption, error)
	ListByGuest(ctx context.Context, tenantID, guestID string) ([]*domain.TaxExemption, error)
	Update(ctx context.Context, tenantID string, te *domain.TaxExemption) error
	Delete(ctx context.Context, tenantID, id string) error
}

// PostgresTaxRateRepository is a PostgreSQL implementation of TaxRateRepository.
type PostgresTaxRateRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewTaxRateRepository creates a new PostgresTaxRateRepository.
func NewTaxRateRepository(pool *db.Pool, metrics *RepositoryMetrics) TaxRateRepository {
	return &PostgresTaxRateRepository{pool: pool, metrics: metrics}
}

func (r *PostgresTaxRateRepository) Create(ctx context.Context, tr *domain.TaxRate) error {
	defer r.metrics.ObserveQuery("tax_rate_create")()
	query := `
		INSERT INTO tax_rates (tenant_id, name, jurisdiction, country_code, state_code, city_code, rate, type, applies_to, is_compound, is_active, effective_from, effective_to)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at`
	return r.pool.QueryRowContext(ctx, query,
		tr.TenantID, tr.Name, tr.Jurisdiction, tr.CountryCode, tr.StateCode, tr.CityCode,
		tr.Rate, tr.Type, tr.AppliesTo, tr.IsCompound, tr.IsActive, tr.EffectiveFrom, tr.EffectiveTo,
	).Scan(&tr.ID, &tr.CreatedAt, &tr.UpdatedAt)
}

func (r *PostgresTaxRateRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.TaxRate, error) {
	defer r.metrics.ObserveQuery("tax_rate_get_by_id")()
	var tr domain.TaxRate
	query := `SELECT id, tenant_id, name, jurisdiction, country_code, state_code, city_code, rate, type, applies_to, is_compound, is_active, effective_from, effective_to, created_at, updated_at FROM tax_rates WHERE tenant_id = $1 AND id = $2`
	err := r.pool.QueryRowContext(ctx, query, tenantID, id).Scan(
		&tr.ID, &tr.TenantID, &tr.Name, &tr.Jurisdiction, &tr.CountryCode, &tr.StateCode, &tr.CityCode,
		&tr.Rate, &tr.Type, &tr.AppliesTo, &tr.IsCompound, &tr.IsActive, &tr.EffectiveFrom, &tr.EffectiveTo,
		&tr.CreatedAt, &tr.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tax rate not found")
	}
	return &tr, err
}

func (r *PostgresTaxRateRepository) ListByJurisdiction(ctx context.Context, tenantID, countryCode, stateCode string) ([]*domain.TaxRate, error) {
	defer r.metrics.ObserveQuery("tax_rate_list_jurisdiction")()
	query := `SELECT id, tenant_id, name, jurisdiction, country_code, state_code, city_code, rate, type, applies_to, is_compound, is_active, effective_from, effective_to, created_at, updated_at FROM tax_rates WHERE tenant_id = $1 AND country_code = $2 AND (state_code = $3 OR state_code IS NULL) AND is_active = true ORDER BY rate DESC`
	rows, err := r.pool.QueryContext(ctx, query, tenantID, countryCode, stateCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []*domain.TaxRate
	for rows.Next() {
		var tr domain.TaxRate
		if err := rows.Scan(&tr.ID, &tr.TenantID, &tr.Name, &tr.Jurisdiction, &tr.CountryCode, &tr.StateCode, &tr.CityCode,
			&tr.Rate, &tr.Type, &tr.AppliesTo, &tr.IsCompound, &tr.IsActive, &tr.EffectiveFrom, &tr.EffectiveTo,
			&tr.CreatedAt, &tr.UpdatedAt); err != nil {
			return nil, err
		}
		rates = append(rates, &tr)
	}
	return rates, rows.Err()
}

func (r *PostgresTaxRateRepository) ListActive(ctx context.Context, tenantID string) ([]*domain.TaxRate, error) {
	defer r.metrics.ObserveQuery("tax_rate_list_active")()
	query := `SELECT id, tenant_id, name, jurisdiction, country_code, state_code, city_code, rate, type, applies_to, is_compound, is_active, effective_from, effective_to, created_at, updated_at FROM tax_rates WHERE tenant_id = $1 AND is_active = true ORDER BY jurisdiction, type`
	rows, err := r.pool.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []*domain.TaxRate
	for rows.Next() {
		var tr domain.TaxRate
		if err := rows.Scan(&tr.ID, &tr.TenantID, &tr.Name, &tr.Jurisdiction, &tr.CountryCode, &tr.StateCode, &tr.CityCode,
			&tr.Rate, &tr.Type, &tr.AppliesTo, &tr.IsCompound, &tr.IsActive, &tr.EffectiveFrom, &tr.EffectiveTo,
			&tr.CreatedAt, &tr.UpdatedAt); err != nil {
			return nil, err
		}
		rates = append(rates, &tr)
	}
	return rates, rows.Err()
}

func (r *PostgresTaxRateRepository) Update(ctx context.Context, tenantID string, tr *domain.TaxRate) error {
	defer r.metrics.ObserveQuery("tax_rate_update")()
	query := `UPDATE tax_rates SET name = $1, jurisdiction = $2, country_code = $3, state_code = $4, city_code = $5, rate = $6, type = $7, applies_to = $8, is_compound = $9, is_active = $10, effective_from = $11, effective_to = $12, updated_at = NOW() WHERE tenant_id = $13 AND id = $14`
	_, err := r.pool.ExecContext(ctx, query,
		tr.Name, tr.Jurisdiction, tr.CountryCode, tr.StateCode, tr.CityCode,
		tr.Rate, tr.Type, tr.AppliesTo, tr.IsCompound, tr.IsActive, tr.EffectiveFrom, tr.EffectiveTo,
		tenantID, tr.ID,
	)
	return err
}

func (r *PostgresTaxRateRepository) Delete(ctx context.Context, tenantID, id string) error {
	defer r.metrics.ObserveQuery("tax_rate_delete")()
	_, err := r.pool.ExecContext(ctx, `DELETE FROM tax_rates WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	return err
}

// PostgresTaxExemptionRepository is a PostgreSQL implementation of TaxExemptionRepository.
type PostgresTaxExemptionRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewTaxExemptionRepository creates a new PostgresTaxExemptionRepository.
func NewTaxExemptionRepository(pool *db.Pool, metrics *RepositoryMetrics) TaxExemptionRepository {
	return &PostgresTaxExemptionRepository{pool: pool, metrics: metrics}
}

func (r *PostgresTaxExemptionRepository) Create(ctx context.Context, te *domain.TaxExemption) error {
	defer r.metrics.ObserveQuery("tax_exemption_create")()
	query := `INSERT INTO tax_exemptions (tenant_id, guest_id, corporate_id, certificate_num, jurisdiction, tax_type, exempt_percent, valid_from, valid_to, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, created_at`
	return r.pool.QueryRowContext(ctx, query,
		te.TenantID, te.GuestID, te.CorporateID, te.CertificateNum, te.Jurisdiction,
		te.TaxType, te.ExemptPercent, te.ValidFrom, te.ValidTo, te.IsActive,
	).Scan(&te.ID, &te.CreatedAt)
}

func (r *PostgresTaxExemptionRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.TaxExemption, error) {
	defer r.metrics.ObserveQuery("tax_exemption_get_by_id")()
	var te domain.TaxExemption
	query := `SELECT id, tenant_id, guest_id, corporate_id, certificate_num, jurisdiction, tax_type, exempt_percent, valid_from, valid_to, is_active, created_at FROM tax_exemptions WHERE tenant_id = $1 AND id = $2`
	err := r.pool.QueryRowContext(ctx, query, tenantID, id).Scan(
		&te.ID, &te.TenantID, &te.GuestID, &te.CorporateID, &te.CertificateNum, &te.Jurisdiction,
		&te.TaxType, &te.ExemptPercent, &te.ValidFrom, &te.ValidTo, &te.IsActive, &te.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tax exemption not found")
	}
	return &te, err
}

func (r *PostgresTaxExemptionRepository) ListByGuest(ctx context.Context, tenantID, guestID string) ([]*domain.TaxExemption, error) {
	defer r.metrics.ObserveQuery("tax_exemption_list_guest")()
	query := `SELECT id, tenant_id, guest_id, corporate_id, certificate_num, jurisdiction, tax_type, exempt_percent, valid_from, valid_to, is_active, created_at FROM tax_exemptions WHERE tenant_id = $1 AND guest_id = $2 AND is_active = true`
	rows, err := r.pool.QueryContext(ctx, query, tenantID, guestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exemptions []*domain.TaxExemption
	for rows.Next() {
		var te domain.TaxExemption
		if err := rows.Scan(&te.ID, &te.TenantID, &te.GuestID, &te.CorporateID, &te.CertificateNum, &te.Jurisdiction,
			&te.TaxType, &te.ExemptPercent, &te.ValidFrom, &te.ValidTo, &te.IsActive, &te.CreatedAt); err != nil {
			return nil, err
		}
		exemptions = append(exemptions, &te)
	}
	return exemptions, rows.Err()
}

func (r *PostgresTaxExemptionRepository) Update(ctx context.Context, tenantID string, te *domain.TaxExemption) error {
	defer r.metrics.ObserveQuery("tax_exemption_update")()
	query := `UPDATE tax_exemptions SET guest_id = $1, corporate_id = $2, certificate_num = $3, jurisdiction = $4, tax_type = $5, exempt_percent = $6, valid_from = $7, valid_to = $8, is_active = $9 WHERE tenant_id = $10 AND id = $11`
	_, err := r.pool.ExecContext(ctx, query,
		te.GuestID, te.CorporateID, te.CertificateNum, te.Jurisdiction,
		te.TaxType, te.ExemptPercent, te.ValidFrom, te.ValidTo, te.IsActive,
		tenantID, te.ID,
	)
	return err
}

func (r *PostgresTaxExemptionRepository) Delete(ctx context.Context, tenantID, id string) error {
	defer r.metrics.ObserveQuery("tax_exemption_delete")()
	_, err := r.pool.ExecContext(ctx, `DELETE FROM tax_exemptions WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	return err
}
