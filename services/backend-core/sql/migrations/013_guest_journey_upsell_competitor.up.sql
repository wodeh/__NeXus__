-- Migration 013: Guest Journey, Upsell, Competitor Rate Shopping
-- Enables AI-powered guest automation, upsell engine, and competitor rate intelligence

-- Guest Journeys (automation sequences)
CREATE TABLE IF NOT EXISTS guest_journeys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    trigger VARCHAR(50) NOT NULL CHECK (trigger IN ('booking_confirmed', 'check_in', 'check_out', 'no_show', 'cancellation')),
    is_active BOOLEAN NOT NULL DEFAULT true,
    steps JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Guest Journey Executions (track per-reservation progress)
CREATE TABLE IF NOT EXISTS guest_journey_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    journey_id UUID NOT NULL REFERENCES guest_journeys(id) ON DELETE CASCADE,
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    guest_phone VARCHAR(30),
    current_step INTEGER NOT NULL DEFAULT 0,
    total_steps INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'running' CHECK (status IN ('running', 'completed', 'cancelled', 'failed')),
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    next_trigger_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Upsell Offers (sellable add-ons)
CREATE TABLE IF NOT EXISTS upsell_offers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(50) NOT NULL CHECK (category IN ('room_upgrade', 'late_checkout', 'early_checkin', 'breakfast', 'spa', 'parking', 'other')),
    price NUMERIC(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    image_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    auto_offer BOOLEAN NOT NULL DEFAULT false,
    conditions JSONB DEFAULT '{}',
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Upsell Purchases (guest transactions)
CREATE TABLE IF NOT EXISTS upsell_purchases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    guest_phone VARCHAR(30),
    offer_id UUID NOT NULL REFERENCES upsell_offers(id) ON DELETE CASCADE,
    offer_name VARCHAR(200) NOT NULL,
    price NUMERIC(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'cancelled', 'refunded')),
    payment_method VARCHAR(20) NOT NULL DEFAULT 'on_bill' CHECK (payment_method IN ('on_bill', 'credit_card', 'whatsapp_pay')),
    folio_posted BOOLEAN NOT NULL DEFAULT false,
    journey_step_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Competitor Hotels (rate shopping targets)
CREATE TABLE IF NOT EXISTS competitor_hotels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100),
    star_rating INTEGER CHECK (star_rating BETWEEN 1 AND 5),
    room_count INTEGER,
    website TEXT,
    booking_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_scraped TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Competitor Rates (scraped pricing data)
CREATE TABLE IF NOT EXISTS competitor_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    competitor_id UUID NOT NULL REFERENCES competitor_hotels(id) ON DELETE CASCADE,
    competitor_name VARCHAR(200) NOT NULL,
    room_type VARCHAR(100) NOT NULL,
    date DATE NOT NULL,
    rate NUMERIC(10,2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    availability INTEGER,
    min_stay INTEGER,
    is_promo BOOLEAN NOT NULL DEFAULT false,
    source VARCHAR(50) NOT NULL,
    scraped_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Rate Recommendations (AI-generated pricing suggestions)
CREATE TABLE IF NOT EXISTS rate_recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    room_type VARCHAR(100) NOT NULL,
    date DATE NOT NULL,
    current_rate NUMERIC(10,2) NOT NULL,
    recommended_rate NUMERIC(10,2) NOT NULL,
    confidence NUMERIC(3,2) NOT NULL CHECK (confidence BETWEEN 0 AND 1),
    reason TEXT NOT NULL,
    factors JSONB NOT NULL DEFAULT '[]',
    applied BOOLEAN NOT NULL DEFAULT false,
    applied_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Rate Shop Configuration (per-tenant settings)
CREATE TABLE IF NOT EXISTS rate_shop_configs (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT true,
    frequency_hours INTEGER NOT NULL DEFAULT 6,
    lookahead_days INTEGER NOT NULL DEFAULT 30,
    auto_adjust BOOLEAN NOT NULL DEFAULT false,
    max_adjustment_pct NUMERIC(4,2) NOT NULL DEFAULT 0.30,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_journeys_tenant ON guest_journeys(tenant_id);
CREATE INDEX IF NOT EXISTS idx_journey_execs_tenant ON guest_journey_executions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_journey_execs_reservation ON guest_journey_executions(reservation_id);
CREATE INDEX IF NOT EXISTS idx_upsell_offers_tenant ON upsell_offers(tenant_id);
CREATE INDEX IF NOT EXISTS idx_upsell_offers_category ON upsell_offers(tenant_id, category);
CREATE INDEX IF NOT EXISTS idx_upsell_purchases_tenant ON upsell_purchases(tenant_id);
CREATE INDEX IF NOT EXISTS idx_upsell_purchases_reservation ON upsell_purchases(reservation_id);
CREATE INDEX IF NOT EXISTS idx_competitors_tenant ON competitor_hotels(tenant_id);
CREATE INDEX IF NOT EXISTS idx_competitor_rates_lookup ON competitor_rates(tenant_id, competitor_id, room_type, date);
CREATE INDEX IF NOT EXISTS idx_rate_recs_pending ON rate_recommendations(tenant_id, applied, date);

-- Triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_guest_journeys_updated_at BEFORE UPDATE ON guest_journeys FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_journey_execs_updated_at BEFORE UPDATE ON guest_journey_executions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_upsell_offers_updated_at BEFORE UPDATE ON upsell_offers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_upsell_purchases_updated_at BEFORE UPDATE ON upsell_purchases FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_competitor_hotels_updated_at BEFORE UPDATE ON competitor_hotels FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_rate_shop_configs_updated_at BEFORE UPDATE ON rate_shop_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- RLS policies
ALTER TABLE guest_journeys ENABLE ROW LEVEL SECURITY;
ALTER TABLE guest_journey_executions ENABLE ROW LEVEL SECURITY;
ALTER TABLE upsell_offers ENABLE ROW LEVEL SECURITY;
ALTER TABLE upsell_purchases ENABLE ROW LEVEL SECURITY;
ALTER TABLE competitor_hotels ENABLE ROW LEVEL SECURITY;
ALTER TABLE competitor_rates ENABLE ROW LEVEL SECURITY;
ALTER TABLE rate_recommendations ENABLE ROW LEVEL SECURITY;
ALTER TABLE rate_shop_configs ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_guest_journeys ON guest_journeys FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_journey_execs ON guest_journey_executions FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_upsell_offers ON upsell_offers FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_upsell_purchases ON upsell_purchases FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_competitors ON competitor_hotels FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_competitor_rates ON competitor_rates FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_rate_recs ON rate_recommendations FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_rate_shop_config ON rate_shop_configs FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
