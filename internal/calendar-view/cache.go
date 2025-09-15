package calendarview

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheManager manages caching for calendar data
type CacheManager struct {
	cache         map[string]*CacheEntry
	mutex         sync.RWMutex
	ttl           time.Duration
	cleanupTicker *time.Ticker
	stopCleanup   chan bool
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Key       string      `json:"key"`
	Data      interface{} `json:"data"`
	CreatedAt time.Time   `json:"created_at"`
	ExpiresAt time.Time   `json:"expires_at"`
	HitCount  int         `json:"hit_count"`
	LastHit   time.Time   `json:"last_hit"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	TotalEntries int           `json:"total_entries"`
	HitRate      float64       `json:"hit_rate"`
	MissRate     float64       `json:"miss_rate"`
	TotalHits    int64         `json:"total_hits"`
	TotalMisses  int64         `json:"total_misses"`
	MemoryUsage  int64         `json:"memory_usage"`
	OldestEntry  *time.Time    `json:"oldest_entry"`
	NewestEntry  *time.Time    `json:"newest_entry"`
	AvgTTL       time.Duration `json:"avg_ttl"`
}

// NewCacheManager creates a new cache manager
func NewCacheManager(ttl time.Duration) *CacheManager {
	if ttl <= 0 {
		ttl = 5 * time.Minute // Default TTL
	}

	cm := &CacheManager{
		cache:         make(map[string]*CacheEntry),
		ttl:           ttl,
		cleanupTicker: time.NewTicker(ttl / 2), // Cleanup every half TTL
		stopCleanup:   make(chan bool),
	}

	// Start cleanup goroutine
	go cm.cleanup()

	return cm
}

// Get retrieves an item from the cache
func (cm *CacheManager) Get(key string) (interface{}, bool) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	entry, exists := cm.cache[key]
	if !exists {
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		// Don't remove here, let cleanup handle it
		return nil, false
	}

	// Update hit statistics
	entry.HitCount++
	entry.LastHit = time.Now()

	return entry.Data, true
}

// Set stores an item in the cache
func (cm *CacheManager) Set(key string, data interface{}) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	entry := &CacheEntry{
		Key:       key,
		Data:      data,
		CreatedAt: now,
		ExpiresAt: now.Add(cm.ttl),
		HitCount:  0,
		LastHit:   now,
	}

	cm.cache[key] = entry
}

// SetWithTTL stores an item in the cache with a custom TTL
func (cm *CacheManager) SetWithTTL(key string, data interface{}, ttl time.Duration) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	entry := &CacheEntry{
		Key:       key,
		Data:      data,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
		HitCount:  0,
		LastHit:   now,
	}

	cm.cache[key] = entry
}

// Delete removes an item from the cache
func (cm *CacheManager) Delete(key string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	delete(cm.cache, key)
}

// Invalidate removes all items from the cache
func (cm *CacheManager) Invalidate() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	cm.cache = make(map[string]*CacheEntry)
}

// InvalidatePattern removes items matching a pattern
func (cm *CacheManager) InvalidatePattern(pattern string) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	toDelete := make([]string, 0)
	for key := range cm.cache {
		if matchesPattern(key, pattern) {
			toDelete = append(toDelete, key)
		}
	}

	for _, key := range toDelete {
		delete(cm.cache, key)
	}
}

// GetStats returns cache statistics
func (cm *CacheManager) GetStats() *CacheStats {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	stats := &CacheStats{
		TotalEntries: len(cm.cache),
	}

	if len(cm.cache) == 0 {
		return stats
	}

	var totalHits, totalMisses int64
	var oldestTime, newestTime time.Time
	var totalTTL time.Duration

	first := true
	for _, entry := range cm.cache {
		totalHits += int64(entry.HitCount)

		if first {
			oldestTime = entry.CreatedAt
			newestTime = entry.CreatedAt
			first = false
		} else {
			if entry.CreatedAt.Before(oldestTime) {
				oldestTime = entry.CreatedAt
			}
			if entry.CreatedAt.After(newestTime) {
				newestTime = entry.CreatedAt
			}
		}

		totalTTL += entry.ExpiresAt.Sub(entry.CreatedAt)
	}

	stats.TotalHits = totalHits
	stats.TotalMisses = totalMisses
	if totalHits+totalMisses > 0 {
		stats.HitRate = float64(totalHits) / float64(totalHits+totalMisses)
		stats.MissRate = float64(totalMisses) / float64(totalHits+totalMisses)
	}

	stats.OldestEntry = &oldestTime
	stats.NewestEntry = &newestTime
	stats.AvgTTL = totalTTL / time.Duration(len(cm.cache))

	// Estimate memory usage (rough calculation)
	stats.MemoryUsage = int64(len(cm.cache)) * 1024 // Rough estimate

	return stats
}

// cleanup removes expired entries periodically
func (cm *CacheManager) cleanup() {
	for {
		select {
		case <-cm.cleanupTicker.C:
			cm.performCleanup()
		case <-cm.stopCleanup:
			return
		}
	}
}

// performCleanup removes expired entries
func (cm *CacheManager) performCleanup() {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	now := time.Now()
	toDelete := make([]string, 0)

	for key, entry := range cm.cache {
		if now.After(entry.ExpiresAt) {
			toDelete = append(toDelete, key)
		}
	}

	for _, key := range toDelete {
		delete(cm.cache, key)
	}
}

// Stop stops the cache manager and cleanup goroutine
func (cm *CacheManager) Stop() {
	if cm.cleanupTicker != nil {
		cm.cleanupTicker.Stop()
	}
	if cm.stopCleanup != nil {
		close(cm.stopCleanup)
	}
}

// Calendar-specific cache methods

// GetCalendarData retrieves cached calendar data
func (cm *CacheManager) GetCalendarData(view *CalendarView) (*CalendarData, bool) {
	key := cm.generateCalendarKey(view)
	data, exists := cm.Get(key)
	if !exists {
		return nil, false
	}

	calendarData, ok := data.(*CalendarData)
	return calendarData, ok
}

// SetCalendarData caches calendar data
func (cm *CacheManager) SetCalendarData(view *CalendarView, data *CalendarData) {
	key := cm.generateCalendarKey(view)
	cm.Set(key, data)
}

// GetEvents retrieves cached events
func (cm *CacheManager) GetEvents(window *ScheduleWindow) (*ScheduleResponse, bool) {
	key := cm.generateEventsKey(window)
	data, exists := cm.Get(key)
	if !exists {
		return nil, false
	}

	response, ok := data.(*ScheduleResponse)
	return response, ok
}

// SetEvents caches events
func (cm *CacheManager) SetEvents(window *ScheduleWindow, response *ScheduleResponse) {
	key := cm.generateEventsKey(window)
	cm.Set(key, response)
}

// GetRules retrieves cached recurring rules
func (cm *CacheManager) GetRules(filter RuleFilter) ([]RecurringRule, bool) {
	key := cm.generateRulesKey(filter)
	data, exists := cm.Get(key)
	if !exists {
		return nil, false
	}

	rules, ok := data.([]RecurringRule)
	return rules, ok
}

// SetRules caches recurring rules
func (cm *CacheManager) SetRules(filter RuleFilter, rules []RecurringRule) {
	key := cm.generateRulesKey(filter)
	cm.Set(key, rules)
}

// InvalidateCalendarData invalidates calendar data cache
func (cm *CacheManager) InvalidateCalendarData() {
	cm.InvalidatePattern("calendar:*")
}

// InvalidateEvents invalidates events cache
func (cm *CacheManager) InvalidateEvents() {
	cm.InvalidatePattern("events:*")
}

// InvalidateRules invalidates rules cache
func (cm *CacheManager) InvalidateRules() {
	cm.InvalidatePattern("rules:*")
}

// Key generation methods

// generateCalendarKey generates a cache key for calendar data
func (cm *CacheManager) generateCalendarKey(view *CalendarView) string {
	data := map[string]interface{}{
		"view_type":    view.ViewType,
		"current_date": view.CurrentDate.Format("2006-01-02"),
		"timezone":     view.Timezone.String(),
		"filter":       view.Filter,
	}

	return cm.generateKey("calendar", data)
}

// generateEventsKey generates a cache key for events
func (cm *CacheManager) generateEventsKey(window *ScheduleWindow) string {
	data := map[string]interface{}{
		"from":       window.From.Format(time.RFC3339),
		"till":       window.Till.Format(time.RFC3339),
		"queue_name": window.QueueName,
		"limit":      window.Limit,
	}

	return cm.generateKey("events", data)
}

// generateRulesKey generates a cache key for recurring rules
func (cm *CacheManager) generateRulesKey(filter RuleFilter) string {
	data := map[string]interface{}{
		"ids":         filter.IDs,
		"queue_names": filter.QueueNames,
		"job_types":   filter.JobTypes,
		"is_active":   filter.IsActive,
		"is_paused":   filter.IsPaused,
	}

	return cm.generateKey("rules", data)
}

// generateKey generates a cache key from prefix and data
func (cm *CacheManager) generateKey(prefix string, data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		// Fallback to simple string representation
		return fmt.Sprintf("%s:%v", prefix, data)
	}

	hash := md5.Sum(jsonData)
	return fmt.Sprintf("%s:%x", prefix, hash)
}

// matchesPattern checks if a key matches a pattern (simple wildcard support)
func matchesPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == "" {
		return key == ""
	}

	// Simple wildcard pattern matching
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(key) >= len(prefix) && key[:len(prefix)] == prefix
	}

	return key == pattern
}

// CacheConfig represents cache configuration
type CacheConfig struct {
	TTL             time.Duration `json:"ttl"`
	CleanupInterval time.Duration `json:"cleanup_interval"`
	MaxEntries      int           `json:"max_entries"`
	EnableStats     bool          `json:"enable_stats"`
}

// DefaultCacheConfig returns default cache configuration
func DefaultCacheConfig() *CacheConfig {
	return &CacheConfig{
		TTL:             5 * time.Minute,
		CleanupInterval: 2*time.Minute + 30*time.Second,
		MaxEntries:      1000,
		EnableStats:     true,
	}
}