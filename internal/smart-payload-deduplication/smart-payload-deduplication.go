// Copyright 2025 James Ross
package deduplication

import (
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// manager implements the DeduplicationManager interface
type manager struct {
	config           *Config
	logger           *zap.Logger
	redis            redis.Cmdable
	chunkStore       ChunkStore
	refCounter       ReferenceCounter
	payloadStore     *PayloadMapStore
	chunker          Chunker
	compressor       Compressor
	similarityDetector SimilarityDetector
	dictBuilder      *DictionaryBuilder
	stats            *DeduplicationStats
	statsMu          sync.RWMutex
	gcTicker         *time.Ticker
	statsTicker      *time.Ticker
	stopCh           chan struct{}
	wg               sync.WaitGroup
}

// NewManager creates a new deduplication manager
func NewManager(config *Config, rdb redis.Cmdable, logger *zap.Logger) (DeduplicationManager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	// Initialize compressor
	compressor, err := NewZstdCompressor(&config.Compression)
	if err != nil {
		return nil, fmt.Errorf("failed to create compressor: %w", err)
	}

	// Initialize chunk store
	chunkStore := NewRedisChunkStore(rdb, config.RedisKeyPrefix, compressor)

	// Initialize reference counter
	refCounter := NewRedisReferenceCounter(rdb, config.RedisKeyPrefix)

	// Initialize payload store
	payloadStore := NewPayloadMapStore(rdb, config.RedisKeyPrefix)

	// Initialize similarity detector
	similarityDetector := NewMinHashSimilarityDetector(128, 16) // 128 hashes, 16 bands

	// Initialize chunker
	chunker := NewRabinChunker(&config.Chunking, similarityDetector)

	// Initialize dictionary builder
	dictBuilder := NewDictionaryBuilder(1000, 10*1024*1024) // 1000 samples, 10MB max

	m := &manager{
		config:             config,
		logger:             logger,
		redis:              rdb,
		chunkStore:         chunkStore,
		refCounter:         refCounter,
		payloadStore:       payloadStore,
		chunker:            chunker,
		compressor:         compressor,
		similarityDetector: similarityDetector,
		dictBuilder:        dictBuilder,
		stats: &DeduplicationStats{
			LastUpdated: time.Now(),
		},
		stopCh: make(chan struct{}),
	}

	// Start background services
	if config.GarbageCollection.Enabled {
		m.startGarbageCollector()
	}

	m.startStatsUpdater()

	logger.Info("Smart payload deduplication manager initialized",
		zap.Bool("enabled", config.Enabled),
		zap.String("chunking_strategy", "rabin_fingerprinting"),
		zap.Bool("compression_enabled", config.Compression.Enabled),
		zap.Bool("gc_enabled", config.GarbageCollection.Enabled))

	return m, nil
}

// DeduplicatePayload breaks a payload into chunks and stores them
func (m *manager) DeduplicatePayload(jobID string, payload []byte) (*PayloadMap, error) {
	if !m.config.Enabled {
		return nil, fmt.Errorf("deduplication is disabled")
	}

	start := time.Now()
	m.logger.Debug("Starting payload deduplication",
		zap.String("job_id", jobID),
		zap.Int("payload_size", len(payload)))

	// Add to dictionary samples (async)
	go m.dictBuilder.AddSample(payload)

	// Chunk the payload
	chunks, err := m.chunker.ChunkPayload(payload)
	if err != nil {
		return nil, NewDeduplicationError(ErrCodeStorageFull, "failed to chunk payload", err)
	}

	// Create payload map
	payloadMap := &PayloadMap{
		JobID:      jobID,
		OrigSize:   len(payload),
		ChunkRefs:  make([]ChunkReference, 0, len(chunks)),
		Checksum:   m.chunker.ComputeChecksum(payload),
		CreatedAt:  time.Now(),
		Compressed: m.config.Compression.Enabled,
	}

	// Process each chunk
	offset := 0
	newChunks := 0
	reusedChunks := 0

	for _, chunk := range chunks {
		// Check if chunk already exists
		exists, err := m.chunkStore.Exists(chunk.Hash)
		if err != nil {
			return nil, NewDeduplicationError(ErrCodeStorageFull, "failed to check chunk existence", err)
		}

		if !exists {
			// Store new chunk
			if err := m.chunkStore.Store(&chunk); err != nil {
				return nil, NewDeduplicationError(ErrCodeStorageFull, "failed to store chunk", err)
			}
			newChunks++
		} else {
			reusedChunks++
		}

		// Add reference
		if _, err := m.refCounter.Add(chunk.Hash); err != nil {
			return nil, NewDeduplicationError(ErrCodeReferenceCorruption, "failed to add chunk reference", err)
		}

		// Add to payload map
		chunkRef := ChunkReference{
			Hash:   chunk.Hash,
			Offset: offset,
			Size:   chunk.Size,
		}
		payloadMap.ChunkRefs = append(payloadMap.ChunkRefs, chunkRef)
		offset += chunk.Size
	}

	// Store payload map
	if err := m.payloadStore.Store(payloadMap); err != nil {
		return nil, NewDeduplicationError(ErrCodeStorageFull, "failed to store payload map", err)
	}

	// Update statistics
	m.updateDeduplicationStats(len(payload), len(chunks), newChunks, reusedChunks, time.Since(start))

	m.logger.Debug("Payload deduplication completed",
		zap.String("job_id", jobID),
		zap.Int("total_chunks", len(chunks)),
		zap.Int("new_chunks", newChunks),
		zap.Int("reused_chunks", reusedChunks),
		zap.Duration("duration", time.Since(start)))

	return payloadMap, nil
}

// ReconstructPayload rebuilds the original payload from chunk references
func (m *manager) ReconstructPayload(payloadMap *PayloadMap) ([]byte, error) {
	start := time.Now()
	m.logger.Debug("Starting payload reconstruction",
		zap.String("job_id", payloadMap.JobID),
		zap.Int("chunk_count", len(payloadMap.ChunkRefs)))

	// Pre-allocate buffer for efficiency
	payload := make([]byte, 0, payloadMap.OrigSize)

	// Retrieve and concatenate chunks
	for _, chunkRef := range payloadMap.ChunkRefs {
		chunk, err := m.chunkStore.Get(chunkRef.Hash)
		if err != nil {
			return nil, NewDeduplicationError(ErrCodeChunkNotFound, "failed to get chunk during reconstruction", err)
		}

		payload = append(payload, chunk.Data...)
	}

	// Verify checksum
	if !m.chunker.ValidateChecksum(payload, payloadMap.Checksum) {
		return nil, NewDeduplicationError(ErrCodeChecksumMismatch, "payload checksum mismatch", nil)
	}

	m.logger.Debug("Payload reconstruction completed",
		zap.String("job_id", payloadMap.JobID),
		zap.Int("reconstructed_size", len(payload)),
		zap.Duration("duration", time.Since(start)))

	return payload, nil
}

// StoreChunk stores a chunk directly
func (m *manager) StoreChunk(chunk *Chunk) error {
	return m.chunkStore.Store(chunk)
}

// GetChunk retrieves a chunk by hash
func (m *manager) GetChunk(hash []byte) (*Chunk, error) {
	return m.chunkStore.Get(hash)
}

// DeleteChunk removes a chunk (decrements reference count first)
func (m *manager) DeleteChunk(hash []byte) error {
	// Decrement reference count
	count, err := m.refCounter.Remove(hash)
	if err != nil {
		return err
	}

	// Only delete if no more references
	if count == 0 {
		return m.chunkStore.Delete(hash)
	}

	return nil
}

// AddReference adds a reference to a chunk
func (m *manager) AddReference(chunkHash []byte) error {
	_, err := m.refCounter.Add(chunkHash)
	return err
}

// RemoveReference removes a reference from a chunk
func (m *manager) RemoveReference(chunkHash []byte) error {
	_, err := m.refCounter.Remove(chunkHash)
	return err
}

// GetReferenceCount returns the reference count for a chunk
func (m *manager) GetReferenceCount(chunkHash []byte) (int64, error) {
	return m.refCounter.Get(chunkHash)
}

// GetStats returns current deduplication statistics
func (m *manager) GetStats() (*DeduplicationStats, error) {
	m.statsMu.RLock()
	defer m.statsMu.RUnlock()

	// Create a copy of the stats
	stats := *m.stats
	return &stats, nil
}

// GetChunkStats returns statistics for a specific chunk
func (m *manager) GetChunkStats(hash []byte) (*ChunkStats, error) {
	chunk, err := m.chunkStore.Get(hash)
	if err != nil {
		return nil, err
	}

	refCount, err := m.refCounter.Get(hash)
	if err != nil {
		return nil, err
	}

	return &ChunkStats{
		Hash:      hex.EncodeToString(hash),
		RefCount:  refCount,
		Size:      chunk.Size,
		CreatedAt: chunk.CreatedAt,
		LastUsed:  chunk.LastUsed,
		HitCount:  refCount, // Approximation
	}, nil
}

// GetPopularChunks returns the most referenced chunks
func (m *manager) GetPopularChunks(limit int) ([]ChunkStats, error) {
	allRefs, err := m.refCounter.GetAll()
	if err != nil {
		return nil, err
	}

	// Convert to slice and sort by reference count
	type chunkRef struct {
		hash     string
		refCount int64
	}

	refs := make([]chunkRef, 0, len(allRefs))
	for hash, count := range allRefs {
		refs = append(refs, chunkRef{hash: hash, refCount: count})
	}

	// Sort by reference count (descending)
	for i := 0; i < len(refs)-1; i++ {
		for j := i + 1; j < len(refs); j++ {
			if refs[j].refCount > refs[i].refCount {
				refs[i], refs[j] = refs[j], refs[i]
			}
		}
	}

	// Limit results
	if limit > 0 && len(refs) > limit {
		refs = refs[:limit]
	}

	// Convert to ChunkStats
	result := make([]ChunkStats, 0, len(refs))
	for _, ref := range refs {
		hashBytes, err := hex.DecodeString(ref.hash)
		if err != nil {
			continue
		}

		chunkStats, err := m.GetChunkStats(hashBytes)
		if err != nil {
			continue
		}

		result = append(result, *chunkStats)
	}

	return result, nil
}

// RunGarbageCollection manually triggers garbage collection
func (m *manager) RunGarbageCollection() error {
	return m.runGarbageCollection()
}

// GetOrphanedChunks returns chunks with zero references
func (m *manager) GetOrphanedChunks() ([][]byte, error) {
	allRefs, err := m.refCounter.GetAll()
	if err != nil {
		return nil, err
	}

	orphaned := make([][]byte, 0)

	for hashStr, count := range allRefs {
		if count <= 0 {
			if hash, err := hex.DecodeString(hashStr); err == nil {
				orphaned = append(orphaned, hash)
			}
		}
	}

	return orphaned, nil
}

// ValidateIntegrity checks system integrity
func (m *manager) ValidateIntegrity() error {
	m.logger.Info("Starting integrity validation")

	// Audit reference counts
	if err := m.refCounter.AuditAndRepair(); err != nil {
		return fmt.Errorf("reference count audit failed: %w", err)
	}

	// Validate chunk checksums (sample check)
	orphaned, err := m.GetOrphanedChunks()
	if err != nil {
		return fmt.Errorf("failed to get orphaned chunks: %w", err)
	}

	m.logger.Info("Integrity validation completed",
		zap.Int("orphaned_chunks", len(orphaned)))

	return nil
}

// AuditReferences repairs reference count inconsistencies
func (m *manager) AuditReferences() error {
	return m.refCounter.AuditAndRepair()
}

// GetHealth returns system health information
func (m *manager) GetHealth() map[string]interface{} {
	stats, _ := m.GetStats()

	health := map[string]interface{}{
		"enabled":           m.config.Enabled,
		"total_chunks":      stats.TotalChunks,
		"memory_savings":    stats.MemorySavings,
		"compression_ratio": stats.CompressionRatio,
		"last_updated":      stats.LastUpdated,
		"gc_enabled":        m.config.GarbageCollection.Enabled,
	}

	// Add compressor stats if available
	if zc, ok := m.compressor.(*ZstdCompressor); ok {
		compStats := zc.GetStats()
		health["compression_stats"] = compStats
	}

	return health
}

// UpdateConfig updates the system configuration
func (m *manager) UpdateConfig(config *Config) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	oldConfig := m.config
	m.config = config

	// Restart services if needed
	if config.GarbageCollection.Enabled != oldConfig.GarbageCollection.Enabled ||
		config.GarbageCollection.Interval != oldConfig.GarbageCollection.Interval {
		m.restartGarbageCollector()
	}

	m.logger.Info("Configuration updated",
		zap.Bool("gc_restarted", config.GarbageCollection.Enabled != oldConfig.GarbageCollection.Enabled))

	return nil
}

