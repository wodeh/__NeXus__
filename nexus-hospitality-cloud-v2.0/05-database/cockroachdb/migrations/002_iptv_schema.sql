-- NHC IPTV Database Schema (CockroachDB)
-- Tenant-isolated schema for hospitality IPTV platform

CREATE SCHEMA IF NOT EXISTS iptv;

-- Content catalog
create table if not exists iptv.content (
    content_id uuid primary key default gen_random_uuid(),
    tenant_id uuid not null,
    content_type varchar(20) not null check (content_type in ('live_channel', 'vod', 'catchup', 'hotel_content', 'promo')),
    title varchar(255) not null,
    description text,
    genre varchar(50),
    rating varchar(10),
    duration_minutes int,
    poster_url varchar(500),
    trailer_url varchar(500),
    release_year int,
    language varchar(10),
    subtitles jsonb,
    audio_tracks jsonb,
    drm_required boolean default false,
    drm_type varchar(20),
    license_window_start timestamp,
    license_window_end timestamp,
    pricing_tier varchar(20) default 'included',
    price decimal(10,2),
    currency char(3) default 'USD',
    metadata jsonb,
    created_at timestamp default now(),
    updated_at timestamp default now(),
    index idx_tenant_type (tenant_id, content_type),
    index idx_rating (rating),
    index idx_genre (genre)
);

-- Live channels
create table if not exists iptv.channels (
    channel_id uuid primary key default gen_random_uuid(),
    tenant_id uuid not null,
    channel_number int not null,
    channel_name varchar(100) not null,
    channel_logo varchar(500),
    genre varchar(50),
    language varchar(10),
    hd boolean default true,
    epg_source varchar(50),
    stream_url varchar(500),
    backup_stream_url varchar(500),
    drm_required boolean default false,
    parental_rating varchar(10) default 'PG',
    is_premium boolean default false,
    is_promotional boolean default false,
    active boolean default true,
    sort_order int default 0,
    created_at timestamp default now(),
    unique (tenant_id, channel_number),
    index idx_tenant_active (tenant_id, active)
);

-- EPG (Electronic Program Guide)
create table if not exists iptv.epg (
    epg_id uuid primary key default gen_random_uuid(),
    tenant_id uuid not null,
    channel_id uuid not null references iptv.channels(channel_id),
    program_title varchar(255) not null,
    program_description text,
    episode_title varchar(255),
    season_number int,
    episode_number int,
    start_time timestamp not null,
    end_time timestamp not null,
    genre varchar(50),
    rating varchar(10),
    poster_url varchar(500),
    is_recordable boolean default true,
    is_catchup_available boolean default true,
    catchup_duration_hours int default 24,
    created_at timestamp default now(),
    index idx_channel_time (channel_id, start_time),
    index idx_tenant_time (tenant_id, start_time)
);

-- Guest viewing history
create table if not exists iptv.viewing_history (
    history_id uuid primary key default gen_random_uuid(),
    tenant_id uuid not null,
    guest_id uuid not null,
    room_id uuid not null,
    session_id uuid not null,
    content_id uuid references iptv.content(content_id),
    channel_id uuid references iptv.channels(channel_id),
    content_type varchar(20) not null,
    watch_start timestamp not null,
    watch_end timestamp,
    duration_seconds int,
    progress_seconds int default 0,
    completed boolean default false,
    device_type varchar(20),
    quality varchar(10),
    rating int check (rating between 1 and 5),
    created_at timestamp default now(),
    index idx_guest (guest_id, watch_start desc),
    index idx_content (content_id, watch_start desc)
);

-- Content entitlements per guest/room
create table if not exists iptv.entitlements (
    entitlement_id uuid primary key default gen_random_uuid(),
    tenant_id uuid not null,
    guest_id uuid,
    room_id uuid not null,
    reservation_id uuid,
    content_id uuid references iptv.content(content_id),
    channel_id uuid references iptv.channels(channel_id),
    entitlement_type varchar(20) not null check (entitlement_type in ('room_type', 'loyalty_tier', 'package', 'purchase', 'promotional')),
    valid_from timestamp not null,
    valid_until timestamp not null,
    max_concurrent_streams int default 3,
    max_quality varchar(10) default '1080p',
    created_at timestamp default now(),
    index idx_guest_valid (guest_id, valid_from, valid_until),
    index idx_room_valid (room_id, valid_from, valid_until)
);

