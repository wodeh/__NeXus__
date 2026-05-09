-- Nexus Hospitality PMS Core Schema
-- Supports multi-tenant, multi-region, CQRS-ready architecture

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "timescaledb";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- ==================== TENANT ISOLATION ====================

CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    tier VARCHAR(50) NOT NULL DEFAULT 'standard',
    region VARCHAR(50) NOT NULL,
    timezone VARCHAR(50) NOT NULL DEFAULT 'UTC',
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    settings JSONB NOT NULL DEFAULT '{}',
    features JSONB NOT NULL DEFAULT '[]',
    billing_profile JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT valid_tier CHECK (tier IN ('standard', 'enterprise', 'sovereign'))
);

CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_region ON tenants(region);

ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;

-- ==================== PROPERTIES (HOTELS) ====================

CREATE TABLE properties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL,
    brand_id UUID,
    address JSONB NOT NULL,
    coordinates POINT,
    timezone VARCHAR(50) NOT NULL,
    property_type VARCHAR(50) NOT NULL,
    star_rating DECIMAL(2,1),
    total_rooms INT NOT NULL DEFAULT 0,
    settings JSONB NOT NULL DEFAULT '{}',
    operational_status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, code)
);

CREATE INDEX idx_properties_tenant ON properties(tenant_id);
CREATE INDEX idx_properties_brand ON properties(brand_id);
CREATE INDEX idx_properties_coordinates ON properties USING GIST(coordinates);

-- ==================== ROOMS ====================

CREATE TABLE room_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    base_occupancy INT NOT NULL DEFAULT 2,
    max_occupancy INT NOT NULL DEFAULT 4,
    size_sqm DECIMAL(6,2),
    bed_configuration JSONB,
    amenities JSONB NOT NULL DEFAULT '[]',
    images JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, property_id, code)
);

CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    room_type_id UUID NOT NULL REFERENCES room_types(id),
    room_number VARCHAR(20) NOT NULL,
    floor INT,
    wing VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'vacant_clean',
    housekeeping_status VARCHAR(50) DEFAULT 'clean',
    maintenance_status VARCHAR(50) DEFAULT 'operational',
    is_accessible BOOLEAN DEFAULT FALSE,
    is_smoking BOOLEAN DEFAULT FALSE,
    has_balcony BOOLEAN DEFAULT FALSE,
    view_type VARCHAR(50),
    iptv_device_id VARCHAR(100),
    iot_gateway_id VARCHAR(100),
    last_checkin TIMESTAMPTZ,
    last_checkout TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, property_id, room_number)
);

CREATE INDEX idx_rooms_property ON rooms(property_id);
CREATE INDEX idx_rooms_status ON rooms(tenant_id, property_id, status);
CREATE INDEX idx_rooms_iptv ON rooms(iptv_device_id) WHERE iptv_device_id IS NOT NULL;

CREATE TABLE room_status_history (
    time TIMESTAMPTZ NOT NULL,
    room_id UUID NOT NULL,
    property_id UUID NOT NULL,
    tenant_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL,
    previous_status VARCHAR(50),
    changed_by VARCHAR(100),
    reason TEXT,
    metadata JSONB
);

SELECT create_hypertable('room_status_history', 'time', chunk_time_interval => INTERVAL '1 day');
CREATE INDEX idx_room_status_history_room ON room_status_history(room_id, time DESC);

-- ==================== GUESTS ====================

CREATE TABLE guests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    external_id VARCHAR(100),
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    date_of_birth DATE,
    nationality VARCHAR(2),
    language_preference VARCHAR(10) DEFAULT 'en',
    id_document_type VARCHAR(50),
    id_document_number VARCHAR(100),
    id_document_expiry DATE,
    id_document_country VARCHAR(2),
    loyalty_member_id UUID,
    preferences JSONB NOT NULL DEFAULT '{}',
    special_needs JSONB NOT NULL DEFAULT '[]',
    marketing_consent BOOLEAN DEFAULT FALSE,
    privacy_consent BOOLEAN DEFAULT FALSE,
    gdpr_status JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_guests_tenant ON guests(tenant_id);
