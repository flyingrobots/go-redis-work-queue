// Copyright 2025 James Ross
package eventhooks

import (
	"sync"
	"time"
)

// EventType represents the type of job lifecycle event
type EventType string

const (
	EventJobEnqueued  EventType = "job_enqueued"
	EventJobStarted   EventType = "job_started"
	EventJobSucceeded EventType = "job_succeeded"
	EventJobFailed    EventType = "job_failed"
	EventJobDLQ       EventType = "job_dlq"
	EventJobRetried   EventType = "job_retried"
)

// JobEvent represents a job lifecycle event with all necessary context
type JobEvent struct {
	Event       EventType  `json:"event"`
	Timestamp   time.Time  `json:"timestamp"`
	JobID       string     `json:"job_id"`
	Queue       string     `json:"queue"`
	Priority    int        `json:"priority"`
	Attempt     int        `json:"attempt"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`

	// Event-specific fields
	Error    string        `json:"error,omitempty"`
	Duration *time.Duration `json:"duration,omitempty"`
	Worker   string        `json:"worker,omitempty"`

	// Optional payload preview (truncated for webhooks)
	Payload interface{} `json:"payload,omitempty"`

	// Correlation tracking
	TraceID   string `json:"trace_id,omitempty"`
	RequestID string `json:"request_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`

	// Deep links for TUI integration
	Links map[string]string `json:"_links,omitempty"`
}

// WebhookSubscription defines a webhook endpoint configuration
type WebhookSubscription struct {
	ID   string `json:"id" redis:"id"`
	Name string `json:"name" redis:"name"`
	URL  string `json:"url" redis:"url"`
	// Secret is never returned in JSON responses
	Secret string `json:"-" redis:"secret"`

	// Filtering rules
	Events      []EventType `json:"events" redis:"events"`
	Queues      []string    `json:"queues" redis:"queues"`
	MinPriority *int        `json:"min_priority,omitempty" redis:"min_priority"`

	// Delivery configuration
	MaxRetries int           `json:"max_retries" redis:"max_retries"`
	Timeout    time.Duration `json:"timeout" redis:"timeout"`
	RateLimit  int           `json:"rate_limit" redis:"rate_limit"`
	Headers    []HeaderPair  `json:"headers" redis:"headers"`

	// Payload configuration
	IncludePayload bool     `json:"include_payload" redis:"include_payload"`
	PayloadFields  []string `json:"payload_fields,omitempty" redis:"payload_fields"`
	RedactFields   []string `json:"redact_fields,omitempty" redis:"redact_fields"`

	// Status tracking
	CreatedAt    time.Time  `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" redis:"updated_at"`
	LastSuccess  *time.Time `json:"last_success,omitempty" redis:"last_success"`
	LastFailure  *time.Time `json:"last_failure,omitempty" redis:"last_failure"`
	FailureCount int        `json:"failure_count" redis:"failure_count"`
	Disabled     bool       `json:"disabled" redis:"disabled"`

	// Runtime state
	mu sync.RWMutex `json:"-" redis:"-"`
}

// HeaderPair represents a custom HTTP header key-value pair
type HeaderPair struct {
	Key   string `json:"key" redis:"key"`
	Value string `json:"value" redis:"value"`
}

// NATSSubscription defines NATS publishing configuration
type NATSSubscription struct {
	ID      string            `json:"id" redis:"id"`
	Name    string            `json:"name" redis:"name"`
	Subject string            `json:"subject" redis:"subject"`
	Events  []EventType       `json:"events" redis:"events"`
	Queues  []string          `json:"queues" redis:"queues"`
	Headers map[string]string `json:"headers" redis:"headers"`

	CreatedAt time.Time `json:"created_at" redis:"created_at"`
	UpdatedAt time.Time `json:"updated_at" redis:"updated_at"`
	Disabled  bool      `json:"disabled" redis:"disabled"`
}

// RetryPolicy defines how delivery failures are retried
type RetryPolicy struct {
	Strategy     string        `json:"strategy" redis:"strategy"`
	InitialDelay time.Duration `json:"initial_delay" redis:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay" redis:"max_delay"`
	Multiplier   float64       `json:"multiplier" redis:"multiplier"`
	MaxRetries   int           `json:"max_retries" redis:"max_retries"`
	Jitter       bool          `json:"jitter" redis:"jitter"`
}

// DeliveryAttempt represents a single webhook delivery attempt
type DeliveryAttempt struct {
	ID             string    `json:"id" redis:"id"`
	SubscriptionID string    `json:"subscription_id" redis:"subscription_id"`
	Event          JobEvent  `json:"event" redis:"event"`
	AttemptNumber  int       `json:"attempt_number" redis:"attempt_number"`
	ScheduledAt    time.Time `json:"scheduled_at" redis:"scheduled_at"`
	AttemptedAt    *time.Time `json:"attempted_at,omitempty" redis:"attempted_at"`

	// Result tracking
	Success      bool   `json:"success" redis:"success"`
	StatusCode   int    `json:"status_code" redis:"status_code"`
	ErrorMessage string `json:"error_message" redis:"error_message"`
	ResponseTime time.Duration `json:"response_time" redis:"response_time"`

	// Metadata
	DeliveryID   string            `json:"delivery_id" redis:"delivery_id"`
	RequestURL   string            `json:"request_url" redis:"request_url"`
	RequestHeaders map[string]string `json:"request_headers" redis:"request_headers"`
	ResponseHeaders map[string]string `json:"response_headers,omitempty" redis:"response_headers"`
}

