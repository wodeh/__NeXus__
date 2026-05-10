-- Migration: Tax Engine — multi-jurisdiction tax rates + exemptions
-- Version: 000008

CREATE TABLE IF NOT EXISTS tax_rates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    jurisdiction VARCHAR(100) NOT NULL,
    country_code CHAR(2) NOT NULL,
    state_code VARCHAR(10),
    city_code VARCHAR(10),
    rate DECIMAL(5,4) NOT NULL DEFAULT 0,
    type VARCHAR(50) NOT NULL DEFAULT 'vat', -- vat, gst, sales_tax, tourism_tax, resort_fee
    applies_to VARCHAR(50) NOT NULL DEFAULT 'all', -- room, fb, service, all
    is_compound BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    effective_from DATE NOT NULL DEFAULT CURRENT_DATE,
    effective_to DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tax_exemptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    guest_id UUID REFERENCES guests(id) ON DELETE SET NULL,
    corporate_id UUID,
    certificate_num VARCHAR(100),
    jurisdiction VARCHAR(100) NOT NULL,
    tax_type VARCHAR(50) NOT NULL,
    exempt_percent DECIMAL(5,2) NOT NULL DEFAULT 0,
    valid_from DATE NOT NULL,
    valid_to DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_tax_rates_tenant ON tax_rates(tenant_id);
CREATE INDEX idx_tax_rates_jurisdiction ON tax_rates(tenant_id, country_code, state_code, city_code);
CREATE INDEX idx_tax_rates_active ON tax_rates(tenant_id, is_active);
CREATE INDEX idx_tax_exemptions_tenant ON tax_exemptions(tenant_id);
CREATE INDEX idx_tax_exemptions_guest ON tax_exemptions(guest_id);

-- Triggers
CREATE TRIGGER update_tax_rates_updated_at
    BEFORE UPDATE ON tax_rates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- RLS
ALTER TABLE tax_rates ENABLE ROW LEVEL SECURITY;
ALTER TABLE tax_exemptions ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_tax_rates ON tax_rates
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY tenant_isolation_tax_exemptions ON tax_exemptions
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);

-- Seed default US sales tax
INSERT INTO tax_rates (tenant_id, name, jurisdiction, country_code, state_code, rate, type, applies_to)
SELECT id, 'US Federal Sales Tax', 'United States', 'US', NULL, 0.00, 'sales_tax', 'all'
FROM tenants LIMIT 1;
