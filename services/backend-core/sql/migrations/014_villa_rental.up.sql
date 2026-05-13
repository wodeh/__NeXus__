-- Migration 014: Villa Rental / Chalet Management
-- Supports private villa rentals with day-based bookings, down payments, and cleaner sensors

-- Villa Properties (bookable units)
CREATE TABLE IF NOT EXISTS villa_properties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    address TEXT,
    city VARCHAR(100),
    country VARCHAR(100),
    latitude NUMERIC(10, 8),
    longitude NUMERIC(11, 8),
    elevation NUMERIC(8, 2),
    bedrooms INTEGER NOT NULL DEFAULT 1,
    bathrooms INTEGER NOT NULL DEFAULT 1,
    max_guests INTEGER NOT NULL DEFAULT 2,
    amenities TEXT[] DEFAULT '{}',
    images TEXT[] DEFAULT '{}',
    price_per_night NUMERIC(10, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    cleaning_fee NUMERIC(10, 2) NOT NULL DEFAULT 0,
    security_deposit NUMERIC(10, 2) NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    status VARCHAR(20) NOT NULL DEFAULT 'available' CHECK (status IN ('available', 'maintenance', 'blocked')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

-- Villa Reservations (day-based bookings)
CREATE TABLE IF NOT EXISTS villa_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    villa_id UUID NOT NULL REFERENCES villa_properties(id) ON DELETE CASCADE,
    villa_name VARCHAR(200) NOT NULL,
    guest_name VARCHAR(200) NOT NULL,
    guest_phone VARCHAR(30),
    guest_email VARCHAR(200),
    guest_count INTEGER NOT NULL DEFAULT 1,
    check_in_date DATE NOT NULL,
    check_out_date DATE NOT NULL,
    nights INTEGER NOT NULL DEFAULT 1,
    total_amount NUMERIC(10, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'reserved', 'cancelled', 'completed')),
    source VARCHAR(20) NOT NULL DEFAULT 'phone' CHECK (source IN ('phone', 'whatsapp', 'walkin', 'website')),
    internal_notes TEXT,
    down_payment JSONB,
    balance_due NUMERIC(10, 2) NOT NULL DEFAULT 0,
    balance_paid BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ
);

-- Cleaner Sensor Logs (phone sensor tracking)
CREATE TABLE IF NOT EXISTS cleaner_sensor_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    villa_id UUID NOT NULL REFERENCES villa_properties(id) ON DELETE CASCADE,
    villa_name VARCHAR(200) NOT NULL,
    cleaner_id VARCHAR(100) NOT NULL,
    cleaner_name VARCHAR(200),
    temperature NUMERIC(5, 2), -- °C
    latitude NUMERIC(10, 8),
    longitude NUMERIC(11, 8),
    altitude NUMERIC(8, 2), -- meters above sea level
    floor INTEGER NOT NULL DEFAULT 0, -- derived from altitude - property elevation
    location_type VARCHAR(20) NOT NULL DEFAULT 'unknown' CHECK (location_type IN ('inside', 'outside', 'balcony', 'rooftop', 'unknown')),
    battery_level NUMERIC(5, 2), -- 0-100
    recorded_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_villa_props_tenant ON villa_properties(tenant_id);
CREATE INDEX IF NOT EXISTS idx_villa_props_status ON villa_properties(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_villa_res_tenant ON villa_reservations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_villa_res_villa ON villa_reservations(villa_id);
CREATE INDEX IF NOT EXISTS idx_villa_res_dates ON villa_reservations(villa_id, check_in_date, check_out_date);
CREATE INDEX IF NOT EXISTS idx_villa_res_status ON villa_reservations(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_sensor_logs_villa ON cleaner_sensor_logs(villa_id);
CREATE INDEX IF NOT EXISTS idx_sensor_logs_cleaner ON cleaner_sensor_logs(villa_id, cleaner_id);
CREATE INDEX IF NOT EXISTS idx_sensor_logs_time ON cleaner_sensor_logs(villa_id, recorded_at DESC);

-- Trigger for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_villa_props_updated_at BEFORE UPDATE ON villa_properties FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_villa_res_updated_at BEFORE UPDATE ON villa_reservations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- RLS policies
ALTER TABLE villa_properties ENABLE ROW LEVEL SECURITY;
ALTER TABLE villa_reservations ENABLE ROW LEVEL SECURITY;
ALTER TABLE cleaner_sensor_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_villa_props ON villa_properties FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_villa_res ON villa_reservations FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_sensor_logs ON cleaner_sensor_logs FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
