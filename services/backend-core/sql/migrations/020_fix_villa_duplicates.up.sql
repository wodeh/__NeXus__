-- Fix duplicate villas caused by missing unique constraint
-- Step 1: Keep only the first villa per (tenant_id, name) and delete duplicates
DELETE FROM villa_properties a
USING villa_properties b
WHERE a.id > b.id
  AND a.tenant_id = b.tenant_id
  AND a.name = b.name;

-- Step 2: Add unique constraint if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'villa_properties_tenant_id_name_key'
    ) THEN
        ALTER TABLE villa_properties ADD CONSTRAINT villa_properties_tenant_id_name_key UNIQUE (tenant_id, name);
    END IF;
END $$;
