package timetraveldebugger

import (
	"time"
)

// DefaultConfig returns a default configuration for the time travel debugger
func DefaultConfig() *CaptureConfig {
	return &CaptureConfig{
		Enabled:            true,
		SamplingRate:       0.01, // 1% of successful jobs
		ForceOnFailure:     true,
		ForceOnRetry:       true,
		MaxEvents:          10000,
		SnapshotInterval:   time.Second * 30,
		CompressionEnabled: true,
		RetentionPolicy: RetentionPolicy{
			FailedJobs:     time.Hour * 24 * 7, // 7 days
			SuccessfulJobs: time.Hour * 24,     // 1 day
			ImportantJobs:  time.Hour * 24 * 30, // 30 days
			MaxRecordings:  10000,
		},
		SensitiveFields: []string{
			"password",
			"token",
			"secret",
			"key",
			"credential",
			"ssn",
			"social_security_number",
			"credit_card",
			"bank_account",
			"routing_number",
		},
	}
}

// ProductionConfig returns a configuration optimized for production use
func ProductionConfig() *CaptureConfig {
	config := DefaultConfig()
	config.SamplingRate = 0.005 // 0.5% sampling to reduce overhead
	config.MaxEvents = 5000     // smaller max events
	config.SnapshotInterval = time.Minute // less frequent snapshots
	config.RetentionPolicy.FailedJobs = time.Hour * 24 * 3 // 3 days
	config.RetentionPolicy.SuccessfulJobs = time.Hour * 12 // 12 hours
	return config
}

// DevelopmentConfig returns a configuration optimized for development
func DevelopmentConfig() *CaptureConfig {
	config := DefaultConfig()
	config.SamplingRate = 0.1         // 10% sampling for more visibility
	config.SnapshotInterval = time.Second * 10 // frequent snapshots for debugging
	config.RetentionPolicy.FailedJobs = time.Hour * 24 * 14 // 2 weeks
	config.RetentionPolicy.SuccessfulJobs = time.Hour * 24 * 3 // 3 days
	return config
}

// TestingConfig returns a configuration optimized for testing
func TestingConfig() *CaptureConfig {
	config := DefaultConfig()
	config.SamplingRate = 1.0 // capture everything
	config.MaxEvents = 1000
	config.SnapshotInterval = time.Second
	config.CompressionEnabled = false // easier to debug
	config.RetentionPolicy.FailedJobs = time.Hour
	config.RetentionPolicy.SuccessfulJobs = time.Hour
	config.RetentionPolicy.MaxRecordings = 100
	return config
}

// Validate checks if the configuration is valid
func (c *CaptureConfig) Validate() error {
	if c.SamplingRate < 0 || c.SamplingRate > 1 {
		return NewConfigError("sampling_rate", "must be between 0.0 and 1.0")
	}
	if c.MaxEvents <= 0 {
		return NewConfigError("max_events", "must be greater than 0")
	}
	if c.SnapshotInterval <= 0 {
		return NewConfigError("snapshot_interval", "must be greater than 0")
	}
	if c.RetentionPolicy.MaxRecordings <= 0 {
		return NewConfigError("retention_policy.max_recordings", "must be greater than 0")
	}
	return nil
}

// ShouldCapture determines if a job should be captured based on the configuration
func (c *CaptureConfig) ShouldCapture(jobStatus string, hasRetries bool, isFailure bool) bool {
	if !c.Enabled {
		return false
	}

	// Always capture failures if configured
	if isFailure && c.ForceOnFailure {
		return true
	}

	// Always capture retries if configured
	if hasRetries && c.ForceOnRetry {
		return true
	}

	// Use sampling rate for other jobs
	// In a real implementation, you'd want a deterministic way to decide
	// based on job ID or other factors to ensure consistent behavior
	return false // Placeholder: would need random sampling logic
}

// GetRetentionPeriod returns the retention period for a given job
func (c *CaptureConfig) GetRetentionPeriod(importance int, isFailure bool) time.Duration {
	if importance >= 8 { // high importance
		return c.RetentionPolicy.ImportantJobs
	}
	if isFailure {
		return c.RetentionPolicy.FailedJobs
	}
	return c.RetentionPolicy.SuccessfulJobs
}

// ConfigError represents a configuration validation error
type ConfigError struct {
	Field   string
	Message string
}

func (e ConfigError) Error() string {
	return "config error for field " + e.Field + ": " + e.Message
}

// NewConfigError creates a new configuration error
func NewConfigError(field, message string) *ConfigError {
	return &ConfigError{
		Field:   field,
		Message: message,
	}
}