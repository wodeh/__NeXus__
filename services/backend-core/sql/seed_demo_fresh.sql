-- ========================================================================
-- Clean all demo data and re-seed fresh (one-shot idempotent run)
-- ========================================================================

DO $$
DECLARE
    demo_tenant_id UUID;
BEGIN
    SELECT id INTO demo_tenant_id FROM tenants WHERE external_id = 'demo';
    IF demo_tenant_id IS NULL THEN
        RAISE EXCEPTION 'Demo tenant not found. Run seed_admin.sql first.';
    END IF;

    -- Delete in dependency order (child tables first)
    DELETE FROM wifi_alerts WHERE tenant_id = demo_tenant_id;
    DELETE FROM wifi_metrics WHERE tenant_id = demo_tenant_id;
    DELETE FROM wifi_floor_summary WHERE tenant_id = demo_tenant_id;
    DELETE FROM lock_events WHERE tenant_id = demo_tenant_id;
    DELETE FROM smart_locks WHERE tenant_id = demo_tenant_id;
    DELETE FROM housekeeping_tasks WHERE tenant_id = demo_tenant_id;
    DELETE FROM housekeeping_staff WHERE tenant_id = demo_tenant_id;
    DELETE FROM reservations WHERE tenant_id = demo_tenant_id;
    DELETE FROM rooms WHERE tenant_id = demo_tenant_id;

    RAISE NOTICE 'Demo data wiped for tenant %', demo_tenant_id;
END $$;

-- Now run the seed
\i seed_demo_data.sql
