CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ==================== TENANTS ====================
CREATE TABLE IF NOT EXISTS tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    license_tier VARCHAR(50) NOT NULL DEFAULT 'core',
    license_status VARCHAR(50) NOT NULL DEFAULT 'active',
    max_rooms INT NOT NULL DEFAULT 50,
    max_users INT NOT NULL DEFAULT 5,
    settings JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_tenants_slug ON tenants(slug);

INSERT INTO tenants (slug, name, license_tier, license_status, max_rooms, max_users, settings)
VALUES 
    ('demo', 'Demo Hotel', 'enterprise', 'active', 5000, 200, '{"timezone":"UTC","currency_code":"USD","language":"en","property_type":"boutique"}'::jsonb),
    ('grand-plaza', 'Grand Plaza', 'revenue', 'active', 500, 50, '{"timezone":"America/New_York","currency_code":"USD","language":"en","property_type":"hotel"}'::jsonb),
    ('sunset-inn', 'Sunset Inn', 'core', 'active', 50, 5, '{"timezone":"America/Los_Angeles","currency_code":"USD","language":"en","property_type":"motel"}'::jsonb)
ON CONFLICT (slug) DO NOTHING;

-- ==================== PROPERTIES ====================
CREATE TABLE IF NOT EXISTS properties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL,
    address JSONB,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    currency_code VARCHAR(3) NOT NULL DEFAULT 'USD',
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    version INT NOT NULL DEFAULT 1,
    UNIQUE(tenant_id, slug)
);

CREATE INDEX idx_properties_tenant ON properties(tenant_id);
CREATE INDEX idx_properties_active ON properties(is_active) WHERE deleted_at IS NULL;

INSERT INTO properties (tenant_id, name, slug, address, timezone, currency_code, contact_email, contact_phone, is_active)
VALUES 
    ('demo', 'Grand Plaza Hotel', 'grand-plaza-hotel', '{"street":"123 Main St","city":"New York","country":"USA","zip":"10001"}', 'America/New_York', 'USD', 'info@grandplaza.com', '+1-555-0100', true),
    ('demo', 'Sunset Inn', 'sunset-inn', '{"street":"456 Beach Rd","city":"Miami","country":"USA","zip":"33101"}', 'America/New_York', 'USD', 'info@sunsetinn.com', '+1-555-0200', true)
ON CONFLICT DO NOTHING;

-- ==================== ROOM TYPES ====================
CREATE TABLE IF NOT EXISTS room_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_occupancy INT NOT NULL DEFAULT 2,
    max_occupancy INT NOT NULL DEFAULT 4,
    amenities TEXT[] NOT NULL DEFAULT '{}',
    images TEXT[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    version INT NOT NULL DEFAULT 1,
    UNIQUE(tenant_id, code)
);

CREATE INDEX idx_room_types_tenant ON room_types(tenant_id);
CREATE INDEX idx_room_types_property ON room_types(property_id);

INSERT INTO room_types (tenant_id, property_id, code, name, description, base_occupancy, max_occupancy, amenities, is_active)
SELECT 'demo', p.id, 'STD-KG', 'Standard King', 'Comfortable king room with city view', 2, 3, ARRAY['wifi','tv','minibar','ac'], true
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

INSERT INTO room_types (tenant_id, property_id, code, name, description, base_occupancy, max_occupancy, amenities, is_active)
SELECT 'demo', p.id, 'DLX-ST', 'Deluxe Suite', 'Spacious suite with separate living area', 2, 4, ARRAY['wifi','tv','minibar','ac','balcony','jacuzzi'], true
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

INSERT INTO room_types (tenant_id, property_id, code, name, description, base_occupancy, max_occupancy, amenities, is_active)
SELECT 'demo', p.id, 'PRES', 'Presidential', 'Top floor presidential suite', 4, 6, ARRAY['wifi','tv','minibar','ac','balcony','jacuzzi','butler'], true
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

-- ==================== ROOMS ====================
CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    room_type_id UUID,
    room_number VARCHAR(50) NOT NULL,
    floor VARCHAR(10) NOT NULL DEFAULT '1',
    status VARCHAR(50) NOT NULL DEFAULT 'available',
    is_smoking BOOLEAN NOT NULL DEFAULT false,
    has_ac BOOLEAN NOT NULL DEFAULT true,
    housekeeping_status VARCHAR(50) NOT NULL DEFAULT 'clean',
    attributes TEXT[] NOT NULL DEFAULT '{}',
    smart_lock_device_id VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    version INT NOT NULL DEFAULT 1,
    UNIQUE(tenant_id, property_id, room_number)
);

CREATE INDEX idx_rooms_tenant ON rooms(tenant_id);
CREATE INDEX idx_rooms_property ON rooms(property_id);
CREATE INDEX idx_rooms_status ON rooms(status);
CREATE INDEX idx_rooms_floor ON rooms(floor);

INSERT INTO rooms (tenant_id, property_id, room_type_id, room_number, floor, status, housekeeping_status, attributes)
SELECT 'demo', p.id, rt.id, '101', '1', 'occupied', 'clean', ARRAY[]
FROM properties p, room_types rt 
WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel' AND rt.code = 'STD-KG'
ON CONFLICT DO NOTHING;

INSERT INTO rooms (tenant_id, property_id, room_type_id, room_number, floor, status, housekeeping_status, attributes)
SELECT 'demo', p.id, rt.id, '102', '1', 'available', 'clean', ARRAY[]
FROM properties p, room_types rt 
WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel' AND rt.code = 'STD-KG'
ON CONFLICT DO NOTHING;

INSERT INTO rooms (tenant_id, property_id, room_type_id, room_number, floor, status, housekeeping_status, attributes)
SELECT 'demo', p.id, rt.id, '201', '2', 'occupied', 'clean', ARRAY[]
FROM properties p, room_types rt 
WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel' AND rt.code = 'DLX-ST'
ON CONFLICT DO NOTHING;

-- ==================== TRIGGERS ====================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY['tenants', 'properties', 'room_types', 'rooms'];
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
