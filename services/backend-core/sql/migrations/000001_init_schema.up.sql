-- Nexus Hospitality Platform — backend-core initial schema
-- Phase 2: foundation tables with tenant isolation, audit columns, soft delete, and optimistic locking.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Tenant registry — the root of all isolation
CREATE TABLE IF NOT EXISTS tenants (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id TEXT NOT NULL UNIQUE, -- user-facing tenant slug
    name        TEXT NOT NULL,
    region      TEXT NOT NULL DEFAULT 'us-east-1',
    tier        TEXT NOT NULL DEFAULT 'enterprise',
    config      JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    version     INTEGER NOT NULL DEFAULT 1
);

CREATE UNIQUE INDEX idx_tenants_external_id ON tenants(external_id) WHERE deleted_at IS NULL;

-- Tenant configuration table (property-level overrides)
CREATE TABLE IF NOT EXISTS tenant_properties (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    property_id TEXT NOT NULL,
    name        TEXT NOT NULL,
    timezone    TEXT NOT NULL DEFAULT 'UTC',
    locale      TEXT NOT NULL DEFAULT 'en',
    currency    TEXT NOT NULL DEFAULT 'USD',
    config      JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    version     INTEGER NOT NULL DEFAULT 1,
    UNIQUE(tenant_id, property_id)
);

CREATE INDEX idx_tenant_properties_tenant_id ON tenant_properties(tenant_id) WHERE deleted_at IS NULL;

-- Audit trigger function (auto-updates updated_at and version)
CREATE OR REPLACE FUNCTION audit_trigger_func()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply audit triggers
CREATE TRIGGER tenants_audit_trigger
    BEFORE UPDATE ON tenants
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_func();

CREATE TRIGGER tenant_properties_audit_trigger
    BEFORE UPDATE ON tenant_properties
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_func();

-- Row-Level Security preparation: enable RLS on all tenant-scoped tables
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE tenant_properties ENABLE ROW LEVEL SECURITY;

-- RLS policies (using current_setting('app.current_tenant') for connection-level tenant binding)
CREATE POLICY tenant_isolation_policy ON tenants
    FOR ALL
    USING (id::text = current_setting('app.current_tenant', true) OR current_setting('app.current_tenant', true) = '');

CREATE POLICY tenant_property_isolation_policy ON tenant_properties
    FOR ALL
    USING (tenant_id::text = current_setting('app.current_tenant', true) OR current_setting('app.current_tenant', true) = '');
