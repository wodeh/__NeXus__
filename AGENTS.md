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
cd services/backend-core
go mod tidy
go build -o bin/server ./cmd/server
go test -race ./...

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

### Frontend

```bash
cd clients/web-portal
npm install
npm run dev      # Vite dev server on :3000
npm run build
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
