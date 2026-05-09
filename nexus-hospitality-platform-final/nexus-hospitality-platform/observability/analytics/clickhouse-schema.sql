-- observability/analytics/clickhouse-schema.sql
-- ============================================================
-- CLICKHOUSE ANALYTICS SCHEMA — Hospitality data platform
-- Supports: real-time streaming, guest 360, executive KPIs,
-- IPTV engagement, AI forecasting, anomaly detection
-- ============================================================

-- Guest 360 Profile (materialized view)
CREATE TABLE IF NOT EXISTS guest_360_profiles (
    tenant_id UUID,
    guest_id UUID,
    property_id UUID,

    -- Identity
    first_name String,
    last_name String,
    email String,
    phone String,
    nationality LowCardinality(String),
    language LowCardinality(String),

    -- Demographics
    age_group LowCardinality(String),
    traveler_type LowCardinality(String), -- business, leisure, family, group

    -- Stay history
    total_nights UInt32,
    total_stays UInt32,
    total_spend Decimal(15, 2),
    average_daily_rate Decimal(10, 2),
    last_checkin Date,
    last_checkout Date,
    preferred_room_type LowCardinality(String),

    -- Preferences
    room_preferences Array(String),
    dietary_restrictions Array(String),
    accessibility_needs Array(String),
    communication_channel LowCardinality(String), -- email, sms, app, phone

    -- Loyalty
    loyalty_tier LowCardinality(String),
    loyalty_points UInt64,
    lifetime_value Decimal(15, 2),

    -- Engagement
    iptv_watch_time_minutes UInt32,
    spa_bookings UInt32,
    dining_visits UInt32,
    concierge_interactions UInt32,

    -- Sentiment
    last_nps_score Int8,
    sentiment_score Float32,
    complaint_count UInt32,
    compliment_count UInt32,

    -- AI predictions
    churn_risk Float32,
    next_booking_probability Float32,
    estimated_next_spend Decimal(10, 2),
    preferred_channel_prediction LowCardinality(String),

    -- Metadata
    updated_at DateTime DEFAULT now(),

    INDEX idx_tenant (tenant_id) TYPE minmax GRANULARITY 4,
    INDEX idx_guest (guest_id) TYPE bloom_filter GRANULARITY 4,
    INDEX idx_email (email) TYPE bloom_filter GRANULARITY 4
)
ENGINE = ReplacingMergeTree(updated_at)
PARTITION BY toYYYYMM(updated_at)
ORDER BY (tenant_id, guest_id)
SETTINGS index_granularity = 8192;

