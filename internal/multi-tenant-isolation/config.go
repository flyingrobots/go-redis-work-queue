// Copyright 2025 James Ross
package multitenantiso

import (
	"fmt"
	"time"
)

// Config represents configuration for multi-tenant isolation
type Config struct {
	// Global settings
	Enabled                    bool          `yaml:"enabled" json:"enabled"`
	DefaultEncryptionEnabled   bool          `yaml:"default_encryption_enabled" json:"default_encryption_enabled"`
	DefaultKEKProvider         string        `yaml:"default_kek_provider" json:"default_kek_provider"`
	DefaultDEKRotationPeriod   time.Duration `yaml:"default_dek_rotation_period" json:"default_dek_rotation_period"`

	// Rate limiting defaults
	DefaultRateLimitingEnabled bool          `yaml:"default_rate_limiting_enabled" json:"default_rate_limiting_enabled"`
	DefaultWindowDuration      time.Duration `yaml:"default_window_duration" json:"default_window_duration"`
	DefaultBurstCapacity       int32         `yaml:"default_burst_capacity" json:"default_burst_capacity"`

	// Quota defaults
	DefaultQuotasConfig        TenantQuotas  `yaml:"default_quotas" json:"default_quotas"`

	// Audit settings
	AuditEnabled              bool          `yaml:"audit_enabled" json:"audit_enabled"`
	AuditRetentionDays        int           `yaml:"audit_retention_days" json:"audit_retention_days"`

	// Security settings
	RequireEncryptionForSensitiveData bool     `yaml:"require_encryption_for_sensitive_data" json:"require_encryption_for_sensitive_data"`
	AllowedKEKProviders              []string `yaml:"allowed_kek_providers" json:"allowed_kek_providers"`
	MinTenantQuotaLimits             TenantQuotas `yaml:"min_tenant_quota_limits" json:"min_tenant_quota_limits"`
	MaxTenantQuotaLimits             TenantQuotas `yaml:"max_tenant_quota_limits" json:"max_tenant_quota_limits"`
}

// DefaultConfig returns a reasonable default configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:                  true,
		DefaultEncryptionEnabled: false,
		DefaultKEKProvider:       "local",
		DefaultDEKRotationPeriod: 24 * time.Hour * 7, // Weekly rotation

		DefaultRateLimitingEnabled: true,
		DefaultWindowDuration:      time.Second,
		DefaultBurstCapacity:       10,

		DefaultQuotasConfig: DefaultQuotas(),

		AuditEnabled:       true,
		AuditRetentionDays: 90,

		RequireEncryptionForSensitiveData: false,
		AllowedKEKProviders:              []string{"local", "aws-kms", "gcp-kms", "azure-kv"},

		MinTenantQuotaLimits: TenantQuotas{
			MaxJobsPerHour:     100,
			MaxJobsPerDay:      1000,
			MaxBacklogSize:     1000,
			MaxJobSizeBytes:    1024,        // 1KB
			MaxQueuesPerTenant: 1,
			MaxWorkersPerQueue: 1,
			MaxStorageBytes:    1024 * 1024, // 1MB
			EnqueueRateLimit:   1,
			DequeueRateLimit:   1,
			SoftLimitThreshold: 0.5,
		},

		MaxTenantQuotaLimits: TenantQuotas{
			MaxJobsPerHour:     1000000,
			MaxJobsPerDay:      10000000,
			MaxBacklogSize:     1000000,
			MaxJobSizeBytes:    100 * 1024 * 1024, // 100MB
			MaxQueuesPerTenant: 1000,
			MaxWorkersPerQueue: 1000,
			MaxStorageBytes:    10 * 1024 * 1024 * 1024, // 10GB
			EnqueueRateLimit:   10000,
			DequeueRateLimit:   10000,
			SoftLimitThreshold: 1.0,
		},
	}
}

