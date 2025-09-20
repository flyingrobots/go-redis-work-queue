//go:build multi_tenant_tests
// +build multi_tenant_tests

// Copyright 2025 James Ross
package multitenantiso

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupIntegrationTest(t *testing.T) (*redis.Client, *TenantHandler, *mux.Router) {
	redisClient := setupTestRedis(t)
	config := DefaultConfig()
	handler := NewTenantHandler(redisClient, config)

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	return redisClient, handler, router
}

func TestTenantLifecycle_Integration(t *testing.T) {
	redisClient, _, router := setupIntegrationTest(t)
	defer redisClient.Close()

	// Test 1: Create tenant
	createReq := map[string]interface{}{
		"id":            "integration-test",
		"name":          "Integration Test Tenant",
		"contact_email": "test@example.com",
		"metadata": map[string]string{
			"environment": "test",
		},
	}

	reqBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/tenants", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "test-user")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var createResp TenantConfig
	err := json.Unmarshal(w.Body.Bytes(), &createResp)
	require.NoError(t, err)
	assert.Equal(t, "integration-test", string(createResp.ID))
	assert.Equal(t, "Integration Test Tenant", createResp.Name)
	assert.Equal(t, TenantStatusActive, createResp.Status)

	// Test 2: Get tenant
	req = httptest.NewRequest("GET", "/tenants/integration-test", nil)
	req.Header.Set("X-User-ID", "test-user")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var getResp TenantConfig
	err = json.Unmarshal(w.Body.Bytes(), &getResp)
	require.NoError(t, err)
	assert.Equal(t, createResp.ID, getResp.ID)

	// Test 3: Update tenant
	updateReq := createResp
	updateReq.Name = "Updated Integration Test"
	updateReq.Status = TenantStatusSuspended

	reqBody, _ = json.Marshal(updateReq)
	req = httptest.NewRequest("PUT", "/tenants/integration-test", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "test-user")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updateResp TenantConfig
	err = json.Unmarshal(w.Body.Bytes(), &updateResp)
	require.NoError(t, err)
	assert.Equal(t, "Updated Integration Test", updateResp.Name)
	assert.Equal(t, TenantStatusSuspended, updateResp.Status)

	// Test 4: List tenants
	req = httptest.NewRequest("GET", "/tenants", nil)
	req.Header.Set("X-User-ID", "test-user")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var listResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	require.NoError(t, err)

	tenants := listResp["tenants"].([]interface{})
	assert.Len(t, tenants, 1)

	// Test 5: Get quota usage
	req = httptest.NewRequest("GET", "/tenants/integration-test/quota-usage", nil)
	req.Header.Set("X-User-ID", "test-user")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var usageResp QuotaUsage
	err = json.Unmarshal(w.Body.Bytes(), &usageResp)
	require.NoError(t, err)
	assert.Equal(t, TenantID("integration-test"), usageResp.TenantID)

	// Test 6: Check quota
	quotaReq := map[string]interface{}{
		"quota_type": "jobs_per_hour",
		"amount":     50,
	}

	reqBody, _ = json.Marshal(quotaReq)
	req = httptest.NewRequest("POST", "/tenants/integration-test/check-quota", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "test-user")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var quotaResp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &quotaResp)
	require.NoError(t, err)
	assert.True(t, quotaResp["allowed"].(bool))

	// Test 7: Delete tenant
	req = httptest.NewRequest("DELETE", "/tenants/integration-test", nil)
	req.Header.Set("X-User-ID", "test-user")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Test 8: Verify deletion
	req = httptest.NewRequest("GET", "/tenants/integration-test", nil)
	req.Header.Set("X-User-ID", "test-user")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code) // Should still exist but marked as deleted

	var deletedResp TenantConfig
	err = json.Unmarshal(w.Body.Bytes(), &deletedResp)
	require.NoError(t, err)
	assert.Equal(t, TenantStatusDeleted, deletedResp.Status)
}

