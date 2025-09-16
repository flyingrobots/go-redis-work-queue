// Copyright 2025 James Ross
package fixtures

import (
	"encoding/json"
	"fmt"
	"time"
)

// TestJobEvent represents a job event for testing
type TestJobEvent struct {
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

// TestWebhookSubscription represents a webhook subscription for testing
type TestWebhookSubscription struct {
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

// TestRetryPolicy represents retry policy configuration for testing
type TestRetryPolicy struct {
	Strategy     string        `json:"strategy"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	Multiplier   float64       `json:"multiplier"`
	MaxRetries   int           `json:"max_retries"`
	Jitter       bool          `json:"jitter"`
}

// Event generators
func NewTestJobEnqueuedEvent() TestJobEvent {
	return TestJobEvent{
		Event:     "job_enqueued",
		Timestamp: time.Now().UTC(),
		JobID:     "job_12345",
		Queue:     "test_queue",
		Priority:  5,
		Attempt:   1,
		TraceID:   "trace_abc123",
		RequestID: "req_xyz789",
		UserID:    "user_456",
		Payload: map[string]interface{}{
			"task_type": "process_file",
			"file_path": "/tmp/test.txt",
			"metadata": map[string]interface{}{
				"size":        1024,
				"content_type": "text/plain",
			},
		},
	}
}

func NewTestJobStartedEvent() TestJobEvent {
	return TestJobEvent{
		Event:     "job_started",
		Timestamp: time.Now().UTC(),
		JobID:     "job_12345",
		Queue:     "test_queue",
		Priority:  5,
		Attempt:   1,
		Worker:    "worker_001",
		TraceID:   "trace_abc123",
		RequestID: "req_xyz789",
		UserID:    "user_456",
	}
}

func NewTestJobSucceededEvent() TestJobEvent {
	duration := 5 * time.Second
	return TestJobEvent{
		Event:     "job_succeeded",
		Timestamp: time.Now().UTC(),
		JobID:     "job_12345",
		Queue:     "test_queue",
		Priority:  5,
		Attempt:   1,
		Worker:    "worker_001",
		Duration:  &duration,
		TraceID:   "trace_abc123",
		RequestID: "req_xyz789",
		UserID:    "user_456",
	}
}

func NewTestJobFailedEvent() TestJobEvent {
	duration := 2 * time.Second
	return TestJobEvent{
		Event:     "job_failed",
		Timestamp: time.Now().UTC(),
		JobID:     "job_12345",
		Queue:     "test_queue",
		Priority:  5,
		Attempt:   1,
		Error:     "Connection timeout to external service",
		Worker:    "worker_001",
		Duration:  &duration,
		TraceID:   "trace_abc123",
		RequestID: "req_xyz789",
		UserID:    "user_456",
	}
}

func NewTestJobDLQEvent() TestJobEvent {
	return TestJobEvent{
		Event:     "job_dlq",
		Timestamp: time.Now().UTC(),
		JobID:     "job_12345",
		Queue:     "test_queue",
		Priority:  5,
		Attempt:   5,
		Error:     "Maximum retries exceeded",
		TraceID:   "trace_abc123",
		RequestID: "req_xyz789",
		UserID:    "user_456",
	}
}

func NewTestJobRetriedEvent() TestJobEvent {
	return TestJobEvent{
		Event:     "job_retried",
		Timestamp: time.Now().UTC(),
		JobID:     "job_12345",
		Queue:     "test_queue",
		Priority:  5,
		Attempt:   2,
		Error:     "Previous attempt failed, retrying",
		TraceID:   "trace_abc123",
		RequestID: "req_xyz789",
		UserID:    "user_456",
	}
}

// Subscription generators
func NewTestWebhookSubscription() TestWebhookSubscription {
	return TestWebhookSubscription{
		ID:             "sub_001",
		Name:           "Test Webhook",
		URL:            "https://example.com/webhook",
		Secret:         "test_secret_key_123",
		Events:         []string{"job_failed", "job_dlq"},
		Queues:         []string{"test_queue", "priority_queue"},
		MaxRetries:     5,
		Timeout:        30 * time.Second,
		RateLimit:      100,
		Headers:        map[string]string{"Content-Type": "application/json"},
		IncludePayload: true,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		Disabled:       false,
	}
}

func NewTestWebhookSubscriptionWithFilters() TestWebhookSubscription {
	minPriority := 8
	return TestWebhookSubscription{
		ID:            "sub_002",
		Name:          "High Priority Alerts",
		URL:           "https://alerts.example.com/webhook",
		Secret:        "alert_secret_456",
		Events:        []string{"job_failed", "job_dlq"},
		Queues:        []string{"*"}, // All queues
		MinPriority:   &minPriority,
		MaxRetries:    3,
		Timeout:       15 * time.Second,
		RateLimit:     50,
		PayloadFields: []string{"error", "queue", "priority"},
		RedactFields:  []string{"user_id", "sensitive_data"},
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
		Disabled:      false,
	}
}

func NewTestWebhookSubscriptionAllEvents() TestWebhookSubscription {
	return TestWebhookSubscription{
		ID:             "sub_003",
		Name:           "Analytics Webhook",
		URL:            "https://analytics.example.com/events",
		Secret:         "analytics_secret_789",
		Events:         []string{"*"}, // All events
		Queues:         []string{"*"}, // All queues
		MaxRetries:     10,
		Timeout:        60 * time.Second,
		RateLimit:      1000,
		IncludePayload: true,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
		Disabled:       false,
	}
}

// Retry policy generators
func NewTestExponentialRetryPolicy() TestRetryPolicy {
	return TestRetryPolicy{
		Strategy:     "exponential",
		InitialDelay: 1 * time.Second,
		MaxDelay:     5 * time.Minute,
		Multiplier:   2.0,
		MaxRetries:   5,
		Jitter:       true,
	}
}

func NewTestLinearRetryPolicy() TestRetryPolicy {
	return TestRetryPolicy{
		Strategy:     "linear",
		InitialDelay: 2 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   1.0,
		MaxRetries:   3,
		Jitter:       false,
	}
}

func NewTestFixedRetryPolicy() TestRetryPolicy {
	return TestRetryPolicy{
		Strategy:     "fixed",
		InitialDelay: 5 * time.Second,
		MaxDelay:     5 * time.Second,
		Multiplier:   1.0,
		MaxRetries:   5,
		Jitter:       false,
	}
}

// Mock data generators
func GenerateJobEvents(count int) []TestJobEvent {
	events := make([]TestJobEvent, count)
	eventTypes := []string{"job_enqueued", "job_started", "job_succeeded", "job_failed", "job_dlq", "job_retried"}
	queues := []string{"test_queue", "priority_queue", "batch_queue"}

	for i := 0; i < count; i++ {
		eventType := eventTypes[i%len(eventTypes)]
		queue := queues[i%len(queues)]

		events[i] = TestJobEvent{
			Event:     eventType,
			Timestamp: time.Now().UTC().Add(-time.Duration(i) * time.Minute),
			JobID:     fmt.Sprintf("job_%d", i),
			Queue:     queue,
			Priority:  (i%10 + 1),
			Attempt:   (i%3 + 1),
			TraceID:   fmt.Sprintf("trace_%d", i),
			RequestID: fmt.Sprintf("req_%d", i),
			UserID:    fmt.Sprintf("user_%d", i%10),
		}

		// Add event-specific fields
		if eventType == "job_failed" || eventType == "job_dlq" {
			events[i].Error = fmt.Sprintf("Test error for job %d", i)
		}

		if eventType == "job_succeeded" || eventType == "job_failed" {
			duration := time.Duration(i%60+1) * time.Second
			events[i].Duration = &duration
		}

		if eventType == "job_started" || eventType == "job_succeeded" || eventType == "job_failed" {
			events[i].Worker = fmt.Sprintf("worker_%d", i%5)
		}

		// Add payload for some events
		if i%3 == 0 {
			events[i].Payload = map[string]interface{}{
				"task_id": fmt.Sprintf("task_%d", i),
				"data": map[string]interface{}{
					"value": i * 10,
					"type":  "test_data",
				},
			}
		}
	}

	return events
}

func GenerateWebhookSubscriptions(count int) []TestWebhookSubscription {
	subscriptions := make([]TestWebhookSubscription, count)

	for i := 0; i < count; i++ {
		subscriptions[i] = TestWebhookSubscription{
			ID:         fmt.Sprintf("sub_%03d", i),
			Name:       fmt.Sprintf("Test Webhook %d", i),
			URL:        fmt.Sprintf("https://webhook%d.example.com/endpoint", i),
			Secret:     fmt.Sprintf("secret_%d", i),
			Events:     []string{"job_failed", "job_dlq"},
			Queues:     []string{fmt.Sprintf("queue_%d", i%3)},
			MaxRetries: 5,
			Timeout:    30 * time.Second,
			RateLimit:  100,
			CreatedAt:  time.Now().UTC().Add(-time.Duration(i) * time.Hour),
			UpdatedAt:  time.Now().UTC().Add(-time.Duration(i%10) * time.Minute),
			Disabled:   i%10 == 0, // 10% disabled
		}
	}

	return subscriptions
}

// Helper functions
func (e TestJobEvent) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (s TestWebhookSubscription) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

func (p TestRetryPolicy) ToJSON() ([]byte, error) {
	return json.Marshal(p)
}

// Event matching helpers for testing filters
func (e TestJobEvent) MatchesSubscription(sub TestWebhookSubscription) bool {
	// Check event type filter
	if !contains(sub.Events, e.Event) && !contains(sub.Events, "*") {
		return false
	}

	// Check queue filter
	if !contains(sub.Queues, e.Queue) && !contains(sub.Queues, "*") {
		return false
	}

	// Check priority filter
	if sub.MinPriority != nil && e.Priority < *sub.MinPriority {
		return false
	}

	return true
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}