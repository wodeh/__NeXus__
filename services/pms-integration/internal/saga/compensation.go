// Package saga provides saga orchestration with compensation infrastructure.
package saga

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/nexus-platform/pms-integration/internal/domain"
	"github.com/nexus-platform/pms-integration/internal/repository"
)

// CompensationExecutor wraps compensators with retry, idempotency, timeout, and observability.
type CompensationExecutor struct {
	compensators map[string]Compensator
	idempotency  domain.IdempotencyStore
	metrics      *CompensationMetrics
	mu           sync.RWMutex
}

// CompensationMetrics tracks compensation execution.
type CompensationMetrics struct {
	total      int
	successful int
	failed     int
	mu         sync.RWMutex
}

// NewCompensationExecutor creates a new compensation executor.
func NewCompensationExecutor(idempotency domain.IdempotencyStore) *CompensationExecutor {
	return &CompensationExecutor{
		compensators: make(map[string]Compensator),
		idempotency:  idempotency,
		metrics:      &CompensationMetrics{},
	}
}

// Register adds a compensator for a service.
func (ce *CompensationExecutor) Register(service string, c Compensator) {
	ce.mu.Lock()
	defer ce.mu.Unlock()
	ce.compensators[service] = c
}

// Execute runs a compensation with bounded retries, timeout, and idempotency.
func (ce *CompensationExecutor) Execute(ctx context.Context, sagaID string, step SagaStep, sagaContext map[string]interface{}) error {
	idempotencyKey := fmt.Sprintf("compensation:%s:%s:%s", sagaID, step.Name, step.CompensateAction)

	// Idempotency check
	processed, err := ce.idempotency.IsProcessed(ctx, idempotencyKey)
	if err != nil {
		slog.Warn("idempotency check failed", slog.String("saga_id", sagaID), slog.String("error", err.Error()))
	}
	if processed {
		slog.Info("compensation already executed, skipping", slog.String("saga_id", sagaID), slog.String("step", step.Name))
		return nil
	}

	ce.mu.RLock()
	compensator, exists := ce.compensators[step.Service]
	ce.mu.RUnlock()
	if !exists {
		return fmt.Errorf("no compensator registered for service %s", step.Service)
	}

	// Timeout context for compensation
	compCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Retry with exponential backoff
	b := backoff.WithMaxRetries(
		backoff.NewExponentialBackOff(
			backoff.WithInitialInterval(500*time.Millisecond),
			backoff.WithMaxInterval(5*time.Second),
			backoff.WithMaxElapsedTime(30*time.Second),
		),
		3,
	)

	var lastErr error
	err = backoff.RetryNotify(
		func() error {
			lastErr = compensator.Compensate(compCtx, sagaID, step, sagaContext)
			return lastErr
		},
		b,
		func(err error, d time.Duration) {
			slog.Warn("compensation retry",
				slog.String("saga_id", sagaID),
				slog.String("step", step.Name),
				slog.Duration("after", d),
			)
		},
	)

	if err != nil {
		ce.metrics.RecordFailure()
		slog.Error("compensation failed permanently",
			slog.String("saga_id", sagaID),
			slog.String("step", step.Name),
			slog.String("error", lastErr.Error()),
		)
		return fmt.Errorf("compensation failed after retries: %w", lastErr)
	}

	// Mark as processed
	if markErr := ce.idempotency.MarkProcessed(ctx, idempotencyKey, 24*time.Hour); markErr != nil {
		slog.Warn("failed to mark compensation processed", slog.String("saga_id", sagaID), slog.String("error", markErr.Error()))
	}

	ce.metrics.RecordSuccess()
	slog.Info("compensation succeeded",
		slog.String("saga_id", sagaID),
		slog.String("step", step.Name),
		slog.String("service", step.Service),
	)
	return nil
}

// Metrics returns current compensation metrics.
func (ce *CompensationExecutor) Metrics() CompensationMetricsSnapshot {
	return ce.metrics.Snapshot()
}

// CompensationMetricsSnapshot is a point-in-time view of compensation metrics.
type CompensationMetricsSnapshot struct {
	Total      int
	Successful int
	Failed     int
}

func (m *CompensationMetrics) RecordSuccess() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.total++
	m.successful++
}

func (m *CompensationMetrics) RecordFailure() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.total++
	m.failed++
}

