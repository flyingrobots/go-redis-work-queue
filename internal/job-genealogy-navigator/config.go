// Copyright 2025 James Ross
package genealogy

import (
	"fmt"
	"time"
)

// Validate validates the genealogy configuration
func (c *GenealogyConfig) Validate() error {
	if c.RedisKeyPrefix == "" {
		return NewValidationError("redis_key_prefix", c.RedisKeyPrefix,
			"Redis key prefix cannot be empty")
	}

	if c.CacheTTL <= 0 {
		return NewValidationError("cache_ttl", c.CacheTTL,
			"Cache TTL must be positive")
	}

	if c.MaxTreeSize <= 0 {
		return NewValidationError("max_tree_size", c.MaxTreeSize,
			"Max tree size must be positive")
	}

	if c.MaxTreeSize > 100000 {
		return NewValidationError("max_tree_size", c.MaxTreeSize,
			"Max tree size is unreasonably large (max 100,000)")
	}

	if c.MaxGenerations <= 0 {
		return NewValidationError("max_generations", c.MaxGenerations,
			"Max generations must be positive")
	}

	if c.MaxGenerations > 1000 {
		return NewValidationError("max_generations", c.MaxGenerations,
			"Max generations is unreasonably large (max 1,000)")
	}

	if !IsValidViewMode(c.DefaultViewMode) {
		return NewValidationError("default_view_mode", c.DefaultViewMode,
			"Invalid default view mode")
	}

	if !IsValidLayoutMode(c.DefaultLayoutMode) {
		return NewValidationError("default_layout_mode", c.DefaultLayoutMode,
			"Invalid default layout mode")
	}

	if c.NodeWidth <= 0 {
		return NewValidationError("node_width", c.NodeWidth,
			"Node width must be positive")
	}

	if c.NodeHeight <= 0 {
		return NewValidationError("node_height", c.NodeHeight,
			"Node height must be positive")
	}

	if c.HorizontalSpacing < 0 {
		return NewValidationError("horizontal_spacing", c.HorizontalSpacing,
			"Horizontal spacing cannot be negative")
	}

	if c.VerticalSpacing < 0 {
		return NewValidationError("vertical_spacing", c.VerticalSpacing,
			"Vertical spacing cannot be negative")
	}

	if c.RefreshInterval <= 0 {
		return NewValidationError("refresh_interval", c.RefreshInterval,
			"Refresh interval must be positive")
	}

	if c.RelationshipTTL <= 0 {
		return NewValidationError("relationship_ttl", c.RelationshipTTL,
			"Relationship TTL must be positive")
	}

	if c.TreeCacheTTL <= 0 {
		return NewValidationError("tree_cache_ttl", c.TreeCacheTTL,
			"Tree cache TTL must be positive")
	}

	if c.CleanupInterval <= 0 {
		return NewValidationError("cleanup_interval", c.CleanupInterval,
			"Cleanup interval must be positive")
	}

	// Sanity checks for reasonable values
	if c.RefreshInterval < time.Second {
		return NewValidationError("refresh_interval", c.RefreshInterval,
			"Refresh interval is too short (min 1s)")
	}

	if c.RelationshipTTL > 30*24*time.Hour {
		return NewValidationError("relationship_ttl", c.RelationshipTTL,
			"Relationship TTL is too long (max 30 days)")
	}

	if c.TreeCacheTTL > time.Hour {
		return NewValidationError("tree_cache_ttl", c.TreeCacheTTL,
			"Tree cache TTL is too long (max 1 hour)")
	}

	return nil
}

