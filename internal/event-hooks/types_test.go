// Copyright 2025 James Ross
package eventhooks

import (
	"testing"
	"time"
)

func TestEventFilter_Matches(t *testing.T) {
	tests := []struct {
		name     string
		filter   EventFilter
		event    JobEvent
		expected bool
	}{
		{
			name: "matches all filters",
			filter: EventFilter{
				Events:      []EventType{EventJobStarted, EventJobSucceeded},
				Queues:      []string{"test-queue"},
				MinPriority: func() *int { i := 5; return &i }(),
			},
			event: JobEvent{
				Event:    EventJobStarted,
				Queue:    "test-queue",
				Priority: 7,
			},
			expected: true,
		},
		{
			name: "fails event type filter",
			filter: EventFilter{
				Events: []EventType{EventJobSucceeded},
				Queues: []string{"test-queue"},
			},
			event: JobEvent{
				Event: EventJobStarted,
				Queue: "test-queue",
			},
			expected: false,
		},
		{
			name: "fails queue filter",
			filter: EventFilter{
				Events: []EventType{EventJobStarted},
				Queues: []string{"other-queue"},
			},
			event: JobEvent{
				Event: EventJobStarted,
				Queue: "test-queue",
			},
			expected: false,
		},
		{
			name: "passes with wildcard queue",
			filter: EventFilter{
				Events: []EventType{EventJobStarted},
				Queues: []string{"*"},
			},
			event: JobEvent{
				Event: EventJobStarted,
				Queue: "any-queue",
			},
			expected: true,
		},
		{
			name: "fails min priority",
			filter: EventFilter{
				Events:      []EventType{EventJobStarted},
				Queues:      []string{"*"},
				MinPriority: func() *int { i := 5; return &i }(),
			},
			event: JobEvent{
				Event:    EventJobStarted,
				Queue:    "test-queue",
				Priority: 3,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.filter.Matches(tt.event)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDefaultRetryPolicy(t *testing.T) {
	policy := DefaultRetryPolicy()

	if policy.Strategy != "exponential" {
		t.Errorf("expected exponential strategy, got %s", policy.Strategy)
	}

	if policy.InitialDelay != 1*time.Second {
		t.Errorf("expected 1s initial delay, got %v", policy.InitialDelay)
	}

	if policy.MaxRetries != 5 {
		t.Errorf("expected 5 max retries, got %d", policy.MaxRetries)
	}
}

func TestWebhookSubscription_Concurrency(t *testing.T) {
	sub := &WebhookSubscription{
		ID:   "test-sub",
		Name: "Test Subscription",
		URL:  "https://example.com/webhook",
	}

	// Test concurrent access doesn't panic
	done := make(chan bool, 2)

	go func() {
		sub.mu.Lock()
		sub.FailureCount++
		sub.mu.Unlock()
		done <- true
	}()

	go func() {
		sub.mu.RLock()
		_ = sub.FailureCount
		sub.mu.RUnlock()
		done <- true
	}()

	<-done
	<-done
}