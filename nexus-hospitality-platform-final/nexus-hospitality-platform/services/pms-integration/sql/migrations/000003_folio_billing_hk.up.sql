-- Folio & Billing Core Migration
-- Adds financial tracking: charges, payments, adjustments, folios, ledger entries

-- ==================== FOLIOS ====================

CREATE TABLE IF NOT EXISTS folios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    reservation_id UUID NOT NULL REFERENCES reservations(id),
    guest_id UUID NOT NULL REFERENCES guests(id),
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    balance DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    total_charges DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    total_payments DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    total_adjustments DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    is_master BOOLEAN NOT NULL DEFAULT FALSE,
    master_folio_id UUID REFERENCES folios(id),
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_folios_tenant ON folios(tenant_id);
CREATE INDEX IF NOT EXISTS idx_folios_reservation ON folios(reservation_id);
CREATE INDEX IF NOT EXISTS idx_folios_guest ON folios(guest_id);
CREATE INDEX IF NOT EXISTS idx_folios_status ON folios(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_folios_master ON folios(master_folio_id) WHERE master_folio_id IS NOT NULL;

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

-- ==================== CHARGES ====================

CREATE TABLE IF NOT EXISTS charges (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    folio_id UUID NOT NULL REFERENCES folios(id),
    reservation_id UUID REFERENCES reservations(id),
    room_id UUID REFERENCES rooms(id),
    charge_type VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT 'room_charge',
    description TEXT NOT NULL,
    quantity DECIMAL(10,2) NOT NULL DEFAULT 1.00,
    unit_price DECIMAL(15,2) NOT NULL,
    total_amount DECIMAL(15,2) NOT NULL,
    tax_amount DECIMAL(15,2) NOT NULL DEFAULT 0.00,
    tax_rate DECIMAL(5,4) NOT NULL DEFAULT 0.0000,
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    charge_date DATE NOT NULL DEFAULT CURRENT_DATE,
    posting_date TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    posted_by VARCHAR(255),
    is_revenue BOOLEAN NOT NULL DEFAULT TRUE,
    revenue_center VARCHAR(100),
    source VARCHAR(50) NOT NULL DEFAULT 'manual',
    is_voided BOOLEAN NOT NULL DEFAULT FALSE,
    voided_at TIMESTAMPTZ,
    voided_by VARCHAR(255),
    void_reason TEXT,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_charges_tenant ON charges(tenant_id);
CREATE INDEX IF NOT EXISTS idx_charges_folio ON charges(folio_id);
CREATE INDEX IF NOT EXISTS idx_charges_reservation ON charges(reservation_id);
CREATE INDEX IF NOT EXISTS idx_charges_date ON charges(charge_date);
CREATE INDEX IF NOT EXISTS idx_charges_type ON charges(tenant_id, charge_type);
CREATE INDEX IF NOT EXISTS idx_charges_voided ON charges(is_voided) WHERE is_voided = TRUE;

ALTER TABLE charges ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'charges' AND policyname = 'charges_tenant_isolation'
    ) THEN
        CREATE POLICY charges_tenant_isolation ON charges
            USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== PAYMENTS ====================

CREATE TABLE IF NOT EXISTS payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    folio_id UUID NOT NULL REFERENCES folios(id),
    reservation_id UUID REFERENCES reservations(id),
    guest_id UUID REFERENCES guests(id),
    payment_method VARCHAR(100) NOT NULL,
    payment_type VARCHAR(100) NOT NULL DEFAULT 'charge',
    amount DECIMAL(15,2) NOT NULL,
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    exchange_rate DECIMAL(15,6) NOT NULL DEFAULT 1.000000,
    reference_number VARCHAR(255),
    transaction_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'completed',
    processor VARCHAR(100),
    last_four_digits VARCHAR(4),
    card_brand VARCHAR(50),
    expiry_month INT,
    expiry_year INT,
    authorization_code VARCHAR(255),
    captured_at TIMESTAMPTZ,
    refunded_at TIMESTAMPTZ,
    refund_amount DECIMAL(15,2),
    refund_reason TEXT,
    processed_by VARCHAR(255),
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_payments_tenant ON payments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_payments_folio ON payments(folio_id);
CREATE INDEX IF NOT EXISTS idx_payments_reservation ON payments(reservation_id);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_payments_method ON payments(tenant_id, payment_method);
CREATE INDEX IF NOT EXISTS idx_payments_refunded ON payments(refunded_at) WHERE refunded_at IS NOT NULL;

ALTER TABLE payments ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'payments' AND policyname = 'payments_tenant_isolation'
    ) THEN
        CREATE POLICY payments_tenant_isolation ON payments
            USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== FOLIO ADJUSTMENTS ====================