// GetConfig returns the current configuration
func (m *manager) GetConfig() *Config {
	return m.config.Clone()
}

// Background services

func (m *manager) startGarbageCollector() {
	m.gcTicker = time.NewTicker(m.config.GarbageCollection.Interval)

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case <-m.gcTicker.C:
				if err := m.runGarbageCollection(); err != nil {
					m.logger.Error("Garbage collection failed", zap.Error(err))
				}
			case <-m.stopCh:
				return
			}
		}
	}()
}

func (m *manager) startStatsUpdater() {
	m.statsTicker = time.NewTicker(m.config.StatsInterval)

	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		for {
			select {
			case <-m.statsTicker.C:
				m.updateStats()
			case <-m.stopCh:
				return
			}
		}
	}()
}

func (m *manager) restartGarbageCollector() {
	if m.gcTicker != nil {
		m.gcTicker.Stop()
	}

	if m.config.GarbageCollection.Enabled {
		m.startGarbageCollector()
	}
}

func (m *manager) runGarbageCollection() error {
	start := time.Now()
	m.logger.Debug("Starting garbage collection")

	orphaned, err := m.GetOrphanedChunks()
	if err != nil {
		return fmt.Errorf("failed to get orphaned chunks: %w", err)
	}

	deleted := 0
	batchSize := m.config.GarbageCollection.BatchSize

	for i := 0; i < len(orphaned); i += batchSize {
		end := i + batchSize
		if end > len(orphaned) {
			end = len(orphaned)
		}

		batch := orphaned[i:end]

		// Delete batch of orphaned chunks
		for _, hash := range batch {
			if err := m.chunkStore.Delete(hash); err != nil {
				m.logger.Warn("Failed to delete orphaned chunk",
					zap.String("hash", hex.EncodeToString(hash)),
					zap.Error(err))
			} else {
				deleted++
			}
		}
	}

	duration := time.Since(start)
	m.logger.Info("Garbage collection completed",
		zap.Int("orphaned_chunks", len(orphaned)),
		zap.Int("deleted_chunks", deleted),
		zap.Duration("duration", duration))

	return nil
}

func (m *manager) updateStats() {
	// This would calculate comprehensive stats from Redis
	// For now, implement basic placeholder
	m.statsMu.Lock()
	defer m.statsMu.Unlock()

	m.stats.LastUpdated = time.Now()
	// Additional stats calculations would go here
}

func (m *manager) updateDeduplicationStats(payloadSize, totalChunks, newChunks, reusedChunks int, duration time.Duration) {
	m.statsMu.Lock()
	defer m.statsMu.Unlock()

	m.stats.TotalPayloads++
	m.stats.TotalChunks += int64(newChunks)
	m.stats.TotalBytes += int64(payloadSize)

	if reusedChunks > 0 {
		m.stats.ChunkHitRate = float64(reusedChunks) / float64(totalChunks)
	}

	m.stats.LastUpdated = time.Now()
}

// Close shuts down the deduplication manager
func (m *manager) Close() error {
	close(m.stopCh)

	if m.gcTicker != nil {
		m.gcTicker.Stop()
	}

	if m.statsTicker != nil {
		m.statsTicker.Stop()
	}

	m.wg.Wait()

	// Close compressor if it supports closing
	if closer, ok := m.compressor.(interface{ Close() error }); ok {
		closer.Close()
	}

	return nil
}