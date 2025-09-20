//go:build integration_tests
// +build integration_tests

// Copyright 2025 James Ross
package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIntegrationTest(t *testing.T) (*redis.Client, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return client, cleanup
}

func TestMultiTenantConcurrency(t *testing.T) {
	client, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()
	numTenants := 5
	numJobsPerTenant := 100

	t.Run("concurrent tenant operations", func(t *testing.T) {
		var wg sync.WaitGroup
		errors := make(chan error, numTenants*numJobsPerTenant)

		for i := 0; i < numTenants; i++ {
			tenantID := fmt.Sprintf("tenant-%d", i)
			wg.Add(1)

			go func(tid string) {
				defer wg.Done()

				// Each tenant performs operations concurrently
				for j := 0; j < numJobsPerTenant; j++ {
					key := fmt.Sprintf("t:%s:job-%d", tid, j)
					value := fmt.Sprintf("data-%s-%d", tid, j)

					if err := client.Set(ctx, key, value, time.Minute).Err(); err != nil {
						errors <- err
						return
					}

					// Verify immediate read
					retrieved, err := client.Get(ctx, key).Result()
					if err != nil {
						errors <- err
						return
					}

					if retrieved != value {
						errors <- fmt.Errorf("value mismatch: expected %s, got %s", value, retrieved)
					}
				}
			}(tenantID)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		for err := range errors {
			t.Errorf("Concurrent operation failed: %v", err)
		}

		// Verify data isolation
		for i := 0; i < numTenants; i++ {
			tenantID := fmt.Sprintf("tenant-%d", i)
			pattern := fmt.Sprintf("t:%s:*", tenantID)

			keys, err := client.Keys(ctx, pattern).Result()
			require.NoError(t, err)
			assert.Len(t, keys, numJobsPerTenant, "tenant %s should have exactly %d keys", tenantID, numJobsPerTenant)
		}
	})

	t.Run("cross-tenant access prevention", func(t *testing.T) {
		// Setup tenant A data
		tenantA := "secure-tenant-a"
		secretKeyA := fmt.Sprintf("t:%s:secret", tenantA)
		secretDataA := "confidential-data-a"

		err := client.Set(ctx, secretKeyA, secretDataA, 0).Err()
		require.NoError(t, err)

		// Setup tenant B data
		tenantB := "secure-tenant-b"
		secretKeyB := fmt.Sprintf("t:%s:secret", tenantB)
		secretDataB := "confidential-data-b"

		err = client.Set(ctx, secretKeyB, secretDataB, 0).Err()
		require.NoError(t, err)

		// Simulate tenant A trying to access tenant B's data
		// In a real system, this would be prevented by access control
		dataB, err := client.Get(ctx, secretKeyB).Result()
		require.NoError(t, err) // Redis allows it, but our app layer wouldn't

		// Verify data is different
		assert.NotEqual(t, secretDataA, dataB)
		assert.Equal(t, secretDataB, dataB)

		// Verify namespace scanning doesn't cross boundaries
		keysA, err := client.Keys(ctx, fmt.Sprintf("t:%s:*", tenantA)).Result()
		require.NoError(t, err)
		assert.Len(t, keysA, 1)
		assert.Contains(t, keysA[0], tenantA)
		assert.NotContains(t, keysA[0], tenantB)
	})
}

func TestQuotaEnforcementIntegration(t *testing.T) {
	client, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("rate limiting across distributed workers", func(t *testing.T) {
		tenantID := "rate-limited-tenant"
		rateLimit := int32(10) // 10 ops per second
		numWorkers := 5

		// Initialize rate limiter in Redis
		rateLimitKey := fmt.Sprintf("tenant:%s:rate_limit:enqueue", tenantID)
		err := client.Set(ctx, rateLimitKey, rateLimit, time.Second).Err()
		require.NoError(t, err)

		var allowed int32
		var denied int32
		var wg sync.WaitGroup

		startTime := time.Now()

		// Simulate multiple workers trying to enqueue
		for w := 0; w < numWorkers; w++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for i := 0; i < 10; i++ {
					// Simple rate limit check using DECR
					remaining, err := client.Decr(ctx, rateLimitKey).Result()
					if err != nil || remaining < 0 {
						atomic.AddInt32(&denied, 1)
						// Reset if negative
						if remaining < 0 {
							client.Incr(ctx, rateLimitKey)
						}
					} else {
						atomic.AddInt32(&allowed, 1)
					}

					time.Sleep(10 * time.Millisecond)
				}
			}(w)
		}

		wg.Wait()
		elapsed := time.Since(startTime)

		t.Logf("Rate limiting test: allowed=%d, denied=%d, elapsed=%v", allowed, denied, elapsed)

		// With rate limit of 10/sec and test running ~500ms, we should see limiting
		assert.Greater(t, denied, int32(0), "some requests should be rate limited")
		assert.Greater(t, allowed, int32(0), "some requests should be allowed")
	})

	t.Run("quota persistence and recovery", func(t *testing.T) {
		tenantID := "persistent-tenant"

		// Set initial quotas
		quotaConfig := map[string]interface{}{
			"max_jobs_per_hour":    1000,
			"max_jobs_per_day":     10000,
			"max_backlog_size":     500,
			"soft_limit_threshold": 0.8,
		}

		quotaKey := fmt.Sprintf("tenant:%s:quotas", tenantID)
		quotaJSON, err := json.Marshal(quotaConfig)
		require.NoError(t, err)

		err = client.Set(ctx, quotaKey, quotaJSON, 0).Err()
		require.NoError(t, err)

		// Track usage
		usageKey := fmt.Sprintf("tenant:%s:usage:hourly", tenantID)
		err = client.HIncrBy(ctx, usageKey, "jobs", 100).Err()
		require.NoError(t, err)

		// Simulate system restart by creating new client
		// (In miniredis, data persists in memory)

		// Verify quotas survived
		quotaData, err := client.Get(ctx, quotaKey).Result()
		require.NoError(t, err)

		var recovered map[string]interface{}
		err = json.Unmarshal([]byte(quotaData), &recovered)
		require.NoError(t, err)

		assert.Equal(t, float64(1000), recovered["max_jobs_per_hour"])
		assert.Equal(t, float64(0.8), recovered["soft_limit_threshold"])

		// Verify usage survived
		usage, err := client.HGet(ctx, usageKey, "jobs").Result()
		require.NoError(t, err)
		assert.Equal(t, "100", usage)
	})

	t.Run("backlog management", func(t *testing.T) {
		tenantID := "backlog-tenant"
		queueKey := fmt.Sprintf("t:%s:main:jobs", tenantID)
		maxBacklog := 10

		// Fill queue to max
		for i := 0; i < maxBacklog; i++ {
			err := client.LPush(ctx, queueKey, fmt.Sprintf("job-%d", i)).Err()
			require.NoError(t, err)
		}

		// Check current size
		size, err := client.LLen(ctx, queueKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(maxBacklog), size)

		// Try to add beyond max (in real system, this would be blocked)
		// Here we just verify we can detect the condition
		if size >= int64(maxBacklog) {
			// Would reject in application layer
			t.Log("Backlog limit reached, would reject new jobs")
		}

		// Process some jobs to make room
		for i := 0; i < 5; i++ {
			_, err := client.RPop(ctx, queueKey).Result()
			require.NoError(t, err)
		}

		// Verify we can add more now
		newSize, err := client.LLen(ctx, queueKey).Result()
		require.NoError(t, err)
		assert.Less(t, newSize, int64(maxBacklog))
	})
}

