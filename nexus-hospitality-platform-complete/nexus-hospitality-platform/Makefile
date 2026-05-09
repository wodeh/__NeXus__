# Nexus Hospitality Platform — Makefile
.PHONY: all dev-up dev-down test build deploy-staging deploy-prod         lint security-scan chaos-test load-test migrate provision

# Variables
REGISTRY := ghcr.io/nexus-platform
VERSION := $(shell git describe --tags --always --dirty)
TERRAFORM_DIR := infrastructure/terraform
HELM_DIR := infrastructure/helm-charts

# Development
all: lint test build

dev-up:
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Development environment ready at http://localhost:3000"

dev-down:
	docker-compose -f docker-compose.dev.yml down -v

# Testing
test:
	@echo "Running unit tests..."
	cd services/backend-core && go test -race ./...
	cd services/iptv-middleware && go test -race ./...
	cd services/ai-platform && go test -race ./...
	cd services/iot-gateway && go test -race ./...
	cd services/pms-integration && go test -race ./...
	@echo "Running E2E tests..."
	cd testing/e2e && npm run test:ci

lint:
	@echo "Running linters..."
	cd services/backend-core && golangci-lint run
	cd clients/web-portal && npm run lint
	cd clients/mobile-apps && npm run lint
	@echo "Running OPA policy checks..."
	conftest test infrastructure/helm-charts/ -p security/policies/opa/

security-scan:
	@echo "Running security scans..."
	trivy fs --severity HIGH,CRITICAL .
	snyk test --all-projects
	checkov -d infrastructure/terraform/

# Chaos & Load Testing
chaos-test:
	@echo "Running chaos engineering suite..."
	kubectl apply -f testing/chaos/experiments/
	chaos-dashboard &

load-test:
	@echo "Running load tests..."
	cd testing/load && locust -f locustfile.py --host=https://api-staging.nexus-platform.com
	cd testing/load && k6 run k6-load-test.js

# Build
build:
	@echo "Building container images..."
	docker build -t $(REGISTRY)/backend-core:$(VERSION) services/backend-core
	docker build -t $(REGISTRY)/iptv-middleware:$(VERSION) services/iptv-middleware
	docker build -t $(REGISTRY)/ai-platform:$(VERSION) services/ai-platform
	docker build -t $(REGISTRY)/iot-gateway:$(VERSION) services/iot-gateway
	docker build -t $(REGISTRY)/pms-integration:$(VERSION) services/pms-integration

# Deployment
deploy-staging:
	@echo "Deploying to staging..."
	cd $(TERRAFORM_DIR)/staging && terraform apply -auto-approve
	cd $(HELM_DIR)/nexus-platform && helm upgrade --install nexus-staging . 		--namespace nexus-staging --values values-staging.yaml

deploy-prod:
	@echo "Deploying to production..."
	@echo "WARNING: This will deploy to PRODUCTION"
	@read -p "Are you sure? [y/N] " confirm && [ $$confirm = y ]
	cd $(TERRAFORM_DIR)/production && terraform apply
	cd $(HELM_DIR)/nexus-platform && helm upgrade --install nexus-prod . 		--namespace nexus-prod --values values-production.yaml

# Migration
migrate:
	@echo "Running database migrations..."
	cd services/pms-integration/sql && psql -f schema.sql
	cd services/pms-integration/sql && ./run-migrations.sh

provision:
	@echo "Provisioning new property..."
	@read -p "Tenant ID: " tenant; 	read -p "Property ID: " property; 	read -p "Region [us-east-1]: " region; 	read -p "Tier [enterprise]: " tier; 	./scripts/bootstrap/zero-touch-provision.sh 		$$tenant $$property $${region:-us-east-1} production $${tier:-enterprise}

# Maintenance
backup:
	@echo "Creating database backup..."
	kubectl exec -it postgres-primary-0 -- pg_dumpall | gzip > backup-$(shell date +%Y%m%d).sql.gz

cleanup:
	@echo "Cleaning up resources..."
	docker system prune -f
	kubectl delete pods --all-namespaces --field-selector=status.phase=Succeeded

# Help
help:
	@echo "Nexus Hospitality Platform — Available Commands"
	@echo ""
	@echo "Development:"
	@echo "  make dev-up          Start local development environment"
	@echo "  make dev-down        Stop local development environment"
	@echo ""
	@echo "Testing:"
	@echo "  make test            Run all tests"
	@echo "  make lint            Run linters"
	@echo "  make security-scan   Run security scanners"
	@echo "  make chaos-test      Run chaos engineering experiments"
	@echo "  make load-test       Run load tests"
	@echo ""
	@echo "Build & Deploy:"
	@echo "  make build           Build container images"
	@echo "  make deploy-staging  Deploy to staging"
	@echo "  make deploy-prod     Deploy to production"
	@echo ""
	@echo "Operations:"
	@echo "  make migrate         Run database migrations"
	@echo "  make provision       Provision new property"
	@echo "  make backup          Create database backup"
	@echo "  make cleanup         Clean up resources"