func TestErrorHandling_Integration(t *testing.T) {
	_, _, router := setupIntegrationTest(t)

	// Test 1: Create tenant with invalid data
	invalidReq := map[string]interface{}{
		"id":   "a", // Too short
		"name": "",  // Empty name
	}

	reqBody, _ := json.Marshal(invalidReq)
	req := httptest.NewRequest("POST", "/tenants", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "test-user")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test 2: Get non-existent tenant
	req = httptest.NewRequest("GET", "/tenants/non-existent", nil)
	req.Header.Set("X-User-ID", "test-user")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	// Test 3: Invalid JSON
	req = httptest.NewRequest("POST", "/tenants", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "test-user")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestQuotaEnforcement_Integration(t *testing.T) {
	redisClient, handler, router := setupIntegrationTest(t)
	defer redisClient.Close()

	// Create a tenant with low quotas for testing
	config := &TenantConfig{
		ID:   "quota-test",
		Name: "Quota Test Tenant",
		Quotas: TenantQuotas{
			MaxJobsPerHour: 5,
			MaxJobsPerDay:  10,
			MaxBacklogSize: 3,
		},
	}

	err := handler.tenantManager.CreateTenant(config)
	require.NoError(t, err)

	// Test quota within limits
	quotaReq := map[string]interface{}{
		"quota_type": "jobs_per_hour",
		"amount":     3,
	}

	reqBody, _ := json.Marshal(quotaReq)
	req := httptest.NewRequest("POST", "/tenants/quota-test/check-quota", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "test-user")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.True(t, resp["allowed"].(bool))

	// Increment usage
	err = handler.tenantManager.IncrementQuotaUsage("quota-test", "jobs_per_hour", 4)
	require.NoError(t, err)

	// Test quota exceeding limits
	quotaReq["amount"] = 3
	reqBody, _ = json.Marshal(quotaReq)
	req = httptest.NewRequest("POST", "/tenants/quota-test/check-quota", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "test-user")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.False(t, resp["allowed"].(bool))
	assert.Contains(t, resp["reason"].(string), "quota exceeded")
}

func TestEncryption_Integration(t *testing.T) {
	redisClient, handler, _ := setupIntegrationTest(t)
	defer redisClient.Close()

	// Create tenant with encryption enabled
	config := &TenantConfig{
		ID:     "encryption-test",
		Name:   "Encryption Test Tenant",
		Quotas: DefaultQuotas(),
		Encryption: TenantEncryption{
			Enabled:     true,
			KEKProvider: "local",
			KEKKeyID:    "test-key-123",
			Algorithm:   "AES-256-GCM",
		},
	}

	err := handler.tenantManager.CreateTenant(config)
	require.NoError(t, err)

	// Test payload encryption
	payload := []byte("sensitive tenant data that needs encryption")

	encrypted, err := handler.tenantManager.encryptor.EncryptPayload(payload, config)
	require.NoError(t, err)
	assert.NotNil(t, encrypted)
	assert.NotEqual(t, payload, encrypted.EncryptedPayload)

	// Test payload decryption
	decrypted, err := handler.tenantManager.encryptor.DecryptPayload(encrypted, config)
	require.NoError(t, err)
	assert.Equal(t, payload, decrypted)

	// Test encryption with disabled tenant
	config.Encryption.Enabled = false
	_, err = handler.tenantManager.encryptor.EncryptPayload(payload, config)
	assert.Error(t, err)
	assert.Equal(t, ErrEncryptionNotEnabled, err)
}

func TestRateLimiting_Integration(t *testing.T) {
	redisClient, handler, _ := setupIntegrationTest(t)
	defer redisClient.Close()

	// Create tenant with rate limiting
	config := &TenantConfig{
		ID:   "ratelimit-test",
		Name: "Rate Limit Test Tenant",
		Quotas: TenantQuotas{
			EnqueueRateLimit: 2, // 2 per second
		},
		RateLimiting: TenantRateLimiting{
			Enabled:        true,
			WindowDuration: time.Second,
		},
	}

	err := handler.tenantManager.CreateTenant(config)
	require.NoError(t, err)

	rateLimitChecker := NewRateLimitChecker(redisClient)

	// Test rate limiting within limits
	err = rateLimitChecker.CheckRateLimit("ratelimit-test", "enqueue", config)
	assert.NoError(t, err)

	err = rateLimitChecker.CheckRateLimit("ratelimit-test", "enqueue", config)
	assert.NoError(t, err)

	// Test rate limiting exceeding limits
	err = rateLimitChecker.CheckRateLimit("ratelimit-test", "enqueue", config)
	assert.Error(t, err)
	assert.True(t, IsRateLimited(err))

	// Test that rate limiting resets after window
	time.Sleep(time.Second + 100*time.Millisecond)

	err = rateLimitChecker.CheckRateLimit("ratelimit-test", "enqueue", config)
	assert.NoError(t, err)
}

func TestAuditTrail_Integration(t *testing.T) {
	redisClient, handler, router := setupIntegrationTest(t)
	defer redisClient.Close()

	// Create tenant (this should generate audit events)
	createReq := map[string]interface{}{
		"id":   "audit-test",
		"name": "Audit Test Tenant",
	}

	reqBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/tenants", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "audit-user")
	req.Header.Set("X-Forwarded-For", "192.168.1.100")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify audit event was logged
	// In a real implementation, you would have an API to query audit events
	// For this test, we'll verify that the Redis keys exist
	auditPattern := "audit:audit-test:*"
	keys, err := redisClient.Keys(context.Background(), auditPattern).Result()
	require.NoError(t, err)
	assert.NotEmpty(t, keys, "Audit events should be logged")

	// Test manual audit logging
	event := &AuditEvent{
		TenantID:  "audit-test",
		UserID:    "test-user",
		Action:    "MANUAL_TEST",
		Resource:  "test-resource",
		Details:   map[string]interface{}{"test": "data"},
		RemoteIP:  "127.0.0.1",
		UserAgent: "test-agent",
		Result:    "SUCCESS",
	}

	err = handler.tenantManager.LogAuditEvent(event)
	assert.NoError(t, err)
}

func TestTenantIsolation_Integration(t *testing.T) {
	redisClient, handler, _ := setupIntegrationTest(t)
	defer redisClient.Close()

	// Create two tenants
	tenant1 := &TenantConfig{
		ID:     "tenant-1",
		Name:   "Tenant 1",
		Quotas: DefaultQuotas(),
	}

	tenant2 := &TenantConfig{
		ID:     "tenant-2",
		Name:   "Tenant 2",
		Quotas: DefaultQuotas(),
	}

	err := handler.tenantManager.CreateTenant(tenant1)
	require.NoError(t, err)

	err = handler.tenantManager.CreateTenant(tenant2)
	require.NoError(t, err)

	// Test key namespace isolation
	ns1 := handler.tenantManager.GetTenantNamespace("tenant-1")
	ns2 := handler.tenantManager.GetTenantNamespace("tenant-2")

	// Verify different key spaces
	assert.NotEqual(t, ns1.QueueKey("queue1"), ns2.QueueKey("queue1"))
	assert.Equal(t, "t:tenant-1:queue1", ns1.QueueKey("queue1"))
	assert.Equal(t, "t:tenant-2:queue1", ns2.QueueKey("queue1"))

	// Simulate data in both tenant namespaces
	err = redisClient.Set(context.Background(), ns1.QueueKey("queue1"), "data1", 0).Err()
	require.NoError(t, err)

	err = redisClient.Set(context.Background(), ns2.QueueKey("queue1"), "data2", 0).Err()
	require.NoError(t, err)

	// Verify isolation
	val1, err := redisClient.Get(context.Background(), ns1.QueueKey("queue1")).Result()
	require.NoError(t, err)
	assert.Equal(t, "data1", val1)

	val2, err := redisClient.Get(context.Background(), ns2.QueueKey("queue1")).Result()
	require.NoError(t, err)
	assert.Equal(t, "data2", val2)

	// Test quota isolation
	err = handler.tenantManager.IncrementQuotaUsage("tenant-1", "jobs_per_hour", 100)
	require.NoError(t, err)

	usage1, err := handler.tenantManager.GetQuotaUsage("tenant-1")
	require.NoError(t, err)
	assert.Equal(t, int64(100), usage1.JobsThisHour)

	usage2, err := handler.tenantManager.GetQuotaUsage("tenant-2")
	require.NoError(t, err)
	assert.Equal(t, int64(0), usage2.JobsThisHour)

	// Test tenant deletion isolation
	err = handler.tenantManager.DeleteTenant("tenant-1")
	require.NoError(t, err)

	// Verify tenant-1 data is cleaned up but tenant-2 data remains
	_, err = redisClient.Get(context.Background(), ns1.QueueKey("queue1")).Result()
	assert.Equal(t, redis.Nil, err) // Should be deleted

	val2, err = redisClient.Get(context.Background(), ns2.QueueKey("queue1")).Result()
	require.NoError(t, err)
	assert.Equal(t, "data2", val2) // Should still exist
}

func TestMiddleware_Integration(t *testing.T) {
	redisClient, handler, _ := setupIntegrationTest(t)
	defer redisClient.Close()

	// Create test tenant
	config := &TenantConfig{
		ID:     "middleware-test",
		Name:   "Middleware Test",
		Status: TenantStatusActive,
		Quotas: DefaultQuotas(),
	}

	err := handler.tenantManager.CreateTenant(config)
	require.NoError(t, err)

	// Create router with middleware
	router := mux.NewRouter()
	router.Use(handler.MiddlewareFunc())

	// Add a test endpoint
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Header.Get("X-Tenant-ID")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("tenant: " + tenantID))
	})

	// Test with valid tenant
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "middleware-test")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "middleware-test")

	// Test with suspended tenant
	config.Status = TenantStatusSuspended
	err = handler.tenantManager.UpdateTenant(config)
	require.NoError(t, err)

	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "middleware-test")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	// Test with non-existent tenant
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", "non-existent")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
