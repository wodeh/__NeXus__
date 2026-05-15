-- ========================================================================
-- Nexus Demo Data Seed — Plain SQL Version
-- No DO block: each statement runs independently with clear error messages
-- Run after seed_admin.sql has created the tenant, users, and properties
-- ========================================================================

-- Get demo tenant UUID once
SELECT id AS demo_tenant_id FROM tenants WHERE external_id = 'demo';

-- =====================================================================
-- 1. ROOMS — 24 rooms across 3 floors + basement, multiple types
-- =====================================================================
INSERT INTO rooms (tenant_id, property_id, number, type, floor, bed_type, status, rate_night, config, created_at, updated_at)
VALUES
    -- Floor 1: Lobby + Standard
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '101', 'standard', '1', 'queen',  'vacant_clean',  199, '{"view":"courtyard","smoking":false}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '102', 'standard', '1', 'queen',  'occupied',      199, '{"view":"courtyard","smoking":false}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '103', 'standard', '1', 'twin',   'vacant_dirty',  189, '{"view":"street","smoking":false}',    NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '104', 'standard', '1', 'twin',   'occupied',      189, '{"view":"street","smoking":false}',    NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '105', 'suite',    '1', 'king',   'vacant_clean',  349, '{"view":"courtyard","smoking":false,"jacuzzi":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '106', 'suite',    '1', 'king',   'blocked',       349, '{"view":"courtyard","smoking":false,"jacuzzi":true}', NOW(), NOW()),
    -- Floor 2: Deluxe + Family
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '201', 'deluxe',   '2', 'king',   'vacant_clean',  279, '{"view":"ocean","smoking":false,"balcony":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '202', 'deluxe',   '2', 'king',   'occupied',      279, '{"view":"ocean","smoking":false,"balcony":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '203', 'deluxe',   '2', 'queen',  'occupied',      259, '{"view":"city","smoking":false,"balcony":true}',  NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '204', 'family',   '2', 'double_queen', 'vacant_clean', 299, '{"view":"ocean","smoking":false,"crib":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '205', 'family',   '2', 'double_queen', 'occupied',     299, '{"view":"ocean","smoking":false,"crib":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '206', 'deluxe',   '2', 'king',   'maintenance',   279, '{"view":"city","smoking":false}',    NOW(), NOW()),
    -- Floor 3: Premium + Presidential
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '301', 'premium',  '3', 'king',   'vacant_clean',  399, '{"view":"panoramic","smoking":false,"butler":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '302', 'premium',  '3', 'king',   'occupied',      399, '{"view":"panoramic","smoking":false,"butler":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '303', 'premium',  '3', 'queen',  'occupied',      379, '{"view":"city","smoking":false,"butler":true}',   NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '304', 'premium',  '3', 'king',   'vacant_dirty',  399, '{"view":"panoramic","smoking":false}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '305', 'presidential', '3', 'king', 'vacant_clean', 899, '{"view":"panoramic","smoking":false,"butler":true,"private_pool":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '306', 'presidential', '3', 'king', 'occupied',    899, '{"view":"panoramic","smoking":false,"butler":true,"private_pool":true}', NOW(), NOW()),
    -- Floor 4: Penthouse
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '401', 'penthouse', '4', 'king',  'vacant_clean', 1299, '{"view":"360","smoking":false,"butler":true,"private_pool":true,"chef":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '402', 'penthouse', '4', 'king',  'blocked',      1299, '{"view":"360","smoking":false,"butler":true,"private_pool":true,"chef":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', '403', 'penthouse', '4', 'king',  'occupied',     1299, '{"view":"360","smoking":false,"butler":true,"private_pool":true,"chef":true}', NOW(), NOW()),
    -- Staff / Service rooms
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'S01', 'service',  'B', 'twin',   'vacant_clean',   0, '{"staff_only":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'S02', 'service',  'B', 'twin',   'occupied',       0, '{"staff_only":true}', NOW(), NOW())
ON CONFLICT (tenant_id, property_id, number) DO NOTHING;

-- =====================================================================
-- 2. RESERVATIONS — Mix of current, upcoming, past, and cancelled
-- =====================================================================
INSERT INTO reservations (tenant_id, property_id, guest_name, email, phone, room_number, room_type, check_in, check_out, adults, children, status, source, total, balance, special_requests, vip, color, created_at, updated_at)
VALUES
    -- Current stays
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'John Mitchell',     'john.mitchell@email.com',    '+1-212-555-1001', '102', 'standard', '2026-05-15', '2026-05-18', 2, 0, 'checked_in',  'direct',   597, 0,  'Late arrival, please hold room', false, '#3b82f6', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'Sarah Chen',        'sarah.chen@email.com',       '+1-212-555-1002', '104', 'standard', '2026-05-14', '2026-05-17', 1, 0, 'checked_in',  'ota',      567, 0,  'Allergic to feathers',            true,  '#ef4444', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'Robert De Niro',    'rdeniro@email.com',          '+1-212-555-1003', '202', 'deluxe',   '2026-05-13', '2026-05-20', 2, 0, 'checked_in',  'direct',  1953, 0,  'VIP — no photography',            true,  '#f59e0b', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'Emma Wilson',       'emma.w@email.com',           '+1-212-555-1004', '205', 'family',   '2026-05-15', '2026-05-19', 2, 2, 'confirmed',   'direct',  1196, 0,  'Kids need cribs',                 false, '#10b981', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'Carlos Rodriguez',  'crodriguez@email.com',       '+1-212-555-1005', '302', 'premium',  '2026-05-14', '2026-05-16', 2, 0, 'checked_in',  'ota',      758, 0,  'Anniversary — champagne please',  true,  '#ec4899', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'Yuki Tanaka',       'yuki.t@email.com',           '+81-90-5551-0001', '306', 'presidential', '2026-05-10', '2026-05-20', 2, 1, 'checked_in', 'agent',  8990, 0, 'Japanese breakfast daily',      true,  '#8b5cf6', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'Ahmed Hassan',      'ahmed.h@email.com',          '+971-50-555-0001', '403', 'penthouse', '2026-05-12', '2026-05-22', 4, 0, 'checked_in', 'direct', 12990, 0, 'Private chef — kosher',         true,  '#6366f1', NOW(), NOW()),
    -- Upcoming arrivals
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'Lisa Park',         'lisa.park@email.com',        '+1-305-555-2001', '101', 'standard', '2026-05-16', '2026-05-19', 2, 0, 'confirmed',   'direct',   597, 597, '', false, '#3b82f6', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'Marcus Johnson',    'marcus.j@email.com',         '+1-212-555-1006', '201', 'deluxe',   '2026-05-17', '2026-05-21', 2, 0, 'confirmed',   'ota',     1116, 1116, 'High floor preferred',           false, '#3b82f6', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'Isabella Romano',   'isabella.r@email.com',       '+39-333-555-0001', '301', 'premium',  '2026-05-18', '2026-05-25', 2, 0, 'confirmed',   'direct',  2793, 2793, 'Honeymoon — rose petals',        true,  '#ec4899', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'David Kim',         'david.kim@email.com',        '+1-212-555-1007', '305', 'presidential', '2026-05-20', '2026-05-23', 2, 0, 'confirmed', 'agent',  2697, 2697, 'Airport pickup arranged',        false, '#f59e0b', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'Olivia Brown',      'olivia.brown@email.com',     '+1-212-555-1008', '401', 'penthouse', '2026-05-22', '2026-05-29', 6, 0, 'confirmed',  'direct', 9093, 9093, 'Corporate retreat — NDA signed',  true,  '#6366f1', NOW(), NOW()),
    -- Past stays
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'James Wilson',      'james.w@email.com',          '+1-212-555-1009', '103', 'standard', '2026-05-10', '2026-05-14', 1, 0, 'checked_out', 'direct',   756, 0,  '', false, '#6b7280', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'Maria Garcia',      'maria.g@email.com',          '+1-212-555-1010', '203', 'deluxe',   '2026-05-08', '2026-05-12', 2, 0, 'checked_out', 'ota',     1036, 0,  'Late checkout approved',         false, '#6b7280', NOW(), NOW()),
    -- Cancelled / No-show
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'Thomas Anderson',   'tanderson@email.com',        '+1-212-555-1011', '101', 'standard', '2026-05-12', '2026-05-15', 1, 0, 'cancelled',   'direct',   597, 0,  'Flight cancelled',               false, '#6b7280', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'main', 'Priya Sharma',      'priya.s@email.com',          '+91-98-555-0001', '204', 'family',   '2026-05-11', '2026-05-14', 2, 1, 'no_show',     'ota',      897, 897, '', false, '#6b7280', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- =====================================================================
-- 3. HOUSEKEEPING STAFF
-- =====================================================================
INSERT INTO housekeeping_staff (tenant_id, name, role, active_shift, max_rooms_per_day, phone, email, active, created_at, updated_at)
VALUES
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Maria Garcia',    'cleaner',     'day',     15, '+1-212-555-2001', 'maria.garcia@nexus.com',     true, NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'James Wilson',    'cleaner',     'day',     12, '+1-212-555-2002', 'james.wilson@nexus.com',     true, NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Ana Petrova',     'cleaner',     'evening', 10, '+1-212-555-2003', 'ana.petrova@nexus.com',      true, NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Liu Wei',         'inspector',   'day',     20, '+1-212-555-2004', 'liu.wei@nexus.com',          true, NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Fatima Al-Rashid','supervisor',  'day',     25, '+1-212-555-2005', 'fatima.alrashid@nexus.com',  true, NOW(), NOW())
ON CONFLICT DO NOTHING;

