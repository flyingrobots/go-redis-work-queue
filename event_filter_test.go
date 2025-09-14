// Copyright 2025 James Ross
package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// EventFilter handles webhook subscription filtering
type EventFilter struct{}

func NewEventFilter() *EventFilter {
	return &EventFilter{}
}

// MatchesSubscription determines if a job event matches a webhook subscription
func (f *EventFilter) MatchesSubscription(event JobEvent, subscription WebhookSubscription) bool {
	// Check event type filter
	if !f.containsString(subscription.Events, event.Event) && !f.containsString(subscription.Events, "*") {
		return false
	}

	// Check queue filter
	if !f.containsString(subscription.Queues, event.Queue) && !f.containsString(subscription.Queues, "*") {
		return false
	}

	// Check priority filter
	if subscription.MinPriority != nil && event.Priority < *subscription.MinPriority {
		return false
	}

	return true
}

// GetMatchingSubscriptions returns all subscriptions that match a given event
func (f *EventFilter) GetMatchingSubscriptions(event JobEvent, subscriptions []WebhookSubscription) []WebhookSubscription {
	var matches []WebhookSubscription
	for _, sub := range subscriptions {
		if f.MatchesSubscription(event, sub) {
			matches = append(matches, sub)
		}
	}
	return matches
}

