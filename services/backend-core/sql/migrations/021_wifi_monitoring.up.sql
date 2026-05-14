-- WiFi Monitoring Module — SNMP-based enterprise AP monitoring
-- Creates tables for access points, metrics, and alerts

-- Access Points inventory
CREATE TABLE IF NOT EXISTS wifi_access_points (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name            VARCHAR(100) NOT NULL,
    floor           VARCHAR(50) NOT NULL,
    location        VARCHAR(200),
    mac_address     VARCHAR(17) NOT NULL,           -- AA:BB:CC:DD:EE:FF
    ip_address      INET,
    model           VARCHAR(100),
    firmware        VARCHAR(50),
    channels_2ghz   INTEGER[],                     -- e.g. {1,6,11}
    channels_5ghz   INTEGER[],
    max_clients     INTEGER DEFAULT 64,
    status          VARCHAR(20) DEFAULT 'online',  -- online, offline, degraded, maintenance
    snmp_community  VARCHAR(50),                   -- community string for SNMP v2c
    snmp_version    VARCHAR(10) DEFAULT 'v2c',      -- v2c, v3
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(tenant_id, mac_address)
);

-- WiFi metrics — polled every 5 minutes via SNMP
CREATE TABLE IF NOT EXISTS wifi_metrics (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    ap_id           UUID NOT NULL REFERENCES wifi_access_points(id) ON DELETE CASCADE,
    floor           VARCHAR(50) NOT NULL,
    timestamp       TIMESTAMPTZ DEFAULT now(),
    rssi_dbm        INTEGER,                       -- negative, closer to 0 = better
    noise_dbm       INTEGER,                       -- noise floor
    snr_db          INTEGER,                       -- signal-to-noise ratio
    quality_score   INTEGER CHECK (quality_score BETWEEN 0 AND 100),  -- composite 0-100
    client_count    INTEGER DEFAULT 0,
    bandwidth_mbps  NUMERIC(10,2),                -- current throughput
    packet_loss_pct NUMERIC(5,2),                  -- 0.00 - 100.00
    channel_util    NUMERIC(5,2),                  -- channel utilization %
    tx_rate_mbps    NUMERIC(10,2),                 -- transmit rate
    rx_rate_mbps    NUMERIC(10,2),                 -- receive rate
    retries_pct     NUMERIC(5,2)                   -- retry rate
);

-- WiFi alerts — active and resolved
CREATE TABLE IF NOT EXISTS wifi_alerts (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    ap_id           UUID REFERENCES wifi_access_points(id) ON DELETE CASCADE,
    floor           VARCHAR(50) NOT NULL,
    type            VARCHAR(30) NOT NULL,          -- dead_zone, channel_conflict, ap_offline, overloaded, poor_signal, interference
    severity        VARCHAR(10) NOT NULL,          -- info, warning, critical
    message         TEXT NOT NULL,
    suggested_fix   TEXT,
    is_resolved     BOOLEAN DEFAULT FALSE,
    resolved_at     TIMESTAMPTZ,
    resolved_by     VARCHAR(100),
    created_at      TIMESTAMPTZ DEFAULT now()
);

-- Floor summary — materialized view for fast dashboard reads
CREATE TABLE IF NOT EXISTS wifi_floor_summary (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    floor           VARCHAR(50) NOT NULL,
    ap_count        INTEGER DEFAULT 0,
    online_aps      INTEGER DEFAULT 0,
    avg_signal_dbm  INTEGER,
    total_clients   INTEGER DEFAULT 0,
    active_alerts   INTEGER DEFAULT 0,
    overall_health  VARCHAR(10) DEFAULT 'good',   -- excellent, good, fair, poor
    updated_at      TIMESTAMPTZ DEFAULT now(),
    UNIQUE(tenant_id, floor)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_wifi_metrics_ap_time ON wifi_metrics(ap_id, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_wifi_metrics_floor_time ON wifi_metrics(floor, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_wifi_alerts_unresolved ON wifi_alerts(tenant_id, is_resolved, severity DESC);
CREATE INDEX IF NOT EXISTS idx_wifi_alerts_floor ON wifi_alerts(tenant_id, floor, is_resolved);

-- RLS policies: tenant isolation
ALTER TABLE wifi_access_points ENABLE ROW LEVEL SECURITY;
ALTER TABLE wifi_metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE wifi_alerts ENABLE ROW LEVEL SECURITY;
ALTER TABLE wifi_floor_summary ENABLE ROW LEVEL SECURITY;

CREATE POLICY wifi_ap_tenant_isolation ON wifi_access_points
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY wifi_metrics_tenant_isolation ON wifi_metrics
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY wifi_alerts_tenant_isolation ON wifi_alerts
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);
CREATE POLICY wifi_floor_summary_tenant_isolation ON wifi_floor_summary
    FOR ALL USING (tenant_id = current_setting('app.current_tenant')::UUID);

-- Trigger to update wifi_floor_summary when metrics are inserted
CREATE OR REPLACE FUNCTION update_wifi_floor_summary()
RETURNS TRIGGER AS $$
DECLARE
    v_tenant_id UUID;
    v_floor VARCHAR(50);
    v_ap_count INTEGER;
    v_online_aps INTEGER;
    v_avg_signal INTEGER;
    v_total_clients INTEGER;
    v_active_alerts INTEGER;
    v_health VARCHAR(10);
BEGIN
    v_tenant_id := NEW.tenant_id;
    v_floor := NEW.floor;

    SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'online')
    INTO v_ap_count, v_online_aps
    FROM wifi_access_points
    WHERE tenant_id = v_tenant_id AND floor = v_floor;

    SELECT AVG(rssi_dbm), SUM(client_count)
    INTO v_avg_signal, v_total_clients
    FROM wifi_metrics
    WHERE tenant_id = v_tenant_id AND floor = v_floor
      AND timestamp > NOW() - INTERVAL '10 minutes';

    SELECT COUNT(*) INTO v_active_alerts
    FROM wifi_alerts
    WHERE tenant_id = v_tenant_id AND floor = v_floor AND is_resolved = FALSE;

    -- Health scoring
    v_avg_signal := COALESCE(v_avg_signal, -100);
    IF v_avg_signal >= -50 THEN v_health := 'excellent';
    ELSIF v_avg_signal >= -60 THEN v_health := 'good';
    ELSIF v_avg_signal >= -70 THEN v_health := 'fair';
    ELSE v_health := 'poor';
    END IF;

    INSERT INTO wifi_floor_summary (tenant_id, floor, ap_count, online_aps, avg_signal_dbm, total_clients, active_alerts, overall_health, updated_at)
    VALUES (v_tenant_id, v_floor, v_ap_count, v_online_aps, v_avg_signal, COALESCE(v_total_clients, 0), v_active_alerts, v_health, NOW())
    ON CONFLICT (tenant_id, floor)
    DO UPDATE SET
        ap_count = EXCLUDED.ap_count,
        online_aps = EXCLUDED.online_aps,
        avg_signal_dbm = EXCLUDED.avg_signal_dbm,
        total_clients = EXCLUDED.total_clients,
        active_alerts = EXCLUDED.active_alerts,
        overall_health = EXCLUDED.overall_health,
        updated_at = NOW();

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS wifi_metrics_summary_trigger ON wifi_metrics;
CREATE TRIGGER wifi_metrics_summary_trigger
    AFTER INSERT ON wifi_metrics
    FOR EACH ROW
    EXECUTE FUNCTION update_wifi_floor_summary();
