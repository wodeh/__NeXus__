-- Migration: Users, Roles, and Permissions for Admin Section
CREATE TABLE IF NOT EXISTS users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    email       TEXT NOT NULL,
    name        TEXT NOT NULL,
    phone       TEXT,
    role        TEXT NOT NULL DEFAULT 'front_desk',
    -- super_admin, admin, manager, front_desk, housekeeping, readonly
    status      TEXT NOT NULL DEFAULT 'active',
    -- active, inactive, suspended
    last_login  TIMESTAMPTZ,
    password_hash TEXT,
    permissions JSONB DEFAULT '[]',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_users_tenant ON users(tenant_id);
CREATE INDEX idx_users_role ON users(role);

CREATE TABLE IF NOT EXISTS roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id   UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    permissions JSONB NOT NULL DEFAULT '[]',
    description TEXT,
    is_system   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, name)
);

CREATE INDEX idx_roles_tenant ON roles(tenant_id);

-- Default roles per tenant
INSERT INTO roles (tenant_id, name, permissions, description, is_system)
SELECT id, 'super_admin', '["*"]'::jsonb, 'Full platform access', true FROM tenants
ON CONFLICT DO NOTHING;

INSERT INTO roles (tenant_id, name, permissions, description, is_system)
SELECT id, 'admin', '["reservations:read","reservations:write","rooms:read","rooms:write","housekeeping:read","housekeeping:write","reports:read","settings:read","settings:write","users:read","users:write"]'::jsonb, 'Hotel administrator', true FROM tenants
ON CONFLICT DO NOTHING;

INSERT INTO roles (tenant_id, name, permissions, description, is_system)
SELECT id, 'front_desk', '["reservations:read","reservations:write","rooms:read","guests:read"]'::jsonb, 'Front desk agent', true FROM tenants
ON CONFLICT DO NOTHING;

INSERT INTO roles (tenant_id, name, permissions, description, is_system)
SELECT id, 'housekeeper', '["housekeeping:read","housekeeping:write","rooms:read"]'::jsonb, 'Housekeeping staff', true FROM tenants
ON CONFLICT DO NOTHING;

INSERT INTO roles (tenant_id, name, permissions, description, is_system)
SELECT id, 'readonly', '["reservations:read","rooms:read","housekeeping:read"]'::jsonb, 'Read-only access', true FROM tenants
ON CONFLICT DO NOTHING;

-- Triggers
CREATE OR REPLACE FUNCTION update_users_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
CREATE TRIGGER trigger_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_users_updated_at();

DROP TRIGGER IF EXISTS trigger_roles_updated_at ON roles;
CREATE TRIGGER trigger_roles_updated_at
    BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_users_updated_at();

-- RLS
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE roles ENABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS users_tenant_isolation ON users;
CREATE POLICY users_tenant_isolation ON users
    USING (tenant_id = current_setting('app.current_tenant')::UUID);

DROP POLICY IF EXISTS roles_tenant_isolation ON roles;
CREATE POLICY roles_tenant_isolation ON roles
    USING (tenant_id = current_setting('app.current_tenant')::UUID);
