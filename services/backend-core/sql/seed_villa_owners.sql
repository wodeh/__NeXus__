-- Seed script for villa owners and their properties
-- Run: psql -U nexus -d nexus_platform -f seed_villa_owners.sql

-- Disable RLS for seeding (re-enable after)
ALTER TABLE IF EXISTS tenants DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS users DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS villa_properties DISABLE ROW LEVEL SECURITY;

-- Create a tenant for villa operations
INSERT INTO tenants (id, external_id, name, region, tier, config)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'villa-owners',
    'Villa Owners Collective',
    'us-east-1',
    'enterprise',
    '{"property_type": "villa", "module": "villa_rental"}'
)
ON CONFLICT (external_id) DO NOTHING;

-- Create 4 villa owner users (password: 'villa123' bcrypt hashed)
INSERT INTO users (id, tenant_id, email, password_hash, name, role, is_active)
VALUES
    ('b1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'owner1@villa.test', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Ahmad Khalil', 'manager', true),
    ('b2eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'owner2@villa.test', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Sarah Nassar', 'manager', true),
    ('b3eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'ramiz@villa.test', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Ramiz Haddad', 'manager', true),
    ('b4eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'owner4@villa.test', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'Layla Farhat', 'manager', true)
ON CONFLICT (email) DO NOTHING;

-- Villa 1: Owner 1 (Ahmad Khalil) — 1 villa
INSERT INTO villa_properties (id, tenant_id, name, description, address, city, country, bedrooms, bathrooms, max_guests, amenities, price_per_night, currency, cleaning_fee, security_deposit, is_active, status)
VALUES
    ('c1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Villa Al-Mashta', 'Luxury mountain villa with panoramic views and private pool', 'Mount of Olives Road 42', 'Ramallah', 'Palestine', 4, 3, 8, ARRAY['wifi', 'pool', 'parking', 'ac', 'kitchen', 'bbq'], 350, 'USD', 75, 500, true, 'available')
ON CONFLICT DO NOTHING;

-- Villa 2-3: Owner 2 (Sarah Nassar) — 2 villas
INSERT INTO villa_properties (id, tenant_id, name, description, address, city, country, bedrooms, bathrooms, max_guests, amenities, price_per_night, currency, cleaning_fee, security_deposit, is_active, status)
VALUES
    ('c2eebc99-9c0b-4ef8-bb6d-6bb9bd380a22', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Chalet Al-Balad', 'Cozy chalet in historic district with modern amenities', 'Star Street 15', 'Bethlehem', 'Palestine', 3, 2, 6, ARRAY['wifi', 'parking', 'ac', 'kitchen', 'fireplace'], 250, 'USD', 50, 300, true, 'available'),
    ('c3eebc99-9c0b-4ef8-bb6d-6bb9bd380a33', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Sea Breeze Cottage', 'Beachfront cottage with direct sea access and sunset views', 'Coastal Road 88', 'Gaza', 'Palestine', 2, 1, 4, ARRAY['wifi', 'beach', 'parking', 'ac', 'kitchen'], 180, 'USD', 40, 200, true, 'available')
ON CONFLICT DO NOTHING;

-- Villa 4-6: Owner 3 (Ramiz Haddad) — 3 villas: Green, Blue, Gold
INSERT INTO villa_properties (id, tenant_id, name, description, address, city, country, bedrooms, bathrooms, max_guests, amenities, price_per_night, currency, cleaning_fee, security_deposit, is_active, status)
VALUES
    ('c4eebc99-9c0b-4ef8-bb6d-6bb9bd380a44', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Green Villa', 'Eco-friendly villa surrounded by olive groves with sustainable design', 'Olive Grove Lane 7', 'Nablus', 'Palestine', 5, 4, 10, ARRAY['wifi', 'pool', 'parking', 'ac', 'kitchen', 'solar', 'garden', 'bbq'], 400, 'USD', 100, 600, true, 'available'),
    ('c5eebc99-9c0b-4ef8-bb6d-6bb9bd380a55', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Blue Villa', 'Modern waterfront villa with infinity pool and smart home features', 'Azure Bay Drive 23', 'Haifa', 'Palestine', 6, 5, 12, ARRAY['wifi', 'pool', 'parking', 'ac', 'kitchen', 'smart_tv', 'gym', 'bbq', 'jacuzzi'], 550, 'USD', 125, 800, true, 'available'),
    ('c6eebc99-9c0b-4ef8-bb6d-6bb9bd380a66', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Gold Villa', 'Premium luxury villa with gold-themed interiors and butler service', 'Crown Heights 1', 'Jerusalem', 'Palestine', 7, 6, 14, ARRAY['wifi', 'pool', 'parking', 'ac', 'kitchen', 'smart_tv', 'gym', 'spa', 'butler', 'helipad'], 800, 'USD', 200, 1000, true, 'available')
ON CONFLICT DO NOTHING;

-- Villa 7-10: Owner 4 (Layla Farhat) — 4 villas
INSERT INTO villa_properties (id, tenant_id, name, description, address, city, country, bedrooms, bathrooms, max_guests, amenities, price_per_night, currency, cleaning_fee, security_deposit, is_active, status)
VALUES
    ('c7eebc99-9c0b-4ef8-bb6d-6bb9bd380a77', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Desert Rose Retreat', 'Serene desert retreat with stargazing deck and camel rides', 'Wadi Rum Trail 5', 'Hebron', 'Palestine', 3, 2, 6, ARRAY['wifi', 'parking', 'ac', 'kitchen', 'stargazing', 'campfire'], 220, 'USD', 60, 250, true, 'available'),
    ('c8eebc99-9c0b-4ef8-bb6d-6bb9bd380a88', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Garden Palace', 'Victorian-style garden villa with rose gardens and tea house', 'Rose Avenue 12', 'Jericho', 'Palestine', 4, 3, 8, ARRAY['wifi', 'pool', 'parking', 'ac', 'kitchen', 'garden', 'tea_house'], 320, 'USD', 80, 400, true, 'available'),
    ('c9eebc99-9c0b-4ef8-bb6d-6bb9bd380a99', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Skyline Penthouse', 'Rooftop penthouse with 360° city views and private elevator', 'Tower Plaza 50', 'Ramallah', 'Palestine', 3, 2, 6, ARRAY['wifi', 'parking', 'ac', 'kitchen', 'smart_tv', 'gym', 'elevator'], 380, 'USD', 90, 450, true, 'available'),
    ('caeebc99-9c0b-4ef8-bb6d-6bb9bd380aaa', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'Forest Hideaway', 'Secluded forest cabin with hiking trails and wildlife viewing', 'Pine Trail 99', 'Nablus', 'Palestine', 2, 1, 4, ARRAY['wifi', 'parking', 'kitchen', 'fireplace', 'hiking'], 150, 'USD', 35, 150, true, 'available')
ON CONFLICT DO NOTHING;

-- Re-enable RLS
ALTER TABLE IF EXISTS tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS users ENABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS villa_properties ENABLE ROW LEVEL SECURITY;

-- Verify counts
SELECT 'Users created:' AS label, COUNT(*) AS count FROM users WHERE tenant_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
UNION ALL
SELECT 'Villas created:' AS label, COUNT(*) AS count FROM villa_properties WHERE tenant_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';