-- =====================================================================
-- 4. HOUSEKEEPING TASKS (generated from room status)
-- =====================================================================
INSERT INTO housekeeping_tasks (tenant_id, room_number, task_type, status, assigned_to, priority, notes, created_at, updated_at)
SELECT
    (SELECT id FROM tenants WHERE external_id = 'demo'),
    r.number,
    CASE WHEN r.status = 'vacant_dirty' THEN 'full_clean'
         WHEN r.status = 'occupied' THEN 'full_clean'
         ELSE 'inspection' END,
    CASE WHEN r.status = 'vacant_clean' THEN 'completed'
         WHEN r.status = 'maintenance' THEN 'blocked'
         ELSE 'pending' END,
    (SELECT id FROM housekeeping_staff WHERE tenant_id = (SELECT id FROM tenants WHERE external_id = 'demo') ORDER BY random() LIMIT 1),
    CASE WHEN r.type IN ('premium','presidential','penthouse') THEN 'high' ELSE 'normal' END,
    CASE WHEN r.status = 'maintenance' THEN 'Waiting for maintenance team — plumbing issue'
         WHEN r.type = 'presidential' THEN 'VIP arrival prep — extra attention required'
         ELSE '' END,
    NOW(),
    NOW()
FROM rooms r
WHERE r.tenant_id = (SELECT id FROM tenants WHERE external_id = 'demo')
  AND r.deleted_at IS NULL
  AND r.type != 'service'
