package worker

import (
    "testing"
    "time"
)

func TestBackoffCaps(t *testing.T) {
    b := backoff(10, 100*time.Millisecond, 1*time.Second)
    if b != 1*time.Second { t.Fatalf("expected cap at 1s, got %v", b) }
}