// Clone creates a deep copy of the configuration
func (c *GenealogyConfig) Clone() GenealogyConfig {
	return GenealogyConfig{
		RedisKeyPrefix:    c.RedisKeyPrefix,
		CacheTTL:          c.CacheTTL,
		MaxTreeSize:       c.MaxTreeSize,
		MaxGenerations:    c.MaxGenerations,
		DefaultViewMode:   c.DefaultViewMode,
		DefaultLayoutMode: c.DefaultLayoutMode,
		NodeWidth:         c.NodeWidth,
		NodeHeight:        c.NodeHeight,
		HorizontalSpacing: c.HorizontalSpacing,
		VerticalSpacing:   c.VerticalSpacing,
		EnableCaching:     c.EnableCaching,
		LazyLoading:       c.LazyLoading,
		BackgroundRefresh: c.BackgroundRefresh,
		RefreshInterval:   c.RefreshInterval,
		RelationshipTTL:   c.RelationshipTTL,
		TreeCacheTTL:      c.TreeCacheTTL,
		CleanupInterval:   c.CleanupInterval,
	}
}

// SetDefaults fills in missing values with defaults
func (c *GenealogyConfig) SetDefaults() {
	defaults := DefaultGenealogyConfig()

	if c.RedisKeyPrefix == "" {
		c.RedisKeyPrefix = defaults.RedisKeyPrefix
	}

	if c.CacheTTL == 0 {
		c.CacheTTL = defaults.CacheTTL
	}

	if c.MaxTreeSize == 0 {
		c.MaxTreeSize = defaults.MaxTreeSize
	}

	if c.MaxGenerations == 0 {
		c.MaxGenerations = defaults.MaxGenerations
	}

	if c.DefaultViewMode == "" {
		c.DefaultViewMode = defaults.DefaultViewMode
	}

	if c.DefaultLayoutMode == "" {
		c.DefaultLayoutMode = defaults.DefaultLayoutMode
	}

	if c.NodeWidth == 0 {
		c.NodeWidth = defaults.NodeWidth
	}

	if c.NodeHeight == 0 {
		c.NodeHeight = defaults.NodeHeight
	}

	if c.RefreshInterval == 0 {
		c.RefreshInterval = defaults.RefreshInterval
	}

	if c.RelationshipTTL == 0 {
		c.RelationshipTTL = defaults.RelationshipTTL
	}

	if c.TreeCacheTTL == 0 {
		c.TreeCacheTTL = defaults.TreeCacheTTL
	}

	if c.CleanupInterval == 0 {
		c.CleanupInterval = defaults.CleanupInterval
	}
}

// String returns a string representation of the configuration
func (c *GenealogyConfig) String() string {
	return fmt.Sprintf("GenealogyConfig{KeyPrefix: %s, CacheTTL: %v, MaxTreeSize: %d, ViewMode: %s, LayoutMode: %s}",
		c.RedisKeyPrefix, c.CacheTTL, c.MaxTreeSize, c.DefaultViewMode, c.DefaultLayoutMode)
}

// Validation functions for enum types

// IsValidViewMode checks if a view mode is valid
func IsValidViewMode(mode ViewMode) bool {
	switch mode {
	case ViewModeFull, ViewModeAncestors, ViewModeDescendants, ViewModeBlamePath, ViewModeImpactZone:
		return true
	default:
		return false
	}
}

// IsValidLayoutMode checks if a layout mode is valid
func IsValidLayoutMode(mode LayoutMode) bool {
	switch mode {
	case LayoutModeTopDown, LayoutModeTimeline, LayoutModeRadial, LayoutModeCompact:
		return true
	default:
		return false
	}
}

// IsValidRelationshipType checks if a relationship type is valid
func IsValidRelationshipType(relType RelationshipType) bool {
	switch relType {
	case RelationshipRetry, RelationshipSpawn, RelationshipFork,
		 RelationshipCallback, RelationshipCompensation,
		 RelationshipContinuation, RelationshipBatchMember:
		return true
	default:
		return false
	}
}

// IsValidJobStatus checks if a job status is valid
func IsValidJobStatus(status JobStatus) bool {
	switch status {
	case JobStatusPending, JobStatusProcessing, JobStatusSuccess,
		 JobStatusFailed, JobStatusRetry, JobStatusCancelled:
		return true
	default:
		return false
	}
}

