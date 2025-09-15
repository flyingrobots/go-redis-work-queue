// Copyright 2025 James Ross
package archives

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// manager implements the Manager interface
type manager struct {
	config           *Config
	redis            *redis.Client
	logger           *zap.Logger
	exporters        map[ExportType]Exporter
	schemaManager    SchemaManager
	retentionManager RetentionManager
	queryTemplates   map[string]QueryTemplate
}

// NewManager creates a new long-term archives manager
func NewManager(config *Config, logger *zap.Logger) (Manager, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
		DB:   config.RedisDB,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	m := &manager{
		config:         config,
		redis:          rdb,
		logger:         logger,
		exporters:      make(map[ExportType]Exporter),
		queryTemplates: make(map[string]QueryTemplate),
	}

	// Initialize exporters
	if err := m.initializeExporters(); err != nil {
		return nil, fmt.Errorf("failed to initialize exporters: %w", err)
	}

	// Initialize schema manager
	m.schemaManager = NewSchemaManager(rdb, &config.Archive, logger)

	// Initialize retention manager
	m.retentionManager = NewRetentionManager(rdb, &config.Archive.Retention, m.exporters, logger)

	// Initialize default schema evolutions
	if err := m.schemaManager.(*schemaManager).InitializeDefaultEvolutions(ctx); err != nil {
		logger.Warn("Failed to initialize default schema evolutions", zap.Error(err))
	}

	// Initialize default query templates
	if err := m.initializeDefaultQueryTemplates(ctx); err != nil {
		logger.Warn("Failed to initialize default query templates", zap.Error(err))
	}

	logger.Info("Long-term archives manager initialized",
		zap.String("redis_addr", config.RedisAddr),
		zap.Bool("archive_enabled", config.Archive.Enabled),
		zap.Int("exporters", len(m.exporters)))

	return m, nil
}

// initializeExporters sets up configured exporters
func (m *manager) initializeExporters() error {
	// Initialize ClickHouse exporter if enabled
	if m.config.Archive.ClickHouse.Enabled {
		clickhouseExporter, err := NewClickHouseExporter(&m.config.Archive.ClickHouse, m.logger)
		if err != nil {
			return fmt.Errorf("failed to initialize ClickHouse exporter: %w", err)
		}
		m.exporters[ExportTypeClickHouse] = clickhouseExporter
		m.logger.Info("ClickHouse exporter initialized")
	}

	// Initialize S3 exporter if enabled
	if m.config.Archive.S3.Enabled {
		s3Exporter, err := NewS3Exporter(&m.config.Archive.S3, m.logger)
		if err != nil {
			return fmt.Errorf("failed to initialize S3 exporter: %w", err)
		}
		m.exporters[ExportTypeS3] = s3Exporter
		m.logger.Info("S3 exporter initialized")
	}

	if len(m.exporters) == 0 {
		m.logger.Warn("No exporters enabled")
	}

	return nil
}

// ArchiveJob archives a completed job
func (m *manager) ArchiveJob(ctx context.Context, job ArchiveJob) error {
	if !m.config.Archive.Enabled {
		return fmt.Errorf("archiving is disabled")
	}

	// Apply sampling
	if rand.Float64() > m.config.Archive.SamplingRate {
		return nil // Skip this job due to sampling
	}

	// Set archive timestamp and schema version
	job.ArchivedAt = time.Now()

	// Validate schema
	if err := m.schemaManager.(*schemaManager).ValidateSchema(ctx, job); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	// Handle payload based on configuration
	if err := m.processPayload(&job); err != nil {
		return fmt.Errorf("payload processing failed: %w", err)
	}

	// Add to Redis stream
	streamKey := m.config.Archive.RedisStreamKey
	if streamKey == "" {
		streamKey = "archive:stream:completed"
	}

	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	_, err = m.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"job_data": string(jobData),
			"job_id":   job.JobID,
			"queue":    job.Queue,
			"outcome":  string(job.Outcome),
		},
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to add to stream: %w", err)
	}

	m.logger.Debug("Job archived to stream",
		zap.String("job_id", job.JobID),
		zap.String("queue", job.Queue),
		zap.String("outcome", string(job.Outcome)))

	return nil
}

