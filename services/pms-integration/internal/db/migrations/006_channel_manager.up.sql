CREATE TABLE IF NOT EXISTS channel_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    source VARCHAR(32) NOT NULL,
    display_name VARCHAR(64) NOT NULL,
    api_key TEXT NOT NULL,
    api_secret TEXT NOT NULL,
    property_id VARCHAR(64),
    is_active BOOLEAN NOT NULL DEFAULT true,
    commission_pct DECIMAL(5,2) NOT NULL DEFAULT 15.00,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    last_sync_status VARCHAR(16),
    last_sync_error TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_channel_connections_tenant ON channel_connections(tenant_id);
CREATE INDEX idx_channel_connections_active ON channel_connections(tenant_id, is_active);

CREATE TABLE IF NOT EXISTS channel_reservations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES channel_connections(id) ON DELETE CASCADE,
    source VARCHAR(32) NOT NULL,
    external_ref VARCHAR(64) NOT NULL,
    guest_name VARCHAR(128) NOT NULL,
    guest_email VARCHAR(128),
    guest_phone VARCHAR(32),
    room_type_id UUID REFERENCES room_types(id),
    room_id UUID REFERENCES rooms(id),
    check_in DATE NOT NULL,
    check_out DATE NOT NULL,
    nights INTEGER NOT NULL,
    adults INTEGER NOT NULL DEFAULT 1,
    children INTEGER NOT NULL DEFAULT 0,
    total_amount DECIMAL(10,2) NOT NULL,
    commission DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    net_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(16) NOT NULL DEFAULT 'confirmed',
    special_requests TEXT,
    raw_payload JSONB,
    mapped_to_reservation_id UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_channel_res_tenant ON channel_reservations(tenant_id);
CREATE INDEX idx_channel_res_channel ON channel_reservations(channel_id);
CREATE INDEX idx_channel_res_external ON channel_reservations(tenant_id, external_ref);
CREATE INDEX idx_channel_res_dates ON channel_reservations(check_in, check_out);
CREATE INDEX idx_channel_res_status ON channel_reservations(tenant_id, status);

CREATE TABLE IF NOT EXISTS channel_sync_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    channel_id UUID NOT NULL REFERENCES channel_connections(id) ON DELETE CASCADE,
    direction VARCHAR(8) NOT NULL,
    status VARCHAR(16) NOT NULL,
    records INTEGER NOT NULL DEFAULT 0,
    error_msg TEXT,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_sync_logs_tenant ON channel_sync_logs(tenant_id);
CREATE INDEX idx_sync_logs_channel ON channel_sync_logs(channel_id);

-- WhatsApp Bot
CREATE TABLE IF NOT EXISTS whatsapp_bot_configs (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    phone_number_id VARCHAR(32),
    access_token TEXT,
    webhook_secret TEXT,
    welcome_message TEXT NOT NULL DEFAULT 'Welcome! How can I help you today?',
    auto_reply_enabled BOOLEAN NOT NULL DEFAULT true,
    booking_enabled BOOLEAN NOT NULL DEFAULT true,
    checkin_enabled BOOLEAN NOT NULL DEFAULT true,
    checkout_enabled BOOLEAN NOT NULL DEFAULT true,
    support_enabled BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS whatsapp_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    guest_phone VARCHAR(32) NOT NULL,
    guest_name VARCHAR(128),
    current_state VARCHAR(32) NOT NULL DEFAULT 'greeting',
    context_json JSONB,
    last_message_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_wa_conv_tenant ON whatsapp_conversations(tenant_id);
CREATE INDEX idx_wa_conv_phone ON whatsapp_conversations(tenant_id, guest_phone);

CREATE TABLE IF NOT EXISTS whatsapp_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    conversation_id UUID NOT NULL REFERENCES whatsapp_conversations(id) ON DELETE CASCADE,
    direction VARCHAR(8) NOT NULL,
    body TEXT NOT NULL,
    message_type VARCHAR(16) NOT NULL DEFAULT 'text',
    button_payload VARCHAR(128),
    sent_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMP WITH TIME ZONE,
    read_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(16) NOT NULL DEFAULT 'sent'
);

CREATE INDEX idx_wa_msg_conv ON whatsapp_messages(conversation_id);
CREATE INDEX idx_wa_msg_tenant ON whatsapp_messages(tenant_id);

-- Direct Booking Engine
CREATE TABLE IF NOT EXISTS promo_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    code VARCHAR(32) NOT NULL,
    discount_type VARCHAR(16) NOT NULL,
    discount_value DECIMAL(10,2) NOT NULL,
    max_uses INTEGER NOT NULL DEFAULT 0,
    uses_count INTEGER NOT NULL DEFAULT 0,
    min_nights INTEGER NOT NULL DEFAULT 1,
    valid_from DATE NOT NULL,
    valid_until DATE NOT NULL,
    applicable_room_types UUID[],
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_promo_tenant ON promo_codes(tenant_id);
CREATE UNIQUE INDEX idx_promo_code ON promo_codes(tenant_id, code);

