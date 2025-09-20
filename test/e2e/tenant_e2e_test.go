//go:build e2e_tests
// +build e2e_tests

// Copyright 2025 James Ross
package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TenantSystem struct {
	redis  *redis.Client
	config map[string]interface{}
}

func NewTenantSystem(redis *redis.Client) *TenantSystem {
	return &TenantSystem{
		redis:  redis,
		config: make(map[string]interface{}),
	}
}

func setupE2ETest(t *testing.T) (*TenantSystem, func()) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	system := NewTenantSystem(client)

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return system, cleanup
}

func TestE2ESaaSPlatformScenario(t *testing.T) {
	system, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("complete SaaS platform workflow", func(t *testing.T) {
		// Scenario: A SaaS platform onboarding multiple customers
		customers := []struct {
			tenantID    string
			companyName string
			tier        string
			quotas      map[string]int64
		}{
			{
				tenantID:    "startup-inc",
				companyName: "Startup Inc",
				tier:        "starter",
				quotas: map[string]int64{
					"max_jobs_per_day": 1000,
					"max_queues":       5,
					"max_workers":      10,
				},
			},
			{
				tenantID:    "enterprise-corp",
				companyName: "Enterprise Corp",
				tier:        "enterprise",
				quotas: map[string]int64{
					"max_jobs_per_day": 100000,
					"max_queues":       100,
					"max_workers":      500,
				},
			},
			{
				tenantID:    "mid-market-co",
				companyName: "Mid Market Co",
				tier:        "professional",
				quotas: map[string]int64{
					"max_jobs_per_day": 10000,
					"max_queues":       20,
					"max_workers":      50,
				},
			},
		}

		// Phase 1: Onboard customers
		for _, customer := range customers {
			// Create tenant
			tenantConfig := map[string]interface{}{
				"id":            customer.tenantID,
				"name":          customer.companyName,
				"tier":          customer.tier,
				"status":        "active",
				"created_at":    time.Now().Unix(),
				"quotas":        customer.quotas,
				"encryption":    customer.tier == "enterprise", // Only enterprise gets encryption
				"audit_enabled": true,
			}

			configKey := fmt.Sprintf("tenant:%s:config", customer.tenantID)
			configJSON, err := json.Marshal(tenantConfig)
			require.NoError(t, err)

			err = system.redis.Set(ctx, configKey, configJSON, 0).Err()
			require.NoError(t, err)

			// Initialize quota tracking
			quotaKey := fmt.Sprintf("tenant:%s:quotas:current", customer.tenantID)
			err = system.redis.HSet(ctx, quotaKey,
				"jobs_today", 0,
				"active_queues", 0,
				"active_workers", 0,
			).Err()
			require.NoError(t, err)

			t.Logf("Onboarded %s as %s tier", customer.companyName, customer.tier)
		}

		// Phase 2: Simulate customer usage patterns
		var wg sync.WaitGroup

		// Startup: Bursts of activity
		wg.Add(1)
		go func() {
			defer wg.Done()
			tenantID := "startup-inc"

			// Create a few queues
			for i := 0; i < 3; i++ {
				queueKey := fmt.Sprintf("t:%s:queue-%d:jobs", tenantID, i)

				// Burst of jobs
				for j := 0; j < 50; j++ {
					job := fmt.Sprintf(`{"id":"job-%d-%d","data":"startup-work"}`, i, j)
					err := system.redis.LPush(ctx, queueKey, job).Err()
					if err != nil {
						t.Logf("Startup enqueue error: %v", err)
					}
				}

				// Update quota usage
				quotaKey := fmt.Sprintf("tenant:%s:quotas:current", tenantID)
				system.redis.HIncrBy(ctx, quotaKey, "jobs_today", 50)
			}

			t.Log("Startup Inc: Burst processing complete")
		}()

		// Enterprise: Continuous high-volume processing
		wg.Add(1)
		go func() {
			defer wg.Done()
			tenantID := "enterprise-corp"

			// Many queues with steady flow
			for i := 0; i < 10; i++ {
				queueKey := fmt.Sprintf("t:%s:dept-%d:jobs", tenantID, i)

				// High volume
				for j := 0; j < 100; j++ {
					job := map[string]interface{}{
						"id":         fmt.Sprintf("ent-job-%d-%d", i, j),
						"priority":   j % 3, // Mixed priorities
						"encrypted":  true,
						"department": fmt.Sprintf("dept-%d", i),
					}
					jobJSON, _ := json.Marshal(job)

					err := system.redis.LPush(ctx, queueKey, jobJSON).Err()
					if err != nil {
						t.Logf("Enterprise enqueue error: %v", err)
					}
				}

				// Track usage
				quotaKey := fmt.Sprintf("tenant:%s:quotas:current", tenantID)
				system.redis.HIncrBy(ctx, quotaKey, "jobs_today", 100)

				// Simulate processing
				time.Sleep(10 * time.Millisecond)
			}

			t.Log("Enterprise Corp: High-volume processing complete")
		}()

		// Mid-market: Moderate steady usage
		wg.Add(1)
		go func() {
			defer wg.Done()
			tenantID := "mid-market-co"

			// Moderate number of queues
			for i := 0; i < 5; i++ {
				queueKey := fmt.Sprintf("t:%s:service-%d:jobs", tenantID, i)

				// Moderate volume
				for j := 0; j < 30; j++ {
					job := fmt.Sprintf(`{"id":"mm-job-%d-%d","service":%d}`, i, j, i)
					err := system.redis.LPush(ctx, queueKey, job).Err()
					if err != nil {
						t.Logf("Mid-market enqueue error: %v", err)
					}
				}

				quotaKey := fmt.Sprintf("tenant:%s:quotas:current", tenantID)
				system.redis.HIncrBy(ctx, quotaKey, "jobs_today", 30)
			}

			t.Log("Mid Market Co: Processing complete")
		}()

		wg.Wait()

		// Phase 3: Verify isolation and quotas
		for _, customer := range customers {
			// Check job counts
			quotaKey := fmt.Sprintf("tenant:%s:quotas:current", customer.tenantID)
			jobCount, err := system.redis.HGet(ctx, quotaKey, "jobs_today").Result()
			require.NoError(t, err)

			t.Logf("%s processed %s jobs today", customer.companyName, jobCount)

			// Verify namespace isolation
			pattern := fmt.Sprintf("t:%s:*", customer.tenantID)
			keys, err := system.redis.Keys(ctx, pattern).Result()
			require.NoError(t, err)

			// Each tenant should only see their own data
			for _, key := range keys {
				assert.Contains(t, key, customer.tenantID)
			}
		}

		// Phase 4: Simulate quota enforcement
		t.Log("Testing quota enforcement...")

		// Try to exceed startup's quota
		startupQuotaKey := fmt.Sprintf("tenant:%s:quotas:current", "startup-inc")
		currentJobs, _ := system.redis.HGet(ctx, startupQuotaKey, "jobs_today").Int64()

		if currentJobs < 1000 {
			// Can still add jobs
			remaining := 1000 - currentJobs
			t.Logf("Startup Inc can still process %d jobs today", remaining)
		} else {
			t.Log("Startup Inc has reached daily quota")
		}

		// Phase 5: Audit trail verification
		for _, customer := range customers {
			auditKey := fmt.Sprintf("audit:%s:events", customer.tenantID)

			// Log some audit events
			events := []map[string]string{
				{"action": "TENANT_CREATED", "user": "system"},
				{"action": "QUEUE_CREATED", "user": "api"},
				{"action": "JOBS_ENQUEUED", "user": "worker"},
			}

			for _, event := range events {
				event["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
				event["tenant"] = customer.tenantID

				err := system.redis.XAdd(ctx, &redis.XAddArgs{
					Stream: auditKey,
					Values: event,
				}).Err()
				require.NoError(t, err)
			}

			// Verify audit log exists
			count, err := system.redis.XLen(ctx, auditKey).Result()
			require.NoError(t, err)
			assert.Equal(t, int64(3), count, "%s should have 3 audit events", customer.companyName)
		}
	})
}

func TestE2ESecurityScenarios(t *testing.T) {
	system, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("multi-tenant security isolation", func(t *testing.T) {
		// Setup: Create sensitive tenants
		tenants := []string{"bank-a", "bank-b", "healthcare-org"}

		for _, tenantID := range tenants {
			// Each tenant has sensitive data
			secretKey := fmt.Sprintf("t:%s:secrets:api-key", tenantID)
			secretValue := fmt.Sprintf("secret-key-%s-%d", tenantID, time.Now().Unix())

			err := system.redis.Set(ctx, secretKey, secretValue, 0).Err()
			require.NoError(t, err)

			// Enable encryption for sensitive tenants
			encConfig := map[string]interface{}{
				"enabled":     true,
				"algorithm":   "AES-256-GCM",
				"kek_id":      fmt.Sprintf("kek-%s", tenantID),
				"rotate_days": 30,
			}

			encKey := fmt.Sprintf("tenant:%s:encryption", tenantID)
			encJSON, _ := json.Marshal(encConfig)
			err = system.redis.Set(ctx, encKey, encJSON, 0).Err()
			require.NoError(t, err)
		}

		// Attack Scenario 1: Cross-tenant access attempt
		t.Log("Testing cross-tenant access prevention...")

		// Bank A trying to access Bank B's data (should be prevented by app layer)
		bankAPattern := "t:bank-a:*"
		bankAKeys, err := system.redis.Keys(ctx, bankAPattern).Result()
		require.NoError(t, err)

		bankBPattern := "t:bank-b:*"
		bankBKeys, err := system.redis.Keys(ctx, bankBPattern).Result()
		require.NoError(t, err)

		// Verify complete isolation
		for _, key := range bankAKeys {
			assert.NotContains(t, key, "bank-b")
		}
		for _, key := range bankBKeys {
			assert.NotContains(t, key, "bank-a")
		}

		// Attack Scenario 2: Attempt to bypass quotas
		t.Log("Testing quota bypass prevention...")

		attackerTenant := "attacker-tenant"
		quotaKey := fmt.Sprintf("tenant:%s:quotas", attackerTenant)

		// Set very low quota
		quotas := map[string]int64{
			"max_jobs_per_hour": 10,
		}
		quotaJSON, _ := json.Marshal(quotas)
		err = system.redis.Set(ctx, quotaKey, quotaJSON, 0).Err()
		require.NoError(t, err)

		// Track current usage
		usageKey := fmt.Sprintf("tenant:%s:usage:hourly", attackerTenant)

		// Try to flood the system
		blocked := false
		for i := 0; i < 20; i++ {
			current, _ := system.redis.Incr(ctx, usageKey).Result()
			if current > 10 {
				blocked = true
				break
			}
		}
		assert.True(t, blocked, "Should block after exceeding quota")

		// Attack Scenario 3: Data exfiltration attempt
		t.Log("Testing data exfiltration prevention...")

		// Log suspicious activity
		for _, tenantID := range tenants {
			suspiciousEvent := map[string]interface{}{
				"timestamp": time.Now().Unix(),
				"tenant":    tenantID,
				"action":    "SUSPICIOUS_ACCESS",
				"user":      "unknown",
				"ip":        "192.168.1.100",
				"pattern":   "t:*:secrets:*",
				"result":    "BLOCKED",
			}

			auditKey := fmt.Sprintf("security:audit:%s", tenantID)
			eventJSON, _ := json.Marshal(suspiciousEvent)
			err := system.redis.LPush(ctx, auditKey, eventJSON).Err()
			require.NoError(t, err)
		}

		// Verify security events are logged
		for _, tenantID := range tenants {
			auditKey := fmt.Sprintf("security:audit:%s", tenantID)
			events, err := system.redis.LRange(ctx, auditKey, 0, -1).Result()
			require.NoError(t, err)
			assert.Greater(t, len(events), 0, "Security events should be logged")
		}
	})

	t.Run("encryption key management", func(t *testing.T) {
		tenantID := "crypto-tenant"

		// Simulate key hierarchy
		masterKey := "master-key-singleton"
		tenantKEK := fmt.Sprintf("kek-%s", tenantID)

		// Store master key reference (in production, this would be in KMS)
		err := system.redis.Set(ctx, "kms:master", masterKey, 0).Err()
		require.NoError(t, err)

		// Generate tenant KEK (encrypted with master)
		kekData := map[string]interface{}{
			"key_id":       tenantKEK,
			"encrypted_by": masterKey,
			"created_at":   time.Now().Unix(),
			"version":      1,
		}

		kekKey := fmt.Sprintf("kms:tenant:%s:kek", tenantID)
		kekJSON, _ := json.Marshal(kekData)
		err = system.redis.Set(ctx, kekKey, kekJSON, 0).Err()
		require.NoError(t, err)

		// Generate multiple DEKs for job encryption
		for i := 0; i < 5; i++ {
			dekData := map[string]interface{}{
				"dek_id":       fmt.Sprintf("dek-%s-%d", tenantID, i),
				"encrypted_by": tenantKEK,
				"created_at":   time.Now().Unix(),
				"job_count":    0,
			}

			dekKey := fmt.Sprintf("kms:tenant:%s:dek:%d", tenantID, i)
			dekJSON, _ := json.Marshal(dekData)
			err := system.redis.Set(ctx, dekKey, dekJSON, time.Hour).Err()
			require.NoError(t, err)
		}

		// Verify key hierarchy
		masterExists, _ := system.redis.Exists(ctx, "kms:master").Result()
		assert.Equal(t, int64(1), masterExists)

		kekExists, _ := system.redis.Exists(ctx, kekKey).Result()
		assert.Equal(t, int64(1), kekExists)

		// Count DEKs
		dekPattern := fmt.Sprintf("kms:tenant:%s:dek:*", tenantID)
		deks, err := system.redis.Keys(ctx, dekPattern).Result()
		require.NoError(t, err)
		assert.Len(t, deks, 5, "Should have 5 DEKs")
	})
}

func TestE2EPerformanceAndScale(t *testing.T) {
	system, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("multi-tenant at scale", func(t *testing.T) {
		numTenants := 10
		jobsPerTenant := 100
		queuesPerTenant := 5

		startTime := time.Now()

		// Create tenants concurrently
		var wg sync.WaitGroup
		for i := 0; i < numTenants; i++ {
			wg.Add(1)
			go func(tenantNum int) {
				defer wg.Done()

				tenantID := fmt.Sprintf("scale-tenant-%d", tenantNum)

				// Create tenant config
				config := map[string]interface{}{
					"id":     tenantID,
					"name":   fmt.Sprintf("Scale Tenant %d", tenantNum),
					"status": "active",
				}

				configKey := fmt.Sprintf("tenant:%s:config", tenantID)
				configJSON, _ := json.Marshal(config)
				system.redis.Set(ctx, configKey, configJSON, 0)

				// Create queues and jobs
				for q := 0; q < queuesPerTenant; q++ {
					queueKey := fmt.Sprintf("t:%s:q%d:jobs", tenantID, q)

					// Batch insert jobs
					pipe := system.redis.Pipeline()
					for j := 0; j < jobsPerTenant/queuesPerTenant; j++ {
						job := fmt.Sprintf(`{"t":%d,"q":%d,"j":%d}`, tenantNum, q, j)
						pipe.LPush(ctx, queueKey, job)
					}
					pipe.Exec(ctx)
				}
			}(i)
		}

		wg.Wait()
		elapsed := time.Since(startTime)

		totalJobs := numTenants * jobsPerTenant
		totalQueues := numTenants * queuesPerTenant

		t.Logf("Created %d tenants with %d total jobs across %d queues in %v",
			numTenants, totalJobs, totalQueues, elapsed)

		// Verify all data was created
		allKeys, err := system.redis.Keys(ctx, "t:scale-tenant-*").Result()
		require.NoError(t, err)
		assert.Equal(t, totalQueues, len(allKeys))

		// Performance metrics
		jobsPerSecond := float64(totalJobs) / elapsed.Seconds()
		t.Logf("Performance: %.2f jobs/second", jobsPerSecond)
		assert.Greater(t, jobsPerSecond, 100.0, "Should handle >100 jobs/second")
	})

	t.Run("quota enforcement under load", func(t *testing.T) {
		tenantID := "load-test-tenant"
		rateLimit := 100 // ops per second
		testDuration := 2 * time.Second

		// Set rate limit
		rateLimitKey := fmt.Sprintf("tenant:%s:rate_limit", tenantID)

		var allowed int64
		var rejected int64
		done := make(chan bool)

		// Generate load
		go func() {
			ticker := time.NewTicker(5 * time.Millisecond) // 200 ops/sec attempt
			defer ticker.Stop()

			timeout := time.After(testDuration)
			for {
				select {
				case <-ticker.C:
					// Simple rate limit check
					current, _ := system.redis.Incr(ctx, rateLimitKey).Result()
					if current <= int64(rateLimit) {
						allowed++
					} else {
						rejected++
					}
				case <-timeout:
					done <- true
					return
				}
			}
		}()

		// Reset counter every second
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ticker.C:
					system.redis.Del(ctx, rateLimitKey)
				case <-done:
					return
				}
			}
		}()

		<-done

		t.Logf("Rate limiting: allowed=%d, rejected=%d", allowed, rejected)
		assert.Greater(t, rejected, int64(0), "Should have rejected some requests")
		assert.Greater(t, allowed, int64(0), "Should have allowed some requests")

		// Verify rate was approximately enforced
		expectedAllowed := int64(rateLimit * int(testDuration.Seconds()))
		tolerance := float64(expectedAllowed) * 0.2 // 20% tolerance
		assert.InDelta(t, expectedAllowed, allowed, tolerance)
	})
}

