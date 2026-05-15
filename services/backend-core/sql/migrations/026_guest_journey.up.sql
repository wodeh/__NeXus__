CREATE TABLE guest_journey_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL,
    reservation_id UUID NOT NULL REFERENCES reservations(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL CHECK (event_type IN ('booking_confirmed', 'pre_arrival_email_sent', 'check_in_completed', 'mid_stay_check', 'check_out_completed', 'post_stay_review_requested', 'post_stay_review_received', 'complaint_raised', 'complaint_resolved')),
    event_data JSONB DEFAULT '{}',
    occurred_at TIMESTAMPTZ DEFAULT now(),
    created_by TEXT,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_guest_journey_tenant ON guest_journey_events(tenant_id);
CREATE INDEX idx_guest_journey_reservation ON guest_journey_events(reservation_id);
CREATE INDEX idx_guest_journey_type ON guest_journey_events(event_type);

ALTER TABLE guest_journey_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY guest_journey_tenant_isolation ON guest_journey_events
    FOR ALL
    USING (tenant_id = current_setting('app.current_tenant', true)::UUID OR current_setting('app.current_tenant', true) = '');
