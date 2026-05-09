# Guru Hospitality Platform — NeXus PMS

**All-in-one property management system for modern hotels.**

Built from scratch to scale to hundreds of millions in revenue within 3 years of launch.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           GURU HOSPITALITY PLATFORM                       │
│                              (NeXus PMS)                                 │
├─────────────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌───────────┐ │
│  │   Frontend   │  │   IPTV       │  │   Mobile     │  │  Nginx    │ │
│  │  Next.js 15  │  │  Guest TV    │  │   App        │  │  Proxy    │ │
│  │  React Query │  │  In-Room     │  │  (future)    │  │  + SSL    │ │
│  │  Tailwind    │  │  Commerce    │  │              │  │           │ │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └─────┬─────┘ │
│         │                 │                 │                │       │
│         └─────────────────┴─────────────────┴────────────────┘       │
│                                    │                                   │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │              PMS Integration API (Go + Chi Router)            │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │   │
│  │  │ Property │ │Reservation│ │  Folio   │ │  Rate    │          │   │
│  │  │ Inventory│ │  Mgmt    │ │ Billing  │ │  Mgmt    │          │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘          │   │
│  │  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐          │   │
│  │  │ House-   │ │Maintenance│ │  Auth    │ │  Smart   │          │   │
│  │  │ keeping │ │   Mgmt    │ │  JWT     │ │  Locks   │          │   │
│  │  └──────────┘ └──────────┘ └──────────┘ └──────────┘          │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                    │                                   │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    Background Worker (Go)                       │   │
│  │   Redis Streams → Folio Recalc → Key Sync → HK Schedule → Reports│   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                    │                                   │
│         ┌──────────────────────────┴──────────────────────────┐        │
│         │                                                     │        │
│  ┌──────┴──────┐                                     ┌──────┴──────┐  │
│  │  PostgreSQL │                                     │    Redis    │  │
│  │  (tenant    │                                     │  (cache,   │  │
│  │   isolation │                                     │   queues)   │  │
│  │   via RLS)  │                                     │             │  │
│  └─────────────┘                                     └─────────────┘  │
│                                                                         │
│  ┌─────────────┐  ┌─────────────┐                                      │
│  │ Prometheus  │  │   Grafana   │                                      │
│  │  Metrics    │  │ Dashboards  │                                      │
│  └─────────────┘  └─────────────┘                                      │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Revenue Model — $100M+ in 3 Years

| Revenue Stream | Year 1 | Year 2 | Year 3 |
|---|---|---|---|
| **PMS SaaS** ($5-15/room/month) | $2M | $8M | $20M |
| **Channel Manager Commission** (1-2% of bookings) | $500K | $3M | $10M |
| **Guest Upsells via IPTV** (room service, amenities) | $200K | $2M | $8M |
| **Smart Lock Hardware SaaS** ($2-5/room/month) | $100K | $1M | $5M |
| **Revenue Management Engine** ($50-200/property/month) | $50K | $500K | $3M |
| **API Licensing** (OTA integrations, custom dev) | $0 | $500K | $2M |
| **Total** | **~$2.9M** | **~$15M** | **~$48M** |

With 20% YoY growth and enterprise expansion, **$100M ARR by Year 5** is achievable.

---

## Tech Stack

| Layer | Technology |
|---|---|
| **Frontend** | Next.js 15, React 19, TypeScript, Tailwind CSS, shadcn/ui |
| **Backend** | Go 1.22, Chi Router, PostgreSQL 16, Redis 7 |
| **Auth** | JWT (access/refresh tokens), bcrypt, tenant isolation via RLS |
| **Realtime** | Redis Streams for async jobs, Server-Sent Events (future) |
| **Observability** | Prometheus, Grafana, structured logging |
| **Infrastructure** | Docker Compose, Nginx, Let's Encrypt |

---

## Quick Start

```bash
# 1. Clone
git clone https://github.com/wodeh/__NeXus__.git
cd __NeXus__/nexus-hospitality-platform-final/nexus-hospitality-platform

# 2. Environment
cp .env.example .env
# Edit .env with your ORBITA_API_KEY, JWT_SECRET, etc.

# 3. Launch everything
docker-compose up --build

# 4. Access
# Frontend:  http://localhost:3000
# API:       http://localhost:8080
# Grafana:   http://localhost:3001 (admin/nexus_admin)
# Prometheus: http://localhost:9090
```

