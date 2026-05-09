-- Folio & Billing + Housekeeping + Maintenance Down Migration

DROP TABLE IF EXISTS folio_adjustments CASCADE;
DROP TABLE IF EXISTS payments CASCADE;
DROP TABLE IF EXISTS charges CASCADE;
DROP TABLE IF EXISTS folios CASCADE;
DROP TABLE IF EXISTS housekeeping_tasks CASCADE;
DROP TABLE IF EXISTS work_orders CASCADE;
DROP TABLE IF EXISTS audit_log CASCADE;
DROP VIEW IF EXISTS v_housekeeping_board CASCADE;
DROP VIEW IF EXISTS v_revenue_dashboard CASCADE;
DROP FUNCTION IF EXISTS recalculate_folio_balance(UUID) CASCADE;

-- Note: Triggers are dropped with tables automatically
