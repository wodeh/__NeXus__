<<<<<<< HEAD
# Nexus Hospitality Platform — Agent Guide

## Repository Structure

This is the **canonical production source** (`nexus-hospitality-platform-final/`). Treat it as the single source of truth for all implementations.

```
services/
  backend-core/         — Core HTTP API server (tenant-aware, observability)
  pms-integration/      — PMS Saga orchestrator, CQRS, event-driven
  ai-platform/          — AI inference gateway with guardrails & RAG
  iot-gateway/          — Multi-protocol IoT gateway (MQTT, Zigbee, BACnet, Modbus, KNX)
  iptv-middleware/      — GPU-accelerated transcode engine with HLS/DVR

clients/
  web-portal/           — React + Vite micro-frontend orchestrator

sdk/
  go-sdk/               — Go client SDK
  typescript-sdk/       — TypeScript client SDK

infrastructure/         — Terraform, Helm, Flux GitOps
observability/          — Prometheus, Grafana, ClickHouse
security/               — Falco, OPA, Vault
```

## Build System

### Go Services

Each Go service is an independent Go module with its own `go.mod`. **Go 1.22** is required.

```bash
# Backend Core
=======
# NeXus Hospitality Platform — Agent Guide

> **Last updated:** 2026-05-09
> **Language of project documentation:** English
> **Reader assumption:** You know nothing about this project.

---

## 1. Project Overview

This repository contains the **NeXus Hospitality Platform** — an enterprise-grade, cloud-native SaaS platform for global hospitality operations. It unifies property management (PMS), guest experience, IPTV/streaming, IoT smart-room infrastructure, and AI-native operations under a single multi-tenant architecture.

The root directory holds **three distinct variants** of the same platform vision. They are organized chronologically from blueprint to canonical source:

| Directory | Purpose | Buildable? |
|-----------|---------|------------|
| `nexus-hospitality-cloud-v2.0/` | Architecture blueprint, PRDs, runbooks, and reference code snippets (numbered sections 01–15). | Partial — contains isolated Go and Python samples, but no top-level build system. |
| `nexus-hospitality-platform-complete/` | High-fidelity scaffold with representative implementations across all domains. | No — missing `go.mod`, `package.json`, `Dockerfile`s, and several referenced directories. |
| `nexus-hospitality-platform-final/` | **Canonical production source.** Contains working Go modules, Makefiles, Docker Compose, CI/CD, and SDKs. | Partial — some backend stubs are unimplemented, and micro-frontend remotes are missing. |

### What each variant contains

- **`nexus-hospitality-cloud-v2.0/`** — Master blueprint (`15-docs/NHC-Master-Blueprint-v2.0.md`), licensing strategy, C4/DDD/sequence diagrams, IPTV architecture, AI/ML service stubs (Python/FastAPI), Terraform modules, Kubernetes manifests, DevSecOps policies (Falco/OPA), and operations runbooks.
- **`nexus-hospitality-platform-complete/`** — Go microservices (AI inference gateway, PMS integration with Saga/CQRS, IPTV transcode engine, IoT gateway), PostgreSQL schema (`schema.sql`), Protocol Buffer event definitions, React web portal scaffold, Terraform multi-region IaC, Helm charts, Prometheus/Grafana configs, and testing suites (Playwright, K6, Locust, Chaos Mesh).
- **`nexus-hospitality-platform-final/`** — The most complete codebase with independent Go modules per service, a working `Makefile`, `docker-compose.dev.yml`, GitHub Actions CI, React + Vite web portal with `package.json`, Go SDK and TypeScript SDK stubs, and operational scripts (zero-touch provisioning, PMS migration factory).

**Treat `nexus-hospitality-platform-final/nexus-hospitality-platform/` as the single source of truth** when implementing or modifying code. The other two directories provide architectural context and historical design rationale.

---

## 2. Technology Stack

### Backend
- **Go 1.22** — Microservices (each service is an independent Go module)
- **Python 3.x** — AI/ML pipelines (`numpy`, `pandas`, `scikit-learn`, `torch`, `transformers`, `sentence-transformers`, `mlflow`)
- **gRPC + Protocol Buffers** — Internal service communication (`services/pms-integration/proto/events/v1/pms_events.proto`)
- **GraphQL** — Flexible client queries (`nexus-hospitality-cloud-v2.0/06-api/graphql/schema.graphql`)
- **REST/OpenAPI** — External integrations (`nexus-hospitality-cloud-v2.0/06-api/rest-specs/openapi.yaml`)

### Frontend
- **React 18** + **TypeScript** (strict mode)
- **Vite** — Build tool and dev server
- **TanStack React Query** — Server-state management
- **Jotai** — Client-state management
- **react-i18next** — Internationalization
- **Vitest** — Unit testing

### Data & Messaging
- **PostgreSQL 16** — Transactional data (with TimescaleDB extensions for hypertables)
- **Redis 7** — Cache, sessions, rate limiting
- **ClickHouse 23.8** — Time-series analytics
- **Apache Kafka 7.5.0** (Confluent) — Event streaming
- **Elasticsearch / MongoDB / Neo4j** — Mentioned in v2.0 architecture

### Infrastructure
- **Docker / Docker Compose** — Local development
- **Kubernetes (EKS)** — Production orchestration
- **Terraform >= 1.6.0** — Multi-region IaC (AWS)
- **Helm + Flux GitOps** — Deployment
- **Karpenter + KEDA** — Autoscaling

### Observability
- **Prometheus + Grafana 10.1** — Metrics and dashboards
- **Loki + Fluentd** — Log aggregation
- **Jaeger / OpenTelemetry** — Distributed tracing

### Security
- **Falco** — Runtime threat detection
- **OPA / OPA Gatekeeper** — Admission control policies
- **HashiCorp Vault** — PKI, secrets, mTLS
- **Istio** — Service mesh with STRICT mTLS
- **Keycloak** — Identity and SSO (OIDC/SAML/LDAP)

### Streaming & Media (IPTV)
- **FFmpeg (NVENC)** — GPU-accelerated transcoding
- **NGINX** — HLS/DASH/RTMP streaming edge
- **Envoy** — Edge proxy
- **Widevine / FairPlay / PlayReady / AES-128** — DRM

### IoT
- **MQTT (Eclipse Paho)** — Primary messaging
- **BACnet, Modbus, KNX, Zigbee** — Device protocols

---

## 3. Repository Structure (Canonical Source)

```
nexus-hospitality-platform-final/nexus-hospitality-platform/
├── .github/workflows/          # CI/CD: ci-backend.yml, ci-ai.yml, ci-iptv.yml, release-orchestrator.yml
├── services/
│   ├── backend-core/           # Core HTTP API server (tenant-aware)
│   │   ├── cmd/server/main.go
│   │   ├── internal/config/
│   │   ├── internal/server/
│   │   ├── internal/health/
│   │   ├── internal/metrics/
│   │   ├── .golangci.yml
│   │   ├── Dockerfile
│   │   ├── Dockerfile.dev
│   │   └── .air.toml
│   ├── pms-integration/        # PMS Saga orchestrator, CQRS, event-driven
│   │   ├── cmd/pms/main.go
│   │   ├── internal/saga/orchestrator.go
│   │   ├── internal/events/
│   │   ├── internal/store/
│   │   ├── internal/config/
│   │   ├── proto/events/v1/pms_events.proto
│   │   ├── sql/schema.sql
│   │   ├── Dockerfile
│   │   └── .air.toml
│   ├── ai-platform/            # AI inference gateway with guardrails & RAG
│   │   ├── inference-gateway/cmd/gateway/main.go
│   │   ├── inference-gateway/internal/config/
│   │   ├── inference-gateway/internal/guardrails/
│   │   ├── inference-gateway/internal/metrics/
│   │   ├── inference-gateway/internal/modelrouter/
│   │   ├── inference-gateway/internal/rag/
│   │   ├── ml-pipelines/requirements.txt
│   │   ├── Dockerfile
│   │   └── .air.toml
│   ├── iot-gateway/            # Multi-protocol IoT gateway
│   │   ├── cmd/gateway/main.go
│   │   ├── internal/config/
│   │   ├── internal/registry/
│   │   ├── internal/telemetry/
│   │   ├── internal/bacnet/
│   │   ├── internal/modbus/
│   │   ├── internal/knx/
│   │   ├── internal/zigbee/
│   │   ├── internal/ota/
│   │   ├── Dockerfile
│   │   └── .air.toml
│   └── iptv-middleware/        # GPU-accelerated transcode engine (HLS/DVR)
│       ├── transcode-engine/cmd/transcoder/main.go
│       ├── transcode-engine/internal/gpu/manager.go
│       ├── transcode-engine/internal/config/
│       ├── Dockerfile
│       └── .air.toml
├── clients/
│   ├── shared/design-system/   # Shared UI tokens and components
│   └── web-portal/             # React + Vite micro-frontend orchestrator
│       ├── src/App.tsx
│       ├── src/hooks/
│       ├── src/layouts/
│       ├── src/components/
│       ├── src/i18n/
│       ├── package.json
│       ├── vite.config.ts
│       ├── tsconfig.json
│       └── Dockerfile.dev
├── sdk/
│   ├── go-sdk/                 # Go client library (go.mod)
│   └── typescript-sdk/         # TypeScript client library (package.json)
├── infrastructure/
│   ├── terraform/              # Multi-region IaC
│   ├── helm-charts/            # K8s deployments
│   └── gitops/                 # Flux CD configuration
├── security/
│   ├── falco/rules/            # Runtime security rules
│   ├── policies/opa/           # Admission control policies
│   └── vault/                  # PKI hierarchy
├── observability/
│   ├── monitoring/             # Prometheus rules + Grafana dashboards
│   └── analytics/              # ClickHouse schemas
├── testing/
│   ├── e2e/playwright/         # E2E tests
│   ├── load/scripts/           # K6 + Locust load tests
│   └── chaos/experiments/      # Chaos engineering experiments
└── scripts/
    ├── bootstrap/              # Zero-touch property provisioning
    └── migration/              # PMS migration factory (OPERA, Mews, Cloudbeds, etc.)
