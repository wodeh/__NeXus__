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

-- Seed demo APs for Grand Plaza
INSERT INTO wifi_access_points (tenant_id, name, floor, location, mac_address, ip_address, model, channels_2ghz, channels_5ghz, status, snmp_community, snmp_version) VALUES
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'Lobby-AP-01', 'Lobby', 'Main entrance ceiling', '00:11:22:33:44:AA', '192.168.10.10', 'Ubiquiti U6-Pro', ARRAY[1,6], ARRAY[36,40], 'online', 'public', 'v2c'),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'Lobby-AP-02', 'Lobby', 'Reception desk', '00:11:22:33:44:AB', '192.168.10.11', 'Ubiquiti U6-Pro', ARRAY[1,11], ARRAY[44,48], 'online', 'public', 'v2c'),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'Lobby-AP-03', 'Lobby', 'Restaurant area', '00:11:22:33:44:AC', '192.168.10.12', 'Ubiquiti U6-Pro', ARRAY[6,11], ARRAY[52,56], 'online', 'public', 'v2c'),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'F1-AP-01', 'Floor 1', 'Hallway east', '00:11:22:33:44:A1', '192.168.10.20', 'Ubiquiti U6-Lite', ARRAY[1], ARRAY[36], 'online', 'public', 'v2c'),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'F1-AP-02', 'Floor 1', 'Hallway west', '00:11:22:33:44:A2', '192.168.10.21', 'Ubiquiti U6-Lite', ARRAY[6], ARRAY[40], 'online', 'public', 'v2c'),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'F2-AP-01', 'Floor 2', 'Hallway east', '00:11:22:33:44:B1', '192.168.10.30', 'Ubiquiti U6-Lite', ARRAY[1], ARRAY[36], 'offline', 'public', 'v2c'),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'F2-AP-02', 'Floor 2', 'Hallway west', '00:11:22:33:44:B2', '192.168.10.31', 'Ubiquiti U6-Lite', ARRAY[6], ARRAY[40], 'online', 'public', 'v2c'),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'F3-AP-01', 'Floor 3', 'Hallway east', '00:11:22:33:44:C1', '192.168.10.40', 'Ubiquiti U6-Pro', ARRAY[1], ARRAY[36], 'online', 'public', 'v2c'),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'F3-AP-02', 'Floor 3', 'Hallway west', '00:11:22:33:44:C2', '192.168.10.41', 'Ubiquiti U6-Pro', ARRAY[6], ARRAY[40], 'online', 'public', 'v2c'),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'F4-AP-01', 'Floor 4', 'Hallway east', '00:11:22:33:44:D1', '192.168.10.50', 'Ubiquiti U6-Lite', ARRAY[1], ARRAY[36], 'online', 'public', 'v2c'),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'F4-AP-02', 'Floor 4', 'Hallway west', '00:11:22:33:44:D2', '192.168.10.51', 'Ubiquiti U6-Lite', ARRAY[6], ARRAY[40], 'online', 'public', 'v2c'),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'F5-AP-01', 'Floor 5', 'Hallway east', '00:11:22:33:44:E1', '192.168.10.60', 'Ubiquiti U6-Lite', ARRAY[1], ARRAY[36], 'online', 'public', 'v2c'),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'F5-AP-02', 'Floor 5', 'Hallway west', '00:11:22:33:44:E2', '192.168.10.61', 'Ubiquiti U6-Lite', ARRAY[11], ARRAY[48], 'online', 'public', 'v2c'),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), 'F5-AP-03', 'Floor 5', 'Hallway center', '00:11:22:33:44:E3', '192.168.10.62', 'Ubiquiti U6-Lite', ARRAY[6], ARRAY[40], 'degraded', 'public', 'v2c');

-- Insert demo metrics (last 30 minutes)
INSERT INTO wifi_metrics (tenant_id, ap_id, floor, rssi_dbm, noise_dbm, snr_db, quality_score, client_count, bandwidth_mbps, packet_loss_pct, channel_util, tx_rate_mbps, rx_rate_mbps) VALUES
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:AA'), 'Lobby', -42, -90, 48, 95, 23, 450.5, 0.1, 35.0, 867.0, 520.0),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:AB'), 'Lobby', -45, -92, 47, 93, 19, 380.2, 0.2, 28.0, 650.0, 410.0),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:AC'), 'Lobby', -48, -91, 43, 88, 18, 410.0, 0.3, 32.0, 780.0, 460.0),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:A1'), 'Floor 1', -52, -93, 41, 85, 12, 320.0, 0.5, 22.0, 520.0, 310.0),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:A2'), 'Floor 1', -55, -94, 39, 82, 10, 290.0, 0.8, 18.0, 450.0, 280.0),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:B2'), 'Floor 2', -58, -95, 37, 75, 15, 250.0, 1.2, 45.0, 400.0, 220.0),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:C1'), 'Floor 3', -50, -92, 42, 87, 14, 350.0, 0.4, 25.0, 600.0, 380.0),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:C2'), 'Floor 3', -53, -93, 40, 84, 13, 330.0, 0.6, 20.0, 550.0, 340.0),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:D1'), 'Floor 4', -56, -94, 38, 80, 11, 300.0, 0.9, 19.0, 480.0, 290.0),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:D2'), 'Floor 4', -54, -93, 39, 83, 9, 310.0, 0.7, 21.0, 500.0, 300.0),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:E1'), 'Floor 5', -68, -96, 28, 55, 8, 150.0, 3.5, 65.0, 200.0, 120.0),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:E2'), 'Floor 5', -72, -97, 25, 48, 7, 120.0, 4.2, 72.0, 180.0, 100.0),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:E3'), 'Floor 5', -65, -95, 30, 60, 6, 170.0, 2.8, 58.0, 220.0, 140.0);

-- Insert demo alerts
INSERT INTO wifi_alerts (tenant_id, ap_id, floor, type, severity, message, suggested_fix, is_resolved) VALUES
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:B1'), 'Floor 2', 'ap_offline', 'critical', 'AP F2-AP-01 (Hallway east) is offline for 15 minutes', 'Check power and Ethernet connection to AP', FALSE),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:E1'), 'Floor 5', 'poor_signal', 'warning', 'AP F5-AP-01 signal degraded to -68 dBm', 'Check AP placement and obstructions. Consider adding an AP nearby.', FALSE),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:E2'), 'Floor 5', 'channel_conflict', 'warning', 'Channel overlap detected on 5GHz band (ch 48)', 'Switch to channel 52 or 56 to avoid interference', FALSE),
((SELECT id FROM tenants WHERE external_id = 'DOWNTOWN'), (SELECT id FROM wifi_access_points WHERE mac_address = '00:11:22:33:44:E3'), 'Floor 5', 'overloaded', 'warning', 'AP F5-AP-03 channel utilization at 58%', 'Reduce client count or enable band steering', FALSE);