ON CONFLICT DO NOTHING;

-- =====================================================================
-- 5. WIFI ACCESS POINTS
-- =====================================================================
INSERT INTO wifi_access_points (tenant_id, name, floor, location, mac_address, ip_address, model, status, firmware_version, config, created_at, updated_at)
VALUES
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'AP-Lobby-01',    'Lobby',  'Main lobby ceiling',         'aa:bb:cc:dd:ee:01', '10.0.1.101', 'Ubiquiti U6-Pro',    'online',  '6.5.50', '{"channel":36,"bandwidth":80,"power":20}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'AP-Lobby-02',    'Lobby',  'Reception desk',             'aa:bb:cc:dd:ee:02', '10.0.1.102', 'Ubiquiti U6-Pro',    'online',  '6.5.50', '{"channel":40,"bandwidth":80,"power":18}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'AP-F1-Corridor', '1',      'Floor 1 east corridor',      'aa:bb:cc:dd:ee:03', '10.0.1.103', 'Ubiquiti U6-Lite',   'online',  '6.5.48', '{"channel":1,"bandwidth":20,"power":22}',  NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'AP-F1-West',     '1',      'Floor 1 west wing',          'aa:bb:cc:dd:ee:04', '10.0.1.104', 'Ubiquiti U6-Lite',   'online',  '6.5.48', '{"channel":6,"bandwidth":20,"power":22}',  NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'AP-F2-East',     '2',      'Floor 2 east corridor',      'aa:bb:cc:dd:ee:05', '10.0.1.105', 'Ubiquiti U6-Pro',    'online',  '6.5.50', '{"channel":36,"bandwidth":80,"power":20}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'AP-F2-West',     '2',      'Floor 2 west wing',          'aa:bb:cc:dd:ee:06', '10.0.1.106', 'Ubiquiti U6-Pro',    'warning', '6.5.47', '{"channel":40,"bandwidth":80,"power":20}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'AP-F3-East',     '3',      'Floor 3 east corridor',      'aa:bb:cc:dd:ee:07', '10.0.1.107', 'Ubiquiti U6-Pro',    'online',  '6.5.50', '{"channel":44,"bandwidth":80,"power":20}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'AP-F3-West',     '3',      'Floor 3 west wing',          'aa:bb:cc:dd:ee:08', '10.0.1.108', 'Ubiquiti U6-Pro',    'online',  '6.5.50', '{"channel":48,"bandwidth":80,"power":20}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'AP-F4-Center',   '4',      'Penthouse floor center',     'aa:bb:cc:dd:ee:09', '10.0.1.109', 'Ubiquiti U6-Enterprise', 'online', '6.5.50', '{"channel":52,"bandwidth":160,"power":22}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'AP-Poolside',    'Pool',   'Pool deck east',             'aa:bb:cc:dd:ee:10', '10.0.1.110', 'Ubiquiti U6-Mesh',   'offline', '6.5.45', '{"channel":11,"bandwidth":20,"power":24}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'AP-Gym',         'Gym',    'Fitness center ceiling',     'aa:bb:cc:dd:ee:11', '10.0.1.111', 'Ubiquiti U6-Lite',   'online',  '6.5.50', '{"channel":1,"bandwidth":20,"power":20}',  NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'AP-Spa',         'Spa',    'Spa reception',              'aa:bb:cc:dd:ee:12', '10.0.1.112', 'Ubiquiti U6-Lite',   'online',  '6.5.50', '{"channel":6,"bandwidth":20,"power":20}',  NOW(), NOW())
ON CONFLICT (tenant_id, mac_address) DO NOTHING;

