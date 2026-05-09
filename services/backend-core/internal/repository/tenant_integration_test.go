//go:build integration

package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*db.Pool, *db.Migrator, func()) {
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

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := db.NewPool(ctx, connStr)
	if err != nil {
		_ = container.Terminate(ctx)
		t.Fatalf("failed to create pool: %v", err)
	}

	migrator, err := db.NewMigrator(connStr, "file://sql/migrations")
	if err != nil {
		pool.Close()
		_ = container.Terminate(ctx)
		t.Fatalf("failed to create migrator: %v", err)
	}

	if err := migrator.Up(ctx); err != nil {
		migrator.Close()
		pool.Close()
		_ = container.Terminate(ctx)
		t.Fatalf("migrate up failed: %v", err)
	}

	cleanup := func() {
		migrator.Close()
		pool.Close()
		_ = container.Terminate(ctx)
	}

	return pool, migrator, cleanup
}

func TestTenantRepository_CreateAndGet(t *testing.T) {
	pool, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewTenantRepository(pool)

	tenant := &Tenant{
		ExternalID: "hotel-group-abc",
		Name:       "ABC Hospitality",
		Region:     "us-west-2",
		Tier:       "enterprise",
	}

	if err := repo.Create(ctx, nil, tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	if tenant.ID == uuid.Nil {
		t.Fatal("expected tenant ID to be generated")
	}

	fetched, err := repo.GetByExternalID(ctx, "hotel-group-abc")
	if err != nil {
		t.Fatalf("get tenant: %v", err)
	}
	if fetched.Name != "ABC Hospitality" {
		t.Fatalf("expected name ABC Hospitality, got %s", fetched.Name)
	}
	if fetched.Region != "us-west-2" {
		t.Fatalf("expected region us-west-2, got %s", fetched.Region)
	}
}

func TestTenantRepository_SoftDeleteOptimisticLock(t *testing.T) {
	pool, _, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewTenantRepository(pool)

	tenant := &Tenant{
		ExternalID: "hotel-group-del",
		Name:       "Deletable Group",
	}
	if err := repo.Create(ctx, nil, tenant); err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	// First soft-delete should succeed
	if err := repo.SoftDelete(ctx, nil, tenant.ID, tenant.Version); err != nil {
		t.Fatalf("first soft delete: %v", err)
	}

	// Second attempt with same version should fail (optimistic lock)
	if err := repo.SoftDelete(ctx, nil, tenant.ID, tenant.Version); err == nil {
		t.Fatal("expected optimistic lock error on second delete")
	}

	// Should not be findable after soft delete
	_, err := repo.GetByExternalID(ctx, "hotel-group-del")
	if err == nil {
		t.Fatal("expected not found after soft delete")
	}
}