CREATE INDEX idx_guests_email ON guests(email) WHERE email IS NOT NULL;
CREATE INDEX idx_guests_loyalty ON guests(loyalty_member_id) WHERE loyalty_member_id IS NOT NULL;
CREATE INDEX idx_guests_name_trgm ON guests USING GIN((first_name || ' ' || last_name) gin_trgm_ops);

-- ==================== RESERVATIONS ====================

CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    guest_id UUID NOT NULL REFERENCES guests(id),
    room_id UUID REFERENCES rooms(id),
    confirmation_number VARCHAR(50) UNIQUE NOT NULL,
    external_reference VARCHAR(100),
    booking_source VARCHAR(50) NOT NULL,
    channel_manager_id VARCHAR(100),
    rate_plan_id UUID,
    market_segment VARCHAR(50),
    booking_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    check_in_date DATE NOT NULL,
    check_out_date DATE NOT NULL,
    arrival_time TIME,
    departure_time TIME,
    num_adults INT NOT NULL DEFAULT 1,
    num_children INT NOT NULL DEFAULT 0,
    num_rooms INT NOT NULL DEFAULT 1,
    room_type_requested UUID REFERENCES room_types(id),
    room_type_assigned UUID REFERENCES room_types(id),
    status VARCHAR(50) NOT NULL DEFAULT 'confirmed',
    payment_status VARCHAR(50) DEFAULT 'pending',
    guarantee_method VARCHAR(50),
    total_amount DECIMAL(15,2) NOT NULL,
    total_tax DECIMAL(15,2) NOT NULL DEFAULT 0,
    currency_code CHAR(3) NOT NULL,
    deposit_amount DECIMAL(15,2),
    deposit_paid BOOLEAN DEFAULT FALSE,
    special_requests TEXT,
    internal_notes TEXT,
    vip_code VARCHAR(50),
    is_group_booking BOOLEAN DEFAULT FALSE,
    group_id UUID,
    cancellation_reason VARCHAR(255),
    cancelled_at TIMESTAMPTZ,
    checked_in_at TIMESTAMPTZ,
    checked_in_by UUID,
    checked_out_at TIMESTAMPTZ,
    checked_out_by UUID,
    actual_arrival TIMESTAMPTZ,
    actual_departure TIMESTAMPTZ,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1
);

CREATE INDEX idx_reservations_tenant ON reservations(tenant_id);
CREATE INDEX idx_reservations_property ON reservations(property_id);
CREATE INDEX idx_reservations_guest ON reservations(guest_id);
CREATE INDEX idx_reservations_dates ON reservations(property_id, check_in_date, check_out_date);
CREATE INDEX idx_reservations_status ON reservations(tenant_id, status);
CREATE INDEX idx_reservations_confirmation ON reservations(confirmation_number);
CREATE INDEX idx_reservations_external ON reservations(external_reference) WHERE external_reference IS NOT NULL;

CREATE TABLE reservation_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    reservation_id UUID NOT NULL REFERENCES reservations(id),
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    changed_by VARCHAR(100),
    field_name VARCHAR(100) NOT NULL,
    old_value TEXT,
    new_value TEXT
);

CREATE INDEX idx_reservation_history ON reservation_history(reservation_id, changed_at DESC);

-- ==================== FOLIO / BILLING ====================

