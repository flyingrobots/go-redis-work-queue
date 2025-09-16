// Copyright 2025 James Ross
package multitenantiso

import (
	"encoding/json"
	"regexp"
	"time"
)

// TenantID represents a unique tenant identifier with validation
type TenantID string

const (
	MaxTenantIDLength = 32
	MinTenantIDLength = 3
)

var tenantIDRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

// Validate checks if the tenant ID follows the required format
func (t TenantID) Validate() error {
	if len(t) < MinTenantIDLength || len(t) > MaxTenantIDLength {
		return ErrInvalidTenantIDLength
	}

	// Only lowercase alphanumeric and hyphens
	if !tenantIDRegex.MatchString(string(t)) {
		return ErrInvalidTenantIDFormat
	}

	// Must start and end with alphanumeric
	if t[0] == '-' || t[len(t)-1] == '-' {
		return ErrTenantIDMustNotStartOrEndWithHyphen
	}

	return nil
}

// String returns the string representation of the tenant ID
func (t TenantID) String() string {
	return string(t)
}

// TenantConfig represents configuration for a tenant
type TenantConfig struct {
	ID                 TenantID            `json:"id"`
	Name               string              `json:"name"`
	Status             TenantStatus        `json:"status"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	Quotas             TenantQuotas        `json:"quotas"`
	Encryption         TenantEncryption    `json:"encryption"`
	RateLimiting       TenantRateLimiting  `json:"rate_limiting"`
	Metadata           map[string]string   `json:"metadata,omitempty"`
	ContactEmail       string              `json:"contact_email,omitempty"`
	BillingReference   string              `json:"billing_reference,omitempty"`
}

// TenantStatus represents the current status of a tenant
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusWarning   TenantStatus = "warning"
	TenantStatusDeleted   TenantStatus = "deleted"
)

// TenantQuotas defines resource limits for a tenant
type TenantQuotas struct {
	// Job limits
	MaxJobsPerHour     int64 `json:"max_jobs_per_hour"`
	MaxJobsPerDay      int64 `json:"max_jobs_per_day"`
	MaxBacklogSize     int64 `json:"max_backlog_size"`
	MaxJobSizeBytes    int64 `json:"max_job_size_bytes"`

	// Resource limits
	MaxQueuesPerTenant int32 `json:"max_queues_per_tenant"`
	MaxWorkersPerQueue int32 `json:"max_workers_per_queue"`
	MaxStorageBytes    int64 `json:"max_storage_bytes"`

	// Rate limits
	EnqueueRateLimit   int32 `json:"enqueue_rate_limit"`   // per second
	DequeueRateLimit   int32 `json:"dequeue_rate_limit"`   // per second

	// Soft limits (warnings)
	SoftLimitThreshold float64 `json:"soft_limit_threshold"` // 0.8 = 80%
}

// DefaultQuotas returns reasonable default quotas for a new tenant
func DefaultQuotas() TenantQuotas {
	return TenantQuotas{
		MaxJobsPerHour:     10000,
		MaxJobsPerDay:      100000,
		MaxBacklogSize:     50000,
		MaxJobSizeBytes:    1024 * 1024, // 1MB
		MaxQueuesPerTenant: 10,
		MaxWorkersPerQueue: 50,
		MaxStorageBytes:    100 * 1024 * 1024, // 100MB
		EnqueueRateLimit:   100,
		DequeueRateLimit:   100,
		SoftLimitThreshold: 0.8,
	}
}

// QuotaUsage tracks current resource usage for a tenant
type QuotaUsage struct {
	TenantID           TenantID  `json:"tenant_id"`
	JobsThisHour       int64     `json:"jobs_this_hour"`
	JobsThisDay        int64     `json:"jobs_this_day"`
	CurrentBacklogSize int64     `json:"current_backlog_size"`
	StorageUsedBytes   int64     `json:"storage_used_bytes"`
	ActiveQueues       int32     `json:"active_queues"`
	ActiveWorkers      int32     `json:"active_workers"`
	LastUpdated        time.Time `json:"last_updated"`
}

// TenantEncryption configures per-tenant encryption settings
type TenantEncryption struct {
	Enabled           bool          `json:"enabled"`
	KEKProvider       string        `json:"kek_provider"`       // "aws-kms", "gcp-kms", "azure-kv", "local"
	KEKKeyID          string        `json:"kek_key_id"`         // Cloud KMS key identifier
	DEKRotationPeriod time.Duration `json:"dek_rotation_period"` // How often to rotate data keys
	Algorithm         string        `json:"algorithm"`          // "AES-256-GCM"
	LastRotation      time.Time     `json:"last_rotation"`
}

// TenantRateLimiting configures rate limiting for a tenant
type TenantRateLimiting struct {
	Enabled              bool                 `json:"enabled"`
	WindowDuration       time.Duration        `json:"window_duration"`
	BurstCapacity        int32               `json:"burst_capacity"`
	CustomLimits         map[string]int32    `json:"custom_limits,omitempty"`
	EnforceAcrossWorkers bool                `json:"enforce_across_workers"`
}

// EncryptedPayload represents an encrypted job payload
type EncryptedPayload struct {
	Version          int    `json:"v"`                    // Encryption version
	EncryptedDEK     []byte `json:"encrypted_dek"`        // DEK encrypted by KEK
	EncryptedPayload []byte `json:"encrypted_payload"`    // Actual job data
	Nonce            []byte `json:"nonce"`               // AES-GCM nonce
	AuthTag          []byte `json:"auth_tag"`            // Authentication tag
	CreatedAt        int64  `json:"created_at"`          // Unix timestamp
}

// TenantPermission defines what actions a user can perform on tenant resources
type TenantPermission struct {
	TenantID    TenantID `json:"tenant_id"`
	Resource    string   `json:"resource"`    // "queues", "workers", "metrics", "config"
	Actions     []string `json:"actions"`     // "read", "write", "admin"
	QueueFilter string   `json:"queue_filter,omitempty"` // Optional queue name pattern
}

// UserTenantAccess defines a user's access permissions across tenants
type UserTenantAccess struct {
	UserID      string             `json:"user_id"`
	Permissions []TenantPermission `json:"permissions"`
	CreatedAt   time.Time          `json:"created_at"`
	ExpiresAt   *time.Time         `json:"expires_at,omitempty"`
}

// AuditEvent represents a logged tenant operation
type AuditEvent struct {
	EventID    string                 `json:"event_id"`    // UUID
	Timestamp  time.Time              `json:"timestamp"`
	TenantID   TenantID               `json:"tenant_id"`
	UserID     string                 `json:"user_id"`     // API key or user ID
	Action     string                 `json:"action"`      // ENQUEUE, DEQUEUE, CREATE, DELETE, etc.
	Resource   string                 `json:"resource"`    // t:tenant:queue, tenant:config, etc.
	Details    map[string]interface{} `json:"details"`     // Action-specific data
	RemoteIP   string                 `json:"remote_ip"`
	UserAgent  string                 `json:"user_agent"`
	Result     string                 `json:"result"`      // SUCCESS, DENIED, ERROR
	ErrorCode  string                 `json:"error_code,omitempty"`
}

// AuditQuery defines parameters for querying audit events
type AuditQuery struct {
	TenantID   *TenantID  `json:"tenant_id,omitempty"`
	UserID     *string    `json:"user_id,omitempty"`
	Actions    []string   `json:"actions,omitempty"`
	StartTime  *time.Time `json:"start_time,omitempty"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	Result     *string    `json:"result,omitempty"`
	Limit      int        `json:"limit,omitempty"`
	Offset     int        `json:"offset,omitempty"`
}

