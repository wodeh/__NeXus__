-- observability/analytics/clickhouse/schema.sql
-- Nexus Hospitality Analytics Platform

CREATE TABLE guest_360_profiles (
    tenant_id String,
    guest_id UUID,
    property_id UUID,
    email String,
    phone String,
    first_name String,
    last_name String,
    lifetime_value Float64,
    total_stays UInt32,
    total_nights UInt32,
    total_spend Float64,
    average_daily_rate Float64,
    room_type_preference String,
    dietary_restrictions Array(String),
    booking_lead_time_avg UInt32,
    cancellation_rate Float32,
    loyalty_tier String,
    loyalty_points_balance UInt32,
    churn_risk_score Float32,
    upsell_propensity_score Float32,
    satisfaction_predicted Float32,
    created_at DateTime64(3),
    updated_at DateTime64(3),
    INDEX idx_tenant (tenant_id) TYPE minmax GRANULARITY 4,
    INDEX idx_guest (guest_id) TYPE bloom_filter GRANULARITY 4
) ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(created_at)
ORDER BY (tenant_id, guest_id)
SETTINGS index_granularity = 8192;

CREATE TABLE iptv_engagement (
    tenant_id String,
    property_id UUID,
    room_id UUID,
    guest_id UUID,
    session_id UUID,
    channel_id String,
    channel_name String,
    content_type String,
    session_start DateTime64(3),
    session_end DateTime64(3),
    duration_seconds UInt32,
    bitrate_avg UInt32,
    bitrate_switches UInt16,
    buffer_events UInt16,
    buffer_duration_ms UInt32,
    startup_time_ms UInt32,
    device_type String,
    device_model String,
    ad_impressions UInt16,
    ad_clicks UInt16,
    ppv_purchased UInt8,
    ppv_amount Float64,
    country String,
    city String,
    created_at DateTime64(3),
    INDEX idx_session (session_id) TYPE bloom_filter GRANULARITY 4,
    INDEX idx_channel (channel_id) TYPE minmax GRANULARITY 4
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(session_start)
ORDER BY (tenant_id, property_id, session_start, guest_id)
SETTINGS index_granularity = 8192;

CREATE TABLE revenue_analytics (
    tenant_id String,
    property_id UUID,
    date Date,
    rooms_available UInt16,
    rooms_sold UInt16,
    occupancy_rate Float32,
    adr Float64,
    revpar Float64,
    trespar Float64,
    transient_revenue Float64,
    group_revenue Float64,
    corporate_revenue Float64,
    ota_revenue Float64,
    direct_revenue Float64,
    forecasted_occupancy Float32,
    forecasted_adr Float64,
    forecasted_revpar Float64,
    ai_demand_forecast UInt16,
    ai_rate_recommendation Float64,
    ai_confidence_score Float32,
    created_at DateTime64(3),
    INDEX idx_property (property_id) TYPE minmax GRANULARITY 4,
    INDEX idx_date (date) TYPE minmax GRANULARITY 4
) ENGINE = ReplacingMergeTree(created_at)
PARTITION BY toYYYYMM(date)
ORDER BY (tenant_id, property_id, date)
SETTINGS index_granularity = 8192;

CREATE TABLE iot_telemetry (
    tenant_id String,
    property_id UUID,
    room_id UUID,
    device_id String,
    device_type String,
    timestamp DateTime64(3),
    energy_consumption_kwh Float64,
    power_watts Float32,
    temperature_indoor Float32,
    temperature_target Float32,
    humidity_indoor Float32,
    hvac_mode String,
    light_level_lux Float32,
    light_brightness UInt8,
    occupancy_detected UInt8,
    motion_events UInt16,
    water_usage_liters Float64,
    co2_ppm UInt16,
    anomaly_detected UInt8,
    anomaly_score Float32,
    maintenance_predicted UInt8,
    energy_optimization_savings Float64,
    firmware_version String,
    battery_level UInt8,
    signal_strength Int8,
    INDEX idx_device (device_id) TYPE bloom_filter GRANULARITY 4,
    INDEX idx_timestamp (timestamp) TYPE minmax GRANULARITY 4
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (tenant_id, property_id, device_id, timestamp)
TTL timestamp + INTERVAL 2 YEAR
SETTINGS index_granularity = 8192;
