-- Migration: Revenue Management ML + Dynamic Pricing
-- Version: 000010

CREATE TABLE IF NOT EXISTS revenue_forecasts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    predicted_occupancy DECIMAL(5,4) NOT NULL DEFAULT 0,
    predicted_adr DECIMAL(10,2) NOT NULL DEFAULT 0,
    predicted_revPAR DECIMAL(10,2) NOT NULL DEFAULT 0,
    confidence DECIMAL(5,4) NOT NULL DEFAULT 0,
    model_version VARCHAR(20) NOT NULL DEFAULT 'v1.0',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS dynamic_pricing_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    room_type_id UUID REFERENCES room_types(id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    min_advance_days INT NOT NULL DEFAULT 0,
    max_advance_days INT,
    min_los INT NOT NULL DEFAULT 1,
    lead_time_discount DECIMAL(5,2) NOT NULL DEFAULT 0,
    last_minute_premium DECIMAL(5,2) NOT NULL DEFAULT 0,
    occupancy_threshold DECIMAL(5,4) NOT NULL DEFAULT 0.8,
    occupancy_premium DECIMAL(5,2) NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS demand_signals (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    local_events JSONB,
    competitor_rates JSONB,
    flight_search_volume INT,
    weather_forecast VARCHAR(50),
    demand_score DECIMAL(5,4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS price_recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    room_type_id UUID REFERENCES room_types(id) ON DELETE SET NULL,
    date DATE NOT NULL,
    current_rate DECIMAL(10,2) NOT NULL DEFAULT 0,
    suggested_rate DECIMAL(10,2) NOT NULL DEFAULT 0,
    change_percent DECIMAL(5,2) NOT NULL DEFAULT 0,
    reason TEXT,
    confidence DECIMAL(5,4) NOT NULL DEFAULT 0,
    applied BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_revenue_forecasts_property ON revenue_forecasts(tenant_id, property_id, date);
CREATE INDEX idx_pricing_rules_tenant ON dynamic_pricing_rules(tenant_id);
CREATE INDEX idx_demand_signals_property ON demand_signals(tenant_id, property_id, date);
CREATE INDEX idx_price_recommendations_pending ON price_recommendations(tenant_id, applied) WHERE applied = FALSE;

-- Triggers
CREATE TRIGGER update_pricing_rules_updated_at
    BEFORE UPDATE ON dynamic_pricing_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- RLS
ALTER TABLE revenue_forecasts ENABLE ROW LEVEL SECURITY;
ALTER TABLE dynamic_pricing_rules ENABLE ROW LEVEL SECURITY;
ALTER TABLE demand_signals ENABLE ROW LEVEL SECURITY;
ALTER TABLE price_recommendations ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_forecasts ON revenue_forecasts
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY tenant_isolation_pricing_rules ON dynamic_pricing_rules
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY tenant_isolation_demand_signals ON demand_signals
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY tenant_isolation_price_recommendations ON price_recommendations
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
