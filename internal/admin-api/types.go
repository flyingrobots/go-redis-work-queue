// Copyright 2025 James Ross
package adminapi

import (
	"time"
)

// Request types

type PeekRequest struct {
	Count int `json:"count" validate:"min=1,max=100"`
}

type BenchRequest struct {
	Count       int    `json:"count" validate:"required,min=1,max=10000"`
	Priority    string `json:"priority" validate:"required,oneof=high low"`
	Rate        int    `json:"rate" validate:"min=1,max=1000"`
	Timeout     int    `json:"timeout_seconds" validate:"min=1,max=300"`
	PayloadSize int    `json:"payload_size_bytes" validate:"min=0,max=1048576"`
}

type PurgeRequest struct {
	Confirmation string `json:"confirmation" validate:"required"`
	Reason       string `json:"reason" validate:"required,min=3,max=500"`
}

// Response types

type ErrorResponse struct {
	Error   string            `json:"error"`
	Code    string            `json:"code,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

type StatsResponse struct {
	Queues          map[string]int64 `json:"queues"`
	ProcessingLists map[string]int64 `json:"processing_lists"`
	Heartbeats      int64            `json:"heartbeats"`
	Timestamp       time.Time        `json:"timestamp"`
}

type StatsKeysResponse struct {
	QueueLengths    map[string]int64 `json:"queue_lengths"`
	ProcessingLists int64            `json:"processing_lists"`
	ProcessingItems int64            `json:"processing_items"`
	Heartbeats      int64            `json:"heartbeats"`
	RateLimitKey    string           `json:"rate_limit_key"`
	RateLimitTTL    string           `json:"rate_limit_ttl,omitempty"`
	Timestamp       time.Time        `json:"timestamp"`
}

type PeekResponse struct {
	Queue     string    `json:"queue"`
	Items     []string  `json:"items"`
	Count     int       `json:"count"`
	Timestamp time.Time `json:"timestamp"`
}

type BenchResponse struct {
	Count      int           `json:"count"`
	Duration   time.Duration `json:"duration"`
	Throughput float64       `json:"throughput_jobs_per_sec"`
	P50        time.Duration `json:"p50_latency"`
	P95        time.Duration `json:"p95_latency"`
	Timestamp  time.Time     `json:"timestamp"`
}

type PurgeResponse struct {
	Success      bool      `json:"success"`
	ItemsDeleted int64     `json:"items_deleted,omitempty"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
}

// DLQ types
type DLQItem struct {
	ID        string    `json:"id"`
	Queue     string    `json:"queue,omitempty"`
	Payload   string    `json:"payload"`
	Reason    string    `json:"reason,omitempty"`
	Attempts  int       `json:"attempts,omitempty"`
	FirstSeen time.Time `json:"first_seen,omitempty"`
	LastSeen  time.Time `json:"last_seen,omitempty"`
}

type DLQListResponse struct {
	Items      []DLQItem `json:"items"`
	NextCursor string    `json:"next_cursor,omitempty"`
	Count      int       `json:"count"`
	Timestamp  time.Time `json:"timestamp"`
}

type DLQRequeueRequest struct {
	Namespace string   `json:"ns"`
	IDs       []string `json:"ids"`
	DestQueue string   `json:"dest_queue,omitempty"`
}

type DLQRequeueResponse struct {
	Requeued  int       `json:"requeued"`
	Timestamp time.Time `json:"timestamp"`
}

type DLQPurgeSelectionRequest struct {
	Namespace string   `json:"ns"`
	IDs       []string `json:"ids"`
}

type DLQPurgeSelectionResponse struct {
	Purged    int       `json:"purged"`
	Timestamp time.Time `json:"timestamp"`
}

// Workers types
type WorkerInfo struct {
	ID            string     `json:"id"`
	LastHeartbeat time.Time  `json:"last_heartbeat"`
	Queue         string     `json:"queue,omitempty"`
	JobID         string     `json:"job_id,omitempty"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	Version       string     `json:"version,omitempty"`
	Host          string     `json:"host,omitempty"`
}

type WorkersResponse struct {
	Workers   []WorkerInfo `json:"workers"`
	Timestamp time.Time    `json:"timestamp"`
}

// Audit log entry
type AuditEntry struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	User      string                 `json:"user"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Result    string                 `json:"result"`
	Reason    string                 `json:"reason,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	IP        string                 `json:"ip"`
	UserAgent string                 `json:"user_agent"`
}

// JWT claims
type Claims struct {
	Subject   string   `json:"sub"`
	Roles     []string `json:"roles"`
	ExpiresAt int64    `json:"exp"`
	IssuedAt  int64    `json:"iat"`
}

// Rate limit info
type RateLimitInfo struct {
	Limit     int
	Remaining int
	ResetAt   time.Time
}