// TenantSummary provides overview information about a tenant
type TenantSummary struct {
	ID           TenantID `json:"id"`
	Name         string   `json:"name"`
	Status       string   `json:"status"`      // active, suspended, warning
	ActiveQueues int      `json:"active_queues"`
	QuotaHealth  string   `json:"quota_health"` // good, warning, critical
	LastActivity time.Time `json:"last_activity"`
}

// TenantContext represents the current tenant context in TUI or API
type TenantContext struct {
	CurrentTenant    *TenantID       `json:"current_tenant,omitempty"`
	AvailableTenants []TenantSummary `json:"available_tenants"`
	LastSwitched     time.Time       `json:"last_switched"`
}

// KeyNamespace defines Redis key patterns for tenant isolation
type KeyNamespace struct {
	TenantID TenantID
}

// QueueKey returns the Redis key for a tenant's queue
func (kn KeyNamespace) QueueKey(queueName string) string {
	return "t:" + string(kn.TenantID) + ":" + queueName
}

// JobsKey returns the Redis key for a tenant's job list
func (kn KeyNamespace) JobsKey(queueName string) string {
	return "t:" + string(kn.TenantID) + ":" + queueName + ":jobs"
}

// DLQKey returns the Redis key for a tenant's dead letter queue
func (kn KeyNamespace) DLQKey(queueName string) string {
	return "t:" + string(kn.TenantID) + ":" + queueName + ":dlq"
}

// WorkersKey returns the Redis key for a tenant's worker registry
func (kn KeyNamespace) WorkersKey(queueName string) string {
	return "t:" + string(kn.TenantID) + ":" + queueName + ":workers"
}

// MetricsKey returns the Redis key for a tenant's queue metrics
func (kn KeyNamespace) MetricsKey(queueName string) string {
	return "t:" + string(kn.TenantID) + ":" + queueName + ":metrics"
}

// ConfigKey returns the Redis key for tenant configuration
func (kn KeyNamespace) ConfigKey() string {
	return "tenant:" + string(kn.TenantID) + ":config"
}

// QuotasKey returns the Redis key for tenant quota tracking
func (kn KeyNamespace) QuotasKey() string {
	return "tenant:" + string(kn.TenantID) + ":quotas"
}

// KeysKey returns the Redis key for tenant encryption key metadata
func (kn KeyNamespace) KeysKey() string {
	return "tenant:" + string(kn.TenantID) + ":keys"
}

// AuditKey returns the Redis key for tenant audit log indices
func (kn KeyNamespace) AuditKey() string {
	return "tenant:" + string(kn.TenantID) + ":audit"
}

// AllKeysPattern returns a Redis pattern to match all tenant keys
func (kn KeyNamespace) AllKeysPattern() string {
	return "t:" + string(kn.TenantID) + ":*"
}

// Validate validates the TenantConfig structure
func (tc *TenantConfig) Validate() error {
	if err := tc.ID.Validate(); err != nil {
		return err
	}

	if tc.Name == "" {
		return ErrTenantNameRequired
	}

	if tc.Status == "" {
		tc.Status = TenantStatusActive
	}

	return nil
}

// ToJSON serializes the TenantConfig to JSON
func (tc *TenantConfig) ToJSON() ([]byte, error) {
	return json.Marshal(tc)
}

// FromJSON deserializes TenantConfig from JSON
func (tc *TenantConfig) FromJSON(data []byte) error {
	return json.Unmarshal(data, tc)
}