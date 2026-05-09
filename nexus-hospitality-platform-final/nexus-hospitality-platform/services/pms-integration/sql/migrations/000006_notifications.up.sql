-- Migration: Notifications — Email, SMS, Guest Messaging
-- Version: 000006

CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    channel VARCHAR(30) NOT NULL,
    subject VARCHAR(255),
    body TEXT NOT NULL,
    variables TEXT[],
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notification_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    template_id UUID REFERENCES notification_templates(id) ON DELETE SET NULL,
    channel VARCHAR(30) NOT NULL,
    recipient VARCHAR(255) NOT NULL,
    subject VARCHAR(255),
    body TEXT,
    status VARCHAR(30) NOT NULL DEFAULT 'queued',
    error_message TEXT,
    sent_at TIMESTAMPTZ,
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS guest_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    guest_id UUID NOT NULL REFERENCES guests(id) ON DELETE CASCADE,
    reservation_id UUID REFERENCES reservations(id) ON DELETE SET NULL,
    direction VARCHAR(20) NOT NULL DEFAULT 'outbound',
    channel VARCHAR(30) NOT NULL DEFAULT 'in_app',
    content TEXT NOT NULL,
    sent_by UUID REFERENCES users(id) ON DELETE SET NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_notification_logs_tenant ON notification_logs(tenant_id);
CREATE INDEX idx_notification_logs_status ON notification_logs(status);
CREATE INDEX idx_guest_messages_guest ON guest_messages(guest_id);
CREATE INDEX idx_guest_messages_tenant ON guest_messages(tenant_id);
CREATE INDEX idx_guest_messages_unread ON guest_messages(tenant_id, is_read) WHERE is_read = FALSE;

-- Triggers
CREATE TRIGGER update_notification_templates_updated_at
    BEFORE UPDATE ON notification_templates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- RLS
ALTER TABLE notification_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE notification_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE guest_messages ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_notification_templates ON notification_templates
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY tenant_isolation_notification_logs ON notification_logs
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);

CREATE POLICY tenant_isolation_guest_messages ON guest_messages
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);

-- Seed default templates
INSERT INTO notification_templates (tenant_id, name, channel, subject, body, variables)
SELECT 
    t.id,
    'Check-in Welcome',
    'email',
    'Welcome to {{property_name}}',
    'Dear {{guest_name}},\n\nWelcome to {{property_name}}! Your room {{room_number}} is ready.\n\nCheck-in: {{check_in_date}}\nCheck-out: {{check_out_date}}\n\nWiFi: {{wifi_ssid}} / Password: {{wifi_password}}\n\nEnjoy your stay!',
    ARRAY['guest_name', 'property_name', 'room_number', 'check_in_date', 'check_out_date', 'wifi_ssid', 'wifi_password']
FROM tenants t
ON CONFLICT DO NOTHING;
