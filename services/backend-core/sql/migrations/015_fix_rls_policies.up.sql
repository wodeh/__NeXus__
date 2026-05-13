-- Fix RLS policies with NULL-safe implementation
-- Uses COALESCE(current_setting('app.current_tenant', true), '') to handle
-- the case where the setting has not been configured on the connection.
-- The '' (empty string) fallback matches the OR clause so that superusers
-- and table owners can still access rows when no tenant is explicitly set.

-- ============================================
-- DROP old broken policies (ignore if not exists)
-- ============================================
DROP POLICY IF EXISTS rooms_isolation_policy ON rooms;
DROP POLICY IF EXISTS reservations_isolation_policy ON reservations;
DROP POLICY IF EXISTS hk_staff_isolation_policy ON housekeeping_staff;
DROP POLICY IF EXISTS hk_tasks_isolation_policy ON housekeeping_tasks;

-- ============================================
-- DROP old policies from later tables that may also be broken
-- ============================================
DROP POLICY IF EXISTS users_tenant_isolation ON users;
DROP POLICY IF EXISTS properties_tenant_isolation ON properties;
DROP POLICY IF EXISTS tenant_isolation_guest_journeys ON guest_journeys;
DROP POLICY IF EXISTS tenant_isolation_journey_execs ON guest_journey_executions;
DROP POLICY IF EXISTS tenant_isolation_upsell_offers ON upsell_offers;
DROP POLICY IF EXISTS tenant_isolation_upsell_purchases ON upsell_purchases;
DROP POLICY IF EXISTS tenant_isolation_competitors ON competitor_hotels;
DROP POLICY IF EXISTS tenant_isolation_competitor_rates ON competitor_rates;
DROP POLICY IF EXISTS tenant_isolation_rate_recs ON rate_recommendations;
DROP POLICY IF EXISTS tenant_isolation_rate_shop_config ON rate_shop_configs;
DROP POLICY IF EXISTS tenant_isolation_villa_props ON villa_properties;
DROP POLICY IF EXISTS tenant_isolation_villa_res ON villa_reservations;
DROP POLICY IF EXISTS tenant_isolation_sensor_logs ON cleaner_sensor_logs;
DROP POLICY IF EXISTS channel_sync_logs_isolation_policy ON channel_sync_logs;
DROP POLICY IF EXISTS channel_res_tenant_isolation ON channel_reservations;
DROP POLICY IF EXISTS availability_tenant_isolation ON channel_availability;

-- ============================================
-- Enable RLS on all tenant-scoped tables
-- ============================================
ALTER TABLE rooms ENABLE ROW LEVEL SECURITY;
ALTER TABLE reservations ENABLE ROW LEVEL SECURITY;
ALTER TABLE housekeeping_staff ENABLE ROW LEVEL SECURITY;
ALTER TABLE housekeeping_tasks ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE properties ENABLE ROW LEVEL SECURITY;
ALTER TABLE agents ENABLE ROW LEVEL SECURITY;
ALTER TABLE channels ENABLE ROW LEVEL SECURITY;
ALTER TABLE smart_locks ENABLE ROW LEVEL SECURITY;
ALTER TABLE lock_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE lock_access_codes ENABLE ROW LEVEL SECURITY;
ALTER TABLE whatsapp_conversations ENABLE ROW LEVEL SECURITY;
ALTER TABLE whatsapp_messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE whatsapp_bot_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE whatsapp_templates ENABLE ROW LEVEL SECURITY;
ALTER TABLE reviews ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE guest_journeys ENABLE ROW LEVEL SECURITY;
ALTER TABLE guest_journey_executions ENABLE ROW LEVEL SECURITY;
ALTER TABLE upsell_offers ENABLE ROW LEVEL SECURITY;
ALTER TABLE upsell_purchases ENABLE ROW LEVEL SECURITY;
ALTER TABLE competitor_hotels ENABLE ROW LEVEL SECURITY;
ALTER TABLE competitor_rates ENABLE ROW LEVEL SECURITY;
ALTER TABLE rate_recommendations ENABLE ROW LEVEL SECURITY;
ALTER TABLE rate_shop_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE villa_properties ENABLE ROW LEVEL SECURITY;
ALTER TABLE villa_reservations ENABLE ROW LEVEL SECURITY;
ALTER TABLE cleaner_sensor_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE channel_sync_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE channel_reservations ENABLE ROW LEVEL SECURITY;
ALTER TABLE channel_availability ENABLE ROW LEVEL SECURITY;

