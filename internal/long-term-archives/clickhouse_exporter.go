// Copyright 2025 James Ross
package archives

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.uber.org/zap"
)

// ClickHouseExporter implements the Exporter interface for ClickHouse
type ClickHouseExporter struct {
	config *ClickHouseConfig
	db     *sql.DB
	logger *zap.Logger
	status *ExportStatus
}

// NewClickHouseExporter creates a new ClickHouse exporter
func NewClickHouseExporter(config *ClickHouseConfig, logger *zap.Logger) (*ClickHouseExporter, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("ClickHouse exporter is disabled")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	exporter := &ClickHouseExporter{
		config: config,
		logger: logger,
		status: &ExportStatus{
			ID:         fmt.Sprintf("clickhouse_%d", time.Now().Unix()),
			Type:       ExportTypeClickHouse,
			Status:     ExportStatusPending,
			StartedAt:  time.Now(),
			Metrics:    ExportMetrics{},
		},
	}

	// Initialize database connection
	if err := exporter.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Ensure table exists
	if err := exporter.ensureTable(); err != nil {
		return nil, fmt.Errorf("failed to ensure table exists: %w", err)
	}

	logger.Info("ClickHouse exporter initialized",
		zap.String("database", config.Database),
		zap.String("table", config.Table))

	return exporter, nil
}

