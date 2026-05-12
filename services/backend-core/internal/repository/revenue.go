package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// RevenueRepository handles revenue analytics and forecasting.
type RevenueRepository struct{ pool *db.Pool }

func NewRevenueRepository(pool *db.Pool) *RevenueRepository { return &RevenueRepository{pool: pool} }

func (r *RevenueRepository) GetDailyStats(ctx context.Context, tenantID uuid.UUID, date time.Time) (*domain.RevenueDashboard, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(total), 0) as total_revenue,
			COUNT(DISTINCT r.id) as occupied_rooms,
			(SELECT COUNT(*) FROM rooms WHERE tenant_id = $1 AND deleted_at IS NULL) as total_rooms,
			COUNT(CASE WHEN check_in = $2 THEN 1 END) as arrivals,
			COUNT(CASE WHEN check_out = $2 THEN 1 END) as departures
		FROM reservations r
		WHERE r.tenant_id = $1 AND r.deleted_at IS NULL 
		  AND r.status IN ('confirmed', 'checked_in')
		  AND $2 BETWEEN r.check_in AND r.check_out`, tenantID, date)

	var d domain.RevenueDashboard
	d.TenantID = tenantID
	d.Date = date
	err := row.Scan(&d.TotalRevenue, &d.OccupiedRooms, &d.TotalRooms, &d.Arrivals, &d.Departures)
	if err != nil { return nil, fmt.Errorf("get daily stats: %w", err) }

	if d.TotalRooms > 0 {
		d.OccupancyRate = float64(d.OccupiedRooms) / float64(d.TotalRooms) * 100
		d.AvailableRooms = d.TotalRooms - d.OccupiedRooms
	}
	if d.OccupiedRooms > 0 {
		d.ADR = d.TotalRevenue / float64(d.OccupiedRooms)
	}
	if d.TotalRooms > 0 {
		d.RevPAR = d.TotalRevenue / float64(d.TotalRooms)
	}
	d.RoomRevenue = d.TotalRevenue * 0.85
	d.TaxRevenue = d.TotalRevenue * 0.10
	d.ExtraRevenue = d.TotalRevenue * 0.05

	return &d, nil
}

func (r *RevenueRepository) GetRangeStats(ctx context.Context, tenantID uuid.UUID, from, to time.Time) ([]domain.RevenueDashboard, error) {
	var out []domain.RevenueDashboard
	for d := from; !d.After(to); d = d.AddDate(0, 0, 1) {
		stat, err := r.GetDailyStats(ctx, tenantID, d)
		if err != nil {
			continue
		}
		out = append(out, *stat)
	}
	return out, nil
}

func (r *RevenueRepository) GetForecast(ctx context.Context, tenantID uuid.UUID, from, to time.Time) ([]domain.RevenueForecast, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT 
			DATE(check_in) as date,
			COUNT(*) as bookings,
			SUM(total) as projected_revenue
		FROM reservations
		WHERE tenant_id = $1 AND deleted_at IS NULL 
		  AND check_in BETWEEN $2 AND $3
		  AND status IN ('confirmed', 'pending')
		GROUP BY DATE(check_in)
		ORDER BY date`, tenantID, from, to)
	if err != nil { return nil, fmt.Errorf("get forecast: %w", err) }
	defer rows.Close()

	var out []domain.RevenueForecast
	for rows.Next() {
		var f domain.RevenueForecast
		var date time.Time
		var bookings int
		var revenue float64
		if err := rows.Scan(&date, &bookings, &revenue); err != nil { continue }
		f.Date = date
		f.ProjectedRevenue = revenue
		f.BookedRooms = bookings
		f.TotalRooms = 40 // Demo total
		if f.TotalRooms > 0 {
			f.ProjectedOccupancy = float64(bookings) / float64(f.TotalRooms) * 100
		}
		f.Confidence = 0.85
		out = append(out, f)
	}
	return out, nil
}

