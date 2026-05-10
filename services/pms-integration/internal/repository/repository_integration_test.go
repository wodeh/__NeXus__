//go:build integration

package repository

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*db.Pool, func()) {
	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("nexus_test"),
		postgres.WithUsername("nexus"),
		postgres.WithPassword("nexus"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	pool, err := db.NewPool(ctx, dsn)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	migrator, err := db.NewMigrator(dsn, "file://../../sql/migrations")
	if err != nil {
		t.Fatalf("failed to create migrator: %v", err)
	}
	defer migrator.Close()

	if err := migrator.Up(ctx); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	cleanup := func() {
		pool.Close()
		_ = pgContainer.Terminate(ctx)
	}

	return pool, cleanup
}

func TestReservationRepository_CRUD(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewReservationRepository(pool, NewRepositoryMetrics())

	res := domain.NewReservation("tenant-1", "prop-1", "guest-1", "room-1",
		time.Now().Add(24*time.Hour), time.Now().Add(48*time.Hour))

	// Create
	if err := repo.Create(ctx, res); err != nil {
		t.Fatalf("create reservation: %v", err)
	}

	// Read
	found, err := repo.GetByID(ctx, res.TenantID, string(res.ID))
	if err != nil {
		t.Fatalf("get reservation: %v", err)
	}
	if found.ID != res.ID {
		t.Errorf("expected id %s, got %s", res.ID, found.ID)
	}

	// Update
	res.Status = domain.ReservationStatusConfirmed
	if err := repo.Update(ctx, res); err != nil {
		t.Fatalf("update reservation: %v", err)
	}

	found, err = repo.GetByID(ctx, res.TenantID, string(res.ID))
	if err != nil {
		t.Fatalf("get updated reservation: %v", err)
	}
	if found.Status != domain.ReservationStatusConfirmed {
		t.Errorf("expected status confirmed, got %s", found.Status)
	}
	if found.Version != 2 {
		t.Errorf("expected version 2, got %d", found.Version)
	}

	// Soft delete
	if err := repo.SoftDelete(ctx, res.TenantID, string(res.ID)); err != nil {
		t.Fatalf("soft delete reservation: %v", err)
	}
	_, err = repo.GetByID(ctx, res.TenantID, string(res.ID))
	if err == nil {
		t.Fatal("expected error after soft delete, got nil")
	}
}

func TestReservationRepository_OptimisticLock(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewReservationRepository(pool, NewRepositoryMetrics())

	res := domain.NewReservation("tenant-1", "prop-1", "guest-1", "room-1",
		time.Now().Add(24*time.Hour), time.Now().Add(48*time.Hour))

	if err := repo.Create(ctx, res); err != nil {
		t.Fatalf("create reservation: %v", err)
	}

	// Simulate stale read
	stale := *res
	stale.Status = domain.ReservationStatusConfirmed

	// Update original
	res.Status = domain.ReservationStatusCancelled
	if err := repo.Update(ctx, res); err != nil {
		t.Fatalf("update reservation: %v", err)
	}

	// Stale update should fail
	if err := repo.Update(ctx, &stale); err == nil {
		t.Fatal("expected optimistic lock conflict, got nil")
	}
}

func TestReservationRepository_TenantIsolation(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewReservationRepository(pool, NewRepositoryMetrics())

	res := domain.NewReservation("tenant-a", "prop-1", "guest-1", "room-1",
		time.Now().Add(24*time.Hour), time.Now().Add(48*time.Hour))

	if err := repo.Create(ctx, res); err != nil {
		t.Fatalf("create reservation: %v", err)
	}

	_, err := repo.GetByID(ctx, "tenant-b", string(res.ID))
	if err == nil {
		t.Fatal("expected tenant isolation error, got nil")
	}
}

func TestGuestRepository_CRUD(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewGuestRepository(pool, NewRepositoryMetrics())

	g := domain.NewGuest("tenant-1", "prop-1", "John", "Doe", "john@example.com")
	g.Phone = "+1234567890"

	if err := repo.Create(ctx, g); err != nil {
		t.Fatalf("create guest: %v", err)
	}

	found, err := repo.GetByID(ctx, g.TenantID, string(g.ID))
	if err != nil {
		t.Fatalf("get guest: %v", err)
	}
	if found.Email != g.Email {
		t.Errorf("expected email %s, got %s", g.Email, found.Email)
	}

	g.FirstName = "Jane"
	if err := repo.Update(ctx, g); err != nil {
		t.Fatalf("update guest: %v", err)
	}

	found, err = repo.GetByID(ctx, g.TenantID, string(g.ID))
	if err != nil {
		t.Fatalf("get updated guest: %v", err)
	}
	if found.FirstName != "Jane" {
		t.Errorf("expected first name Jane, got %s", found.FirstName)
	}
	if found.Version != 2 {
		t.Errorf("expected version 2, got %d", found.Version)
	}
}

func TestFolioRepository_CRUD(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewFolioRepository(pool, NewRepositoryMetrics())

	f := domain.NewFolio("tenant-1", "prop-1", "res-1", "guest-1", "USD")

	if err := repo.Create(ctx, f); err != nil {
		t.Fatalf("create folio: %v", err)
	}

	found, err := repo.GetByID(ctx, f.TenantID, string(f.ID))
	if err != nil {
		t.Fatalf("get folio: %v", err)
	}
	if found.Balance != 0 {
		t.Errorf("expected balance 0, got %f", found.Balance)
	}

	charge := domain.Charge{ID: "charge-1", Description: "Room Rate", Amount: 150.0, PostedAt: time.Now()}
	if err := repo.AddCharge(ctx, f.TenantID, string(f.ID), charge); err != nil {
		t.Fatalf("add charge: %v", err)
	}

	found, err = repo.GetByID(ctx, f.TenantID, string(f.ID))
	if err != nil {
		t.Fatalf("get folio after charge: %v", err)
	}
	if len(found.Charges) != 1 {
		t.Errorf("expected 1 charge, got %d", len(found.Charges))
	}
}

func TestTransaction_Rollback(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	ctx := context.Background()
	repo := NewReservationRepository(pool, NewRepositoryMetrics())

	res := domain.NewReservation("tenant-1", "prop-1", "guest-1", "room-1",
		time.Now().Add(24*time.Hour), time.Now().Add(48*time.Hour))

	err := WithTransaction(ctx, pool, 5*time.Second, func(ctx context.Context, tx pgx.Tx) error {
		// Insert within transaction
		if _, err := tx.Exec(ctx, `
			INSERT INTO reservations (id, tenant_id, property_id, guest_id, room_id, check_in_date, check_out_date, status, created_at, updated_at, version)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`, res.ID, res.TenantID, res.PropertyID, res.GuestID, res.RoomID, res.CheckInDate, res.CheckOutDate, res.Status, res.CreatedAt, res.UpdatedAt, res.Version); err != nil {
			return err
		}
		// Force rollback
		return fmt.Errorf("intentional rollback")
	})
	if err == nil {
		t.Fatal("expected transaction error")
	}

	// Verify record does not exist
	_, err = repo.GetByID(ctx, res.TenantID, string(res.ID))
	if err == nil {
		t.Fatal("expected record not found after rollback")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
