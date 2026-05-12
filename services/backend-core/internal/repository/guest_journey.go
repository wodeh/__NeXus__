package repository

import (
	"context"
	"time"

	"nexus-hospitality-platform/services/backend-core/internal/db"
	"nexus-hospitality-platform/services/backend-core/internal/domain"
)

type GuestJourneyRepository struct {
	pool *db.Pool
}

func NewGuestJourneyRepository(pool *db.Pool) *GuestJourneyRepository {
	return &GuestJourneyRepository{pool: pool}
}

// Journey CRUD
func (r *GuestJourneyRepository) ListJourneys(ctx context.Context, tenantID string) ([]domain.GuestJourney, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, trigger, is_active, steps, created_at, updated_at
		FROM guest_journeys WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var journeys []domain.GuestJourney
	for rows.Next() {
		var j domain.GuestJourney
		var stepsJSON []byte
		err := rows.Scan(&j.ID, &j.TenantID, &j.Name, &j.Trigger, &j.IsActive, &stepsJSON, &j.CreatedAt, &j.UpdatedAt)
		if err != nil {
			continue
		}
		// stepsJSON would be unmarshaled in real implementation
		journeys = append(journeys, j)
	}
	return journeys, nil
}

func (r *GuestJourneyRepository) CreateJourney(ctx context.Context, j *domain.GuestJourney) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO guest_journeys (id, tenant_id, name, trigger, is_active, steps, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		j.ID, j.TenantID, j.Name, j.Trigger, j.IsActive, "[]", j.CreatedAt, j.UpdatedAt)
	return err
}

func (r *GuestJourneyRepository) UpdateJourney(ctx context.Context, j *domain.GuestJourney) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE guest_journeys SET name = $1, trigger = $2, is_active = $3, steps = $4, updated_at = $5
		WHERE id = $6 AND tenant_id = $7`,
		j.Name, j.Trigger, j.IsActive, "[]", time.Now(), j.ID, j.TenantID)
	return err
}

func (r *GuestJourneyRepository) DeleteJourney(ctx context.Context, id, tenantID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM guest_journeys WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

// Execution tracking
func (r *GuestJourneyRepository) ListExecutions(ctx context.Context, tenantID string) ([]domain.GuestJourneyExecution, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, journey_id, reservation_id, guest_phone, current_step, total_steps, status, started_at, completed_at, next_trigger_at, created_at, updated_at
		FROM guest_journey_executions WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var execs []domain.GuestJourneyExecution
	for rows.Next() {
		var e domain.GuestJourneyExecution
		err := rows.Scan(&e.ID, &e.TenantID, &e.JourneyID, &e.ReservationID, &e.GuestPhone, &e.CurrentStep, &e.TotalSteps, &e.Status, &e.StartedAt, &e.CompletedAt, &e.NextTriggerAt, &e.CreatedAt, &e.UpdatedAt)
		if err != nil {
			continue
		}
		execs = append(execs, e)
	}
	return execs, nil
}

func (r *GuestJourneyRepository) CreateExecution(ctx context.Context, e *domain.GuestJourneyExecution) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO guest_journey_executions (id, tenant_id, journey_id, reservation_id, guest_phone, current_step, total_steps, status, started_at, next_trigger_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
		e.ID, e.TenantID, e.JourneyID, e.ReservationID, e.GuestPhone, e.CurrentStep, e.TotalSteps, e.Status, e.StartedAt, e.NextTriggerAt, e.CreatedAt, e.UpdatedAt)
	return err
}

func (r *GuestJourneyRepository) UpdateExecution(ctx context.Context, e *domain.GuestJourneyExecution) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE guest_journey_executions SET current_step = $1, status = $2, completed_at = $3, next_trigger_at = $4, updated_at = $5
		WHERE id = $6 AND tenant_id = $7`,
		e.CurrentStep, e.Status, e.CompletedAt, e.NextTriggerAt, time.Now(), e.ID, e.TenantID)
	return err
}

