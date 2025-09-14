// Copyright 2025 James Ross
package backpressure

import (
	"time"
)

// ShouldAllow determines if the circuit breaker should allow a request
func (cb *CircuitBreaker) ShouldAllow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.State {
	case Closed:
		// Check if we should trip due to failures in the window
		if cb.FailureCount >= cb.Config.FailureThreshold {
			if now.Sub(cb.LastFailureTime) <= cb.Config.TripWindow {
				cb.State = Open
				cb.LastFailureTime = now
				return false
			} else {
				// Failure window expired, reset failure count
				cb.FailureCount = 0
				return true
			}
		}
		return true

	case Open:
		// Check if we should transition to half-open
		if now.Sub(cb.LastFailureTime) >= cb.Config.RecoveryTimeout {
			cb.State = HalfOpen
			cb.SuccessCount = 0
			cb.LastProbe = now
			return true // Allow probe request
		}
		return false

	case HalfOpen:
		// Allow periodic probes
		if now.Sub(cb.LastProbe) >= cb.Config.ProbeInterval {
			cb.LastProbe = now
			return true
		}
		return false

	default:
		// Unknown state - assume closed
		cb.State = Closed
		return true
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.State {
	case Closed:
		// Reset failure count on success
		cb.FailureCount = 0

	case HalfOpen:
		// Count successes in half-open state
		cb.SuccessCount++
		if cb.SuccessCount >= cb.Config.RecoveryThreshold {
			// Enough successes - close the circuit
			cb.State = Closed
			cb.FailureCount = 0
			cb.SuccessCount = 0
		}

	case Open:
		// Unexpected success in open state - shouldn't happen
		// but we'll treat it as moving to half-open
		cb.State = HalfOpen
		cb.SuccessCount = 1
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	now := time.Now()

	switch cb.State {
	case Closed:
		cb.FailureCount++
		cb.LastFailureTime = now

		// Check if we should trip
		if cb.FailureCount >= cb.Config.FailureThreshold {
			cb.State = Open
		}

	case HalfOpen:
		// Failure in half-open immediately trips back to open
		cb.State = Open
		cb.FailureCount++
		cb.LastFailureTime = now
		cb.SuccessCount = 0

	case Open:
		// Already open, just update failure time
		cb.LastFailureTime = now
	}
}

// GetState returns the current circuit state (thread-safe)
func (cb *CircuitBreaker) GetState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.State
}

// GetFailureCount returns the current failure count (thread-safe)
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.FailureCount
}

// GetSuccessCount returns the current success count in half-open state (thread-safe)
func (cb *CircuitBreaker) GetSuccessCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.SuccessCount
}

// GetLastFailureTime returns the time of the last failure (thread-safe)
func (cb *CircuitBreaker) GetLastFailureTime() time.Time {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.LastFailureTime
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.State = Closed
	cb.FailureCount = 0
	cb.SuccessCount = 0
	cb.LastFailureTime = time.Time{}
	cb.LastProbe = time.Time{}
}

// ForceOpen forces the circuit breaker to open state
func (cb *CircuitBreaker) ForceOpen() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.State = Open
	cb.LastFailureTime = time.Now()
}

// ForceClose forces the circuit breaker to closed state
func (cb *CircuitBreaker) ForceClose() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.State = Closed
	cb.FailureCount = 0
	cb.SuccessCount = 0
}

// GetStats returns current circuit breaker statistics
func (cb *CircuitBreaker) GetStats() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":              cb.State.String(),
		"failure_count":      cb.FailureCount,
		"success_count":      cb.SuccessCount,
		"last_failure_time":  cb.LastFailureTime,
		"last_probe_time":    cb.LastProbe,
		"failure_threshold":  cb.Config.FailureThreshold,
		"recovery_threshold": cb.Config.RecoveryThreshold,
		"trip_window":        cb.Config.TripWindow,
		"recovery_timeout":   cb.Config.RecoveryTimeout,
		"probe_interval":     cb.Config.ProbeInterval,
	}
}

// IsHealthy returns true if the circuit breaker is allowing requests
func (cb *CircuitBreaker) IsHealthy() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.State != Open
}

// TimeUntilRecovery returns the time remaining until the circuit can transition to half-open
func (cb *CircuitBreaker) TimeUntilRecovery() time.Duration {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	if cb.State != Open {
		return 0
	}

	remaining := cb.Config.RecoveryTimeout - time.Since(cb.LastFailureTime)
	if remaining < 0 {
		return 0
	}
	return remaining
}