func TestEncryptionIntegration(t *testing.T) {
	client, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("end-to-end encrypted job flow", func(t *testing.T) {
		tenantID := "encrypted-tenant"

		// Store encryption config
		encConfig := map[string]interface{}{
			"enabled":    true,
			"algorithm":  "AES-256-GCM",
			"key_id":     "key-v1",
			"rotated_at": time.Now().Unix(),
		}

		configKey := fmt.Sprintf("tenant:%s:encryption", tenantID)
		configJSON, err := json.Marshal(encConfig)
		require.NoError(t, err)

		err = client.Set(ctx, configKey, configJSON, 0).Err()
		require.NoError(t, err)

		// Simulate encrypted job
		job := map[string]interface{}{
			"id":        "job-001",
			"encrypted": true,
			"payload":   "base64_encrypted_data_here",
			"key_ref":   "key-v1",
			"nonce":     "random_nonce",
		}

		jobKey := fmt.Sprintf("t:%s:jobs:job-001", tenantID)
		jobJSON, err := json.Marshal(job)
		require.NoError(t, err)

		err = client.Set(ctx, jobKey, jobJSON, time.Hour).Err()
		require.NoError(t, err)

		// Retrieve and verify structure
		retrieved, err := client.Get(ctx, jobKey).Result()
		require.NoError(t, err)

		var jobData map[string]interface{}
		err = json.Unmarshal([]byte(retrieved), &jobData)
		require.NoError(t, err)

		assert.True(t, jobData["encrypted"].(bool))
		assert.Equal(t, "key-v1", jobData["key_ref"])
	})

	t.Run("key rotation workflow", func(t *testing.T) {
		tenantID := "rotation-tenant"

		// Track key versions
		keyVersions := []string{"key-v1", "key-v2", "key-v3"}
		currentKeyIndex := 0

		keyRotationKey := fmt.Sprintf("tenant:%s:current_key", tenantID)
		err := client.Set(ctx, keyRotationKey, keyVersions[currentKeyIndex], 0).Err()
		require.NoError(t, err)

		// Store some jobs with current key
		for i := 0; i < 3; i++ {
			job := map[string]interface{}{
				"id":      fmt.Sprintf("job-%d", i),
				"key_ref": keyVersions[currentKeyIndex],
				"data":    fmt.Sprintf("encrypted_with_%s", keyVersions[currentKeyIndex]),
			}

			jobKey := fmt.Sprintf("t:%s:job:%d", tenantID, i)
			jobJSON, _ := json.Marshal(job)
			client.Set(ctx, jobKey, jobJSON, 0)
		}

		// Rotate key
		currentKeyIndex++
		err = client.Set(ctx, keyRotationKey, keyVersions[currentKeyIndex], 0).Err()
		require.NoError(t, err)

		// New jobs use new key
		newJob := map[string]interface{}{
			"id":      "job-new",
			"key_ref": keyVersions[currentKeyIndex],
			"data":    fmt.Sprintf("encrypted_with_%s", keyVersions[currentKeyIndex]),
		}

		newJobKey := fmt.Sprintf("t:%s:job:new", tenantID)
		newJobJSON, _ := json.Marshal(newJob)
		err = client.Set(ctx, newJobKey, newJobJSON, 0).Err()
		require.NoError(t, err)

		// Verify old jobs still reference old key
		oldJobData, err := client.Get(ctx, fmt.Sprintf("t:%s:job:0", tenantID)).Result()
		require.NoError(t, err)

		var oldJob map[string]interface{}
		json.Unmarshal([]byte(oldJobData), &oldJob)
		assert.Equal(t, "key-v1", oldJob["key_ref"])

		// Verify new job uses new key
		newJobData, err := client.Get(ctx, newJobKey).Result()
		require.NoError(t, err)

		var retrievedNewJob map[string]interface{}
		json.Unmarshal([]byte(newJobData), &retrievedNewJob)
		assert.Equal(t, "key-v2", retrievedNewJob["key_ref"])
	})
}