// Upsell CRUD
func (r *GuestJourneyRepository) ListOffers(ctx context.Context, tenantID string) ([]domain.UpsellOffer, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, description, category, price, currency, image_url, is_active, auto_offer, display_order, created_at, updated_at
		FROM upsell_offers WHERE tenant_id = $1 ORDER BY display_order`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var offers []domain.UpsellOffer
	for rows.Next() {
		var o domain.UpsellOffer
		err := rows.Scan(&o.ID, &o.TenantID, &o.Name, &o.Description, &o.Category, &o.Price, &o.Currency, &o.ImageURL, &o.IsActive, &o.AutoOffer, &o.DisplayOrder, &o.CreatedAt, &o.UpdatedAt)
		if err != nil {
			continue
		}
		// Compute fields would be queried in real implementation
		o.TotalSold = 0
		o.RevenueGenerated = 0
		offers = append(offers, o)
	}
	return offers, nil
}

func (r *GuestJourneyRepository) CreateOffer(ctx context.Context, o *domain.UpsellOffer) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO upsell_offers (id, tenant_id, name, description, category, price, currency, image_url, is_active, auto_offer, display_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		o.ID, o.TenantID, o.Name, o.Description, o.Category, o.Price, o.Currency, o.ImageURL, o.IsActive, o.AutoOffer, o.DisplayOrder, o.CreatedAt, o.UpdatedAt)
	return err
}

func (r *GuestJourneyRepository) UpdateOffer(ctx context.Context, o *domain.UpsellOffer) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE upsell_offers SET name = $1, description = $2, category = $3, price = $4, currency = $5, image_url = $6, is_active = $7, auto_offer = $8, display_order = $9, updated_at = $10
		WHERE id = $11 AND tenant_id = $12`,
		o.Name, o.Description, o.Category, o.Price, o.Currency, o.ImageURL, o.IsActive, o.AutoOffer, o.DisplayOrder, time.Now(), o.ID, o.TenantID)
	return err
}

func (r *GuestJourneyRepository) DeleteOffer(ctx context.Context, id, tenantID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM upsell_offers WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

// Upsell Purchase CRUD
func (r *GuestJourneyRepository) ListPurchases(ctx context.Context, tenantID string) ([]domain.UpsellPurchase, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, reservation_id, guest_phone, offer_id, offer_name, price, currency, status, payment_method, folio_posted, journey_step_id, created_at, updated_at
		FROM upsell_purchases WHERE tenant_id = $1 ORDER BY created_at DESC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var purchases []domain.UpsellPurchase
	for rows.Next() {
		var p domain.UpsellPurchase
		err := rows.Scan(&p.ID, &p.TenantID, &p.ReservationID, &p.GuestPhone, &p.OfferID, &p.OfferName, &p.Price, &p.Currency, &p.Status, &p.PaymentMethod, &p.FolioPosted, &p.JourneyStepID, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			continue
		}
		purchases = append(purchases, p)
	}
	return purchases, nil
}

func (r *GuestJourneyRepository) CreatePurchase(ctx context.Context, p *domain.UpsellPurchase) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO upsell_purchases (id, tenant_id, reservation_id, guest_phone, offer_id, offer_name, price, currency, status, payment_method, folio_posted, journey_step_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		p.ID, p.TenantID, p.ReservationID, p.GuestPhone, p.OfferID, p.OfferName, p.Price, p.Currency, p.Status, p.PaymentMethod, p.FolioPosted, p.JourneyStepID, p.CreatedAt, p.UpdatedAt)
	return err
}

// Competitor CRUD
func (r *GuestJourneyRepository) ListCompetitors(ctx context.Context, tenantID string) ([]domain.CompetitorHotel, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, address, city, country, star_rating, room_count, website, booking_url, is_active, last_scraped, created_at, updated_at
		FROM competitor_hotels WHERE tenant_id = $1 AND is_active = true ORDER BY name`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var hotels []domain.CompetitorHotel
	for rows.Next() {
		var h domain.CompetitorHotel
		err := rows.Scan(&h.ID, &h.TenantID, &h.Name, &h.Address, &h.City, &h.Country, &h.StarRating, &h.RoomCount, &h.Website, &h.BookingURL, &h.IsActive, &h.LastScraped, &h.CreatedAt, &h.UpdatedAt)
		if err != nil {
			continue
		}
		hotels = append(hotels, h)
	}
	return hotels, nil
}

func (r *GuestJourneyRepository) CreateCompetitor(ctx context.Context, h *domain.CompetitorHotel) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO competitor_hotels (id, tenant_id, name, address, city, country, star_rating, room_count, website, booking_url, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		h.ID, h.TenantID, h.Name, h.Address, h.City, h.Country, h.StarRating, h.RoomCount, h.Website, h.BookingURL, h.IsActive, h.CreatedAt, h.UpdatedAt)
	return err
}

func (r *GuestJourneyRepository) DeleteCompetitor(ctx context.Context, id, tenantID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM competitor_hotels WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

// Competitor Rates
func (r *GuestJourneyRepository) GetLatestRates(ctx context.Context, tenantID string) ([]domain.CompetitorRate, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, competitor_id, competitor_name, room_type, date, rate, currency, availability, min_stay, is_promo, source, scraped_at, created_at
		FROM competitor_rates WHERE tenant_id = $1
		AND scraped_at = (SELECT MAX(scraped_at) FROM competitor_rates cr2 WHERE cr2.competitor_id = competitor_rates.competitor_id AND cr2.room_type = competitor_rates.room_type AND cr2.date = competitor_rates.date)
		ORDER BY competitor_name, room_type, date`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rates []domain.CompetitorRate
	for rows.Next() {
		var rate domain.CompetitorRate
		err := rows.Scan(&rate.ID, &rate.TenantID, &rate.CompetitorID, &rate.CompetitorName, &rate.RoomType, &rate.Date, &rate.Rate, &rate.Currency, &rate.Availability, &rate.MinStay, &rate.IsPromo, &rate.Source, &rate.ScrapedAt, &rate.CreatedAt)
		if err != nil {
			continue
		}
		rates = append(rates, rate)
	}
	return rates, nil
}

