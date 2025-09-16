package calendarview

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheManager_NewCacheManager(t *testing.T) {
	ttl := 5 * time.Minute
	cm := NewCacheManager(ttl)

	assert.NotNil(t, cm)
	assert.Equal(t, ttl, cm.ttl)
	assert.NotNil(t, cm.cache)
	assert.NotNil(t, cm.cleanupTicker)
	assert.NotNil(t, cm.stopCleanup)

	// Test default TTL
	cm2 := NewCacheManager(0)
	assert.Equal(t, 5*time.Minute, cm2.ttl)

	// Cleanup
	cm.Stop()
	cm2.Stop()
}

func TestCacheManager_SetAndGet(t *testing.T) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	key := "test-key"
	value := "test-value"

	// Test Set and Get
	cm.Set(key, value)
	retrieved, exists := cm.Get(key)

	assert.True(t, exists)
	assert.Equal(t, value, retrieved)

	// Test non-existent key
	_, exists = cm.Get("non-existent")
	assert.False(t, exists)
}

func TestCacheManager_SetWithTTL(t *testing.T) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	key := "test-key"
	value := "test-value"
	customTTL := 100 * time.Millisecond

	cm.SetWithTTL(key, value, customTTL)

	// Should exist immediately
	retrieved, exists := cm.Get(key)
	assert.True(t, exists)
	assert.Equal(t, value, retrieved)

	// Should expire after custom TTL
	time.Sleep(150 * time.Millisecond)
	_, exists = cm.Get(key)
	assert.False(t, exists)
}

func TestCacheManager_Delete(t *testing.T) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	key := "test-key"
	value := "test-value"

	cm.Set(key, value)
	_, exists := cm.Get(key)
	assert.True(t, exists)

	cm.Delete(key)
	_, exists = cm.Get(key)
	assert.False(t, exists)
}

func TestCacheManager_Invalidate(t *testing.T) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	// Add multiple items
	cm.Set("key1", "value1")
	cm.Set("key2", "value2")
	cm.Set("key3", "value3")

	// Verify they exist
	_, exists := cm.Get("key1")
	assert.True(t, exists)
	_, exists = cm.Get("key2")
	assert.True(t, exists)

	// Invalidate all
	cm.Invalidate()

	// Verify they're gone
	_, exists = cm.Get("key1")
	assert.False(t, exists)
	_, exists = cm.Get("key2")
	assert.False(t, exists)
	_, exists = cm.Get("key3")
	assert.False(t, exists)
}

func TestCacheManager_InvalidatePattern(t *testing.T) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	// Add items with different prefixes
	cm.Set("calendar:month:2023-06", "month data")
	cm.Set("calendar:week:2023-06-15", "week data")
	cm.Set("events:2023-06-15", "events data")
	cm.Set("rules:active", "rules data")

	// Invalidate calendar items
	cm.InvalidatePattern("calendar:*")

	// Calendar items should be gone
	_, exists := cm.Get("calendar:month:2023-06")
	assert.False(t, exists)
	_, exists = cm.Get("calendar:week:2023-06-15")
	assert.False(t, exists)

	// Other items should still exist
	_, exists = cm.Get("events:2023-06-15")
	assert.True(t, exists)
	_, exists = cm.Get("rules:active")
	assert.True(t, exists)
}

func TestCacheManager_GetStats(t *testing.T) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	// Test empty cache
	stats := cm.GetStats()
	assert.Equal(t, 0, stats.TotalEntries)
	assert.Equal(t, int64(0), stats.TotalHits)

	// Add some items and generate hits
	cm.Set("key1", "value1")
	cm.Set("key2", "value2")

	// Generate some hits
	cm.Get("key1")
	cm.Get("key1")
	cm.Get("key2")

	stats = cm.GetStats()
	assert.Equal(t, 2, stats.TotalEntries)
	assert.Equal(t, int64(3), stats.TotalHits)
	assert.NotNil(t, stats.OldestEntry)
	assert.NotNil(t, stats.NewestEntry)
}

func TestCacheManager_CalendarSpecificMethods(t *testing.T) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	// Test CalendarData caching
	view := &CalendarView{
		ViewType:    ViewTypeMonth,
		CurrentDate: time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
		Timezone:    time.UTC,
		Filter:      EventFilter{},
	}

	calendarData := &CalendarData{
		TotalEvents: 10,
		PeakDensity: 5.0,
		Cells:       [][]CalendarCell{},
	}

	// Set and get calendar data
	cm.SetCalendarData(view, calendarData)
	retrieved, exists := cm.GetCalendarData(view)

	assert.True(t, exists)
	assert.Equal(t, calendarData.TotalEvents, retrieved.TotalEvents)
	assert.Equal(t, calendarData.PeakDensity, retrieved.PeakDensity)
}

func TestCacheManager_EventsCaching(t *testing.T) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	window := &ScheduleWindow{
		From:      time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC),
		Till:      time.Date(2023, 6, 30, 23, 59, 59, 0, time.UTC),
		QueueName: "test-queue",
		Limit:     100,
	}

	response := &ScheduleResponse{
		Events: []CalendarEvent{
			{ID: "event1", QueueName: "test-queue"},
			{ID: "event2", QueueName: "test-queue"},
		},
		TotalCount: 2,
		HasMore:    false,
	}

	// Set and get events
	cm.SetEvents(window, response)
	retrieved, exists := cm.GetEvents(window)

	assert.True(t, exists)
	assert.Equal(t, response.TotalCount, retrieved.TotalCount)
	assert.Equal(t, len(response.Events), len(retrieved.Events))
}

