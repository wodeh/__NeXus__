package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// RateRuleRepository provides rate rule data access.
type RateRuleRepository struct {
	pool *db.Pool
}

// NewRateRuleRepository creates a rate rule repository.
func NewRateRuleRepository(pool *db.Pool) *RateRuleRepository {
	return &RateRuleRepository{pool: pool}
}

// Create adds a new rate rule.
func (r *RateRuleRepository) Create(ctx context.Context, tenantID uuid.UUID, req domain.RateRuleCreateRequest) (*domain.RateRule, error) {
	query := `
		INSERT INTO rate_rules (tenant_id, name, room_type, condition_type, condition_value, rate_adjustment_type, rate_adjustment_value, min_nights, max_nights, start_date, end_date, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, tenant_id, name, room_type, condition_type, condition_value, rate_adjustment_type, rate_adjustment_value, min_nights, max_nights, start_date, end_date, priority, active, created_at, updated_at
	`
	var rule domain.RateRule
	row := r.pool.QueryRow(ctx, query, tenantID, req.Name, strPtrOrNil(req.RoomType), req.ConditionType, req.ConditionValue, req.RateAdjustmentType, req.RateAdjustmentValue, req.MinNights, intPtrOrNil(req.MaxNights), strPtrOrNil(req.StartDate), strPtrOrNil(req.EndDate), req.Priority)
	if err := row.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.RoomType, &rule.ConditionType, &rule.ConditionValue, &rule.RateAdjustmentType, &rule.RateAdjustmentValue, &rule.MinNights, &rule.MaxNights, &rule.StartDate, &rule.EndDate, &rule.Priority, &rule.Active, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
		return nil, fmt.Errorf("create rate rule: %w", err)
	}
	return &rule, nil
}

// List returns all active rate rules for a tenant.
func (r *RateRuleRepository) List(ctx context.Context, tenantID uuid.UUID) ([]domain.RateRule, error) {
	query := `
		SELECT id, tenant_id, name, room_type, condition_type, condition_value, rate_adjustment_type, rate_adjustment_value, min_nights, max_nights, start_date, end_date, priority, active, created_at, updated_at
		FROM rate_rules
		WHERE tenant_id = $1 AND deleted_at IS NULL
		ORDER BY priority DESC, created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("list rate rules: %w", err)
	}
	defer rows.Close()

	var rules []domain.RateRule
	for rows.Next() {
		var rule domain.RateRule
		if err := rows.Scan(&rule.ID, &rule.TenantID, &rule.Name, &rule.RoomType, &rule.ConditionType, &rule.ConditionValue, &rule.RateAdjustmentType, &rule.RateAdjustmentValue, &rule.MinNights, &rule.MaxNights, &rule.StartDate, &rule.EndDate, &rule.Priority, &rule.Active, &rule.CreatedAt, &rule.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan rate rule: %w", err)
		}
		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rate rule rows: %w", err)
	}
	return rules, nil
}

// Delete soft-deletes a rate rule.
func (r *RateRuleRepository) Delete(ctx context.Context, tenantID uuid.UUID, id uuid.UUID) error {
	query := `UPDATE rate_rules SET deleted_at = now() WHERE id = $1 AND tenant_id = $2`
	_, err := r.pool.Exec(ctx, query, id, tenantID)
	if err != nil {
		return fmt.Errorf("delete rate rule: %w", err)
	}
	return nil
}

// CalculateRate computes the final rate for given parameters.
func (r *RateRuleRepository) CalculateRate(ctx context.Context, tenantID uuid.UUID, req domain.RateCalculateRequest) (*domain.RateCalculateResponse, error) {
	baseRate := 10000 // $100 base per night
	nights := daysBetween(req.CheckIn, req.CheckOut)
	if nights <= 0 {
		nights = 1
	}

	baseTotal := baseRate * nights
	resp := &domain.RateCalculateResponse{
		BaseRate: baseTotal,
		Nights:   nights,
		FinalRate: baseTotal,
	}

	// Fetch active rules
	rules, err := r.List(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	for _, rule := range rules {
		if !rule.Active {
			continue
		}
		if rule.RoomType != nil && *rule.RoomType != req.RoomType {
			continue
		}
		if rule.MinNights > 0 && nights < rule.MinNights {
			continue
		}
		if rule.MaxNights != nil && nights > *rule.MaxNights {
			continue
		}
		if rule.StartDate != nil && req.CheckIn < *rule.StartDate {
			continue
		}
		if rule.EndDate != nil && req.CheckOut > *rule.EndDate {
			continue
		}

		// Check condition
		matches := false
		switch rule.ConditionType {
		case "season":
			matches = true // simplified: always match if dates overlap
		case "day_of_week":
			matches = matchesDayOfWeek(req.CheckIn, rule.ConditionValue)
		case "length_of_stay":
			matches = fmt.Sprintf("%d", nights) == rule.ConditionValue
		case "advance_booking":
			matches = true // simplified
		case "occupancy":
			matches = true // simplified
		}

		if matches {
			var amount int
			switch rule.RateAdjustmentType {
			case "fixed_amount":
				amount = rule.RateAdjustmentValue * nights
			case "percentage":
				amount = (baseTotal * rule.RateAdjustmentValue) / 100
			case "fixed_rate":
				amount = rule.RateAdjustmentValue*nights - baseTotal
			}
			resp.Adjustments = append(resp.Adjustments, domain.RateAdjustment{
				RuleName: rule.Name,
				Type:     rule.RateAdjustmentType,
				Value:    rule.RateAdjustmentValue,
				Amount:   amount,
			})
			resp.FinalRate += amount
		}
	}

	return resp, nil
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func intPtrOrNil(n int) *int {
	if n == 0 {
		return nil
	}
	return &n
}

func daysBetween(a, b string) int {
	layout := "2006-01-02"
	ta, _ := time.Parse(layout, a)
	tb, _ := time.Parse(layout, b)
	return int(tb.Sub(ta).Hours() / 24)
}

func matchesDayOfWeek(checkIn, value string) bool {
	layout := "2006-01-02"
	t, _ := time.Parse(layout, checkIn)
	return t.Weekday().String() == value
}
