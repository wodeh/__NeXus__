# Nexus Hospitality Platform
## Enterprise-Grade SaaS for Global Hospitality Operations

### Architecture Overview
```
┌─────────────────────────────────────────────────────────────┐
│                    GLOBAL LOAD BALANCER                       │
│              (Latency-based DNS + Health Checks)              │
└──────────────┬──────────────────────────────┬───────────────┘
               │                              │
    ┌──────────▼──────────┐      ┌───────────▼──────────┐
    │   US-EAST-1         │      │   EU-WEST-1         │
    │   (Primary)         │      │   (DR/Active)       │
    │                     │      │                     │
    │  ┌───────────────┐  │      │  ┌───────────────┐  │
    │  │ EKS Cluster   │  │      │  │ EKS Cluster   │  │
    │  │               │  │      │  │               │  │
    │  │ • Backend API │  │      │  │ • Backend API │  │
    │  │ • IPTV Trans  │  │      │  │ • IPTV Trans  │  │
    │  │ • AI Inference│  │      │  │ • AI Inference│  │
    │  │ • IoT Gateway │  │      │  │ • IoT Gateway │  │
    │  │ • PMS Core    │  │      │  │ • PMS Core    │  │
    │  └───────────────┘  │      │  └───────────────┘  │
    │                     │      │                     │
    │  ┌───────────────┐  │      │  ┌───────────────┐  │
    │  │ Data Layer    │  │      │  │ Data Layer    │  │
    │  │ • PostgreSQL  │  │      │  │ • PostgreSQL  │  │
    │  │ • Redis       │  │      │  │ • Redis       │  │
    │  │ • Kafka       │  │      │  │ • Kafka       │  │
    │  │ • ClickHouse  │  │      │  │ • ClickHouse  │  │
    │  └───────────────┘  │      │  └───────────────┘  │
    └─────────────────────┘      └─────────────────────┘
```

### Repository Structure
```
nexus-hospitality-platform/
├── .github/                    # CI/CD workflows, CODEOWNERS
├── services/
│   ├── backend-core/           # Go microservices
│   ├── iptv-middleware/        # FFmpeg, NGINX, Envoy
│   ├── ai-platform/            # MLFlow, Inference Gateway
│   ├── iot-gateway/            # MQTT, BACnet, Modbus
│   └── pms-integration/        # CQRS, Saga, Events
├── clients/
│   ├── web-portal/             # React, Micro-frontends
│   ├── mobile-apps/            # React Native
│   ├── tv-apps/                # Android TV, webOS, Tizen, tvOS
│   └── staff-console/          # Operations dashboard
├── infrastructure/
│   ├── terraform/              # Multi-region IaC
│   ├── helm-charts/            # Kubernetes deployments
│   └── gitops/                 # Flux CD configuration
├── security/
│   ├── falco/                  # Runtime security rules
│   ├── policies/opa/           # Admission control
│   └── vault/                  # PKI hierarchy
├── observability/
│   ├── monitoring/             # Prometheus, Grafana
│   ├── logging/                # Loki, Fluentd
│   └── tracing/                # Tempo, OpenTelemetry
├── testing/
│   ├── e2e/                    # Cypress, Playwright
│   ├── load/                   # Locust, K6
│   └── chaos/                  # Chaos Mesh experiments
├── sdk/
│   ├── go-sdk/                 # Go client library
│   └── typescript-sdk/         # TypeScript client
└── scripts/
    ├── bootstrap/              # Zero-touch provisioning
    ├── migration/              # PMS migration factory
    └── maintenance/            # Operational scripts
```

### Quick Start
```bash
# Clone repository
git clone https://github.com/nexus-platform/nexus-hospitality-platform.git
cd nexus-hospitality-platform

# Bootstrap local development
make dev-up

# Run tests
make test

# Deploy to staging
make deploy-staging

# Provision new property
./scripts/bootstrap/zero-touch-provision.sh     tenant-123 property-456 us-east-1 production enterprise
```

### Supported PMS Migrations
- Oracle OPERA (v5, Cloud)
- Mews Commander
- Cloudbeds
- protel
- Fidelio/Opera v4
- SynXis
- HotelRunner

### Compliance
- PCI-DSS Level 1
- GDPR (EU)
- SOC 2 Type II
- ISO 27001
- HIPAA (healthcare hospitality)
- FedRAMP (US government)

### Support
- Documentation: https://docs.nexus-platform.com
- API Reference: https://api.nexus-platform.com/docs
- Status Page: https://status.nexus-platform.com
- Security: security@nexus-platform.com

---

## NeXus PMS Integration Service

The `services/pms-integration` directory contains the core Property Management System — a Go backend with a Next.js 15 frontend.

### Backend (Go 1.22)

**Tech Stack:** Chi router, PostgreSQL 16 (RLS multi-tenant), pgx v5, JWT, bcrypt, in-memory repositories for demo.

