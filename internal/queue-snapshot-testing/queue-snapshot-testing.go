// Copyright 2025 James Ross
package queuesnapshotesting

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// SnapshotManager manages queue snapshots
type SnapshotManager struct {
	config     *SnapshotConfig
	storage    Storage
	redis      *redis.Client
	logger     *zap.Logger
	mu         sync.RWMutex
}

// Storage interface for snapshot persistence
type Storage interface {
	Save(snapshot *Snapshot) error
	Load(id string) (*Snapshot, error)
	Delete(id string) error
	List(filter *SnapshotFilter) ([]*SnapshotMetadata, error)
	Exists(id string) bool
}

// NewSnapshotManager creates a new snapshot manager
func NewSnapshotManager(config *SnapshotConfig, redis *redis.Client, logger *zap.Logger) (*SnapshotManager, error) {
	if config == nil {
		config = &SnapshotConfig{
			StoragePath:      "./snapshots",
			MaxSnapshots:     100,
			RetentionDays:    30,
			CompressLevel:    gzip.BestSpeed,
			IgnoreTimestamps: true,
			IgnoreIDs:        true,
			MaxJobsPerSnapshot: 10000,
			SampleRate:       1.0,
			TimeoutSeconds:   30,
		}
	}

	// Create storage directory
	if err := os.MkdirAll(config.StoragePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	storage := NewFileStorage(config.StoragePath, logger)

	return &SnapshotManager{
		config:  config,
		storage: storage,
		redis:   redis,
		logger:  logger,
	}, nil
}

// CaptureSnapshot captures the current queue state
func (sm *SnapshotManager) CaptureSnapshot(ctx context.Context, name, description string, tags []string) (*Snapshot, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	snapshot := &Snapshot{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Version:     "1.0.0",
		CreatedAt:   time.Now(),
		CreatedBy:   os.Getenv("USER"),
		Tags:        tags,
		Context:     make(map[string]interface{}),
		Environment: os.Getenv("ENVIRONMENT"),
	}

	// Capture queue states
	if err := sm.captureQueues(ctx, snapshot); err != nil {
		return nil, fmt.Errorf("failed to capture queues: %w", err)
	}

	// Capture jobs
	if err := sm.captureJobs(ctx, snapshot); err != nil {
		return nil, fmt.Errorf("failed to capture jobs: %w", err)
	}

	// Capture workers
	if err := sm.captureWorkers(ctx, snapshot); err != nil {
		return nil, fmt.Errorf("failed to capture workers: %w", err)
	}

	// Capture metrics
	if err := sm.captureMetrics(ctx, snapshot); err != nil {
		sm.logger.Warn("Failed to capture metrics", zap.Error(err))
	}

	// Calculate checksum
	snapshot.Checksum = sm.calculateChecksum(snapshot)

	// Compress if configured
	if sm.config.CompressLevel > 0 {
		snapshot.Compressed = true
	}

	// Save snapshot
	if err := sm.storage.Save(snapshot); err != nil {
		return nil, fmt.Errorf("failed to save snapshot: %w", err)
	}

	sm.logger.Info("Snapshot captured",
		zap.String("id", snapshot.ID),
		zap.String("name", name),
		zap.Int("queues", len(snapshot.Queues)),
		zap.Int("jobs", len(snapshot.Jobs)),
		zap.Int("workers", len(snapshot.Workers)))

	return snapshot, nil
}

// LoadSnapshot loads a snapshot by ID
func (sm *SnapshotManager) LoadSnapshot(id string) (*Snapshot, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.storage.Load(id)
}

// RestoreSnapshot restores the queue state from a snapshot
func (sm *SnapshotManager) RestoreSnapshot(ctx context.Context, id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	snapshot, err := sm.storage.Load(id)
	if err != nil {
		return fmt.Errorf("failed to load snapshot: %w", err)
	}

	// Clear existing state
	if err := sm.clearCurrentState(ctx); err != nil {
		return fmt.Errorf("failed to clear current state: %w", err)
	}

	// Restore queues
	for _, queue := range snapshot.Queues {
		if err := sm.restoreQueue(ctx, &queue); err != nil {
			return fmt.Errorf("failed to restore queue %s: %w", queue.Name, err)
		}
	}

	// Restore jobs
	for _, job := range snapshot.Jobs {
		if err := sm.restoreJob(ctx, &job); err != nil {
			return fmt.Errorf("failed to restore job %s: %w", job.ID, err)
		}
	}

	sm.logger.Info("Snapshot restored",
		zap.String("id", snapshot.ID),
		zap.String("name", snapshot.Name))

	return nil
}

// CompareSnapshots compares two snapshots
func (sm *SnapshotManager) CompareSnapshots(leftID, rightID string) (*DiffResult, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	left, err := sm.storage.Load(leftID)
	if err != nil {
		return nil, fmt.Errorf("failed to load left snapshot: %w", err)
	}

	right, err := sm.storage.Load(rightID)
	if err != nil {
		return nil, fmt.Errorf("failed to load right snapshot: %w", err)
	}

	differ := NewDiffer(sm.config)
	return differ.Compare(left, right)
}

// AssertSnapshot asserts that current state matches a snapshot
func (sm *SnapshotManager) AssertSnapshot(ctx context.Context, snapshotID string) (*AssertionResult, error) {
	// Capture current state
	current, err := sm.CaptureSnapshot(ctx, "temp-assertion", "Temporary snapshot for assertion", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to capture current state: %w", err)
	}
	defer sm.storage.Delete(current.ID)

	// Compare with expected
	expected, err := sm.storage.Load(snapshotID)
	if err != nil {
		return nil, fmt.Errorf("failed to load expected snapshot: %w", err)
	}

	differ := NewDiffer(sm.config)
	diff, err := differ.Compare(expected, current)
	if err != nil {
		return nil, fmt.Errorf("failed to compare snapshots: %w", err)
	}

	result := &AssertionResult{
		Passed:      diff.TotalChanges == 0,
		Timestamp:   time.Now(),
		Differences: []Change{},
	}

	if !result.Passed {
		result.Message = fmt.Sprintf("Found %d differences", diff.TotalChanges)
		result.Differences = append(result.Differences, diff.QueueChanges...)
		result.Differences = append(result.Differences, diff.JobChanges...)
		result.Differences = append(result.Differences, diff.WorkerChanges...)
	} else {
		result.Message = "Snapshot assertion passed"
	}

	return result, nil
}

// ListSnapshots lists available snapshots
func (sm *SnapshotManager) ListSnapshots(filter *SnapshotFilter) ([]*SnapshotMetadata, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.storage.List(filter)
}

// DeleteSnapshot deletes a snapshot
func (sm *SnapshotManager) DeleteSnapshot(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	return sm.storage.Delete(id)
}

// Helper methods

func (sm *SnapshotManager) captureQueues(ctx context.Context, snapshot *Snapshot) error {
	// Get all queue names
	queueKeys, err := sm.redis.Keys(ctx, "queue:*").Result()
	if err != nil {
		return err
	}

	snapshot.Queues = make([]QueueState, 0, len(queueKeys))

	for _, key := range queueKeys {
		queueName := strings.TrimPrefix(key, "queue:")

		// Get queue length
		length, err := sm.redis.LLen(ctx, key).Result()
		if err != nil {
			sm.logger.Warn("Failed to get queue length", zap.String("queue", queueName), zap.Error(err))
			continue
		}

		// Get queue config
		configKey := fmt.Sprintf("queue:config:%s", queueName)
		configData, _ := sm.redis.HGetAll(ctx, configKey).Result()

		config := make(map[string]interface{})
		for k, v := range configData {
			config[k] = v
		}

		queue := QueueState{
			Name:   queueName,
			Type:   "redis-list",
			Length: length,
			Config: config,
		}

		snapshot.Queues = append(snapshot.Queues, queue)
	}

	// Sort for deterministic output
	sort.Slice(snapshot.Queues, func(i, j int) bool {
		return snapshot.Queues[i].Name < snapshot.Queues[j].Name
	})

	return nil
}

func (sm *SnapshotManager) captureJobs(ctx context.Context, snapshot *Snapshot) error {
	snapshot.Jobs = make([]JobState, 0)

	// Limit number of jobs to capture
	jobCount := 0

	for _, queue := range snapshot.Queues {
		if jobCount >= sm.config.MaxJobsPerSnapshot {
			break
		}

		// Get jobs from queue
		key := fmt.Sprintf("queue:%s", queue.Name)
		jobs, err := sm.redis.LRange(ctx, key, 0, 100).Result()
		if err != nil {
			sm.logger.Warn("Failed to get jobs", zap.String("queue", queue.Name), zap.Error(err))
			continue
		}

		for _, jobData := range jobs {
			if jobCount >= sm.config.MaxJobsPerSnapshot {
				break
			}

			var job JobState
			if err := json.Unmarshal([]byte(jobData), &job); err != nil {
				// Create minimal job state
				job = JobState{
					ID:        fmt.Sprintf("job-%d", jobCount),
					QueueName: queue.Name,
					Status:    "pending",
					CreatedAt: time.Now(),
					Payload:   map[string]interface{}{"raw": jobData},
				}
			}

			snapshot.Jobs = append(snapshot.Jobs, job)
			jobCount++
		}
	}

	// Sort for deterministic output
	sort.Slice(snapshot.Jobs, func(i, j int) bool {
		if snapshot.Jobs[i].QueueName != snapshot.Jobs[j].QueueName {
			return snapshot.Jobs[i].QueueName < snapshot.Jobs[j].QueueName
		}
		return snapshot.Jobs[i].ID < snapshot.Jobs[j].ID
	})

	return nil
}

func (sm *SnapshotManager) captureWorkers(ctx context.Context, snapshot *Snapshot) error {
	// Get all worker keys
	workerKeys, err := sm.redis.Keys(ctx, "worker:*").Result()
	if err != nil {
		return err
	}

	snapshot.Workers = make([]WorkerState, 0, len(workerKeys))

	for _, key := range workerKeys {
		workerID := strings.TrimPrefix(key, "worker:")

		// Get worker data
		workerData, err := sm.redis.HGetAll(ctx, key).Result()
		if err != nil {
			sm.logger.Warn("Failed to get worker data", zap.String("worker", workerID), zap.Error(err))
			continue
		}

		worker := WorkerState{
			ID:       workerID,
			Status:   workerData["status"],
			Metadata: make(map[string]string),
		}

		// Parse additional fields
		if jobID, ok := workerData["current_job"]; ok {
			worker.CurrentJobID = jobID
		}

		if lastSeen, ok := workerData["last_seen"]; ok {
			if ts, err := time.Parse(time.RFC3339, lastSeen); err == nil {
				worker.LastSeen = ts
			}
		}

		snapshot.Workers = append(snapshot.Workers, worker)
	}

	// Sort for deterministic output
	sort.Slice(snapshot.Workers, func(i, j int) bool {
		return snapshot.Workers[i].ID < snapshot.Workers[j].ID
	})

	return nil
}

func (sm *SnapshotManager) captureMetrics(ctx context.Context, snapshot *Snapshot) error {
	snapshot.Metrics = make(map[string]interface{})

	// Capture key metrics
	metricsKeys := []string{
		"metrics:total_processed",
		"metrics:total_failed",
		"metrics:avg_latency",
	}

	for _, key := range metricsKeys {
		value, err := sm.redis.Get(ctx, key).Result()
		if err == nil {
			metricName := strings.TrimPrefix(key, "metrics:")
			snapshot.Metrics[metricName] = value
		}
	}

	return nil
}

func (sm *SnapshotManager) calculateChecksum(snapshot *Snapshot) string {
	// Serialize snapshot deterministically
	data, _ := json.Marshal(snapshot)

	// Calculate SHA256
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (sm *SnapshotManager) clearCurrentState(ctx context.Context) error {
	// Clear all queues
	queueKeys, err := sm.redis.Keys(ctx, "queue:*").Result()
	if err != nil {
		return err
	}

	for _, key := range queueKeys {
		if err := sm.redis.Del(ctx, key).Err(); err != nil {
			return err
		}
	}

	// Clear workers
	workerKeys, err := sm.redis.Keys(ctx, "worker:*").Result()
	if err != nil {
		return err
	}

	for _, key := range workerKeys {
		if err := sm.redis.Del(ctx, key).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (sm *SnapshotManager) restoreQueue(ctx context.Context, queue *QueueState) error {
	// Restore queue config
	if len(queue.Config) > 0 {
		configKey := fmt.Sprintf("queue:config:%s", queue.Name)
		configMap := make(map[string]interface{})
		for k, v := range queue.Config {
			configMap[k] = v
		}
		if err := sm.redis.HMSet(ctx, configKey, configMap).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (sm *SnapshotManager) restoreJob(ctx context.Context, job *JobState) error {
	// Serialize job
	jobData, err := json.Marshal(job)
	if err != nil {
		return err
	}

	// Add to queue
	queueKey := fmt.Sprintf("queue:%s", job.QueueName)
	return sm.redis.RPush(ctx, queueKey, string(jobData)).Err()
}

// FileStorage implements file-based snapshot storage
type FileStorage struct {
	basePath string
	logger   *zap.Logger
	mu       sync.RWMutex
}

// NewFileStorage creates a new file-based storage
func NewFileStorage(basePath string, logger *zap.Logger) *FileStorage {
	return &FileStorage{
		basePath: basePath,
		logger:   logger,
	}
}

// Save saves a snapshot to disk
func (fs *FileStorage) Save(snapshot *Snapshot) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	filePath := filepath.Join(fs.basePath, fmt.Sprintf("%s.json", snapshot.ID))

	// Serialize snapshot
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize snapshot: %w", err)
	}

	// Compress if needed
	if snapshot.Compressed {
		var buf bytes.Buffer
		gz := gzip.NewWriter(&buf)
		if _, err := gz.Write(data); err != nil {
			return fmt.Errorf("failed to compress snapshot: %w", err)
		}
		gz.Close()
		data = buf.Bytes()
		filePath += ".gz"
	}

	snapshot.SizeBytes = int64(len(data))

	// Write to file
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	// Save metadata
	metaPath := filepath.Join(fs.basePath, fmt.Sprintf("%s.meta.json", snapshot.ID))
	metadata := &SnapshotMetadata{
		ID:          snapshot.ID,
		Name:        snapshot.Name,
		Description: snapshot.Description,
		CreatedAt:   snapshot.CreatedAt,
		SizeBytes:   snapshot.SizeBytes,
		Tags:        snapshot.Tags,
		Environment: snapshot.Environment,
	}

	metaData, _ := json.MarshalIndent(metadata, "", "  ")
	os.WriteFile(metaPath, metaData, 0644)

	return nil
}

// Load loads a snapshot from disk
func (fs *FileStorage) Load(id string) (*Snapshot, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Try compressed first
	filePath := filepath.Join(fs.basePath, fmt.Sprintf("%s.json.gz", id))
	compressed := true

	data, err := os.ReadFile(filePath)
	if err != nil {
		// Try uncompressed
		filePath = filepath.Join(fs.basePath, fmt.Sprintf("%s.json", id))
		data, err = os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("snapshot not found: %w", err)
		}
		compressed = false
	}

	// Decompress if needed
	if compressed {
		gz, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to decompress snapshot: %w", err)
		}
		defer gz.Close()

		data, err = io.ReadAll(gz)
		if err != nil {
			return nil, fmt.Errorf("failed to read compressed snapshot: %w", err)
		}
	}

	// Deserialize
	var snapshot Snapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, fmt.Errorf("failed to deserialize snapshot: %w", err)
	}

	return &snapshot, nil
}

// Delete deletes a snapshot from disk
func (fs *FileStorage) Delete(id string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Try to delete both compressed and uncompressed versions
	os.Remove(filepath.Join(fs.basePath, fmt.Sprintf("%s.json", id)))
	os.Remove(filepath.Join(fs.basePath, fmt.Sprintf("%s.json.gz", id)))
	os.Remove(filepath.Join(fs.basePath, fmt.Sprintf("%s.meta.json", id)))

	return nil
}

// List lists available snapshots
func (fs *FileStorage) List(filter *SnapshotFilter) ([]*SnapshotMetadata, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	pattern := filepath.Join(fs.basePath, "*.meta.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var results []*SnapshotMetadata

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var meta SnapshotMetadata
		if err := json.Unmarshal(data, &meta); err != nil {
			continue
		}

		// Apply filters
		if filter != nil {
			if filter.Name != "" && !strings.Contains(meta.Name, filter.Name) {
				continue
			}

			if len(filter.Tags) > 0 && !hasAnyTag(meta.Tags, filter.Tags) {
				continue
			}

			if filter.Environment != "" && meta.Environment != filter.Environment {
				continue
			}

			if !filter.CreatedAfter.IsZero() && meta.CreatedAt.Before(filter.CreatedAfter) {
				continue
			}

			if !filter.CreatedBefore.IsZero() && meta.CreatedAt.After(filter.CreatedBefore) {
				continue
			}
		}

		results = append(results, &meta)
	}

	// Sort by creation time (newest first)
	sort.Slice(results, func(i, j int) bool {
		return results[i].CreatedAt.After(results[j].CreatedAt)
	})

	// Apply max results
	if filter != nil && filter.MaxResults > 0 && len(results) > filter.MaxResults {
		results = results[:filter.MaxResults]
	}

	return results, nil
}

// Exists checks if a snapshot exists
func (fs *FileStorage) Exists(id string) bool {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	// Check both compressed and uncompressed
	if _, err := os.Stat(filepath.Join(fs.basePath, fmt.Sprintf("%s.json", id))); err == nil {
		return true
	}

	if _, err := os.Stat(filepath.Join(fs.basePath, fmt.Sprintf("%s.json.gz", id))); err == nil {
		return true
	}

	return false
}

func hasAnyTag(tags, filterTags []string) bool {
	tagMap := make(map[string]bool)
	for _, tag := range tags {
		tagMap[tag] = true
	}

	for _, filterTag := range filterTags {
		if tagMap[filterTag] {
			return true
		}
	}

	return false
}