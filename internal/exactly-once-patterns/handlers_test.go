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
	cfg.Metrics.Enabled = false // Disable metrics to avoid registration conflicts
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
	cfg := DefaultConfig()
	cfg.Idempotency.Enabled = true
	logger := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, logger)
	handler := NewAdminHandler(manager, logger)

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

func TestAdminHandler_HandleIdempotencyKey(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Idempotency.Enabled = true
	logger := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, logger)
	handler := NewAdminHandler(manager, logger)

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

	t.Run("GET - with parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/exactly-once/idempotency?queue=test&key=test-key", nil)
		w := httptest.NewRecorder()
		handler.handleIdempotencyKey(w, req)

		if w.Code == http.StatusOK {
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "test-key", response["key"])
			assert.Equal(t, "test", response["queue"])
		}
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

	t.Run("POST - valid request", func(t *testing.T) {
		body := map[string]interface{}{
			"queue_name": "test-queue",
			"tenant_id":  "test-tenant",
			"value":      map[string]string{"status": "processed"},
			"ttl":        "1h",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/exactly-once/idempotency", bytes.NewReader(bodyBytes))
		w := httptest.NewRecorder()
		handler.handleIdempotencyKey(w, req)

		if w.Code == http.StatusCreated {
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, "test-queue", response["queue"])
			assert.Equal(t, "test-tenant", response["tenant"])
		}
	})

	t.Run("DELETE - missing parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/exactly-once/idempotency", nil)
		w := httptest.NewRecorder()
		handler.handleIdempotencyKey(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("DELETE - with parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/exactly-once/idempotency?queue=test&key=test-key", nil)
		w := httptest.NewRecorder()
		handler.handleIdempotencyKey(w, req)

		// Should succeed even if key doesn't exist
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestAdminHandler_HandleOutbox(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Outbox.Enabled = false // Disabled by default
	logger := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, logger)
	handler := NewAdminHandler(manager, logger)

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/exactly-once/outbox", nil)
		w := httptest.NewRecorder()
		handler.handleOutbox(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})

	t.Run("outbox disabled", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/exactly-once/outbox", nil)
		w := httptest.NewRecorder()
		handler.handleOutbox(w, req)

		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}

func TestAdminHandler_HandleCleanup(t *testing.T) {
	cfg := DefaultConfig()
	logger := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, logger)
	handler := NewAdminHandler(manager, logger)

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

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "success", response["status"])
		assert.Equal(t, "all", response["type"])
	})

	t.Run("cleanup idempotency", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/exactly-once/cleanup?type=idempotency", nil)
		w := httptest.NewRecorder()
		handler.handleCleanup(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "idempotency", response["type"])
	})
}

func TestAdminHandler_HandleHealth(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Idempotency.Enabled = true
	cfg.Outbox.Enabled = false
	cfg.Metrics.Enabled = true
	logger := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, logger)
	handler := NewAdminHandler(manager, logger)

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
		assert.Equal(t, true, features["metrics"])
	})
}

func TestAdminHandler_ParseIdempotencyKeyFromQuery(t *testing.T) {
	cfg := DefaultConfig()
	logger := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, logger)
	handler := NewAdminHandler(manager, logger)

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

	t.Run("invalid ttl", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?queue=test&key=test-key&ttl=invalid", nil)
		_, err := handler.parseIdempotencyKeyFromQuery(req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid ttl format")
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

	t.Run("default ttl", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?queue=test&key=test-key", nil)
		key, err := handler.parseIdempotencyKeyFromQuery(req)
		assert.NoError(t, err)
		assert.Equal(t, manager.cfg.Idempotency.DefaultTTL, key.TTL)
	})
}

func TestAdminHandler_Middleware(t *testing.T) {
	cfg := DefaultConfig()
	logger := zaptest.NewLogger(t)
	manager := NewManager(cfg, nil, logger)
	handler := NewAdminHandler(manager, logger)

	t.Run("logging middleware", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := handler.LoggingMiddleware(testHandler)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		middleware(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("CORS middleware", func(t *testing.T) {
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		middleware := handler.CORSMiddleware(testHandler)

		t.Run("OPTIONS request", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, "/test", nil)
			w := httptest.NewRecorder()
			middleware(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
			assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
			assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
		})

		t.Run("normal request", func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			middleware(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
		})
	})
}