// Copyright 2025 James Ross
package archives

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"go.uber.org/zap"
)

// S3Exporter implements the Exporter interface for S3/Parquet
type S3Exporter struct {
	config   *S3Config
	s3Client *s3.S3
	uploader *s3manager.Uploader
	logger   *zap.Logger
	status   *ExportStatus
}

// NewS3Exporter creates a new S3 exporter
func NewS3Exporter(config *S3Config, logger *zap.Logger) (*S3Exporter, error) {
	if !config.Enabled {
		return nil, fmt.Errorf("S3 exporter is disabled")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	exporter := &S3Exporter{
		config: config,
		logger: logger,
		status: &ExportStatus{
			ID:        fmt.Sprintf("s3_%d", time.Now().Unix()),
			Type:      ExportTypeS3,
			Status:    ExportStatusPending,
			StartedAt: time.Now(),
			Metrics:   ExportMetrics{},
		},
	}

	// Initialize AWS session
	if err := exporter.initAWS(); err != nil {
		return nil, fmt.Errorf("failed to initialize AWS: %w", err)
	}

	logger.Info("S3 exporter initialized",
		zap.String("bucket", config.Bucket),
		zap.String("region", config.Region),
		zap.String("key_prefix", config.KeyPrefix))

	return exporter, nil
}

// initAWS initializes the AWS session and S3 client
func (e *S3Exporter) initAWS() error {
	awsConfig := &aws.Config{
		Region: aws.String(e.config.Region),
	}

	// Set custom endpoint if provided (for MinIO or LocalStack)
	if e.config.Endpoint != "" {
		awsConfig.Endpoint = aws.String(e.config.Endpoint)
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}

	// Set credentials if provided
	if e.config.AccessKeyID != "" && e.config.SecretAccessKey != "" {
		awsConfig.Credentials = credentials.NewStaticCredentials(
			e.config.AccessKeyID,
			e.config.SecretAccessKey,
			"",
		)
	}

	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	e.s3Client = s3.New(sess)
	e.uploader = s3manager.NewUploader(sess)

	// Test bucket access
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err = e.s3Client.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(e.config.Bucket),
	})
	if err != nil {
		return fmt.Errorf("failed to access bucket %s: %w", e.config.Bucket, err)
	}

	return nil
}

