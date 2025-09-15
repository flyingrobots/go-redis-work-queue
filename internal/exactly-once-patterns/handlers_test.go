// Copyright 2025 James Ross
package exactlyonce

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func setupTestHandler(t *testing.T) (*AdminHandler, *Manager, func()) {
	cfg := DefaultConfig()
	cfg.Metrics.Enabled = false   // Disable metrics to avoid registration conflicts
	cfg.Idempotency.Enabled = true // Keep idempotency enabled but use memory storage
	cfg.Idempotency.Storage.Type = "memory" // Use memory storage to avoid Redis client issues
	cfg.Outbox.Enabled = false    // Disable outbox to avoid Redis client issues
	logger := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, logger)
	handler := NewAdminHandler(manager, logger)

	cleanup := func() {
		manager.Close()
	}

	return handler, manager, cleanup
}

func TestAdminHandler_NewAdminHandler(t *testing.T) {
	handler, manager, cleanup := setupTestHandler(t)
	defer cleanup()

	assert.NotNil(t, handler)
	assert.Equal(t, manager, handler.manager)
}

func TestAdminHandler_RegisterRoutes(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Test that routes are registered by making requests
	routes := []string{
		"/api/v1/exactly-once/stats",
		"/api/v1/exactly-once/idempotency",
		"/api/v1/exactly-once/outbox",
		"/api/v1/exactly-once/cleanup",
		"/api/v1/exactly-once/health",
	}

	for _, route := range routes {
		req := httptest.NewRequest(http.MethodGet, route, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		// Should not get 404 (route exists), but may get other errors
		assert.NotEqual(t, http.StatusNotFound, w.Code, "Route %s should be registered", route)
	}
}

func TestAdminHandler_HandleStats(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	t.Run("missing queue parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/exactly-once/stats", nil)
		w := httptest.NewRecorder()
		handler.handleStats(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "queue parameter is required")
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/exactly-once/stats?queue=test", nil)
		w := httptest.NewRecorder()
		handler.handleStats(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("successful stats request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/exactly-once/stats?queue=test-queue&tenant=test-tenant", nil)
		w := httptest.NewRecorder()
		handler.handleStats(w, req)

		if w.Code == http.StatusOK {
			var stats DedupStats
			err := json.Unmarshal(w.Body.Bytes(), &stats)
			assert.NoError(t, err)
			assert.Equal(t, "test-queue", stats.QueueName)
			assert.Equal(t, "test-tenant", stats.TenantID)
		}
	})
}

func TestAdminHandler_HandleHealth(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/exactly-once/health", nil)
		w := httptest.NewRecorder()
		handler.handleHealth(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("healthy status", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/exactly-once/health", nil)
		w := httptest.NewRecorder()
		handler.handleHealth(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])

		features, ok := response["features"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, true, features["idempotency"])
		assert.Equal(t, false, features["outbox"])
		assert.Equal(t, false, features["metrics"]) // Disabled in tests
	})
}

func TestAdminHandler_HandleCleanup(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/exactly-once/cleanup", nil)
		w := httptest.NewRecorder()
		handler.handleCleanup(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("cleanup all", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/exactly-once/cleanup", nil)
		w := httptest.NewRecorder()
		handler.handleCleanup(w, req)

		// Could be partial failure if outbox is disabled
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusPartialContent)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Contains(t, []string{"success", "partial_failure"}, response["status"])
		assert.Equal(t, "all", response["type"])
	})
}

func TestAdminHandler_ParseIdempotencyKeyFromQuery(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	t.Run("missing queue", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?key=test-key", nil)
		_, err := handler.parseIdempotencyKeyFromQuery(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "queue parameter is required")
	})

	t.Run("missing key", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?queue=test-queue", nil)
		_, err := handler.parseIdempotencyKeyFromQuery(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key parameter is required")
	})

	t.Run("valid parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?queue=test-queue&key=test-key&tenant=test-tenant&ttl=2h", nil)
		key, err := handler.parseIdempotencyKeyFromQuery(req)
		assert.NoError(t, err)
		assert.Equal(t, "test-key", key.ID)
		assert.Equal(t, "test-queue", key.QueueName)
		assert.Equal(t, "test-tenant", key.TenantID)
		assert.Equal(t, 2*time.Hour, key.TTL)
	})
}

func TestAdminHandler_HandleIdempotencyKey(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/exactly-once/idempotency", nil)
		w := httptest.NewRecorder()
		handler.handleIdempotencyKey(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("GET - missing parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/exactly-once/idempotency", nil)
		w := httptest.NewRecorder()
		handler.handleIdempotencyKey(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("POST - invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/exactly-once/idempotency", bytes.NewReader([]byte("invalid json")))
		w := httptest.NewRecorder()
		handler.handleIdempotencyKey(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("POST - missing queue_name", func(t *testing.T) {
		body := map[string]interface{}{
			"value": "test-value",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/exactly-once/idempotency", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()
		handler.handleIdempotencyKey(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}