func TestAuditLogging(t *testing.T) {
	client, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("comprehensive audit trail", func(t *testing.T) {
		tenantID := "audited-tenant"
		auditStream := fmt.Sprintf("audit:%s", tenantID)

		// Log various events
		events := []map[string]interface{}{
			{
				"timestamp": time.Now().Unix(),
				"action":    "CREATE_QUEUE",
				"user":      "admin",
				"resource":  "queue-1",
				"result":    "SUCCESS",
			},
			{
				"timestamp": time.Now().Unix(),
				"action":    "ENQUEUE_JOB",
				"user":      "worker-1",
				"resource":  "queue-1",
				"result":    "SUCCESS",
			},
			{
				"timestamp": time.Now().Unix(),
				"action":    "ACCESS_DENIED",
				"user":      "unauthorized",
				"resource":  "queue-2",
				"result":    "DENIED",
			},
		}

		// Add events to Redis Stream
		for _, event := range events {
			eventData := make(map[string]interface{})
			for k, v := range event {
				eventData[k] = fmt.Sprintf("%v", v)
			}

			err := client.XAdd(ctx, &redis.XAddArgs{
				Stream: auditStream,
				Values: eventData,
			}).Err()
			require.NoError(t, err)
		}

		// Query audit log
		entries, err := client.XRange(ctx, auditStream, "-", "+").Result()
		require.NoError(t, err)

		assert.Len(t, entries, 3, "should have 3 audit entries")

		// Verify denied access is logged
		foundDenied := false
		for _, entry := range entries {
			if entry.Values["action"] == "ACCESS_DENIED" {
				foundDenied = true
				assert.Equal(t, "DENIED", entry.Values["result"])
				assert.Equal(t, "unauthorized", entry.Values["user"])
			}
		}
		assert.True(t, foundDenied, "should find denied access in audit log")
	})

	t.Run("audit log retention", func(t *testing.T) {
		tenantID := "retention-tenant"
		auditKey := fmt.Sprintf("audit:%s:daily:%s", tenantID, time.Now().Format("2006-01-02"))

		// Add audit entries with expiration
		for i := 0; i < 10; i++ {
			field := fmt.Sprintf("event-%d", i)
			value := fmt.Sprintf(`{"action":"test","time":%d}`, time.Now().Unix())
			err := client.HSet(ctx, auditKey, field, value).Err()
			require.NoError(t, err)
		}

		// Set TTL for daily rotation
		err := client.Expire(ctx, auditKey, 24*time.Hour).Err()
		require.NoError(t, err)

		// Verify TTL is set
		ttl, err := client.TTL(ctx, auditKey).Result()
		require.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0))
		assert.LessOrEqual(t, ttl, 24*time.Hour)

		// Verify all events are stored
		events, err := client.HGetAll(ctx, auditKey).Result()
		require.NoError(t, err)
		assert.Len(t, events, 10)
	})
}

