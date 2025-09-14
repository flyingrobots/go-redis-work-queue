// Copyright 2025 James Ross
package integration

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/exactly_once"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDuplicateInjectionUnderLoad tests duplicate handling under concurrent load
func TestDuplicateInjectionUnderLoad(t *testing.T) {
	// Setup Redis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	ctx := context.Background()
	manager := exactly_once.NewRedisIdempotencyManager(client, "load_test", time.Hour)

	// Test parameters
	numWorkers := 50
	numJobsPerWorker := 100
	numUniqueJobs := 500 // Less than total to ensure duplicates

	// Track results
	var processedCount int32
	var duplicateCount int32
	processedJobs := sync.Map{}

	// Generate job IDs with intentional duplicates
	jobIDs := make([]string, numWorkers*numJobsPerWorker)
	for i := range jobIDs {
		jobIDs[i] = fmt.Sprintf("job_%d", i%numUniqueJobs)
	}

	// Worker function
	worker := func(workerID int) {
		for i := 0; i < numJobsPerWorker; i++ {
			jobID := jobIDs[workerID*numJobsPerWorker+i]

			isDuplicate, err := manager.CheckAndReserve(ctx, jobID, time.Hour)
			if err != nil {
				continue
			}

			if isDuplicate {
				atomic.AddInt32(&duplicateCount, 1)
			} else {
				atomic.AddInt32(&processedCount, 1)
				processedJobs.Store(jobID, true)
			}
		}
	}

	// Launch workers
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker(id)
		}(i)
	}
	wg.Wait()
	duration := time.Since(start)

	// Verify results
	totalAttempts := int32(numWorkers * numJobsPerWorker)
	assert.Equal(t, totalAttempts, processedCount+duplicateCount, "all jobs should be accounted for")
	assert.Equal(t, int32(numUniqueJobs), processedCount, "should process exactly unique job count")
	assert.Greater(t, duplicateCount, int32(0), "should have detected duplicates")

	// Count actual unique jobs processed
	uniqueCount := 0
	processedJobs.Range(func(key, value interface{}) bool {
		uniqueCount++
		return true
	})
	assert.Equal(t, numUniqueJobs, uniqueCount, "should have processed all unique jobs")

	// Performance check
	throughput := float64(totalAttempts) / duration.Seconds()
	t.Logf("Processed %d attempts in %v (%.0f ops/sec)", totalAttempts, duration, throughput)
	t.Logf("Unique: %d, Duplicates: %d", processedCount, duplicateCount)

	// Get stats
	stats, err := manager.Stats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(processedCount), stats.Processed)
	assert.Equal(t, int64(duplicateCount), stats.Duplicates)
}

// TestPaymentProcessingScenario simulates e-commerce payment processing with exactly-once
func TestPaymentProcessingScenario(t *testing.T) {
	// Setup Redis
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	ctx := context.Background()

	// Use content-based key generator for payments
	keyGen := exactly_once.NewContentHashGenerator("payments", nil)
	manager := exactly_once.NewRedisIdempotencyManager(client, "payments", time.Hour)

	// Payment processor
	type Payment struct {
		UserID    string  `json:"user_id"`
		Amount    float64 `json:"amount"`
		OrderID   string  `json:"order_id"`
		Timestamp int64   `json:"timestamp"`
	}

	processedPayments := sync.Map{}
	chargedAmount := float64(0)
	var chargedMutex sync.Mutex

	processPayment := func(payment Payment) error {
		// Generate idempotency key based on payment content
		key := keyGen.Generate(payment)

		isDuplicate, err := manager.CheckAndReserve(ctx, key, time.Hour)
		if err != nil {
			return err
		}

		if isDuplicate {
			return fmt.Errorf("duplicate payment detected for order %s", payment.OrderID)
		}

		// Simulate payment processing
		time.Sleep(10 * time.Millisecond)

		// Record successful payment
		processedPayments.Store(payment.OrderID, payment)

		chargedMutex.Lock()
		chargedAmount += payment.Amount
		chargedMutex.Unlock()

		// Confirm successful processing
		return manager.Confirm(ctx, key)
	}

	// Simulate double-click scenario
	t.Run("double click protection", func(t *testing.T) {
		payment := Payment{
			UserID:    "user_123",
			Amount:    99.99,
			OrderID:   "order_456",
			Timestamp: time.Now().Unix(),
		}

		// Multiple concurrent attempts (simulating double clicks)
		var wg sync.WaitGroup
		successCount := 0
		var successMutex sync.Mutex

		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := processPayment(payment)
				if err == nil {
					successMutex.Lock()
					successCount++
					successMutex.Unlock()
				}
			}()
		}
		wg.Wait()

		assert.Equal(t, 1, successCount, "only one payment should succeed")

		// Verify charged amount
		assert.Equal(t, 99.99, chargedAmount, "should charge exactly once")
	})

	// Reset for next test
	chargedAmount = 0

	// Simulate retry with different order IDs but same content
	t.Run("content-based deduplication", func(t *testing.T) {
		// Same payment details, but system might generate different order IDs
		basePayment := Payment{
			UserID:    "user_789",
			Amount:    149.99,
			Timestamp: 1234567890, // Fixed timestamp for consistent hash
		}

		// Process with first order ID
		payment1 := basePayment
		payment1.OrderID = "order_001"
		err := processPayment(payment1)
		assert.NoError(t, err)

		// Try with different order ID but same other fields
		payment2 := basePayment
		payment2.OrderID = "order_002"
		err = processPayment(payment2)
		assert.Error(t, err, "should detect duplicate based on content")

		// Only one charge should have been made
		assert.Equal(t, 149.99, chargedAmount, "should charge only once despite different order IDs")
	})
}

