-- PMS Domain Foundation Migration
-- Creates core tables for Reservation, Guest, and Folio aggregates.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ==================== GUESTS ====================

CREATE TABLE IF NOT EXISTS guests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    property_id UUID NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    loyalty_member_id VARCHAR(100),
    preferences JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_guests_tenant ON guests(tenant_id);
CREATE INDEX IF NOT EXISTS idx_guests_email ON guests(email) WHERE email IS NOT NULL;

ALTER TABLE guests ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'guests' AND policyname = 'guests_tenant_isolation'
    ) THEN
        CREATE POLICY guests_tenant_isolation ON guests
        USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== RESERVATIONS ====================

CREATE TABLE IF NOT EXISTS reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    property_id UUID NOT NULL,
    guest_id UUID NOT NULL REFERENCES guests(id),
    room_id UUID,
    confirmation_number VARCHAR(50) UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    check_in_date DATE NOT NULL,
    check_out_date DATE NOT NULL,
    special_requests TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_reservations_tenant ON reservations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_reservations_guest ON reservations(guest_id);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(tenant_id, status);

ALTER TABLE reservations ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'reservations' AND policyname = 'reservations_tenant_isolation'
    ) THEN
        CREATE POLICY reservations_tenant_isolation ON reservations
        USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== FOLIOS ====================

CREATE TABLE IF NOT EXISTS folios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    property_id UUID NOT NULL,
    reservation_id UUID NOT NULL REFERENCES reservations(id),
    guest_id UUID NOT NULL REFERENCES guests(id),
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    balance DECIMAL(15,2) NOT NULL DEFAULT 0,
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_folios_tenant ON folios(tenant_id);
CREATE INDEX IF NOT EXISTS idx_folios_reservation ON folios(reservation_id);

ALTER TABLE folios ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'folios' AND policyname = 'folios_tenant_isolation'
    ) THEN
        CREATE POLICY folios_tenant_isolation ON folios
        USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== FOLIO TRANSACTIONS ====================

CREATE TABLE IF NOT EXISTS folio_transactions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    folio_id UUID NOT NULL REFERENCES folios(id),
    transaction_type VARCHAR(50) NOT NULL,
    description VARCHAR(255) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    posting_date DATE NOT NULL,
    posted_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_voided BOOLEAN DEFAULT FALSE,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_folio_transactions_folio ON folio_transactions(folio_id, posting_date DESC);

ALTER TABLE folio_transactions ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'folio_transactions' AND policyname = 'folio_transactions_tenant_isolation'
    ) THEN
        CREATE POLICY folio_transactions_tenant_isolation ON folio_transactions
        USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== TRIGGERS ====================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY['guests', 'reservations', 'folios', 'folio_transactions'];
BEGIN
    FOREACH tbl IN ARRAY tables
    LOOP
        IF NOT EXISTS (
            SELECT 1 FROM pg_trigger
            WHERE tgname = 'update_' || tbl || '_updated_at'
        ) THEN
            EXECUTE format('
                CREATE TRIGGER update_%I_updated_at
                BEFORE UPDATE ON %I
                FOR EACH ROW
                EXECUTE FUNCTION update_updated_at_column();
            ', tbl, tbl);
        END IF;
    END LOOP;
END $$;
