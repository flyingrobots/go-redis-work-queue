// Copyright 2025 James Ross
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// DeadLetterHook manages failed webhook deliveries and replay functionality
type DeadLetterHook struct {
	storage    DLHStorage
	replayMgr  *ReplayManager
	config     DLHConfig
	mu         sync.RWMutex
}

// DLHStorage interface for dead letter hook storage
type DLHStorage interface {
	Store(ctx context.Context, entry DLHEntry) (*DLHEntry, error)
	GetByID(ctx context.Context, id string) (*DLHEntry, error)
	List(ctx context.Context, filter DLHFilter) ([]DLHEntry, error)
	UpdateStatus(ctx context.Context, id string, status DLHStatus) error
	Delete(ctx context.Context, id string) error
	GetMetrics(ctx context.Context) (DLHMetrics, error)
}

// InMemoryDLHStorage is an in-memory implementation for testing
type InMemoryDLHStorage struct {
	entries map[string]DLHEntry
	mu      sync.RWMutex
}

// ReplayManager handles replaying dead letter hook entries
type ReplayManager struct {
	storage       DLHStorage
	webhookClient WebhookClient
	config        ReplayConfig
}

// WebhookClient interface for webhook delivery
type WebhookClient interface {
	DeliverWebhook(ctx context.Context, url, secret string, payload []byte, headers map[string]string) error
}

// MockWebhookClient is a mock implementation for testing
type MockWebhookClient struct {
	responses     map[string]error
	deliveryCount map[string]int
	deliveries    []WebhookDelivery
	mu            sync.RWMutex
}

