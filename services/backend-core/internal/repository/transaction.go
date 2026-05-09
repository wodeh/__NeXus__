// Package repository provides tenant-aware data access for backend-core.
package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nexus-platform/backend-core/internal/db"
)

// TxFn is a function that executes within a database transaction.
type TxFn func(ctx context.Context, tx pgx.Tx) error

// WithTransaction executes fn inside a pgx transaction with timeout and rollback safety.
// If fn returns an error, the transaction is rolled back. Otherwise it is committed.
func WithTransaction(ctx context.Context, pool *db.Pool, timeout time.Duration, fn TxFn) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
		AccessMode: pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	// Ensure rollback on panic or error
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // re-panic after cleanup
		}
	}()

	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			slog.Error("transaction rollback failed", slog.String("error", rbErr.Error()))
		}
		return fmt.Errorf("tx fn: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}
	return nil
}

// ExecTimeout executes a query with a bounded timeout.
func ExecTimeout(ctx context.Context, pool *db.Pool, timeout time.Duration, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return pool.Exec(ctx, sql, args...)
}

// QueryRowTimeout executes a single-row query with a bounded timeout.
func QueryRowTimeout(ctx context.Context, pool *db.Pool, timeout time.Duration, sql string, args ...interface{}) pgx.Row {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return pool.QueryRow(ctx, sql, args...)
}
