CREATE TABLE IF NOT EXISTS smart_locks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    room_id UUID REFERENCES rooms(id) ON DELETE SET NULL,
    device_name VARCHAR(255) NOT NULL,
    device_serial VARCHAR(255) NOT NULL,
    manufacturer VARCHAR(50) NOT NULL DEFAULT 'orbitatech',
    model VARCHAR(100) NOT NULL DEFAULT 'OT-SL300',
    firmware_version VARCHAR(50),
    status VARCHAR(50) NOT NULL DEFAULT 'online',
    battery_level INT NOT NULL DEFAULT 100,
    connection_type VARCHAR(50) NOT NULL DEFAULT 'wifi',
    ip_address INET,
    mac_address VARCHAR(17),
    last_communication_at TIMESTAMPTZ,
    last_unlock_at TIMESTAMPTZ,
    last_lock_at TIMESTAMPTZ,
    remote_unlock_enabled BOOLEAN NOT NULL DEFAULT true,
    auto_lock_enabled BOOLEAN NOT NULL DEFAULT true,
    auto_lock_delay_seconds INT NOT NULL DEFAULT 30,
    settings JSONB NOT NULL DEFAULT '{}',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, device_serial)
);

CREATE INDEX idx_smart_locks_tenant ON smart_locks(tenant_id);
CREATE INDEX idx_smart_locks_property ON smart_locks(property_id);
CREATE INDEX idx_smart_locks_room ON smart_locks(room_id);
CREATE INDEX idx_smart_locks_status ON smart_locks(status);

CREATE TABLE IF NOT EXISTS lock_access_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    smart_lock_id UUID NOT NULL REFERENCES smart_locks(id) ON DELETE CASCADE,
    reservation_id UUID REFERENCES reservations(id) ON DELETE SET NULL,
    guest_id UUID REFERENCES guests(id) ON DELETE SET NULL,
    code VARCHAR(20) NOT NULL,
    code_type VARCHAR(50) NOT NULL DEFAULT 'temporary',
    label VARCHAR(255),
    valid_from TIMESTAMPTZ NOT NULL,
    valid_until TIMESTAMPTZ,
    max_uses INT,
    uses_count INT NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_master BOOLEAN NOT NULL DEFAULT false,
    created_by UUID,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    revoked_at TIMESTAMPTZ,
    revoked_reason TEXT,
    metadata JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_lock_access_codes_tenant ON lock_access_codes(tenant_id);
CREATE INDEX idx_lock_access_codes_lock ON lock_access_codes(smart_lock_id);
CREATE INDEX idx_lock_access_codes_reservation ON lock_access_codes(reservation_id);
CREATE INDEX idx_lock_access_codes_active ON lock_access_codes(is_active) WHERE is_active = true;

CREATE TABLE IF NOT EXISTS lock_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    smart_lock_id UUID NOT NULL REFERENCES smart_locks(id) ON DELETE CASCADE,
    access_code_id UUID REFERENCES lock_access_codes(id) ON DELETE SET NULL,
    event_type VARCHAR(50) NOT NULL,
    event_source VARCHAR(50) NOT NULL DEFAULT 'device',
    user_id UUID,
    guest_id UUID,
    details JSONB NOT NULL DEFAULT '{}',
    occurred_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_lock_events_tenant ON lock_events(tenant_id);
CREATE INDEX idx_lock_events_lock ON lock_events(smart_lock_id);
CREATE INDEX idx_lock_events_type ON lock_events(event_type);
CREATE INDEX idx_lock_events_occurred ON lock_events(occurred_at DESC);

-- Demo data
INSERT INTO smart_locks (tenant_id, property_id, room_id, device_name, device_serial, manufacturer, model, firmware_version, status, battery_level, remote_unlock_enabled, auto_lock_enabled, settings)
SELECT 'demo', p.id, r.id, 'Lock-' || r.room_number, 'OT-' || r.room_number || '-2026', 'orbitatech', 'OT-SL300', 'v2.1.4', 'online', 87, true, true, '{"keypad_backlight":true,"sound_enabled":true,"tamper_alert":true}'::jsonb
FROM properties p, rooms r
WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel' AND r.tenant_id = 'demo' AND r.property_id = p.id
ON CONFLICT DO NOTHING;

-- Demo access codes
INSERT INTO lock_access_codes (tenant_id, property_id, smart_lock_id, reservation_id, guest_id, code, code_type, label, valid_from, valid_until, max_uses, is_active)
SELECT 'demo', p.id, sl.id, res.id, g.id, '1234', 'temporary', 'John Smith Stay', res.check_in_date::timestamptz, res.check_out_date::timestamptz + INTERVAL '1 hour', NULL, true
FROM properties p, smart_locks sl, reservations res, guests g, rooms r
WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
  AND sl.tenant_id = 'demo' AND sl.property_id = p.id AND sl.room_id = r.id AND r.room_number = '101'
  AND res.tenant_id = 'demo' AND res.room_id = r.id AND res.status = 'checked_in'
  AND g.id = res.guest_id AND g.email = 'john.smith@email.com'
ON CONFLICT DO NOTHING;

-- Demo events
INSERT INTO lock_events (tenant_id, property_id, smart_lock_id, event_type, event_source, details)
SELECT 'demo', p.id, sl.id, 'device_online', 'system', '{"battery":87,"signal_strength":-45}'::jsonb
FROM properties p, smart_locks sl
WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel' AND sl.tenant_id = 'demo' AND sl.property_id = p.id
ON CONFLICT DO NOTHING;

DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY['smart_locks', 'lock_access_codes', 'lock_events'];
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
