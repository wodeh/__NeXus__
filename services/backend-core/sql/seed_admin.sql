-- Seed admin users, tenant, properties, and system config for demo
-- Compatible with schema from migrations 000001 + 009 + 010

-- 1. Ensure demo tenant exists
INSERT INTO tenants (external_id, name, region, tier, config)
VALUES ('demo', 'Demo Tenant', 'us-east-1', 'enterprise', '{}')
ON CONFLICT (external_id) DO NOTHING;

-- 2. Seed users (no phone, no status, no permissions columns in actual schema)
INSERT INTO users (tenant_id, email, password_hash, name, role, is_active, created_at, updated_at)
VALUES
  ((SELECT id FROM tenants WHERE external_id = 'demo'), 'super@nexus.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQYe7bSb1n3J9Zg5j7z1J0Z5J0Zq', 'System Administrator', 'super_admin', true, NOW(), NOW()),
  ((SELECT id FROM tenants WHERE external_id = 'demo'), 'admin@grandplaza.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQYe7bSb1n3J9Zg5j7z1J0Z5J0Zq', 'Alex Rivera', 'admin', true, NOW(), NOW()),
  ((SELECT id FROM tenants WHERE external_id = 'demo'), 'manager@grandplaza.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQYe7bSb1n3J9Zg5j7z1J0Z5J0Zq', 'Sarah Chen', 'manager', true, NOW(), NOW()),
  ((SELECT id FROM tenants WHERE external_id = 'demo'), 'frontdesk@grandplaza.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQYe7bSb1n3J9Zg5j7z1J0Z5J0Zq', 'James Wilson', 'front_desk', true, NOW(), NOW()),
  ((SELECT id FROM tenants WHERE external_id = 'demo'), 'housekeeper@grandplaza.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQYe7bSb1n3J9Zg5j7z1J0Z5J0Zq', 'Maria Garcia', 'housekeeper', true, NOW(), NOW())
ON CONFLICT (tenant_id, email) DO NOTHING;

-- 3. Seed properties (no code column, status -> is_active)
INSERT INTO properties (tenant_id, name, address, city, country, phone, email, timezone, currency, star_rating, is_active, created_at, updated_at)
VALUES
  ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Grand Plaza Downtown', '123 Main Street', 'New York', 'USA', '+1-212-555-0100', 'downtown@grandplaza.com', 'America/New_York', 'USD', 5, true, NOW(), NOW()),
  ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Grand Plaza Beachfront', '456 Ocean Avenue', 'Miami', 'USA', '+1-305-555-0200', 'beach@grandplaza.com', 'America/New_York', 'USD', 4, true, NOW(), NOW()),
  ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Grand Plaza Mountain Lodge', '789 Summit Road', 'Denver', 'USA', '+1-720-555-0300', 'mountain@grandplaza.com', 'America/Denver', 'USD', 4, true, NOW(), NOW()),
  ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Grand Plaza Dubai Marina', '321 Sheikh Zayed Road', 'Dubai', 'UAE', '+971-4-555-0400', 'dubai@grandplaza.com', 'Asia/Dubai', 'AED', 5, true, NOW(), NOW())
ON CONFLICT (tenant_id, name) DO NOTHING;

-- 4. Seed system defaults into tenant_properties (tenant_id must be UUID, property_id is TEXT)
INSERT INTO tenant_properties (tenant_id, property_id, name, timezone, locale, currency, config)
VALUES (
  (SELECT id FROM tenants WHERE external_id = 'demo'),
  'system',
  'System Defaults',
  'UTC',
  'en',
  'USD',
  '{"default_check_in_time":"15:00","default_check_out_time":"11:00","auto_confirm":true,"require_deposit":true,"deposit_percent":20,"allow_walk_in":true,"overbooking_enabled":false}'::jsonb
)
ON CONFLICT (tenant_id, property_id) DO NOTHING;
