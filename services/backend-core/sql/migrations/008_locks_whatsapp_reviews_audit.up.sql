-- Nexus Hospitality Platform — Locks, WhatsApp, Reviews, Audit migration
-- Created: 2026-05-12

-- ============================================
-- SMART LOCKS
-- ============================================
CREATE TABLE IF NOT EXISTS smart_locks (
    id                    UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id             UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    room_id               UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    room_number           TEXT NOT NULL,
    serial_number         TEXT NOT NULL,
    model                 TEXT NOT NULL DEFAULT 'OT-SL300',
    manufacturer          TEXT NOT NULL DEFAULT 'OrbitaTech',
    status                TEXT NOT NULL DEFAULT 'online',
    -- online, offline, low_battery, warning
    battery_level         INTEGER NOT NULL DEFAULT 100,
    last_communication_at TIMESTAMPTZ,
    last_unlock_at        TIMESTAMPTZ,
    last_lock_at          TIMESTAMPTZ,
    firmware_version      TEXT NOT NULL DEFAULT '3.2.1',
    remote_unlock_enabled BOOLEAN NOT NULL DEFAULT true,
    auto_lock_enabled     BOOLEAN NOT NULL DEFAULT true,
    config                JSONB NOT NULL DEFAULT '{}',
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, serial_number)
);

CREATE INDEX idx_smart_locks_tenant_id ON smart_locks(tenant_id);
CREATE INDEX idx_smart_locks_status ON smart_locks(status);
CREATE INDEX idx_smart_locks_room_id ON smart_locks(room_id);

-- ============================================
-- LOCK EVENTS
-- ============================================
CREATE TABLE IF NOT EXISTS lock_events (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id     UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    lock_id       UUID NOT NULL REFERENCES smart_locks(id) ON DELETE CASCADE,
    event_type    TEXT NOT NULL,
    -- unlock, lock, code_created, code_deleted, alert
    event_source  TEXT NOT NULL,
    -- guest_keycard, staff_master, remote, auto_lock, access_code
    details       TEXT,
    occurred_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata      JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_lock_events_lock_id ON lock_events(lock_id);
CREATE INDEX idx_lock_events_tenant_id ON lock_events(tenant_id);
CREATE INDEX idx_lock_events_occurred ON lock_events(occurred_at DESC);

-- ============================================
-- LOCK ACCESS CODES
-- ============================================
CREATE TABLE IF NOT EXISTS lock_access_codes (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    lock_id     UUID NOT NULL REFERENCES smart_locks(id) ON DELETE CASCADE,
    code        TEXT NOT NULL,
    label       TEXT NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    valid_from  TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    max_uses    INTEGER,
    use_count   INTEGER NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lock_access_codes_lock_id ON lock_access_codes(lock_id);
CREATE INDEX idx_lock_access_codes_active ON lock_access_codes(is_active) WHERE is_active = true;

-- ============================================
-- WHATSAPP CONVERSATIONS
-- ============================================
CREATE TABLE IF NOT EXISTS whatsapp_conversations (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    guest_phone     TEXT NOT NULL,
    guest_name      TEXT,
    current_state   TEXT NOT NULL DEFAULT 'greeting',
    -- greeting, ask_dates, ask_guests, ask_room, confirm, complete, support
    booking_created BOOLEAN NOT NULL DEFAULT false,
    reservation_id  UUID REFERENCES reservations(id) ON DELETE SET NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, guest_phone)
);

CREATE INDEX idx_whatsapp_conversations_tenant_id ON whatsapp_conversations(tenant_id);
CREATE INDEX idx_whatsapp_conversations_state ON whatsapp_conversations(current_state);

-- ============================================
-- WHATSAPP MESSAGES
-- ============================================
CREATE TABLE IF NOT EXISTS whatsapp_messages (
    id               UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id        UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    conversation_id  UUID NOT NULL REFERENCES whatsapp_conversations(id) ON DELETE CASCADE,
    direction        TEXT NOT NULL,
    -- inbound, outbound
    body             TEXT NOT NULL,
    status           TEXT NOT NULL DEFAULT 'sent',
    -- sent, delivered, read, failed
    sent_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata         JSONB NOT NULL DEFAULT '{}'
);

CREATE INDEX idx_whatsapp_messages_conversation ON whatsapp_messages(conversation_id);
CREATE INDEX idx_whatsapp_messages_sent ON whatsapp_messages(sent_at DESC);

-- ============================================
-- WHATSAPP BOT CONFIG
-- ============================================
CREATE TABLE IF NOT EXISTS whatsapp_bot_configs (
    id                UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    bot_enabled       BOOLEAN NOT NULL DEFAULT false,
    booking_enabled   BOOLEAN NOT NULL DEFAULT false,
    auto_reply_enabled BOOLEAN NOT NULL DEFAULT false,
    welcome_message   TEXT NOT NULL DEFAULT 'Welcome to our hotel! How can I help you today?',
    phone_number_id   TEXT,
    access_token      TEXT,
    webhook_url       TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id)
);

-- ============================================
-- WHATSAPP TEMPLATES
-- ============================================
CREATE TABLE IF NOT EXISTS whatsapp_templates (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id  UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    trigger    TEXT NOT NULL,
    response   TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, trigger)
);

-- ============================================
-- REVIEWS
-- ============================================
CREATE TABLE IF NOT EXISTS reviews (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    reservation_id  UUID REFERENCES reservations(id) ON DELETE SET NULL,
    guest_name      TEXT NOT NULL,
    room_number     TEXT,
    rating          INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    cleanliness     INTEGER NOT NULL DEFAULT 5 CHECK (cleanliness >= 1 AND cleanliness <= 5),
    service         INTEGER NOT NULL DEFAULT 5 CHECK (service >= 1 AND service <= 5),
    location        INTEGER NOT NULL DEFAULT 5 CHECK (location >= 1 AND location <= 5),
    value           INTEGER NOT NULL DEFAULT 5 CHECK (value >= 1 AND value <= 5),
    comment         TEXT,
    staff_reply     TEXT,
    replied_at      TIMESTAMPTZ,
    is_published    BOOLEAN NOT NULL DEFAULT false,
    source          TEXT NOT NULL DEFAULT 'direct',
    -- direct, ota, email
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reviews_tenant_id ON reviews(tenant_id);
CREATE INDEX idx_reviews_rating ON reviews(rating);
CREATE INDEX idx_reviews_published ON reviews(is_published) WHERE is_published = false;

-- ============================================
-- AUDIT LOGS
-- ============================================
CREATE TABLE IF NOT EXISTS audit_logs (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id     UUID,
    user_email  TEXT,
    action      TEXT NOT NULL,
    resource    TEXT NOT NULL,
    resource_id TEXT,
    details     JSONB,
    ip_address  TEXT,
    user_agent  TEXT,
    success     BOOLEAN NOT NULL DEFAULT true,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_logs_tenant_id ON audit_logs(tenant_id);
CREATE INDEX idx_audit_logs_user ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