CREATE TABLE IF NOT EXISTS booking_widget_configs (
    tenant_id UUID PRIMARY KEY REFERENCES tenants(id) ON DELETE CASCADE,
    enabled BOOLEAN NOT NULL DEFAULT false,
    property_id UUID REFERENCES properties(id),
    theme_color VARCHAR(7) NOT NULL DEFAULT '#6366f1',
    logo_url TEXT,
    title VARCHAR(128) NOT NULL DEFAULT 'Book Your Stay',
    subtitle VARCHAR(256),
    show_promo_code BOOLEAN NOT NULL DEFAULT true,
    show_extras BOOLEAN NOT NULL DEFAULT true,
    upsell_enabled BOOLEAN NOT NULL DEFAULT true,
    require_deposit BOOLEAN NOT NULL DEFAULT false,
    deposit_pct DECIMAL(5,2) NOT NULL DEFAULT 25.00,
    success_redirect_url TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS booking_upsells (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    per_night BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_upsells_tenant ON booking_upsells(tenant_id);

CREATE TABLE IF NOT EXISTS direct_booking_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID REFERENCES properties(id),
    widget_config_id UUID,
    guest_name VARCHAR(128) NOT NULL,
    guest_email VARCHAR(128) NOT NULL,
    guest_phone VARCHAR(32),
    room_type_id UUID REFERENCES room_types(id),
    check_in DATE NOT NULL,
    check_out DATE NOT NULL,
    nights INTEGER NOT NULL,
    adults INTEGER NOT NULL DEFAULT 2,
    children INTEGER NOT NULL DEFAULT 0,
    base_rate DECIMAL(10,2) NOT NULL,
    upsells_json JSONB,
    promo_code_id UUID REFERENCES promo_codes(id),
    discount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    total DECIMAL(10,2) NOT NULL,
    deposit_amount DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dbs_tenant ON direct_booking_sessions(tenant_id);
CREATE INDEX idx_dbs_status ON direct_booking_sessions(status, expires_at);

-- Guest Reviews
CREATE TABLE IF NOT EXISTS guest_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    guest_id UUID REFERENCES guests(id),
    guest_name VARCHAR(128) NOT NULL,
    reservation_id UUID,
    room_number VARCHAR(8),
    channel VARCHAR(16) NOT NULL,
    external_ref VARCHAR(64),
    overall_rating INTEGER NOT NULL CHECK (overall_rating BETWEEN 1 AND 5),
    cleanliness INTEGER CHECK (cleanliness BETWEEN 1 AND 5),
    service INTEGER CHECK (service BETWEEN 1 AND 5),
    location INTEGER CHECK (location BETWEEN 1 AND 5),
    value INTEGER CHECK (value BETWEEN 1 AND 5),
    amenities INTEGER CHECK (amenities BETWEEN 1 AND 5),
    comment TEXT,
    is_published BOOLEAN NOT NULL DEFAULT false,
    staff_response TEXT,
    responded_by UUID,
    responded_at TIMESTAMP WITH TIME ZONE,
    sent_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_reviews_tenant ON guest_reviews(tenant_id);
CREATE INDEX idx_reviews_guest ON guest_reviews(guest_id);
CREATE INDEX idx_reviews_res ON guest_reviews(reservation_id);
CREATE INDEX idx_reviews_channel ON guest_reviews(tenant_id, channel);

CREATE TABLE IF NOT EXISTS review_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    guest_id UUID,
    reservation_id UUID,
    channel VARCHAR(16) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'scheduled',
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE,
    opened_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    bounce_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_rev_req_tenant ON review_requests(tenant_id);
CREATE INDEX idx_rev_req_sched ON review_requests(scheduled_at, status);

-- Communications
CREATE TABLE IF NOT EXISTS communication_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    subject VARCHAR(256),
    body TEXT NOT NULL,
    channel VARCHAR(8) NOT NULL,
    category VARCHAR(16) NOT NULL,
    variables VARCHAR(32)[],
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comm_tmpl_tenant ON communication_templates(tenant_id);
CREATE INDEX idx_comm_tmpl_cat ON communication_templates(tenant_id, category);

CREATE TABLE IF NOT EXISTS communication_sequences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(128) NOT NULL,
    description TEXT,
    trigger VARCHAR(32) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comm_seq_tenant ON communication_sequences(tenant_id);

CREATE TABLE IF NOT EXISTS communication_sequence_steps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence_id UUID NOT NULL REFERENCES communication_sequences(id) ON DELETE CASCADE,
    template_id UUID NOT NULL REFERENCES communication_templates(id),
    step_order INTEGER NOT NULL,
    delay_hours INTEGER NOT NULL DEFAULT 0,
    channel VARCHAR(8) NOT NULL,
    conditions_json JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comm_step_seq ON communication_sequence_steps(sequence_id);

CREATE TABLE IF NOT EXISTS scheduled_communications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    reservation_id UUID,
    guest_id UUID,
    guest_phone VARCHAR(32),
    guest_email VARCHAR(128),
    template_id UUID REFERENCES communication_templates(id),
    sequence_id UUID REFERENCES communication_sequences(id),
    sequence_step_id UUID REFERENCES communication_sequence_steps(id),
    channel VARCHAR(8) NOT NULL,
    subject VARCHAR(256),
    body_rendered TEXT NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'scheduled',
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,
    sent_at TIMESTAMP WITH TIME ZONE,
    delivered_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sched_comm_tenant ON scheduled_communications(tenant_id);
CREATE INDEX idx_sched_comm_status ON scheduled_communications(status, scheduled_at);

CREATE TABLE IF NOT EXISTS communication_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    guest_id UUID,
    reservation_id UUID,
    channel VARCHAR(8) NOT NULL,
    direction VARCHAR(8) NOT NULL,
    subject VARCHAR(256),
    body TEXT NOT NULL,
    status VARCHAR(16) NOT NULL,
    external_id VARCHAR(128),
    sent_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_comm_logs_tenant ON communication_logs(tenant_id);
