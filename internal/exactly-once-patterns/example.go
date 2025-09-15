// Copyright 2025 James Ross
package exactlyonce

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Example demonstrates how to use the exactly-once patterns

// ProcessJobWithIdempotency shows how to process a job with idempotency guarantees
func ExampleProcessJob(manager *Manager, jobID string, jobData map[string]interface{}) error {
	// Generate or receive idempotency key
	key := manager.GenerateIdempotencyKey("file-processing", "", jobID)

	// Process the job with idempotency
	result, err := manager.ProcessWithIdempotency(context.Background(), key, func() (interface{}, error) {
		// This is where you put your actual job processing logic
		fmt.Printf("Processing job %s with data: %+v\n", jobID, jobData)

		// Simulate some processing time
		time.Sleep(100 * time.Millisecond)

		// Return the result
		return map[string]interface{}{
			"status": "completed",
			"processed_at": time.Now().UTC(),
			"job_id": jobID,
		}, nil
	})

	if err != nil {
		return fmt.Errorf("job processing failed: %w", err)
	}

	fmt.Printf("Job %s result: %+v\n", jobID, result)
	return nil
}

// ExampleTransactionalOutbox shows how to use the transactional outbox pattern
func ExampleTransactionalOutbox(manager *Manager) error {
	ctx := context.Background()

	// Create an outbox event
	event := OutboxEvent{
		AggregateID: "user-123",
		EventType:   "user.created",
		Payload:     json.RawMessage(`{"user_id": "123", "email": "user@example.com"}`),
		Headers: map[string]string{
			"source": "user-service",
			"version": "1.0",
		},
	}

	// In a real application, you would do this within a database transaction
	// For this example, we'll just simulate it
	fmt.Println("Starting database transaction...")

	// Store the event in the outbox as part of the transaction
	err := manager.StoreInOutbox(ctx, event) // tx would be your database transaction
	if err != nil {
		return fmt.Errorf("failed to store outbox event: %w", err)
	}

	fmt.Println("Event stored in outbox successfully")

	// Later, publish pending outbox events
	err = manager.PublishOutboxEvents(ctx)
	if err != nil {
		return fmt.Errorf("failed to publish outbox events: %w", err)
	}

	fmt.Println("Outbox events published successfully")
	return nil
}

// ExampleIdempotencyHook shows how to implement and register processing hooks
type LoggingHook struct {
	log *zap.Logger
}

func (h *LoggingHook) BeforeProcessing(ctx context.Context, jobID string, idempotencyKey IdempotencyKey) error {
	h.log.Info("Starting job processing",
		zap.String("job_id", jobID),
		zap.String("idempotency_key", idempotencyKey.ID),
		zap.String("queue", idempotencyKey.QueueName))
	return nil
}

func (h *LoggingHook) AfterProcessing(ctx context.Context, jobID string, result interface{}, err error) error {
	if err != nil {
		h.log.Error("Job processing failed",
			zap.String("job_id", jobID),
			zap.Error(err))
	} else {
		h.log.Info("Job processing completed",
			zap.String("job_id", jobID),
			zap.Any("result", result))
	}
	return nil
}

func (h *LoggingHook) OnDuplicate(ctx context.Context, jobID string, existingResult interface{}) error {
	h.log.Info("Duplicate job detected, returning cached result",
		zap.String("job_id", jobID),
		zap.Any("cached_result", existingResult))
	return nil
}

func ExampleWithHooks(rdb *redis.Client, log *zap.Logger) error {
	// Create manager with default config
	cfg := DefaultConfig()
	manager := NewManager(cfg, rdb, log)

	// Register hooks
	hook := &LoggingHook{log: log}
	manager.RegisterHook(hook)

	// Process a job
	return ExampleProcessJob(manager, "job-123", map[string]interface{}{
		"file_path": "/tmp/example.txt",
		"priority": "high",
	})
}

