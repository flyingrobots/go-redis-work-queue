// Copyright 2025 James Ross
package pluginpanel

import (
	"fmt"
	"time"
)

// ValidateConfig validates a plugin configuration
func (c *PluginConfig) Validate() error {
	if c.PluginDir == "" {
		return fmt.Errorf("plugin directory is required")
	}

	if c.MaxPlugins <= 0 {
		return fmt.Errorf("max plugins must be positive")
	}

	if c.EventQueueSize <= 0 {
		return fmt.Errorf("event queue size must be positive")
	}

	if c.PluginTimeout <= 0 {
		return fmt.Errorf("plugin timeout must be positive")
	}

	if c.GCInterval <= 0 {
		return fmt.Errorf("GC interval must be positive")
	}

	return nil
}

// SetDefaults sets default values for missing configuration
func (c *PluginConfig) SetDefaults() {
	defaults := DefaultPluginConfig()

	if c.PluginDir == "" {
		c.PluginDir = defaults.PluginDir
	}
	if c.MaxPlugins == 0 {
		c.MaxPlugins = defaults.MaxPlugins
	}
	if c.EventQueueSize == 0 {
		c.EventQueueSize = defaults.EventQueueSize
	}
	if c.PluginTimeout == 0 {
		c.PluginTimeout = defaults.PluginTimeout
	}
	if c.GCInterval == 0 {
		c.GCInterval = defaults.GCInterval
	}
	if c.LogLevel == "" {
		c.LogLevel = defaults.LogLevel
	}
	if len(c.DefaultPermissions) == 0 {
		c.DefaultPermissions = defaults.DefaultPermissions
	}
}

// Clone creates a deep copy of the configuration
func (c *PluginConfig) Clone() PluginConfig {
	clone := *c

	// Deep copy slices
	clone.TrustedPlugins = make([]string, len(c.TrustedPlugins))
	copy(clone.TrustedPlugins, c.TrustedPlugins)

	clone.DefaultPermissions = make([]Capability, len(c.DefaultPermissions))
	copy(clone.DefaultPermissions, c.DefaultPermissions)

	return clone
}

// String returns a string representation of the configuration
func (c *PluginConfig) String() string {
	return fmt.Sprintf("PluginConfig{Dir: %s, MaxPlugins: %d, HotReload: %v, Sandbox: %v}",
		c.PluginDir, c.MaxPlugins, c.HotReload, c.SandboxEnabled)
}

// Development configuration presets

// DevelopmentConfig returns a configuration optimized for development
func DevelopmentConfig() PluginConfig {
	config := DefaultPluginConfig()
	config.HotReload = true
	config.SandboxEnabled = false // Disable sandbox for easier debugging
	config.LogLevel = "debug"
	config.PermissionTimeout = 60 * time.Second // Longer timeout for manual approval
	config.PluginTimeout = 300 * time.Second   // Longer timeout for debugging
	return config
}

// ProductionConfig returns a configuration optimized for production
func ProductionConfig() PluginConfig {
	config := DefaultPluginConfig()
	config.HotReload = false
	config.SandboxEnabled = true
	config.LogLevel = "info"
	config.PermissionTimeout = 10 * time.Second
	config.PluginTimeout = 30 * time.Second
	config.MaxPlugins = 50
	return config
}

// TestingConfig returns a configuration optimized for testing
func TestingConfig() PluginConfig {
	config := DefaultPluginConfig()
	config.HotReload = false
	config.SandboxEnabled = false
	config.LogLevel = "warn"
	config.MaxPlugins = 5
	config.EventQueueSize = 100
	config.PluginTimeout = 5 * time.Second
	config.GCInterval = 1 * time.Second
	return config
}

// Security configuration helpers

// RestrictiveConfig returns a configuration with minimal permissions
func RestrictiveConfig() PluginConfig {
	config := DefaultPluginConfig()
	config.SandboxEnabled = true
	config.DefaultPermissions = []Capability{} // No default permissions
	config.TrustedPlugins = []string{}          // No trusted plugins
	return config
}

// TrustedConfig returns a configuration for trusted environments
func TrustedConfig() PluginConfig {
	config := DefaultPluginConfig()
	config.SandboxEnabled = false
	config.DefaultPermissions = []Capability{
		CapabilityReadStats,
		CapabilityReadKeys,
		CapabilityReadSelection,
		CapabilityReadTimers,
		CapabilityReadQueues,
		CapabilityReadJobs,
		CapabilityRenderPanel,
		CapabilityKeyEvents,
		CapabilityMouseEvents,
	}
	return config
}