// processPayload handles job payload according to configuration
func (m *manager) processPayload(job *ArchiveJob) error {
	payloadConfig := m.config.Archive.PayloadHandling

	if !payloadConfig.IncludePayload {
		job.PayloadSnapshot = nil
		return nil
	}

	// Check size limits
	if payloadConfig.MaxPayloadSize > 0 && int64(len(job.PayloadSnapshot)) > payloadConfig.MaxPayloadSize {
		if payloadConfig.HashOnly {
			// Just keep the hash
			hash := sha256.Sum256(job.PayloadSnapshot)
			job.PayloadHash = hex.EncodeToString(hash[:])
			job.PayloadSnapshot = nil
		} else {
			// Truncate
			job.PayloadSnapshot = job.PayloadSnapshot[:payloadConfig.MaxPayloadSize]
		}
	}

	// Hash the payload if configured
	if payloadConfig.HashOnly {
		hash := sha256.Sum256(job.PayloadSnapshot)
		job.PayloadHash = hex.EncodeToString(hash[:])
		job.PayloadSnapshot = nil
	} else if job.PayloadSnapshot != nil {
		// Generate hash for verification even if keeping payload
		hash := sha256.Sum256(job.PayloadSnapshot)
		job.PayloadHash = hex.EncodeToString(hash[:])
	}

	return nil
}

// ExportJobs exports a batch of jobs using the specified exporter
func (m *manager) ExportJobs(ctx context.Context, jobs []ArchiveJob, exportType ExportType) (*ExportStatus, error) {
	exporter, exists := m.exporters[exportType]
	if !exists {
		return nil, fmt.Errorf("exporter %s not available", exportType)
	}

	if len(jobs) == 0 {
		return nil, fmt.Errorf("no jobs to export")
	}

	// Create batch
	batch := ArchiveBatch{
		ID:        fmt.Sprintf("batch_%d_%s", time.Now().Unix(), exportType),
		Jobs:      jobs,
		CreatedAt: time.Now(),
		Size:      m.calculateBatchSize(jobs),
	}

	// Generate checksum
	batch.Checksum = m.generateChecksum(batch)

	// Export the batch
	err := exporter.Export(ctx, batch)
	if err != nil {
		return nil, fmt.Errorf("export failed: %w", err)
	}

	// Get status
	status, err := exporter.GetStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get export status: %w", err)
	}

	// Store export status
	statusKey := fmt.Sprintf("archive:export:%s", status.ID)
	statusData, _ := json.Marshal(status)
	m.redis.Set(ctx, statusKey, statusData, 24*time.Hour)

	m.logger.Info("Export completed",
		zap.String("batch_id", batch.ID),
		zap.String("export_type", string(exportType)),
		zap.Int("job_count", len(jobs)),
		zap.String("status", string(status.Status)))

	return status, nil
}

// calculateBatchSize calculates the total size of a batch
func (m *manager) calculateBatchSize(jobs []ArchiveJob) int64 {
	var size int64
	for _, job := range jobs {
		size += job.PayloadSize
		size += int64(len(job.PayloadSnapshot))
		// Add estimated size for other fields
		size += 1024 // Rough estimate for metadata
	}
	return size
}

// generateChecksum generates a checksum for the batch
func (m *manager) generateChecksum(batch ArchiveBatch) string {
	hasher := sha256.New()
	hasher.Write([]byte(batch.ID))
	hasher.Write([]byte(batch.CreatedAt.Format(time.RFC3339)))
	for _, job := range batch.Jobs {
		hasher.Write([]byte(job.JobID))
		hasher.Write([]byte(job.PayloadHash))
	}
	return hex.EncodeToString(hasher.Sum(nil))
}

// GetExportStatus returns the status of an export operation
func (m *manager) GetExportStatus(ctx context.Context, id string) (*ExportStatus, error) {
	statusKey := fmt.Sprintf("archive:export:%s", id)

	data, err := m.redis.Get(ctx, statusKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("export status not found")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get export status: %w", err)
	}

	var status ExportStatus
	err = json.Unmarshal([]byte(data), &status)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal export status: %w", err)
	}

	return &status, nil
}