CREATE TABLE folios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    reservation_id UUID NOT NULL REFERENCES reservations(id),
    guest_id UUID NOT NULL REFERENCES guests(id),
    room_id UUID REFERENCES rooms(id),
    folio_number VARCHAR(50) UNIQUE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    balance DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_charges DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_payments DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_adjustments DECIMAL(15,2) NOT NULL DEFAULT 0,
    currency_code CHAR(3) NOT NULL,
    is_master BOOLEAN DEFAULT FALSE,
    parent_folio_id UUID REFERENCES folios(id),
    closed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE folio_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    folio_id UUID NOT NULL REFERENCES folios(id),
    transaction_type VARCHAR(50) NOT NULL,
    category VARCHAR(100) NOT NULL,
    description VARCHAR(255) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    currency_code CHAR(3) NOT NULL,
    posting_date DATE NOT NULL,
    posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    posted_by UUID,
    revenue_center VARCHAR(100),
    source_system VARCHAR(100),
    reference_number VARCHAR(100),
    is_voided BOOLEAN DEFAULT FALSE,
    voided_at TIMESTAMPTZ,
    voided_by UUID,
    void_reason TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_folios_reservation ON folios(reservation_id);
CREATE INDEX idx_folio_transactions_folio ON folio_transactions(folio_id, posting_date DESC);

-- ==================== HOUSEKEEPING ====================

CREATE TABLE housekeeping_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    room_id UUID NOT NULL REFERENCES rooms(id),
    task_type VARCHAR(50) NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'normal',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    assigned_to UUID,
    scheduled_at TIMESTAMPTZ NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    verified_at TIMESTAMPTZ,
    verified_by UUID,
    duration_minutes INT,
    checklist JSONB NOT NULL DEFAULT '[]',
    issues_found JSONB NOT NULL DEFAULT '[]',
    notes TEXT,
    score INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_hk_tasks_property ON housekeeping_tasks(property_id, scheduled_at);
CREATE INDEX idx_hk_tasks_assigned ON housekeeping_tasks(assigned_to, status);
CREATE INDEX idx_hk_tasks_room ON housekeeping_tasks(room_id, scheduled_at DESC);

-- ==================== MAINTENANCE ====================

CREATE TABLE maintenance_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    room_id UUID REFERENCES rooms(id),
    category VARCHAR(50) NOT NULL,
    priority VARCHAR(20) NOT NULL DEFAULT 'medium',
    status VARCHAR(50) NOT NULL DEFAULT 'reported',
    title VARCHAR(255) NOT NULL,
    description TEXT,
    reported_by VARCHAR(100),
    reported_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    assigned_to UUID,
    assigned_at TIMESTAMPTZ,
    scheduled_date DATE,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    resolution_notes TEXT,
    parts_used JSONB NOT NULL DEFAULT '[]',
    cost DECIMAL(15,2),
    downtime_minutes INT,
    photos JSONB NOT NULL DEFAULT '[]',
    guest_impact BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_maintenance_property ON maintenance_requests(property_id, status);
CREATE INDEX idx_maintenance_room ON maintenance_requests(room_id);
CREATE INDEX idx_maintenance_priority ON maintenance_requests(priority, status);

-- ==================== INVENTORY ====================

CREATE TABLE inventory_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    sku VARCHAR(100) NOT NULL,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    unit_of_measure VARCHAR(50) NOT NULL,
    par_level INT NOT NULL DEFAULT 0,
    reorder_point INT NOT NULL DEFAULT 0,
    reorder_quantity INT NOT NULL DEFAULT 0,
    current_quantity INT NOT NULL DEFAULT 0,
    location VARCHAR(100),
    supplier_id UUID,
    cost_per_unit DECIMAL(15,4),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, property_id, sku)
);

CREATE TABLE inventory_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    item_id UUID NOT NULL REFERENCES inventory_items(id),
    transaction_type VARCHAR(50) NOT NULL,
    quantity INT NOT NULL,
    unit_cost DECIMAL(15,4),
    reference_type VARCHAR(50),
    reference_id UUID,
    performed_by UUID,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_inventory_property ON inventory_items(property_id, category);
CREATE INDEX idx_inventory_low ON inventory_items(property_id) WHERE current_quantity <= reorder_point;

-- ==================== LOYALTY ====================

