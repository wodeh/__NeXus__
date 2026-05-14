-- Seed admin users and roles for demo

-- Demo super admin
INSERT INTO users (tenant_id, email, name, phone, role, status, password_hash, permissions, created_at, updated_at)
VALUES 
  ('demo', 'super@nexus.com', 'System Administrator', '+1-555-0000', 'super_admin', 'active', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQYe7bSb1n3J9Zg5j7z1J0Z5J0Zq', '["*"]', NOW(), NOW()),
  ('demo', 'admin@grandplaza.com', 'Alex Rivera', '+1-555-0100', 'admin', 'active', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQYe7bSb1n3J9Zg5j7z1J0Z5J0Zq', '["reservations:read","reservations:write","rooms:read","rooms:write","housekeeping:read","housekeeping:write","reports:read","settings:read","settings:write","users:read","users:write"]', NOW(), NOW()),
  ('demo', 'manager@grandplaza.com', 'Sarah Chen', '+1-555-0101', 'manager', 'active', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQYe7bSb1n3J9Zg5j7z1J0Z5J0Zq', '["reservations:read","reservations:write","housekeeping:read","reports:read"]', NOW(), NOW()),
  ('demo', 'frontdesk@grandplaza.com', 'James Wilson', '+1-555-0102', 'front_desk', 'active', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQYe7bSb1n3J9Zg5j7z1J0Z5J0Zq', '["reservations:read","reservations:write","guests:read"]', NOW(), NOW()),
  ('demo', 'housekeeper@grandplaza.com', 'Maria Garcia', '+1-555-0103', 'housekeeper', 'active', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQYe7bSb1n3J9Zg5j7z1J0Z5J0Zq', '["housekeeping:read","housekeeping:write","rooms:read"]', NOW(), NOW())
ON CONFLICT (tenant_id, email) DO NOTHING;

-- Demo properties with full details
INSERT INTO properties (tenant_id, name, code, address, city, country, phone, email, timezone, currency, star_rating, status, created_at, updated_at)
VALUES
  ('demo', 'Grand Plaza Downtown', 'DOWNTOWN', '123 Main Street', 'New York', 'USA', '+1-212-555-0100', 'downtown@grandplaza.com', 'America/New_York', 'USD', 5, 'active', NOW(), NOW()),
  ('demo', 'Grand Plaza Beachfront', 'BEACH', '456 Ocean Avenue', 'Miami', 'USA', '+1-305-555-0200', 'beach@grandplaza.com', 'America/New_York', 'USD', 4, 'active', NOW(), NOW()),
  ('demo', 'Grand Plaza Mountain Lodge', 'MOUNTAIN', '789 Summit Road', 'Denver', 'USA', '+1-720-555-0300', 'mountain@grandplaza.com', 'America/Denver', 'USD', 4, 'active', NOW(), NOW()),
  ('demo', 'Grand Plaza Dubai Marina', 'DUBAI', '321 Sheikh Zayed Road', 'Dubai', 'UAE', '+971-4-555-0400', 'dubai@grandplaza.com', 'Asia/Dubai', 'AED', 5, 'active', NOW(), NOW())
ON CONFLICT (tenant_id, code) DO NOTHING;

-- System config
INSERT INTO tenant_properties (tenant_id, property_id, name, timezone, locale, currency, config)
VALUES ('demo', 'system', 'System Defaults', 'UTC', 'en', 'USD', '{"default_check_in_time":"15:00","default_check_out_time":"11:00","auto_confirm":true,"require_deposit":true,"deposit_percent":20,"allow_walk_in":true,"overbooking_enabled":false}'::jsonb)
ON CONFLICT (tenant_id, property_id) DO NOTHING;
