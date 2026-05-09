// Package db provides migration scaffolding for PMS integration.
package db

import (
	"context"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Migrator wraps golang-migrate for PMS schema management.
type Migrator struct {
	m *migrate.Migrate
}

// NewMigrator creates a migrator from a database DSN and migrations path.
func NewMigrator(databaseURL, migrationsPath string) (*Migrator, error) {
	m, err := migrate.New(
		"file://"+migrationsPath,
		databaseURL,
	)
	if err != nil {
		return nil, fmt.Errorf("create migrator: %w", err)
	}
	return &Migrator{m: m}, nil
}

// Up runs all pending migrations.
func (m *Migrator) Up(ctx context.Context) error {
	if err := m.m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate up: %w", err)
	}
	return nil
}

// Down rolls back one migration.
func (m *Migrator) Down(ctx context.Context) error {
	if err := m.m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate down: %w", err)
	}
	return nil
}

// Close cleans up migrator resources.
func (m *Migrator) Close() error {
	srcErr, dbErr := m.m.Close()
	if srcErr != nil {
		return fmt.Errorf("close source: %w", srcErr)
	}
	if dbErr != nil {
		return fmt.Errorf("close db: %w", dbErr)
	}
	return nil
}