CREATE TABLE loyalty_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    member_number VARCHAR(50) UNIQUE NOT NULL,
    guest_id UUID REFERENCES guests(id),
    tier VARCHAR(50) NOT NULL DEFAULT 'member',
    total_points BIGINT NOT NULL DEFAULT 0,
    available_points BIGINT NOT NULL DEFAULT 0,
    lifetime_nights INT NOT NULL DEFAULT 0,
    lifetime_spend DECIMAL(15,2) NOT NULL DEFAULT 0,
    tier_qualifying_nights INT NOT NULL DEFAULT 0,
    tier_qualifying_spend DECIMAL(15,2) NOT NULL DEFAULT 0,
    tier_effective_date DATE,
    tier_expiry_date DATE,
    preferences JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE loyalty_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    member_id UUID NOT NULL REFERENCES loyalty_members(id),
    transaction_type VARCHAR(50) NOT NULL,
    points BIGINT NOT NULL,
    description VARCHAR(255),
    reference_type VARCHAR(50),
    reference_id UUID,
    balance_after BIGINT NOT NULL,
    expiry_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_loyalty_member ON loyalty_transactions(member_id, created_at DESC);

-- ==================== RATE MANAGEMENT ====================

CREATE TABLE rate_plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    code VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    room_type_id UUID REFERENCES room_types(id),
    rate_type VARCHAR(50) NOT NULL,
    base_rate DECIMAL(15,2) NOT NULL,
    currency_code CHAR(3) NOT NULL,
    min_stay INT DEFAULT 1,
    max_stay INT,
    advance_booking_days INT,
    cancellation_policy JSONB,
    deposit_policy JSONB,
    included_services JSONB NOT NULL DEFAULT '[]',
    restrictions JSONB NOT NULL DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, property_id, code)
);

CREATE TABLE daily_rates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    rate_plan_id UUID NOT NULL REFERENCES rate_plans(id),
    room_type_id UUID NOT NULL REFERENCES room_types(id),
    rate_date DATE NOT NULL,
    rate DECIMAL(15,2) NOT NULL,
    availability INT NOT NULL,
    min_stay INT DEFAULT 1,
    closed_to_arrival BOOLEAN DEFAULT FALSE,
    closed_to_departure BOOLEAN DEFAULT FALSE,
    closed BOOLEAN DEFAULT FALSE,
    last_modified_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_modified_by UUID,
    UNIQUE(tenant_id, property_id, rate_plan_id, room_type_id, rate_date)
);

CREATE INDEX idx_daily_rates_lookup ON daily_rates(property_id, rate_date, room_type_id);
CREATE INDEX idx_daily_rates_rate_plan ON daily_rates(rate_plan_id, rate_date);

-- ==================== EVENTS / BANQUETS ====================

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    event_name VARCHAR(255) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    organizer_name VARCHAR(255),
    organizer_contact JSONB,
    start_datetime TIMESTAMPTZ NOT NULL,
    end_datetime TIMESTAMPTZ NOT NULL,
    setup_time_minutes INT DEFAULT 60,
    teardown_time_minutes INT DEFAULT 60,
    num_attendees INT,
    venue_spaces JSONB NOT NULL DEFAULT '[]',
    catering_details JSONB,
    av_requirements JSONB,
    room_block JSONB,
    total_revenue DECIMAL(15,2),
    deposit_amount DECIMAL(15,2),
    status VARCHAR(50) NOT NULL DEFAULT 'inquiry',
    assigned_coordinator UUID,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ==================== WORKFORCE SCHEDULING ====================

CREATE TABLE staff_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    employee_number VARCHAR(50) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    department VARCHAR(100) NOT NULL,
    job_title VARCHAR(100) NOT NULL,
    employment_type VARCHAR(50) NOT NULL,
    hire_date DATE NOT NULL,
    termination_date DATE,
    skills JSONB NOT NULL DEFAULT '[]',
    certifications JSONB NOT NULL DEFAULT '[]',
    wage_rate DECIMAL(10,2),
    wage_type VARCHAR(20),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, property_id, employee_number)
);

