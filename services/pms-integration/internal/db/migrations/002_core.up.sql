CREATE TABLE IF NOT EXISTS guests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    loyalty_member_id VARCHAR(255),
    preferences JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    version INT NOT NULL DEFAULT 1
);

CREATE INDEX idx_guests_tenant ON guests(tenant_id);
CREATE INDEX idx_guests_email ON guests(email) WHERE email IS NOT NULL AND deleted_at IS NULL;

INSERT INTO guests (tenant_id, property_id, first_name, last_name, email, phone, preferences)
SELECT 'demo', p.id, 'John', 'Smith', 'john.smith@email.com', '+1-555-0101', '{"room_pref":"high_floor","smoking":false}'::jsonb
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

INSERT INTO guests (tenant_id, property_id, first_name, last_name, email, phone, preferences)
SELECT 'demo', p.id, 'Sarah', 'Johnson', 'sarah.j@email.com', '+1-555-0102', '{"room_pref":"quiet","dietary":"vegetarian"}'::jsonb
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    guest_id UUID NOT NULL,
    room_id UUID,
    check_in_date DATE NOT NULL,
    check_out_date DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'confirmed',
    special_requests TEXT[] NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    version INT NOT NULL DEFAULT 1
);

CREATE INDEX idx_reservations_tenant ON reservations(tenant_id);
CREATE INDEX idx_reservations_guest ON reservations(guest_id);
CREATE INDEX idx_reservations_dates ON reservations(check_in_date, check_out_date);
CREATE INDEX idx_reservations_status ON reservations(status);
CREATE INDEX idx_reservations_room ON reservations(room_id);

INSERT INTO reservations (tenant_id, property_id, guest_id, room_id, check_in_date, check_out_date, status, special_requests)
SELECT 'demo', p.id, g.id, r.id, CURRENT_DATE - INTERVAL '2 days', CURRENT_DATE + INTERVAL '2 days', 'checked_in', ARRAY['Late arrival, please leave key at reception']
FROM properties p, guests g, rooms r
WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel' AND g.email = 'john.smith@email.com' AND r.room_number = '101'
ON CONFLICT DO NOTHING;

INSERT INTO reservations (tenant_id, property_id, guest_id, room_id, check_in_date, check_out_date, status, special_requests)
SELECT 'demo', p.id, g.id, r.id, CURRENT_DATE + INTERVAL '1 day', CURRENT_DATE + INTERVAL '4 days', 'confirmed', ARRAY['Extra bed for child']
FROM properties p, guests g, rooms r
WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel' AND g.email = 'sarah.j@email.com' AND r.room_number = '201'
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS folios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    reservation_id UUID NOT NULL,
    guest_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    balance DECIMAL(12,2) NOT NULL DEFAULT 0,
    currency_code VARCHAR(3) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    version INT NOT NULL DEFAULT 1
);

CREATE INDEX idx_folios_tenant ON folios(tenant_id);
CREATE INDEX idx_folios_reservation ON folios(reservation_id);
CREATE INDEX idx_folios_status ON folios(status);

INSERT INTO folios (tenant_id, property_id, reservation_id, guest_id, status, balance, currency_code)
SELECT 'demo', p.id, r.id, g.id, 'open', 516.00, 'USD'
FROM properties p, reservations r, guests g
WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel' AND g.email = 'john.smith@email.com' AND r.guest_id = g.id AND r.status = 'checked_in'
ON CONFLICT DO NOTHING;

INSERT INTO folios (tenant_id, property_id, reservation_id, guest_id, status, balance, currency_code)
SELECT 'demo', p.id, r.id, g.id, 'open', 567.00, 'USD'
FROM properties p, reservations r, guests g
WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel' AND g.email = 'sarah.j@email.com' AND r.guest_id = g.id AND r.status = 'confirmed'
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    user_id UUID,
    guest_id UUID,
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    resource_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_tenant ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource, resource_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);

INSERT INTO audit_logs (tenant_id, action, resource, resource_id, new_values)
SELECT 'demo', 'property_created', 'property', p.id, jsonb_build_object('name', p.name)
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel';

INSERT INTO audit_logs (tenant_id, action, resource, resource_id, new_values)
SELECT 'demo', 'guest_created', 'guest', g.id, jsonb_build_object('name', g.first_name || ' ' || g.last_name)
FROM guests g WHERE g.tenant_id = 'demo' AND g.email = 'john.smith@email.com';

DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY['guests', 'reservations', 'folios'];
BEGIN
    FOREACH tbl IN ARRAY tables
    LOOP
        IF NOT EXISTS (
            SELECT 1 FROM pg_trigger
            WHERE tgname = 'update_' || tbl || '_updated_at'
        ) THEN
            EXECUTE format('
                CREATE TRIGGER update_%I_updated_at
                BEFORE UPDATE ON %I
                FOR EACH ROW
                EXECUTE FUNCTION update_updated_at_column();
            ', tbl, tbl);
        END IF;
    END LOOP;
END $$;
