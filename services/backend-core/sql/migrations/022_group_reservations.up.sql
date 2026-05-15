-- Add group reservation support to reservations table
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS group_id UUID;
ALTER TABLE reservations ADD COLUMN IF NOT EXISTS group_name TEXT;

CREATE INDEX IF NOT EXISTS idx_reservations_group_id ON reservations(group_id) WHERE deleted_at IS NULL;