-- =====================================================================
-- 6. WIFI ALERTS (generated from AP status)
-- =====================================================================
INSERT INTO wifi_alerts (tenant_id, ap_id, alert_type, severity, message, resolved, created_at, updated_at)
SELECT
    (SELECT id FROM tenants WHERE external_id = 'demo'),
    ap.id,
    CASE WHEN ap.status = 'offline' THEN 'device_offline'
         WHEN ap.firmware_version < '6.5.50' THEN 'firmware_outdated'
         WHEN ap.status = 'warning' THEN 'signal_degraded'
         ELSE 'interference_detected' END,
    CASE WHEN ap.status = 'offline' THEN 'critical'
         WHEN ap.firmware_version < '6.5.50' THEN 'high'
         ELSE 'medium' END,
    CASE WHEN ap.status = 'offline' THEN 'Access point unreachable for > 15 minutes'
         WHEN ap.firmware_version < '6.5.50' THEN 'Firmware ' || ap.firmware_version || ' has known security vulnerabilities'
         WHEN ap.status = 'warning' THEN 'Signal strength below threshold on 5GHz band'
         ELSE 'Channel congestion detected — recommend switching to channel 149' END,
    false,
    NOW() - (random() * interval '6 hours'),
    NOW()
FROM wifi_access_points ap
WHERE ap.tenant_id = (SELECT id FROM tenants WHERE external_id = 'demo')
  AND (ap.status != 'online' OR ap.firmware_version < '6.5.50')
ON CONFLICT DO NOTHING;

