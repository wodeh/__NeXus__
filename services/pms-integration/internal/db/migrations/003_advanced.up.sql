CREATE TABLE IF NOT EXISTS room_blocks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    block_type VARCHAR(50) NOT NULL,
    reason TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    affected_rooms JSONB NOT NULL DEFAULT '[]',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_room_blocks_tenant ON room_blocks(tenant_id);
CREATE INDEX idx_room_blocks_dates ON room_blocks(start_date, end_date);
CREATE INDEX idx_room_blocks_status ON room_blocks(status);

INSERT INTO room_blocks (tenant_id, property_id, name, block_type, reason, start_date, end_date, status, affected_rooms)
SELECT 'demo', p.id, 'Maintenance Block Q2', 'maintenance', 'HVAC system upgrade', CURRENT_DATE, CURRENT_DATE + INTERVAL '5 days', 'active', '["301","302","303"]'::jsonb
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

INSERT INTO room_blocks (tenant_id, property_id, name, block_type, reason, start_date, end_date, status, affected_rooms)
SELECT 'demo', p.id, 'VIP Event Hold', 'vip', 'Celebrity booking', CURRENT_DATE + INTERVAL '7 days', CURRENT_DATE + INTERVAL '10 days', 'active', '["PRES-01"]'::jsonb
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS group_reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    group_name VARCHAR(255) NOT NULL,
    contact_name VARCHAR(255),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'tentative',
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    num_rooms INT NOT NULL DEFAULT 1,
    room_type_ids JSONB NOT NULL DEFAULT '[]',
    rate_agreement DECIMAL(10,2),
    total_value DECIMAL(12,2) NOT NULL DEFAULT 0,
    currency_code VARCHAR(3) NOT NULL DEFAULT 'USD',
    notes TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_group_reservations_tenant ON group_reservations(tenant_id);
CREATE INDEX idx_group_reservations_dates ON group_reservations(start_date, end_date);
CREATE INDEX idx_group_reservations_status ON group_reservations(status);

INSERT INTO group_reservations (tenant_id, property_id, group_name, contact_name, contact_email, status, start_date, end_date, num_rooms, rate_agreement, total_value, currency_code, notes)
SELECT 'demo', p.id, 'Tech Conference 2026', 'Alice Wang', 'alice@techconf.com', 'definite', CURRENT_DATE + INTERVAL '30 days', CURRENT_DATE + INTERVAL '33 days', 15, 115.00, 5175.00, 'USD', '15 standard rooms, group breakfast included'
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS guest_profiles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    guest_id UUID NOT NULL,
    segment VARCHAR(50),
    lifetime_value DECIMAL(12,2) NOT NULL DEFAULT 0,
    satisfaction_score INT,
    last_stay_date DATE,
    next_predicted_stay DATE,
    communications JSONB NOT NULL DEFAULT '[]',
    preferences JSONB NOT NULL DEFAULT '{}',
    tags JSONB NOT NULL DEFAULT '[]',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_guest_profiles_tenant ON guest_profiles(tenant_id);
CREATE INDEX idx_guest_profiles_guest ON guest_profiles(guest_id);

INSERT INTO guest_profiles (tenant_id, guest_id, segment, lifetime_value, satisfaction_score, last_stay_date, next_predicted_stay, tags, preferences)
SELECT 'demo', g.id, 'business', 8765.50, 9, CURRENT_DATE - INTERVAL '30 days', CURRENT_DATE + INTERVAL '45 days', '["VIP","business_traveler"]'::jsonb, '{"room_pref":"high_floor","dietary":"none"}'::jsonb
FROM guests g WHERE g.tenant_id = 'demo' AND g.email = 'john.smith@email.com'
ON CONFLICT DO NOTHING;

INSERT INTO guest_profiles (tenant_id, guest_id, segment, lifetime_value, satisfaction_score, last_stay_date, next_predicted_stay, tags, preferences)
SELECT 'demo', g.id, 'leisure', 3200.00, 8, CURRENT_DATE - INTERVAL '60 days', CURRENT_DATE + INTERVAL '90 days', '["family","returning"]'::jsonb, '{"room_pref":"quiet","dietary":"vegetarian"}'::jsonb
FROM guests g WHERE g.tenant_id = 'demo' AND g.email = 'sarah.j@email.com'
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    agent_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    commission_rate DECIMAL(5,2) NOT NULL DEFAULT 0,
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    api_key VARCHAR(255),
    webhook_url TEXT,
    settings JSONB NOT NULL DEFAULT '{}',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_agents_tenant ON agents(tenant_id);
CREATE INDEX idx_agents_status ON agents(status);
CREATE INDEX idx_agents_type ON agents(agent_type);

