//go:build integration

package db

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNewPool_Integration(t *testing.T) {
	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("nexus_test"),
		postgres.WithUsername("nexus"),
		postgres.WithPassword("nexus_test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("container termination warning: %v", termErr)
		}
	}()

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := NewPool(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	if pingErr := pool.Ping(ctx); pingErr != nil {
		t.Fatalf("failed to ping: %v", pingErr)
	}

	// Verify tenant RLS binding helpers work at the SQL level.
	if err := pool.SetTenant(ctx, "tenant-123"); err != nil {
		t.Fatalf("failed to set tenant: %v", err)
	}
	if err := pool.ResetTenant(ctx); err != nil {
		t.Fatalf("failed to reset tenant: %v", err)
	}
}

func TestMigrator_Integration(t *testing.T) {
	ctx := context.Background()

	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("nexus_test"),
		postgres.WithUsername("nexus"),
		postgres.WithPassword("nexus_test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}
	defer func() {
		if termErr := container.Terminate(ctx); termErr != nil {
			t.Logf("container termination warning: %v", termErr)
		}
	}()

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := NewPool(ctx, connStr)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	// Migrations are relative to module root; in test we assume cwd is module root.
	migrator, err := NewMigrator(connStr, "file://sql/migrations")
	if err != nil {
		t.Fatalf("failed to create migrator: %v", err)
	}
	defer migrator.Close()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("migrate up failed: %v", err)
	}

	if err := migrator.Down(ctx); err != nil {
		t.Fatalf("migrate down failed: %v", err)
	}
}