// DLHEntry represents a failed webhook delivery stored in DLH
type DLHEntry struct {
	ID           string                 `json:"id"`
	WebhookID    string                 `json:"webhook_id"`
	URL          string                 `json:"url"`
	EventID      string                 `json:"event_id"`
	Payload      json.RawMessage        `json:"payload"`
	Headers      map[string]string      `json:"headers"`
	FailureCount int                    `json:"failure_count"`
	LastError    string                 `json:"last_error"`
	Status       DLHStatus              `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	LastAttempt  *time.Time             `json:"last_attempt,omitempty"`
	NextRetry    *time.Time             `json:"next_retry,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// DLHStatus represents the status of a DLH entry
type DLHStatus string

const (
	DLHStatusPending   DLHStatus = "pending"
	DLHStatusRetrying  DLHStatus = "retrying"
	DLHStatusExhausted DLHStatus = "exhausted"
	DLHStatusReplaying DLHStatus = "replaying"
	DLHStatusCompleted DLHStatus = "completed"
	DLHStatusArchived  DLHStatus = "archived"
)

// DLHFilter defines filtering criteria for DLH entries
type DLHFilter struct {
	WebhookID    string     `json:"webhook_id,omitempty"`
	Status       DLHStatus  `json:"status,omitempty"`
	CreatedAfter *time.Time `json:"created_after,omitempty"`
	CreatedBefore *time.Time `json:"created_before,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// DLHMetrics provides statistics about DLH entries
type DLHMetrics struct {
	TotalEntries     int64              `json:"total_entries"`
	StatusCounts     map[DLHStatus]int64 `json:"status_counts"`
	AvgFailureCount  float64            `json:"avg_failure_count"`
	OldestEntry      *time.Time         `json:"oldest_entry,omitempty"`
	RecentActivity   []DLHActivity      `json:"recent_activity"`
}

// DLHActivity represents recent DLH activity
type DLHActivity struct {
	Action    string    `json:"action"`
	EntryID   string    `json:"entry_id"`
	Timestamp time.Time `json:"timestamp"`
	Details   string    `json:"details,omitempty"`
}

// DLHConfig configures the Dead Letter Hook system
type DLHConfig struct {
	MaxRetries       int           `json:"max_retries"`
	RetryDelay       time.Duration `json:"retry_delay"`
	CleanupInterval  time.Duration `json:"cleanup_interval"`
	ArchiveAfter     time.Duration `json:"archive_after"`
	EnableReplay     bool          `json:"enable_replay"`
	ReplayBatchSize  int           `json:"replay_batch_size"`
}

// ReplayConfig configures replay behavior
type ReplayConfig struct {
	BatchSize       int           `json:"batch_size"`
	DelayBetweenBatches time.Duration `json:"delay_between_batches"`
	MaxConcurrent   int           `json:"max_concurrent"`
	TimeoutPerItem  time.Duration `json:"timeout_per_item"`
}

// WebhookDelivery represents a webhook delivery attempt
type WebhookDelivery struct {
	URL       string            `json:"url"`
	Payload   []byte            `json:"payload"`
	Headers   map[string]string `json:"headers"`
	Timestamp time.Time         `json:"timestamp"`
	Error     error             `json:"error,omitempty"`
}

// NewInMemoryDLHStorage creates a new in-memory DLH storage
func NewInMemoryDLHStorage() *InMemoryDLHStorage {
	return &InMemoryDLHStorage{
		entries: make(map[string]DLHEntry),
	}
}

// Store stores a DLH entry
func (s *InMemoryDLHStorage) Store(ctx context.Context, entry DLHEntry) (*DLHEntry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if entry.ID == "" {
		entry.ID = fmt.Sprintf("dlh_%d", time.Now().UnixNano())
	}

	entry.CreatedAt = time.Now()
	entry.UpdatedAt = time.Now()

	s.entries[entry.ID] = entry
	return &entry, nil
}

// GetByID retrieves a DLH entry by ID
func (s *InMemoryDLHStorage) GetByID(ctx context.Context, id string) (*DLHEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.entries[id]
	if !exists {
		return nil, fmt.Errorf("DLH entry not found: %s", id)
	}

	return &entry, nil
}

// List retrieves DLH entries based on filter
func (s *InMemoryDLHStorage) List(ctx context.Context, filter DLHFilter) ([]DLHEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var results []DLHEntry

	for _, entry := range s.entries {
		if s.matchesFilter(entry, filter) {
			results = append(results, entry)
		}
	}

	// Apply limit and offset
	if filter.Offset > 0 && filter.Offset < len(results) {
		results = results[filter.Offset:]
	}

	if filter.Limit > 0 && filter.Limit < len(results) {
		results = results[:filter.Limit]
	}

	return results, nil
}

// UpdateStatus updates the status of a DLH entry
func (s *InMemoryDLHStorage) UpdateStatus(ctx context.Context, id string, status DLHStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.entries[id]
	if !exists {
		return fmt.Errorf("DLH entry not found: %s", id)
	}

	entry.Status = status
	entry.UpdatedAt = time.Now()
	s.entries[id] = entry

	return nil
}

// Delete removes a DLH entry
func (s *InMemoryDLHStorage) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.entries[id]; !exists {
		return fmt.Errorf("DLH entry not found: %s", id)
	}

	delete(s.entries, id)
	return nil
}

// GetMetrics returns DLH metrics
func (s *InMemoryDLHStorage) GetMetrics(ctx context.Context) (DLHMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := DLHMetrics{
		StatusCounts: make(map[DLHStatus]int64),
	}

	var totalFailures int64
	var oldestTime *time.Time

	for _, entry := range s.entries {
		metrics.TotalEntries++
		metrics.StatusCounts[entry.Status]++
		totalFailures += int64(entry.FailureCount)

		if oldestTime == nil || entry.CreatedAt.Before(*oldestTime) {
			oldestTime = &entry.CreatedAt
		}
	}

	if metrics.TotalEntries > 0 {
		metrics.AvgFailureCount = float64(totalFailures) / float64(metrics.TotalEntries)
	}

	metrics.OldestEntry = oldestTime

	return metrics, nil
}

// matchesFilter checks if an entry matches the filter criteria
func (s *InMemoryDLHStorage) matchesFilter(entry DLHEntry, filter DLHFilter) bool {
	if filter.WebhookID != "" && entry.WebhookID != filter.WebhookID {
		return false
	}

	if filter.Status != "" && entry.Status != filter.Status {
		return false
	}

	if filter.CreatedAfter != nil && entry.CreatedAt.Before(*filter.CreatedAfter) {
		return false
	}

	if filter.CreatedBefore != nil && entry.CreatedAt.After(*filter.CreatedBefore) {
		return false
	}

	return true
}

// NewMockWebhookClient creates a new mock webhook client
func NewMockWebhookClient() *MockWebhookClient {
	return &MockWebhookClient{
		responses:     make(map[string]error),
		deliveryCount: make(map[string]int),
	}
}

// SetResponse configures how the client responds to a URL
func (c *MockWebhookClient) SetResponse(url string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.responses[url] = err
}

// DeliverWebhook simulates webhook delivery
func (c *MockWebhookClient) DeliverWebhook(ctx context.Context, url, secret string, payload []byte, headers map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deliveryCount[url]++

	delivery := WebhookDelivery{
		URL:       url,
		Payload:   payload,
		Headers:   headers,
		Timestamp: time.Now(),
	}

	if err, exists := c.responses[url]; exists {
		delivery.Error = err
		c.deliveries = append(c.deliveries, delivery)
		return err
	}

	c.deliveries = append(c.deliveries, delivery)
	return nil
}

// GetDeliveryCount returns the number of deliveries for a URL
func (c *MockWebhookClient) GetDeliveryCount(url string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.deliveryCount[url]
}

// GetDeliveries returns all deliveries
func (c *MockWebhookClient) GetDeliveries() []WebhookDelivery {
	c.mu.RLock()
	defer c.mu.RUnlock()
	deliveries := make([]WebhookDelivery, len(c.deliveries))
	copy(deliveries, c.deliveries)
	return deliveries
}

// ClearDeliveries clears delivery history
func (c *MockWebhookClient) ClearDeliveries() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.deliveries = nil
	c.deliveryCount = make(map[string]int)
}

// NewDeadLetterHook creates a new Dead Letter Hook system
func NewDeadLetterHook(storage DLHStorage, config DLHConfig) *DeadLetterHook {
	return &DeadLetterHook{
		storage: storage,
		config:  config,
	}
}

// NewReplayManager creates a new replay manager
func NewReplayManager(storage DLHStorage, client WebhookClient, config ReplayConfig) *ReplayManager {
	return &ReplayManager{
		storage:       storage,
		webhookClient: client,
		config:        config,
	}
}

// StoreFailedDelivery stores a failed webhook delivery in DLH
func (dlh *DeadLetterHook) StoreFailedDelivery(ctx context.Context, webhookID, url, eventID string, payload []byte, err error) error {
	entry := DLHEntry{
		WebhookID:    webhookID,
		URL:          url,
		EventID:      eventID,
		Payload:      json.RawMessage(payload),
		FailureCount: 1,
		LastError:    err.Error(),
		Status:       DLHStatusPending,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
	}

	_, err = dlh.storage.Store(ctx, entry)
	return err
}

// GetEntries retrieves DLH entries based on filter
func (dlh *DeadLetterHook) GetEntries(ctx context.Context, filter DLHFilter) ([]DLHEntry, error) {
	return dlh.storage.List(ctx, filter)
}

// GetMetrics returns DLH metrics
func (dlh *DeadLetterHook) GetMetrics(ctx context.Context) (DLHMetrics, error) {
	return dlh.storage.GetMetrics(ctx)
}

// ArchiveEntry archives a DLH entry
func (dlh *DeadLetterHook) ArchiveEntry(ctx context.Context, id string) error {
	return dlh.storage.UpdateStatus(ctx, id, DLHStatusArchived)
}

// ReplayEntry replays a single DLH entry
func (rm *ReplayManager) ReplayEntry(ctx context.Context, entryID string) error {
	entry, err := rm.storage.GetByID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to get DLH entry: %w", err)
	}

	// Update status to replaying
	err = rm.storage.UpdateStatus(ctx, entryID, DLHStatusReplaying)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Attempt delivery
	err = rm.webhookClient.DeliverWebhook(ctx, entry.URL, "", entry.Payload, entry.Headers)

	if err != nil {
		// Update failure count and status
		entry.FailureCount++
		entry.LastError = err.Error()
		entry.LastAttempt = func() *time.Time { t := time.Now(); return &t }()

		if entry.FailureCount >= 5 { // Max retries
			entry.Status = DLHStatusExhausted
		} else {
			entry.Status = DLHStatusPending
			nextRetry := time.Now().Add(time.Duration(entry.FailureCount) * time.Minute)
			entry.NextRetry = &nextRetry
		}

		// Store updated entry
		_, err = rm.storage.Store(ctx, *entry)
		return err
	}

	// Success - mark as completed
	return rm.storage.UpdateStatus(ctx, entryID, DLHStatusCompleted)
}

// ReplayBatch replays a batch of DLH entries
func (rm *ReplayManager) ReplayBatch(ctx context.Context, filter DLHFilter) (int, error) {
	entries, err := rm.storage.List(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to list entries: %w", err)
	}

	var successful int

	for _, entry := range entries {
		if entry.Status != DLHStatusPending {
			continue
		}

		err := rm.ReplayEntry(ctx, entry.ID)
		if err == nil {
			successful++
		}

		// Add delay between replays
		if rm.config.DelayBetweenBatches > 0 {
			time.Sleep(rm.config.DelayBetweenBatches)
		}
	}

	return successful, nil
}

// Integration Tests

func TestDLH_BasicStorage(t *testing.T) {
	storage := NewInMemoryDLHStorage()
	ctx := context.Background()

	// Store entry
	entry := DLHEntry{
		WebhookID:    "webhook_001",
		URL:          "https://example.com/webhook",
		EventID:      "event_123",
		Payload:      json.RawMessage(`{"event": "job_failed"}`),
		FailureCount: 1,
		LastError:    "Connection timeout",
		Status:       DLHStatusPending,
	}

	storedEntry, err := storage.Store(ctx, entry)
	assert.NoError(t, err)
	assert.NotEmpty(t, storedEntry.ID)
	entry = *storedEntry

	// Retrieve entry by ID
	retrieved, err := storage.GetByID(ctx, entry.ID)
	assert.NoError(t, err)
	assert.Equal(t, entry.WebhookID, retrieved.WebhookID)
	assert.Equal(t, entry.URL, retrieved.URL)
	assert.Equal(t, entry.Status, retrieved.Status)

	// Update status
	err = storage.UpdateStatus(ctx, entry.ID, DLHStatusCompleted)
	assert.NoError(t, err)

	updated, err := storage.GetByID(ctx, entry.ID)
	assert.NoError(t, err)
	assert.Equal(t, DLHStatusCompleted, updated.Status)
}

func TestDLH_FilteringAndList(t *testing.T) {
	storage := NewInMemoryDLHStorage()
	ctx := context.Background()

	// Store multiple entries
	entries := []DLHEntry{
		{WebhookID: "webhook_001", Status: DLHStatusPending, URL: "https://example.com/1"},
		{WebhookID: "webhook_001", Status: DLHStatusCompleted, URL: "https://example.com/2"},
		{WebhookID: "webhook_002", Status: DLHStatusPending, URL: "https://example.com/3"},
		{WebhookID: "webhook_002", Status: DLHStatusExhausted, URL: "https://example.com/4"},
	}

	for _, entry := range entries {
		_, err := storage.Store(ctx, entry)
		assert.NoError(t, err)
	}

	// Filter by webhook ID
	filter := DLHFilter{WebhookID: "webhook_001"}
	results, err := storage.List(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// Filter by status
	filter = DLHFilter{Status: DLHStatusPending}
	results, err = storage.List(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, results, 2)

	// Filter with limit
	filter = DLHFilter{Limit: 2}
	results, err = storage.List(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestDLH_Metrics(t *testing.T) {
	storage := NewInMemoryDLHStorage()
	ctx := context.Background()

	// Store entries with different statuses
	entries := []DLHEntry{
		{Status: DLHStatusPending, FailureCount: 1},
		{Status: DLHStatusPending, FailureCount: 2},
		{Status: DLHStatusCompleted, FailureCount: 1},
		{Status: DLHStatusExhausted, FailureCount: 5},
	}

	for _, entry := range entries {
		_, err := storage.Store(ctx, entry)
		assert.NoError(t, err)
	}

	metrics, err := storage.GetMetrics(ctx)
	assert.NoError(t, err)

	assert.Equal(t, int64(4), metrics.TotalEntries)
	assert.Equal(t, int64(2), metrics.StatusCounts[DLHStatusPending])
	assert.Equal(t, int64(1), metrics.StatusCounts[DLHStatusCompleted])
	assert.Equal(t, int64(1), metrics.StatusCounts[DLHStatusExhausted])
	assert.Equal(t, float64(2.25), metrics.AvgFailureCount) // (1+2+1+5)/4
	assert.NotNil(t, metrics.OldestEntry)
}

func TestDLH_StoreFailedDelivery(t *testing.T) {
	storage := NewInMemoryDLHStorage()
	config := DLHConfig{MaxRetries: 3}
	dlh := NewDeadLetterHook(storage, config)

	ctx := context.Background()
	err := dlh.StoreFailedDelivery(ctx, "webhook_001", "https://example.com/webhook", "event_123", []byte(`{"test": true}`), fmt.Errorf("connection failed"))
	assert.NoError(t, err)

	entries, err := dlh.GetEntries(ctx, DLHFilter{})
	assert.NoError(t, err)
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "webhook_001", entry.WebhookID)
	assert.Equal(t, "https://example.com/webhook", entry.URL)
	assert.Equal(t, "event_123", entry.EventID)
	assert.Equal(t, DLHStatusPending, entry.Status)
	assert.Equal(t, 1, entry.FailureCount)
	assert.Contains(t, entry.LastError, "connection failed")
}

func TestReplayManager_SingleEntryReplay(t *testing.T) {
	storage := NewInMemoryDLHStorage()
	client := NewMockWebhookClient()
	config := ReplayConfig{BatchSize: 10, DelayBetweenBatches: time.Millisecond}

	rm := NewReplayManager(storage, client, config)
	ctx := context.Background()

	// Store a failed entry
	entry := DLHEntry{
		WebhookID: "webhook_001",
		URL:       "https://example.com/webhook",
		EventID:   "event_123",
		Payload:   json.RawMessage(`{"event": "job_failed"}`),
		Status:    DLHStatusPending,
		Headers:   map[string]string{"Content-Type": "application/json"},
	}

	storedEntry, err := storage.Store(ctx, entry)
	assert.NoError(t, err)
	entry = *storedEntry

	// Configure client to succeed
	client.SetResponse("https://example.com/webhook", nil)

	// Replay the entry
	err = rm.ReplayEntry(ctx, entry.ID)
	assert.NoError(t, err)

	// Verify webhook was called
	assert.Equal(t, 1, client.GetDeliveryCount("https://example.com/webhook"))

	// Verify status was updated
	updated, err := storage.GetByID(ctx, entry.ID)
	assert.NoError(t, err)
	assert.Equal(t, DLHStatusCompleted, updated.Status)
}

func TestReplayManager_FailedReplay(t *testing.T) {
	storage := NewInMemoryDLHStorage()
	client := NewMockWebhookClient()
	config := ReplayConfig{BatchSize: 10}

	rm := NewReplayManager(storage, client, config)
	ctx := context.Background()

	// Store a failed entry
	entry := DLHEntry{
		WebhookID:    "webhook_001",
		URL:          "https://example.com/webhook",
		EventID:      "event_123",
		Payload:      json.RawMessage(`{"event": "job_failed"}`),
		Status:       DLHStatusPending,
		FailureCount: 1,
	}

	storedEntry, err := storage.Store(ctx, entry)
	assert.NoError(t, err)
	entry = *storedEntry

	// Configure client to fail
	client.SetResponse("https://example.com/webhook", fmt.Errorf("service unavailable"))

	// Replay the entry
	err = rm.ReplayEntry(ctx, entry.ID)
	assert.NoError(t, err) // No error returned even if delivery fails

	// Verify failure count was incremented
	updated, err := storage.GetByID(ctx, entry.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, updated.FailureCount)
	assert.Equal(t, DLHStatusPending, updated.Status)
	assert.Contains(t, updated.LastError, "service unavailable")
	assert.NotNil(t, updated.LastAttempt)
	assert.NotNil(t, updated.NextRetry)
}

func TestReplayManager_ExhaustedRetries(t *testing.T) {
	storage := NewInMemoryDLHStorage()
	client := NewMockWebhookClient()
	config := ReplayConfig{BatchSize: 10}

	rm := NewReplayManager(storage, client, config)
	ctx := context.Background()

	// Store an entry that's already failed 4 times
	entry := DLHEntry{
		WebhookID:    "webhook_001",
		URL:          "https://example.com/webhook",
		EventID:      "event_123",
		Payload:      json.RawMessage(`{"event": "job_failed"}`),
		Status:       DLHStatusPending,
		FailureCount: 4,
	}

	storedEntry, err := storage.Store(ctx, entry)
	assert.NoError(t, err)
	entry = *storedEntry

	// Configure client to fail
	client.SetResponse("https://example.com/webhook", fmt.Errorf("persistent failure"))

	// Replay the entry (this will be the 5th failure)
	err = rm.ReplayEntry(ctx, entry.ID)
	assert.NoError(t, err)

	// Verify entry is marked as exhausted
	updated, err := storage.GetByID(ctx, entry.ID)
	assert.NoError(t, err)
	assert.Equal(t, 5, updated.FailureCount)
	assert.Equal(t, DLHStatusExhausted, updated.Status)
}

func TestReplayManager_BatchReplay(t *testing.T) {
	storage := NewInMemoryDLHStorage()
	client := NewMockWebhookClient()
	config := ReplayConfig{
		BatchSize:           10,
		DelayBetweenBatches: time.Millisecond,
	}

	rm := NewReplayManager(storage, client, config)
	ctx := context.Background()

	// Store multiple pending entries
	for i := 0; i < 5; i++ {
		entry := DLHEntry{
			WebhookID: fmt.Sprintf("webhook_%d", i),
			URL:       fmt.Sprintf("https://example.com/webhook%d", i),
			EventID:   fmt.Sprintf("event_%d", i),
			Payload:   json.RawMessage(`{"event": "job_failed"}`),
			Status:    DLHStatusPending,
		}

		_, err := storage.Store(ctx, entry)
		assert.NoError(t, err)

		// Configure some to succeed, some to fail
		if i%2 == 0 {
			client.SetResponse(entry.URL, nil) // Success
		} else {
			client.SetResponse(entry.URL, fmt.Errorf("failure")) // Failure
		}
	}

	// Store one completed entry (should be skipped)
	completedEntry := DLHEntry{
		WebhookID: "webhook_completed",
		URL:       "https://example.com/completed",
		Status:    DLHStatusCompleted,
	}
	_, err := storage.Store(ctx, completedEntry)
	assert.NoError(t, err)

	// Replay batch
	successful, err := rm.ReplayBatch(ctx, DLHFilter{})
	assert.NoError(t, err)
	assert.Equal(t, 3, successful) // 3 out of 5 should succeed (0, 2, 4)

	// Verify delivery counts
	for i := 0; i < 5; i++ {
		url := fmt.Sprintf("https://example.com/webhook%d", i)
		assert.Equal(t, 1, client.GetDeliveryCount(url))
	}

	// Completed entry should not be called
	assert.Equal(t, 0, client.GetDeliveryCount("https://example.com/completed"))
}

func TestDLH_ArchiveEntry(t *testing.T) {
	storage := NewInMemoryDLHStorage()
	config := DLHConfig{}
	dlh := NewDeadLetterHook(storage, config)

	ctx := context.Background()

	// Store an entry
	entry := DLHEntry{
		WebhookID: "webhook_001",
		URL:       "https://example.com/webhook",
		Status:    DLHStatusExhausted,
	}

	storedEntry, err := storage.Store(ctx, entry)
	assert.NoError(t, err)
	entry = *storedEntry

	// Archive the entry
	err = dlh.ArchiveEntry(ctx, entry.ID)
	assert.NoError(t, err)

	// Verify status
	updated, err := storage.GetByID(ctx, entry.ID)
	assert.NoError(t, err)
	assert.Equal(t, DLHStatusArchived, updated.Status)
}

func TestDLH_ConcurrentOperations(t *testing.T) {
	storage := NewInMemoryDLHStorage()
	config := DLHConfig{}
	dlh := NewDeadLetterHook(storage, config)

	ctx := context.Background()
	const numOperations = 50

	var wg sync.WaitGroup

	// Concurrent stores
	for i := 0; i < numOperations; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			err := dlh.StoreFailedDelivery(ctx,
				fmt.Sprintf("webhook_%d", index),
				fmt.Sprintf("https://example.com/webhook%d", index),
				fmt.Sprintf("event_%d", index),
				[]byte(`{"test": true}`),
				fmt.Errorf("error %d", index))

			assert.NoError(t, err)
		}(i)
	}

	wg.Wait()

	// Verify all entries were stored
	entries, err := dlh.GetEntries(ctx, DLHFilter{})
	assert.NoError(t, err)
	assert.Len(t, entries, numOperations)

	// Verify metrics
	metrics, err := dlh.GetMetrics(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(numOperations), metrics.TotalEntries)
	assert.Equal(t, int64(numOperations), metrics.StatusCounts[DLHStatusPending])
}

// Benchmark Tests

func BenchmarkDLH_StoreEntry(b *testing.B) {
	storage := NewInMemoryDLHStorage()
	ctx := context.Background()

	entry := DLHEntry{
		WebhookID: "webhook_benchmark",
		URL:       "https://example.com/webhook",
		EventID:   "event_benchmark",
		Payload:   json.RawMessage(`{"event": "job_failed", "benchmark": true}`),
		Status:    DLHStatusPending,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entry.ID = fmt.Sprintf("dlh_%d", i)
		storage.Store(ctx, entry)
	}
}

func BenchmarkReplayManager_SingleReplay(b *testing.B) {
	storage := NewInMemoryDLHStorage()
	client := NewMockWebhookClient()
	config := ReplayConfig{BatchSize: 1}

	rm := NewReplayManager(storage, client, config)
	ctx := context.Background()

	// Pre-populate storage
	for i := 0; i < b.N; i++ {
		entry := DLHEntry{
			ID:        fmt.Sprintf("dlh_%d", i),
			WebhookID: "webhook_benchmark",
			URL:       "https://example.com/webhook",
			Status:    DLHStatusPending,
			Payload:   json.RawMessage(`{"benchmark": true}`),
		}
		storage.Store(ctx, entry)
	}

	client.SetResponse("https://example.com/webhook", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.ReplayEntry(ctx, fmt.Sprintf("dlh_%d", i))
	}
}