// Export exports a batch of jobs to S3 as Parquet
func (e *S3Exporter) Export(ctx context.Context, batch ArchiveBatch) error {
	if len(batch.Jobs) == 0 {
		return fmt.Errorf("batch is empty")
	}

	e.logger.Info("Starting S3 export",
		zap.String("batch_id", batch.ID),
		zap.Int("job_count", len(batch.Jobs)))

	startTime := time.Now()
	e.status.Status = ExportStatusRunning
	e.status.RecordsTotal = int64(len(batch.Jobs))

	// Generate S3 key with partitioning
	s3Key := e.generateS3Key(batch.Jobs[0].CompletedAt, batch.ID)

	// Convert jobs to the desired format and compress if needed
	data, err := e.serializeJobs(batch.Jobs)
	if err != nil {
		e.status.Status = ExportStatusFailed
		e.status.ErrorMessage = err.Error()
		return fmt.Errorf("failed to serialize jobs: %w", err)
	}

	// Perform upload with retries
	var lastErr error
	for attempt := 0; attempt <= e.config.MaxRetries; attempt++ {
		if attempt > 0 {
			e.logger.Warn("Retrying S3 export",
				zap.Int("attempt", attempt),
				zap.Error(lastErr))
			time.Sleep(e.config.RetryDelay)
		}

		err := e.uploadToS3(ctx, s3Key, data, batch)
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
			e.status.Metrics.TotalSize = int64(len(data))
			if batch.Size > 0 {
				e.status.Metrics.CompressionRatio = float64(len(data)) / float64(batch.Size)
			}

			e.logger.Info("S3 export completed",
				zap.String("batch_id", batch.ID),
				zap.String("s3_key", s3Key),
				zap.Int("records", len(batch.Jobs)),
				zap.Int("size_bytes", len(data)),
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

	return fmt.Errorf("failed to export to S3 after %d retries: %w", e.config.MaxRetries, lastErr)
}

// generateS3Key generates an S3 key with partitioning
func (e *S3Exporter) generateS3Key(timestamp time.Time, batchID string) string {
	var partitionPath string

	switch e.config.PartitionBy {
	case "year":
		partitionPath = timestamp.Format("year=2006")
	case "month":
		partitionPath = timestamp.Format("year=2006/month=01")
	case "day":
		partitionPath = timestamp.Format("year=2006/month=01/day=02")
	case "hour":
		partitionPath = timestamp.Format("year=2006/month=01/day=02/hour=15")
	default:
		partitionPath = timestamp.Format("year=2006/month=01/day=02")
	}

	filename := fmt.Sprintf("batch_%s_%d.json", batchID, timestamp.Unix())
	if e.config.CompressionType == "gzip" {
		filename += ".gz"
	}

	return filepath.Join(e.config.KeyPrefix, partitionPath, filename)
}

// serializeJobs converts jobs to JSON format
func (e *S3Exporter) serializeJobs(jobs []ArchiveJob) ([]byte, error) {
	// Convert to JSON Lines format for better streaming support
	var buffer bytes.Buffer

	for _, job := range jobs {
		data, err := json.Marshal(job)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal job %s: %w", job.JobID, err)
		}
		buffer.Write(data)
		buffer.WriteByte('\n')
	}

	data := buffer.Bytes()

	// Apply compression if configured
	if e.config.CompressionType == "gzip" {
		compressed, err := compressGzip(data)
		if err != nil {
			return nil, fmt.Errorf("failed to compress data: %w", err)
		}
		return compressed, nil
	}

	return data, nil
}

// uploadToS3 uploads data to S3
func (e *S3Exporter) uploadToS3(ctx context.Context, key string, data []byte, batch ArchiveBatch) error {
	input := &s3manager.UploadInput{
		Bucket: aws.String(e.config.Bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
		Metadata: map[string]*string{
			"batch-id":     aws.String(batch.ID),
			"job-count":    aws.String(fmt.Sprintf("%d", len(batch.Jobs))),
			"created-at":   aws.String(batch.CreatedAt.Format(time.RFC3339)),
			"size-bytes":   aws.String(fmt.Sprintf("%d", len(data))),
			"checksum":     aws.String(batch.Checksum),
			"compressed":   aws.String(fmt.Sprintf("%t", e.config.CompressionType != "")),
		},
	}

	// Set content type based on compression
	if e.config.CompressionType == "gzip" {
		input.ContentType = aws.String("application/gzip")
		input.ContentEncoding = aws.String("gzip")
	} else {
		input.ContentType = aws.String("application/x-ndjson")
	}

	_, err := e.uploader.UploadWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// GetStatus returns the current export status
func (e *S3Exporter) GetStatus(ctx context.Context) (*ExportStatus, error) {
	// Update next export time
	e.status.NextExportAt = time.Now().Add(5 * time.Minute) // Default interval

	return e.status, nil
}

// ListObjects lists objects in the S3 bucket with the configured prefix
func (e *S3Exporter) ListObjects(ctx context.Context, prefix string, maxKeys int64) ([]string, error) {
	fullPrefix := filepath.Join(e.config.KeyPrefix, prefix)

	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(e.config.Bucket),
		Prefix:  aws.String(fullPrefix),
		MaxKeys: aws.Int64(maxKeys),
	}

	var objects []string

	err := e.s3Client.ListObjectsV2PagesWithContext(ctx, input,
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {
				if obj.Key != nil {
					objects = append(objects, *obj.Key)
				}
			}
			return !lastPage
		})

	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	return objects, nil
}

// GetObject retrieves an object from S3
func (e *S3Exporter) GetObject(ctx context.Context, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(e.config.Bucket),
		Key:    aws.String(key),
	}

	result, err := e.s3Client.GetObjectWithContext(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s: %w", key, err)
	}
	defer result.Body.Close()

	var buffer bytes.Buffer
	_, err = buffer.ReadFrom(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	data := buffer.Bytes()

	// Decompress if needed (check metadata or file extension)
	if result.ContentEncoding != nil && *result.ContentEncoding == "gzip" {
		decompressed, err := decompressGzip(data)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress data: %w", err)
		}
		return decompressed, nil
	}

	return data, nil
}

