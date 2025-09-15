// Copyright 2025 James Ross
package backpressure

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.NoError(t, config.Validate())
	assert.Greater(t, config.Thresholds.HighPriority.Green, 0)
	assert.Greater(t, config.Circuit.FailureThreshold, 0)
	assert.Greater(t, config.Polling.Interval, time.Duration(0))
}

func TestBacklogThresholdsValidation(t *testing.T) {
	tests := []struct {
		name      string
		threshold BacklogThresholds
		wantErr   bool
	}{
		{
			name:      "valid thresholds",
			threshold: DefaultThresholds(),
			wantErr:   false,
		},
		{
			name: "negative green",
			threshold: BacklogThresholds{
				HighPriority: BacklogWindow{Green: -1, Yellow: 100, Red: 200},
			},
			wantErr: true,
		},
		{
			name: "yellow not greater than green",
			threshold: BacklogThresholds{
				HighPriority: BacklogWindow{Green: 100, Yellow: 100, Red: 200},
			},
			wantErr: true,
		},
		{
			name: "red not greater than yellow",
			threshold: BacklogThresholds{
				HighPriority: BacklogWindow{Green: 100, Yellow: 200, Red: 200},
			},
			wantErr: true,
		},
		{
			name: "red threshold too high",
			threshold: BacklogThresholds{
				HighPriority: BacklogWindow{Green: 100, Yellow: 200, Red: 2000000},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.threshold.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCircuitConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config CircuitConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultCircuitConfig(),
			wantErr: false,
		},
		{
			name: "zero failure threshold",
			config: CircuitConfig{
				FailureThreshold:  0,
				RecoveryThreshold: 3,
				TripWindow:        30 * time.Second,
				RecoveryTimeout:   60 * time.Second,
				ProbeInterval:     5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "zero recovery threshold",
			config: CircuitConfig{
				FailureThreshold:  5,
				RecoveryThreshold: 0,
				TripWindow:        30 * time.Second,
				RecoveryTimeout:   60 * time.Second,
				ProbeInterval:     5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative trip window",
			config: CircuitConfig{
				FailureThreshold:  5,
				RecoveryThreshold: 3,
				TripWindow:        -1 * time.Second,
				RecoveryTimeout:   60 * time.Second,
				ProbeInterval:     5 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "probe interval too short",
			config: CircuitConfig{
				FailureThreshold:  5,
				RecoveryThreshold: 3,
				TripWindow:        30 * time.Second,
				RecoveryTimeout:   60 * time.Second,
				ProbeInterval:     500 * time.Millisecond,
			},
			wantErr: true,
		},
		{
			name: "unreasonably high failure threshold",
			config: CircuitConfig{
				FailureThreshold:  2000,
				RecoveryThreshold: 3,
				TripWindow:        30 * time.Second,
				RecoveryTimeout:   60 * time.Second,
				ProbeInterval:     5 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPollingConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config PollingConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultPollingConfig(),
			wantErr: false,
		},
		{
			name: "zero interval",
			config: PollingConfig{
				Interval:   0,
				Jitter:     1 * time.Second,
				Timeout:    3 * time.Second,
				MaxBackoff: 60 * time.Second,
				CacheTTL:   30 * time.Second,
				Enabled:    true,
			},
			wantErr: true,
		},
		{
			name: "negative jitter",
			config: PollingConfig{
				Interval:   5 * time.Second,
				Jitter:     -1 * time.Second,
				Timeout:    3 * time.Second,
				MaxBackoff: 60 * time.Second,
				CacheTTL:   30 * time.Second,
				Enabled:    true,
			},
			wantErr: true,
		},
		{
			name: "interval too short",
			config: PollingConfig{
				Interval:   500 * time.Millisecond,
				Jitter:     100 * time.Millisecond,
				Timeout:    200 * time.Millisecond,
				MaxBackoff: 60 * time.Second,
				CacheTTL:   30 * time.Second,
				Enabled:    true,
			},
			wantErr: true,
		},
		{
			name: "jitter exceeds interval",
			config: PollingConfig{
				Interval:   5 * time.Second,
				Jitter:     6 * time.Second,
				Timeout:    3 * time.Second,
				MaxBackoff: 60 * time.Second,
				CacheTTL:   30 * time.Second,
				Enabled:    true,
			},
			wantErr: true,
		},
		{
			name: "timeout exceeds interval",
			config: PollingConfig{
				Interval:   5 * time.Second,
				Jitter:     1 * time.Second,
				Timeout:    6 * time.Second,
				MaxBackoff: 60 * time.Second,
				CacheTTL:   30 * time.Second,
				Enabled:    true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetBacklogWindow(t *testing.T) {
	thresholds := DefaultThresholds()

	highWindow := thresholds.GetBacklogWindow(HighPriority)
	assert.Equal(t, thresholds.HighPriority, highWindow)

	mediumWindow := thresholds.GetBacklogWindow(MediumPriority)
	assert.Equal(t, thresholds.MediumPriority, mediumWindow)

	lowWindow := thresholds.GetBacklogWindow(LowPriority)
	assert.Equal(t, thresholds.LowPriority, lowWindow)

	// Unknown priority should default to medium
	unknownWindow := thresholds.GetBacklogWindow(Priority(-1))
	assert.Equal(t, thresholds.MediumPriority, unknownWindow)
}

func TestConfigClone(t *testing.T) {
	original := DefaultConfig()
	clone := original.Clone()

	assert.Equal(t, original, clone)

	// Modify clone and ensure original is unchanged
	clone.Circuit.FailureThreshold = 999
	assert.NotEqual(t, original.Circuit.FailureThreshold, clone.Circuit.FailureThreshold)
}

func TestConfigSetDefaults(t *testing.T) {
	config := BackpressureConfig{}

	// Should have zero values initially
	assert.Equal(t, 0, config.Thresholds.HighPriority.Green)
	assert.Equal(t, 0, config.Circuit.FailureThreshold)
	assert.Equal(t, time.Duration(0), config.Polling.Interval)

	config.SetDefaults()

	// Should now have default values
	assert.Greater(t, config.Thresholds.HighPriority.Green, 0)
	assert.Greater(t, config.Circuit.FailureThreshold, 0)
	assert.Greater(t, config.Polling.Interval, time.Duration(0))
}

func TestParsePriority(t *testing.T) {
	tests := []struct {
		input    string
		expected Priority
		wantErr  bool
	}{
		{"low", LowPriority, false},
		{"medium", MediumPriority, false},
		{"high", HighPriority, false},
		{"invalid", Priority(-1), true},
		{"", Priority(-1), true},
		{"LOW", Priority(-1), true}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParsePriority(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsValidPriority(t *testing.T) {
	assert.True(t, IsValidPriority(LowPriority))
	assert.True(t, IsValidPriority(MediumPriority))
	assert.True(t, IsValidPriority(HighPriority))
	assert.False(t, IsValidPriority(Priority(-1)))
	assert.False(t, IsValidPriority(Priority(999)))
}

func TestIsValidQueueName(t *testing.T) {
	assert.True(t, IsValidQueueName("valid-queue"))
	assert.True(t, IsValidQueueName("a"))
	assert.False(t, IsValidQueueName(""))

	// Test very long name (over 255 chars)
	longName := string(make([]byte, 256))
	for i := range longName {
		longName = longName[:i] + "a" + longName[i+1:]
	}
	assert.False(t, IsValidQueueName(longName))
}

func TestPriorityString(t *testing.T) {
	assert.Equal(t, "low", LowPriority.String())
	assert.Equal(t, "medium", MediumPriority.String())
	assert.Equal(t, "high", HighPriority.String())
	assert.Equal(t, "unknown", Priority(-1).String())
}

func TestCircuitStateString(t *testing.T) {
	assert.Equal(t, "closed", Closed.String())
	assert.Equal(t, "open", Open.String())
	assert.Equal(t, "half-open", HalfOpen.String())
	assert.Equal(t, "unknown", CircuitState(-1).String())
}

func TestRecoveryStrategyValidation(t *testing.T) {
	tests := []struct {
		name     string
		strategy RecoveryStrategy
		wantErr  bool
	}{
		{
			name:     "valid strategy",
			strategy: DefaultRecoveryStrategy(),
			wantErr:  false,
		},
		{
			name: "negative graceful degrade",
			strategy: RecoveryStrategy{
				GracefulDegrade: -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "unreasonably long graceful degrade",
			strategy: RecoveryStrategy{
				GracefulDegrade: 25 * time.Hour,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.strategy.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigString(t *testing.T) {
	config := DefaultConfig()
	str := config.String()

	assert.Contains(t, str, "BackpressureConfig")
	assert.Contains(t, str, "Thresholds")
	assert.Contains(t, str, "Circuit")
	assert.Contains(t, str, "Polling")
	assert.Contains(t, str, "Recovery")
}