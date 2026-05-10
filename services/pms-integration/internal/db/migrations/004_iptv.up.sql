CREATE TABLE IF NOT EXISTS iptv_channels (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    channel_number INT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    stream_url TEXT,
    icon_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_premium BOOLEAN NOT NULL DEFAULT false,
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, property_id, channel_number)
);

CREATE INDEX idx_iptv_channels_tenant ON iptv_channels(tenant_id);
CREATE INDEX idx_iptv_channels_property ON iptv_channels(property_id);
CREATE INDEX idx_iptv_channels_category ON iptv_channels(category);

CREATE TABLE IF NOT EXISTS iptv_content (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    title VARCHAR(255) NOT NULL,
    content_type VARCHAR(50) NOT NULL DEFAULT 'movie',
    category VARCHAR(50) NOT NULL DEFAULT 'entertainment',
    description TEXT,
    poster_url TEXT,
    stream_url TEXT,
    duration_minutes INT,
    is_premium BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_iptv_content_tenant ON iptv_content(tenant_id);
CREATE INDEX idx_iptv_content_property ON iptv_content(property_id);
CREATE INDEX idx_iptv_content_type ON iptv_content(content_type);

CREATE TABLE IF NOT EXISTS iptv_room_bindings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    device_type VARCHAR(50) NOT NULL DEFAULT 'smart_tv',
    welcome_screen_enabled BOOLEAN NOT NULL DEFAULT true,
    welcome_message TEXT,
    guest_name_display BOOLEAN NOT NULL DEFAULT true,
    checkout_reminder_enabled BOOLEAN NOT NULL DEFAULT true,
    language_override VARCHAR(10),
    channels_enabled JSONB NOT NULL DEFAULT '[]',
    content_enabled JSONB NOT NULL DEFAULT '[]',
    last_sync_at TIMESTAMPTZ,
    status VARCHAR(50) NOT NULL DEFAULT 'online',
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(tenant_id, property_id, room_id, device_id)
);

CREATE INDEX idx_iptv_room_bindings_tenant ON iptv_room_bindings(tenant_id);
CREATE INDEX idx_iptv_room_bindings_room ON iptv_room_bindings(room_id);
CREATE INDEX idx_iptv_room_bindings_status ON iptv_room_bindings(status);

CREATE TABLE IF NOT EXISTS iptv_guest_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id VARCHAR(255) NOT NULL,
    property_id UUID NOT NULL,
    room_binding_id UUID NOT NULL REFERENCES iptv_room_bindings(id) ON DELETE CASCADE,
    reservation_id UUID,
    guest_id UUID,
    session_started_at TIMESTAMPTZ DEFAULT NOW(),
    session_ended_at TIMESTAMPTZ,
    channel_watch_history JSONB NOT NULL DEFAULT '[]',
    content_watch_history JSONB NOT NULL DEFAULT '[]',
    welcome_shown_at TIMESTAMPTZ,
    checkout_reminder_shown_at TIMESTAMPTZ,
    total_watch_minutes INT NOT NULL DEFAULT 0,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_iptv_guest_sessions_tenant ON iptv_guest_sessions(tenant_id);
CREATE INDEX idx_iptv_guest_sessions_binding ON iptv_guest_sessions(room_binding_id);
CREATE INDEX idx_iptv_guest_sessions_guest ON iptv_guest_sessions(guest_id);

-- Demo data
INSERT INTO iptv_channels (tenant_id, property_id, name, channel_number, category, stream_url, is_active, language)
SELECT 'demo', p.id, 'Nexus Welcome', 1, 'welcome', 'https://nexus.tv/welcome', true, 'en'
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

INSERT INTO iptv_channels (tenant_id, property_id, name, channel_number, category, stream_url, is_active, language)
SELECT 'demo', p.id, 'CNN International', 2, 'news', 'https://cnn.com/live', true, 'en'
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

INSERT INTO iptv_channels (tenant_id, property_id, name, channel_number, category, stream_url, is_active, language)
SELECT 'demo', p.id, 'BBC World', 3, 'news', 'https://bbc.com/live', true, 'en'
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

INSERT INTO iptv_channels (tenant_id, property_id, name, channel_number, category, stream_url, is_active, language)
SELECT 'demo', p.id, 'HBO', 4, 'entertainment', 'https://hbo.com/live', true, 'en'
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

INSERT INTO iptv_channels (tenant_id, property_id, name, channel_number, category, stream_url, is_active, is_premium, language)
SELECT 'demo', p.id, 'Nexus Premium Movies', 5, 'movies', 'https://nexus.tv/premium', true, true, 'en'
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

INSERT INTO iptv_content (tenant_id, property_id, title, content_type, category, description, duration_minutes, is_active)
SELECT 'demo', p.id, 'Local Attractions Guide', 'info', 'hospitality', 'Discover the best local restaurants, attractions, and hidden gems curated by our concierge team.', 15, true
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

INSERT INTO iptv_content (tenant_id, property_id, title, content_type, category, description, duration_minutes, is_premium, is_active)
SELECT 'demo', p.id, 'Premium Movie Collection', 'movie', 'entertainment', 'Curated selection of award-winning films available exclusively to suite guests.', 120, true, true
FROM properties p WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel'
ON CONFLICT DO NOTHING;

-- Room bindings for demo rooms
INSERT INTO iptv_room_bindings (tenant_id, property_id, room_id, device_id, device_type, welcome_screen_enabled, welcome_message, guest_name_display, status)
SELECT 'demo', p.id, r.id, 'TV-' || r.room_number, 'smart_tv', true, 'Welcome to Grand Plaza Hotel', true, 'online'
FROM properties p, rooms r
WHERE p.tenant_id = 'demo' AND p.slug = 'grand-plaza-hotel' AND r.tenant_id = 'demo' AND r.property_id = p.id
ON CONFLICT DO NOTHING;

DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY['iptv_channels', 'iptv_content', 'iptv_room_bindings', 'iptv_guest_sessions'];
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
