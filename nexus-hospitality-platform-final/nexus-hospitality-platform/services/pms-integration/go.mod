module github.com/nexus-platform/pms-integration

go 1.22

require (
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/go-chi/chi/v5 v5.0.12
	github.com/golang-migrate/migrate/v4 v4.17.0
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.5
	github.com/prometheus/client_golang v1.19.0
	github.com/segmentio/kafka-go v0.4.47
	github.com/testcontainers/testcontainers-go v0.30.0
	github.com/testcontainers/testcontainers-go/modules/postgres v0.30.0
	go.opentelemetry.io/otel v1.24.0
	go.opentelemetry.io/otel/trace v1.24.0
)