func (r *GuestJourneyRepository) AddRate(ctx context.Context, rate *domain.CompetitorRate) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO competitor_rates (id, tenant_id, competitor_id, competitor_name, room_type, date, rate, currency, availability, min_stay, is_promo, source, scraped_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		rate.ID, rate.TenantID, rate.CompetitorID, rate.CompetitorName, rate.RoomType, rate.Date, rate.Rate, rate.Currency, rate.Availability, rate.MinStay, rate.IsPromo, rate.Source, rate.ScrapedAt, rate.CreatedAt)
	return err
}

// Rate Recommendations
func (r *GuestJourneyRepository) ListRecommendations(ctx context.Context, tenantID string) ([]domain.RateRecommendation, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, room_type, date, current_rate, recommended_rate, confidence, reason, factors, applied, applied_at, created_at
		FROM rate_recommendations WHERE tenant_id = $1 AND applied = false ORDER BY date`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var recs []domain.RateRecommendation
	for rows.Next() {
		var rec domain.RateRecommendation
		var factorsJSON []byte
		err := rows.Scan(&rec.ID, &rec.TenantID, &rec.RoomType, &rec.Date, &rec.CurrentRate, &rec.RecommendedRate, &rec.Confidence, &rec.Reason, &factorsJSON, &rec.Applied, &rec.AppliedAt, &rec.CreatedAt)
		if err != nil {
			continue
		}
		recs = append(recs, rec)
	}
	return recs, nil
}

func (r *GuestJourneyRepository) CreateRecommendation(ctx context.Context, rec *domain.RateRecommendation) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO rate_recommendations (id, tenant_id, room_type, date, current_rate, recommended_rate, confidence, reason, factors, applied, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		rec.ID, rec.TenantID, rec.RoomType, rec.Date, rec.CurrentRate, rec.RecommendedRate, rec.Confidence, rec.Reason, "[]", rec.Applied, rec.CreatedAt)
	return err
}

func (r *GuestJourneyRepository) ApplyRecommendation(ctx context.Context, id, tenantID string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE rate_recommendations SET applied = true, applied_at = $1 WHERE id = $2 AND tenant_id = $3`,
		time.Now(), id, tenantID)
	return err
}

// Rate Shop Config
func (r *GuestJourneyRepository) GetRateShopConfig(ctx context.Context, tenantID string) (*domain.RateShopConfig, error) {
	var c domain.RateShopConfig
	err := r.pool.QueryRow(ctx, `
		SELECT tenant_id, enabled, frequency_hours, lookahead_days, auto_adjust, max_adjustment_pct, created_at, updated_at
		FROM rate_shop_configs WHERE tenant_id = $1`, tenantID).Scan(
		&c.TenantID, &c.Enabled, &c.FrequencyHours, &c.LookaheadDays, &c.AutoAdjust, &c.MaxAdjustmentPct, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *GuestJourneyRepository) SaveRateShopConfig(ctx context.Context, c *domain.RateShopConfig) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO rate_shop_configs (tenant_id, enabled, frequency_hours, lookahead_days, auto_adjust, max_adjustment_pct, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (tenant_id) DO UPDATE SET
			enabled = EXCLUDED.enabled, frequency_hours = EXCLUDED.frequency_hours, lookahead_days = EXCLUDED.lookahead_days,
			auto_adjust = EXCLUDED.auto_adjust, max_adjustment_pct = EXCLUDED.max_adjustment_pct, updated_at = EXCLUDED.updated_at`,
		c.TenantID, c.Enabled, c.FrequencyHours, c.LookaheadDays, c.AutoAdjust, c.MaxAdjustmentPct, c.CreatedAt, c.UpdatedAt)
	return err
}