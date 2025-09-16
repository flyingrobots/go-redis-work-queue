// Copyright 2025 James Ross
package exactlyonce

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"time"
)

// IdempotencyKey represents a unique identifier for ensuring exactly-once processing
type IdempotencyKey struct {
	ID        string    `json:"id"`
	QueueName string    `json:"queue_name"`
	TenantID  string    `json:"tenant_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	TTL       time.Duration `json:"ttl"`
}

// DedupEntry represents an entry in the deduplication storage
type DedupEntry struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	QueueName string    `json:"queue_name"`
	TenantID  string    `json:"tenant_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// OutboxEvent represents an event that needs to be published from the outbox pattern
type OutboxEvent struct {
	ID             string                 `json:"id"`
	AggregateID    string                 `json:"aggregate_id"`
	AggregateType  string                 `json:"aggregate_type,omitempty"`
	EventType      string                 `json:"event_type"`
	Payload        json.RawMessage        `json:"payload"`
	Headers        map[string]string      `json:"headers,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	PublishedAt    *time.Time             `json:"published_at,omitempty"`
	Retries        int                    `json:"retries"`
	RetryCount     int                    `json:"retry_count"`
	MaxRetries     int                    `json:"max_retries"`
	NextRetryAt    *time.Time             `json:"next_retry_at,omitempty"`
	LastError      string                 `json:"last_error,omitempty"`
	LastAttemptAt  *time.Time             `json:"last_attempt_at,omitempty"`
	Status         string                 `json:"status,omitempty"`
}

// DedupStats represents statistics about deduplication
type DedupStats struct {
	QueueName        string    `json:"queue_name"`
	TenantID         string    `json:"tenant_id,omitempty"`
	TotalKeys        int64     `json:"total_keys"`
	HitRate          float64   `json:"hit_rate"`
	TotalRequests    int64     `json:"total_requests"`
	DuplicatesAvoided int64    `json:"duplicates_avoided"`
	LastUpdated      time.Time `json:"last_updated"`
}

// IdempotencyResult indicates the result of an idempotency check
type IdempotencyResult struct {
	IsFirstTime bool        `json:"is_first_time"`
	ExistingValue interface{} `json:"existing_value,omitempty"`
	Key         string      `json:"key"`
}

// ProcessingStatus represents the status of job processing
type ProcessingStatus int

const (
	StatusPending ProcessingStatus = iota
	StatusProcessing
	StatusCompleted
	StatusFailed
)

func (s ProcessingStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusProcessing:
		return "processing"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// Value implements driver.Valuer for database compatibility
func (s ProcessingStatus) Value() (driver.Value, error) {
	return int64(s), nil
}

// IdempotencyStorage defines the interface for idempotency storage operations
type IdempotencyStorage interface {
	// Check verifies if a key has been processed before
	Check(ctx context.Context, key IdempotencyKey) (*IdempotencyResult, error)

	// Set marks a key as processed with optional result value
	Set(ctx context.Context, key IdempotencyKey, value interface{}) error

	// Delete removes a key from the idempotency store
	Delete(ctx context.Context, key IdempotencyKey) error

	// Stats returns statistics about the deduplication store
	Stats(ctx context.Context, queueName, tenantID string) (*DedupStats, error)
}


// ProcessingHook defines hooks that can be called during processing
type ProcessingHook interface {
	// BeforeProcessing is called before job processing starts
	BeforeProcessing(ctx context.Context, jobID string, idempotencyKey IdempotencyKey) error

	// AfterProcessing is called after job processing completes
	AfterProcessing(ctx context.Context, jobID string, result interface{}, err error) error

	// OnDuplicate is called when a duplicate job is detected
	OnDuplicate(ctx context.Context, jobID string, existingResult interface{}) error
}