CREATE TABLE IF NOT EXISTS folio_adjustments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    folio_id UUID NOT NULL REFERENCES folios(id),
    charge_id UUID REFERENCES charges(id),
    adjustment_type VARCHAR(100) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    currency_code CHAR(3) NOT NULL DEFAULT 'USD',
    reason TEXT NOT NULL,
    approved_by VARCHAR(255),
    approved_at TIMESTAMPTZ,
    is_approved BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_folio_adjustments_tenant ON folio_adjustments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_folio_adjustments_folio ON folio_adjustments(folio_id);
CREATE INDEX IF NOT EXISTS idx_folio_adjustments_approved ON folio_adjustments(is_approved);

ALTER TABLE folio_adjustments ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'folio_adjustments' AND policyname = 'folio_adjustments_tenant_isolation'
    ) THEN
        CREATE POLICY folio_adjustments_tenant_isolation ON folio_adjustments
            USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== HOUSEKEEPING TASKS ====================

CREATE TABLE IF NOT EXISTS housekeeping_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    property_id UUID NOT NULL REFERENCES properties(id),
    room_id UUID NOT NULL REFERENCES rooms(id),
    task_type VARCHAR(100) NOT NULL,
    priority VARCHAR(50) NOT NULL DEFAULT 'normal',
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    assigned_to VARCHAR(255),
    scheduled_at TIMESTAMPTZ,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    estimated_duration_minutes INT,
    actual_duration_minutes INT,
    notes TEXT,
    inspection_score INT,
    inspection_notes TEXT,
    inspected_by VARCHAR(255),
    created_by VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_hk_tasks_tenant ON housekeeping_tasks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_hk_tasks_property ON housekeeping_tasks(property_id);
CREATE INDEX IF NOT EXISTS idx_hk_tasks_room ON housekeeping_tasks(room_id);
CREATE INDEX IF NOT EXISTS idx_hk_tasks_status ON housekeeping_tasks(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_hk_tasks_assigned ON housekeeping_tasks(assigned_to);
CREATE INDEX IF NOT EXISTS idx_hk_tasks_scheduled ON housekeeping_tasks(scheduled_at);
CREATE INDEX IF NOT EXISTS idx_hk_tasks_priority ON housekeeping_tasks(priority);

ALTER TABLE housekeeping_tasks ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'housekeeping_tasks' AND policyname = 'hk_tasks_tenant_isolation'
    ) THEN
        CREATE POLICY hk_tasks_tenant_isolation ON housekeeping_tasks
            USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== MAINTENANCE WORK ORDERS ====================

CREATE TABLE IF NOT EXISTS work_orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    property_id UUID NOT NULL REFERENCES properties(id),
    room_id UUID REFERENCES rooms(id),
    asset_id UUID,
    work_order_number VARCHAR(100) NOT NULL,
    category VARCHAR(100) NOT NULL,
    priority VARCHAR(50) NOT NULL DEFAULT 'medium',
    status VARCHAR(50) NOT NULL DEFAULT 'open',
    title VARCHAR(255) NOT NULL,
    description TEXT,
    reported_by VARCHAR(255),
    assigned_to VARCHAR(255),
    estimated_cost DECIMAL(15,2),
    actual_cost DECIMAL(15,2),
    scheduled_date DATE,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    resolution_notes TEXT,
    parts_used JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    version INT NOT NULL DEFAULT 1,
    deleted_at TIMESTAMPTZ,
    UNIQUE(tenant_id, work_order_number)
);

