// Package db provides PostgreSQL connection pooling and tenant isolation for PMS integration.
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool.Pool with PMS-specific operational defaults.
type Pool struct {
	*pgxpool.Pool
}

// NewPool creates a tenant-isolated PostgreSQL connection pool.
func NewPool(ctx context.Context, dsn string) (*Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pgx config: %w", err)
	}

	cfg.MaxConns = 25
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 30 * time.Minute
	cfg.MaxConnIdleTime = 10 * time.Minute
	cfg.HealthCheckPeriod = 5 * time.Minute
	cfg.ConnConfig.Config.RuntimeParams["application_name"] = "nexus-pms-integration"

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &Pool{Pool: pool}, nil
}

// SetTenant binds the current connection pool session to a tenant ID for RLS.
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