// ParseViewMode converts a string to a ViewMode
func ParseViewMode(s string) (ViewMode, error) {
	switch s {
	case "full":
		return ViewModeFull, nil
	case "ancestors":
		return ViewModeAncestors, nil
	case "descendants":
		return ViewModeDescendants, nil
	case "blame":
		return ViewModeBlamePath, nil
	case "impact":
		return ViewModeImpactZone, nil
	default:
		return "", fmt.Errorf("invalid view mode %q: must be one of full, ancestors, descendants, blame, impact", s)
	}
}

// ParseLayoutMode converts a string to a LayoutMode
func ParseLayoutMode(s string) (LayoutMode, error) {
	switch s {
	case "topdown":
		return LayoutModeTopDown, nil
	case "timeline":
		return LayoutModeTimeline, nil
	case "radial":
		return LayoutModeRadial, nil
	case "compact":
		return LayoutModeCompact, nil
	default:
		return "", fmt.Errorf("invalid layout mode %q: must be one of topdown, timeline, radial, compact", s)
	}
}

// ParseRelationshipType converts a string to a RelationshipType
func ParseRelationshipType(s string) (RelationshipType, error) {
	switch s {
	case "retry":
		return RelationshipRetry, nil
	case "spawn":
		return RelationshipSpawn, nil
	case "fork":
		return RelationshipFork, nil
	case "callback":
		return RelationshipCallback, nil
	case "compensation":
		return RelationshipCompensation, nil
	case "continuation":
		return RelationshipContinuation, nil
	case "batch_member":
		return RelationshipBatchMember, nil
	default:
		return "", fmt.Errorf("invalid relationship type %q", s)
	}
}

// ParseJobStatus converts a string to a JobStatus
func ParseJobStatus(s string) (JobStatus, error) {
	switch s {
	case "pending":
		return JobStatusPending, nil
	case "processing":
		return JobStatusProcessing, nil
	case "success":
		return JobStatusSuccess, nil
	case "failed":
		return JobStatusFailed, nil
	case "retry":
		return JobStatusRetry, nil
	case "cancelled":
		return JobStatusCancelled, nil
	default:
		return "", fmt.Errorf("invalid job status %q", s)
	}
}

// Configuration presets

// DevelopmentConfig returns a configuration optimized for development
func DevelopmentConfig() GenealogyConfig {
	config := DefaultGenealogyConfig()
	config.CacheTTL = 1 * time.Minute
	config.RefreshInterval = 10 * time.Second
	config.RelationshipTTL = 2 * time.Hour
	config.BackgroundRefresh = true
	config.EnableCaching = true
	config.LazyLoading = true
	return config
}

// ProductionConfig returns a configuration optimized for production
func ProductionConfig() GenealogyConfig {
	config := DefaultGenealogyConfig()
	config.CacheTTL = 10 * time.Minute
	config.RefreshInterval = 5 * time.Minute
	config.RelationshipTTL = 7 * 24 * time.Hour // 1 week
	config.TreeCacheTTL = 30 * time.Minute
	config.BackgroundRefresh = true
	config.EnableCaching = true
	config.LazyLoading = true
	config.MaxTreeSize = 50000
	return config
}

// TestingConfig returns a configuration optimized for testing
func TestingConfig() GenealogyConfig {
	config := DefaultGenealogyConfig()
	config.CacheTTL = 1 * time.Second
	config.RefreshInterval = 100 * time.Millisecond
	config.RelationshipTTL = 5 * time.Minute
	config.TreeCacheTTL = 10 * time.Second
	config.CleanupInterval = 30 * time.Second
	config.BackgroundRefresh = false
	config.EnableCaching = false
	config.LazyLoading = false
	config.MaxTreeSize = 1000
	return config
}