// ExampleCustomConfiguration shows how to create a custom configuration
func ExampleCustomConfiguration() *Config {
	cfg := DefaultConfig()

	// Customize idempotency settings
	cfg.Idempotency.DefaultTTL = 2 * time.Hour
	cfg.Idempotency.KeyPrefix = "myapp:idempotency:"
	cfg.Idempotency.CleanupInterval = 30 * time.Minute

	// Customize Redis storage settings
	cfg.Idempotency.Storage.Redis.UseHashes = true
	cfg.Idempotency.Storage.Redis.KeyPattern = "{queue}:dedupe:{tenant}:{key}"
	cfg.Idempotency.Storage.Redis.HashKeyPattern = "{queue}:dedupe:{tenant}"

	// Enable outbox pattern
	cfg.Outbox.Enabled = true
	cfg.Outbox.BatchSize = 25
	cfg.Outbox.PollInterval = 10 * time.Second

	// Configure metrics
	cfg.Metrics.CollectionInterval = 15 * time.Second
	cfg.Metrics.HistogramBuckets = []float64{0.01, 0.1, 0.5, 1.0, 2.5, 5.0, 10.0}

	return cfg
}

// ExampleAdminAPI shows how to set up the admin API
func ExampleAdminAPI(manager *Manager, log *zap.Logger) error {
	// Create admin handler
	handler := NewAdminHandler(manager, log)

	// Create HTTP server
	mux := http.NewServeMux()

	// Register routes with middleware
	mux.HandleFunc("/api/v1/exactly-once/stats",
		handler.CORSMiddleware(handler.LoggingMiddleware(handler.handleStats)))
	mux.HandleFunc("/api/v1/exactly-once/idempotency",
		handler.CORSMiddleware(handler.LoggingMiddleware(handler.handleIdempotencyKey)))
	mux.HandleFunc("/api/v1/exactly-once/outbox",
		handler.CORSMiddleware(handler.LoggingMiddleware(handler.handleOutbox)))
	mux.HandleFunc("/api/v1/exactly-once/cleanup",
		handler.CORSMiddleware(handler.LoggingMiddleware(handler.handleCleanup)))
	mux.HandleFunc("/api/v1/exactly-once/health",
		handler.CORSMiddleware(handler.LoggingMiddleware(handler.handleHealth)))

	log.Info("Starting admin API server on :8080")

	// In a real application, you'd want to configure timeouts, TLS, etc.
	// return http.ListenAndServe(":8080", mux)
	return nil // Just an example
}

// ExampleBatchProcessing shows how to process multiple jobs efficiently
func ExampleBatchProcessing(manager *Manager, jobs []map[string]interface{}) error {
	ctx := context.Background()

	for i, _ := range jobs {
		jobID := fmt.Sprintf("batch-job-%d", i)

		// Generate idempotency key
		key := manager.GenerateIdempotencyKey("batch-processing", "", jobID)

		// Process with idempotency
		_, err := manager.ProcessWithIdempotency(ctx, key, func() (interface{}, error) {
			fmt.Printf("Processing batch job %s\n", jobID)

			// Simulate processing
			time.Sleep(50 * time.Millisecond)

			return map[string]interface{}{
				"job_id": jobID,
				"status": "completed",
				"batch_index": i,
			}, nil
		})

		if err != nil {
			return fmt.Errorf("batch job %s failed: %w", jobID, err)
		}
	}

	fmt.Printf("Successfully processed %d batch jobs\n", len(jobs))
	return nil
}

// ExampleMonitoring shows how to get metrics and stats
func ExampleMonitoring(manager *Manager) error {
	ctx := context.Background()

	// Get deduplication stats for a queue
	stats, err := manager.GetDedupStats(ctx, "file-processing", "")
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	fmt.Printf("Queue: %s\n", stats.QueueName)
	fmt.Printf("Total Keys: %d\n", stats.TotalKeys)
	fmt.Printf("Hit Rate: %.2f%%\n", stats.HitRate * 100)
	fmt.Printf("Duplicates Avoided: %d\n", stats.DuplicatesAvoided)
	fmt.Printf("Last Updated: %s\n", stats.LastUpdated.Format(time.RFC3339))

	return nil
}