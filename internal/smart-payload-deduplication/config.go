// Copyright 2025 James Ross
package deduplication

import (
	"encoding/json"
	"fmt"
	"time"
)

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Enabled:        true,
		RedisKeyPrefix: "go-redis-work-queue:dedup:",
		RedisDB:        0,
		Chunking: ChunkingConfig{
			MinChunkSize:        DefaultMinChunkSize,
			MaxChunkSize:        DefaultMaxChunkSize,
			AvgChunkSize:        DefaultAvgChunkSize,
			WindowSize:          DefaultWindowSize,
			Polynomial:          DefaultPolynomial,
			SimilarityThreshold: 0.8,
		},
		Compression: CompressionConfig{
			Enabled:       true,
			Level:         DefaultCompressionLevel,
			DictionarySize: DefaultDictionarySize,
			UseDictionary: true,
		},
		GarbageCollection: GarbageCollectionConfig{
			Enabled:           true,
			Interval:          DefaultGCInterval,
			OrphanThreshold:   DefaultOrphanThreshold,
			BatchSize:         DefaultBatchSize,
			ConcurrentWorkers: DefaultConcurrentWorkers,
		},
		SafetyMode:     true,
		MigrationRatio: 1.0, // 100% migration by default
		MaxMemoryMB:    DefaultMaxMemoryMB,
		StatsInterval:  DefaultStatsInterval,
	}
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Chunking.MinChunkSize <= 0 {
		return fmt.Errorf("min_chunk_size must be positive, got %d", c.Chunking.MinChunkSize)
	}

	if c.Chunking.MaxChunkSize <= c.Chunking.MinChunkSize {
		return fmt.Errorf("max_chunk_size (%d) must be greater than min_chunk_size (%d)",
			c.Chunking.MaxChunkSize, c.Chunking.MinChunkSize)
	}

	if c.Chunking.AvgChunkSize < c.Chunking.MinChunkSize ||
		c.Chunking.AvgChunkSize > c.Chunking.MaxChunkSize {
		return fmt.Errorf("avg_chunk_size (%d) must be between min (%d) and max (%d)",
			c.Chunking.AvgChunkSize, c.Chunking.MinChunkSize, c.Chunking.MaxChunkSize)
	}

	if c.Chunking.WindowSize <= 0 || c.Chunking.WindowSize > 256 {
		return fmt.Errorf("window_size must be between 1 and 256, got %d", c.Chunking.WindowSize)
	}

	if c.Chunking.SimilarityThreshold < 0.0 || c.Chunking.SimilarityThreshold > 1.0 {
		return fmt.Errorf("similarity_threshold must be between 0.0 and 1.0, got %.2f",
			c.Chunking.SimilarityThreshold)
	}

	if c.Compression.Enabled {
		if c.Compression.Level < 1 || c.Compression.Level > 19 {
			return fmt.Errorf("compression_level must be between 1 and 19, got %d",
				c.Compression.Level)
		}

		if c.Compression.DictionarySize <= 0 || c.Compression.DictionarySize > 1<<20 {
			return fmt.Errorf("dictionary_size must be between 1 and 1MB, got %d",
				c.Compression.DictionarySize)
		}
	}

	if c.GarbageCollection.Enabled {
		if c.GarbageCollection.Interval <= 0 {
			return fmt.Errorf("gc_interval must be positive, got %v",
				c.GarbageCollection.Interval)
		}

		if c.GarbageCollection.OrphanThreshold <= 0 {
			return fmt.Errorf("orphan_threshold must be positive, got %v",
				c.GarbageCollection.OrphanThreshold)
		}

		if c.GarbageCollection.BatchSize <= 0 {
			return fmt.Errorf("gc_batch_size must be positive, got %d",
				c.GarbageCollection.BatchSize)
		}

		if c.GarbageCollection.ConcurrentWorkers <= 0 {
			return fmt.Errorf("concurrent_workers must be positive, got %d",
				c.GarbageCollection.ConcurrentWorkers)
		}
	}

	if c.MigrationRatio < 0.0 || c.MigrationRatio > 1.0 {
		return fmt.Errorf("migration_ratio must be between 0.0 and 1.0, got %.2f",
			c.MigrationRatio)
	}

	if c.MaxMemoryMB <= 0 {
		return fmt.Errorf("max_memory_mb must be positive, got %d", c.MaxMemoryMB)
	}

	if c.StatsInterval <= 0 {
		return fmt.Errorf("stats_interval must be positive, got %v", c.StatsInterval)
	}

	return nil
}