// ValidateConfig validates the configuration
func (c *Config) ValidateConfig() error {
	if c.DefaultDEKRotationPeriod <= 0 {
		return NewConfigurationError("multi-tenant-isolation", "DEK rotation period must be positive", "")
	}

	if c.DefaultWindowDuration <= 0 {
		return NewConfigurationError("multi-tenant-isolation", "rate limiting window duration must be positive", "")
	}

	if c.AuditRetentionDays <= 0 {
		return NewConfigurationError("multi-tenant-isolation", "audit retention days must be positive", "")
	}

	if len(c.AllowedKEKProviders) == 0 {
		return NewConfigurationError("multi-tenant-isolation", "at least one KEK provider must be allowed", "")
	}

	// Validate that min limits are less than max limits
	if c.MinTenantQuotaLimits.MaxJobsPerHour > c.MaxTenantQuotaLimits.MaxJobsPerHour {
		return NewConfigurationError("multi-tenant-isolation", "min jobs per hour exceeds max", "")
	}

	return nil
}

// GetDefaultTenantConfig creates a default tenant configuration
func (c *Config) GetDefaultTenantConfig(tenantID TenantID, tenantName string) *TenantConfig {
	return &TenantConfig{
		ID:     tenantID,
		Name:   tenantName,
		Status: TenantStatusActive,
		Quotas: c.DefaultQuotasConfig,
		Encryption: TenantEncryption{
			Enabled:           c.DefaultEncryptionEnabled,
			KEKProvider:       c.DefaultKEKProvider,
			DEKRotationPeriod: c.DefaultDEKRotationPeriod,
			Algorithm:         "AES-256-GCM",
		},
		RateLimiting: TenantRateLimiting{
			Enabled:              c.DefaultRateLimitingEnabled,
			WindowDuration:       c.DefaultWindowDuration,
			BurstCapacity:        c.DefaultBurstCapacity,
			EnforceAcrossWorkers: true,
		},
		Metadata: make(map[string]string),
	}
}

// IsKEKProviderAllowed checks if a KEK provider is allowed
func (c *Config) IsKEKProviderAllowed(provider string) bool {
	for _, allowed := range c.AllowedKEKProviders {
		if allowed == provider {
			return true
		}
	}
	return false
}

// ValidateTenantQuotas checks if tenant quotas are within allowed limits
func (c *Config) ValidateTenantQuotas(quotas *TenantQuotas) error {
	if quotas.MaxJobsPerHour < c.MinTenantQuotaLimits.MaxJobsPerHour {
		return NewValidationError("max_jobs_per_hour", fmt.Sprintf("%d", quotas.MaxJobsPerHour), "below minimum limit")
	}
	if quotas.MaxJobsPerHour > c.MaxTenantQuotaLimits.MaxJobsPerHour {
		return NewValidationError("max_jobs_per_hour", fmt.Sprintf("%d", quotas.MaxJobsPerHour), "exceeds maximum limit")
	}

	if quotas.MaxJobsPerDay < c.MinTenantQuotaLimits.MaxJobsPerDay {
		return NewValidationError("max_jobs_per_day", fmt.Sprintf("%d", quotas.MaxJobsPerDay), "below minimum limit")
	}
	if quotas.MaxJobsPerDay > c.MaxTenantQuotaLimits.MaxJobsPerDay {
		return NewValidationError("max_jobs_per_day", fmt.Sprintf("%d", quotas.MaxJobsPerDay), "exceeds maximum limit")
	}

	if quotas.MaxBacklogSize < c.MinTenantQuotaLimits.MaxBacklogSize {
		return NewValidationError("max_backlog_size", fmt.Sprintf("%d", quotas.MaxBacklogSize), "below minimum limit")
	}
	if quotas.MaxBacklogSize > c.MaxTenantQuotaLimits.MaxBacklogSize {
		return NewValidationError("max_backlog_size", fmt.Sprintf("%d", quotas.MaxBacklogSize), "exceeds maximum limit")
	}

	if quotas.MaxStorageBytes < c.MinTenantQuotaLimits.MaxStorageBytes {
		return NewValidationError("max_storage_bytes", fmt.Sprintf("%d", quotas.MaxStorageBytes), "below minimum limit")
	}
	if quotas.MaxStorageBytes > c.MaxTenantQuotaLimits.MaxStorageBytes {
		return NewValidationError("max_storage_bytes", fmt.Sprintf("%d", quotas.MaxStorageBytes), "exceeds maximum limit")
	}

	return nil
}