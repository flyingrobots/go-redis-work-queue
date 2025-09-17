package timetraveldebugger

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// EventCapture handles recording job execution events
type EventCapture struct {
	config *CaptureConfig
	redis  *redis.Client
	logger *zap.Logger

	// In-memory buffers for active recordings
	recordings map[string]*activeRecording
	mu         sync.RWMutex

	// Background worker for async operations
	eventChan chan Event
	snapChan  chan StateSnapshot
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// activeRecording represents an in-progress recording
type activeRecording struct {
	record       *ExecutionRecord
	lastSnapshot time.Time
	mu           sync.Mutex
}

// NewEventCapture creates a new event capture instance
func NewEventCapture(config *CaptureConfig, redis *redis.Client, logger *zap.Logger) *EventCapture {
	ctx, cancel := context.WithCancel(context.Background())

	ec := &EventCapture{
		config:     config,
		redis:      redis,
		logger:     logger,
		recordings: make(map[string]*activeRecording),
		eventChan:  make(chan Event, 1000),
		snapChan:   make(chan StateSnapshot, 100),
		ctx:        ctx,
		cancel:     cancel,
	}

	// Start background worker
	ec.wg.Add(1)
	go ec.processEvents()

	return ec
}

// Close shuts down the event capture system
func (ec *EventCapture) Close() error {
	ec.cancel()
	close(ec.eventChan)
	close(ec.snapChan)
	ec.wg.Wait()
	return nil
}

// StartRecording begins recording events for a job
func (ec *EventCapture) StartRecording(jobID string, reason string, importance int) error {
	if !ec.config.Enabled {
		return ErrCaptureDisabled
	}

	recordID := ec.generateRecordID()
	now := time.Now()

	record := &ExecutionRecord{
		JobID:     jobID,
		StartTime: now,
		Events:    make([]Event, 0, ec.config.MaxEvents),
		Snapshots: make(map[string]StateSnapshot),
		Metadata: RecordMetadata{
			RecordID:      recordID,
			CreatedAt:     now,
			RecordedBy:    "time-travel-debugger",
			Reason:        reason,
			Importance:    importance,
			EventCount:    0,
			SnapshotCount: 0,
			Retention:     ec.config.GetRetentionPeriod(importance, false),
		},
		Compressed: false,
		Tags:       []string{reason},
	}

	activeRec := &activeRecording{
		record:       record,
		lastSnapshot: now,
	}

	ec.mu.Lock()
	ec.recordings[jobID] = activeRec
	ec.mu.Unlock()

	ec.logger.Debug("Started recording job execution",
		zap.String("job_id", jobID),
		zap.String("record_id", recordID),
		zap.String("reason", reason),
		zap.Int("importance", importance),
	)

	return nil
}

// CaptureEvent records an event for a job
func (ec *EventCapture) CaptureEvent(jobID string, eventType EventType, jobState *JobState, systemChanges []SystemChange, context map[string]interface{}) error {
	if !ec.config.Enabled {
		return ErrCaptureDisabled
	}

	ec.mu.RLock()
	activeRec, exists := ec.recordings[jobID]
	ec.mu.RUnlock()

	if !exists {
		// Try to auto-start recording if this is a failure or retry
		shouldAutoStart := (eventType == EventFailed || eventType == EventRetrying) &&
			(ec.config.ForceOnFailure || ec.config.ForceOnRetry)

		if shouldAutoStart {
			importance := 5 // medium importance for auto-started recordings
			if eventType == EventFailed {
				importance = 7 // higher importance for failures
			}

			if err := ec.StartRecording(jobID, string(eventType), importance); err != nil {
				return NewCaptureError(jobID, eventType, "failed to auto-start recording", err)
			}

			ec.mu.RLock()
			activeRec, exists = ec.recordings[jobID]
			ec.mu.RUnlock()
		}

		if !exists {
			return NewCaptureError(jobID, eventType, "no active recording found", nil)
		}
	}

	activeRec.mu.Lock()
	defer activeRec.mu.Unlock()

	// Check if we've exceeded max events
	if len(activeRec.record.Events) >= ec.config.MaxEvents {
		return NewCaptureError(jobID, eventType, "max events exceeded", nil)
	}

	// Create performance snapshot
	perfSnapshot := ec.capturePerformanceSnapshot(jobID)

	// Prepare state diff
	var jobStateBefore *JobState
	if len(activeRec.record.Events) > 0 {
		lastEvent := activeRec.record.Events[len(activeRec.record.Events)-1]
		if lastEvent.StateChange.JobStateAfter != nil {
			jobStateBefore = lastEvent.StateChange.JobStateAfter
		}
	}

	event := Event{
		ID:        ec.generateEventID(),
		Timestamp: time.Now(),
		Type:      eventType,
		JobID:     jobID,
		StateChange: StateDiff{
			JobStateBefore:  jobStateBefore,
			JobStateAfter:   jobState,
			SystemChanges:   systemChanges,
			PerformanceData: perfSnapshot,
		},
		Context: context,
	}

	// Add event to recording
	activeRec.record.Events = append(activeRec.record.Events, event)
	activeRec.record.Metadata.EventCount++

	// Check if we need a snapshot
	now := time.Now()
	if now.Sub(activeRec.lastSnapshot) >= ec.config.SnapshotInterval {
		snapshot := ec.createStateSnapshot(jobID, jobState, perfSnapshot)
		timestampKey := now.Format(time.RFC3339Nano)
		activeRec.record.Snapshots[timestampKey] = snapshot
		activeRec.record.Metadata.SnapshotCount++
		activeRec.lastSnapshot = now
	}

	// Send event for async processing (non-blocking)
	select {
	case ec.eventChan <- event:
	default:
		ec.logger.Warn("Event channel full, dropping event",
			zap.String("job_id", jobID),
			zap.String("event_type", string(eventType)),
		)
	}

	return nil
}

// FinishRecording completes and stores a job's recording
func (ec *EventCapture) FinishRecording(jobID string, finalJobState *JobState) error {
	ec.mu.Lock()
	activeRec, exists := ec.recordings[jobID]
	if exists {
		delete(ec.recordings, jobID)
	}
	ec.mu.Unlock()

	if !exists {
		return NewRecordingError("", jobID, "no active recording found", nil)
	}

	activeRec.mu.Lock()
	defer activeRec.mu.Unlock()

	// Set end time and final state
	now := time.Now()
	activeRec.record.EndTime = &now

	// Add final snapshot
	if finalJobState != nil {
		perfSnapshot := ec.capturePerformanceSnapshot(jobID)
		finalSnapshot := ec.createStateSnapshot(jobID, finalJobState, perfSnapshot)
		timestampKey := now.Format(time.RFC3339Nano)
		activeRec.record.Snapshots[timestampKey] = finalSnapshot
		activeRec.record.Metadata.SnapshotCount++
	}

	// Update retention policy based on final state
	isFailure := finalJobState != nil && finalJobState.Status == "failed"
	activeRec.record.Metadata.Retention = ec.config.GetRetentionPeriod(
		activeRec.record.Metadata.Importance,
		isFailure,
	)

	// Calculate compressed size estimate
	data, err := json.Marshal(activeRec.record)
	if err != nil {
		return NewRecordingError(activeRec.record.Metadata.RecordID, jobID, "failed to marshal recording", err)
	}
	activeRec.record.Metadata.Size = int64(len(data))

	// Store to Redis asynchronously
	go ec.storeRecording(activeRec.record)

	ec.logger.Info("Finished recording job execution",
		zap.String("job_id", jobID),
		zap.String("record_id", activeRec.record.Metadata.RecordID),
		zap.Int("event_count", activeRec.record.Metadata.EventCount),
		zap.Int("snapshot_count", activeRec.record.Metadata.SnapshotCount),
		zap.Duration("duration", activeRec.record.Duration()),
		zap.Int64("size_bytes", activeRec.record.Metadata.Size),
	)

	return nil
}

// GetActiveRecordings returns a list of currently active recordings
func (ec *EventCapture) GetActiveRecordings() []string {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	jobIDs := make([]string, 0, len(ec.recordings))
	for jobID := range ec.recordings {
		jobIDs = append(jobIDs, jobID)
	}
	return jobIDs
}

// processEvents handles background processing of events
func (ec *EventCapture) processEvents() {
	defer ec.wg.Done()

	for {
		select {
		case event, ok := <-ec.eventChan:
			if !ok {
				return
			}
			ec.processEvent(event)

		case snapshot, ok := <-ec.snapChan:
			if !ok {
				return
			}
			ec.processSnapshot(snapshot)

		case <-ec.ctx.Done():
			return
		}
	}
}

// processEvent handles background processing of individual events
func (ec *EventCapture) processEvent(event Event) {
	// Add any background processing here (metrics, alerts, etc.)
	ec.logger.Debug("Processed event",
		zap.String("job_id", event.JobID),
		zap.String("event_type", string(event.Type)),
		zap.Time("timestamp", event.Timestamp),
	)
}

// processSnapshot handles background processing of state snapshots
func (ec *EventCapture) processSnapshot(snapshot StateSnapshot) {
	// Add any background processing here
	ec.logger.Debug("Processed snapshot",
		zap.Time("timestamp", snapshot.Timestamp),
	)
}

// storeRecording saves a completed recording to Redis
func (ec *EventCapture) storeRecording(record *ExecutionRecord) {
	data, err := json.Marshal(record)
	if err != nil {
		ec.logger.Error("Failed to marshal recording for storage",
			zap.String("record_id", record.Metadata.RecordID),
			zap.Error(err),
		)
		return
	}

	// Compress if enabled
	if ec.config.CompressionEnabled {
		compressed, err := ec.compressData(data)
		if err != nil {
			ec.logger.Error("Failed to compress recording",
				zap.String("record_id", record.Metadata.RecordID),
				zap.Error(err),
			)
			// Continue with uncompressed data
		} else {
			data = compressed
			record.Compressed = true
			record.Metadata.Size = int64(len(data))
		}
	}

	// Store in Redis with expiration
	key := fmt.Sprintf("time_travel:recording:%s", record.Metadata.RecordID)
	err = ec.redis.Set(ec.ctx, key, data, record.Metadata.Retention).Err()
	if err != nil {
		ec.logger.Error("Failed to store recording in Redis",
			zap.String("record_id", record.Metadata.RecordID),
			zap.String("key", key),
			zap.Error(err),
		)
		return
	}

	// Store job ID mapping for easier lookup
	jobKey := fmt.Sprintf("time_travel:job_index:%s", record.JobID)
	err = ec.redis.Set(ec.ctx, jobKey, record.Metadata.RecordID, record.Metadata.Retention).Err()
	if err != nil {
		ec.logger.Error("Failed to store job index",
			zap.String("job_id", record.JobID),
			zap.String("record_id", record.Metadata.RecordID),
			zap.Error(err),
		)
	}

	ec.logger.Info("Stored recording to Redis",
		zap.String("record_id", record.Metadata.RecordID),
		zap.String("job_id", record.JobID),
		zap.Int64("size_bytes", record.Metadata.Size),
		zap.Bool("compressed", record.Compressed),
		zap.Duration("retention", record.Metadata.Retention),
	)
}

// Helper functions

func (ec *EventCapture) generateRecordID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return "rec_" + hex.EncodeToString(bytes)
}

