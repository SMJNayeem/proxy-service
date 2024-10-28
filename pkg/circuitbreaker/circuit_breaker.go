package circuitbreaker

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// CircuitBreaker errors
var (
	ErrCircuitOpen = fmt.Errorf("circuit breaker is open")
)

// Additional helper types and constants
type State string

const (
	StateClosed   State = "closed"
	StateOpen     State = "open"
	StateHalfOpen State = "half-open"
)

// CircuitBreakerConfig holds configuration for circuit breaker
type CircuitBreakerConfig struct {
	MaxFailures   int
	Timeout       time.Duration
	OnStateChange func(from, to State)
}

type CircuitBreaker struct {
	mutex       sync.RWMutex
	failures    int
	lastFailure time.Time
	state       string // closed, open, half-open
	maxFailures int
	timeout     time.Duration
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       "closed",
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

func (cb *CircuitBreaker) Execute(ctx context.Context, cmd func() error) error {
	cb.mutex.RLock()
	if cb.state == "open" {
		if time.Since(cb.lastFailure) > cb.timeout {
			cb.mutex.RUnlock()
			cb.mutex.Lock()
			cb.state = "half-open"
			cb.mutex.Unlock()
		} else {
			cb.mutex.RUnlock()
			return ErrCircuitOpen
		}
	} else {
		cb.mutex.RUnlock()
	}

	err := cmd()

	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = "open"
		}
		return err
	}

	if cb.state == "half-open" {
		cb.state = "closed"
	}
	cb.failures = 0
	return nil
}

// New creates a new circuit breaker with the given configuration
func New(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		state:       string(StateClosed),
		maxFailures: config.MaxFailures,
		timeout:     config.Timeout,
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() State {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	return State(cb.state)
}