func TestCacheManager_RulesCaching(t *testing.T) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	filter := RuleFilter{
		QueueNames: []string{"test-queue"},
		IsActive:   func() *bool { active := true; return &active }(),
	}

	rules := []RecurringRule{
		{
			ID:       "rule1",
			Name:     "Test Rule 1",
			IsActive: true,
		},
		{
			ID:       "rule2",
			Name:     "Test Rule 2",
			IsActive: true,
		},
	}

	// Set and get rules
	cm.SetRules(filter, rules)
	retrieved, exists := cm.GetRules(filter)

	assert.True(t, exists)
	assert.Equal(t, len(rules), len(retrieved))
	assert.Equal(t, rules[0].ID, retrieved[0].ID)
	assert.Equal(t, rules[1].ID, retrieved[1].ID)
}

func TestCacheManager_SpecificInvalidations(t *testing.T) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	// Add items with different prefixes
	cm.Set("calendar:test", "calendar data")
	cm.Set("events:test", "events data")
	cm.Set("rules:test", "rules data")
	cm.Set("other:test", "other data")

	// Test specific invalidations
	cm.InvalidateCalendarData()
	_, exists := cm.Get("calendar:test")
	assert.False(t, exists)

	cm.InvalidateEvents()
	_, exists = cm.Get("events:test")
	assert.False(t, exists)

	cm.InvalidateRules()
	_, exists = cm.Get("rules:test")
	assert.False(t, exists)

	// Other data should still exist
	_, exists = cm.Get("other:test")
	assert.True(t, exists)
}

func TestCacheManager_KeyGeneration(t *testing.T) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	// Test calendar key generation
	view1 := &CalendarView{
		ViewType:    ViewTypeMonth,
		CurrentDate: time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
		Timezone:    time.UTC,
		Filter:      EventFilter{QueueNames: []string{"queue1"}},
	}

	view2 := &CalendarView{
		ViewType:    ViewTypeMonth,
		CurrentDate: time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
		Timezone:    time.UTC,
		Filter:      EventFilter{QueueNames: []string{"queue2"}}, // Different filter
	}

	key1 := cm.generateCalendarKey(view1)
	key2 := cm.generateCalendarKey(view2)

	// Keys should be different due to different filters
	assert.NotEqual(t, key1, key2)

	// Same view should generate same key
	key1Duplicate := cm.generateCalendarKey(view1)
	assert.Equal(t, key1, key1Duplicate)
}

func TestCacheManager_Cleanup(t *testing.T) {
	// Use very short TTL for testing
	cm := NewCacheManager(50 * time.Millisecond)
	defer cm.Stop()

	// Add an item
	cm.Set("test-key", "test-value")

	// Should exist immediately
	_, exists := cm.Get("test-key")
	assert.True(t, exists)

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, exists = cm.Get("test-key")
	assert.False(t, exists)
}

func TestCacheManager_HitCounting(t *testing.T) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	key := "test-key"
	value := "test-value"

	cm.Set(key, value)

	// Generate multiple hits
	for i := 0; i < 5; i++ {
		retrieved, exists := cm.Get(key)
		assert.True(t, exists)
		assert.Equal(t, value, retrieved)
	}

	// Check the entry directly to verify hit counting
	cm.mutex.RLock()
	entry, exists := cm.cache[key]
	cm.mutex.RUnlock()

	assert.True(t, exists)
	assert.Equal(t, 5, entry.HitCount)
	assert.False(t, entry.LastHit.IsZero())
}

func TestMatchesPattern(t *testing.T) {
	testCases := []struct {
		key     string
		pattern string
		matches bool
	}{
		{"calendar:month", "calendar:*", true},
		{"calendar:week", "calendar:*", true},
		{"events:today", "calendar:*", false},
		{"anything", "*", true},
		{"exact", "exact", true},
		{"exact", "other", false},
		{"", "", true},
		{"test", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.key+"_"+tc.pattern, func(t *testing.T) {
			result := matchesPattern(tc.key, tc.pattern)
			assert.Equal(t, tc.matches, result)
		})
	}
}

func TestDefaultCacheConfig(t *testing.T) {
	config := DefaultCacheConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 5*time.Minute, config.TTL)
	assert.Equal(t, 2*time.Minute+30*time.Second, config.CleanupInterval)
	assert.Equal(t, 1000, config.MaxEntries)
	assert.True(t, config.EnableStats)
}

func TestCacheEntry(t *testing.T) {
	now := time.Now()
	entry := &CacheEntry{
		Key:       "test-key",
		Data:      "test-data",
		CreatedAt: now,
		ExpiresAt: now.Add(5 * time.Minute),
		HitCount:  0,
		LastHit:   now,
	}

	assert.Equal(t, "test-key", entry.Key)
	assert.Equal(t, "test-data", entry.Data)
	assert.Equal(t, now, entry.CreatedAt)
	assert.Equal(t, 0, entry.HitCount)
}

// Benchmark tests

func BenchmarkCacheManager_Set(b *testing.B) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i)
		cm.Set(key, "test-value")
	}
}

func BenchmarkCacheManager_Get(b *testing.B) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%d", i)
		cm.Set(key, "test-value")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("key%d", i%1000)
		cm.Get(key)
	}
}

func BenchmarkCacheManager_GetCalendarData(b *testing.B) {
	cm := NewCacheManager(5 * time.Minute)
	defer cm.Stop()

	view := &CalendarView{
		ViewType:    ViewTypeMonth,
		CurrentDate: time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC),
		Timezone:    time.UTC,
		Filter:      EventFilter{},
	}

	calendarData := &CalendarData{
		TotalEvents: 100,
		PeakDensity: 10.0,
		Cells:       make([][]CalendarCell, 6),
	}

	cm.SetCalendarData(view, calendarData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.GetCalendarData(view)
	}
}