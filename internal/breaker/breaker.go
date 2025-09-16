// Copyright 2025 James Ross
// Copyright 2025 James Ross
package breaker

import (
	"sync"
	"time"
)

type State int

const (
	Closed State = iota
	HalfOpen
	Open
)

type result struct {
	t  time.Time
	ok bool
}

// CircuitBreaker with sliding window and cooldown.
type CircuitBreaker struct {
	mu               sync.Mutex
	state            State
	window           time.Duration
	cooldown         time.Duration
	failureThresh    float64
	minSamples       int
	lastTransition   time.Time
	results          []result
	halfOpenInFlight bool
}

func New(window time.Duration, cooldown time.Duration, failureThresh float64, minSamples int) *CircuitBreaker {
	return &CircuitBreaker{state: Closed, window: window, cooldown: cooldown, failureThresh: failureThresh, minSamples: minSamples, lastTransition: time.Now()}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	switch cb.state {
	case Open:
		if time.Since(cb.lastTransition) >= cb.cooldown {
			cb.state = HalfOpen
			cb.lastTransition = time.Now()
			cb.halfOpenInFlight = false
			// allow exactly one probe once we enter HalfOpen; next branch handles flag
			cb.halfOpenInFlight = true
			return true
		}
		return false
	case HalfOpen:
		if cb.halfOpenInFlight {
			return false
		}
		cb.halfOpenInFlight = true
		return true
	default:
		return true
	}
}

func (cb *CircuitBreaker) Record(ok bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	now := time.Now()
	// purge old
	cutoff := now.Add(-cb.window)
	filtered := cb.results[:0]
	for _, r := range cb.results {
		if r.t.After(cutoff) {
			filtered = append(filtered, r)
		}
	}
	cb.results = append(filtered, result{t: now, ok: ok})

	// compute failure rate
	total := len(cb.results)
	if total < cb.minSamples {
		if cb.state == HalfOpen {
			if ok {
				cb.state = Closed
				cb.lastTransition = now
			} else {
				cb.state = Open
				cb.lastTransition = now
			}
		}
		return
	}
	fails := 0
	for _, r := range cb.results {
		if !r.ok {
			fails++
		}
	}
	rate := float64(fails) / float64(total)
	switch cb.state {
	case Closed:
		if rate >= cb.failureThresh {
			cb.state = Open
			cb.lastTransition = now
		}
	case HalfOpen:
		if ok {
			cb.state = Closed
		} else {
			cb.state = Open
		}
		// the single probe completed; allow a future probe after cooldown or next Allow
		cb.halfOpenInFlight = false
		cb.lastTransition = now
	case Open:
		// handled in Allow()
	}
}
