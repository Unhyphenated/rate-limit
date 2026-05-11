package breaker

import (
	"errors"
	"sync"
	"testing"
	"time"
)

func TestBreakerOpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		FailureThreshold:  3,
		CooldownPeriod:    100 * time.Millisecond,
		HalfOpenSuccesses: 2,
	})

	failingFn := func() (any, error) {
		return nil, errors.New("boom")
	}

	for i := 0; i < 3; i++ {
		_, err := cb.Execute(failingFn)
		if err == ErrCircuitOpen {
			t.Fatalf("breaker should not be open yet, iter %d", i)
		}
		if err == nil {
			t.Fatalf("expected failure, iter %d", i)
		}
	}

	_, err := cb.Execute(failingFn)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestBreakerDoesNotOpenBelowThreshold(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		FailureThreshold:  5,
		CooldownPeriod:    time.Second,
		HalfOpenSuccesses: 1,
	})

	fail := func() (any, error) { return nil, errors.New("x") }
	ok := func() (any, error) { return 1, nil }

	for i := 0; i < 4; i++ {
		_, err := cb.Execute(fail)
		if errors.Is(err, ErrCircuitOpen) {
			t.Fatalf("unexpected open at %d", i)
		}
	}
	_, err := cb.Execute(ok)
	if err != nil {
		t.Fatalf("success should clear failures: %v", err)
	}
	for i := 0; i < 4; i++ {
		_, err := cb.Execute(fail)
		if errors.Is(err, ErrCircuitOpen) {
			t.Fatalf("unexpected open after reset, iter %d", i)
		}
	}
}

func TestBreakerHalfOpenRecoversWithSuccesses(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		FailureThreshold:  2,
		CooldownPeriod:    20 * time.Millisecond,
		HalfOpenSuccesses: 2,
	})

	fail := func() (any, error) { return nil, errors.New("x") }
	ok := func() (any, error) { return "ok", nil }

	for i := 0; i < 2; i++ {
		_, _ = cb.Execute(fail)
	}
	_, err := cb.Execute(fail)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected open, got %v", err)
	}

	time.Sleep(25 * time.Millisecond)

	v, err := cb.Execute(ok)
	if err != nil {
		t.Fatalf("half-open probe 1: %v", err)
	}
	if v != "ok" {
		t.Fatalf("got %v", v)
	}

	v, err = cb.Execute(ok)
	if err != nil {
		t.Fatalf("half-open probe 2: %v", err)
	}
	if v != "ok" {
		t.Fatalf("got %v", v)
	}

	// closed again: failures should count from zero
	_, _ = cb.Execute(fail)
	_, err = cb.Execute(fail)
	if errors.Is(err, ErrCircuitOpen) {
		t.Fatal("should need FailureThreshold failures after close")
	}
}

func TestBreakerHalfOpenFailureReopensImmediately(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		FailureThreshold:  1,
		CooldownPeriod:    20 * time.Millisecond,
		HalfOpenSuccesses: 3,
	})

	fail := func() (any, error) { return nil, errors.New("x") }
	ok := func() (any, error) { return 1, nil }

	_, _ = cb.Execute(fail)
	_, err := cb.Execute(fail)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected open: %v", err)
	}

	time.Sleep(25 * time.Millisecond)

	_, err = cb.Execute(ok)
	if err != nil {
		t.Fatalf("first half-open success: %v", err)
	}

	_, err = cb.Execute(fail)
	if err == nil || errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected inner error, not open short-circuit: %v", err)
	}

	_, err = cb.Execute(fail)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen after half-open fail, got %v", err)
	}
}

func TestBreakerOpenShortCircuitsBeforeCooldown(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		FailureThreshold:  1,
		CooldownPeriod:    200 * time.Millisecond,
		HalfOpenSuccesses: 1,
	})

	fail := func() (any, error) { return nil, errors.New("x") }

	_, _ = cb.Execute(fail)
	_, err := cb.Execute(fail)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Fatalf("expected open: %v", err)
	}

	_, err = cb.Execute(fail)
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("still cooling down: expected ErrCircuitOpen, got %v", err)
	}
}

func TestBreakerExecuteConcurrent(t *testing.T) {
	cb := NewCircuitBreaker(Config{
		FailureThreshold:  10,
		CooldownPeriod:    time.Second,
		HalfOpenSuccesses: 1,
	})
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = cb.Execute(func() (any, error) {
				return 1, nil
			})
		}()
	}
	wg.Wait()
}