---

## API Overview

All endpoints are prefixed with `/api/v1`.

### Auth
- `POST /auth/register` — Create tenant + admin user
- `POST /auth/login` — Authenticate, receive JWT
- `POST /auth/refresh` — Refresh access token
- `GET /auth/me` — Current user profile

### Property Inventory
- `GET/POST /tenants/{tenantId}/properties`
- `GET/POST /tenants/{tenantId}/room-types`
- `GET/POST /tenants/{tenantId}/rooms`

### Reservations
- `POST /tenants/{tenantId}/reservations`
- `GET /tenants/{tenantId}/reservations/{id}`
- `POST /tenants/{tenantId}/reservations/{id}/checkin`
- `POST /tenants/{tenantId}/reservations/{id}/checkout`

### Rate Management
- `GET/POST /tenants/{tenantId}/rate-plans`
- `GET/POST /tenants/{tenantId}/rate-plans/{id}/daily-rates`

### Folio & Billing
- `POST /tenants/{tenantId}/folios`
- `GET /tenants/{tenantId}/folios/{id}`
- `POST /tenants/{tenantId}/charges`
- `POST /tenants/{tenantId}/payments`

### Housekeeping
- `GET/POST /tenants/{tenantId}/housekeeping-tasks`
- `POST /tenants/{tenantId}/housekeeping-tasks/{id}/start`
- `POST /tenants/{tenantId}/housekeeping-tasks/{id}/complete`

### Maintenance
- `GET/POST /tenants/{tenantId}/work-orders`
- `POST /tenants/{tenantId}/work-orders/{id}/assign`

### Smart Locks
- `GET/POST /tenants/{tenantId}/smart-locks`
- `POST /tenants/{tenantId}/digital-keys`
- `POST /tenants/{tenantId}/digital-keys/{id}/revoke`
- `POST /webhooks/tenants/{tenantId}/lock-events`

---

## Database Schema

Three migration waves:

1. **000001** — Core inventory (properties, rooms, room_types, reservations, guests)
2. **000002** — Rate management (rate_plans, daily_rates)
3. **000003** — Folio, billing, housekeeping, maintenance + audit log

All tables include `tenant_id` with Row-Level Security (RLS) policies for complete multi-tenant isolation.

---

## Smart Lock Integration

Pluggable service layer for Orbita API (and future vendors):

- **Create Digital Key** — Generate key code + PIN, sync to Orbita
- **Revoke Digital Key** — Disable key on Orbita, mark revoked locally
- **Webhook Receiver** — Receive lock events (unlock, low battery, tamper)
- **Guest Phone Key** — SMS/APP push with key code and PIN

---

## Background Jobs

Redis Streams-based worker processes:

| Job | Interval | Description |
|---|---|---|
| **Folio Recalc** | 30s | Recalculate all open folio balances |
| **Key Sync** | 60s | Sync pending digital keys with Orbita API |
| **HK Schedule** | 5min | Auto-create cleaning tasks for checked-out rooms |
| **Nightly Report** | 24h | Generate revenue and occupancy reports |

---

## Frontend Screens

| Screen | Path | Features |
|---|---|---|
| **Dashboard** | `/` | Stats, occupancy, revenue, quick actions |
| **Reservations** | `/reservations` | Grid, search, status filters |
| **Rooms** | `/rooms` | Status board, floor filters |
| **Housekeeping** | `/housekeeping` | Task list, status filters |
| **Rates** | `/rates` | Monthly calendar, plan selector |
| **Login** | `/login` | JWT auth, tenant onboarding |
| **IPTV** | `/iptv` | Guest in-room experience |
| **IPTV Lock** | `/iptv/lock` | Door lock/unlock |
| **IPTV Dining** | `/iptv/dining` | Room service ordering |

---

## Security

- JWT access tokens (15min TTL) + refresh tokens (7 days)
- Bcrypt password hashing
- Tenant isolation via PostgreSQL RLS
- Rate limiting (future: Redis-based)
- Audit log for all mutations
- HTTPS via Nginx + Let's Encrypt

---

## License

Proprietary — Guru Hospitality Platform.

Built with ❤️ to redefine hotel operations.

---

*"Don't worry. Even if the world forgets, I'll remember for you."* — NeXus
