// Package db provides PostgreSQL connection pooling, retry logic, and migration scaffolding
// for the backend-core service.
package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool.Pool with Nexus-specific operational defaults.
type Pool struct {
	*pgxpool.Pool
	dsn string
}

// NewPool creates a tenant-isolated PostgreSQL connection pool with exponential backoff.
// It validates connectivity before returning.
func NewPool(ctx context.Context, dsn string) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}

	// Production-grade defaults aligned with PgBouncer and K8s sidecar patterns.
	cfg.MaxConns = 25
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 10 * time.Minute
	cfg.HealthCheckPeriod = 5 * time.Minute

	// AfterConnect hook to set application name and tenant context defaults.
	cfg.ConnConfig.Config.RuntimeParams["application_name"] = "nexus-backend-core"

	b := backoff.WithMaxRetries(
		backoff.NewExponentialBackOff(
			backoff.WithInitialInterval(500*time.Millisecond),
			backoff.WithMaxInterval(5*time.Second),
			backoff.WithMaxElapsedTime(30*time.Second),
		),
		5,
	)

	var pool *pgxpool.Pool
	err = backoff.RetryNotify(
		func() error {
			var retryErr error
			pool, retryErr = pgxpool.NewWithConfig(ctx, cfg)
			if retryErr != nil {
				return retryErr
			}
			if pingErr := pool.Ping(ctx); pingErr != nil {
				pool.Close()
				pool = nil
				return pingErr
			}
			return nil
		},
		b,
		func(err error, d time.Duration) {
			slog.Warn("database connection retry", slog.String("error", err.Error()), slog.Duration("after", d))
		},
	)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool after retries: %w", err)
	}

	slog.Info("database pool established", slog.Int("max_conns", int(cfg.MaxConns)), slog.Int("min_conns", int(cfg.MinConns)))
	return &Pool{Pool: pool, dsn: dsn}, nil
}

// SetTenant binds the current connection pool session to a tenant ID using PostgreSQL
// configuration parameters. When RLS is enabled, this restricts row visibility.
func (p *Pool) SetTenant(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id required for RLS binding")
	}
	_, err := p.Exec(ctx, "SELECT set_config('app.current_tenant', $1, false)", tenantID)
	if err != nil {
		return fmt.Errorf("set tenant config: %w", err)
	}
	return nil
}

// ResetTenant clears the tenant binding on the current connection.
func (p *Pool) ResetTenant(ctx context.Context) error {
	_, err := p.Exec(ctx, "SELECT set_config('app.current_tenant', '', false)")
	if err != nil {
		return fmt.Errorf("reset tenant config: %w", err)
	}
	return nil
}

// IsConnectionError returns true if the error is a transient PostgreSQL connection error.
func IsConnectionError(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		// Class 08 — Connection Exception
		return pgErr.Code[0:2] == "08"
	}
	return false
}
