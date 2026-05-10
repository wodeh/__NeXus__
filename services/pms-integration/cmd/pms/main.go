package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nexus-platform/pms-integration/internal/api"
	"github.com/nexus-platform/pms-integration/internal/config"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
	"github.com/nexus-platform/pms-integration/internal/events"
	"github.com/nexus-platform/pms-integration/internal/license"
	"github.com/nexus-platform/pms-integration/internal/metrics"
	"github.com/nexus-platform/pms-integration/internal/repository"
	"github.com/nexus-platform/pms-integration/internal/saga"
	"github.com/nexus-platform/pms-integration/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}
	if err := cfg.Validate(); err != nil {
		logger.Error("invalid config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Initialize metrics collector.
	metricsCollector := metrics.NewCollector(metrics.Config{Enabled: true, Port: cfg.MetricsPort})
	defer metricsCollector.Shutdown(context.Background())

	// Initialize database pool and repository store.
	var repoStore *repository.Store
	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pool, err := db.NewPool(ctx, cfg.DatabaseURL)
		if err != nil {
			logger.Error("failed to connect to database", slog.String("error", err.Error()))
			os.Exit(1)
		}

		migrator, err := db.NewMigrator(cfg.DatabaseURL, "file://sql/migrations")
		if err != nil {
			logger.Error("failed to create migrator", slog.String("error", err.Error()))
			os.Exit(1)
		}
		if err := migrator.Up(ctx); err != nil {
			logger.Error("failed to run migrations", slog.String("error", err.Error()))
			os.Exit(1)
		}
		migrator.Close()

		repoStore = repository.NewStore(pool)
		logger.Info("database connected")
	}

	// Initialize event store.
	var eventStore events.Store
	if len(cfg.KafkaBrokers) > 0 {
		mapper := events.NewDomainEventMapper("pms-integration")
		kafkaStore, err := events.NewKafkaEventStore(cfg.KafkaBrokers, cfg.KafkaTopic, mapper)
		if err != nil {
			logger.Error("failed to create kafka event store", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer kafkaStore.Close()
		eventStore = kafkaStore
		logger.Info("kafka event store initialized", slog.Any("brokers", cfg.KafkaBrokers))
	} else {
		eventStore = &inMemoryEventStore{}
	}

	// Initialize saga orchestrator.
	commandBus := &inMemoryCommandBus{}
	sagaStore := store.NewInMemorySagaStore()
	orchestrator := saga.NewSagaOrchestrator(eventStore, commandBus, sagaStore)

	// Wire compensation executor with real compensators when repositories are available.
	if repoStore != nil {
		compExecutor := saga.NewCompensationExecutor(domain.NewInMemoryIdempotencyStore(24 * time.Hour))
		compExecutor.Register("reservation", &saga.ReservationCompensator{Reservations: repoStore.Reservations})
		compExecutor.Register("billing", &saga.BillingCompensator{Folios: repoStore.Folios})
		compExecutor.Register("access_control", &saga.AccessControlCompensator{})
		compExecutor.Register("housekeeping", &saga.HousekeepingCompensator{})
		compExecutor.Register("iot_gateway", &saga.IoTCompensator{})
		orchestrator.SetCompensationExecutor(compExecutor)
	}

	// Initialize license service.
	licenseSvc := license.NewService()

	// HTTP router.
	handler := api.NewHandler(repoStore, licenseSvc)
	router := handler.Router()

	server := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("pms integration starting", slog.String("port", cfg.HTTPPort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", slog.String("error", err.Error()))
	}
	if repoStore != nil {
		repoStore.Close()
	}
	logger.Info("pms integration shutdown complete")
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

type inMemoryEventStore struct{}

func (s *inMemoryEventStore) Publish(ctx context.Context, event events.DomainEvent) error {
	slog.Info("event published", slog.String("type", event.Type), slog.String("tenant", event.TenantID))
	return nil
}

type inMemoryCommandBus struct{}

func (b *inMemoryCommandBus) Send(ctx context.Context, command events.Command) (*events.CommandResponse, error) {
	slog.Info("command sent", slog.String("type", command.Type), slog.String("service", command.Service))
	return &events.CommandResponse{Success: true, Data: map[string]interface{}{}}, nil
}
