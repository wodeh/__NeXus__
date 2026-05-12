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
    elevation NUMERIC(8, 2), -- meters above sea level (baseline for floor detection)
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
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
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
CREATE TRIGGER update_villa_props_updated_at BEFORE UPDATE ON villa_properties FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_villa_res_updated_at BEFORE UPDATE ON villa_reservations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- RLS policies
ALTER TABLE villa_properties ENABLE ROW LEVEL SECURITY;
ALTER TABLE villa_reservations ENABLE ROW LEVEL SECURITY;
ALTER TABLE cleaner_sensor_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_villa_props ON villa_properties FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_villa_res ON villa_reservations FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY tenant_isolation_sensor_logs ON cleaner_sensor_logs FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);

-- Seed demo villa data
INSERT INTO villa_properties (id, tenant_id, name, description, address, city, country, latitude, longitude, elevation, bedrooms, bathrooms, max_guests, amenities, price_per_night, currency, cleaning_fee, security_deposit, is_active, status)
VALUES
    ('a0000001-0000-0000-0000-000000000001', 'b0000001-0000-0000-0000-000000000001', 'Villa Al-Mashta', 'Luxury villa with pool and mountain view', 'Al-Mashta Street 15', 'Ramallah', 'Palestine', 31.8996, 35.2042, 850.00, 3, 2, 6, ARRAY['pool', 'wifi', 'parking', 'bbq'], 350.00, 'USD', 50.00, 500.00, true, 'available'),
    ('a0000002-0000-0000-0000-000000000001', 'b0000001-0000-0000-0000-000000000001', 'Chalet Al-Balad', 'Traditional chalet in Bethlehem old city', 'Star Street 8', 'Bethlehem', 'Palestine', 31.7043, 35.2075, 775.00, 4, 3, 8, ARRAY['fireplace', 'wifi', 'kitchen', 'garden'], 450.00, 'USD', 75.00, 750.00, true, 'available')
ON CONFLICT (id) DO NOTHING;

-- Seed demo reservations
INSERT INTO villa_reservations (id, tenant_id, villa_id, villa_name, guest_name, guest_phone, guest_email, guest_count, check_in_date, check_out_date, nights, total_amount, currency, status, source, down_payment, balance_due, balance_paid)
VALUES
    ('b0000001-0000-0000-0000-000000000001', 'c0000001-0000-0000-0000-000000000001', 'a0000002-0000-0000-0000-000000000001', 'Chalet Al-Balad', 'Ahmad Khalil', '+970599123456', 'ahmad@email.com', 4, '2024-08-15', '2024-08-20', 5, 2500.00, 'USD', 'reserved', 'whatsapp', '{"amount": 500, "method": "visa", "status": "received", "received_at": "2024-08-10T10:00:00Z", "reference": "visa-001"}'::jsonb, 2000.00, false),
    ('b0000002-0000-0000-0000-000000000001', 'c0000001-0000-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001', 'Villa Al-Mashta', 'Sarah Nassar', '+970599987654', 'sarah@email.com', 2, '2024-09-01', '2024-09-05', 4, 1400.00, 'USD', 'pending', 'phone', NULL, 1400.00, false)
ON CONFLICT (id) DO NOTHING;

-- Seed demo sensor logs
INSERT INTO cleaner_sensor_logs (id, tenant_id, villa_id, villa_name, cleaner_id, cleaner_name, temperature, latitude, longitude, altitude, floor, location_type, battery_level, recorded_at)
VALUES
    ('d0000001-0000-0000-0000-000000000001', 'e0000001-0000-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001', 'Villa Al-Mashta', 'cleaner-1', 'Muhammad', 31.50, 31.8996, 35.2042, 850.00, 0, 'outside', 78.00, NOW() - INTERVAL '30 minutes'),
    ('d0000002-0000-0000-0000-000000000001', 'e0000001-0000-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001', 'Villa Al-Mashta', 'cleaner-1', 'Muhammad', 24.20, 31.8996, 35.2042, 854.00, 1, 'inside', 76.00, NOW() - INTERVAL '15 minutes'),
    ('d0000003-0000-0000-0000-000000000001', 'e0000001-0000-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001', 'Villa Al-Mashta', 'cleaner-1', 'Muhammad', 22.80, 31.8996, 35.2042, 858.00, 2, 'inside', 74.00, NOW() - INTERVAL '5 minutes')
ON CONFLICT (id) DO NOTHING;