// ToJSON serializes the configuration to JSON
func (c *Config) ToJSON() ([]byte, error) {
	return json.MarshalIndent(c, "", "  ")
}

// FromJSON deserializes configuration from JSON
func (c *Config) FromJSON(data []byte) error {
	return json.Unmarshal(data, c)
}

// Clone creates a deep copy of the configuration
func (c *Config) Clone() *Config {
	clone := *c
	return &clone
}

// GetChunkingBits calculates the number of bits for chunk boundary detection
func (c *Config) GetChunkingBits() int {
	// Calculate bits needed to achieve target average chunk size
	// Using: avg_chunk_size = 2^bits
	bits := 0
	target := c.Chunking.AvgChunkSize
	for (1 << bits) < target {
		bits++
	}
	return bits
}

// GetChunkingMask returns the mask for chunk boundary detection
func (c *Config) GetChunkingMask() uint64 {
	bits := c.GetChunkingBits()
	return (1 << bits) - 1
}

// EstimateMemoryUsage estimates memory usage based on configuration
func (c *Config) EstimateMemoryUsage(numChunks int64, avgPayloadSize int) int64 {
	// Estimate based on chunk size, compression ratio, and metadata overhead
	avgChunkSize := int64(c.Chunking.AvgChunkSize)
	compressionRatio := 0.7 // Assume 70% compression ratio

	chunkDataSize := numChunks * avgChunkSize
	if c.Compression.Enabled {
		chunkDataSize = int64(float64(chunkDataSize) * compressionRatio)
	}

	// Add metadata overhead (hash + reference counts + stats)
	metadataPerChunk := int64(32 + 8 + 64) // 32-byte hash + 8-byte refcount + 64-byte metadata
	metadataSize := numChunks * metadataPerChunk

	// Add payload map overhead
	avgChunksPerPayload := int64(avgPayloadSize / c.Chunking.AvgChunkSize)
	payloadMapSize := numChunks / avgChunksPerPayload * 256 // 256 bytes per payload map

	totalSize := chunkDataSize + metadataSize + payloadMapSize

	return totalSize
}

// OptimizeForWorkload adjusts configuration based on workload characteristics
func (c *Config) OptimizeForWorkload(avgPayloadSize int, repetitionRate float64) {
	// Adjust chunk size based on payload size
	if avgPayloadSize < 4096 {
		// Small payloads - use smaller chunks
		c.Chunking.MinChunkSize = 256
		c.Chunking.AvgChunkSize = 1024
		c.Chunking.MaxChunkSize = 8192
	} else if avgPayloadSize > 1<<20 {
		// Large payloads - use larger chunks
		c.Chunking.MinChunkSize = 8192
		c.Chunking.AvgChunkSize = 32768
		c.Chunking.MaxChunkSize = 262144
	}

	// Adjust similarity threshold based on repetition rate
	if repetitionRate > 0.8 {
		// High repetition - more aggressive similarity detection
		c.Chunking.SimilarityThreshold = 0.7
	} else if repetitionRate < 0.3 {
		// Low repetition - less aggressive similarity detection
		c.Chunking.SimilarityThreshold = 0.9
	}

	// Adjust GC frequency based on workload
	if repetitionRate > 0.8 {
		// High repetition - less frequent GC (chunks are reused)
		c.GarbageCollection.Interval = 4 * time.Hour
		c.GarbageCollection.OrphanThreshold = 48 * time.Hour
	} else {
		// Low repetition - more frequent GC (chunks become orphaned quickly)
		c.GarbageCollection.Interval = 30 * time.Minute
		c.GarbageCollection.OrphanThreshold = 6 * time.Hour
	}
}