```

---

## 4. Build and Test Commands

All commands below assume you are inside `nexus-hospitality-platform-final/nexus-hospitality-platform/`.

### Local Development

```bash
# Start the full local stack (Postgres, Redis, Kafka, ClickHouse, all services, observability)
make dev-up

# Tear down local stack
make dev-down
```

### Go Services

Each service is an independent module. **Go 1.22 is required.**

```bash
# Example: backend-core
>>>>>>> phase1/security-stability
cd services/backend-core
go mod tidy
go build -o bin/server ./cmd/server
go test -race ./...
<<<<<<< HEAD

# PMS Integration
cd services/pms-integration
go mod tidy
go build -o bin/pms ./cmd/pms
go test -race ./...

# AI Platform
cd services/ai-platform
go mod tidy
go build -o bin/gateway ./inference-gateway/cmd/gateway
go test -race ./...

# IoT Gateway
cd services/iot-gateway
go mod tidy
go build -o bin/iot-gateway ./cmd/gateway
go test -race ./...

# IPTV Middleware
cd services/iptv-middleware
go mod tidy
go build -o bin/transcoder ./transcode-engine/cmd/transcoder
go test -race ./...
```

=======
```

Repeat the above pattern for:
- `services/pms-integration`
- `services/ai-platform`
- `services/iot-gateway`
- `services/iptv-middleware`

