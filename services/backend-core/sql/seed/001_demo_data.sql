-- Nexus Hospitality Platform — Seed Data
-- Created: 2026-05-12

-- Tenant
INSERT INTO tenants (id, external_id, name, region, tier, config)
VALUES (
  'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
  'grand-plaza',
  'Grand Plaza Hotel',
  'us-east-1',
  'enterprise',
  '{"timezone":"America/New_York","locale":"en","currency":"USD","modules":["reservations","housekeeping","iptv","locks","communications","revenue","reviews","audit","whatsapp","channel_manager","booking_engine"]}'
)
ON CONFLICT (id) DO NOTHING;

-- Users
INSERT INTO users (id, tenant_id, email, password_hash, name, role, is_active)
VALUES
  ('a0000001-0000-0000-0000-000000000001', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'admin@grandplaza.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQzBZN0UfGNEJHILxN5GJAnWL7SC', 'Alex Rivera', 'admin', true),
  ('a0000002-0000-0000-0000-000000000002', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'manager@grandplaza.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQzBZN0UfGNEJHILxN5GJAnWL7SC', 'Sarah Chen', 'manager', true),
  ('a0000003-0000-0000-0000-000000000003', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'frontdesk@grandplaza.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQzBZN0UfGNEJHILxN5GJAnWL7SC', 'James Wilson', 'staff', true),
  ('a0000004-0000-0000-0000-000000000004', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'cleaner1@grandplaza.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQzBZN0UfGNEJHILxN5GJAnWL7SC', 'Maria Garcia', 'cleaner', true),
  ('a0000005-0000-0000-0000-000000000005', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'cleaner2@grandplaza.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqQzBZN0UfGNEJHILxN5GJAnWL7SC', 'Liu Wei', 'cleaner', true)
ON CONFLICT (id) DO NOTHING;

