CREATE TABLE rate_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    property_id UUID,
    name TEXT NOT NULL,
    room_type TEXT,
    condition_type TEXT NOT NULL CHECK (condition_type IN ('season', 'day_of_week', 'length_of_stay', 'advance_booking', 'occupancy')),
    condition_value TEXT NOT NULL, -- JSON or simple string depending on type
    rate_adjustment_type TEXT NOT NULL CHECK (rate_adjustment_type IN ('fixed_amount', 'percentage', 'fixed_rate')),
    rate_adjustment_value INT NOT NULL, -- in cents or percentage
    min_nights INT DEFAULT 1,
    max_nights INT,
    start_date DATE,
    end_date DATE,
    priority INT DEFAULT 0,
    active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_rate_rules_tenant ON rate_rules(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_rate_rules_active ON rate_rules(active) WHERE deleted_at IS NULL;
CREATE INDEX idx_rate_rules_dates ON rate_rules(start_date, end_date) WHERE deleted_at IS NULL;

ALTER TABLE rate_rules ENABLE ROW LEVEL SECURITY;

CREATE POLICY rate_rules_tenant_isolation ON rate_rules
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID OR current_setting('app.current_tenant', true) = '');

CREATE TRIGGER trg_rate_rules_updated_at
    BEFORE UPDATE ON rate_rules
    FOR EACH ROW
    EXECUTE FUNCTION update_reservations_updated_at();
