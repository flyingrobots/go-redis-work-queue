// Copyright 2025 James Ross
package breaker

import (
    "testing"
    "time"
)

func TestBreakerTransitions(t *testing.T) {
    cb := New(2*time.Second, 200*time.Millisecond, 0.5, 2)
    if cb.State() != Closed { t.Fatal("expected closed") }
    cb.Record(false)
    cb.Record(false)
    time.Sleep(10 * time.Millisecond)
    if cb.State() != Open { t.Fatal("expected open") }
    if cb.Allow() != false { t.Fatal("should not allow until cooldown") }
    time.Sleep(250 * time.Millisecond)
    if cb.Allow() != true { t.Fatal("should allow probe in half-open") }
    cb.Record(true)
    if cb.State() != Closed { t.Fatal("expected closed after probe success") }
}