// TestDatabaseOutboxIntegration tests the transactional outbox pattern with a real database
func TestDatabaseOutboxIntegration(t *testing.T) {
	// Setup database
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE business_data (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id TEXT NOT NULL,
			action TEXT NOT NULL,
			amount REAL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE outbox_events (
			id VARCHAR(255) PRIMARY KEY,
			queue_name VARCHAR(255) NOT NULL,
			payload TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL,
			processed_at TIMESTAMP,
			status VARCHAR(50) NOT NULL DEFAULT 'pending',
			attempts INT NOT NULL DEFAULT 0,
			last_error TEXT
		);
	`)
	require.NoError(t, err)

	// Setup Redis for idempotency
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	// Setup managers
	ctx := context.Background()
	dedupManager := exactly_once.NewRedisIdempotencyManager(client, "outbox_test", time.Hour)

	// Mock queue to track enqueued jobs
	enqueuedJobs := sync.Map{}
	mockQueue := &mockQueueIntegration{
		enqueuedJobs: &enqueuedJobs,
	}

	outboxManager := exactly_once.NewSQLOutboxManager(db, mockQueue, dedupManager)

	// Business operation that must be atomic with event publishing
	executeBusinessOperation := func(userID string, action string, amount float64) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()

		// Business logic
		businessLogic := func(tx *sql.Tx) error {
			_, err := tx.Exec(
				"INSERT INTO business_data (user_id, action, amount) VALUES (?, ?, ?)",
				userID, action, amount,
			)
			return err
		}

		// Create outbox event
		eventPayload, _ := json.Marshal(map[string]interface{}{
			"user_id": userID,
			"action":  action,
			"amount":  amount,
		})

		event := exactly_once.OutboxEvent{
			ID:        fmt.Sprintf("%s_%s_%d", userID, action, time.Now().UnixNano()),
			QueueName: "business_events",
			Payload:   eventPayload,
		}

		// Execute with outbox
		if err := outboxManager.ExecuteWithOutbox(ctx, tx, businessLogic, event); err != nil {
			return err
		}

		return tx.Commit()
	}

	t.Run("atomic business operation with event", func(t *testing.T) {
		// Execute business operation
		err := executeBusinessOperation("user_1", "purchase", 299.99)
		require.NoError(t, err)

		// Verify business data was saved
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM business_data WHERE user_id = 'user_1'").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "business data should be saved")

		// Verify outbox event was created
		err = db.QueryRow("SELECT COUNT(*) FROM outbox_events WHERE status = 'pending'").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "outbox event should be created")

		// Process outbox events
		err = outboxManager.ProcessPending(ctx)
		require.NoError(t, err)

		// Verify event was enqueued
		enqueuedCount := 0
		enqueuedJobs.Range(func(key, value interface{}) bool {
			enqueuedCount++
			return true
		})
		assert.Equal(t, 1, enqueuedCount, "event should be enqueued")

		// Verify idempotency - process again
		err = outboxManager.ProcessPending(ctx)
		require.NoError(t, err)

		// Should still have only one enqueued job
		enqueuedCount = 0
		enqueuedJobs.Range(func(key, value interface{}) bool {
			enqueuedCount++
			return true
		})
		assert.Equal(t, 1, enqueuedCount, "should not enqueue duplicate")
	})

	t.Run("failure rollback atomicity", func(t *testing.T) {
		// Simulate a failure in business logic
		tx, err := db.Begin()
		require.NoError(t, err)

		failingLogic := func(tx *sql.Tx) error {
			// Start inserting data
			_, err := tx.Exec(
				"INSERT INTO business_data (user_id, action, amount) VALUES (?, ?, ?)",
				"user_2", "refund", 100.00,
			)
			if err != nil {
				return err
			}
			// Simulate failure after partial work
			return fmt.Errorf("simulated business error")
		}

		event := exactly_once.OutboxEvent{
			ID:        "failed_event",
			QueueName: "business_events",
			Payload:   json.RawMessage(`{}`),
		}

		err = outboxManager.ExecuteWithOutbox(ctx, tx, failingLogic, event)
		assert.Error(t, err)

		tx.Rollback()

		// Verify nothing was committed
		var businessCount, eventCount int
		err = db.QueryRow("SELECT COUNT(*) FROM business_data WHERE user_id = 'user_2'").Scan(&businessCount)
		require.NoError(t, err)
		err = db.QueryRow("SELECT COUNT(*) FROM outbox_events WHERE id = 'failed_event'").Scan(&eventCount)
		require.NoError(t, err)

		assert.Equal(t, 0, businessCount, "business data should not be saved")
		assert.Equal(t, 0, eventCount, "outbox event should not be saved")
	})
}

// mockQueueIntegration implements the Queue interface for integration tests
type mockQueueIntegration struct {
	enqueuedJobs *sync.Map
}

func (m *mockQueueIntegration) Enqueue(ctx context.Context, queueName string, payload []byte, idempotencyKey string) error {
	m.enqueuedJobs.Store(idempotencyKey, struct {
		QueueName string
		Payload   []byte
	}{
		QueueName: queueName,
		Payload:   payload,
	})
	return nil
}

// BenchmarkIdempotencyCheck benchmarks the idempotency check performance
func BenchmarkIdempotencyCheck(b *testing.B) {
	mr, err := miniredis.Run()
	require.NoError(b, err)
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	defer client.Close()

	ctx := context.Background()
	manager := exactly_once.NewRedisIdempotencyManager(client, "bench", time.Hour)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("bench_key_%d", i)
			_, _ = manager.CheckAndReserve(ctx, key, time.Hour)
			i++
		}
	})

	b.ReportMetric(float64(b.N)/b.Elapsed().Seconds(), "ops/s")
}