func (ec *EventCapture) generateEventID() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return "evt_" + hex.EncodeToString(bytes)
}

func (ec *EventCapture) capturePerformanceSnapshot(jobID string) *PerformanceSnapshot {
	// In a real implementation, this would gather actual system metrics
	return &PerformanceSnapshot{
		QueueLength:      0, // Would query Redis
		WorkerCount:      0, // Would query worker registry
		MemoryUsageMB:    0, // Would query system metrics
		CPUUsagePercent:  0, // Would query system metrics
		RedisConnections: 0, // Would query Redis info
		ErrorRate:        0, // Would calculate from recent events
	}
}

func (ec *EventCapture) createStateSnapshot(jobID string, jobState *JobState, perfSnapshot *PerformanceSnapshot) StateSnapshot {
	return StateSnapshot{
		Timestamp: time.Now(),
		JobState:  jobState,
		QueueState: &QueueState{
			Name:        "default", // Would determine from job
			Length:      0,         // Would query Redis
			LastUpdated: time.Now(),
		},
		SystemMetrics: perfSnapshot,
	}
}

func (ec *EventCapture) compressData(data []byte) ([]byte, error) {
	var compressed bytes.Buffer

	// Use gzip compression
	writer := gzip.NewWriter(&compressed)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return compressed.Bytes(), nil
}
