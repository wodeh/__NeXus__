package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nexus-platform/backend-core/internal/config"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/events"
	"github.com/nexus-platform/backend-core/internal/metrics"
	"github.com/nexus-platform/backend-core/internal/repository"
	"github.com/nexus-platform/backend-core/internal/server"
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

	// Initialize metrics collector first so event metrics can register on the same registry.
	metricsCollector := metrics.NewCollector(metrics.Config{Enabled: true, Port: cfg.MetricsPort})

	var repo *repository.Store
	if cfg.DatabaseURL != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		pool, err := db.NewPool(ctx, cfg.DatabaseURL)
		if err != nil {
			logger.Error("failed to connect to database", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer pool.Close()
		migrator, err := db.NewMigrator(cfg.DatabaseURL, "file://sql/migrations")
		if err != nil {
			logger.Error("failed to create migrator", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer migrator.Close()
		if err := migrator.Up(ctx); err != nil {
			logger.Error("failed to run migrations", slog.String("error", err.Error()))
			os.Exit(1)
		}
		repo = repository.NewStore(pool)
		logger.Info("database connected")
	}

	// Initialize event producer if Kafka brokers are configured.
	var producer *events.ProducerManager
	if len(cfg.KafkaBrokers) > 0 {
		var err error
		producer, err = events.NewProducerManager(events.ProducerConfig{
			Brokers: cfg.KafkaBrokers,
			Topic:   cfg.KafkaTopic,
			Source:  "backend-core",
			Metrics: events.NewEventMetrics(metricsCollector.Registry()),
		})
		if err != nil {
			logger.Error("failed to create event producer", slog.String("error", err.Error()))
			os.Exit(1)
		}
		defer producer.Close()
		logger.Info("event producer initialized", slog.Any("brokers", cfg.KafkaBrokers))
	}

	srv := server.New(cfg, repo, producer, metricsCollector)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := srv.Start(ctx); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown error", slog.String("error", err.Error()))
	}
	logger.Info("backend-core shutdown complete")
}