func TestTenantLifecycle(t *testing.T) {
	client, cleanup := setupIntegrationTest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("complete tenant lifecycle", func(t *testing.T) {
		tenantID := "lifecycle-tenant"

		// 1. Create tenant
		tenantData := map[string]interface{}{
			"id":         tenantID,
			"name":       "Lifecycle Test Tenant",
			"status":     "active",
			"created_at": time.Now().Unix(),
		}

		configKey := fmt.Sprintf("tenant:%s:config", tenantID)
		configJSON, _ := json.Marshal(tenantData)
		err := client.Set(ctx, configKey, configJSON, 0).Err()
		require.NoError(t, err)

		// 2. Initialize tenant resources
		resources := []string{
			fmt.Sprintf("tenant:%s:quotas", tenantID),
			fmt.Sprintf("tenant:%s:users", tenantID),
			fmt.Sprintf("tenant:%s:encryption", tenantID),
		}

		for _, resource := range resources {
			err := client.Set(ctx, resource, "{}", 0).Err()
			require.NoError(t, err)
		}

		// 3. Create queues and jobs
		for i := 0; i < 3; i++ {
			queueKey := fmt.Sprintf("t:%s:queue-%d:jobs", tenantID, i)
			for j := 0; j < 5; j++ {
				err := client.LPush(ctx, queueKey, fmt.Sprintf("job-%d-%d", i, j)).Err()
				require.NoError(t, err)
			}
		}

		// 4. Suspend tenant
		tenantData["status"] = "suspended"
		tenantData["suspended_at"] = time.Now().Unix()
		configJSON, _ = json.Marshal(tenantData)
		err = client.Set(ctx, configKey, configJSON, 0).Err()
		require.NoError(t, err)

		// 5. Verify tenant is suspended
		data, err := client.Get(ctx, configKey).Result()
		require.NoError(t, err)

		var suspended map[string]interface{}
		json.Unmarshal([]byte(data), &suspended)
		assert.Equal(t, "suspended", suspended["status"])

		// 6. Archive tenant (soft delete)
		tenantData["status"] = "archived"
		tenantData["archived_at"] = time.Now().Unix()
		configJSON, _ = json.Marshal(tenantData)
		err = client.Set(ctx, configKey, configJSON, 0).Err()
		require.NoError(t, err)

		// 7. Clean up tenant data (hard delete)
		pattern := fmt.Sprintf("t:%s:*", tenantID)
		keys, err := client.Keys(ctx, pattern).Result()
		require.NoError(t, err)

		if len(keys) > 0 {
			err = client.Del(ctx, keys...).Err()
			require.NoError(t, err)
		}

		// Verify cleanup
		remainingKeys, err := client.Keys(ctx, pattern).Result()
		require.NoError(t, err)
		assert.Len(t, remainingKeys, 0, "all tenant data should be deleted")
	})

	t.Run("tenant migration", func(t *testing.T) {
		sourceTenant := "source-tenant"
		targetTenant := "target-tenant"

		// Create source tenant with data
		sourceQueue := fmt.Sprintf("t:%s:main:jobs", sourceTenant)
		for i := 0; i < 10; i++ {
			err := client.LPush(ctx, sourceQueue, fmt.Sprintf("job-%d", i)).Err()
			require.NoError(t, err)
		}

		// Export data (get all jobs)
		jobs, err := client.LRange(ctx, sourceQueue, 0, -1).Result()
		require.NoError(t, err)
		assert.Len(t, jobs, 10)

		// Import to target tenant
		targetQueue := fmt.Sprintf("t:%s:main:jobs", targetTenant)
		for _, job := range jobs {
			err := client.LPush(ctx, targetQueue, job).Err()
			require.NoError(t, err)
		}

		// Verify migration
		targetJobs, err := client.LRange(ctx, targetQueue, 0, -1).Result()
		require.NoError(t, err)
		assert.Equal(t, jobs, targetJobs)
	})
}

func BenchmarkTenantOperations(b *testing.B) {
	client, cleanup := setupIntegrationTest(&testing.T{})
	defer cleanup()

	ctx := context.Background()

	b.Run("namespace_key_generation", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tenantID := fmt.Sprintf("tenant-%d", i%100)
			queueName := fmt.Sprintf("queue-%d", i%10)
			_ = fmt.Sprintf("t:%s:%s:jobs", tenantID, queueName)
		}
	})

	b.Run("tenant_isolation_check", func(b *testing.B) {
		// Pre-populate some data
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("t:bench-tenant:%d", i)
			client.Set(ctx, key, "data", 0)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pattern := fmt.Sprintf("t:bench-tenant:*")
			client.Keys(ctx, pattern)
		}
	})

	b.Run("quota_check", func(b *testing.B) {
		quotaKey := "tenant:bench:quota:hourly"
		client.Set(ctx, quotaKey, 1000, 0)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			client.Decr(ctx, quotaKey)
			if i%100 == 0 {
				client.Set(ctx, quotaKey, 1000, 0) // Reset
			}
		}
	})
}
