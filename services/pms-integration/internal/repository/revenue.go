package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// RevenueForecastRepository defines forecast storage.
type RevenueForecastRepository interface {
	Create(ctx context.Context, f *domain.RevenueForecast) error
	ListByProperty(ctx context.Context, tenantID, propertyID string, start, end time.Time) ([]*domain.RevenueForecast, error)
	GetLatest(ctx context.Context, tenantID, propertyID string) (*domain.RevenueForecast, error)
}

// DynamicPricingRuleRepository defines rule storage.
type DynamicPricingRuleRepository interface {
	Create(ctx context.Context, r *domain.DynamicPricingRule) error
	List(ctx context.Context, tenantID string) ([]*domain.DynamicPricingRule, error)
	GetByID(ctx context.Context, tenantID, id string) (*domain.DynamicPricingRule, error)
	Update(ctx context.Context, tenantID string, r *domain.DynamicPricingRule) error
	Delete(ctx context.Context, tenantID, id string) error
}

// PriceRecommendationRepository defines recommendation storage.
type PriceRecommendationRepository interface {
	Create(ctx context.Context, pr *domain.PriceRecommendation) error
	ListPending(ctx context.Context, tenantID string, limit int) ([]*domain.PriceRecommendation, error)
	MarkApplied(ctx context.Context, tenantID, id string) error
}

// PostgresRevenueForecastRepository is a PostgreSQL implementation.
type PostgresRevenueForecastRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewRevenueForecastRepository creates a new repository.
func NewRevenueForecastRepository(pool *db.Pool, metrics *RepositoryMetrics) RevenueForecastRepository {
	return &PostgresRevenueForecastRepository{pool: pool, metrics: metrics}
}

func (r *PostgresRevenueForecastRepository) Create(ctx context.Context, f *domain.RevenueForecast) error {
	defer r.metrics.ObserveQuery("revenue_forecast_create")()
	query := `INSERT INTO revenue_forecasts (tenant_id, property_id, date, predicted_occupancy, predicted_adr, predicted_revPAR, confidence, model_version) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, f.TenantID, f.PropertyID, f.Date, f.PredictedOccupancy, f.PredictedADR, f.PredictedRevPAR, f.Confidence, f.ModelVersion).Scan(&f.ID, &f.CreatedAt)
}

func (r *PostgresRevenueForecastRepository) ListByProperty(ctx context.Context, tenantID, propertyID string, start, end time.Time) ([]*domain.RevenueForecast, error) {
	defer r.metrics.ObserveQuery("revenue_forecast_list")()
	query := `SELECT id, tenant_id, property_id, date, predicted_occupancy, predicted_adr, predicted_revPAR, confidence, model_version, created_at FROM revenue_forecasts WHERE tenant_id = $1 AND property_id = $2 AND date BETWEEN $3 AND $4 ORDER BY date`
	rows, err := r.pool.Query(ctx, query, tenantID, propertyID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var forecasts []*domain.RevenueForecast
	for rows.Next() {
		var f domain.RevenueForecast
		if err := rows.Scan(&f.ID, &f.TenantID, &f.PropertyID, &f.Date, &f.PredictedOccupancy, &f.PredictedADR, &f.PredictedRevPAR, &f.Confidence, &f.ModelVersion, &f.CreatedAt); err != nil {
			return nil, err
		}
		forecasts = append(forecasts, &f)
	}
	return forecasts, rows.Err()
}

func (r *PostgresRevenueForecastRepository) GetLatest(ctx context.Context, tenantID, propertyID string) (*domain.RevenueForecast, error) {
	defer r.metrics.ObserveQuery("revenue_forecast_latest")()
	var f domain.RevenueForecast
	query := `SELECT id, tenant_id, property_id, date, predicted_occupancy, predicted_adr, predicted_revPAR, confidence, model_version, created_at FROM revenue_forecasts WHERE tenant_id = $1 AND property_id = $2 ORDER BY date DESC LIMIT 1`
	err := r.pool.QueryRow(ctx, query, tenantID, propertyID).Scan(&f.ID, &f.TenantID, &f.PropertyID, &f.Date, &f.PredictedOccupancy, &f.PredictedADR, &f.PredictedRevPAR, &f.Confidence, &f.ModelVersion, &f.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no forecast found")
	}
	return &f, err
}

// PostgresDynamicPricingRuleRepository is a PostgreSQL implementation.
type PostgresDynamicPricingRuleRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewDynamicPricingRuleRepository creates a new repository.
func NewDynamicPricingRuleRepository(pool *db.Pool, metrics *RepositoryMetrics) DynamicPricingRuleRepository {
	return &PostgresDynamicPricingRuleRepository{pool: pool, metrics: metrics}
}

func (r *PostgresDynamicPricingRuleRepository) Create(ctx context.Context, rule *domain.DynamicPricingRule) error {
	defer r.metrics.ObserveQuery("pricing_rule_create")()
	query := `INSERT INTO dynamic_pricing_rules (tenant_id, property_id, room_type_id, name, min_advance_days, max_advance_days, min_los, lead_time_discount, last_minute_premium, occupancy_threshold, occupancy_premium, is_active) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id, created_at, updated_at`
	return r.pool.QueryRow(ctx, query, rule.TenantID, rule.PropertyID, rule.RoomTypeID, rule.Name, rule.MinAdvanceDays, rule.MaxAdvanceDays, rule.MinLOS, rule.LeadTimeDiscount, rule.LastMinutePremium, rule.OccupancyThreshold, rule.OccupancyPremium, rule.IsActive).Scan(&rule.ID, &rule.CreatedAt, &rule.UpdatedAt)
}

func (r *PostgresDynamicPricingRuleRepository) List(ctx context.Context, tenantID string) ([]*domain.DynamicPricingRule, error) {
	defer r.metrics.ObserveQuery("pricing_rule_list")()
	query := `SELECT id, tenant_id, property_id, room_type_id, name, min_advance_days, max_advance_days, min_los, lead_time_discount, last_minute_premium, occupancy_threshold, occupancy_premium, is_active, created_at, updated_at FROM dynamic_pricing_rules WHERE tenant_id = $1 ORDER BY name`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []*domain.DynamicPricingRule
	for rows.Next() {
		var rule domain.DynamicPricingRule
		if err := rows.Scan(&rule.ID, &rule.TenantID, &rule.PropertyID, &rule.RoomTypeID, &rule.Name, &rule.MinAdvanceDays, &rule.MaxAdvanceDays, &rule.MinLOS, &rule.LeadTimeDiscount, &rule.LastMinutePremium, &rule.OccupancyThreshold, &rule.OccupancyPremium, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, &rule)
	}
	return rules, rows.Err()
}

func (r *PostgresDynamicPricingRuleRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.DynamicPricingRule, error) {
	defer r.metrics.ObserveQuery("pricing_rule_get")()
	var rule domain.DynamicPricingRule
	query := `SELECT id, tenant_id, property_id, room_type_id, name, min_advance_days, max_advance_days, min_los, lead_time_discount, last_minute_premium, occupancy_threshold, occupancy_premium, is_active, created_at, updated_at FROM dynamic_pricing_rules WHERE tenant_id = $1 AND id = $2`
	err := r.pool.QueryRow(ctx, query, tenantID, id).Scan(&rule.ID, &rule.TenantID, &rule.PropertyID, &rule.RoomTypeID, &rule.Name, &rule.MinAdvanceDays, &rule.MaxAdvanceDays, &rule.MinLOS, &rule.LeadTimeDiscount, &rule.LastMinutePremium, &rule.OccupancyThreshold, &rule.OccupancyPremium, &rule.IsActive, &rule.CreatedAt, &rule.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("pricing rule not found")
	}
	return &rule, err
}

func (r *PostgresDynamicPricingRuleRepository) Update(ctx context.Context, tenantID string, rule *domain.DynamicPricingRule) error {
	defer r.metrics.ObserveQuery("pricing_rule_update")()
	query := `UPDATE dynamic_pricing_rules SET name = $1, min_advance_days = $2, max_advance_days = $3, min_los = $4, lead_time_discount = $5, last_minute_premium = $6, occupancy_threshold = $7, occupancy_premium = $8, is_active = $9, updated_at = NOW() WHERE tenant_id = $10 AND id = $11`
	_, err := r.pool.Exec(ctx, query, rule.Name, rule.MinAdvanceDays, rule.MaxAdvanceDays, rule.MinLOS, rule.LeadTimeDiscount, rule.LastMinutePremium, rule.OccupancyThreshold, rule.OccupancyPremium, rule.IsActive, tenantID, rule.ID)
	return err
}

func (r *PostgresDynamicPricingRuleRepository) Delete(ctx context.Context, tenantID, id string) error {
	defer r.metrics.ObserveQuery("pricing_rule_delete")()
	_, err := r.pool.Exec(ctx, `DELETE FROM dynamic_pricing_rules WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	return err
}