// FilterEventsBySubscription returns events that match a subscription's filters
func (f *EventFilter) FilterEventsBySubscription(events []JobEvent, subscription WebhookSubscription) []JobEvent {
	var filtered []JobEvent
	for _, event := range events {
		if f.MatchesSubscription(event, subscription) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// ValidateSubscriptionFilters checks if subscription filters are valid
func (f *EventFilter) ValidateSubscriptionFilters(subscription WebhookSubscription) error {
	if len(subscription.Events) == 0 {
		return &FilterError{Type: "validation", Message: "at least one event type must be specified"}
	}

	if len(subscription.Queues) == 0 {
		return &FilterError{Type: "validation", Message: "at least one queue must be specified"}
	}

	// Validate event types
	validEvents := []string{"job_enqueued", "job_started", "job_succeeded", "job_failed", "job_dlq", "job_retried", "*"}
	for _, event := range subscription.Events {
		if !f.containsString(validEvents, event) {
			return &FilterError{Type: "validation", Message: "invalid event type: " + event}
		}
	}

	// Validate priority range
	if subscription.MinPriority != nil && (*subscription.MinPriority < 1 || *subscription.MinPriority > 10) {
		return &FilterError{Type: "validation", Message: "priority must be between 1 and 10"}
	}

	return nil
}

// Helper function to check if slice contains string
func (f *EventFilter) containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// FilterError represents a filtering error
type FilterError struct {
	Type    string
	Message string
}

func (e *FilterError) Error() string {
	return e.Message
}

// JobEvent represents a job lifecycle event
type JobEvent struct {
	Event       string                 `json:"event"`
	Timestamp   time.Time              `json:"timestamp"`
	JobID       string                 `json:"job_id"`
	Queue       string                 `json:"queue"`
	Priority    int                    `json:"priority"`
	Attempt     int                    `json:"attempt"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Duration    *time.Duration         `json:"duration,omitempty"`
	Worker      string                 `json:"worker,omitempty"`
	Payload     map[string]interface{} `json:"payload,omitempty"`
	TraceID     string                 `json:"trace_id,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
}

// WebhookSubscription represents a webhook subscription configuration
type WebhookSubscription struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	URL             string            `json:"url"`
	Secret          string            `json:"-"`
	Events          []string          `json:"events"`
	Queues          []string          `json:"queues"`
	MinPriority     *int              `json:"min_priority,omitempty"`
	MaxRetries      int               `json:"max_retries"`
	Timeout         time.Duration     `json:"timeout"`
	RateLimit       int               `json:"rate_limit"`
	Headers         map[string]string `json:"headers"`
	IncludePayload  bool              `json:"include_payload"`
	PayloadFields   []string          `json:"payload_fields,omitempty"`
	RedactFields    []string          `json:"redact_fields,omitempty"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	LastSuccess     *time.Time        `json:"last_success,omitempty"`
	LastFailure     *time.Time        `json:"last_failure,omitempty"`
	FailureCount    int               `json:"failure_count"`
	Disabled        bool              `json:"disabled"`
}

// Tests for Event Filter Matching

func TestEventFilter_MatchesSubscription(t *testing.T) {
	filter := NewEventFilter()

	t.Run("exact event and queue match", func(t *testing.T) {
		event := JobEvent{
			Event:    "job_failed",
			Queue:    "test_queue",
			Priority: 5,
		}

		subscription := WebhookSubscription{
			Events: []string{"job_failed"},
			Queues: []string{"test_queue"},
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.True(t, matches)
	})

	t.Run("wildcard event matching", func(t *testing.T) {
		event := JobEvent{
			Event:    "job_succeeded",
			Queue:    "test_queue",
			Priority: 5,
		}

		subscription := WebhookSubscription{
			Events: []string{"*"},
			Queues: []string{"test_queue"},
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.True(t, matches)
	})

	t.Run("wildcard queue matching", func(t *testing.T) {
		event := JobEvent{
			Event:    "job_failed",
			Queue:    "any_queue",
			Priority: 5,
		}

		subscription := WebhookSubscription{
			Events: []string{"job_failed"},
			Queues: []string{"*"},
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.True(t, matches)
	})

	t.Run("multiple events and queues", func(t *testing.T) {
		event := JobEvent{
			Event:    "job_dlq",
			Queue:    "priority_queue",
			Priority: 8,
		}

		subscription := WebhookSubscription{
			Events: []string{"job_failed", "job_dlq", "job_retried"},
			Queues: []string{"test_queue", "priority_queue", "batch_queue"},
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.True(t, matches)
	})

	t.Run("priority filter - above minimum", func(t *testing.T) {
		minPriority := 7
		event := JobEvent{
			Event:    "job_failed",
			Queue:    "test_queue",
			Priority: 9,
		}

		subscription := WebhookSubscription{
			Events:      []string{"job_failed"},
			Queues:      []string{"test_queue"},
			MinPriority: &minPriority,
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.True(t, matches)
	})

	t.Run("priority filter - below minimum", func(t *testing.T) {
		minPriority := 7
		event := JobEvent{
			Event:    "job_failed",
			Queue:    "test_queue",
			Priority: 5,
		}

		subscription := WebhookSubscription{
			Events:      []string{"job_failed"},
			Queues:      []string{"test_queue"},
			MinPriority: &minPriority,
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.False(t, matches)
	})

	t.Run("priority filter - exact minimum", func(t *testing.T) {
		minPriority := 5
		event := JobEvent{
			Event:    "job_failed",
			Queue:    "test_queue",
			Priority: 5,
		}

		subscription := WebhookSubscription{
			Events:      []string{"job_failed"},
			Queues:      []string{"test_queue"},
			MinPriority: &minPriority,
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.True(t, matches)
	})

	t.Run("no priority filter", func(t *testing.T) {
		event := JobEvent{
			Event:    "job_failed",
			Queue:    "test_queue",
			Priority: 1, // Very low priority
		}

		subscription := WebhookSubscription{
			Events:      []string{"job_failed"},
			Queues:      []string{"test_queue"},
			MinPriority: nil, // No priority filter
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.True(t, matches)
	})

	t.Run("event mismatch", func(t *testing.T) {
		event := JobEvent{
			Event:    "job_succeeded",
			Queue:    "test_queue",
			Priority: 5,
		}

		subscription := WebhookSubscription{
			Events: []string{"job_failed", "job_dlq"},
			Queues: []string{"test_queue"},
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.False(t, matches)
	})

	t.Run("queue mismatch", func(t *testing.T) {
		event := JobEvent{
			Event:    "job_failed",
			Queue:    "different_queue",
			Priority: 5,
		}

		subscription := WebhookSubscription{
			Events: []string{"job_failed"},
			Queues: []string{"test_queue", "priority_queue"},
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.False(t, matches)
	})
}

func TestEventFilter_GetMatchingSubscriptions(t *testing.T) {
	filter := NewEventFilter()

	event := JobEvent{
		Event:    "job_failed",
		Queue:    "test_queue",
		Priority: 8,
	}

	subscriptions := []WebhookSubscription{
		{
			ID:     "sub_001",
			Events: []string{"job_failed"},
			Queues: []string{"test_queue"},
		},
		{
			ID:     "sub_002",
			Events: []string{"*"},
			Queues: []string{"*"},
		},
		{
			ID:     "sub_003",
			Events: []string{"job_succeeded"},
			Queues: []string{"test_queue"},
		},
		{
			ID:     "sub_004",
			Events: []string{"job_failed"},
			Queues: []string{"different_queue"},
		},
		{
			ID:          "sub_005",
			Events:      []string{"job_failed"},
			Queues:      []string{"test_queue"},
			MinPriority: func() *int { p := 9; return &p }(),
		},
		{
			ID:          "sub_006",
			Events:      []string{"job_failed"},
			Queues:      []string{"test_queue"},
			MinPriority: func() *int { p := 7; return &p }(),
		},
	}

	t.Run("multiple matching subscriptions", func(t *testing.T) {
		matches := filter.GetMatchingSubscriptions(event, subscriptions)

		// Should match: sub_001 (exact), sub_002 (wildcard), sub_006 (priority 7 <= 8)
		// Should not match: sub_003 (wrong event), sub_004 (wrong queue), sub_005 (priority 9 > 8)
		assert.Len(t, matches, 3)

		matchedIDs := make([]string, len(matches))
		for i, match := range matches {
			matchedIDs[i] = match.ID
		}

		assert.Contains(t, matchedIDs, "sub_001")
		assert.Contains(t, matchedIDs, "sub_002")
		assert.Contains(t, matchedIDs, "sub_006")
		assert.NotContains(t, matchedIDs, "sub_003")
		assert.NotContains(t, matchedIDs, "sub_004")
		assert.NotContains(t, matchedIDs, "sub_005")
	})

	t.Run("no matching subscriptions", func(t *testing.T) {
		nonMatchingEvent := JobEvent{
			Event:    "job_enqueued",
			Queue:    "unknown_queue",
			Priority: 1,
		}

		matches := filter.GetMatchingSubscriptions(nonMatchingEvent, subscriptions)

		// Only sub_002 (wildcard) should match
		assert.Len(t, matches, 1)
		assert.Equal(t, "sub_002", matches[0].ID)
	})

	t.Run("empty subscriptions list", func(t *testing.T) {
		matches := filter.GetMatchingSubscriptions(event, []WebhookSubscription{})
		assert.Empty(t, matches)
	})
}

func TestEventFilter_FilterEventsBySubscription(t *testing.T) {
	filter := NewEventFilter()

	events := []JobEvent{
		{Event: "job_enqueued", Queue: "test_queue", Priority: 5},
		{Event: "job_started", Queue: "test_queue", Priority: 5},
		{Event: "job_succeeded", Queue: "test_queue", Priority: 5},
		{Event: "job_failed", Queue: "test_queue", Priority: 8},
		{Event: "job_dlq", Queue: "priority_queue", Priority: 9},
		{Event: "job_failed", Queue: "different_queue", Priority: 7},
	}

	t.Run("filter by specific events and queues", func(t *testing.T) {
		subscription := WebhookSubscription{
			Events: []string{"job_failed", "job_dlq"},
			Queues: []string{"test_queue", "priority_queue"},
		}

		filtered := filter.FilterEventsBySubscription(events, subscription)

		assert.Len(t, filtered, 2)
		assert.Equal(t, "job_failed", filtered[0].Event)
		assert.Equal(t, "test_queue", filtered[0].Queue)
		assert.Equal(t, "job_dlq", filtered[1].Event)
		assert.Equal(t, "priority_queue", filtered[1].Queue)
	})

	t.Run("filter with priority threshold", func(t *testing.T) {
		minPriority := 7
		subscription := WebhookSubscription{
			Events:      []string{"*"},
			Queues:      []string{"*"},
			MinPriority: &minPriority,
		}

		filtered := filter.FilterEventsBySubscription(events, subscription)

		assert.Len(t, filtered, 3)
		for _, event := range filtered {
			assert.GreaterOrEqual(t, event.Priority, 7)
		}
	})

	t.Run("wildcard filters", func(t *testing.T) {
		subscription := WebhookSubscription{
			Events: []string{"*"},
			Queues: []string{"*"},
		}

		filtered := filter.FilterEventsBySubscription(events, subscription)

		assert.Len(t, filtered, len(events))
	})

	t.Run("no matching events", func(t *testing.T) {
		subscription := WebhookSubscription{
			Events: []string{"job_retried"},
			Queues: []string{"nonexistent_queue"},
		}

		filtered := filter.FilterEventsBySubscription(events, subscription)

		assert.Empty(t, filtered)
	})
}

func TestEventFilter_ValidateSubscriptionFilters(t *testing.T) {
	filter := NewEventFilter()

	t.Run("valid subscription", func(t *testing.T) {
		subscription := WebhookSubscription{
			Events: []string{"job_failed", "job_dlq"},
			Queues: []string{"test_queue"},
		}

		err := filter.ValidateSubscriptionFilters(subscription)
		assert.NoError(t, err)
	})

	t.Run("valid subscription with wildcard", func(t *testing.T) {
		subscription := WebhookSubscription{
			Events: []string{"*"},
			Queues: []string{"*"},
		}

		err := filter.ValidateSubscriptionFilters(subscription)
		assert.NoError(t, err)
	})

	t.Run("valid subscription with priority filter", func(t *testing.T) {
		minPriority := 5
		subscription := WebhookSubscription{
			Events:      []string{"job_failed"},
			Queues:      []string{"test_queue"},
			MinPriority: &minPriority,
		}

		err := filter.ValidateSubscriptionFilters(subscription)
		assert.NoError(t, err)
	})

	t.Run("empty events list", func(t *testing.T) {
		subscription := WebhookSubscription{
			Events: []string{},
			Queues: []string{"test_queue"},
		}

		err := filter.ValidateSubscriptionFilters(subscription)
		require.Error(t, err)

		var filterErr *FilterError
		require.ErrorAs(t, err, &filterErr)
		assert.Equal(t, "validation", filterErr.Type)
		assert.Contains(t, filterErr.Message, "at least one event type must be specified")
	})

	t.Run("empty queues list", func(t *testing.T) {
		subscription := WebhookSubscription{
			Events: []string{"job_failed"},
			Queues: []string{},
		}

		err := filter.ValidateSubscriptionFilters(subscription)
		require.Error(t, err)

		var filterErr *FilterError
		require.ErrorAs(t, err, &filterErr)
		assert.Equal(t, "validation", filterErr.Type)
		assert.Contains(t, filterErr.Message, "at least one queue must be specified")
	})

	t.Run("invalid event type", func(t *testing.T) {
		subscription := WebhookSubscription{
			Events: []string{"job_failed", "invalid_event"},
			Queues: []string{"test_queue"},
		}

		err := filter.ValidateSubscriptionFilters(subscription)
		require.Error(t, err)

		var filterErr *FilterError
		require.ErrorAs(t, err, &filterErr)
		assert.Equal(t, "validation", filterErr.Type)
		assert.Contains(t, filterErr.Message, "invalid event type: invalid_event")
	})

	t.Run("invalid priority - too low", func(t *testing.T) {
		minPriority := 0
		subscription := WebhookSubscription{
			Events:      []string{"job_failed"},
			Queues:      []string{"test_queue"},
			MinPriority: &minPriority,
		}

		err := filter.ValidateSubscriptionFilters(subscription)
		require.Error(t, err)

		var filterErr *FilterError
		require.ErrorAs(t, err, &filterErr)
		assert.Equal(t, "validation", filterErr.Type)
		assert.Contains(t, filterErr.Message, "priority must be between 1 and 10")
	})

	t.Run("invalid priority - too high", func(t *testing.T) {
		minPriority := 11
		subscription := WebhookSubscription{
			Events:      []string{"job_failed"},
			Queues:      []string{"test_queue"},
			MinPriority: &minPriority,
		}

		err := filter.ValidateSubscriptionFilters(subscription)
		require.Error(t, err)

		var filterErr *FilterError
		require.ErrorAs(t, err, &filterErr)
		assert.Equal(t, "validation", filterErr.Type)
		assert.Contains(t, filterErr.Message, "priority must be between 1 and 10")
	})
}

// Edge cases and performance tests

func TestEventFilter_EdgeCases(t *testing.T) {
	filter := NewEventFilter()

	t.Run("empty event type", func(t *testing.T) {
		event := JobEvent{
			Event:    "",
			Queue:    "test_queue",
			Priority: 5,
		}

		subscription := WebhookSubscription{
			Events: []string{""},
			Queues: []string{"test_queue"},
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.True(t, matches)
	})

	t.Run("zero priority", func(t *testing.T) {
		event := JobEvent{
			Event:    "job_failed",
			Queue:    "test_queue",
			Priority: 0,
		}

		minPriority := 1
		subscription := WebhookSubscription{
			Events:      []string{"job_failed"},
			Queues:      []string{"test_queue"},
			MinPriority: &minPriority,
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.False(t, matches)
	})

	t.Run("negative priority", func(t *testing.T) {
		event := JobEvent{
			Event:    "job_failed",
			Queue:    "test_queue",
			Priority: -5,
		}

		minPriority := 1
		subscription := WebhookSubscription{
			Events:      []string{"job_failed"},
			Queues:      []string{"test_queue"},
			MinPriority: &minPriority,
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.False(t, matches)
	})

	t.Run("case sensitivity", func(t *testing.T) {
		event := JobEvent{
			Event:    "JOB_FAILED",
			Queue:    "TEST_QUEUE",
			Priority: 5,
		}

		subscription := WebhookSubscription{
			Events: []string{"job_failed"},
			Queues: []string{"test_queue"},
		}

		matches := filter.MatchesSubscription(event, subscription)
		assert.False(t, matches, "Matching should be case-sensitive")
	})
}

// Benchmark tests for performance validation

func BenchmarkEventFilter_MatchesSubscription(b *testing.B) {
	filter := NewEventFilter()
	event := JobEvent{
		Event:    "job_failed",
		Queue:    "test_queue",
		Priority: 8,
	}

	subscription := WebhookSubscription{
		Events: []string{"job_failed", "job_dlq", "job_retried"},
		Queues: []string{"test_queue", "priority_queue", "batch_queue"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.MatchesSubscription(event, subscription)
	}
}

func BenchmarkEventFilter_GetMatchingSubscriptions(b *testing.B) {
	filter := NewEventFilter()
	event := JobEvent{
		Event:    "job_failed",
		Queue:    "test_queue",
		Priority: 8,
	}

	// Create 100 subscriptions
	subscriptions := make([]WebhookSubscription, 100)
	for i := 0; i < 100; i++ {
		subscriptions[i] = WebhookSubscription{
			ID:     "sub_" + string(rune(i)),
			Events: []string{"job_failed"},
			Queues: []string{"test_queue"},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.GetMatchingSubscriptions(event, subscriptions)
	}
}

func BenchmarkEventFilter_FilterEventsBySubscription(b *testing.B) {
	filter := NewEventFilter()

	// Create 1000 events
	events := make([]JobEvent, 1000)
	eventTypes := []string{"job_enqueued", "job_started", "job_succeeded", "job_failed", "job_dlq"}
	queues := []string{"test_queue", "priority_queue", "batch_queue"}

	for i := 0; i < 1000; i++ {
		events[i] = JobEvent{
			Event:    eventTypes[i%len(eventTypes)],
			Queue:    queues[i%len(queues)],
			Priority: (i%10 + 1),
		}
	}

	subscription := WebhookSubscription{
		Events: []string{"job_failed", "job_dlq"},
		Queues: []string{"test_queue", "priority_queue"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filter.FilterEventsBySubscription(events, subscription)
	}
}