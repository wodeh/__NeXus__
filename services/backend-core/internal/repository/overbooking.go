package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/db"
)

// OverbookingRepository provides overbooking confidence calculations.
type OverbookingRepository struct {
	pool *db.Pool
}

// NewOverbookingRepository creates an overbooking repository.
func NewOverbookingRepository(pool *db.Pool) *OverbookingRepository {
	return &OverbookingRepository{pool: pool}
}

// ConfidenceScore represents overbooking safety for a room type.
type ConfidenceScore struct {
	RoomType          string  `json:"room_type"`
	Confidence        float64 `json:"confidence"`
	NoShowRate        float64 `json:"no_show_rate"`
	CancellationRate  float64 `json:"cancellation_rate"`
	CurrentOccupancy  float64 `json:"current_occupancy"`
	DayOfWeekRisk     float64 `json:"day_of_week_risk"`
	SuggestedOverbook int     `json:"suggested_overbook"`
}

// GetConfidenceScores calculates overbooking confidence for all room types.
func (r *OverbookingRepository) GetConfidenceScores(ctx context.Context, tenantID uuid.UUID, date string) ([]ConfidenceScore, error) {
	// Get room types
	roomTypes, err := r.getRoomTypes(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Get historical rates (last 90 days)
	noShowRates, cancelRates, err := r.getHistoricalRates(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// Get current occupancy by room type
	occupancy, err := r.getCurrentOccupancy(ctx, tenantID, date)
	if err != nil {
		return nil, err
	}

	// Day of week risk factor
	dowRisk := r.getDayOfWeekRisk(date)

	var scores []ConfidenceScore
	for _, rt := range roomTypes {
		nsRate := noShowRates[rt]
		cRate := cancelRates[rt]
		occ := occupancy[rt]

		// Formula: confidence = 100 - weighted risk factors
		confidence := 100.0 - (nsRate*40.0 + cRate*30.0 + occ*20.0 + dowRisk*10.0)
		if confidence < 0 {
			confidence = 0
		}
		if confidence > 100 {
			confidence = 100
		}

		// Suggested overbook rooms based on confidence
		suggested := 0
		if confidence > 85 {
			suggested = 2
		} else if confidence > 70 {
			suggested = 1
		}

		scores = append(scores, ConfidenceScore{
			RoomType:          rt,
			Confidence:        confidence,
			NoShowRate:        nsRate * 100,
			CancellationRate:  cRate * 100,
			CurrentOccupancy:  occ * 100,
			DayOfWeekRisk:     dowRisk * 100,
			SuggestedOverbook: suggested,
		})
	}

	return scores, nil
}

func (r *OverbookingRepository) getRoomTypes(ctx context.Context, tenantID uuid.UUID) ([]string, error) {
	query := `SELECT DISTINCT room_type FROM rooms WHERE tenant_id = $1 AND deleted_at IS NULL`
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var types []string
	for rows.Next() {
		var rt string
		if err := rows.Scan(&rt); err == nil {
			types = append(types, rt)
		}
	}
	return types, nil
}

func (r *OverbookingRepository) getHistoricalRates(ctx context.Context, tenantID uuid.UUID) (map[string]float64, map[string]float64, error) {
	// No-show rate by room type (last 90 days)
	noShowQuery := `
		SELECT room_type, 
			COUNT(*) FILTER (WHERE status = 'no_show')::float / NULLIF(COUNT(*), 0) as rate
		FROM reservations
		WHERE tenant_id = $1 AND check_in >= now() - interval '90 days'
		GROUP BY room_type
	`
	noShowRates := make(map[string]float64)
	rows, err := r.pool.Query(ctx, noShowQuery, tenantID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var rt string
		var rate float64
		if err := rows.Scan(&rt, &rate); err == nil {
			noShowRates[rt] = rate
		}
	}

	// Cancellation rate by room type
	cancelQuery := `
		SELECT room_type, 
			COUNT(*) FILTER (WHERE status = 'cancelled')::float / NULLIF(COUNT(*), 0) as rate
		FROM reservations
		WHERE tenant_id = $1 AND created_at >= now() - interval '90 days'
		GROUP BY room_type
	`
	cancelRates := make(map[string]float64)
	rows2, err := r.pool.Query(ctx, cancelQuery, tenantID)
	if err != nil {
		return nil, nil, err
	}
	defer rows2.Close()
	for rows2.Next() {
		var rt string
		var rate float64
		if err := rows2.Scan(&rt, &rate); err == nil {
			cancelRates[rt] = rate
		}
	}

	return noShowRates, cancelRates, nil
}

func (r *OverbookingRepository) getCurrentOccupancy(ctx context.Context, tenantID uuid.UUID, date string) (map[string]float64, error) {
	query := `
		SELECT r.room_type,
			COUNT(*) FILTER (WHERE res.status IN ('confirmed', 'checked_in'))::float / NULLIF(COUNT(*), 0) as occ
		FROM rooms r
		LEFT JOIN reservations res ON res.room_number = r.number 
			AND res.tenant_id = r.tenant_id
			AND res.check_in <= $2::date AND res.check_out > $2::date
			AND res.status NOT IN ('cancelled', 'no_show')
		WHERE r.tenant_id = $1 AND r.deleted_at IS NULL
		GROUP BY r.room_type
	`
	rows, err := r.pool.Query(ctx, query, tenantID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	occ := make(map[string]float64)
	for rows.Next() {
		var rt string
		var rate float64
		if err := rows.Scan(&rt, &rate); err == nil {
			occ[rt] = rate
		}
	}
	return occ, nil
}

func (r *OverbookingRepository) getDayOfWeekRisk(dateStr string) float64 {
	t, _ := time.Parse("2006-01-02", dateStr)
	wday := t.Weekday()
	// Weekends (Fri-Sat) are higher risk for no-shows
	if wday == time.Friday || wday == time.Saturday {
		return 0.6
	}
	if wday == time.Sunday {
		return 0.4
	}
	return 0.2
}
