// Package main provides background job processing for the PMS.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nexus-platform/pms-integration/internal/repository"
	"github.com/redis/go-redis/v9"
)

type JobType string

const (
	JobFolioRecalc        JobType = "folio_recalc"
	JobKeySync            JobType = "key_sync"
	JobHousekeepingSchedule JobType = "housekeeping_schedule"
	JobNightlyReport      JobType = "nightly_report"
)

type Job struct {
	Type      JobType                `json:"type"`
	TenantID  string                 `json:"tenant_id"`
	Payload   map[string]interface{} `json:"payload"`
	CreatedAt time.Time              `json:"created_at"`
}

type JobProcessor struct {
	rdb   *redis.Client
	store *repository.Store
}

func NewJobProcessor(rdb *redis.Client, store *repository.Store) *JobProcessor {
	return &JobProcessor{rdb: rdb, store: store}
}

func (p *JobProcessor) publishJob(ctx context.Context, job Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return p.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "nexus:jobs",
		Values: map[string]interface{}{
			"data": string(data),
		},
	}).Err()
}

func (p *JobProcessor) consumeJobs(ctx context.Context, jobType JobType, handler func(context.Context, Job) error) {
	group := fmt.Sprintf("workers:%s", jobType)
	consumer := fmt.Sprintf("consumer-%d", time.Now().UnixNano())

	_ = p.rdb.XGroupCreateMkStream(ctx, "nexus:jobs", group, "0").Err()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		streams, err := p.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{"nexus:jobs", ">"},
			Count:    10,
			Block:    5 * time.Second,
		}).Result()

		if err != nil {
			if err != redis.Nil {
				log.Printf("error reading stream: %v", err)
			}
			continue
		}

		for _, stream := range streams {
			for _, msg := range stream.Messages {
				var job Job
				if data, ok := msg.Values["data"].(string); ok {
					if err := json.Unmarshal([]byte(data), &job); err != nil {
						log.Printf("failed to unmarshal job: %v", err)
						_ = p.rdb.XAck(ctx, "nexus:jobs", group, msg.ID).Err()
						continue
					}
				}

				if job.Type == jobType {
					if err := handler(ctx, job); err != nil {
						log.Printf("job %s failed: %v", jobType, err)
						// Retry logic could go here
					}
				}

				_ = p.rdb.XAck(ctx, "nexus:jobs", group, msg.ID).Err()
			}
		}
	}
}

// RunFolioRecalc periodically recalculates folio balances.
func (p *JobProcessor) RunFolioRecalc(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Publish recalc jobs for all open folios
			rows, err := p.store.pool.Query(ctx, `
				SELECT DISTINCT tenant_id, id FROM folios WHERE status = 'open'
			`)
			if err != nil {
				log.Printf("folio recalc query failed: %v", err)
				continue
			}
			for rows.Next() {
				var tenantID, folioID string
				_ = rows.Scan(&tenantID, &folioID)
				_ = p.publishJob(ctx, Job{
					Type:      JobFolioRecalc,
					TenantID:  tenantID,
					Payload:   map[string]interface{}{"folio_id": folioID},
					CreatedAt: time.Now().UTC(),
				})
			}
			rows.Close()
		}
	}
}

// RunKeySync syncs digital keys with Orbita API.
func (p *JobProcessor) RunKeySync(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Find keys pending sync
			rows, err := p.store.pool.Query(ctx, `
				SELECT tenant_id, id, smart_lock_id, guest_phone
				FROM digital_keys
				WHERE status = 'active' AND synced_at IS NULL
			`)
			if err != nil {
				log.Printf("key sync query failed: %v", err)
				continue
			}
			for rows.Next() {
				var tenantID, keyID, lockID, phone string
				_ = rows.Scan(&tenantID, &keyID, &lockID, &phone)
				_ = p.publishJob(ctx, Job{
					Type:      JobKeySync,
					TenantID:  tenantID,
					Payload:   map[string]interface{}{"key_id": keyID, "lock_id": lockID, "phone": phone},
					CreatedAt: time.Now().UTC(),
				})
			}
			rows.Close()
		}
	}
}

// RunHousekeepingSchedule auto-creates housekeeping tasks for check-outs.
func (p *JobProcessor) RunHousekeepingSchedule(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := time.Now().UTC()
			rows, err := p.store.pool.Query(ctx, `
				SELECT r.tenant_id, r.room_id, r.check_out_date
				FROM reservations r
				WHERE r.status = 'checked_in'
				  AND r.check_out_date <= $1
			`, now.Format("2006-01-02"))
			if err != nil {
				log.Printf("hk schedule query failed: %v", err)
				continue
			}
			for rows.Next() {
				var tenantID, roomID, checkoutDate string
				_ = rows.Scan(&tenantID, &roomID, &checkoutDate)
				_ = p.publishJob(ctx, Job{
					Type:      JobHousekeepingSchedule,
					TenantID:  tenantID,
					Payload:   map[string]interface{}{"room_id": roomID, "checkout_date": checkoutDate},
					CreatedAt: time.Now().UTC(),
				})
			}
			rows.Close()
		}
	}
}

// RunNightlyReport generates daily revenue and occupancy reports.
func (p *JobProcessor) RunNightlyReport(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = p.publishJob(ctx, Job{
				Type:      JobNightlyReport,
				TenantID:  "*",
				Payload:   map[string]interface{}{"date": time.Now().UTC().Format("2006-01-02")},
				CreatedAt: time.Now().UTC(),
			})
		}
	}
}