-- Real-time streaming analytics
CREATE TABLE IF NOT EXISTS iptv_stream_analytics (
    timestamp DateTime64(3),
    tenant_id UUID,
    property_id UUID,
    room_id String,
    guest_id UUID,

    -- Stream metadata
    stream_id String,
    channel_id UInt32,
    channel_name LowCardinality(String),
    content_type LowCardinality(String), -- live, vod, catchup

    -- Quality metrics
    bitrate UInt32,
    resolution LowCardinality(String),
    buffering_events UInt32,
    buffering_duration_ms UInt32,
    startup_time_ms UInt32,

    -- Engagement
    watch_duration_seconds UInt32,
    pause_count UInt32,
    seek_count UInt32,
    volume_changes UInt32,

    -- Device
    device_type LowCardinality(String), -- tv, mobile, tablet, laptop
    device_os LowCardinality(String),

    -- Network
    cdn_edge LowCardinality(String),
    network_type LowCardinality(String), -- wifi, ethernet, cellular

    -- QoE score (0-100)
    qoe_score UInt8
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (tenant_id, timestamp, room_id)
TTL timestamp + INTERVAL 90 DAY
SETTINGS index_granularity = 8192;

-- IoT telemetry analytics
CREATE TABLE IF NOT EXISTS iot_telemetry_analytics (
    timestamp DateTime64(3),
    tenant_id UUID,
    property_id UUID,
    room_id String,
    device_id String,
    device_type LowCardinality(String),

    -- Sensor readings
    temperature Float32,
    humidity Float32,
    occupancy Boolean,
    light_level UInt8,
    power_consumption_watts Float32,

    -- HVAC
    hvac_mode LowCardinality(String),
    setpoint_temperature Float32,
    fan_speed LowCardinality(String),

    -- Energy
    energy_kwh Decimal(10, 3),
    water_liters Decimal(10, 2),

    -- Anomaly flags
    temperature_anomaly Boolean,
    occupancy_anomaly Boolean,
    energy_anomaly Boolean
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (tenant_id, timestamp, room_id, device_id)
TTL timestamp + INTERVAL 30 DAY
SETTINGS index_granularity = 8192;

-- Executive KPI engine
CREATE TABLE IF NOT EXISTS executive_kpis (
    date Date,
    tenant_id UUID,
    property_id UUID,

    -- Revenue
    total_revenue Decimal(15, 2),
    room_revenue Decimal(15, 2),
    f_b_revenue Decimal(15, 2),
    spa_revenue Decimal(15, 2),
    other_revenue Decimal(15, 2),

    -- Occupancy
    available_rooms UInt32,
    sold_rooms UInt32,
    occupancy_rate Float32,
    adr Decimal(10, 2),
    revpar Decimal(10, 2),

    -- Operations
    checkins UInt32,
    checkouts UInt32,
    no_shows UInt32,
    cancellations UInt32,
    walk_ins UInt32,

    -- Housekeeping
    rooms_cleaned UInt32,
    avg_cleaning_time_minutes Float32,
    cleaning_quality_score Float32,

    -- Maintenance
    requests_created UInt32,
    requests_completed UInt32,
    avg_resolution_time_hours Float32,
    guest_impact_count UInt32,

    -- Staff
    scheduled_hours UInt32,
    actual_hours UInt32,
    overtime_hours UInt32,
    labor_cost Decimal(10, 2),

    -- Guest satisfaction
    nps_score Float32,
    review_count UInt32,
    avg_rating Float32,
    complaint_count UInt32,

    -- IPTV
    total_streaming_hours Float32,
    unique_viewers UInt32,
    vod_revenue Decimal(10, 2),

    -- AI insights
    ai_recommendation_acceptance_rate Float32,
    ai_concierge_sessions UInt32,
    predicted_demand_accuracy Float32
)
ENGINE = SummingMergeTree()
PARTITION BY toYYYYMM(date)
ORDER BY (tenant_id, property_id, date)
SETTINGS index_granularity = 8192;

-- Materialized view: Guest 360 aggregation
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_guest_360_aggregation
TO guest_360_profiles
AS
SELECT
    tenant_id,
    guest_id,
    any(property_id) as property_id,
    any(first_name) as first_name,
    any(last_name) as last_name,
    any(email) as email,
    any(phone) as phone,
    any(nationality) as nationality,
    any(language) as language,

    -- Aggregations
    sum(total_nights) as total_nights,
    count() as total_stays,
    sum(total_spend) as total_spend,
    avg(adr) as average_daily_rate,
    max(last_checkin) as last_checkin,
    max(last_checkout) as last_checkout,

    now() as updated_at
FROM reservations
GROUP BY tenant_id, guest_id;

-- Materialized view: Real-time occupancy
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_realtime_occupancy
ENGINE = AggregatingMergeTree()
PARTITION BY toYYYYMMDD(now())
ORDER BY (tenant_id, property_id, room_id)
AS
SELECT
    tenant_id,
    property_id,
    room_id,
    argMaxState(status, timestamp) as current_status,
    countState() as status_changes_today
FROM room_status_history
GROUP BY tenant_id, property_id, room_id;

-- Anomaly detection table
CREATE TABLE IF NOT EXISTS anomaly_detections (
    timestamp DateTime64(3),
    tenant_id UUID,
    property_id UUID,

    anomaly_type LowCardinality(String),
    entity_type LowCardinality(String), -- room, device, guest, staff
    entity_id String,

    -- Detection
    expected_value Float32,
    actual_value Float32,
    deviation_score Float32,
    severity LowCardinality(String), -- low, medium, high, critical

    -- Context
    model_version String,
    detection_method LowCardinality(String), -- statistical, ml, rule

    -- Resolution
    acknowledged Boolean DEFAULT false,
    acknowledged_by UUID,
    acknowledged_at DateTime,
    resolved Boolean DEFAULT false,
    resolved_at DateTime,
    resolution_notes String
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (tenant_id, timestamp, anomaly_type)
TTL timestamp + INTERVAL 1 YEAR;

-- Data lineage tracking
CREATE TABLE IF NOT EXISTS data_lineage (
    timestamp DateTime64(3),

    source_system LowCardinality(String),
    source_table String,
    source_columns Array(String),

    target_system LowCardinality(String),
    target_table String,
    target_columns Array(String),

    transformation_logic String,
    transformation_type LowCardinality(String), -- etl, elt, streaming, sync

    record_count UInt64,
    duration_ms UInt32,
    status LowCardinality(String), -- success, failed, partial
    error_message String
)
ENGINE = MergeTree()
PARTITION BY toYYYYMMDD(timestamp)
ORDER BY (timestamp, source_system, target_system);
