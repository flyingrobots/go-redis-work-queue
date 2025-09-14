// Copyright 2025 James Ross
package exactlyonce

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// MemoryIdempotencyStorage implements IdempotencyStorage using in-memory storage
// This is primarily for testing and development environments
type MemoryIdempotencyStorage struct {
	cfg     *Config
	log     *zap.Logger
	data    map[string]*memoryEntry
	stats   map[string]*memoryStats
	mu      sync.RWMutex
	stopCh  chan struct{}
	stopped bool
}

type memoryEntry struct {
	Value     interface{}
	CreatedAt time.Time
	ExpiresAt time.Time
}

type memoryStats struct {
	QueueName         string
	TenantID          string
	TotalRequests     int64
	Hits              int64
	LastUpdated       time.Time
}

// NewMemoryIdempotencyStorage creates a new memory-based idempotency storage
func NewMemoryIdempotencyStorage(cfg *Config, log *zap.Logger) *MemoryIdempotencyStorage {
	m := &MemoryIdempotencyStorage{
		cfg:    cfg,
		log:    log,
		data:   make(map[string]*memoryEntry),
		stats:  make(map[string]*memoryStats),
		stopCh: make(chan struct{}),
	}

	// Start cleanup goroutine
	go m.cleanupLoop()

	return m
}

// Check verifies if an idempotency key has been processed before
func (m *MemoryIdempotencyStorage) Check(ctx context.Context, key IdempotencyKey) (*IdempotencyResult, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.stopped {
		return nil, fmt.Errorf("storage is stopped")
	}

	memKey := m.buildKey(key)
	statsKey := m.buildStatsKey(key.QueueName, key.TenantID)

	// Update stats
	if stats, ok := m.stats[statsKey]; ok {
		stats.TotalRequests++
		stats.LastUpdated = time.Now().UTC()
	} else {
		m.stats[statsKey] = &memoryStats{
			QueueName:     key.QueueName,
			TenantID:      key.TenantID,
			TotalRequests: 1,
			Hits:          0,
			LastUpdated:   time.Now().UTC(),
		}
	}

	entry, exists := m.data[memKey]
	if !exists {
		// First time processing
		return &IdempotencyResult{
			IsFirstTime: true,
			Key:         memKey,
		}, nil
	}

	// Check if entry has expired
	if time.Now().After(entry.ExpiresAt) {
		// Expired - treat as first time
		delete(m.data, memKey)
		return &IdempotencyResult{
			IsFirstTime: true,
			Key:         memKey,
		}, nil
	}

	// Entry exists and is valid - mark as hit in stats
	if stats := m.stats[statsKey]; stats != nil {
		stats.Hits++
	}

	return &IdempotencyResult{
		IsFirstTime:   false,
		ExistingValue: entry.Value,
		Key:           memKey,
	}, nil
}

// Set marks an idempotency key as processed with optional result value
func (m *MemoryIdempotencyStorage) Set(ctx context.Context, key IdempotencyKey, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stopped {
		return fmt.Errorf("storage is stopped")
	}

	memKey := m.buildKey(key)

	// Check if we're at capacity
	maxKeys := m.cfg.Idempotency.Storage.Memory.MaxKeys
	if maxKeys > 0 && len(m.data) >= maxKeys {
		if err := m.evictEntries(1); err != nil {
			return fmt.Errorf("failed to evict entries: %w", err)
		}
	}

	// Store the entry
	m.data[memKey] = &memoryEntry{
		Value:     value,
		CreatedAt: key.CreatedAt,
		ExpiresAt: key.CreatedAt.Add(key.TTL),
	}

	return nil
}

// Delete removes an idempotency key from storage
func (m *MemoryIdempotencyStorage) Delete(ctx context.Context, key IdempotencyKey) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stopped {
		return fmt.Errorf("storage is stopped")
	}

	memKey := m.buildKey(key)
	delete(m.data, memKey)

	return nil
}

