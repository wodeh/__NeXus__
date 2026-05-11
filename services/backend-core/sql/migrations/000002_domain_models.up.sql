-- Nexus Hospitality Platform — domain models migration
-- Rooms, Reservations, Housekeeping

-- ============================================
-- ROOMS
-- ============================================
CREATE TABLE IF NOT EXISTS rooms (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id TEXT NOT NULL DEFAULT 'main',
    number      TEXT NOT NULL,
    type        TEXT NOT NULL,
    floor       TEXT NOT NULL,
    bed_type    TEXT,
    status      TEXT NOT NULL DEFAULT 'vacant_clean',
    -- vacant_clean, vacant_dirty, occupied, blocked, maintenance
    rate_night  INTEGER NOT NULL DEFAULT 0,
    config      JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    version     INTEGER NOT NULL DEFAULT 1,
    UNIQUE(tenant_id, property_id, number)
);

CREATE INDEX idx_rooms_tenant_id ON rooms(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_rooms_status ON rooms(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_rooms_floor ON rooms(floor) WHERE deleted_at IS NULL;

-- ============================================
-- RESERVATIONS
-- ============================================
CREATE TABLE IF NOT EXISTS reservations (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id        UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id      TEXT NOT NULL DEFAULT 'main',
    guest_name       TEXT NOT NULL,
    email            TEXT,
    phone            TEXT,
    room_number      TEXT,
    room_type        TEXT,
    check_in         DATE NOT NULL,
    check_out        DATE NOT NULL,
    adults           INTEGER NOT NULL DEFAULT 1,
    children         INTEGER NOT NULL DEFAULT 0,
    status           TEXT NOT NULL DEFAULT 'confirmed',
    -- confirmed, checked_in, checked_out, cancelled, no_show
    source           TEXT NOT NULL DEFAULT 'direct',
    -- walk_in, ota, direct, agent
    total            INTEGER NOT NULL DEFAULT 0,
    balance          INTEGER NOT NULL DEFAULT 0,
    special_requests TEXT,
    vip              BOOLEAN NOT NULL DEFAULT FALSE,
    color            TEXT,
    config           JSONB NOT NULL DEFAULT '{}',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at       TIMESTAMPTZ,
    version          INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_reservations_tenant_id ON reservations(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_reservations_status ON reservations(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_reservations_check_in ON reservations(check_in) WHERE deleted_at IS NULL;
CREATE INDEX idx_reservations_room ON reservations(room_number) WHERE deleted_at IS NULL;

-- ============================================
-- HOUSEKEEPING STAFF
-- ============================================
CREATE TABLE IF NOT EXISTS housekeeping_staff (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name              TEXT NOT NULL,
    role              TEXT NOT NULL DEFAULT 'cleaner',
    -- cleaner, inspector, supervisor
    active_shift      TEXT DEFAULT 'day',
    -- day, evening, night
    max_rooms_per_day INTEGER NOT NULL DEFAULT 15,
    phone             TEXT,
    email             TEXT,
    active            BOOLEAN NOT NULL DEFAULT TRUE,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ,
    version           INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_hk_staff_tenant_id ON housekeeping_staff(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_hk_staff_active ON housekeeping_staff(active) WHERE deleted_at IS NULL;

-- ============================================
-- HOUSEKEEPING TASKS
-- ============================================
CREATE TABLE IF NOT EXISTS housekeeping_tasks (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id     UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    room_number   TEXT NOT NULL,
    task_type     TEXT NOT NULL DEFAULT 'full_clean',
    -- full_clean, refill, maintenance, inspection
    status        TEXT NOT NULL DEFAULT 'pending',
    -- pending, in_progress, completed, skipped, blocked
    assigned_to   UUID REFERENCES housekeeping_staff(id),
    priority      TEXT NOT NULL DEFAULT 'normal',
    -- low, normal, high, urgent
    notes         TEXT,
    started_at    TIMESTAMPTZ,
    completed_at  TIMESTAMPTZ,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ,
    version       INTEGER NOT NULL DEFAULT 1
);

CREATE INDEX idx_hk_tasks_tenant_id ON housekeeping_tasks(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_hk_tasks_status ON housekeeping_tasks(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_hk_tasks_assigned ON housekeeping_tasks(assigned_to) WHERE deleted_at IS NULL;
CREATE INDEX idx_hk_tasks_room ON housekeeping_tasks(room_number) WHERE deleted_at IS NULL;

-- ============================================
-- TRIGGERS
-- ============================================
CREATE TRIGGER rooms_audit_trigger
    BEFORE UPDATE ON rooms
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_func();

CREATE TRIGGER reservations_audit_trigger
    BEFORE UPDATE ON reservations
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_func();

CREATE TRIGGER housekeeping_staff_audit_trigger
    BEFORE UPDATE ON housekeeping_staff
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_func();

CREATE TRIGGER housekeeping_tasks_audit_trigger
    BEFORE UPDATE ON housekeeping_tasks
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_func();

-- ============================================
-- RLS
-- ============================================
ALTER TABLE rooms ENABLE ROW LEVEL SECURITY;
ALTER TABLE reservations ENABLE ROW LEVEL SECURITY;
ALTER TABLE housekeeping_staff ENABLE ROW LEVEL SECURITY;
ALTER TABLE housekeeping_tasks ENABLE ROW LEVEL SECURITY;

CREATE POLICY rooms_isolation_policy ON rooms
    FOR ALL
    USING (tenant_id::text = current_setting('app.current_tenant', true) OR current_setting('app.current_tenant', true) = '');

CREATE POLICY reservations_isolation_policy ON reservations
    FOR ALL
    USING (tenant_id::text = current_setting('app.current_tenant', true) OR current_setting('app.current_tenant', true) = '');

CREATE POLICY hk_staff_isolation_policy ON housekeeping_staff
    FOR ALL
    USING (tenant_id::text = current_setting('app.current_tenant', true) OR current_setting('app.current_tenant', true) = '');

CREATE POLICY hk_tasks_isolation_policy ON housekeeping_tasks
    FOR ALL
    USING (tenant_id::text = current_setting('app.current_tenant', true) OR current_setting('app.current_tenant', true) = '');