INSERT INTO agents (tenant_id, name, agent_type, status, commission_rate, contact_email, settings)
VALUES
    ('demo', 'Booking.com', 'ota', 'active', 15.00, 'partner@booking.com', '{"channel_manager":"mapped","auto_confirm":true}'::jsonb),
    ('demo', 'Expedia', 'ota', 'active', 18.00, 'partner@expedia.com', '{"channel_manager":"mapped","auto_confirm":false}'::jsonb),
    ('demo', 'Corporate Travel Inc', 'corporate', 'active', 10.00, 'reservations@corptravel.com', '{"contract":"annual","payment_terms":"net30"}'::jsonb),
    ('demo', 'Airbnb', 'ota', 'pending', 3.00, 'partner@airbnb.com', '{"channel_manager":"pending","auto_confirm":false}'::jsonb)
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS rate_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    description TEXT,
    seasonal_rates JSONB NOT NULL DEFAULT '[]',
    restrictions JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, code)
);

CREATE INDEX idx_rate_plans_tenant ON rate_plans(tenant_id);
CREATE INDEX idx_rate_plans_property ON rate_plans(property_id);

INSERT INTO rate_plans (tenant_id, property_id, name, code, description, seasonal_rates, restrictions, is_active)
SELECT 'demo', p.id, 'Summer 2026 Standard', 'SUM-STD', 'Standard room summer rate',
    '[{"room_type_id":"rt-001","start_date":"2026-06-01","end_date":"2026-08-31","base_price":129.00},{"room_type_id":"rt-002","start_date":"2026-06-01","end_date":"2026-08-31","base_price":189.00}]'::jsonb,
    '{"min_stay":1,"max_stay":14}'::jsonb,
    true
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS checkins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    reservation_id UUID NOT NULL,
    guest_id UUID NOT NULL,
    room_id UUID,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    id_verified BOOLEAN NOT NULL DEFAULT false,
    payment_collected BOOLEAN NOT NULL DEFAULT false,
    key_issued BOOLEAN NOT NULL DEFAULT false,
    room_inspected BOOLEAN NOT NULL DEFAULT false,
    welcome_sent BOOLEAN NOT NULL DEFAULT false,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_checkins_tenant ON checkins(tenant_id);
CREATE INDEX idx_checkins_reservation ON checkins(reservation_id);
CREATE INDEX idx_checkins_status ON checkins(status);

CREATE TABLE IF NOT EXISTS checkouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    reservation_id UUID NOT NULL,
    guest_id UUID NOT NULL,
    room_id UUID,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    balance_settled BOOLEAN NOT NULL DEFAULT false,
    key_returned BOOLEAN NOT NULL DEFAULT false,
    room_inspected BOOLEAN NOT NULL DEFAULT false,
    feedback_sent BOOLEAN NOT NULL DEFAULT false,
    completed_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_checkouts_tenant ON checkouts(tenant_id);
CREATE INDEX idx_checkouts_reservation ON checkouts(reservation_id);
CREATE INDEX idx_checkouts_status ON checkouts(status);

CREATE TABLE IF NOT EXISTS invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    folio_id UUID NOT NULL,
    guest_id UUID NOT NULL,
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    line_items JSONB NOT NULL DEFAULT '[]',
    subtotal DECIMAL(12,2) NOT NULL DEFAULT 0,
    tax_total DECIMAL(12,2) NOT NULL DEFAULT 0,
    grand_total DECIMAL(12,2) NOT NULL DEFAULT 0,
    currency_code VARCHAR(3) NOT NULL DEFAULT 'USD',
    issue_date DATE NOT NULL DEFAULT CURRENT_DATE,
    due_date DATE,
    paid_date DATE,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_invoices_tenant ON invoices(tenant_id);
CREATE INDEX idx_invoices_folio ON invoices(folio_id);
CREATE INDEX idx_invoices_status ON invoices(status);

INSERT INTO invoices (tenant_id, folio_id, guest_id, invoice_number, status, line_items, subtotal, tax_total, grand_total, currency_code, issue_date, due_date, notes)
SELECT 'demo', f.id, g.id, 'INV-000001', 'draft',
    '[{"description":"Room 101 - 4 nights","quantity":4,"unit_price":129.00,"tax_rate":14.0,"total":516.00},{"description":"Breakfast buffet","quantity":4,"unit_price":25.00,"tax_rate":14.0,"total":100.00}]'::jsonb,
    616.00, 86.24, 702.24, 'USD',
    CURRENT_DATE, CURRENT_DATE + INTERVAL '7 days',
    'Please settle by due date'
FROM folios f, guests g, reservations r
WHERE g.email = 'john.smith@email.com' AND r.guest_id = g.id AND r.status = 'checked_in' AND f.reservation_id = r.id
ON CONFLICT DO NOTHING;

CREATE TABLE IF NOT EXISTS gdpr_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    guest_id UUID NOT NULL,
    request_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    requested_at TIMESTAMPTZ DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    data_export JSONB,
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_gdpr_tenant ON gdpr_requests(tenant_id);
CREATE INDEX idx_gdpr_guest ON gdpr_requests(guest_id);
CREATE INDEX idx_gdpr_status ON gdpr_requests(status);