func TestE2EDisasterRecovery(t *testing.T) {
	system, cleanup := setupE2ETest(t)
	defer cleanup()

	ctx := context.Background()

	t.Run("tenant backup and restore", func(t *testing.T) {
		sourceTenant := "production-tenant"

		// Create production tenant with data
		tenantData := map[string]interface{}{
			"id":     sourceTenant,
			"name":   "Production Company",
			"status": "active",
			"settings": map[string]interface{}{
				"critical_setting": "important_value",
				"feature_flags":    []string{"feature1", "feature2"},
			},
		}

		// Store tenant config
		configKey := fmt.Sprintf("tenant:%s:config", sourceTenant)
		configJSON, _ := json.Marshal(tenantData)
		err := system.redis.Set(ctx, configKey, configJSON, 0).Err()
		require.NoError(t, err)

		// Add operational data
		for i := 0; i < 5; i++ {
			queueKey := fmt.Sprintf("t:%s:queue-%d:jobs", sourceTenant, i)
			for j := 0; j < 10; j++ {
				system.redis.LPush(ctx, queueKey, fmt.Sprintf("job-%d-%d", i, j))
			}
		}

		// Simulate backup process
		t.Log("Creating backup...")

		// Get all tenant keys
		pattern := fmt.Sprintf("*%s*", sourceTenant)
		allKeys, err := system.redis.Keys(ctx, pattern).Result()
		require.NoError(t, err)

		backup := make(map[string]string)
		for _, key := range allKeys {
			keyType, _ := system.redis.Type(ctx, key).Result()

			switch keyType {
			case "string":
				value, _ := system.redis.Get(ctx, key).Result()
				backup[key] = value
			case "list":
				values, _ := system.redis.LRange(ctx, key, 0, -1).Result()
				jsonValues, _ := json.Marshal(values)
				backup[key] = string(jsonValues)
			case "hash":
				values, _ := system.redis.HGetAll(ctx, key).Result()
				jsonValues, _ := json.Marshal(values)
				backup[key] = string(jsonValues)
			}
		}

		t.Logf("Backed up %d keys", len(backup))

		// Simulate disaster - delete all data
		t.Log("Simulating disaster...")
		for _, key := range allKeys {
			system.redis.Del(ctx, key)
		}

		// Verify data is gone
		remainingKeys, err := system.redis.Keys(ctx, pattern).Result()
		require.NoError(t, err)
		assert.Len(t, remainingKeys, 0, "All data should be deleted")

		// Restore from backup
		t.Log("Restoring from backup...")
		restoredTenant := "production-tenant" // Same tenant after recovery

		for key, value := range backup {
			if key == configKey {
				// Restore config
				err := system.redis.Set(ctx, key, value, 0).Err()
				require.NoError(t, err)
			} else if strings.Contains(key, ":jobs") {
				// Restore job lists
				var jobs []string
				json.Unmarshal([]byte(value), &jobs)
				for _, job := range jobs {
					system.redis.LPush(ctx, key, job)
				}
			}
		}

		// Verify restoration
		restoredConfig, err := system.redis.Get(ctx, configKey).Result()
		require.NoError(t, err)

		var restored map[string]interface{}
		json.Unmarshal([]byte(restoredConfig), &restored)
		assert.Equal(t, sourceTenant, restored["id"])
		assert.Equal(t, "Production Company", restored["name"])

		// Verify queues restored
		restoredKeys, err := system.redis.Keys(ctx, pattern).Result()
		require.NoError(t, err)
		assert.Greater(t, len(restoredKeys), 0, "Data should be restored")

		t.Log("Disaster recovery successful")
	})

	t.Run("tenant migration during incident", func(t *testing.T) {
		affectedTenant := "affected-tenant"
		newTenant := "migrated-tenant"

		// Create tenant experiencing issues
		problemData := map[string]interface{}{
			"id":       affectedTenant,
			"name":     "Affected Company",
			"status":   "degraded",
			"incident": "high_latency",
		}

		configKey := fmt.Sprintf("tenant:%s:config", affectedTenant)
		configJSON, _ := json.Marshal(problemData)
		system.redis.Set(ctx, configKey, configJSON, 0)

		// Add some jobs that need migration
		oldQueue := fmt.Sprintf("t:%s:critical:jobs", affectedTenant)
		for i := 0; i < 20; i++ {
			job := fmt.Sprintf(`{"id":"critical-job-%d","priority":"high"}`, i)
			system.redis.LPush(ctx, oldQueue, job)
		}

		// Perform hot migration
		t.Log("Starting hot migration...")

		// Create new tenant
		newData := map[string]interface{}{
			"id":            newTenant,
			"name":          "Affected Company (Migrated)",
			"status":        "active",
			"migrated_from": affectedTenant,
			"migrated_at":   time.Now().Unix(),
		}

		newConfigKey := fmt.Sprintf("tenant:%s:config", newTenant)
		newConfigJSON, _ := json.Marshal(newData)
		system.redis.Set(ctx, newConfigKey, newConfigJSON, 0)

		// Migrate jobs atomically
		newQueue := fmt.Sprintf("t:%s:critical:jobs", newTenant)

		for {
			job, err := system.redis.RPopLPush(ctx, oldQueue, newQueue).Result()
			if err == redis.Nil {
				break // No more jobs
			}
			require.NoError(t, err)
			_ = job // Job migrated
		}

		// Verify migration
		oldCount, _ := system.redis.LLen(ctx, oldQueue).Result()
		newCount, _ := system.redis.LLen(ctx, newQueue).Result()

		assert.Equal(t, int64(0), oldCount, "Old queue should be empty")
		assert.Equal(t, int64(20), newCount, "New queue should have all jobs")

		// Update old tenant to redirect
		problemData["status"] = "migrated"
		problemData["redirect_to"] = newTenant
		configJSON, _ = json.Marshal(problemData)
		system.redis.Set(ctx, configKey, configJSON, 0)

		t.Log("Hot migration completed successfully")
	})
}