// connect establishes a connection to ClickHouse
func (e *ClickHouseExporter) connect() error {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{e.config.DSN},
		Auth: clickhouse.Auth{
			Database: e.config.Database,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:      time.Duration(30) * time.Second,
		MaxOpenConns:     e.config.MaxOpenConns,
		MaxIdleConns:     e.config.MaxIdleConns,
		ConnMaxLifetime:  e.config.ConnMaxLife,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := conn.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	e.db = conn
	return nil
}

// ensureTable creates the archive table if it doesn't exist
func (e *ClickHouseExporter) ensureTable() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	createTableSQL := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.%s (
			job_id String,
			queue String,
			priority Int32,
			enqueued_at DateTime64(3),
			started_at Nullable(DateTime64(3)),
			completed_at DateTime64(3),
			outcome LowCardinality(String),
			retry_count UInt32,
			worker_id String,
			payload_size UInt64,
			trace_id String,
			error_message String,
			error_code String,
			processing_time_ms UInt64,
			payload_hash String,
			payload_snapshot String,
			tags Map(String, String),
			schema_version UInt32,
			archived_at DateTime64(3),
			tenant String,
			job_type String
		) ENGINE = MergeTree()
		PARTITION BY toYYYYMM(completed_at)
		ORDER BY (queue, completed_at, job_id)
		TTL completed_at + INTERVAL 1 YEAR DELETE
		SETTINGS index_granularity = 8192
	`, e.config.Database, e.config.Table)

	_, err := e.db.ExecContext(ctx, createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	e.logger.Info("ClickHouse table ensured",
		zap.String("database", e.config.Database),
		zap.String("table", e.config.Table))

	return nil
}

// Export exports a batch of jobs to ClickHouse
func (e *ClickHouseExporter) Export(ctx context.Context, batch ArchiveBatch) error {
	if len(batch.Jobs) == 0 {
		return fmt.Errorf("batch is empty")
	}

	e.logger.Info("Starting ClickHouse export",
		zap.String("batch_id", batch.ID),
		zap.Int("job_count", len(batch.Jobs)))

	startTime := time.Now()
	e.status.Status = ExportStatusRunning
	e.status.RecordsTotal = int64(len(batch.Jobs))

	// Prepare batch insert with retries
	var lastErr error
	for attempt := 0; attempt <= e.config.MaxRetries; attempt++ {
		if attempt > 0 {
			e.logger.Warn("Retrying ClickHouse export",
				zap.Int("attempt", attempt),
				zap.Error(lastErr))
			time.Sleep(e.config.RetryDelay)
		}

		err := e.executeBatchInsert(ctx, batch.Jobs)
		if err == nil {
			// Success
			e.status.Status = ExportStatusCompleted
			e.status.RecordsExported = int64(len(batch.Jobs))
			completedAt := time.Now()
			e.status.CompletedAt = &completedAt
			e.status.LastExportAt = completedAt

			// Update metrics
			duration := completedAt.Sub(startTime)
			e.status.Metrics.AvgExportTime = float64(duration.Milliseconds())
			e.status.Metrics.AvgBatchSize = float64(len(batch.Jobs))
			e.status.Metrics.SuccessRate = 1.0
			e.status.Metrics.TotalSize = batch.Size

			e.logger.Info("ClickHouse export completed",
				zap.String("batch_id", batch.ID),
				zap.Int("records", len(batch.Jobs)),
				zap.Duration("duration", duration))

			return nil
		}

		lastErr = err
		e.status.RecordsFailed = int64(len(batch.Jobs))
	}

	// All retries failed
	e.status.Status = ExportStatusFailed
	e.status.ErrorMessage = lastErr.Error()
	completedAt := time.Now()
	e.status.CompletedAt = &completedAt

	// Update metrics for failure
	e.status.Metrics.ErrorRate = 1.0
	e.status.Metrics.SuccessRate = 0.0

	return fmt.Errorf("failed to export to ClickHouse after %d retries: %w", e.config.MaxRetries, lastErr)
}

// executeBatchInsert performs the actual batch insert to ClickHouse
func (e *ClickHouseExporter) executeBatchInsert(ctx context.Context, jobs []ArchiveJob) error {
	tx, err := e.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	insertSQL := fmt.Sprintf(`
		INSERT INTO %s.%s (
			job_id, queue, priority, enqueued_at, started_at, completed_at,
			outcome, retry_count, worker_id, payload_size, trace_id,
			error_message, error_code, processing_time_ms, payload_hash,
			payload_snapshot, tags, schema_version, archived_at, tenant, job_type
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, e.config.Database, e.config.Table)

	stmt, err := tx.PrepareContext(ctx, insertSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, job := range jobs {
		_, err := stmt.ExecContext(ctx,
			job.JobID,
			job.Queue,
			job.Priority,
			job.EnqueuedAt,
			job.StartedAt,
			job.CompletedAt,
			string(job.Outcome),
			job.RetryCount,
			job.WorkerID,
			job.PayloadSize,
			job.TraceID,
			job.ErrorMessage,
			job.ErrorCode,
			job.ProcessingTime,
			job.PayloadHash,
			string(job.PayloadSnapshot),
			job.Tags,
			job.SchemaVersion,
			job.ArchivedAt,
			job.Tenant,
			job.JobType,
		)
		if err != nil {
			return fmt.Errorf("failed to insert job %s: %w", job.JobID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetStatus returns the current export status
func (e *ClickHouseExporter) GetStatus(ctx context.Context) (*ExportStatus, error) {
	// Update next export time
	e.status.NextExportAt = time.Now().Add(5 * time.Minute) // Default interval

	return e.status, nil
}

// QueryJobs queries jobs from ClickHouse
func (e *ClickHouseExporter) QueryJobs(ctx context.Context, query SearchQuery) ([]ArchiveJob, error) {
	whereClause, args := e.buildWhereClause(query)

	querySQL := fmt.Sprintf(`
		SELECT
			job_id, queue, priority, enqueued_at, started_at, completed_at,
			outcome, retry_count, worker_id, payload_size, trace_id,
			error_message, error_code, processing_time_ms, payload_hash,
			payload_snapshot, tags, schema_version, archived_at, tenant, job_type
		FROM %s.%s
		%s
		ORDER BY %s %s
		LIMIT %d OFFSET %d
	`, e.config.Database, e.config.Table, whereClause,
		query.OrderBy, query.OrderDir, query.Limit, query.Offset)

	rows, err := e.db.QueryContext(ctx, querySQL, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query jobs: %w", err)
	}
	defer rows.Close()

	var jobs []ArchiveJob
	for rows.Next() {
		var job ArchiveJob
		var outcomeStr string
		var payloadSnapshot sql.NullString

		err := rows.Scan(
			&job.JobID,
			&job.Queue,
			&job.Priority,
			&job.EnqueuedAt,
			&job.StartedAt,
			&job.CompletedAt,
			&outcomeStr,
			&job.RetryCount,
			&job.WorkerID,
			&job.PayloadSize,
			&job.TraceID,
			&job.ErrorMessage,
			&job.ErrorCode,
			&job.ProcessingTime,
			&job.PayloadHash,
			&payloadSnapshot,
			&job.Tags,
			&job.SchemaVersion,
			&job.ArchivedAt,
			&job.Tenant,
			&job.JobType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		job.Outcome = JobOutcome(outcomeStr)
		if payloadSnapshot.Valid {
			job.PayloadSnapshot = []byte(payloadSnapshot.String)
		}

		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return jobs, nil
}

// GetStats returns statistics from ClickHouse
func (e *ClickHouseExporter) GetStats(ctx context.Context, window time.Duration) (*ArchiveStats, error) {
	since := time.Now().Add(-window)

	statsSQL := fmt.Sprintf(`
		SELECT
			count() as total_jobs,
			sum(payload_size) as total_size,
			outcome,
			queue,
			avg(processing_time_ms) as avg_processing_time,
			min(completed_at) as oldest_job,
			max(completed_at) as newest_job
		FROM %s.%s
		WHERE completed_at >= ?
		GROUP BY outcome, queue
	`, e.config.Database, e.config.Table)

	rows, err := e.db.QueryContext(ctx, statsSQL, since)
	if err != nil {
		return nil, fmt.Errorf("failed to query stats: %w", err)
	}
	defer rows.Close()

	stats := &ArchiveStats{
		JobsByOutcome: make(map[JobOutcome]int64),
		JobsByQueue:   make(map[string]int64),
		LastExportAt:  e.status.LastExportAt,
	}

	for rows.Next() {
		var totalJobs, totalSize int64
		var outcome, queue string
		var avgProcessingTime float64
		var oldestJob, newestJob time.Time

		err := rows.Scan(
			&totalJobs,
			&totalSize,
			&outcome,
			&queue,
			&avgProcessingTime,
			&oldestJob,
			&newestJob,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stats: %w", err)
		}

		stats.TotalJobs += totalJobs
		stats.TotalSize += totalSize
		stats.JobsByOutcome[JobOutcome(outcome)] += totalJobs
		stats.JobsByQueue[queue] += totalJobs

		if stats.AvgProcessTime == 0 {
			stats.AvgProcessTime = avgProcessingTime
		} else {
			stats.AvgProcessTime = (stats.AvgProcessTime + avgProcessingTime) / 2
		}

		if stats.OldestJob.IsZero() || oldestJob.Before(stats.OldestJob) {
			stats.OldestJob = oldestJob
		}

		if stats.NewestJob.IsZero() || newestJob.After(stats.NewestJob) {
			stats.NewestJob = newestJob
		}
	}

	// Calculate error rate
	if stats.TotalJobs > 0 {
		failedJobs := stats.JobsByOutcome[OutcomeFailed] + stats.JobsByOutcome[OutcomeTimeout]
		stats.ErrorRate = float64(failedJobs) / float64(stats.TotalJobs)
	}

	// Calculate export lag
	if !stats.NewestJob.IsZero() {
		stats.ExportLag = time.Since(stats.NewestJob)
	}

	return stats, nil
}

// buildWhereClause builds the WHERE clause for queries
func (e *ClickHouseExporter) buildWhereClause(query SearchQuery) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if len(query.JobIDs) > 0 {
		conditions = append(conditions, "job_id IN (?)")
		args = append(args, query.JobIDs)
	}

	if query.Queue != "" {
		conditions = append(conditions, "queue = ?")
		args = append(args, query.Queue)
	}

	if query.Outcome != "" {
		conditions = append(conditions, "outcome = ?")
		args = append(args, string(query.Outcome))
	}

	if query.WorkerID != "" {
		conditions = append(conditions, "worker_id = ?")
		args = append(args, query.WorkerID)
	}

	if query.TraceID != "" {
		conditions = append(conditions, "trace_id = ?")
		args = append(args, query.TraceID)
	}

	if query.StartTime != nil {
		conditions = append(conditions, "completed_at >= ?")
		args = append(args, *query.StartTime)
	}

	if query.EndTime != nil {
		conditions = append(conditions, "completed_at <= ?")
		args = append(args, *query.EndTime)
	}

	if len(query.Tags) > 0 {
		for key, value := range query.Tags {
			conditions = append(conditions, "tags[?] = ?")
			args = append(args, key, value)
		}
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + conditions[0]
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
	}

	return whereClause, args
}

// Close closes the ClickHouse connection
func (e *ClickHouseExporter) Close() error {
	if e.db != nil {
		return e.db.Close()
	}
	return nil
}

// GetTableInfo returns information about the ClickHouse table
func (e *ClickHouseExporter) GetTableInfo(ctx context.Context) (map[string]interface{}, error) {
	infoSQL := fmt.Sprintf(`
		SELECT
			table,
			engine,
			total_rows,
			total_bytes,
			formatReadableSize(total_bytes) as total_size
		FROM system.tables
		WHERE database = ? AND name = ?
	`, e.config.Database, e.config.Table)

	row := e.db.QueryRowContext(ctx, infoSQL, e.config.Database, e.config.Table)

	var table, engine, totalSize string
	var totalRows, totalBytes int64

	err := row.Scan(&table, &engine, &totalRows, &totalBytes, &totalSize)
	if err != nil {
		return nil, fmt.Errorf("failed to get table info: %w", err)
	}

	return map[string]interface{}{
		"table":       table,
		"engine":      engine,
		"total_rows":  totalRows,
		"total_bytes": totalBytes,
		"total_size":  totalSize,
		"database":    e.config.Database,
	}, nil
}

// OptimizeTable runs OPTIMIZE TABLE to compact data
func (e *ClickHouseExporter) OptimizeTable(ctx context.Context) error {
	optimizeSQL := fmt.Sprintf("OPTIMIZE TABLE %s.%s", e.config.Database, e.config.Table)

	_, err := e.db.ExecContext(ctx, optimizeSQL)
	if err != nil {
		return fmt.Errorf("failed to optimize table: %w", err)
	}

	e.logger.Info("ClickHouse table optimized",
		zap.String("database", e.config.Database),
		zap.String("table", e.config.Table))

	return nil
}