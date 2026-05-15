CREATE TABLE IF NOT EXISTS waitlist (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id UUID,
    guest_name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    adults INT NOT NULL DEFAULT 1,
    children INT NOT NULL DEFAULT 0,
    room_type TEXT,
    requested_check_in DATE NOT NULL,
    requested_check_out DATE NOT NULL,
    priority INT NOT NULL DEFAULT 0,
    notes TEXT,
    status TEXT NOT NULL DEFAULT 'waiting',
    assigned_room_number TEXT,
    assigned_reservation_id UUID REFERENCES reservations(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_waitlist_tenant ON waitlist(tenant_id);
CREATE INDEX IF NOT EXISTS idx_waitlist_status ON waitlist(status);
CREATE INDEX IF NOT EXISTS idx_waitlist_dates ON waitlist(requested_check_in, requested_check_out);

ALTER TABLE waitlist ENABLE ROW LEVEL SECURITY;

CREATE POLICY waitlist_tenant_isolation ON waitlist
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID OR current_setting('app.current_tenant', true) = '');

CREATE OR REPLACE FUNCTION update_waitlist_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_waitlist_updated_at ON waitlist;
CREATE TRIGGER trg_waitlist_updated_at
    BEFORE UPDATE ON waitlist
    FOR EACH ROW
    EXECUTE FUNCTION update_waitlist_updated_at();
