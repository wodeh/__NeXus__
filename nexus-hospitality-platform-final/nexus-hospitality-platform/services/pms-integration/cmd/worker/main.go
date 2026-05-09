// cmd/worker/main.go — Background job processor using Redis Streams.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/repository"
	"github.com/redis/go-redis/v9"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://nexus:nexus_secret_2024@localhost:5432/nexus_pms?sslmode=disable"
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	pool, err := db.NewPool(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("failed to parse redis url: %v", err)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()

	store := repository.NewStore(pool)

	processor := NewJobProcessor(rdb, store)

	// Start workers
	go processor.RunFolioRecalc(ctx, 30*time.Second)
	go processor.RunKeySync(ctx, 60*time.Second)
	go processor.RunHousekeepingSchedule(ctx, 5*time.Minute)
	go processor.RunNightlyReport(ctx, 24*time.Hour)

	log.Println("worker started. waiting for jobs...")
	<-ctx.Done()
	log.Println("worker shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = shutdownCtx

	log.Println("worker stopped")
}