-- =====================================================================
-- 7. IPTV CHANNELS
-- =====================================================================
INSERT INTO iptv_channels (tenant_id, name, category, number, url, icon_url, is_active, is_premium, language, config, created_at, updated_at)
VALUES
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'CNN International',    'news',       1,  'https://cnn-international.stream/live.m3u8',     '', true, false, 'en', '{}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'BBC World News',       'news',       2,  'https://bbc-world.stream/live.m3u8',            '', true, false, 'en', '{}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Al Jazeera English',   'news',       3,  'https://aljazeera.stream/live.m3u8',           '', true, false, 'en', '{}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'ESPN',                 'sports',     10, 'https://espn.stream/live.m3u8',                '', true, true,  'en', '{"hd":true,"dolby":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Sky Sports',           'sports',     11, 'https://skysports.stream/live.m3u8',           '', true, true,  'en', '{"hd":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'HBO',                  'movies',     20, 'https://hbo.stream/live.m3u8',                 '', true, true,  'en', '{"hd":true,"dolby":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Netflix Hotel',        'movies',     21, 'https://netflix-hotel.stream/live.m3u8',       '', true, true,  'en', '{"4k":true,"dolby_vision":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Cartoon Network',      'kids',       30, 'https://cn.stream/live.m3u8',                    '', true, false, 'en', '{}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Disney Channel',       'kids',       31, 'https://disney.stream/live.m3u8',              '', true, false, 'en', '{}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Local Guide',          'hotel',      50, 'https://hotel-guide.stream/welcome.m3u8',       '', true, false, 'en', '{"loop":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Room Service Menu',    'hotel',      51, 'https://hotel-menu.stream/menu.m3u8',          '', true, false, 'en', '{"interactive":true}', NOW(), NOW()),
    ((SELECT id FROM tenants WHERE external_id = 'demo'), 'Spa & Wellness',       'hotel',      52, 'https://hotel-spa.stream/wellness.m3u8',       '', true, true,  'en', '{"booking_integration":true}', NOW(), NOW())
ON CONFLICT DO NOTHING;

-- =====================================================================
-- 8. SMART LOCKS (generated per room)
-- =====================================================================
INSERT INTO smart_locks (tenant_id, device_id, room_number, status, battery_level, firmware_version, config, created_at, updated_at)
SELECT
    (SELECT id FROM tenants WHERE external_id = 'demo'),
    'LOCK-' || r.number,
    r.number,
    CASE WHEN r.status IN ('occupied','checked_in') THEN 'locked'
         WHEN r.status = 'maintenance' THEN 'maintenance'
         ELSE 'unlocked' END,
    (70 + (random() * 30))::int,
    '2.4.1',
    '{"auto_lock":true,"auto_lock_delay":30,"master_code_enabled":true,"audit_log":true}',
    NOW(),
    NOW()
FROM rooms r
WHERE r.tenant_id = (SELECT id FROM tenants WHERE external_id = 'demo')
  AND r.deleted_at IS NULL
  AND r.type != 'service'
ON CONFLICT (tenant_id, device_id) DO NOTHING;

-- =====================================================================
-- 9. WIFI FLOOR SUMMARIES (trigger auto-populates, but seed some initial metrics)
-- =====================================================================
INSERT INTO wifi_floor_summaries (tenant_id, floor, ap_count, online_count, offline_count, avg_signal, avg_clients, avg_throughput, last_updated)
SELECT
    (SELECT id FROM tenants WHERE external_id = 'demo'),
    ap.floor,
    COUNT(*)::int,
    COUNT(*) FILTER (WHERE ap.status = 'online')::int,
    COUNT(*) FILTER (WHERE ap.status = 'offline')::int,
    (70 + random() * 25)::int,
    (15 + random() * 40)::int,
    (50 + random() * 150)::int,
    NOW()
FROM wifi_access_points ap
WHERE ap.tenant_id = (SELECT id FROM tenants WHERE external_id = 'demo')
GROUP BY ap.floor
ON CONFLICT (tenant_id, floor) DO UPDATE SET
    ap_count = EXCLUDED.ap_count,
    online_count = EXCLUDED.online_count,
    offline_count = EXCLUDED.offline_count,
    avg_signal = EXCLUDED.avg_signal,
    avg_clients = EXCLUDED.avg_clients,
    avg_throughput = EXCLUDED.avg_throughput,
    last_updated = NOW();
