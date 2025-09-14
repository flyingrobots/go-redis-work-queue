// Copyright 2025 James Ross
package exactlyonce

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
)

// AdminHandler provides HTTP endpoints for managing exactly-once patterns
type AdminHandler struct {
	manager *Manager
	log     *zap.Logger
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(manager *Manager, log *zap.Logger) *AdminHandler {
	return &AdminHandler{
		manager: manager,
		log:     log,
	}
}

// RegisterRoutes registers the admin API routes
func (h *AdminHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/exactly-once/stats", h.handleStats)
	mux.HandleFunc("/api/v1/exactly-once/idempotency", h.handleIdempotencyKey)
	mux.HandleFunc("/api/v1/exactly-once/outbox", h.handleOutbox)
	mux.HandleFunc("/api/v1/exactly-once/cleanup", h.handleCleanup)
	mux.HandleFunc("/api/v1/exactly-once/health", h.handleHealth)
}

// handleStats returns deduplication statistics
func (h *AdminHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	queueName := r.URL.Query().Get("queue")
	tenantID := r.URL.Query().Get("tenant")

	if queueName == "" {
		http.Error(w, "queue parameter is required", http.StatusBadRequest)
		return
	}

	stats, err := h.manager.GetDedupStats(r.Context(), queueName, tenantID)
	if err != nil {
		h.log.Error("Failed to get dedup stats", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to get stats: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		h.log.Error("Failed to encode stats response", zap.Error(err))
	}
}

// handleIdempotencyKey manages idempotency keys
func (h *AdminHandler) handleIdempotencyKey(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.handleGetIdempotencyKey(w, r)
	case http.MethodPost:
		h.handleCreateIdempotencyKey(w, r)
	case http.MethodDelete:
		h.handleDeleteIdempotencyKey(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleGetIdempotencyKey checks if an idempotency key exists
func (h *AdminHandler) handleGetIdempotencyKey(w http.ResponseWriter, r *http.Request) {
	key, err := h.parseIdempotencyKeyFromQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !h.manager.cfg.Idempotency.Enabled {
		http.Error(w, "Idempotency checking is disabled", http.StatusServiceUnavailable)
		return
	}

	result, err := h.manager.storage.Check(r.Context(), *key)
	if err != nil {
		h.log.Error("Failed to check idempotency key", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to check key: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"key":           key.ID,
		"queue":         key.QueueName,
		"tenant":        key.TenantID,
		"is_first_time": result.IsFirstTime,
		"existing_value": result.ExistingValue,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCreateIdempotencyKey creates a new idempotency key
func (h *AdminHandler) handleCreateIdempotencyKey(w http.ResponseWriter, r *http.Request) {
	var req struct {
		QueueName string      `json:"queue_name"`
		TenantID  string      `json:"tenant_id,omitempty"`
		KeyID     string      `json:"key_id,omitempty"`
		Value     interface{} `json:"value"`
		TTL       string      `json:"ttl,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.QueueName == "" {
		http.Error(w, "queue_name is required", http.StatusBadRequest)
		return
	}

	key := h.manager.GenerateIdempotencyKey(req.QueueName, req.TenantID)
	if req.KeyID != "" {
		key.ID = req.KeyID
	}

	if req.TTL != "" {
		ttl, err := time.ParseDuration(req.TTL)
		if err != nil {
			http.Error(w, "Invalid TTL format", http.StatusBadRequest)
			return
		}
		key.TTL = ttl
	}

	if err := h.manager.storage.Set(r.Context(), key, req.Value); err != nil {
		h.log.Error("Failed to set idempotency key", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to set key: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"key":        key.ID,
		"queue":      key.QueueName,
		"tenant":     key.TenantID,
		"created_at": key.CreatedAt,
		"ttl":        key.TTL.String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// handleDeleteIdempotencyKey deletes an idempotency key
func (h *AdminHandler) handleDeleteIdempotencyKey(w http.ResponseWriter, r *http.Request) {
	key, err := h.parseIdempotencyKeyFromQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.manager.storage.Delete(r.Context(), *key); err != nil {
		h.log.Error("Failed to delete idempotency key", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to delete key: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleOutbox manages outbox events
func (h *AdminHandler) handleOutbox(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.handlePublishOutbox(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePublishOutbox triggers outbox event publishing
func (h *AdminHandler) handlePublishOutbox(w http.ResponseWriter, r *http.Request) {
	if !h.manager.cfg.Outbox.Enabled {
		http.Error(w, "Outbox pattern is disabled", http.StatusServiceUnavailable)
		return
	}

	startTime := time.Now()
	err := h.manager.PublishOutboxEvents(r.Context())
	duration := time.Since(startTime)

	if err != nil {
		h.log.Error("Failed to publish outbox events", zap.Error(err))
		http.Error(w, fmt.Sprintf("Failed to publish events: %v", err), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"status":           "success",
		"duration_seconds": duration.Seconds(),
		"timestamp":        time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleCleanup triggers cleanup operations
func (h *AdminHandler) handleCleanup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cleanupType := r.URL.Query().Get("type")
	if cleanupType == "" {
		cleanupType = "all"
	}

	var errors []string

	if cleanupType == "idempotency" || cleanupType == "all" {
		if err := h.manager.CleanupExpiredKeys(r.Context()); err != nil {
			errors = append(errors, fmt.Sprintf("idempotency cleanup failed: %v", err))
		}
	}

	if cleanupType == "outbox" || cleanupType == "all" {
		if err := h.manager.CleanupOutboxEvents(r.Context()); err != nil {
			errors = append(errors, fmt.Sprintf("outbox cleanup failed: %v", err))
		}
	}

	response := map[string]interface{}{
		"status":    "success",
		"timestamp": time.Now().UTC(),
		"type":      cleanupType,
	}

	if len(errors) > 0 {
		response["status"] = "partial_failure"
		response["errors"] = errors
		w.WriteHeader(http.StatusPartialContent)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleHealth returns health status
func (h *AdminHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"features": map[string]bool{
			"idempotency": h.manager.cfg.Idempotency.Enabled,
			"outbox":      h.manager.cfg.Outbox.Enabled,
			"metrics":     h.manager.cfg.Metrics.Enabled,
		},
	}

	// Basic health checks
	if h.manager.cfg.Idempotency.Enabled && h.manager.storage == nil {
		health["status"] = "degraded"
		health["issues"] = []string{"idempotency storage not available"}
	}

	if h.manager.cfg.Outbox.Enabled && h.manager.outbox == nil {
		if health["issues"] == nil {
			health["issues"] = []string{}
		}
		health["issues"] = append(health["issues"].([]string), "outbox storage not available")
		health["status"] = "degraded"
	}

	statusCode := http.StatusOK
	if health["status"] == "degraded" {
		statusCode = http.StatusPartialContent
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(health)
}

// parseIdempotencyKeyFromQuery parses an idempotency key from query parameters
func (h *AdminHandler) parseIdempotencyKeyFromQuery(r *http.Request) (*IdempotencyKey, error) {
	queueName := r.URL.Query().Get("queue")
	tenantID := r.URL.Query().Get("tenant")
	keyID := r.URL.Query().Get("key")
	ttlStr := r.URL.Query().Get("ttl")

	if queueName == "" {
		return nil, fmt.Errorf("queue parameter is required")
	}

	if keyID == "" {
		return nil, fmt.Errorf("key parameter is required")
	}

	ttl := h.manager.cfg.Idempotency.DefaultTTL
	if ttlStr != "" {
		parsedTTL, err := time.ParseDuration(ttlStr)
		if err != nil {
			return nil, fmt.Errorf("invalid ttl format: %v", err)
		}
		ttl = parsedTTL
	}

	return &IdempotencyKey{
		ID:        keyID,
		QueueName: queueName,
		TenantID:  tenantID,
		CreatedAt: time.Now().UTC(),
		TTL:       ttl,
	}, nil
}

// Middleware for request logging
func (h *AdminHandler) LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler
		next(w, r)

		// Log the request
		h.log.Info("Admin API request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("query", r.URL.RawQuery),
			zap.Duration("duration", time.Since(start)),
			zap.String("user_agent", r.UserAgent()),
		)
	}
}

// Middleware for CORS headers
func (h *AdminHandler) CORSMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}