>>>>>>> phase1/security-stability
### Frontend

```bash
cd clients/web-portal
npm install
npm run dev      # Vite dev server on :3000
npm run build
<<<<<<< HEAD
npm run test
```

### Docker (Development)

```bash
make dev-up      # docker-compose -f docker-compose.dev.yml up -d
make dev-down    # docker-compose -f docker-compose.dev.yml down -v
make build       # Build all production images
make test        # Run all Go tests + E2E
```

## Service Architecture Rules

1. **Tenant Isolation**: Every service must extract `X-Tenant-ID` header and enforce isolation at the data layer. No cross-tenant queries.
2. **Observability**: Every service exposes `/health`, `/ready`, `/live` and Prometheus metrics on a separate port.
3. **Structured Logging**: Use `log/slog` with JSON handler. Include `tenant_id` in every log entry where applicable.
4. **Graceful Shutdown**: All services handle `SIGINT`/`SIGTERM` with a 15-second timeout for in-flight requests.
5. **Concurrency Safety**: All shared state must use `sync.RWMutex` or channels. No unprotected maps.
6. **Configuration**: Environment-variable based with `internal/config` package. Validate at startup.

## Internal Package Conventions

| Package      | Purpose                              | Required In |
|--------------|--------------------------------------|-------------|
| `config`     | Env-based config, validation         | Every service |
| `metrics`    | Prometheus Collector (tenant-aware)  | Every service |
| `health`     | K8s probe endpoints                  | backend-core, iptv-middleware |
| `server`     | HTTP router with middleware          | backend-core |

## Phase 1 Status

### Completed
- [x] `go.mod` + `go.sum` (empty, run `go mod tidy`) for all 5 Go services
- [x] Missing internal package scaffolding (config, metrics, health, etc.)
- [x] Missing entry points: `backend-core/cmd/server/main.go`, `pms-integration/cmd/pms/main.go`
- [x] Dockerfile.dev for all services + web-portal
- [x] `.air.toml` hot-reload configs
- [x] Web-portal `package.json`, `vite.config.ts`, `tsconfig.json`
- [x] Web-portal missing hook/component stubs
- [x] PMS saga circular dependency resolved (`SagaStore` interface moved into `saga` package)
- [x] IoT gateway protocol handler interface mismatch fixed
- [x] Unit tests for all new internal packages

