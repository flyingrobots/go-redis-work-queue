// Copyright 2025 James Ross
package backpressure

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCircuitBreaker(t *testing.T) {
	config := DefaultCircuitConfig()
	cb := NewCircuitBreaker(config)

	assert.Equal(t, Closed, cb.State)
	assert.Equal(t, 0, cb.FailureCount)
	assert.Equal(t, 0, cb.SuccessCount)
	assert.Equal(t, config, cb.Config)
}

func TestCircuitBreakerClosedState(t *testing.T) {
	config := CircuitConfig{
		FailureThreshold:  3,
		RecoveryThreshold: 2,
		TripWindow:        30 * time.Second,
		RecoveryTimeout:   60 * time.Second,
		ProbeInterval:     5 * time.Second,
	}
	cb := NewCircuitBreaker(config)

	// Initially closed and should allow requests
	assert.True(t, cb.ShouldAllow())
	assert.Equal(t, Closed, cb.GetState())

	// Record success - should remain closed
	cb.RecordSuccess()
	assert.True(t, cb.ShouldAllow())
	assert.Equal(t, Closed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())

	// Record some failures but not enough to trip
	cb.RecordFailure()
	assert.True(t, cb.ShouldAllow())
	assert.Equal(t, Closed, cb.GetState())
	assert.Equal(t, 1, cb.GetFailureCount())

	cb.RecordFailure()
	assert.True(t, cb.ShouldAllow())
	assert.Equal(t, Closed, cb.GetState())
	assert.Equal(t, 2, cb.GetFailureCount())

	// One more failure should trip the circuit
	cb.RecordFailure()
	assert.False(t, cb.ShouldAllow()) // Now trips to Open
	assert.Equal(t, Open, cb.GetState())
	assert.Equal(t, 3, cb.GetFailureCount())
}

func TestCircuitBreakerOpenState(t *testing.T) {
	config := CircuitConfig{
		FailureThreshold:  2,
		RecoveryThreshold: 2,
		TripWindow:        30 * time.Second,
		RecoveryTimeout:   100 * time.Millisecond, // Short timeout for testing
		ProbeInterval:     5 * time.Second,
	}
	cb := NewCircuitBreaker(config)

	// Trip the circuit
	cb.RecordFailure()
	cb.RecordFailure()
	assert.False(t, cb.ShouldAllow())
	assert.Equal(t, Open, cb.GetState())

	// Should not allow requests while in recovery timeout
	assert.False(t, cb.ShouldAllow())
	assert.Equal(t, Open, cb.GetState())

	// Wait for recovery timeout
	time.Sleep(150 * time.Millisecond)

	// Should transition to half-open and allow probe
	assert.True(t, cb.ShouldAllow())
	assert.Equal(t, HalfOpen, cb.GetState())
}