// DeadLetterHook represents a failed webhook delivery in the DLH queue
type DeadLetterHook struct {
	ID             string           `json:"id" redis:"id"`
	SubscriptionID string           `json:"subscription_id" redis:"subscription_id"`
	Event          JobEvent         `json:"event" redis:"event"`
	Attempts       []DeliveryAttempt `json:"attempts" redis:"attempts"`
	FinalError     string           `json:"final_error" redis:"final_error"`
	CreatedAt      time.Time        `json:"created_at" redis:"created_at"`

	// Replay tracking
	Replayed   bool       `json:"replayed" redis:"replayed"`
	ReplayedAt *time.Time `json:"replayed_at,omitempty" redis:"replayed_at"`
	ReplayedBy string     `json:"replayed_by,omitempty" redis:"replayed_by"`
}

// EventMetrics tracks performance and health metrics
type EventMetrics struct {
	EventsEmitted       int64             `json:"events_emitted"`
	WebhookDeliveries   int64             `json:"webhook_deliveries"`
	WebhookFailures     int64             `json:"webhook_failures"`
	RetryAttempts       int64             `json:"retry_attempts"`
	DLHSize            int64             `json:"dlh_size"`
	DeliveryLatencyP95  time.Duration     `json:"delivery_latency_p95"`
	SubscriptionHealth  map[string]float64 `json:"subscription_health"`

	// Rate limiting metrics
	RateLimitViolations int64 `json:"rate_limit_violations"`
	CircuitBreakerTrips int64 `json:"circuit_breaker_trips"`
}

// SubscriptionHealthStatus represents the health of a webhook subscription
type SubscriptionHealthStatus struct {
	SubscriptionID   string        `json:"subscription_id"`
	SuccessRate      float64       `json:"success_rate"`
	LastDelivery     *time.Time    `json:"last_delivery,omitempty"`
	LastSuccess      *time.Time    `json:"last_success,omitempty"`
	LastFailure      *time.Time    `json:"last_failure,omitempty"`
	ConsecutiveFailures int        `json:"consecutive_failures"`
	AverageLatency   time.Duration `json:"average_latency"`
	TotalDeliveries  int64         `json:"total_deliveries"`
	TotalFailures    int64         `json:"total_failures"`
}

// EventFilter provides methods for filtering events
type EventFilter struct {
	Events      []EventType `json:"events"`
	Queues      []string    `json:"queues"`
	MinPriority *int        `json:"min_priority,omitempty"`
	MaxPriority *int        `json:"max_priority,omitempty"`
}

// Matches checks if an event matches the filter criteria
func (f *EventFilter) Matches(event JobEvent) bool {
	// Check event type filter
	if len(f.Events) > 0 {
		found := false
		for _, eventType := range f.Events {
			if eventType == event.Event {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check queue filter
	if len(f.Queues) > 0 {
		found := false
		for _, queue := range f.Queues {
			if queue == "*" || queue == event.Queue {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check priority filters
	if f.MinPriority != nil && event.Priority < *f.MinPriority {
		return false
	}
	if f.MaxPriority != nil && event.Priority > *f.MaxPriority {
		return false
	}

	return true
}

// DefaultRetryPolicy returns the default exponential backoff retry policy
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		Strategy:     "exponential",
		InitialDelay: 1 * time.Second,
		MaxDelay:     5 * time.Minute,
		Multiplier:   2.0,
		MaxRetries:   5,
		Jitter:       true,
	}
}

// EventSubscriber interface for different transport mechanisms
type EventSubscriber interface {
	ID() string
	Name() string
	ProcessEvent(event JobEvent) error
	IsHealthy() bool
	GetFilter() EventFilter
	Close() error
}

// EventBusConfig configures the event bus behavior
type EventBusConfig struct {
	BufferSize       int           `json:"buffer_size"`
	WorkerPoolSize   int           `json:"worker_pool_size"`
	DefaultTimeout   time.Duration `json:"default_timeout"`
	MetricsInterval  time.Duration `json:"metrics_interval"`
	EnablePersistence bool         `json:"enable_persistence"`
	MaxRetryDelay    time.Duration `json:"max_retry_delay"`
}

// DefaultEventBusConfig returns sensible defaults for the event bus
func DefaultEventBusConfig() EventBusConfig {
	return EventBusConfig{
		BufferSize:       10000,
		WorkerPoolSize:   10,
		DefaultTimeout:   30 * time.Second,
		MetricsInterval:  60 * time.Second,
		EnablePersistence: false,
		MaxRetryDelay:    5 * time.Minute,
	}
}