### Remaining Blockers
1. **go.sum population**: Empty `go.sum` files must be populated by running `go mod tidy` in each service directory (requires Go 1.22 installed).
2. **Web-portal micro-frontends**: The `App.tsx` lazy-loads federated modules (`guestPortal/GuestPortal`, etc.) that do not exist. These are declared as Module Federation remotes — the host/remote configs need to be added for a full build.
3. **AI backend stubs**: `callOpenAI`, `callAnthropic`, `callLocalModel`, `callCustomModel` in `ai-platform` are empty stubs. They compile but do not perform inference.
4. **IoT device identity validation**: `validateDeviceIdentity` and `verifyTenantIsolation` in `iot-gateway` are empty stubs.
5. **PMS compensators**: All compensator implementations (`ReservationCompensator`, etc.) are empty stubs.
6. **GPU manager NVML dependency**: `iptv-middleware` imports `github.com/NVIDIA/go-nvml`. This requires CUDA/NVML headers for compilation; builds will fail without them. Use build tags (`//go:build nvml`) to gate this if needed.
7. **Backend-core database integration**: No actual database connection or repository layer exists yet.

## Next Recommended Steps
1. Run `go mod tidy` in all 5 service directories to populate `go.sum`.
2. Verify compilation with `go build ./...` in each service.
3. Implement the missing micro-frontend remotes or provide fallback components.
4. Add PostgreSQL connection pooling and migration runner to `backend-core`.
5. Implement concrete compensators for PMS Saga with idempotency keys.
6. Add Kafka/Redpanda producer/consumer implementations for the event bus.
=======
npm run test     # Vitest
npm run lint
```

### Testing (Root Makefile)

| Command | Description |
|---------|-------------|
| `make test` | Run all Go unit tests with race detector + E2E tests |
| `make lint` | Run `golangci-lint`, `npm run lint`, OPA `conftest` |
| `make security-scan` | Run Trivy, Snyk, Checkov |
| `make chaos-test` | Apply Chaos Mesh experiments |
| `make load-test` | Run Locust + K6 against staging |

### Docker Images

```bash
make build       # Build all 5 service images → ghcr.io/nexus-platform
```

### Deployment

```bash
make deploy-staging     # Terraform + Helm to staging
make deploy-prod        # Terraform + Helm to production (interactive confirm)
```

### Operations

```bash
make migrate            # Run PostgreSQL schema + migration scripts
make provision          # Interactive zero-touch property provisioning
make backup             # pg_dump from Kubernetes
```

---

## 5. Code Style Guidelines

### Go
- Follow **Effective Go**.
- Run `golangci-lint` before committing (configured via `.golangci.yml` in `services/backend-core/`).
- Enabled linters: `errcheck`, `gosimple`, `govet`, `ineffassign`, `staticcheck`, `unused`, `gocritic`, `gofmt`, `goimports`, `misspell`, `revive`, `unconvert`.
- Use `log/slog` with **JSON handler** for structured logging.
- Include `tenant_id` in every log entry where applicable.
- Environment-based configuration only; validate at startup in `internal/config`.
- Handle `SIGINT`/`SIGTERM` with a **15-second timeout** for graceful shutdown.
- Use `sync.RWMutex` or channels for shared state. **No unprotected maps.**

### TypeScript / React
- **Strict mode** enabled.
- ESLint + Prettier for formatting.
- Use functional components and hooks.

### SQL
- **Migrations only.** Never modify an existing migration file after it has been committed.
- PostgreSQL Row-Level Security (RLS) must be used for multi-tenant tables.
- The PMS schema (`services/pms-integration/sql/schema.sql`) includes TimescaleDB hypertables, RLS policies, audit triggers, and `updated_at` automation.

### Terraform
- Run `terraform fmt` and `tflint` before committing.

### Commit Message Format

```
type(scope): subject

body (optional)

footer (optional)
```

**Types:** `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, `chore`

**Example:**
```
feat(iptv): add SCTE-35 ad insertion support

Implements splice_insert and time_signal marker types.
Supports pre-roll and mid-roll ad breaks.

Closes: ABC-123
```

---

## 6. Testing Instructions

### Coverage Requirements
- **Unit tests: >80% coverage** (enforced in CI).
- **Integration tests:** Required for all API changes.
- **E2E tests (Playwright):** Required for all UI changes.
- **Chaos tests:** Required for all infrastructure changes.

