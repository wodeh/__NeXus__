-- Revert 000001_init_schema
DROP TRIGGER IF EXISTS tenant_properties_audit_trigger ON tenant_properties;
DROP TRIGGER IF EXISTS tenants_audit_trigger ON tenants;
DROP FUNCTION IF EXISTS audit_trigger_func();
DROP TABLE IF EXISTS tenant_properties;
DROP TABLE IF EXISTS tenants;