func (r *RevenueRepository) GetChannelBreakdown(ctx context.Context, tenantID uuid.UUID, from, to time.Time) ([]domain.ChannelRevenueBreakdown, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT 
			COALESCE(source, 'unknown') as channel,
			COUNT(*) as bookings,
			COALESCE(SUM(total), 0) as revenue
		FROM reservations
		WHERE tenant_id = $1 AND deleted_at IS NULL 
		  AND check_in BETWEEN $2 AND $3
		  AND status IN ('confirmed', 'checked_in', 'checked_out')
		GROUP BY source`, tenantID, from, to)
	if err != nil { return nil, fmt.Errorf("get channel breakdown: %w", err) }
	defer rows.Close()

	var out []domain.ChannelRevenueBreakdown
	for rows.Next() {
		var b domain.ChannelRevenueBreakdown
		if err := rows.Scan(&b.Channel, &b.Bookings, &b.Revenue); err != nil { continue }
		b.Commission = b.Revenue * 0.15
		b.NetRevenue = b.Revenue - b.Commission
		if b.Bookings > 0 {
			b.AvgRate = b.Revenue / float64(b.Bookings)
		}
		out = append(out, b)
	}
	return out, nil
}

func (r *RevenueRepository) GetRoomTypeRevenue(ctx context.Context, tenantID uuid.UUID, from, to time.Time) ([]domain.RoomTypeRevenue, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT 
			room_type,
			COUNT(*) as nights_sold,
			COALESCE(SUM(total), 0) as revenue
		FROM reservations
		WHERE tenant_id = $1 AND deleted_at IS NULL 
		  AND check_in BETWEEN $2 AND $3
		  AND status IN ('confirmed', 'checked_in', 'checked_out')
		GROUP BY room_type`, tenantID, from, to)
	if err != nil { return nil, fmt.Errorf("get room type revenue: %w", err) }
	defer rows.Close()

	var out []domain.RoomTypeRevenue
	for rows.Next() {
		var rt domain.RoomTypeRevenue
		if err := rows.Scan(&rt.RoomType, &rt.NightsSold, &rt.Revenue); err != nil { continue }
		if rt.NightsSold > 0 {
			rt.AvgRate = rt.Revenue / float64(rt.NightsSold)
		}
		rt.OccupancyPct = 75.0 // Demo
		out = append(out, rt)
	}
	return out, nil
}

func (r *RevenueRepository) ListPricingRules(ctx context.Context, tenantID uuid.UUID) ([]domain.PricingRule, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, room_type, condition, trigger_value, adjustment_type, adjustment_value, min_rate, max_rate, is_active, priority, config, created_at, updated_at
		FROM pricing_rules WHERE tenant_id = $1 AND deleted_at IS NULL ORDER BY priority DESC`, tenantID)
	if err != nil { return nil, fmt.Errorf("list pricing rules: %w", err) }
	defer rows.Close()

	var out []domain.PricingRule
	for rows.Next() {
		var p domain.PricingRule
		var configJSON []byte
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.RoomType, &p.Condition, &p.TriggerValue, &p.AdjustmentType, &p.AdjustmentValue, &p.MinRate, &p.MaxRate, &p.IsActive, &p.Priority, &configJSON, &p.CreatedAt, &p.UpdatedAt); err != nil { continue }
		_ = unmarshalJSON(configJSON, &p.Config)
		out = append(out, p)
	}
	return out, nil
}

func (r *RevenueRepository) CreatePricingRule(ctx context.Context, p *domain.PricingRule) error {
	if p.ID == uuid.Nil { p.ID = uuid.New() }
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now

	_, err := r.pool.Exec(ctx, `
		INSERT INTO pricing_rules (id, tenant_id, name, room_type, condition, trigger_value, adjustment_type, adjustment_value, min_rate, max_rate, is_active, priority, config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		p.ID, p.TenantID, p.Name, p.RoomType, p.Condition, p.TriggerValue, p.AdjustmentType, p.AdjustmentValue, p.MinRate, p.MaxRate, p.IsActive, p.Priority, marshalJSON(p.Config), p.CreatedAt, p.UpdatedAt)
	if err != nil { return fmt.Errorf("create pricing rule: %w", err) }
	return nil
}
