-- Create agents table
CREATE TABLE IF NOT EXISTS agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('ota', 'travel_agent', 'corporate', 'direct')),
    commission_pct INTEGER NOT NULL DEFAULT 0 CHECK (commission_pct >= 0 AND commission_pct <= 100),
    contact_name VARCHAR(255),
    contact_email VARCHAR(255),
    contact_phone VARCHAR(50),
    contract_ref VARCHAR(100),
    is_active BOOLEAN NOT NULL DEFAULT true,
    source_code VARCHAR(20) NOT NULL,
    config JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, source_code)
);

CREATE INDEX idx_agents_tenant ON agents(tenant_id);
CREATE INDEX idx_agents_type ON agents(type);
CREATE INDEX idx_agents_active ON agents(tenant_id, is_active);

-- Create channels table
CREATE TABLE IF NOT EXISTS channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    source VARCHAR(50) NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    commission_pct INTEGER NOT NULL DEFAULT 0 CHECK (commission_pct >= 0 AND commission_pct <= 100),
    last_sync_at TIMESTAMPTZ,
    last_sync_status VARCHAR(20) NOT NULL DEFAULT 'n/a' CHECK (last_sync_status IN ('success', 'warning', 'error', 'n/a')),
    config JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(tenant_id, source)
);

CREATE INDEX idx_channels_tenant ON channels(tenant_id);
CREATE INDEX idx_channels_active ON channels(tenant_id, is_active);

-- Create channel_sync_logs table
CREATE TABLE IF NOT EXISTS channel_sync_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id),
    channel_id UUID NOT NULL REFERENCES channels(id),
    channel_source VARCHAR(50) NOT NULL,
    direction VARCHAR(20) NOT NULL CHECK (direction IN ('pull', 'push')),
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'partial', 'error')),
    records INTEGER NOT NULL DEFAULT 0,
    duration VARCHAR(20),
    error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sync_logs_tenant ON channel_sync_logs(tenant_id);
CREATE INDEX idx_sync_logs_channel ON channel_sync_logs(channel_id);
CREATE INDEX idx_sync_logs_created ON channel_sync_logs(created_at DESC);