// ListExports returns a list of export operations
func (m *manager) ListExports(ctx context.Context, limit int, offset int) ([]ExportStatus, error) {
	pattern := "archive:export:*"
	keys, err := m.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list export keys: %w", err)
	}

	// Get export statuses
	var exports []ExportStatus
	for i, key := range keys {
		if i < offset {
			continue
		}
		if len(exports) >= limit {
			break
		}

		data, err := m.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var status ExportStatus
		err = json.Unmarshal([]byte(data), &status)
		if err != nil {
			continue
		}

		exports = append(exports, status)
	}

	// Sort by creation time (most recent first)
	for i := 0; i < len(exports)-1; i++ {
		for j := i + 1; j < len(exports); j++ {
			if exports[i].StartedAt.Before(exports[j].StartedAt) {
				exports[i], exports[j] = exports[j], exports[i]
			}
		}
	}

	return exports, nil
}

// CancelExport cancels an export operation
func (m *manager) CancelExport(ctx context.Context, id string) error {
	// Update status to canceled
	statusKey := fmt.Sprintf("archive:export:%s", id)

	data, err := m.redis.Get(ctx, statusKey).Result()
	if err == redis.Nil {
		return fmt.Errorf("export not found")
	} else if err != nil {
		return fmt.Errorf("failed to get export status: %w", err)
	}

	var status ExportStatus
	err = json.Unmarshal([]byte(data), &status)
	if err != nil {
		return fmt.Errorf("failed to unmarshal export status: %w", err)
	}

	if status.Status == ExportStatusCompleted || status.Status == ExportStatusFailed {
		return fmt.Errorf("cannot cancel export in status %s", status.Status)
	}

	status.Status = ExportStatusCanceled
	now := time.Now()
	status.CompletedAt = &now

	statusData, _ := json.Marshal(status)
	err = m.redis.Set(ctx, statusKey, statusData, 24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to update export status: %w", err)
	}

	m.logger.Info("Export canceled", zap.String("export_id", id))
	return nil
}

// GetArchivedJob retrieves an archived job by ID
func (m *manager) GetArchivedJob(ctx context.Context, jobID string) (*ArchiveJob, error) {
	// Try to find in exporters
	for _, exporter := range m.exporters {
		if chExporter, ok := exporter.(*ClickHouseExporter); ok {
			query := SearchQuery{
				JobIDs: []string{jobID},
				Limit:  1,
			}
			jobs, err := chExporter.QueryJobs(ctx, query)
			if err == nil && len(jobs) > 0 {
				return &jobs[0], nil
			}
		}
	}

	return nil, fmt.Errorf("job %s not found in archives", jobID)
}

// SearchJobs searches for archived jobs
func (m *manager) SearchJobs(ctx context.Context, query SearchQuery) ([]ArchiveJob, error) {
	// Use ClickHouse exporter for search if available
	if chExporter, exists := m.exporters[ExportTypeClickHouse]; exists {
		if ch, ok := chExporter.(*ClickHouseExporter); ok {
			return ch.QueryJobs(ctx, query)
		}
	}

	// Fallback to simple Redis-based search (limited)
	return m.searchInRedis(ctx, query)
}

