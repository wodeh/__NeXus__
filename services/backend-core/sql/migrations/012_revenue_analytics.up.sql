-- Migration: Revenue analytics tables
-- Created: 2024-05-12

CREATE TABLE IF NOT EXISTS pricing_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    room_type VARCHAR(50) NOT NULL,
    condition VARCHAR(50) NOT NULL, -- occupancy_based, advance_booking, length_of_stay, seasonal
    trigger_value NUMERIC(10,2) DEFAULT 0,
    adjustment_type VARCHAR(20) NOT NULL, -- percentage, fixed_amount
    adjustment_value NUMERIC(10,2) DEFAULT 0,
    min_rate NUMERIC(10,2) DEFAULT 0,
    max_rate NUMERIC(10,2) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    priority INT DEFAULT 0,
    config JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_pricing_rules_tenant ON pricing_rules(tenant_id);
CREATE INDEX idx_pricing_rules_active ON pricing_rules(tenant_id, is_active) WHERE deleted_at IS NULL;

-- Trigger
CREATE OR REPLACE FUNCTION update_pricing_rules_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_pricing_rules_updated_at ON pricing_rules;
CREATE TRIGGER trg_pricing_rules_updated_at
    BEFORE UPDATE ON pricing_rules
    FOR EACH ROW EXECUTE FUNCTION update_pricing_rules_updated_at();

-- RLS
ALTER TABLE pricing_rules ENABLE ROW LEVEL SECURITY;

CREATE POLICY pricing_rules_tenant_isolation ON pricing_rules
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));