**Modules:**

| Module | Endpoints | Capability |
|---|---|---|
| Properties | GET/POST/PUT `/tenants/:id/properties` | `core:properties` |
| Rooms | GET/POST/PUT/PATCH `/tenants/:id/properties/:pid/rooms` | `core:rooms` |
| Room Types | GET/POST `/tenants/:id/properties/:pid/room-types` | `core:rooms` |
| Guests | GET/POST/PATCH `/tenants/:id/guests` | `core:guests` |
| Reservations | GET/POST/PATCH `/tenants/:id/reservations` | `core:reservations` |
| Folios | GET/POST `/tenants/:id/folios` | `core:reservations` |
| Housekeeping | PATCH `/tenants/:id/properties/:pid/rooms/:rid/housekeeping` | `core:housekeeping` |
| Rate Plans | GET/POST/PATCH/DELETE `/tenants/:id/rate-plans` | `revenue:dynamic_pricing` |
| Revenue Forecasts | GET `/tenants/:id/revenue/forecasts` | `revenue:revenue_forecasting` |
| Dynamic Pricing | GET/POST `/tenants/:id/revenue/pricing-rules` | `revenue:dynamic_pricing` |
| Room Blocks | GET/POST/PATCH/DELETE `/tenants/:id/room-blocks` | `operations:room_blocks` |
| Group Reservations | GET/POST/PATCH/DELETE `/tenants/:id/groups` | `operations:group_reservations` |
| Guest CRM | GET/POST/PATCH `/tenants/:id/guest-profiles` | `enterprise:advanced_crm` |
| Agents | GET/POST/PATCH/DELETE `/tenants/:id/agents` | `revenue:agent_management` |
| Audit Logs | GET/POST `/tenants/:id/audit` | `core:audit_logs` |
| GDPR | GET/POST/DELETE `/tenants/:id/gdpr` | `core:audit_logs` |
| Check-In/Out | GET/POST/PATCH `/tenants/:id/checkins`, `/checkouts` | `core:reservations` |
| Invoices | GET/POST/PATCH/DELETE `/tenants/:id/invoices` | `core:reservations` |

**License Tiers:**

| Tier | Price | Rooms | Users | Key Features |
|---|---|---|---|---|
| Core | $99/mo | 50 | 5 | Reservations, guests, properties, rooms, housekeeping |
| Operations | $199/mo | 150 | 20 | + Floor dashboard, room blocks, groups, maintenance, front desk |
| Revenue | $349/mo | 500 | 50 | + Dynamic pricing, OTA integration, forecasting, agent management |
| Enterprise | $599/mo | 5000 | 200 | + Multi-property, advanced CRM, API access, white label, custom reports |

**Property Types:** boutique, motel, resort, hostel, aparthotel, bnb — each gets type-specific capabilities.

### Frontend (Next.js 15)

**Tech Stack:** React 19, TypeScript, Tailwind CSS v4, Lucide icons.

**Pages:**

| Page | Path | License Required |
|---|---|---|
| Dashboard | `/` | Core |
| Floor Plan | `/floor` | Operations |
| Reservations | `/reservations` | Core |
| Guests | `/guests` | Core |
| Room Blocks | `/room-blocks` | Operations |
| Group Reservations | `/groups` | Operations |
| Guest CRM | `/crm` | Enterprise |
| Agents & Partners | `/agents` | Revenue |
| Properties | `/properties` | Core |
| Rooms | `/rooms` | Core |
| Housekeeping | `/housekeeping` | Core |
| Revenue | `/revenue` | Revenue |
| Audit | `/audit` | Core |
| Settings | `/settings` | Core |

### Running Locally

```bash
# Backend
cd services/pms-integration
go build -o pms-api ./cmd/pms
DATABASE_URL=postgres://user:pass@localhost:5432/pms ./pms-api

# Frontend
cd apps/web
npm install
npm run dev
```

Backend runs on `localhost:8080`. Frontend runs on `localhost:3000` and proxies API calls to `:8080`.

### API Testing

```powershell
# Health checks
Invoke-RestMethod http://localhost:8080/health
Invoke-RestMethod http://localhost:8080/ready
Invoke-RestMethod http://localhost:8080/live

# Tenant config (no DB needed)
Invoke-RestMethod http://localhost:8080/tenants/demo/config

# List guests (needs PostgreSQL)
Invoke-RestMethod http://localhost:8080/tenants/demo/guests
```

### Docker Compose

```bash
docker-compose up -d
```

Services: PostgreSQL 16, Redis 7, PMS API (:8080), Next.js frontend (:3000), Prometheus (:9090), Grafana (:3001).

### Demo Credentials
- Tenant: `demo` (Enterprise tier, all features)
- Property: `p-001` (Grand Plaza Hotel)
- Pre-seeded with 15 rooms, 2 guests, 2 reservations, 2 room blocks, 2 groups, 2 guest profiles, 4 agents.
