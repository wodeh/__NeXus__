-- Migration: Channel Manager — OTA integrations
-- Version: 000005

CREATE TABLE IF NOT EXISTS channel_integrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    channel_type VARCHAR(30) NOT NULL,
    name VARCHAR(100) NOT NULL,
    api_key TEXT,
    api_secret TEXT,
    hotel_id VARCHAR(100),
    rate_plan_map JSONB DEFAULT '{}',
    room_type_map JSONB DEFAULT '{}',
    commission_pct DECIMAL(5,2) NOT NULL DEFAULT 15.00,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    status VARCHAR(30) NOT NULL DEFAULT 'active',
    last_sync_at TIMESTAMPTZ,
    last_error TEXT,
    last_error_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS channel_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    channel_integration_id UUID NOT NULL REFERENCES channel_integrations(id) ON DELETE CASCADE,
    ota_confirmation_number VARCHAR(100) NOT NULL,
    reservation_id UUID REFERENCES reservations(id) ON DELETE SET NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    guest_name VARCHAR(200),
    guest_email VARCHAR(255),
    guest_phone VARCHAR(50),
    room_type_id UUID REFERENCES room_types(id) ON DELETE SET NULL,
    rate_plan_id UUID REFERENCES rate_plans(id) ON DELETE SET NULL,
    check_in_date DATE NOT NULL,
    check_out_date DATE NOT NULL,
    adults INTEGER NOT NULL DEFAULT 1,
    children INTEGER NOT NULL DEFAULT 0,
    total_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    currency_code VARCHAR(3) DEFAULT 'USD',
    commission_amount DECIMAL(12,2) NOT NULL DEFAULT 0,
    raw_data TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS availability_push_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    channel_integration_id UUID NOT NULL REFERENCES channel_integrations(id) ON DELETE CASCADE,
    room_type_id UUID NOT NULL REFERENCES room_types(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    available_rooms INTEGER NOT NULL,
    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_channel_integrations_tenant ON channel_integrations(tenant_id);
CREATE INDEX idx_channel_integrations_property ON channel_integrations(property_id);
CREATE INDEX idx_channel_reservations_tenant ON channel_reservations(tenant_id);
CREATE INDEX idx_channel_reservations_status ON channel_reservations(status);
CREATE INDEX idx_availability_push_logs_tenant ON availability_push_logs(tenant_id);

-- Triggers
CREATE TRIGGER update_channel_integrations_updated_at
    BEFORE UPDATE ON channel_integrations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_channel_reservations_updated_at
    BEFORE UPDATE ON channel_reservations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- RLS
ALTER TABLE channel_integrations ENABLE ROW LEVEL SECURITY;
ALTER TABLE channel_reservations ENABLE ROW LEVEL SECURITY;
ALTER TABLE availability_push_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_channel_integrations ON channel_integrations
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY tenant_isolation_channel_reservations ON channel_reservations
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY tenant_isolation_availability_push_logs ON availability_push_logs
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