-- Rooms
INSERT INTO rooms (id, tenant_id, property_id, number, type, floor, bed_type, status, rate_night, config)
VALUES
  ('a1000001-0000-0000-0000-000000000001', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', '101', 'standard', '1', 'queen', 'occupied', 149, '{"tv":true,"wifi":true,"minibar":true,"balcony":false}'),
  ('a1000002-0000-0000-0000-000000000002', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', '102', 'standard', '1', 'queen', 'vacant_clean', 149, '{"tv":true,"wifi":true,"minibar":true}'),
  ('a1000003-0000-0000-0000-000000000003', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', '103', 'standard', '1', 'king', 'occupied', 169, '{"tv":true,"wifi":true,"minibar":true}'),
  ('a1000004-0000-0000-0000-000000000004', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', '201', 'deluxe', '2', 'king', 'occupied', 219, '{"tv":true,"wifi":true,"minibar":true,"balcony":true,"jacuzzi":true}'),
  ('a1000005-0000-0000-0000-000000000005', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', '202', 'deluxe', '2', 'queen', 'vacant_dirty', 199, '{"tv":true,"wifi":true,"minibar":true,"balcony":true}'),
  ('a1000006-0000-0000-0000-000000000006', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', '203', 'suite', '2', 'king', 'occupied', 299, '{"tv":true,"wifi":true,"minibar":true,"balcony":true,"kitchenette":true}'),
  ('a1000007-0000-0000-0000-000000000007', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', '301', 'penthouse', '3', 'king', 'blocked', 499, '{"tv":true,"wifi":true,"minibar":true,"balcony":true,"pool":true,"butler":true}'),
  ('a1000008-0000-0000-0000-000000000008', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', '302', 'standard', '3', 'twin', 'maintenance', 139, '{"tv":true,"wifi":true}')
ON CONFLICT (id) DO NOTHING;

-- Reservations
INSERT INTO reservations (id, tenant_id, property_id, guest_name, email, phone, room_number, room_type, check_in, check_out, adults, children, status, source, total, balance, special_requests, vip)
VALUES
  ('a2000001-0000-0000-0000-000000000001', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', 'Alice Chen', 'alice@email.com', '+1-555-0101', '101', 'standard', '2026-05-10', '2026-05-14', 2, 0, 'checked_in', 'direct', 596, 0, 'Extra pillows, high floor', false),
  ('a2000002-0000-0000-0000-000000000002', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', 'Bob Jones', 'bob@email.com', '+1-555-0102', '103', 'standard', '2026-05-09', '2026-05-13', 1, 0, 'checked_in', 'booking.com', 676, 0, 'Late check-in', false),
  ('a2000003-0000-0000-0000-000000000003', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', 'Carol White', 'carol@email.com', '+1-555-0103', '201', 'deluxe', '2026-05-11', '2026-05-15', 2, 1, 'checked_in', 'direct', 876, 200, 'Crib needed', false),
  ('a2000004-0000-0000-0000-000000000004', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', 'David Kim', 'david@email.com', '+1-555-0104', '203', 'suite', '2026-05-12', '2026-05-16', 2, 0, 'confirmed', 'expedia', 1196, 0, 'Anniversary - champagne', true),
  ('a2000005-0000-0000-0000-000000000005', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', 'Emma Watson', 'emma@email.com', '+1-555-0105', '102', 'standard', '2026-05-15', '2026-05-18', 2, 0, 'confirmed', 'direct', 447, 447, '', false),
  ('a2000006-0000-0000-0000-000000000006', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', 'Frank Lee', 'frank@email.com', '+1-555-0106', '102', 'standard', '2026-05-14', '2026-05-17', 1, 0, 'confirmed', 'booking.com', 447, 0, 'Quiet room', false),
  ('a2000007-0000-0000-0000-000000000007', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'main', 'Grace Ho', 'grace@email.com', '+1-555-0107', '202', 'deluxe', '2026-05-10', '2026-05-11', 2, 0, 'checked_out', 'walk_in', 199, 0, '', false)
ON CONFLICT (id) DO NOTHING;

-- Smart Locks
INSERT INTO smart_locks (id, tenant_id, room_id, room_number, serial_number, model, status, battery_level, firmware_version, config)
VALUES
  ('a3000001-0000-0000-0000-000000000001', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a1000001-0000-0000-0000-000000000001', '101', 'OT-SL300-1001', 'OT-SL300', 'online', 87, '3.2.1', '{"auto_lock_seconds":30}'),
  ('a3000002-0000-0000-0000-000000000002', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a1000003-0000-0000-0000-000000000003', '103', 'OT-SL300-1002', 'OT-SL300', 'online', 92, '3.2.1', '{"auto_lock_seconds":30}'),
  ('a3000003-0000-0000-0000-000000000003', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a1000004-0000-0000-0000-000000000004', '201', 'OT-SL300-1003', 'OT-SL300', 'online', 45, '3.2.1', '{"auto_lock_seconds":60}'),
  ('a3000004-0000-0000-0000-000000000004', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a1000006-0000-0000-0000-000000000006', '203', 'OT-SL300-1004', 'OT-SL300', 'warning', 12, '3.1.9', '{"auto_lock_seconds":30}'),
  ('a3000005-0000-0000-0000-000000000005', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a1000002-0000-0000-0000-000000000002', '102', 'OT-SL300-1005', 'OT-SL300', 'online', 98, '3.2.1', '{"auto_lock_seconds":30}'),
  ('a3000006-0000-0000-0000-000000000006', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a1000005-0000-0000-0000-000000000005', '202', 'OT-SL300-1006', 'OT-SL300', 'offline', 0, '3.1.8', '{"auto_lock_seconds":30}')
ON CONFLICT (id) DO NOTHING;

-- Lock Events
INSERT INTO lock_events (tenant_id, lock_id, event_type, event_source, details, occurred_at)
VALUES
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a3000001-0000-0000-0000-000000000001', 'unlock', 'guest_keycard', 'Room 101 unlocked by guest keycard', NOW() - INTERVAL '2 hours'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a3000001-0000-0000-0000-000000000001', 'lock', 'auto_lock', 'Auto-locked after 30 seconds', NOW() - INTERVAL '2 hours'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a3000003-0000-0000-0000-000000000003', 'unlock', 'guest_keycard', 'Room 201 unlocked by guest keycard', NOW() - INTERVAL '4 hours'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a3000004-0000-0000-0000-000000000004', 'alert', 'system', 'Low battery warning: 12%', NOW() - INTERVAL '1 day'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a3000006-0000-0000-0000-000000000006', 'alert', 'system', 'Lock offline - no communication for 24h', NOW() - INTERVAL '2 days');

-- Lock Access Codes
INSERT INTO lock_access_codes (tenant_id, lock_id, code, label, is_active, valid_from, valid_until)
VALUES
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a3000001-0000-0000-0000-000000000001', '2847', 'Guest Alice Chen', true, '2026-05-10', '2026-05-14'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a3000001-0000-0000-0000-000000000001', '9999', 'Master Override', true, NULL, NULL),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a3000003-0000-0000-0000-000000000003', '5932', 'Guest Carol White', true, '2026-05-11', '2026-05-15'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a3000005-0000-0000-0000-000000000005', '1122', 'Cleaning Card', true, NULL, NULL),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a3000005-0000-0000-0000-000000000005', '4455', 'Restock Card', true, NULL, NULL);

-- Reviews
INSERT INTO reviews (id, tenant_id, reservation_id, guest_name, room_number, rating, cleanliness, service, location, value, comment, staff_reply, is_published, source)
VALUES
  ('a4000001-0000-0000-0000-000000000001', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a2000007-0000-0000-0000-000000000007', 'Grace Ho', '202', 5, 5, 5, 4, 5, 'Absolutely wonderful stay! The staff was incredibly attentive and the room was spotless. Will definitely return.', 'Thank you Grace! We''re thrilled you enjoyed your stay. See you next time!', true, 'direct'),
  ('a4000002-0000-0000-0000-000000000002', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NULL, 'Michael Brown', NULL, 4, 4, 5, 4, 4, 'Great location and friendly staff. Room was a bit small but very clean. Breakfast was excellent.', NULL, true, 'ota'),
  ('a4000003-0000-0000-0000-000000000003', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NULL, 'Jennifer Park', NULL, 2, 2, 3, 4, 2, 'The AC was broken in our room and it took 2 hours to get it fixed. Not acceptable for a hotel of this caliber.', 'We sincerely apologize for the AC issue. We have addressed this with our maintenance team and upgraded your room.', false, 'email'),
  ('a4000004-0000-0000-0000-000000000004', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NULL, 'Robert Taylor', NULL, 5, 5, 5, 5, 5, 'Best hotel experience I''ve had in years. From check-in to check-out, everything was perfect. The concierge was especially helpful.', NULL, false, 'ota'),
  ('a4000005-0000-0000-0000-000000000005', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NULL, 'Sophie Martin', NULL, 3, 3, 4, 3, 3, 'Decent hotel but overpriced for what you get. The gym was outdated and the pool was closed during our stay.', NULL, true, 'ota')
ON CONFLICT (id) DO NOTHING;

-- WhatsApp Conversations
INSERT INTO whatsapp_conversations (id, tenant_id, guest_phone, guest_name, current_state, booking_created, created_at)
VALUES
  ('a5000001-0000-0000-0000-000000000001', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '+1-555-0201', 'John Smith', 'complete', true, NOW() - INTERVAL '2 days'),
  ('a5000002-0000-0000-0000-000000000002', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '+1-555-0202', 'Lisa Wong', 'support', false, NOW() - INTERVAL '1 hour'),
  ('a5000003-0000-0000-0000-000000000003', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '+1-555-0203', 'Ahmed Hassan', 'ask_dates', false, NOW() - INTERVAL '30 minutes'),
  ('a5000004-0000-0000-0000-000000000004', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', '+1-555-0204', 'Nina Petrova', 'greeting', false, NOW() - INTERVAL '5 minutes')
ON CONFLICT (id) DO NOTHING;

-- WhatsApp Messages
INSERT INTO whatsapp_messages (tenant_id, conversation_id, direction, body, status, sent_at)
VALUES
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a5000001-0000-0000-0000-000000000001', 'outbound', 'Welcome to Grand Plaza! How can I help you today?', 'delivered', NOW() - INTERVAL '2 days'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a5000001-0000-0000-0000-000000000001', 'inbound', 'I want to book a room for May 15-18', 'read', NOW() - INTERVAL '2 days'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a5000001-0000-0000-0000-000000000001', 'outbound', 'Great! I found availability. How many guests?', 'delivered', NOW() - INTERVAL '2 days'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a5000001-0000-0000-0000-000000000001', 'inbound', '2 adults', 'read', NOW() - INTERVAL '2 days'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a5000001-0000-0000-0000-000000000001', 'outbound', 'Perfect! Your reservation for Room 102 is confirmed. Check-in: May 15, 3:00 PM.', 'delivered', NOW() - INTERVAL '2 days'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a5000002-0000-0000-0000-000000000002', 'outbound', 'Welcome to Grand Plaza! How can I help you today?', 'delivered', NOW() - INTERVAL '1 hour'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a5000002-0000-0000-0000-000000000002', 'inbound', 'My AC is not working in room 201', 'read', NOW() - INTERVAL '50 minutes'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a5000002-0000-0000-0000-000000000002', 'outbound', 'I''m sorry to hear that. Our maintenance team is on their way. Estimated arrival: 15 minutes.', 'delivered', NOW() - INTERVAL '48 minutes'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a5000003-0000-0000-0000-000000000003', 'outbound', 'Welcome to Grand Plaza! How can I help you today?', 'delivered', NOW() - INTERVAL '30 minutes'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a5000003-0000-0000-0000-000000000003', 'inbound', 'Book a room', 'read', NOW() - INTERVAL '28 minutes'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a5000003-0000-0000-0000-000000000003', 'outbound', 'When would you like to check in?', 'delivered', NOW() - INTERVAL '27 minutes'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a5000004-0000-0000-0000-000000000004', 'outbound', 'Welcome to Grand Plaza! How can I help you today?', 'sent', NOW() - INTERVAL '5 minutes');

-- WhatsApp Bot Config
INSERT INTO whatsapp_bot_configs (tenant_id, bot_enabled, booking_enabled, auto_reply_enabled, welcome_message, phone_number_id)
VALUES
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', true, true, true, 'Welcome to Grand Plaza Hotel! 🏨\nI can help you with:\n• BOOK - Make a reservation\n• CHECKIN - Check-in information\n• WIFI - WiFi password\n• HELP - Talk to front desk', '15551234567')
ON CONFLICT (tenant_id) DO NOTHING;

-- WhatsApp Templates
INSERT INTO whatsapp_templates (tenant_id, trigger, response)
VALUES
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'book', 'I can help you book a room! What dates are you looking for? (Format: YYYY-MM-DD to YYYY-MM-DD)'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'checkin', 'Check-in time is 3:00 PM. Your digital key will be activated automatically. Room number and WiFi details will be sent 1 hour before arrival.'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'wifi', 'Network: NexusGuest\nPassword: Welcome2026'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'help', 'Connecting you to our front desk team. Someone will assist you shortly. For urgent matters, call: +1-555-0199'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'checkout', 'Check-out time is 11:00 AM. You can check out via the app or at the front desk. We hope you enjoyed your stay!');

-- Audit Logs
INSERT INTO audit_logs (tenant_id, user_id, user_email, action, resource, resource_id, details, success, created_at)
VALUES
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0000001-0000-0000-0000-000000000001', 'admin@grandplaza.com', 'login', 'auth', NULL, '{"ip":"192.168.1.100","method":"password"}', true, NOW() - INTERVAL '2 hours'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0000002-0000-0000-0000-000000000002', 'manager@grandplaza.com', 'check_in', 'reservation', 'a2000004-0000-0000-0000-000000000004', '{"room":"203","guest":"David Kim"}', true, NOW() - INTERVAL '4 hours'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0000003-0000-0000-0000-000000000003', 'frontdesk@grandplaza.com', 'update_reservation', 'reservation', 'a2000001-0000-0000-0000-000000000001', '{"field":"special_requests","old":"","new":"Extra pillows, high floor"}', true, NOW() - INTERVAL '6 hours'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0000004-0000-0000-0000-000000000004', 'cleaner1@grandplaza.com', 'mark_clean', 'room', 'a1000002-0000-0000-0000-000000000002', '{"room":"102","status":"vacant_clean"}', true, NOW() - INTERVAL '1 hour'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0000001-0000-0000-0000-000000000001', 'admin@grandplaza.com', 'update_lock', 'smart_lock', 'a3000004-0000-0000-0000-000000000004', '{"field":"firmware","old":"3.1.9","new":"3.2.1"}', true, NOW() - INTERVAL '30 minutes'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', NULL, NULL, 'failed_login', 'auth', NULL, '{"ip":"10.0.0.55","email":"hacker@evil.com","reason":"invalid_password"}', false, NOW() - INTERVAL '1 hour'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0000002-0000-0000-0000-000000000002', 'manager@grandplaza.com', 'publish_review', 'review', 'a4000001-0000-0000-0000-000000000001', '{"published":true}', true, NOW() - INTERVAL '2 days'),
  ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0000003-0000-0000-0000-000000000003', 'frontdesk@grandplaza.com', 'create_access_code', 'smart_lock', 'a3000001-0000-0000-0000-000000000001', '{"code":"2847","label":"Guest Alice Chen"}', true, NOW() - INTERVAL '1 day');
