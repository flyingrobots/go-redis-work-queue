// Copyright 2025 James Ross
package backpressure

import (
	"fmt"
	"time"
)

// Validate validates the backpressure configuration
func (c *BackpressureConfig) Validate() error {
	if err := c.Thresholds.Validate(); err != nil {
		return fmt.Errorf("thresholds validation failed: %w", err)
	}

	if err := c.Circuit.Validate(); err != nil {
		return fmt.Errorf("circuit configuration validation failed: %w", err)
	}

	if err := c.Polling.Validate(); err != nil {
		return fmt.Errorf("polling configuration validation failed: %w", err)
	}

	if err := c.Recovery.Validate(); err != nil {
		return fmt.Errorf("recovery configuration validation failed: %w", err)
	}

	return nil
}

// Validate validates backlog thresholds
func (bt *BacklogThresholds) Validate() error {
	if err := bt.HighPriority.Validate("high_priority"); err != nil {
		return err
	}
	if err := bt.MediumPriority.Validate("medium_priority"); err != nil {
		return err
	}
	if err := bt.LowPriority.Validate("low_priority"); err != nil {
		return err
	}
	return nil
}

// Validate validates a backlog window
func (bw *BacklogWindow) Validate(name string) error {
	if bw.Green < 0 {
		return NewConfigurationError(name+".green_max", bw.Green,
			"green threshold must be non-negative")
	}
	if bw.Yellow <= bw.Green {
		return NewConfigurationError(name+".yellow_max", bw.Yellow,
			"yellow threshold must be greater than green threshold")
	}
	if bw.Red <= bw.Yellow {
		return NewConfigurationError(name+".red_max", bw.Red,
			"red threshold must be greater than yellow threshold")
	}
	if bw.Red > 1000000 {
		return NewConfigurationError(name+".red_max", bw.Red,
			"red threshold is unreasonably high (max 1M)")
	}
	return nil
}

// Validate validates circuit breaker configuration
func (cc *CircuitConfig) Validate() error {
	if cc.FailureThreshold <= 0 {
		return NewConfigurationError("circuit.failure_threshold", cc.FailureThreshold,
			"failure threshold must be positive")
	}
	if cc.RecoveryThreshold <= 0 {
		return NewConfigurationError("circuit.recovery_threshold", cc.RecoveryThreshold,
			"recovery threshold must be positive")
	}
	if cc.TripWindow <= 0 {
		return NewConfigurationError("circuit.trip_window", cc.TripWindow,
			"trip window must be positive")
	}
	if cc.RecoveryTimeout <= 0 {
		return NewConfigurationError("circuit.recovery_timeout", cc.RecoveryTimeout,
			"recovery timeout must be positive")
	}
	if cc.ProbeInterval <= 0 {
		return NewConfigurationError("circuit.probe_interval", cc.ProbeInterval,
			"probe interval must be positive")
	}

	// Sanity checks for reasonable values
	if cc.FailureThreshold > 1000 {
		return NewConfigurationError("circuit.failure_threshold", cc.FailureThreshold,
			"failure threshold is unreasonably high (max 1000)")
	}
	if cc.TripWindow > 24*time.Hour {
		return NewConfigurationError("circuit.trip_window", cc.TripWindow,
			"trip window is unreasonably long (max 24h)")
	}
	if cc.RecoveryTimeout > 24*time.Hour {
		return NewConfigurationError("circuit.recovery_timeout", cc.RecoveryTimeout,
			"recovery timeout is unreasonably long (max 24h)")
	}
	if cc.ProbeInterval < time.Second {
		return NewConfigurationError("circuit.probe_interval", cc.ProbeInterval,
			"probe interval is too short (min 1s)")
	}

	return nil
}