### CI/CD Pipeline (`.github/workflows/ci-backend.yml`)
1. **Lint** — `golangci-lint` (10-minute timeout).
2. **Test** — Go race-detector tests with coverage upload to Codecov. Runs with PostgreSQL, Redis, and Kafka service containers.
3. **Security Scan** — Trivy filesystem scan (SARIF output uploaded to GitHub).
4. **Build** — Docker Buildx with SBOM and provenance, pushed to `ghcr.io/nexus-platform`.

Triggers on push to `main`, `release/*`, `hotfix/*`.

### Load Test Thresholds
- P95 latency < 500ms
- P99 latency < 1s
- Error rate < 1%

---

## 7. Security Considerations

### Secrets Management
- **Never commit secrets.** Use `external-secrets` operator or HashiCorp Vault.
- All commits must be **GPG signed**.

### Runtime Security (Falco)
Rules detect:
- Unauthorized database access
- PII exfiltration
- IPTV DRM tampering
- IoT firmware tampering
- Crypto-mining
- Reverse shells and container escapes
- Tenant isolation breaches

### Admission Control (OPA/Rego)
Policies enforce:
- Resource limits required
- No privileged/root containers
- Read-only root filesystem
- Drop ALL capabilities
- seccomp profiles
- No `latest` image tags
- Mandatory liveness/readiness probes
- Trusted registries only
- SSL for PMS services, MQTT TLS for IoT

### Network
- Istio **STRICT mTLS** between services.
- Tenant-isolated Kubernetes `NetworkPolicies`.

### Compliance Targets
- PCI-DSS Level 1
- GDPR
- SOC 2 Type II
- ISO 27001
- HIPAA
- FedRAMP

---

## 8. Service Architecture Rules (Canonical Source)

1. **Tenant Isolation** — Extract `X-Tenant-ID` header and enforce data-layer isolation. No cross-tenant queries.
2. **Observability** — Every service exposes `/health`, `/ready`, `/live` and Prometheus metrics on a separate port.
3. **Structured Logging** — `log/slog` JSON handler with `tenant_id`.
4. **Graceful Shutdown** — 15-second timeout for in-flight requests on `SIGINT`/`SIGTERM`.
5. **Concurrency Safety** — `sync.RWMutex` or channels. No unprotected maps.
6. **Configuration** — Environment variables via `internal/config`. Validate at startup.

### Required Internal Packages

| Package | Purpose | Required In |
|---------|---------|-------------|
| `config` | Env-based config, validation | Every service |
| `metrics` | Prometheus Collector (tenant-aware) | Every service |
| `health` | K8s probe endpoints | backend-core, iptv-middleware |
| `server` | HTTP router with middleware | backend-core |

---

## 9. Deployment Conventions

- **Canary rollout:** `5% → 25% → 50% → 100%`
- **Automated rollback** on error rate > 1%
- **Post-deployment monitoring:** 30 minutes
- **Hotfix branches:** `hotfix/ABC-456-critical-fix` with fast-track review (1 approval)

---

## 10. Known Blockers and Incomplete Areas

The canonical source (`nexus-hospitality-platform-final/`) has the following known gaps:

1. **Web-portal micro-frontends** — `App.tsx` lazy-loads federated modules (`guestPortal/GuestPortal`, etc.) that do not exist. Module Federation host/remote configs are incomplete; `vite.config.ts` does not include the module federation plugin.
2. **AI backend stubs** — `callOpenAI`, `callAnthropic`, `callLocalModel`, `callCustomModel` in `ai-platform` are empty stubs.
3. **IoT device identity validation** — `validateDeviceIdentity` and `verifyTenantIsolation` in `iot-gateway` are empty stubs.
4. **PMS compensators** — All compensator implementations (`ReservationCompensator`, `BillingCompensator`, etc.) are empty stubs.
5. **GPU manager NVML dependency** — `iptv-middleware` imports `github.com/NVIDIA/go-nvml`. Requires CUDA/NVML headers. Use build tags (`//go:build nvml`) if compiling without them.
6. **Backend-core database integration** — No actual database connection or repository layer exists yet.
7. **Kafka/Redpanda producer and consumer** — Event bus implementations are missing.
8. **CI workflow paths** — `.github/workflows/ci-backend.yml` references `services/notification-engine/**` which does not exist.
9. **Makefile references missing directory** — `make lint` references `clients/mobile-apps` which does not exist in the canonical source.

---

## 11. Next Recommended Steps for Agents

1. Verify compilation with `go build ./...` in each service directory.
2. Implement missing micro-frontend remotes or provide fallback components in `web-portal`.
3. Add PostgreSQL connection pooling and migration runner to `backend-core`.
4. Implement concrete compensators for PMS Saga with idempotency keys.
5. Add Kafka producer/consumer implementations for the event bus.
>>>>>>> phase1/security-stability
