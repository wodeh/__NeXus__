package saga

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/pms-integration/internal/events"
)

type SagaOrchestrator struct {
	eventStore    events.Store
	commandBus    events.CommandBus
	sagaStore     SagaStore
	activeSagas   map[string]*SagaInstance
	mu            sync.RWMutex
	compensators  map[string]Compensator
	compExecutor  *CompensationExecutor
}

type SagaInstance struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Status        SagaStatus             `json:"status"`
	Steps         []SagaStep             `json:"steps"`
	CurrentStep   int                    `json:"current_step"`
	Context       map[string]interface{} `json:"context"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	TenantID      string                 `json:"tenant_id"`
	CorrelationID string                 `json:"correlation_id"`
}

type SagaStatus string

const (
	SagaStatusPending      SagaStatus = "pending"
	SagaStatusRunning      SagaStatus = "running"
	SagaStatusCompleted    SagaStatus = "completed"
	SagaStatusCompensating SagaStatus = "compensating"
	SagaStatusFailed       SagaStatus = "failed"
)

type SagaStep struct {
	Name             string                 `json:"name"`
	Service          string                 `json:"service"`
	Action           string                 `json:"action"`
	CompensateAction string                 `json:"compensate_action"`
	Status           StepStatus             `json:"status"`
	Input            map[string]interface{} `json:"input"`
	Output           map[string]interface{} `json:"output"`
	Error            string                 `json:"error,omitempty"`
	StartedAt        *time.Time             `json:"started_at,omitempty"`
	CompletedAt      *time.Time             `json:"completed_at,omitempty"`
	RetryCount       int                    `json:"retry_count"`
	MaxRetries       int                    `json:"max_retries"`
}

type StepStatus string

const (
	StepStatusPending     StepStatus = "pending"
	StepStatusRunning     StepStatus = "running"
	StepStatusCompleted   StepStatus = "completed"
	StepStatusFailed      StepStatus = "failed"
	StepStatusCompensated StepStatus = "compensated"
)

type Compensator interface {
	Compensate(ctx context.Context, sagaID string, step SagaStep, context map[string]interface{}) error
}

type SagaDefinition struct {
	Name  string
	Steps []SagaStepDefinition
}

type SagaStepDefinition struct {
	Name             string
	Service          string
	Action           string
	CompensateAction string
	MaxRetries       int
	InputMapper      func(ctx map[string]interface{}) map[string]interface{}
}

var GuestCheckInSaga = SagaDefinition{
	Name: "guest_checkin",
	Steps: []SagaStepDefinition{
		{Name: "validate_reservation", Service: "reservation", Action: "validate_reservation", MaxRetries: 3},
		{Name: "assign_room", Service: "reservation", Action: "assign_room", CompensateAction: "unassign_room", MaxRetries: 3},
		{Name: "create_folio", Service: "billing", Action: "create_folio", CompensateAction: "void_folio", MaxRetries: 3},
		{Name: "activate_room_key", Service: "access_control", Action: "activate_key_card", CompensateAction: "deactivate_key_card", MaxRetries: 3},
		{Name: "update_housekeeping", Service: "housekeeping", Action: "mark_room_occupied", CompensateAction: "mark_room_vacant", MaxRetries: 3},
		{Name: "send_welcome_notification", Service: "notification", Action: "send_guest_welcome", MaxRetries: 2},
		{Name: "initialize_iot_room", Service: "iot_gateway", Action: "initialize_room_state", MaxRetries: 3},
	},
}

var RoomChangeSaga = SagaDefinition{
	Name: "room_change",
	Steps: []SagaStepDefinition{
		{Name: "validate_new_room", Service: "reservation", Action: "validate_room_availability", MaxRetries: 3},
		{Name: "transfer_folio", Service: "billing", Action: "transfer_folio_items", CompensateAction: "reverse_transfer", MaxRetries: 3},
		{Name: "deactivate_old_key", Service: "access_control", Action: "deactivate_key_card", CompensateAction: "reactivate_key_card", MaxRetries: 3},
		{Name: "activate_new_key", Service: "access_control", Action: "activate_new_key_card", CompensateAction: "deactivate_new_key_card", MaxRetries: 3},
		{Name: "update_housekeeping_old", Service: "housekeeping", Action: "schedule_room_cleaning", MaxRetries: 3},
		{Name: "update_housekeeping_new", Service: "housekeeping", Action: "mark_room_occupied", CompensateAction: "mark_room_vacant", MaxRetries: 3},
		{Name: "update_iot_room", Service: "iot_gateway", Action: "transfer_room_state", MaxRetries: 3},
	},
}

// SagaStore persists saga instances.
type SagaStore interface {
	Save(ctx context.Context, instance *SagaInstance) error
}

// NewSagaOrchestrator creates a new saga orchestrator.
func NewSagaOrchestrator(eventStore events.Store, commandBus events.CommandBus, sagaStore SagaStore) *SagaOrchestrator {
	so := &SagaOrchestrator{
		eventStore:   eventStore,
		commandBus:   commandBus,
		sagaStore:    sagaStore,
		activeSagas:  make(map[string]*SagaInstance),
		compensators: make(map[string]Compensator),
	}

	// Register stub compensators for backward compatibility.
	so.compensators["reservation"] = &stubCompensator{service: "reservation"}
	so.compensators["billing"] = &stubCompensator{service: "billing"}
	so.compensators["access_control"] = &stubCompensator{service: "access_control"}
	so.compensators["housekeeping"] = &stubCompensator{service: "housekeeping"}
	so.compensators["iot_gateway"] = &stubCompensator{service: "iot_gateway"}

	return so
}

// SetCompensationExecutor wires a real compensation executor into the orchestrator.
func (so *SagaOrchestrator) SetCompensationExecutor(executor *CompensationExecutor) {
	so.mu.Lock()
	defer so.mu.Unlock()
	so.compExecutor = executor
}

func (so *SagaOrchestrator) StartSaga(ctx context.Context, definition SagaDefinition, initialContext map[string]interface{}) (*SagaInstance, error) {
	sagaID := uuid.New().String()
	correlationID := ctx.Value("correlation_id")
	if correlationID == nil {
		correlationID = sagaID
	}

	steps := make([]SagaStep, len(definition.Steps))
	for i, def := range definition.Steps {
		steps[i] = SagaStep{
			Name:             def.Name,
			Service:          def.Service,
			Action:           def.Action,
			CompensateAction: def.CompensateAction,
			Status:           StepStatusPending,
			MaxRetries:       def.MaxRetries,
		}
		if def.InputMapper != nil {
			steps[i].Input = def.InputMapper(initialContext)
		}
	}

	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok || tenantID == "" {
		return nil, fmt.Errorf("missing or invalid tenant_id in context")
	}
	correlationIDStr, ok := correlationID.(string)
	if !ok {
		correlationIDStr = sagaID
	}

	instance := &SagaInstance{
		ID:            sagaID,
		Type:          definition.Name,
		Status:        SagaStatusRunning,
		Steps:         steps,
		CurrentStep:   0,
		Context:       initialContext,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		TenantID:      tenantID,
		CorrelationID: correlationIDStr,
	}

	so.mu.Lock()
	so.activeSagas[sagaID] = instance
	so.mu.Unlock()

	if err := so.sagaStore.Save(ctx, instance); err != nil {
		return nil, fmt.Errorf("failed to persist saga: %w", err)
	}

	so.publishSagaEvent(ctx, instance, "saga_started")
	go so.executeStep(instance, 0)

	return instance, nil
}

func (so *SagaOrchestrator) executeStep(instance *SagaInstance, stepIndex int) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "correlation_id", instance.CorrelationID)
	ctx = context.WithValue(ctx, "tenant_id", instance.TenantID)
	ctx = context.WithValue(ctx, "saga_id", instance.ID)

	if stepIndex >= len(instance.Steps) {
		so.completeSaga(instance)
		return
	}

	step := &instance.Steps[stepIndex]
	now := time.Now()
	step.StartedAt = &now
	step.Status = StepStatusRunning
	instance.CurrentStep = stepIndex
	instance.UpdatedAt = now

	so.sagaStore.Save(ctx, instance)

	command := events.Command{
		Type:          step.Action,
		Service:       step.Service,
		SagaID:        instance.ID,
		StepIndex:     stepIndex,
		CorrelationID: instance.CorrelationID,
		TenantID:      instance.TenantID,
		Payload:       so.buildStepPayload(instance, step),
	}

	response, err := so.commandBus.Send(ctx, command)
	if err != nil {
		so.handleStepFailure(instance, stepIndex, err)
		return
	}

	if response.Success {
		step.Status = StepStatusCompleted
		step.Output = response.Data
		completedAt := time.Now()
		step.CompletedAt = &completedAt

		for k, v := range response.Data {
			instance.Context[k] = v
		}

		so.sagaStore.Save(ctx, instance)
		go so.executeStep(instance, stepIndex+1)
	} else {
		so.handleStepFailure(instance, stepIndex, fmt.Errorf(response.Error))
	}
}

func (so *SagaOrchestrator) handleStepFailure(instance *SagaInstance, stepIndex int, err error) {
	step := &instance.Steps[stepIndex]
	step.Status = StepStatusFailed
	step.Error = err.Error()

	if step.RetryCount < step.MaxRetries {
		step.RetryCount++
		log.Printf("Retrying saga %s step %s (attempt %d/%d)", instance.ID, step.Name, step.RetryCount, step.MaxRetries)
		time.Sleep(time.Duration(step.RetryCount) * time.Second)
		go so.executeStep(instance, stepIndex)
		return
	}

	log.Printf("Saga %s step %s failed permanently, starting compensation", instance.ID, step.Name)
	so.compensateSaga(instance, stepIndex)
}

func (so *SagaOrchestrator) compensateSaga(instance *SagaInstance, failedStepIndex int) {
	instance.Status = SagaStatusCompensating
	instance.UpdatedAt = time.Now()
	so.sagaStore.Save(context.Background(), instance)

	for i := failedStepIndex; i >= 0; i-- {
		step := instance.Steps[i]
		if step.CompensateAction == "" {
			continue
		}

		ctx := context.Background()
		ctx = context.WithValue(ctx, "correlation_id", instance.CorrelationID)
		ctx = context.WithValue(ctx, "tenant_id", instance.TenantID)

		// Use compensation executor if available, otherwise fall back to direct compensator.
		if so.compExecutor != nil {
			if err := so.compExecutor.Execute(ctx, instance.ID, step, instance.Context); err != nil {
				log.Printf("Compensation failed for saga %s step %s: %v", instance.ID, step.Name, err)
				so.alertCompensationFailure(instance, step, err)
			}
		} else {
			compensator, exists := so.compensators[step.Service]
			if !exists {
				log.Printf("No compensator found for service %s", step.Service)
				continue
			}
			if err := compensator.Compensate(ctx, instance.ID, step, instance.Context); err != nil {
				log.Printf("Compensation failed for saga %s step %s: %v", instance.ID, step.Name, err)
				so.alertCompensationFailure(instance, step, err)
			}
		}

		step.Status = StepStatusCompensated
		instance.UpdatedAt = time.Now()
		so.sagaStore.Save(ctx, instance)
	}

	instance.Status = SagaStatusFailed
	failedAt := time.Now()
	instance.CompletedAt = &failedAt
	so.sagaStore.Save(context.Background(), instance)
	so.publishSagaEvent(context.Background(), instance, "saga_failed")
}

func (so *SagaOrchestrator) completeSaga(instance *SagaInstance) {
	instance.Status = SagaStatusCompleted
	completedAt := time.Now()
	instance.CompletedAt = &completedAt
	instance.UpdatedAt = completedAt

	so.mu.Lock()
	delete(so.activeSagas, instance.ID)
	so.mu.Unlock()

	so.sagaStore.Save(context.Background(), instance)
	so.publishSagaEvent(context.Background(), instance, "saga_completed")

	log.Printf("Saga %s completed successfully", instance.ID)
}

func (so *SagaOrchestrator) buildStepPayload(instance *SagaInstance, step *SagaStep) map[string]interface{} {
	payload := make(map[string]interface{})
	for k, v := range instance.Context {
		payload[k] = v
	}
	for k, v := range step.Input {
		payload[k] = v
	}
	payload["_saga_id"] = instance.ID
	payload["_step_index"] = instance.CurrentStep
	payload["_step_name"] = step.Name
	return payload
}

func (so *SagaOrchestrator) publishSagaEvent(ctx context.Context, instance *SagaInstance, eventType string) {
	event := events.DomainEvent{
		Type:          eventType,
		AggregateID:   instance.ID,
		AggregateType: "saga",
		TenantID:      instance.TenantID,
		CorrelationID: instance.CorrelationID,
		OccurredAt:    time.Now(),
		Payload: map[string]interface{}{
			"saga_type":    instance.Type,
			"saga_status":  string(instance.Status),
			"current_step": instance.CurrentStep,
			"total_steps":  len(instance.Steps),
		},
	}
	so.eventStore.Publish(ctx, event)
}

func (so *SagaOrchestrator) alertCompensationFailure(instance *SagaInstance, step SagaStep, err error) {
	log.Printf("ALERT: Compensation failure for saga %s, step %s: %v", instance.ID, step.Name, err)
}

// stubCompensator provides a backward-compatible compensator that logs but does not fail.
type stubCompensator struct {
	service string
}

func (c *stubCompensator) Compensate(ctx context.Context, sagaID string, step SagaStep, context map[string]interface{}) error {
	log.Printf("stub compensator for service %s executed: action=%s saga=%s", c.service, step.CompensateAction, sagaID)
	return nil
}
