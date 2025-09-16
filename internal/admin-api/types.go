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
	Count    int    `json:"count" validate:"required,min=1,max=10000"`
	Priority string `json:"priority" validate:"required,oneof=high low"`
	Rate     int    `json:"rate" validate:"min=1,max=1000"`
	Timeout  int    `json:"timeout_seconds" validate:"min=1,max=300"`
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