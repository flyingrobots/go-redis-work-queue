package timetraveldebugger

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.True(t, config.Enabled)
	assert.Equal(t, 0.01, config.SamplingRate)
	assert.True(t, config.ForceOnFailure)
	assert.True(t, config.ForceOnRetry)
	assert.Equal(t, 10000, config.MaxEvents)
	assert.Equal(t, 30*time.Second, config.SnapshotInterval)
	assert.True(t, config.CompressionEnabled)

	// Test retention policy
	assert.Equal(t, 7*24*time.Hour, config.RetentionPolicy.FailedJobs)
	assert.Equal(t, 24*time.Hour, config.RetentionPolicy.SuccessfulJobs)
	assert.Equal(t, 30*24*time.Hour, config.RetentionPolicy.ImportantJobs)
	assert.Equal(t, 10000, config.RetentionPolicy.MaxRecordings)

	// Test sensitive fields
	assert.Contains(t, config.SensitiveFields, "password")
	assert.Contains(t, config.SensitiveFields, "token")
	assert.Contains(t, config.SensitiveFields, "credit_card")
}

func TestProductionConfig(t *testing.T) {
	config := ProductionConfig()

	// Production should have lower sampling and shorter retention
	assert.Equal(t, 0.005, config.SamplingRate)
	assert.Equal(t, 5000, config.MaxEvents)
	assert.Equal(t, time.Minute, config.SnapshotInterval)
	assert.Equal(t, 3*24*time.Hour, config.RetentionPolicy.FailedJobs)
	assert.Equal(t, 12*time.Hour, config.RetentionPolicy.SuccessfulJobs)
}

func TestDevelopmentConfig(t *testing.T) {
	config := DevelopmentConfig()

	// Development should have higher sampling and longer retention
	assert.Equal(t, 0.1, config.SamplingRate)
	assert.Equal(t, 10*time.Second, config.SnapshotInterval)
	assert.Equal(t, 14*24*time.Hour, config.RetentionPolicy.FailedJobs)
	assert.Equal(t, 3*24*time.Hour, config.RetentionPolicy.SuccessfulJobs)
}

func TestTestingConfig(t *testing.T) {
	config := TestingConfig()

	// Testing should capture everything with minimal retention
	assert.Equal(t, 1.0, config.SamplingRate)
	assert.Equal(t, 1000, config.MaxEvents)
	assert.Equal(t, time.Second, config.SnapshotInterval)
	assert.False(t, config.CompressionEnabled)
	assert.Equal(t, time.Hour, config.RetentionPolicy.FailedJobs)
	assert.Equal(t, time.Hour, config.RetentionPolicy.SuccessfulJobs)
	assert.Equal(t, 100, config.RetentionPolicy.MaxRecordings)
}

func TestCaptureConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *CaptureConfig
		expectError bool
		errorField  string
	}{
		{
			name:        "valid config",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name: "invalid sampling rate - negative",
			config: &CaptureConfig{
				SamplingRate: -0.1,
				MaxEvents:    1000,
				SnapshotInterval: time.Second,
				RetentionPolicy: RetentionPolicy{MaxRecordings: 100},
			},
			expectError: true,
			errorField:  "sampling_rate",
		},
		{
			name: "invalid sampling rate - too high",
			config: &CaptureConfig{
				SamplingRate: 1.5,
				MaxEvents:    1000,
				SnapshotInterval: time.Second,
				RetentionPolicy: RetentionPolicy{MaxRecordings: 100},
			},
			expectError: true,
			errorField:  "sampling_rate",
		},
		{
			name: "invalid max events",
			config: &CaptureConfig{
				SamplingRate: 0.1,
				MaxEvents:    0,
				SnapshotInterval: time.Second,
				RetentionPolicy: RetentionPolicy{MaxRecordings: 100},
			},
			expectError: true,
			errorField:  "max_events",
		},
		{
			name: "invalid snapshot interval",
			config: &CaptureConfig{
				SamplingRate: 0.1,
				MaxEvents:    1000,
				SnapshotInterval: 0,
				RetentionPolicy: RetentionPolicy{MaxRecordings: 100},
			},
			expectError: true,
			errorField:  "snapshot_interval",
		},
		{
			name: "invalid max recordings",
			config: &CaptureConfig{
				SamplingRate: 0.1,
				MaxEvents:    1000,
				SnapshotInterval: time.Second,
				RetentionPolicy: RetentionPolicy{MaxRecordings: 0},
			},
			expectError: true,
			errorField:  "retention_policy.max_recordings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.expectError {
				require.Error(t, err)
				configErr, ok := err.(*ConfigError)
				require.True(t, ok, "Expected ConfigError")
				assert.Equal(t, tt.errorField, configErr.Field)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestShouldCapture(t *testing.T) {
	config := &CaptureConfig{
		Enabled:        true,
		SamplingRate:   0.1, // 10%
		ForceOnFailure: true,
		ForceOnRetry:   true,
	}

	tests := []struct {
		name       string
		enabled    bool
		jobStatus  string
		hasRetries bool
		isFailure  bool
		expected   bool
	}{
		{
			name:      "disabled config",
			enabled:   false,
			jobStatus: "completed",
			expected:  false,
		},
		{
			name:      "force on failure",
			enabled:   true,
			jobStatus: "failed",
			isFailure: true,
			expected:  true,
		},
		{
			name:       "force on retry",
			enabled:    true,
			jobStatus:  "processing",
			hasRetries: true,
			expected:   true,
		},
		{
			name:      "normal job - would use sampling",
			enabled:   true,
			jobStatus: "completed",
			expected:  false, // Placeholder implementation returns false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config.Enabled = tt.enabled
			result := config.ShouldCapture(tt.jobStatus, tt.hasRetries, tt.isFailure)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetRetentionPeriod(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		name       string
		importance int
		isFailure  bool
		expected   time.Duration
	}{
		{
			name:       "high importance",
			importance: 8,
			isFailure:  false,
			expected:   config.RetentionPolicy.ImportantJobs,
		},
		{
			name:       "failure job",
			importance: 5,
			isFailure:  true,
			expected:   config.RetentionPolicy.FailedJobs,
		},
		{
			name:       "successful job",
			importance: 5,
			isFailure:  false,
			expected:   config.RetentionPolicy.SuccessfulJobs,
		},
		{
			name:       "high importance failure",
			importance: 9,
			isFailure:  true,
			expected:   config.RetentionPolicy.ImportantJobs, // importance takes precedence
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := config.GetRetentionPeriod(tt.importance, tt.isFailure)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestConfigError(t *testing.T) {
	err := NewConfigError("test_field", "test message")

	assert.Equal(t, "test_field", err.Field)
	assert.Equal(t, "test message", err.Message)
	assert.Contains(t, err.Error(), "test_field")
	assert.Contains(t, err.Error(), "test message")
}

func TestConfigSensitiveFields(t *testing.T) {
	config := DefaultConfig()

	sensitiveFields := []string{
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
	}

	for _, field := range sensitiveFields {
		assert.Contains(t, config.SensitiveFields, field,
			"Config should include sensitive field: %s", field)
	}
}

func TestConfigEnvironmentSpecific(t *testing.T) {
	// Test that environment-specific configs have appropriate settings
	production := ProductionConfig()
	development := DevelopmentConfig()
	testing := TestingConfig()

	// Production should be more conservative
	assert.True(t, production.SamplingRate < development.SamplingRate)
	assert.True(t, production.RetentionPolicy.FailedJobs < development.RetentionPolicy.FailedJobs)

	// Testing should capture everything
	assert.Equal(t, 1.0, testing.SamplingRate)
	assert.False(t, testing.CompressionEnabled) // For easier debugging

	// Development should have frequent snapshots for debugging
	assert.True(t, development.SnapshotInterval < production.SnapshotInterval)
}