// ConfigurationManager handles dynamic configuration updates
type ConfigurationManager struct {
	config   *Config
	handlers []ConfigChangeHandler
}

// ConfigChangeHandler is called when configuration changes
type ConfigChangeHandler func(old, new *Config) error

// NewConfigurationManager creates a new configuration manager
func NewConfigurationManager(config *Config) *ConfigurationManager {
	return &ConfigurationManager{
		config:   config.Clone(),
		handlers: make([]ConfigChangeHandler, 0),
	}
}

// GetConfig returns the current configuration
func (cm *ConfigurationManager) GetConfig() *Config {
	return cm.config.Clone()
}

// UpdateConfig updates the configuration and notifies handlers
func (cm *ConfigurationManager) UpdateConfig(newConfig *Config) error {
	if err := newConfig.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	oldConfig := cm.config.Clone()
	cm.config = newConfig.Clone()

	// Notify all handlers
	for _, handler := range cm.handlers {
		if err := handler(oldConfig, cm.config); err != nil {
			// Rollback on error
			cm.config = oldConfig
			return fmt.Errorf("configuration update failed: %w", err)
		}
	}

	return nil
}

// AddConfigChangeHandler adds a handler for configuration changes
func (cm *ConfigurationManager) AddConfigChangeHandler(handler ConfigChangeHandler) {
	cm.handlers = append(cm.handlers, handler)
}

// ConfigDiff represents the difference between two configurations
type ConfigDiff struct {
	ChunkingChanged         bool `json:"chunking_changed"`
	CompressionChanged      bool `json:"compression_changed"`
	GarbageCollectionChanged bool `json:"garbage_collection_changed"`
	SafetyModeChanged       bool `json:"safety_mode_changed"`
	MigrationRatioChanged   bool `json:"migration_ratio_changed"`
	MemoryLimitChanged      bool `json:"memory_limit_changed"`
}

// DiffConfigs computes the difference between two configurations
func DiffConfigs(old, new *Config) *ConfigDiff {
	return &ConfigDiff{
		ChunkingChanged:          old.Chunking != new.Chunking,
		CompressionChanged:       old.Compression != new.Compression,
		GarbageCollectionChanged: old.GarbageCollection != new.GarbageCollection,
		SafetyModeChanged:        old.SafetyMode != new.SafetyMode,
		MigrationRatioChanged:    old.MigrationRatio != new.MigrationRatio,
		MemoryLimitChanged:       old.MaxMemoryMB != new.MaxMemoryMB,
	}
}

// HasSignificantChanges returns true if the configuration changes require system restart
func (cd *ConfigDiff) HasSignificantChanges() bool {
	return cd.ChunkingChanged || cd.CompressionChanged
}

// GetChangeSummary returns a human-readable summary of changes
func (cd *ConfigDiff) GetChangeSummary() string {
	changes := []string{}

	if cd.ChunkingChanged {
		changes = append(changes, "chunking configuration")
	}
	if cd.CompressionChanged {
		changes = append(changes, "compression settings")
	}
	if cd.GarbageCollectionChanged {
		changes = append(changes, "garbage collection")
	}
	if cd.SafetyModeChanged {
		changes = append(changes, "safety mode")
	}
	if cd.MigrationRatioChanged {
		changes = append(changes, "migration ratio")
	}
	if cd.MemoryLimitChanged {
		changes = append(changes, "memory limits")
	}

	if len(changes) == 0 {
		return "no changes"
	}

	return fmt.Sprintf("changed: %v", changes)
}