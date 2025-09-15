// Copyright 2025 James Ross
package adminapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	exactlyonce "github.com/flyingrobots/go-redis-work-queue/internal/exactly-once-patterns"
	"github.com/flyingrobots/go-redis-work-queue/internal/exactly_once"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestHandler(t *testing.T) (*ExactlyOnceHandler, *redis.Client, func()) {
	// Setup miniredis
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create handler with custom config to disable metrics
	logger := zap.NewNop()
	handler := newExactlyOnceHandlerWithConfig(client, logger, false)

	cleanup := func() {
		client.Close()
		mr.Close()
	}

	return handler, client, cleanup
}

// Helper function to create handler with configurable metrics
func newExactlyOnceHandlerWithConfig(redisClient *redis.Client, logger *zap.Logger, enableMetrics bool) *ExactlyOnceHandler {
	// Import the exactly-once-patterns package
	cfg := exactlyonce.DefaultConfig()
	cfg.Metrics.Enabled = enableMetrics // Disable metrics for tests
	manager := exactlyonce.NewManager(cfg, redisClient, logger)

	idempManager := exactly_once.NewRedisIdempotencyManager(
		redisClient,
		"admin",
		24*time.Hour,
	)

	return &ExactlyOnceHandler{
		manager:      manager,
		idempManager: idempManager,
		redisClient:  redisClient,
		logger:       logger,
	}
}

func TestGetStats(t *testing.T) {
	handler, client, cleanup := setupTestHandler(t)
	defer cleanup()

	ctx := context.Background()

	// Create some test data
	idempManager := exactly_once.NewRedisIdempotencyManager(client, "test", time.Hour)

	// Add some processed keys
	for i := 0; i < 5; i++ {
		key := exactly_once.NewUUIDKeyGenerator("test", "prefix").Generate(nil)
		isDuplicate, err := idempManager.CheckAndReserve(ctx, key, time.Hour)
		require.NoError(t, err)
		assert.False(t, isDuplicate)
	}

	// Create duplicate attempts
	testKey := "duplicate-test-key"
	_, _ = idempManager.CheckAndReserve(ctx, testKey, time.Hour)
	isDuplicate, _ := idempManager.CheckAndReserve(ctx, testKey, time.Hour)
	assert.True(t, isDuplicate)

	// Test the endpoint
	req := httptest.NewRequest("GET", "/api/v1/exactly-once/stats", nil)
	w := httptest.NewRecorder()

	handler.GetStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "deduplication")
	assert.Contains(t, response, "timestamp")

	dedup := response["deduplication"].(map[string]interface{})
	assert.Contains(t, dedup, "processed")
	assert.Contains(t, dedup, "duplicates")
	assert.Contains(t, dedup, "hit_rate")
}

func TestGetDedupStats(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/exactly-once/dedup/stats", nil)
	w := httptest.NewRecorder()

	handler.GetDedupStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var stats exactly_once.DedupStats
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, stats.Processed, int64(0))
	assert.GreaterOrEqual(t, stats.Duplicates, int64(0))
}

func TestGetPendingOutboxEvents(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/exactly-once/outbox/pending?limit=50", nil)
	w := httptest.NewRecorder()

	handler.GetPendingOutboxEvents(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "pending_events")
	assert.Contains(t, response, "count")
	assert.Contains(t, response, "limit")
	assert.Contains(t, response, "timestamp")
}

func TestPublishOutboxEvents(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/v1/exactly-once/outbox/publish", nil)
	w := httptest.NewRecorder()

	handler.PublishOutboxEvents(w, req)

	// Should return error since outbox is disabled by default
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Outbox is disabled", response["error"])
}

func TestCleanupOutboxEvents(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("POST", "/api/v1/exactly-once/outbox/cleanup", nil)
	w := httptest.NewRecorder()

	handler.CleanupOutboxEvents(w, req)

	// Should return error since outbox is disabled by default
	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "error")
	assert.Equal(t, "Outbox is disabled", response["error"])
}

func TestGetConfig(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/exactly-once/config", nil)
	w := httptest.NewRecorder()

	handler.GetConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "idempotency")
	assert.Contains(t, response, "outbox")
	assert.Contains(t, response, "metrics")

	idempotency := response["idempotency"].(map[string]interface{})
	assert.Equal(t, true, idempotency["enabled"])
	assert.Equal(t, "redis", idempotency["storage_type"])
}

func TestUpdateConfig(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	configUpdate := map[string]interface{}{
		"idempotency": map[string]interface{}{
			"enabled": false,
		},
	}

	body, _ := json.Marshal(configUpdate)
	req := httptest.NewRequest("PUT", "/api/v1/exactly-once/config", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.UpdateConfig(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "success", response["status"])
	assert.Contains(t, response, "message")
}

func TestHealthCheck(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/api/v1/exactly-once/health", nil)
	w := httptest.NewRecorder()

	handler.HealthCheck(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "healthy", response["status"])
	assert.Contains(t, response, "components")

	components := response["components"].(map[string]interface{})
	assert.Contains(t, components, "redis")
	assert.Contains(t, components, "deduplication")
	assert.Contains(t, components, "outbox")

	redis := components["redis"].(map[string]interface{})
	assert.Equal(t, true, redis["healthy"])
}

func TestRegisterRoutes(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	router := mux.NewRouter()
	handler.RegisterRoutes(router)

	// Test that routes are registered
	routes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/exactly-once/stats"},
		{"GET", "/api/v1/exactly-once/dedup/stats"},
		{"GET", "/api/v1/exactly-once/outbox/pending"},
		{"POST", "/api/v1/exactly-once/outbox/publish"},
		{"POST", "/api/v1/exactly-once/outbox/cleanup"},
		{"GET", "/api/v1/exactly-once/config"},
		{"PUT", "/api/v1/exactly-once/config"},
		{"GET", "/api/v1/exactly-once/health"},
	}

	for _, route := range routes {
		req := httptest.NewRequest(route.method, route.path, nil)
		match := &mux.RouteMatch{}
		matched := router.Match(req, match)
		assert.True(t, matched, "Route %s %s should be registered", route.method, route.path)
	}
}

// BenchmarkGetStats benchmarks the GetStats endpoint
func BenchmarkGetStats(b *testing.B) {
	handler, _, cleanup := setupTestHandler(&testing.T{})
	defer cleanup()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/exactly-once/stats", nil)
		w := httptest.NewRecorder()
		handler.GetStats(w, req)
	}
}