// Validate validates polling configuration
func (pc *PollingConfig) Validate() error {
	if pc.Interval <= 0 {
		return NewConfigurationError("polling.interval", pc.Interval,
			"polling interval must be positive")
	}
	if pc.Jitter < 0 {
		return NewConfigurationError("polling.jitter", pc.Jitter,
			"jitter must be non-negative")
	}
	if pc.Timeout <= 0 {
		return NewConfigurationError("polling.timeout", pc.Timeout,
			"timeout must be positive")
	}
	if pc.MaxBackoff <= 0 {
		return NewConfigurationError("polling.max_backoff", pc.MaxBackoff,
			"max backoff must be positive")
	}
	if pc.CacheTTL <= 0 {
		return NewConfigurationError("polling.cache_ttl", pc.CacheTTL,
			"cache TTL must be positive")
	}

	// Sanity checks
	if pc.Interval < time.Second {
		return NewConfigurationError("polling.interval", pc.Interval,
			"polling interval is too short (min 1s)")
	}
	if pc.Jitter > pc.Interval {
		return NewConfigurationError("polling.jitter", pc.Jitter,
			"jitter cannot exceed polling interval")
	}
	if pc.Timeout > pc.Interval {
		return NewConfigurationError("polling.timeout", pc.Timeout,
			"timeout cannot exceed polling interval")
	}
	if pc.MaxBackoff > 24*time.Hour {
		return NewConfigurationError("polling.max_backoff", pc.MaxBackoff,
			"max backoff is unreasonably long (max 24h)")
	}
	if pc.CacheTTL > time.Hour {
		return NewConfigurationError("polling.cache_ttl", pc.CacheTTL,
			"cache TTL is unreasonably long (max 1h)")
	}

	return nil
}

// Validate validates recovery strategy configuration
func (rs *RecoveryStrategy) Validate() error {
	if rs.GracefulDegrade < 0 {
		return NewConfigurationError("recovery.graceful_degrade", rs.GracefulDegrade,
			"graceful degrade duration must be non-negative")
	}
	if rs.GracefulDegrade > 24*time.Hour {
		return NewConfigurationError("recovery.graceful_degrade", rs.GracefulDegrade,
			"graceful degrade duration is unreasonably long (max 24h)")
	}
	return nil
}

// GetBacklogWindow returns the appropriate backlog window for a priority
func (bt *BacklogThresholds) GetBacklogWindow(priority Priority) BacklogWindow {
	switch priority {
	case HighPriority:
		return bt.HighPriority
	case MediumPriority:
		return bt.MediumPriority
	case LowPriority:
		return bt.LowPriority
	default:
		// Default to medium priority for unknown priorities
		return bt.MediumPriority
	}
}

// Clone creates a deep copy of the configuration
func (c *BackpressureConfig) Clone() BackpressureConfig {
	return BackpressureConfig{
		Thresholds: c.Thresholds,
		Circuit:    c.Circuit,
		Polling:    c.Polling,
		Recovery:   c.Recovery,
	}
}

// SetDefaults fills in missing values with defaults
func (c *BackpressureConfig) SetDefaults() {
	defaults := DefaultConfig()

	// Set threshold defaults if zero values
	if c.Thresholds.HighPriority.Green == 0 {
		c.Thresholds = defaults.Thresholds
	}

	// Set circuit defaults if zero values
	if c.Circuit.FailureThreshold == 0 {
		c.Circuit = defaults.Circuit
	}

	// Set polling defaults if zero values
	if c.Polling.Interval == 0 {
		c.Polling = defaults.Polling
	}

	// Set recovery defaults if zero values
	if c.Recovery.GracefulDegrade == 0 {
		c.Recovery = defaults.Recovery
	}
}

// String returns a string representation of the configuration
func (c *BackpressureConfig) String() string {
	return fmt.Sprintf("BackpressureConfig{Thresholds: %+v, Circuit: %+v, Polling: %+v, Recovery: %+v}",
		c.Thresholds, c.Circuit, c.Polling, c.Recovery)
}

// IsValidPriority checks if a priority level is valid
func IsValidPriority(priority Priority) bool {
	return priority >= LowPriority && priority <= HighPriority
}

// ParsePriority converts a string to a Priority
func ParsePriority(s string) (Priority, error) {
	switch s {
	case "low":
		return LowPriority, nil
	case "medium":
		return MediumPriority, nil
	case "high":
		return HighPriority, nil
	default:
		return -1, fmt.Errorf("invalid priority %q: must be low, medium, or high", s)
	}
}

// IsValidQueueName validates a queue name
func IsValidQueueName(name string) bool {
	return len(name) > 0 && len(name) <= 255
}