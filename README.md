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