func (m *CompensationMetrics) Snapshot() CompensationMetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return CompensationMetricsSnapshot{
		Total:      m.total,
		Successful: m.successful,
		Failed:     m.failed,
	}
}

// ==================== Concrete Compensators ====================

// ReservationCompensator handles reservation-related compensations.
type ReservationCompensator struct {
	Reservations repository.ReservationRepository
}

// Compensate reverts a reservation operation.
func (c *ReservationCompensator) Compensate(ctx context.Context, sagaID string, step SagaStep, sagaContext map[string]interface{}) error {
	reservationID, _ := sagaContext["reservation_id"].(string)
	if reservationID == "" {
		return fmt.Errorf("reservation_id not found in saga context")
	}

	switch step.CompensateAction {
	case "unassign_room":
		// Reset room assignment
		res, err := c.Reservations.GetByID(ctx, sagaContext["tenant_id"].(string), reservationID)
		if err != nil {
			return fmt.Errorf("get reservation for unassign: %w", err)
		}
		res.RoomID = ""
		if err := c.Reservations.Update(ctx, res); err != nil {
			return fmt.Errorf("unassign room: %w", err)
		}
	case "cancel_reservation":
		res, err := c.Reservations.GetByID(ctx, sagaContext["tenant_id"].(string), reservationID)
		if err != nil {
			return fmt.Errorf("get reservation for cancel: %w", err)
		}
		if err := res.Cancel(); err != nil {
			return fmt.Errorf("cancel reservation: %w", err)
		}
		if err := c.Reservations.Update(ctx, res); err != nil {
			return fmt.Errorf("persist cancelled reservation: %w", err)
		}
	default:
		return fmt.Errorf("unknown reservation compensate action: %s", step.CompensateAction)
	}
	return nil
}

// BillingCompensator handles billing-related compensations.
type BillingCompensator struct {
	Folios repository.FolioRepository
}

// Compensate reverts a billing operation.
func (c *BillingCompensator) Compensate(ctx context.Context, sagaID string, step SagaStep, sagaContext map[string]interface{}) error {
	folioID, _ := sagaContext["folio_id"].(string)
	if folioID == "" {
		return fmt.Errorf("folio_id not found in saga context")
	}

	switch step.CompensateAction {
	case "void_folio":
		folio, err := c.Folios.GetByID(ctx, sagaContext["tenant_id"].(string), folioID)
		if err != nil {
			return fmt.Errorf("get folio for void: %w", err)
		}
		if err := folio.Close(); err != nil {
			return fmt.Errorf("close folio: %w", err)
		}
		if err := c.Folios.Update(ctx, folio); err != nil {
			return fmt.Errorf("persist voided folio: %w", err)
		}
	default:
		return fmt.Errorf("unknown billing compensate action: %s", step.CompensateAction)
	}
	return nil
}

// AccessControlCompensator handles access control compensations.
type AccessControlCompensator struct{}

// Compensate reverts an access control operation.
func (c *AccessControlCompensator) Compensate(ctx context.Context, sagaID string, step SagaStep, sagaContext map[string]interface{}) error {
	slog.Info("access control compensation executed",
		slog.String("saga_id", sagaID),
		slog.String("action", step.CompensateAction),
	)
	// Phase 3: access control service integration not yet implemented.
	// Compensation is logged and treated as best-effort.
	return nil
}

// HousekeepingCompensator handles housekeeping compensations.
type HousekeepingCompensator struct{}

// Compensate reverts a housekeeping operation.
func (c *HousekeepingCompensator) Compensate(ctx context.Context, sagaID string, step SagaStep, sagaContext map[string]interface{}) error {
	slog.Info("housekeeping compensation executed",
		slog.String("saga_id", sagaID),
		slog.String("action", step.CompensateAction),
	)
	// Phase 3: housekeeping service integration not yet implemented.
	// Compensation is logged and treated as best-effort.
	return nil
}

// IoTCompensator handles IoT device compensations.
type IoTCompensator struct{}

// Compensate reverts an IoT operation.
func (c *IoTCompensator) Compensate(ctx context.Context, sagaID string, step SagaStep, sagaContext map[string]interface{}) error {
	slog.Info("iot compensation executed",
		slog.String("saga_id", sagaID),
		slog.String("action", step.CompensateAction),
	)
	// Phase 3: IoT gateway integration not yet implemented.
	// Compensation is logged and treated as best-effort.
	return nil
}