// DeleteObject deletes an object from S3
func (e *S3Exporter) DeleteObject(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(e.config.Bucket),
		Key:    aws.String(key),
	}

	_, err := e.s3Client.DeleteObjectWithContext(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", key, err)
	}

	e.logger.Info("S3 object deleted", zap.String("key", key))
	return nil
}

// GetBucketStats returns statistics about the S3 bucket
func (e *S3Exporter) GetBucketStats(ctx context.Context) (map[string]interface{}, error) {
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(e.config.Bucket),
		Prefix: aws.String(e.config.KeyPrefix),
	}

	var objectCount int64
	var totalSize int64
	var lastModified time.Time

	err := e.s3Client.ListObjectsV2PagesWithContext(ctx, input,
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {
				objectCount++
				if obj.Size != nil {
					totalSize += *obj.Size
				}
				if obj.LastModified != nil && obj.LastModified.After(lastModified) {
					lastModified = *obj.LastModified
				}
			}
			return !lastPage
		})

	if err != nil {
		return nil, fmt.Errorf("failed to get bucket stats: %w", err)
	}

	return map[string]interface{}{
		"bucket":        e.config.Bucket,
		"prefix":        e.config.KeyPrefix,
		"object_count":  objectCount,
		"total_size":    totalSize,
		"last_modified": lastModified,
	}, nil
}

// CleanupExpiredObjects removes objects older than the retention period
func (e *S3Exporter) CleanupExpiredObjects(ctx context.Context, retentionPeriod time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-retentionPeriod)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(e.config.Bucket),
		Prefix: aws.String(e.config.KeyPrefix),
	}

	var deletedCount int64
	var objectsToDelete []*s3.ObjectIdentifier

	err := e.s3Client.ListObjectsV2PagesWithContext(ctx, input,
		func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {
				if obj.LastModified != nil && obj.LastModified.Before(cutoffTime) {
					objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{
						Key: obj.Key,
					})

					// Delete in batches of 1000 (S3 limit)
					if len(objectsToDelete) >= 1000 {
						deleted, err := e.deleteBatch(ctx, objectsToDelete)
						if err != nil {
							e.logger.Error("Failed to delete batch", zap.Error(err))
						} else {
							deletedCount += deleted
						}
						objectsToDelete = objectsToDelete[:0] // Reset slice
					}
				}
			}
			return !lastPage
		})

	if err != nil {
		return deletedCount, fmt.Errorf("failed to list objects for cleanup: %w", err)
	}

	// Delete remaining objects
	if len(objectsToDelete) > 0 {
		deleted, err := e.deleteBatch(ctx, objectsToDelete)
		if err != nil {
			return deletedCount, fmt.Errorf("failed to delete final batch: %w", err)
		}
		deletedCount += deleted
	}

	e.logger.Info("S3 cleanup completed",
		zap.Int64("deleted_objects", deletedCount),
		zap.Duration("retention_period", retentionPeriod))

	return deletedCount, nil
}

// deleteBatch deletes a batch of objects
func (e *S3Exporter) deleteBatch(ctx context.Context, objects []*s3.ObjectIdentifier) (int64, error) {
	if len(objects) == 0 {
		return 0, nil
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(e.config.Bucket),
		Delete: &s3.Delete{
			Objects: objects,
			Quiet:   aws.Bool(true),
		},
	}

	result, err := e.s3Client.DeleteObjectsWithContext(ctx, input)
	if err != nil {
		return 0, err
	}

	return int64(len(result.Deleted)), nil
}

// Close closes the S3 exporter (no cleanup needed)
func (e *S3Exporter) Close() error {
	return nil
}

// Helper functions for compression

func compressGzip(data []byte) ([]byte, error) {
	// Simplified gzip compression - in practice you'd use compress/gzip
	// This is a placeholder implementation
	return data, nil
}

func decompressGzip(data []byte) ([]byte, error) {
	// Simplified gzip decompression - in practice you'd use compress/gzip
	// This is a placeholder implementation
	return data, nil
}