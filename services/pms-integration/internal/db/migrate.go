package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

)

// Migration represents a single database migration.
type Migration struct {
	Version string
	Name    string
	Up      string
	Down    string
}

// Migrator handles database migrations.
type Migrator struct {
	pool *Pool
}

// NewMigrator creates a new migrator.
func NewMigrator(pool *Pool) *Migrator {
	return &Migrator{pool: pool}
}

// EnsureMigrationsTable creates the migrations tracking table.
func (m *Migrator) EnsureMigrationsTable(ctx context.Context) error {
	_, err := m.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		)
	`)
	return err
}

// LoadMigrationsFromDir reads migration files from a directory.
func LoadMigrationsFromDir(dir string) ([]Migration, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var migrations []Migration
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".up.sql") {
			continue
		}
		base := strings.TrimSuffix(entry.Name(), ".up.sql")
		parts := strings.SplitN(base, "_", 2)
		if len(parts) != 2 {
			continue
		}
		version := parts[0]
		name := parts[1]

		upBytes, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}

		downPath := filepath.Join(dir, base+".down.sql")
		downBytes := []byte{}
		if _, err := os.Stat(downPath); err == nil {
			downBytes, err = os.ReadFile(downPath)
			if err != nil {
				return nil, err
			}
		}

		migrations = append(migrations, Migration{
			Version: version,
			Name:    name,
			Up:      string(upBytes),
			Down:    string(downBytes),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

// GetAppliedVersions returns already-applied migration versions.
func (m *Migrator) GetAppliedVersions(ctx context.Context) (map[string]bool, error) {
	rows, err := m.pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// Up runs all pending migrations.
func (m *Migrator) Up(ctx context.Context, migrations []Migration) error {
	if err := m.EnsureMigrationsTable(ctx); err != nil {
		return fmt.Errorf("ensure migrations table: %w", err)
	}

	applied, err := m.GetAppliedVersions(ctx)
	if err != nil {
		return fmt.Errorf("get applied versions: %w", err)
	}

	for _, mig := range migrations {
		if applied[mig.Version] {
			continue
		}

		if err := m.runMigration(ctx, mig); err != nil {
			return fmt.Errorf("migration %s (%s): %w", mig.Version, mig.Name, err)
		}
	}

	return nil
}

func (m *Migrator) runMigration(ctx context.Context, mig Migration) error {
	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, mig.Up); err != nil {
		return fmt.Errorf("up sql: %w", err)
	}

	if _, err := tx.Exec(ctx,
		"INSERT INTO schema_migrations (version) VALUES ($1)",
		mig.Version,
	); err != nil {
		return fmt.Errorf("record migration: %w", err)
	}

	return tx.Commit(ctx)
}

// Down rolls back the last migration.
func (m *Migrator) Down(ctx context.Context, migrations []Migration) error {
	applied, err := m.GetAppliedVersions(ctx)
	if err != nil {
		return err
	}

	if len(applied) == 0 {
		return nil
	}

	// Find last applied
	var lastMigration *Migration
	for i := len(migrations) - 1; i >= 0; i-- {
		if applied[migrations[i].Version] {
			lastMigration = &migrations[i]
			break
		}
	}

	if lastMigration == nil || lastMigration.Down == "" {
		return nil
	}

	tx, err := m.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, lastMigration.Down); err != nil {
		return fmt.Errorf("down sql: %w", err)
	}

	if _, err := tx.Exec(ctx,
		"DELETE FROM schema_migrations WHERE version = $1",
		lastMigration.Version,
	); err != nil {
		return fmt.Errorf("remove migration record: %w", err)
	}

	return tx.Commit(ctx)
}