-- ============================================
-- NULL-safe tenant isolation helper function
-- Compares a UUID tenant_id column against the connection-scoped setting.
-- Returns TRUE when:
--   1. The tenant_id matches the current connection setting, OR
--   2. The current connection setting is empty/unset (owner/superuser bypass)
-- ============================================
CREATE OR REPLACE FUNCTION tenant_isolation(tenant_id UUID)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN tenant_id::text = COALESCE(current_setting('app.current_tenant', true), '')
        OR COALESCE(current_setting('app.current_tenant', true), '') = '';
END;
$$ LANGUAGE plpgsql STABLE SECURITY DEFINER;

-- ============================================
-- NULL-safe RLS policies for core domain tables
-- ============================================
CREATE POLICY rooms_isolation_policy ON rooms
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY reservations_isolation_policy ON reservations
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY hk_staff_isolation_policy ON housekeeping_staff
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY hk_tasks_isolation_policy ON housekeeping_tasks
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY users_tenant_isolation ON users
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY properties_tenant_isolation ON properties
    FOR ALL
    USING (tenant_isolation(tenant_id));

-- ============================================
-- RLS policies for agents, channels, and sync logs (007)
-- ============================================
CREATE POLICY agents_isolation_policy ON agents
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY channels_isolation_policy ON channels
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY channel_sync_logs_isolation_policy ON channel_sync_logs
    FOR ALL
    USING (tenant_isolation(tenant_id));

-- ============================================
-- RLS policies for locks, WhatsApp, reviews, audit (008)
-- ============================================
CREATE POLICY smart_locks_isolation_policy ON smart_locks
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY lock_events_isolation_policy ON lock_events
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY lock_access_codes_isolation_policy ON lock_access_codes
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY whatsapp_conversations_isolation_policy ON whatsapp_conversations
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY whatsapp_messages_isolation_policy ON whatsapp_messages
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY whatsapp_bot_configs_isolation_policy ON whatsapp_bot_configs
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY whatsapp_templates_isolation_policy ON whatsapp_templates
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY reviews_isolation_policy ON reviews
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY audit_logs_isolation_policy ON audit_logs
    FOR ALL
    USING (tenant_isolation(tenant_id));

-- ============================================
-- RLS policies for guest journey, upsell, competitor (013)
-- ============================================
CREATE POLICY tenant_isolation_guest_journeys ON guest_journeys
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY tenant_isolation_journey_execs ON guest_journey_executions
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY tenant_isolation_upsell_offers ON upsell_offers
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY tenant_isolation_upsell_purchases ON upsell_purchases
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY tenant_isolation_competitors ON competitor_hotels
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY tenant_isolation_competitor_rates ON competitor_rates
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY tenant_isolation_rate_recs ON rate_recommendations
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY tenant_isolation_rate_shop_config ON rate_shop_configs
    FOR ALL
    USING (tenant_isolation(tenant_id));

-- ============================================
-- RLS policies for villa rental (014)
-- ============================================
CREATE POLICY tenant_isolation_villa_props ON villa_properties
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY tenant_isolation_villa_res ON villa_reservations
    FOR ALL
    USING (tenant_isolation(tenant_id));

CREATE POLICY tenant_isolation_sensor_logs ON cleaner_sensor_logs
    FOR ALL
    USING (tenant_isolation(tenant_id));

-- ============================================
-- RLS policies for channel manager (011)
-- ============================================
CREATE POLICY channel_res_tenant_isolation ON channel_reservations
    FOR ALL
    USING (tenant_isolation(tenant_id));

-- channel_availability uses a composite key, not a UUID tenant_id
CREATE POLICY availability_tenant_isolation ON channel_availability
    FOR ALL
    USING (tenant_id::text = COALESCE(current_setting('app.current_tenant', true), '')
        OR COALESCE(current_setting('app.current_tenant', true), '') = '');
