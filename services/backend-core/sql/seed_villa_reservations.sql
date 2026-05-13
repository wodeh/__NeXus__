-- Seed villa reservations for IPTV welcome screen testing
-- Add active reservations for today so the welcome screen has real data

-- Disable RLS for seeding
ALTER TABLE IF EXISTS villa_reservations DISABLE ROW LEVEL SECURITY;

-- Reservation for Green Villa (c4eebc99...) — active today
INSERT INTO villa_reservations (
    id, tenant_id, villa_id, villa_name, guest_name, guest_phone, guest_email,
    guest_count, check_in_date, check_out_date, nights, total_amount, currency,
    status, source, internal_notes, balance_due, balance_paid, created_at, updated_at
) VALUES (
    'd1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'c4eebc99-9c0b-4ef8-bb6d-6bb9bd380a44',
    'Green Villa',
    'John Smith',
    '+1-555-0101',
    'john.smith@email.com',
    4,
    (CURRENT_DATE - INTERVAL '2 days')::date,
    (CURRENT_DATE + INTERVAL '5 days')::date,
    7,
    2800.00,
    'USD',
    'reserved',
    'website',
    'Honeymoon trip, requested late checkout',
    2800.00,
    false,
    NOW(),
    NOW()
)
ON CONFLICT DO NOTHING;

-- Reservation for Blue Villa (c5eebc99...) — active today
INSERT INTO villa_reservations (
    id, tenant_id, villa_id, villa_name, guest_name, guest_phone, guest_email,
    guest_count, check_in_date, check_out_date, nights, total_amount, currency,
    status, source, internal_notes, balance_due, balance_paid, created_at, updated_at
) VALUES (
    'd2eebc99-9c0b-4ef8-bb6d-6bb9bd380a22',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'c5eebc99-9c0b-4ef8-bb6d-6bb9bd380a55',
    'Blue Villa',
    'Maria Garcia',
    '+1-555-0202',
    'maria.g@email.com',
    6,
    (CURRENT_DATE - INTERVAL '1 day')::date,
    (CURRENT_DATE + INTERVAL '3 days')::date,
    4,
    2200.00,
    'USD',
    'reserved',
    'phone',
    'Family vacation, needs baby cot',
    2200.00,
    false,
    NOW(),
    NOW()
)
ON CONFLICT DO NOTHING;

-- Reservation for Gold Villa (c6eebc99...) — active today
INSERT INTO villa_reservations (
    id, tenant_id, villa_id, villa_name, guest_name, guest_phone, guest_email,
    guest_count, check_in_date, check_out_date, nights, total_amount, currency,
    status, source, internal_notes, balance_due, balance_paid, created_at, updated_at
) VALUES (
    'd3eebc99-9c0b-4ef8-bb6d-6bb9bd380a33',
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'c6eebc99-9c0b-4ef8-bb6d-6bb9bd380a66',
    'Gold Villa',
    'Ahmed Hassan',
    '+1-555-0303',
    'ahmed.h@email.com',
    8,
    CURRENT_DATE::date,
    (CURRENT_DATE + INTERVAL '7 days')::date,
    7,
    5600.00,
    'USD',
    'reserved',
    'walkin',
    'VIP guest, arrange airport pickup',
    5600.00,
    false,
    NOW(),
    NOW()
)
ON CONFLICT DO NOTHING;

-- Re-enable RLS
ALTER TABLE IF EXISTS villa_reservations ENABLE ROW LEVEL SECURITY;

-- Verify
SELECT villa_name, guest_name, check_in_date, check_out_date, status
FROM villa_reservations
WHERE tenant_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
AND check_in_date <= CURRENT_DATE AND check_out_date > CURRENT_DATE;