-- VOD purchases / PPV transactions
create table if not exists iptv.vod_purchases (
    purchase_id uuid primary key default gen_random_uuid(),
    tenant_id uuid not null,
    guest_id uuid not null,
    room_id uuid not null,
    reservation_id uuid not null,
    content_id uuid not null references iptv.content(content_id),
    purchase_price decimal(10,2) not null,
    currency char(3) default 'USD',
    payment_status varchar(20) default 'pending',
    payment_method varchar(50),
    transaction_reference varchar(100),
    folio_item_id uuid,
    watch_count int default 0,
    max_watch_count int default 1,
    valid_until timestamp,
    purchased_at timestamp default now(),
    index idx_guest (guest_id, purchased_at desc),
    index idx_reservation (reservation_id)
);

-- Casting sessions
create table if not exists iptv.casting_sessions (
    cast_id uuid primary key default gen_random_uuid(),
    tenant_id uuid not null,
    room_id uuid not null,
    guest_id uuid not null,
    session_id uuid not null,
    cast_type varchar(20) not null check (cast_type in ('chromecast', 'airplay', 'miracast', 'dlna')),
    device_name varchar(100),
    device_model varchar(100),
    device_ip varchar(50),
    paired_at timestamp default now(),
    disconnected_at timestamp,
    status varchar(20) default 'active',
    bandwidth_limit_mbps int default 50,
    index idx_session (session_id),
    index idx_room (room_id, status)
);

-- Emergency broadcasts log
create table if not exists iptv.emergency_broadcasts (
    broadcast_id uuid primary key default gen_random_uuid(),
    tenant_id uuid not null,
    property_id uuid not null,
    level varchar(20) not null check (level in ('critical', 'urgent', 'important', 'advisory')),
    message text not null,
    sender_id uuid not null,
    sender_type varchar(20) not null,
    affected_rooms int,
    delivered_count int default 0,
    acknowledged_count int default 0,
    duration_seconds int,
    auto_resume boolean default true,
    broadcast_at timestamp default now(),
    cleared_at timestamp,
    index idx_property (property_id, broadcast_at desc),
    index idx_level (level, broadcast_at desc)
);

-- Content recommendations (AI-generated)
create table if not exists iptv.recommendations (
    recommendation_id uuid primary key default gen_random_uuid(),
    tenant_id uuid not null,
    guest_id uuid not null,
    content_id uuid not null references iptv.content(content_id),
    recommendation_type varchar(20) not null check (recommendation_type in ('trending', 'similar', 'personalized', 'staff_pick', 'ai_generated')),
    confidence_score decimal(3,2),
    reason text,
    displayed boolean default false,
    clicked boolean default false,
    watched boolean default false,
    created_at timestamp default now(),
    index idx_guest (guest_id, confidence_score desc),
    index idx_type (recommendation_type, created_at desc)
);

-- Row-level security policies
alter table iptv.content enable row level security;
alter table iptv.channels enable row level security;
alter table iptv.epg enable row level security;
alter table iptv.viewing_history enable row level security;
alter table iptv.entitlements enable row level security;
alter table iptv.vod_purchases enable row level security;

-- Tenant isolation policies
create policy tenant_isolation_content on iptv.content
    using (tenant_id = current_setting('app.current_tenant')::uuid);
create policy tenant_isolation_channels on iptv.channels
    using (tenant_id = current_setting('app.current_tenant')::uuid);
create policy tenant_isolation_epg on iptv.epg
    using (tenant_id = current_setting('app.current_tenant')::uuid);
create policy tenant_isolation_history on iptv.viewing_history
    using (tenant_id = current_setting('app.current_tenant')::uuid);
create policy tenant_isolation_entitlements on iptv.entitlements
    using (tenant_id = current_setting('app.current_tenant')::uuid);