// Stats returns statistics about the deduplication store
func (m *MemoryIdempotencyStorage) Stats(ctx context.Context, queueName, tenantID string) (*DedupStats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.stopped {
		return nil, fmt.Errorf("storage is stopped")
	}

	statsKey := m.buildStatsKey(queueName, tenantID)

	// Count total keys for this queue/tenant
	totalKeys := int64(0)
	for key := range m.data {
		if m.keyMatchesQueueTenant(key, queueName, tenantID) {
			totalKeys++
		}
	}

	stats := m.stats[statsKey]
	if stats == nil {
		stats = &memoryStats{
			QueueName:     queueName,
			TenantID:      tenantID,
			TotalRequests: 0,
			Hits:          0,
			LastUpdated:   time.Now().UTC(),
		}
	}

	hitRate := 0.0
	if stats.TotalRequests > 0 {
		hitRate = float64(stats.Hits) / float64(stats.TotalRequests)
	}

	return &DedupStats{
		QueueName:         queueName,
		TenantID:          tenantID,
		TotalKeys:         totalKeys,
		HitRate:           hitRate,
		TotalRequests:     stats.TotalRequests,
		DuplicatesAvoided: stats.Hits,
		LastUpdated:       stats.LastUpdated,
	}, nil
}

// Close stops the storage and cleans up resources
func (m *MemoryIdempotencyStorage) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.stopped {
		close(m.stopCh)
		m.stopped = true
		m.data = nil
		m.stats = nil
	}

	return nil
}

// cleanupLoop periodically removes expired entries
func (m *MemoryIdempotencyStorage) cleanupLoop() {
	ticker := time.NewTicker(m.cfg.Idempotency.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.cleanupExpired()
		}
	}
}

// cleanupExpired removes expired entries
func (m *MemoryIdempotencyStorage) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	toDelete := make([]string, 0)

	for key, entry := range m.data {
		if now.After(entry.ExpiresAt) {
			toDelete = append(toDelete, key)
		}
	}

	for _, key := range toDelete {
		delete(m.data, key)
	}

	if len(toDelete) > 0 {
		m.log.Debug("Cleaned up expired idempotency keys",
			zap.Int("count", len(toDelete)))
	}
}

// evictEntries removes entries based on eviction policy
func (m *MemoryIdempotencyStorage) evictEntries(count int) error {
	if count <= 0 {
		return nil
	}

	policy := m.cfg.Idempotency.Storage.Memory.EvictionPolicy

	switch policy {
	case "lru":
		return m.evictLRU(count)
	case "fifo":
		return m.evictFIFO(count)
	default:
		return m.evictOldest(count)
	}
}

// evictLRU evicts least recently used entries (simplified - just uses creation time)
func (m *MemoryIdempotencyStorage) evictLRU(count int) error {
	return m.evictOldest(count)
}

// evictFIFO evicts first-in-first-out entries
func (m *MemoryIdempotencyStorage) evictFIFO(count int) error {
	return m.evictOldest(count)
}

// evictOldest evicts the oldest entries
func (m *MemoryIdempotencyStorage) evictOldest(count int) error {
	type entryWithKey struct {
		key   string
		entry *memoryEntry
	}

	entries := make([]entryWithKey, 0, len(m.data))
	for key, entry := range m.data {
		entries = append(entries, entryWithKey{key: key, entry: entry})
	}

	// Sort by creation time (oldest first)
	for i := 0; i < len(entries)-1; i++ {
		for j := i + 1; j < len(entries); j++ {
			if entries[i].entry.CreatedAt.After(entries[j].entry.CreatedAt) {
				entries[i], entries[j] = entries[j], entries[i]
			}
		}
	}

	// Remove the oldest entries
	evicted := 0
	for _, entry := range entries {
		if evicted >= count {
			break
		}
		delete(m.data, entry.key)
		evicted++
	}

	m.log.Debug("Evicted idempotency keys", zap.Int("count", evicted))
	return nil
}

// buildKey constructs the memory key for an idempotency key
func (m *MemoryIdempotencyStorage) buildKey(key IdempotencyKey) string {
	if key.TenantID != "" {
		return fmt.Sprintf("%s:%s:%s", key.QueueName, key.TenantID, key.ID)
	}
	return fmt.Sprintf("%s:%s", key.QueueName, key.ID)
}

// buildStatsKey constructs the stats key for a queue/tenant
func (m *MemoryIdempotencyStorage) buildStatsKey(queueName, tenantID string) string {
	if tenantID != "" {
		return fmt.Sprintf("stats:%s:%s", queueName, tenantID)
	}
	return fmt.Sprintf("stats:%s", queueName)
}

// keyMatchesQueueTenant checks if a key matches the given queue/tenant
func (m *MemoryIdempotencyStorage) keyMatchesQueueTenant(key, queueName, tenantID string) bool {
	expectedPrefix := queueName + ":"
	if tenantID != "" {
		expectedPrefix = queueName + ":" + tenantID + ":"
	}

	return len(key) > len(expectedPrefix) && key[:len(expectedPrefix)] == expectedPrefix
}