func TestCircuitBreakerHalfOpenState(t *testing.T) {
	config := CircuitConfig{
		FailureThreshold:  2,
		RecoveryThreshold: 2,
		TripWindow:        30 * time.Second,
		RecoveryTimeout:   10 * time.Millisecond,
		ProbeInterval:     50 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Trip the circuit and wait for half-open
	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	cb.ShouldAllow() // Transition to half-open

	assert.Equal(t, HalfOpen, cb.GetState())
	assert.Equal(t, 0, cb.GetSuccessCount())

	// Record success
	cb.RecordSuccess()
	assert.Equal(t, HalfOpen, cb.GetState())
	assert.Equal(t, 1, cb.GetSuccessCount())

	// One more success should close the circuit
	cb.RecordSuccess()
	assert.Equal(t, Closed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())
	assert.Equal(t, 0, cb.GetSuccessCount())

	// Should allow requests when closed
	assert.True(t, cb.ShouldAllow())
}

func TestCircuitBreakerHalfOpenFailure(t *testing.T) {
	config := CircuitConfig{
		FailureThreshold:  2,
		RecoveryThreshold: 2,
		TripWindow:        30 * time.Second,
		RecoveryTimeout:   10 * time.Millisecond,
		ProbeInterval:     5 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Trip the circuit and wait for half-open
	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	cb.ShouldAllow() // Transition to half-open

	assert.Equal(t, HalfOpen, cb.GetState())

	// Record failure in half-open state
	cb.RecordFailure()

	// Should immediately trip back to open
	assert.Equal(t, Open, cb.GetState())
	assert.False(t, cb.ShouldAllow())
}

func TestCircuitBreakerProbeInterval(t *testing.T) {
	config := CircuitConfig{
		FailureThreshold:  2,
		RecoveryThreshold: 2,
		TripWindow:        30 * time.Second,
		RecoveryTimeout:   10 * time.Millisecond,
		ProbeInterval:     50 * time.Millisecond,
	}
	cb := NewCircuitBreaker(config)

	// Trip the circuit and transition to half-open
	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(20 * time.Millisecond)
	cb.ShouldAllow() // First probe, transitions to half-open

	assert.Equal(t, HalfOpen, cb.GetState())

	// Immediate second call should not be allowed (probe interval not reached)
	assert.False(t, cb.ShouldAllow())

	// Wait for probe interval
	time.Sleep(60 * time.Millisecond)

	// Now should allow probe
	assert.True(t, cb.ShouldAllow())
}

func TestCircuitBreakerTripWindow(t *testing.T) {
	config := CircuitConfig{
		FailureThreshold:  2,
		RecoveryThreshold: 2,
		TripWindow:        50 * time.Millisecond, // Short window
		RecoveryTimeout:   60 * time.Second,
		ProbeInterval:     5 * time.Second,
	}
	cb := NewCircuitBreaker(config)

	// Record one failure
	cb.RecordFailure()
	assert.Equal(t, Closed, cb.GetState())
	assert.Equal(t, 1, cb.GetFailureCount())

	// Wait for trip window to expire
	time.Sleep(60 * time.Millisecond)

	// Record another failure - should not trip since window expired
	cb.RecordFailure()
	assert.Equal(t, Closed, cb.GetState()) // Failure count should reset
	assert.Equal(t, 1, cb.GetFailureCount()) // Count reset, this is the new first failure

	// Add another failure quickly
	cb.RecordFailure()
	assert.Equal(t, Open, cb.GetState()) // Now should trip
}

func TestCircuitBreakerReset(t *testing.T) {
	config := DefaultCircuitConfig()
	cb := NewCircuitBreaker(config)

	// Trip the circuit
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	assert.Equal(t, Open, cb.GetState())
	assert.Greater(t, cb.GetFailureCount(), 0)

	// Reset the circuit
	cb.Reset()

	assert.Equal(t, Closed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())
	assert.Equal(t, 0, cb.GetSuccessCount())
	assert.True(t, cb.GetLastFailureTime().IsZero())
	assert.True(t, cb.ShouldAllow())
}

func TestCircuitBreakerForceOpen(t *testing.T) {
	config := DefaultCircuitConfig()
	cb := NewCircuitBreaker(config)

	assert.Equal(t, Closed, cb.GetState())
	assert.True(t, cb.ShouldAllow())

	cb.ForceOpen()

	assert.Equal(t, Open, cb.GetState())
	assert.False(t, cb.ShouldAllow())
	assert.False(t, cb.GetLastFailureTime().IsZero())
}

func TestCircuitBreakerForceClose(t *testing.T) {
	config := DefaultCircuitConfig()
	cb := NewCircuitBreaker(config)

	// Trip the circuit
	for i := 0; i < config.FailureThreshold+1; i++ {
		cb.RecordFailure()
	}
	assert.Equal(t, Open, cb.GetState())

	cb.ForceClose()

	assert.Equal(t, Closed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())
	assert.Equal(t, 0, cb.GetSuccessCount())
	assert.True(t, cb.ShouldAllow())
}

func TestCircuitBreakerGetStats(t *testing.T) {
	config := DefaultCircuitConfig()
	cb := NewCircuitBreaker(config)

	stats := cb.GetStats()

	require.NotNil(t, stats)
	assert.Equal(t, "closed", stats["state"])
	assert.Equal(t, 0, stats["failure_count"])
	assert.Equal(t, 0, stats["success_count"])
	assert.Equal(t, config.FailureThreshold, stats["failure_threshold"])
	assert.Equal(t, config.RecoveryThreshold, stats["recovery_threshold"])
	assert.Equal(t, config.TripWindow, stats["trip_window"])
	assert.Equal(t, config.RecoveryTimeout, stats["recovery_timeout"])
	assert.Equal(t, config.ProbeInterval, stats["probe_interval"])

	// Trip circuit and check stats change
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()

	stats = cb.GetStats()
	assert.Equal(t, "open", stats["state"])
	assert.Greater(t, stats["failure_count"], 0)
}

func TestCircuitBreakerIsHealthy(t *testing.T) {
	config := DefaultCircuitConfig()
	cb := NewCircuitBreaker(config)

	// Initially healthy (closed)
	assert.True(t, cb.IsHealthy())

	// Trip circuit
	for i := 0; i < config.FailureThreshold+1; i++ {
		cb.RecordFailure()
	}

	// Not healthy when open
	assert.False(t, cb.IsHealthy())

	// Force to half-open
	cb.Reset()
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()
	cb.RecordFailure()
	time.Sleep(time.Millisecond) // Ensure time passes
	// Set to half-open manually for test
	cb.mu.Lock()
	cb.State = HalfOpen
	cb.mu.Unlock()

	// Healthy in half-open (allowing probes)
	assert.True(t, cb.IsHealthy())
}

func TestCircuitBreakerTimeUntilRecovery(t *testing.T) {
	config := CircuitConfig{
		FailureThreshold:  2,
		RecoveryThreshold: 2,
		TripWindow:        30 * time.Second,
		RecoveryTimeout:   1 * time.Second,
		ProbeInterval:     5 * time.Second,
	}
	cb := NewCircuitBreaker(config)

	// Not open, should be 0
	assert.Equal(t, time.Duration(0), cb.TimeUntilRecovery())

	// Trip the circuit
	cb.RecordFailure()
	cb.RecordFailure()

	assert.Equal(t, Open, cb.GetState())

	// Should have time remaining
	remaining := cb.TimeUntilRecovery()
	assert.Greater(t, remaining, time.Duration(0))
	assert.LessOrEqual(t, remaining, config.RecoveryTimeout)

	// Wait for recovery timeout
	time.Sleep(config.RecoveryTimeout + 10*time.Millisecond)

	// Should be 0 or negative (ready to try)
	remaining = cb.TimeUntilRecovery()
	assert.Equal(t, time.Duration(0), remaining)
}

func TestCircuitBreakerConcurrency(t *testing.T) {
	config := DefaultCircuitConfig()
	cb := NewCircuitBreaker(config)

	// Test concurrent access to circuit breaker
	done := make(chan bool, 100)

	// Start multiple goroutines that access the circuit breaker
	for i := 0; i < 50; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 10; j++ {
				cb.ShouldAllow()
				cb.RecordSuccess()
				cb.RecordFailure()
				cb.GetState()
				cb.GetStats()
			}
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 50; i++ {
		select {
		case <-done:
			// Continue
		case <-time.After(5 * time.Second):
			t.Fatal("Timeout waiting for concurrent operations")
		}
	}

	// Circuit breaker should still be in a valid state
	state := cb.GetState()
	assert.Contains(t, []CircuitState{Closed, Open, HalfOpen}, state)
}