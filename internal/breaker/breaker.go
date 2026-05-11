package breaker

import (
	"errors"
	"sync"
	"time"
)

// ErrCircuitOpen is returned when the breaker is open and the cooldown has not elapsed.
var ErrCircuitOpen = errors.New("circuit breaker is open")

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

type Config struct {
	FailureThreshold  int
	CooldownPeriod    time.Duration
	HalfOpenSuccesses int
}

type CircuitBreaker struct {
	mu         sync.Mutex
	state      State
	failures   int
	successes  int
	openedAt   time.Time
	config     Config
}

// NewCircuitBreaker returns a breaker in the closed state. Config must use positive thresholds/durations.
func NewCircuitBreaker(config Config) *CircuitBreaker {
	return &CircuitBreaker{
		state:  StateClosed,
		config: config,
	}
}

func (cb *CircuitBreaker) Execute(fn func() (any, error)) (any, error) {
	cb.mu.Lock()
	switch cb.state {
	case StateOpen:
		if time.Since(cb.openedAt) >= cb.config.CooldownPeriod {
			cb.state = StateHalfOpen
			cb.successes = 0
		} else {
			cb.mu.Unlock()
			return nil, ErrCircuitOpen
		}
	case StateHalfOpen, StateClosed:
		// proceed
	}
	cb.mu.Unlock()

	result, err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()
	if err != nil {
		cb.onFailure()
	} else {
		cb.onSuccess()
	}
	return result, err
}

func (cb *CircuitBreaker) onFailure() {
	switch cb.state {
	case StateClosed:
		cb.failures++
		if cb.failures >= cb.config.FailureThreshold {
			cb.state = StateOpen
			cb.openedAt = time.Now()
			cb.failures = 0
		}
	case StateHalfOpen:
		cb.state = StateOpen
		cb.openedAt = time.Now()
		cb.successes = 0
		cb.failures = 0
	}
}

func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		cb.failures = 0
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.HalfOpenSuccesses {
			cb.state = StateClosed
			cb.successes = 0
			cb.failures = 0
		}
	}
}
