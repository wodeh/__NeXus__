// Package provider implements circuit breaker pattern for inference backends.
package provider

import (
	"fmt"
	"sync"
	"time"
)

// CircuitBreaker prevents cascading failures by rejecting requests after consecutive errors.
type CircuitBreaker struct {
	name           string
	failureThreshold int
	resetTimeout   time.Duration
	state          CircuitState
	failures       int
	lastFailure    time.Time
	mu             sync.RWMutex
}

// CircuitState represents the breaker state.
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// NewCircuitBreaker creates a circuit breaker.
func NewCircuitBreaker(name string, failureThreshold int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		name:             name,
		failureThreshold: failureThreshold,
		resetTimeout:     resetTimeout,
		state:            StateClosed,
	}
}

// Allow returns nil if the breaker is closed, otherwise returns an error.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.failures = 0
			return nil
		}
		return fmt.Errorf("circuit breaker %s is open", cb.name)
	}
	return nil
}

// RecordSuccess marks a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures = 0
	cb.state = StateClosed
}

// RecordFailure marks a failed request.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.failures++
	cb.lastFailure = time.Now()
	if cb.failures >= cb.failureThreshold {
		cb.state = StateOpen
	}
}

// State returns the current breaker state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