CREATE INDEX IF NOT EXISTS idx_work_orders_tenant ON work_orders(tenant_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_property ON work_orders(property_id);
CREATE INDEX IF NOT EXISTS idx_work_orders_status ON work_orders(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_work_orders_priority ON work_orders(priority);
CREATE INDEX IF NOT EXISTS idx_work_orders_assigned ON work_orders(assigned_to);

ALTER TABLE work_orders ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'work_orders' AND policyname = 'work_orders_tenant_isolation'
    ) THEN
        CREATE POLICY work_orders_tenant_isolation ON work_orders
            USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== AUDIT LOG ====================

CREATE TABLE IF NOT EXISTS audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL,
    entity_type VARCHAR(100) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    performed_by VARCHAR(255),
    performed_by_id UUID,
    ip_address INET,
    user_agent TEXT,
    old_values JSONB,
    new_values JSONB,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_tenant ON audit_log(tenant_id);
CREATE INDEX IF NOT EXISTS idx_audit_entity ON audit_log(entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_log(action);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_log(created_at);

ALTER TABLE audit_log ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'audit_log' AND policyname = 'audit_log_tenant_isolation'
    ) THEN
        CREATE POLICY audit_log_tenant_isolation ON audit_log
            USING (tenant_id = current_setting('app.current_tenant', true)::UUID);
    END IF;
END $$;

-- ==================== TRIGGERS ====================

DO $$
DECLARE
    tbl TEXT;
    tables TEXT[] := ARRAY['folios', 'charges', 'payments', 'folio_adjustments', 'housekeeping_tasks', 'work_orders'];
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

-- ==================== FOLIO BALANCE FUNCTION ====================

CREATE OR REPLACE FUNCTION recalculate_folio_balance(folio_id UUID)
RETURNS VOID AS $$
BEGIN
    UPDATE folios
    SET total_charges = (
        SELECT COALESCE(SUM(total_amount), 0)
        FROM charges
        WHERE charges.folio_id = $1 AND is_voided = FALSE
    ),
    total_payments = (
        SELECT COALESCE(SUM(amount), 0)
        FROM payments
        WHERE payments.folio_id = $1 AND status = 'completed' AND refunded_at IS NULL
    ),
    total_adjustments = (
        SELECT COALESCE(SUM(amount), 0)
        FROM folio_adjustments
        WHERE folio_adjustments.folio_id = $1 AND is_approved = TRUE
    ),
    balance = total_charges - total_payments - total_adjustments,
    updated_at = NOW()
    WHERE id = $1;
END;
$$ LANGUAGE plpgsql;

-- ==================== HOUSEKEEPING PREDICTIVE VIEW ====================

CREATE OR REPLACE VIEW v_housekeeping_board AS
SELECT
    t.id AS task_id,
    t.tenant_id,
    t.property_id,
    p.name AS property_name,
    t.room_id,
    r.room_number,
    r.floor,
    rt.name AS room_type_name,
    t.task_type,
    t.priority,
    t.status,
    t.assigned_to,
    t.scheduled_at,
    t.started_at,
    t.completed_at,
    t.estimated_duration_minutes,
    t.actual_duration_minutes,
    t.inspection_score,
    res.id AS current_reservation_id,
    res.check_out_date AS expected_check_out,
    res.check_in_date AS expected_check_in,
    CASE
        WHEN t.status = 'completed' THEN 'done'
        WHEN t.status = 'in_progress' THEN 'working'
        WHEN t.scheduled_at < NOW() THEN 'overdue'
        WHEN res.check_out_date = CURRENT_DATE THEN 'check_out_today'
        WHEN res.check_in_date = CURRENT_DATE THEN 'check_in_today'
        ELSE 'standard'
    END AS board_category,
    t.created_at,
    t.updated_at
FROM housekeeping_tasks t
JOIN rooms r ON t.room_id = r.id
JOIN properties p ON t.property_id = p.id
LEFT JOIN room_types rt ON r.room_type_id = rt.id
LEFT JOIN reservations res ON res.room_id = r.id
    AND res.status IN ('confirmed', 'checked_in')
    AND res.deleted_at IS NULL
WHERE t.deleted_at IS NULL;

-- ==================== REVENUE DASHBOARD VIEW ====================

CREATE OR REPLACE VIEW v_revenue_dashboard AS
SELECT
    c.tenant_id,
    p.id AS property_id,
    p.name AS property_name,
    DATE_TRUNC('day', c.charge_date) AS charge_day,
    DATE_TRUNC('month', c.charge_date) AS charge_month,
    c.charge_type,
    c.category,
    c.revenue_center,
    COUNT(*) AS transaction_count,
    SUM(c.total_amount) AS gross_revenue,
    SUM(c.tax_amount) AS tax_collected,
    SUM(c.total_amount - c.tax_amount) AS net_revenue
FROM charges c
JOIN folios f ON c.folio_id = f.id
JOIN reservations res ON f.reservation_id = res.id
JOIN properties p ON res.property_id = p.id
WHERE c.is_voided = FALSE
GROUP BY c.tenant_id, p.id, p.name, DATE_TRUNC('day', c.charge_date), DATE_TRUNC('month', c.charge_date), c.charge_type, c.category, c.revenue_center;
