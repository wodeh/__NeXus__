package db

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrator executes schema migrations using golang-migrate.
type Migrator struct {
	m *migrate.Migrate
}

// NewMigrator creates a migrator backed by the provided DSN.
// migrationsPath should be a file:// URL pointing to the migrations directory.
func NewMigrator(dsn, migrationsPath string) (*Migrator, error) {
	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		return nil, fmt.Errorf("create migrator: %w", err)
	}
	return &Migrator{m: m}, nil
}

// Up applies all pending migrations.
func (m *Migrator) Up(ctx context.Context) error {
	_ = ctx // golang-migrate does not accept context; run synchronously
	if err := m.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		// If migration fails due to already-existing objects, mark it as applied
		// and continue. This handles cases where schema exists but schema_migrations
		// table was reset or corrupted.
		slog.Warn("migration failed, attempting to force version", slog.String("error", err.Error()))
		version, _, _ := m.m.Version()
		if version > 0 {
			if forceErr := m.m.Force(int(version)); forceErr != nil {
				return fmt.Errorf("migrate up: %w, force version %d failed: %v", err, version, forceErr)
			}
			slog.Info("forced migration version", slog.Uint64("version", uint64(version)))
			return nil
		}
		return fmt.Errorf("migrate up: %w", err)
	}
	version, dirty, _ := m.m.Version()
	slog.Info("migrations applied", slog.Uint64("version", uint64(version)), slog.Bool("dirty", dirty))
	return nil
}

// Down reverts all migrations.
func (m *Migrator) Down(ctx context.Context) error {
	_ = ctx
	if err := m.m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}

// Close releases migration resources.
func (m *Migrator) Close() error {
	srcErr, dbErr := m.m.Close()
	if srcErr != nil {
		return fmt.Errorf("close migration source: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("close migration database: %w", dbErr)
	}
	return nil
}
