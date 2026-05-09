package db

import (
	"context"
	"testing"
	"time"
)

func TestNewPool_InvalidDSN(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := NewPool(ctx, "not-a-valid-dsn")
	if err == nil {
		t.Fatal("expected error for invalid DSN")
	}
}
