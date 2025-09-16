// Copyright 2025 James Ross
package breaker

import (
	"sync"
	"testing"
	"time"
)

// Test that in HalfOpen under concurrent load, only a single probe is allowed at a time.
func TestBreakerHalfOpenSingleProbeUnderLoad(t *testing.T) {
	cb := New(20*time.Millisecond, 50*time.Millisecond, 0.5, 2)
	if cb.State() != Closed {
		t.Fatal("expected closed")
	}
	cb.Record(false)
	cb.Record(false)
	if cb.State() != Open {
		t.Fatal("expected open after 2 failures")
	}

	// Wait for cooldown to enter HalfOpen
	time.Sleep(60 * time.Millisecond)

	// Concurrently call Allow; only one should be allowed
	const N = 100
	var wg sync.WaitGroup
	wg.Add(N)
	trues := 0
	var mu sync.Mutex
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			if cb.Allow() {
				mu.Lock()
				trues++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if trues != 1 {
		t.Fatalf("expected exactly 1 allowed probe, got %d", trues)
	}

	// Fail the probe to remain Open
	cb.Record(false)
	if cb.State() != Open {
		t.Fatalf("expected open after failed probe, got %v", cb.State())
	}

	// Wait again to HalfOpen and check single probe again
	time.Sleep(60 * time.Millisecond)
	trues = 0
	wg.Add(N)
	for i := 0; i < N; i++ {
		go func() {
			defer wg.Done()
			if cb.Allow() {
				mu.Lock()
				trues++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if trues != 1 {
		t.Fatalf("expected exactly 1 allowed probe in second cycle, got %d", trues)
	}

	// Succeed the probe to close
	cb.Record(true)
	if cb.State() != Closed {
		t.Fatalf("expected closed after successful probe, got %v", cb.State())
	}
}
