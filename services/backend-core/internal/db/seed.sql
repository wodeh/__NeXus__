-- Seed demo data for Guru Hospitality Platform
-- Run with: psql $DATABASE_URL -f seed.sql

-- Insert demo rooms
INSERT INTO rooms (tenant_id, number, floor, type, bed_type, max_occupancy, status, rate, features, last_cleaned, last_inspected, notes, created_at, updated_at)
VALUES
  ('demo', '101', 1, 'Standard', 'Queen', 2, 'vacant_clean', 89, '{"wifi":true,"tv":true,"ac":true}', NOW(), NOW(), 'Street view', NOW(), NOW()),
  ('demo', '102', 1, 'Standard', 'Queen', 2, 'occupied', 89, '{"wifi":true,"tv":true,"ac":true}', NOW(), NOW(), 'Street view', NOW(), NOW()),
  ('demo', '103', 1, 'Deluxe', 'King', 2, 'vacant_clean', 129, '{"wifi":true,"tv":true,"ac":true,"balcony":true}', NOW(), NOW(), 'Garden view', NOW(), NOW()),
  ('demo', '104', 1, 'Deluxe', 'King', 2, 'vacant_dirty', 129, '{"wifi":true,"tv":true,"ac":true,"balcony":true}', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days', 'Garden view', NOW(), NOW()),
  ('demo', '105', 1, 'Suite', 'King', 4, 'occupied', 229, '{"wifi":true,"tv":true,"ac":true,"balcony":true,"jacuzzi":true,"minibar":true}', NOW(), NOW(), 'Executive suite', NOW(), NOW()),
  ('demo', '201', 2, 'Standard', 'Queen', 2, 'vacant_clean', 89, '{"wifi":true,"tv":true,"ac":true}', NOW(), NOW(), 'Pool view', NOW(), NOW()),
  ('demo', '202', 2, 'Standard', 'Twin', 2, 'occupied', 89, '{"wifi":true,"tv":true,"ac":true}', NOW(), NOW(), 'Pool view', NOW(), NOW()),
  ('demo', '203', 2, 'Deluxe', 'King', 2, 'blocked', 129, '{"wifi":true,"tv":true,"ac":true,"balcony":true}', NOW(), NOW(), 'Maintenance blocked', NOW(), NOW()),
  ('demo', '204', 2, 'Deluxe', 'King', 2, 'vacant_clean', 129, '{"wifi":true,"tv":true,"ac":true,"balcony":true}', NOW(), NOW(), 'Garden view', NOW(), NOW()),
  ('demo', '205', 2, 'Suite', 'King', 4, 'occupied', 229, '{"wifi":true,"tv":true,"ac":true,"balcony":true,"jacuzzi":true,"minibar":true}', NOW(), NOW(), 'Penthouse suite', NOW(), NOW()),
  ('demo', '301', 3, 'Standard', 'Queen', 2, 'vacant_dirty', 89, '{"wifi":true,"tv":true,"ac":true}', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day', 'Needs deep clean', NOW(), NOW()),
  ('demo', '302', 3, 'Standard', 'Twin', 2, 'vacant_clean', 89, '{"wifi":true,"tv":true,"ac":true}', NOW(), NOW(), 'Quiet corner', NOW(), NOW()),
  ('demo', '303', 3, 'Deluxe', 'King', 2, 'occupied', 129, '{"wifi":true,"tv":true,"ac":true,"balcony":true}', NOW(), NOW(), 'City view', NOW(), NOW()),
  ('demo', '304', 3, 'Suite', 'King', 4, 'vacant_clean', 229, '{"wifi":true,"tv":true,"ac":true,"balcony":true,"jacuzzi":true,"minibar":true}', NOW(), NOW(), 'Bridal suite', NOW(), NOW()),
  ('demo', '305', 3, 'Deluxe', 'Queen', 2, 'maintenance', 129, '{"wifi":true,"tv":true,"ac":true}', NOW(), NOW(), 'AC repair scheduled', NOW(), NOW())
ON CONFLICT (tenant_id, number) DO NOTHING;

-- Insert demo reservations
INSERT INTO reservations (tenant_id, guest_name, email, phone, room_type, room_number, check_in, check_out, adults, children, status, source, total_amount, special_requests, vip, created_at, updated_at)
VALUES
  ('demo', 'Alice Johnson', 'alice@example.com', '+1-555-0101', 'Standard', '102', '2026-05-09', '2026-05-12', 2, 0, 'checked_in', 'direct', 267, '', false, NOW(), NOW()),
  ('demo', 'Bob Smith', 'bob@example.com', '+1-555-0102', 'Suite', '105', '2026-05-08', '2026-05-13', 2, 2, 'checked_in', 'ota_booking', 1145, 'Extra towels', true, NOW(), NOW()),
  ('demo', 'Carol Davis', 'carol@example.com', '+1-555-0103', 'Standard', '202', '2026-05-10', '2026-05-14', 1, 0, 'checked_in', 'direct', 356, 'High floor preferred', false, NOW(), NOW()),
  ('demo', 'David Wilson', 'david@example.com', '+1-555-0104', 'Deluxe', '303', '2026-05-07', '2026-05-11', 2, 0, 'checked_in', 'ota_booking', 516, 'Late checkout requested', false, NOW(), NOW()),
  ('demo', 'Eva Martinez', 'eva@example.com', '+1-555-0105', 'Suite', '205', '2026-05-11', '2026-05-16', 2, 1, 'confirmed', 'direct', 1145, 'Anniversary setup', true, NOW(), NOW()),
  ('demo', 'Frank Brown', 'frank@example.com', '+1-555-0106', 'Standard', '201', '2026-05-12', '2026-05-15', 2, 0, 'confirmed', 'walk_in', 267, '', false, NOW(), NOW()),
  ('demo', 'Grace Lee', 'grace@example.com', '+1-555-0107', 'Deluxe', '103', '2026-05-13', '2026-05-18', 2, 0, 'confirmed', 'ota_booking', 645, 'Allergic to feathers', false, NOW(), NOW()),
  ('demo', 'Henry Taylor', 'henry@example.com', '+1-555-0108', 'Standard', '101', '2026-05-11', '2026-05-12', 1, 0, 'confirmed', 'direct', 89, '', false, NOW(), NOW()),
  ('demo', 'Ivy Chen', 'ivy@example.com', '+1-555-0109', 'Deluxe', '204', '2026-05-14', '2026-05-19', 2, 2, 'confirmed', 'ota_booking', 645, 'Connecting rooms if possible', false, NOW(), NOW()),
  ('demo', 'Jack Anderson', 'jack@example.com', '+1-555-0110', 'Suite', '304', '2026-05-15', '2026-05-20', 4, 0, 'confirmed', 'direct', 1145, 'Honeymoon package', true, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Insert housekeeping staff
INSERT INTO housekeeping_staff (tenant_id, name, role, phone, email, shift, status, workload_limit, created_at, updated_at)
VALUES
  ('demo', 'Maria Garcia', 'room_attendant', '+1-555-0201', 'maria@hotel.com', 'morning', 'active', 15, NOW(), NOW()),
  ('demo', 'John Kim', 'room_attendant', '+1-555-0202', 'john@hotel.com', 'afternoon', 'active', 15, NOW(), NOW()),
  ('demo', 'Sarah Johnson', 'supervisor', '+1-555-0203', 'sarah@hotel.com', 'morning', 'active', 25, NOW(), NOW()),
  ('demo', 'Liu Wei', 'room_attendant', '+1-555-0204', 'liu@hotel.com', 'night', 'active', 12, NOW(), NOW()),
  ('demo', 'Emma Thompson', 'inspector', '+1-555-0205', 'emma@hotel.com', 'morning', 'active', 20, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- Insert housekeeping tasks
INSERT INTO housekeeping_tasks (tenant_id, room_number, staff_id, task_type, priority, status, notes, scheduled_date, started_at, completed_at, verified_by, created_at, updated_at)
SELECT
  'demo',
  r.number,
  s.id,
  CASE WHEN r.status = 'vacant_dirty' THEN 'cleaning'
       WHEN r.status = 'maintenance' THEN 'maintenance'
       ELSE 'cleaning' END,
  CASE WHEN r.status = 'maintenance' THEN 'urgent' ELSE 'normal' END,
  CASE WHEN r.status = 'vacant_clean' THEN 'completed'
       WHEN r.status = 'occupied' THEN 'pending'
       ELSE 'pending' END,
  CASE WHEN r.status = 'maintenance' THEN 'AC repair needed'
       WHEN r.number = '104' THEN 'Deep clean required'
       WHEN r.number = '301' THEN 'Post-event cleanup'
       ELSE 'Regular turnover' END,
  CURRENT_DATE,
  CASE WHEN r.status = 'vacant_clean' THEN NOW() - INTERVAL '2 hours' ELSE NULL END,
  CASE WHEN r.status = 'vacant_clean' THEN NOW() - INTERVAL '1 hour' ELSE NULL END,
  CASE WHEN r.status = 'vacant_clean' THEN (SELECT id FROM housekeeping_staff WHERE role = 'inspector' LIMIT 1) ELSE NULL END,
  NOW(),
  NOW()
FROM rooms r
CROSS JOIN (SELECT id FROM housekeeping_staff WHERE role = 'room_attendant' ORDER BY random() LIMIT 1) s
WHERE r.tenant_id = 'demo';

-- Insert demo properties (for the multi-property feature)
INSERT INTO properties (tenant_id, name, code, address, city, country, timezone, status, created_at, updated_at)
VALUES
  ('demo', 'Guru Downtown', 'DOWNTOWN', '123 Main St', 'New York', 'USA', 'America/New_York', 'active', NOW(), NOW()),
  ('demo', 'Guru Beachfront', 'BEACH', '456 Ocean Ave', 'Miami', 'USA', 'America/New_York', 'active', NOW(), NOW()),
  ('demo', 'Guru Mountain Lodge', 'MOUNTAIN', '789 Summit Rd', 'Denver', 'USA', 'America/Denver', 'active', NOW(), NOW())
ON CONFLICT DO NOTHING;
