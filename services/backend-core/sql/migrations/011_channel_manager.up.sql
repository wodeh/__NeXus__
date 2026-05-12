-- Migration: Channel Manager tables
-- Created: 2024-05-12

CREATE TABLE IF NOT EXISTS channel_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    channel_source VARCHAR(50) NOT NULL, -- booking_com, expedia, airbnb, email, whatsapp, walk_in
    external_ref VARCHAR(255),
    guest_name VARCHAR(255) NOT NULL,
    guest_email VARCHAR(255),
    guest_phone VARCHAR(50),
    room_type VARCHAR(50) NOT NULL,
    room_number VARCHAR(10),
    check_in DATE NOT NULL,
    check_out DATE NOT NULL,
    adults INT DEFAULT 1,
    children INT DEFAULT 0,
    total NUMERIC(10,2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'pending', -- pending, confirmed, cancelled, no_show, checked_in, checked_out
    special_requests TEXT,
    raw_payload JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_channel_res_tenant ON channel_reservations(tenant_id);
CREATE INDEX idx_channel_res_source ON channel_reservations(channel_source);
CREATE INDEX idx_channel_res_status ON channel_reservations(status);
CREATE INDEX idx_channel_res_dates ON channel_reservations(check_in, check_out);
CREATE INDEX idx_channel_res_external ON channel_reservations(external_ref);

CREATE TABLE IF NOT EXISTS channel_availability (
    tenant_id UUID NOT NULL,
    room_type VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    total_rooms INT DEFAULT 0,
    booked_rooms INT DEFAULT 0,
    blocked_rooms INT DEFAULT 0,
    available INT DEFAULT 0,
    rate NUMERIC(10,2) DEFAULT 0,
    currency VARCHAR(3) DEFAULT 'USD',
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (tenant_id, room_type, date)
);

CREATE INDEX idx_avail_tenant ON channel_availability(tenant_id);
CREATE INDEX idx_avail_date ON channel_availability(date);

-- Triggers
CREATE OR REPLACE FUNCTION update_channel_reservations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_channel_reservations_updated_at ON channel_reservations;
CREATE TRIGGER trg_channel_reservations_updated_at
    BEFORE UPDATE ON channel_reservations
    FOR EACH ROW EXECUTE FUNCTION update_channel_reservations_updated_at();

-- RLS
ALTER TABLE channel_reservations ENABLE ROW LEVEL SECURITY;
ALTER TABLE channel_availability ENABLE ROW LEVEL SECURITY;

CREATE POLICY channel_res_tenant_isolation ON channel_reservations
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));

CREATE POLICY availability_tenant_isolation ON channel_availability
    USING (tenant_id::TEXT = current_setting('app.current_tenant', true));