CREATE TABLE work_schedules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    staff_id UUID NOT NULL REFERENCES staff_members(id),
    schedule_date DATE NOT NULL,
    shift_start TIMESTAMPTZ NOT NULL,
    shift_end TIMESTAMPTZ NOT NULL,
    break_duration_minutes INT DEFAULT 30,
    department VARCHAR(100),
    role VARCHAR(100),
    status VARCHAR(50) DEFAULT 'scheduled',
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_schedules_staff ON work_schedules(staff_id, schedule_date);
CREATE INDEX idx_schedules_date ON work_schedules(property_id, schedule_date, department);

-- ==================== AUDIT & COMPLIANCE ====================

CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    time TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    tenant_id UUID NOT NULL,
    table_name VARCHAR(100) NOT NULL,
    record_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    old_data JSONB,
    new_data JSONB,
    performed_by UUID,
    performed_by_type VARCHAR(50),
    ip_address INET,
    user_agent TEXT,
    correlation_id UUID,
    session_id VARCHAR(100)
);

SELECT create_hypertable('audit_log', 'time', chunk_time_interval => INTERVAL '7 days');

CREATE INDEX idx_audit_tenant ON audit_log(tenant_id, time DESC);
CREATE INDEX idx_audit_record ON audit_log(table_name, record_id, time DESC);

-- ==================== ROW LEVEL SECURITY ====================

DO $$
DECLARE
    tbl RECORD;
BEGIN
    FOR tbl IN
        SELECT tablename FROM pg_tables
        WHERE schemaname = 'public'
        AND tablename NOT IN ('tenants', 'audit_log', 'room_status_history')
    LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', tbl.tablename);
        EXECUTE format('
            CREATE POLICY tenant_isolation_policy ON %I
            USING (tenant_id = current_setting(''app.current_tenant'', true)::UUID);
        ', tbl.tablename);
    END LOOP;
END $$;

-- ==================== FUNCTIONS & TRIGGERS ====================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

DO $$
DECLARE
    tbl RECORD;
BEGIN
    FOR tbl IN
        SELECT tablename FROM information_schema.columns
        WHERE column_name = 'updated_at'
        AND table_schema = 'public'
    LOOP
        EXECUTE format('
            CREATE TRIGGER update_%I_updated_at
            BEFORE UPDATE ON %I
            FOR EACH ROW
            EXECUTE FUNCTION update_updated_at_column();
        ', tbl.tablename, tbl.tablename);
    END LOOP;
END $$;

CREATE OR REPLACE FUNCTION increment_reservation_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER reservation_version_trigger
    BEFORE UPDATE ON reservations
    FOR EACH ROW
    EXECUTE FUNCTION increment_reservation_version();

CREATE OR REPLACE FUNCTION audit_trigger_func()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO audit_log (tenant_id, table_name, record_id, action, old_data, performed_by)
        VALUES (OLD.tenant_id, TG_TABLE_NAME, OLD.id, 'delete', row_to_json(OLD), current_setting('app.current_user', true)::UUID);
        RETURN OLD;
    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO audit_log (tenant_id, table_name, record_id, action, old_data, new_data, performed_by)
        VALUES (NEW.tenant_id, TG_TABLE_NAME, NEW.id, 'update', row_to_json(OLD), row_to_json(NEW), current_setting('app.current_user', true)::UUID);
        RETURN NEW;
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO audit_log (tenant_id, table_name, record_id, action, new_data, performed_by)
        VALUES (NEW.tenant_id, TG_TABLE_NAME, NEW.id, 'insert', row_to_json(NEW), current_setting('app.current_user', true)::UUID);
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ language 'plpgsql';

CREATE TRIGGER reservations_audit_trigger AFTER INSERT OR UPDATE OR DELETE ON reservations
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();

CREATE TRIGGER guests_audit_trigger AFTER INSERT OR UPDATE OR DELETE ON guests
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();

CREATE TRIGGER folios_audit_trigger AFTER INSERT OR UPDATE OR DELETE ON folios
    FOR EACH ROW EXECUTE FUNCTION audit_trigger_func();
