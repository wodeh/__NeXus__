-- Property & Inventory Foundation Migration
-- Adds properties, room_types, rooms, rate_plans, daily_rates for a shippable PMS.

-- ==================== PROPERTIES ====================

CREATE TABLE IF NOT EXISTS properties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) NOT NULL,
    address JSONB NOT NULL DEFAULT '{}',
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ,
    UNIQUE(tenant_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_properties_tenant ON properties(tenant_id);
CREATE INDEX IF NOT EXISTS idx_properties_active ON properties(tenant_id, is_active) WHERE is_active = TRUE;

ALTER TABLE properties ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'properties' AND policyname = 'properties_tenant_isolation'
    ) THEN
        CREATE POLICY properties_tenant_isolation ON properties
        USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== ROOM TYPES ====================

CREATE TABLE IF NOT EXISTS room_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    property_id UUID NOT NULL REFERENCES properties(id),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_occupancy INT NOT NULL DEFAULT 2,
    max_occupancy INT NOT NULL DEFAULT 4,
    amenities JSONB NOT NULL DEFAULT '[]',
    images JSONB NOT NULL DEFAULT '[]',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ,
    UNIQUE(tenant_id, property_id, code)
);

CREATE INDEX IF NOT EXISTS idx_room_types_tenant ON room_types(tenant_id);
CREATE INDEX IF NOT EXISTS idx_room_types_property ON room_types(property_id);

ALTER TABLE room_types ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'room_types' AND policyname = 'room_types_tenant_isolation'
    ) THEN
        CREATE POLICY room_types_tenant_isolation ON room_types
        USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== ROOMS ====================

CREATE TABLE IF NOT EXISTS rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    property_id UUID NOT NULL REFERENCES properties(id),
    room_type_id UUID NOT NULL REFERENCES room_types(id),
    room_number VARCHAR(50) NOT NULL,
    floor VARCHAR(20),
    status VARCHAR(50) NOT NULL DEFAULT 'available',
    is_smoking BOOLEAN NOT NULL DEFAULT FALSE,
    has_ac BOOLEAN NOT NULL DEFAULT TRUE,
    housekeeping_status VARCHAR(50) NOT NULL DEFAULT 'clean',
    attributes JSONB NOT NULL DEFAULT '[]',
    smart_lock_device_id VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ,
    UNIQUE(tenant_id, property_id, room_number)
);

CREATE INDEX IF NOT EXISTS idx_rooms_tenant ON rooms(tenant_id);
CREATE INDEX IF NOT EXISTS idx_rooms_property ON rooms(property_id);
CREATE INDEX IF NOT EXISTS idx_rooms_type ON rooms(room_type_id);
CREATE INDEX IF NOT EXISTS idx_rooms_status ON rooms(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_rooms_housekeeping ON rooms(tenant_id, housekeeping_status);

ALTER TABLE rooms ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'rooms' AND policyname = 'rooms_tenant_isolation'
    ) THEN
        CREATE POLICY rooms_tenant_isolation ON rooms
        USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== RATE PLANS ====================

CREATE TABLE IF NOT EXISTS rate_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    property_id UUID NOT NULL REFERENCES properties(id),
    room_type_id UUID NOT NULL REFERENCES room_types(id),
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_rate DECIMAL(15,2) NOT NULL,
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    min_los INT NOT NULL DEFAULT 1,
    max_los INT,
    advance_booking_days INT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    restrictions JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ,
    UNIQUE(tenant_id, property_id, code)
);

CREATE INDEX IF NOT EXISTS idx_rate_plans_tenant ON rate_plans(tenant_id);
CREATE INDEX IF NOT EXISTS idx_rate_plans_room_type ON rate_plans(room_type_id);

ALTER TABLE rate_plans ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'rate_plans' AND policyname = 'rate_plans_tenant_isolation'
    ) THEN
        CREATE POLICY rate_plans_tenant_isolation ON rate_plans
        USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== DAILY RATES ====================

CREATE TABLE IF NOT EXISTS daily_rates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    rate_plan_id UUID NOT NULL REFERENCES rate_plans(id),
    rate_date DATE NOT NULL,
    rate DECIMAL(15,2) NOT NULL,
    availability INT NOT NULL DEFAULT 0,
    min_stay INT NOT NULL DEFAULT 1,
    close_to_arrival BOOLEAN NOT NULL DEFAULT FALSE,
    close_to_departure BOOLEAN NOT NULL DEFAULT FALSE,
    stop_sell BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(rate_plan_id, rate_date)
);

CREATE INDEX IF NOT EXISTS idx_daily_rates_plan ON daily_rates(rate_plan_id);
CREATE INDEX IF NOT EXISTS idx_daily_rates_date ON daily_rates(rate_date);

ALTER TABLE daily_rates ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'daily_rates' AND policyname = 'daily_rates_tenant_isolation'
    ) THEN
        CREATE POLICY daily_rates_tenant_isolation ON daily_rates
        USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== TRIGGERS ====================

DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY['properties', 'room_types', 'rooms', 'rate_plans'];
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
