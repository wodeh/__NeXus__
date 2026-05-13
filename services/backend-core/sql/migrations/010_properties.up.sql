-- Nexus Hospitality Platform — Properties table
-- Created: 2026-05-12

CREATE TABLE IF NOT EXISTS properties (
    id          UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    address     TEXT,
    city        TEXT,
    country     TEXT,
    phone       TEXT,
    email       TEXT,
    timezone    TEXT NOT NULL DEFAULT 'UTC',
    currency    TEXT NOT NULL DEFAULT 'USD',
    star_rating INTEGER NOT NULL DEFAULT 3,
    is_active   BOOLEAN NOT NULL DEFAULT true,
    config      JSONB NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ,
    version     INTEGER NOT NULL DEFAULT 1,
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_properties_tenant_id ON properties(tenant_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_properties_active ON properties(is_active) WHERE deleted_at IS NULL;

CREATE TRIGGER properties_audit_trigger
    BEFORE UPDATE ON properties
    FOR EACH ROW
    EXECUTE FUNCTION audit_trigger_func();

ALTER TABLE properties ENABLE ROW LEVEL SECURITY;

CREATE POLICY properties_tenant_isolation ON properties
    FOR ALL
    USING (tenant_id::text = current_setting('app.current_tenant', true) OR current_setting('app.current_tenant', true) = '');