// searchInRedis performs a simple search in Redis stream
func (m *manager) searchInRedis(ctx context.Context, query SearchQuery) ([]ArchiveJob, error) {
	streamKey := m.config.Archive.RedisStreamKey
	if streamKey == "" {
		streamKey = "archive:stream:completed"
	}

	// Read from stream
	result, err := m.redis.XRead(ctx, &redis.XReadArgs{
		Streams: []string{streamKey, "0"},
		Count:   int64(query.Limit),
	}).Result()

	if err != nil {
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	var jobs []ArchiveJob
	for _, stream := range result {
		for _, message := range stream.Messages {
			if jobData, ok := message.Values["job_data"].(string); ok {
				var job ArchiveJob
				err := json.Unmarshal([]byte(jobData), &job)
				if err != nil {
					continue
				}

				// Apply filters
				if m.jobMatchesQuery(job, query) {
					jobs = append(jobs, job)
				}
			}
		}
	}

	return jobs, nil
}

// jobMatchesQuery checks if a job matches the search query
func (m *manager) jobMatchesQuery(job ArchiveJob, query SearchQuery) bool {
	if len(query.JobIDs) > 0 {
		found := false
		for _, id := range query.JobIDs {
			if job.JobID == id {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if query.Queue != "" && job.Queue != query.Queue {
		return false
	}

	if query.Outcome != "" && job.Outcome != query.Outcome {
		return false
	}

	if query.WorkerID != "" && job.WorkerID != query.WorkerID {
		return false
	}

	if query.TraceID != "" && job.TraceID != query.TraceID {
		return false
	}

	if query.StartTime != nil && job.CompletedAt.Before(*query.StartTime) {
		return false
	}

	if query.EndTime != nil && job.CompletedAt.After(*query.EndTime) {
		return false
	}

	// Check tags
	for key, value := range query.Tags {
		if jobValue, exists := job.Tags[key]; !exists || jobValue != value {
			return false
		}
	}

	return true
}

// GetStats returns archive statistics
func (m *manager) GetStats(ctx context.Context, window time.Duration) (*ArchiveStats, error) {
	// Use ClickHouse exporter for stats if available
	if chExporter, exists := m.exporters[ExportTypeClickHouse]; exists {
		if ch, ok := chExporter.(*ClickHouseExporter); ok {
			return ch.GetStats(ctx, window)
		}
	}

	// Fallback to Redis-based stats
	return m.getRedisStats(ctx, window)
}

// getRedisStats gets statistics from Redis
func (m *manager) getRedisStats(ctx context.Context, window time.Duration) (*ArchiveStats, error) {
	streamKey := m.config.Archive.RedisStreamKey
	if streamKey == "" {
		streamKey = "archive:stream:completed"
	}

	// Get stream info
	info, err := m.redis.XInfoStream(ctx, streamKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get stream info: %w", err)
	}

	return &ArchiveStats{
		TotalJobs:       info.Length,
		JobsByOutcome:   make(map[JobOutcome]int64),
		JobsByQueue:     make(map[string]int64),
		OldestJob:       time.Unix(0, info.FirstEntry.ID.Time*int64(time.Millisecond)),
		NewestJob:       time.Unix(0, info.LastEntry.ID.Time*int64(time.Millisecond)),
		LastExportAt:    time.Now(),
		ExportLag:       time.Since(time.Unix(0, info.LastEntry.ID.Time*int64(time.Millisecond))),
	}, nil
}

// GetSchemaVersion returns the current schema version
func (m *manager) GetSchemaVersion(ctx context.Context) (int, error) {
	return m.schemaManager.GetCurrentVersion(ctx)
}

// UpgradeSchema upgrades to a new schema version
func (m *manager) UpgradeSchema(ctx context.Context, newVersion int) error {
	return m.schemaManager.Upgrade(ctx, newVersion)
}

// GetSchemaEvolution returns schema evolution history
func (m *manager) GetSchemaEvolution(ctx context.Context) ([]SchemaEvolution, error) {
	return m.schemaManager.GetEvolution(ctx)
}

// CleanupExpired removes expired data
func (m *manager) CleanupExpired(ctx context.Context) (int64, error) {
	return m.retentionManager.Cleanup(ctx)
}

// ProcessGDPRDelete processes a GDPR deletion request
func (m *manager) ProcessGDPRDelete(ctx context.Context, request GDPRDeleteRequest) error {
	return m.retentionManager.ProcessGDPRDelete(ctx, request)
}

// GetHealth returns health status of the archive system
func (m *manager) GetHealth(ctx context.Context) (map[string]interface{}, error) {
	health := map[string]interface{}{
		"service":   "long-term-archives",
		"timestamp": time.Now(),
		"status":    "ok",
	}

	// Check Redis connectivity
	if err := m.redis.Ping(ctx).Err(); err != nil {
		health["status"] = "degraded"
		health["redis_error"] = err.Error()
	} else {
		health["redis"] = "connected"
	}

	// Check exporters
	exporterStatus := make(map[string]interface{})
	for exportType, exporter := range m.exporters {
		status, err := exporter.GetStatus(ctx)
		if err != nil {
			exporterStatus[string(exportType)] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
		} else {
			exporterStatus[string(exportType)] = map[string]interface{}{
				"status":           status.Status,
				"last_export":      status.LastExportAt,
				"records_exported": status.RecordsExported,
				"success_rate":     status.Metrics.SuccessRate,
			}
		}
	}
	health["exporters"] = exporterStatus

	// Get schema version
	version, err := m.GetSchemaVersion(ctx)
	if err == nil {
		health["schema_version"] = version
	}

	return health, nil
}

// Close closes the manager and cleans up resources
func (m *manager) Close() error {
	var lastErr error

	// Close exporters
	for _, exporter := range m.exporters {
		if err := exporter.Close(); err != nil {
			lastErr = err
		}
	}

	// Close Redis connection
	if err := m.redis.Close(); err != nil {
		lastErr = err
	}

	return lastErr
}

// Additional methods for query templates and batch processing

// initializeDefaultQueryTemplates sets up default query templates
func (m *manager) initializeDefaultQueryTemplates(ctx context.Context) error {
	defaultTemplates := []QueryTemplate{
		{
			Name:        "failed_jobs_last_hour",
			Description: "Failed jobs in the last hour",
			SQL:         "SELECT * FROM archives WHERE outcome = 'failed' AND completed_at >= now() - interval 1 hour",
			Parameters:  []QueryParameter{},
			Tags:        []string{"monitoring", "failures"},
			CreatedAt:   time.Now(),
		},
		{
			Name:        "jobs_by_queue",
			Description: "Jobs by queue in a time range",
			SQL:         "SELECT queue, count(*) as job_count FROM archives WHERE completed_at BETWEEN ? AND ? GROUP BY queue",
			Parameters: []QueryParameter{
				{Name: "start_time", Type: "timestamp", Description: "Start time", Required: true},
				{Name: "end_time", Type: "timestamp", Description: "End time", Required: true},
			},
			Tags:      []string{"analytics", "queues"},
			CreatedAt: time.Now(),
		},
	}

	for _, template := range defaultTemplates {
		err := m.AddQueryTemplate(ctx, template)
		if err != nil {
			m.logger.Warn("Failed to add default query template",
				zap.String("name", template.Name),
				zap.Error(err))
		}
	}

	return nil
}

// AddQueryTemplate adds a new query template
func (m *manager) AddQueryTemplate(ctx context.Context, template QueryTemplate) error {
	templateKey := fmt.Sprintf("archive:query_template:%s", template.Name)

	templateData, err := json.Marshal(template)
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	err = m.redis.Set(ctx, templateKey, templateData, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to store template: %w", err)
	}

	m.queryTemplates[template.Name] = template

	m.logger.Info("Query template added",
		zap.String("name", template.Name),
		zap.String("description", template.Description))

	return nil
}

// GetQueryTemplates returns all available query templates
func (m *manager) GetQueryTemplates(ctx context.Context) ([]QueryTemplate, error) {
	pattern := "archive:query_template:*"
	keys, err := m.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get template keys: %w", err)
	}

	var templates []QueryTemplate
	for _, key := range keys {
		data, err := m.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var template QueryTemplate
		err = json.Unmarshal([]byte(data), &template)
		if err != nil {
			continue
		}

		templates = append(templates, template)
	}

	return templates, nil
}

// ExecuteQuery executes a query template with parameters
func (m *manager) ExecuteQuery(ctx context.Context, templateName string, params map[string]interface{}) (interface{}, error) {
	// Get template
	templateKey := fmt.Sprintf("archive:query_template:%s", templateName)
	data, err := m.redis.Get(ctx, templateKey).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("template not found: %s", templateName)
	} else if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	var template QueryTemplate
	err = json.Unmarshal([]byte(data), &template)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal template: %w", err)
	}

	// Validate parameters
	for _, param := range template.Parameters {
		if param.Required {
			if _, exists := params[param.Name]; !exists {
				return nil, fmt.Errorf("required parameter missing: %s", param.Name)
			}
		}
	}

	// Execute query using ClickHouse exporter if available
	if chExporter, exists := m.exporters[ExportTypeClickHouse]; exists {
		if ch, ok := chExporter.(*ClickHouseExporter); ok {
			// In a real implementation, you'd parse the SQL and execute it
			// For now, we'll return a placeholder result
			return map[string]interface{}{
				"template": templateName,
				"params":   params,
				"executed": time.Now(),
				"result":   "Query execution not implemented in this demo",
			}, nil
		}
	}

	return nil, fmt.Errorf("no suitable query executor available")
}