ALTER TABLE reservations ADD COLUMN IF NOT EXISTS pre_arrival_ready BOOLEAN DEFAULT FALSE;
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS deposit_paid INT DEFAULT 0;
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS special_requests_acknowledged BOOLEAN DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_reservations_pre_arrival ON reservations(pre_arrival_ready) WHERE deleted_at IS NULL;