// PostgresPriceRecommendationRepository is a PostgreSQL implementation.
type PostgresPriceRecommendationRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewPriceRecommendationRepository creates a new repository.
func NewPriceRecommendationRepository(pool *db.Pool, metrics *RepositoryMetrics) PriceRecommendationRepository {
	return &PostgresPriceRecommendationRepository{pool: pool, metrics: metrics}
}

func (r *PostgresPriceRecommendationRepository) Create(ctx context.Context, pr *domain.PriceRecommendation) error {
	defer r.metrics.ObserveQuery("price_recommendation_create")()
	query := `INSERT INTO price_recommendations (tenant_id, property_id, room_type_id, date, current_rate, suggested_rate, change_percent, reason, confidence, applied) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, created_at`
	return r.pool.QueryRow(ctx, query, pr.TenantID, pr.PropertyID, pr.RoomTypeID, pr.Date, pr.CurrentRate, pr.SuggestedRate, pr.ChangePercent, pr.Reason, pr.Confidence, pr.Applied).Scan(&pr.ID, &pr.CreatedAt)
}

func (r *PostgresPriceRecommendationRepository) ListPending(ctx context.Context, tenantID string, limit int) ([]*domain.PriceRecommendation, error) {
	defer r.metrics.ObserveQuery("price_recommendation_pending")()
	query := `SELECT id, tenant_id, property_id, room_type_id, date, current_rate, suggested_rate, change_percent, reason, confidence, applied, created_at FROM price_recommendations WHERE tenant_id = $1 AND applied = FALSE ORDER BY created_at DESC LIMIT $2`
	rows, err := r.pool.Query(ctx, query, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var recs []*domain.PriceRecommendation
	for rows.Next() {
		var pr domain.PriceRecommendation
		if err := rows.Scan(&pr.ID, &pr.TenantID, &pr.PropertyID, &pr.RoomTypeID, &pr.Date, &pr.CurrentRate, &pr.SuggestedRate, &pr.ChangePercent, &pr.Reason, &pr.Confidence, &pr.Applied, &pr.CreatedAt); err != nil {
			return nil, err
		}
		recs = append(recs, &pr)
	}
	return recs, rows.Err()
}

func (r *PostgresPriceRecommendationRepository) MarkApplied(ctx context.Context, tenantID, id string) error {
	defer r.metrics.ObserveQuery("price_recommendation_apply")()
	_, err := r.pool.Exec(ctx, `UPDATE price_recommendations SET applied = TRUE WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	return err
}
