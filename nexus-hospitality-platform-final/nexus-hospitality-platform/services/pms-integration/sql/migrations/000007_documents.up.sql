-- Migration: Document Storage — S3-backed file metadata
-- Version: 000007

CREATE TABLE IF NOT EXISTS documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    type VARCHAR(50) NOT NULL,
    file_name VARCHAR(255) NOT NULL,
    content_type VARCHAR(100),
    size_bytes BIGINT NOT NULL DEFAULT 0,
    s3_bucket VARCHAR(100),
    s3_key TEXT,
    s3_url TEXT,
    uploaded_by UUID REFERENCES users(id) ON DELETE SET NULL,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_documents_tenant ON documents(tenant_id);
CREATE INDEX idx_documents_entity ON documents(tenant_id, entity_type, entity_id);
CREATE INDEX idx_documents_type ON documents(tenant_id, type);

-- Triggers
CREATE TRIGGER update_documents_updated_at
    BEFORE UPDATE ON documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- RLS
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation_documents ON documents
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
