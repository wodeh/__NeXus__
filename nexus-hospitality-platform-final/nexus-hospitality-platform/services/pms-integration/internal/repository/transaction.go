// Package repository provides tenant-aware data access for PMS integration.
package repository

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
)

// TxFn is a function that executes within a database transaction.
type TxFn func(ctx context.Context, tx pgx.Tx) error

// WithTransaction executes fn inside a pgx transaction with timeout and rollback safety.
func WithTransaction(ctx context.Context, pool *db.Pool, timeout time.Duration, fn TxFn) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:       pgx.ReadCommitted,
		AccessMode:     pgx.ReadWrite,
		DeferrableMode: pgx.NotDeferrable,
	})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
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
