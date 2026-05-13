-- Fix audit trigger: add missing version column to tables that use audit_trigger_func
-- but were created without it.

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'version'
    ) THEN
        ALTER TABLE users ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'properties' AND column_name = 'version'
    ) THEN
        ALTER TABLE properties ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
    END IF;
END $$;
