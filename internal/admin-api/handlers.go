// Copyright 2025 James Ross
package adminapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/admin"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// Handler holds the API handler dependencies
type Handler struct {
	cfg       *config.Config
	apiCfg    *Config
	rdb       *redis.Client
	logger    *zap.Logger
	auditLog  *AuditLogger
}

// NewHandler creates a new API handler
func NewHandler(cfg *config.Config, apiCfg *Config, rdb *redis.Client, logger *zap.Logger, auditLog *AuditLogger) *Handler {
	return &Handler{
		cfg:      cfg,
		apiCfg:   apiCfg,
		rdb:      rdb,
		logger:   logger,
		auditLog: auditLog,
	}
}

// GetStats handles GET /api/v1/stats
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	stats, err := admin.Stats(ctx, h.cfg, h.rdb)
	if err != nil {
		h.logger.Error("Failed to get stats", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "STATS_ERROR", "Failed to retrieve statistics")
		return
	}

	response := StatsResponse{
		Queues:          stats.Queues,
		ProcessingLists: stats.ProcessingLists,
		Heartbeats:      stats.Heartbeats,
		Timestamp:       time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// GetStatsKeys handles GET /api/v1/stats/keys
func (h *Handler) GetStatsKeys(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	stats, err := admin.StatsKeys(ctx, h.cfg, h.rdb)
	if err != nil {
		h.logger.Error("Failed to get stats keys", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "STATS_ERROR", "Failed to retrieve key statistics")
		return
	}

	response := StatsKeysResponse{
		QueueLengths:    stats.QueueLengths,
		ProcessingLists: stats.ProcessingLists,
		ProcessingItems: stats.ProcessingItems,
		Heartbeats:      stats.Heartbeats,
		RateLimitKey:    stats.RateLimitKey,
		RateLimitTTL:    stats.RateLimitTTL,
		Timestamp:       time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// PeekQueue handles GET /api/v1/queues/{queue}/peek
func (h *Handler) PeekQueue(w http.ResponseWriter, r *http.Request) {
	// Extract queue name from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		writeError(w, http.StatusBadRequest, "INVALID_PATH", "Invalid path format")
		return
	}
	queue := parts[4]

	// Get count parameter
	count := 10
	if c := r.URL.Query().Get("count"); c != "" {
		if n, err := strconv.Atoi(c); err == nil && n > 0 && n <= 100 {
			count = n
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	result, err := admin.Peek(ctx, h.cfg, h.rdb, queue, int64(count))
	if err != nil {
		h.logger.Error("Failed to peek queue", zap.Error(err), zap.String("queue", queue))
		writeError(w, http.StatusBadRequest, "PEEK_ERROR", err.Error())
		return
	}

	response := PeekResponse{
		Queue:     result.Queue,
		Items:     result.Items,
		Count:     len(result.Items),
		Timestamp: time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// PurgeDLQ handles DELETE /api/v1/queues/dlq
func (h *Handler) PurgeDLQ(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req PurgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate confirmation
	if req.Confirmation != h.apiCfg.ConfirmationPhrase {
		writeError(w, http.StatusBadRequest, "CONFIRMATION_FAILED",
			fmt.Sprintf("Confirmation phrase must be '%s'", h.apiCfg.ConfirmationPhrase))
		return
	}

	if req.Reason == "" || len(req.Reason) < 3 {
		writeError(w, http.StatusBadRequest, "REASON_REQUIRED", "A valid reason is required for this operation")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	// Get DLQ length before purge
	dlqLen, _ := h.rdb.LLen(ctx, h.cfg.Worker.DeadLetterList).Result()

	// Perform purge
	err := admin.PurgeDLQ(ctx, h.cfg, h.rdb)
	if err != nil {
		h.logger.Error("Failed to purge DLQ", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "PURGE_ERROR", "Failed to purge dead letter queue")
		return
	}

	// Log audit entry
	if h.auditLog != nil {
		entry := AuditEntry{
			ID:        generateID(),
			Timestamp: time.Now(),
			Action:    "PURGE_DLQ",
			Resource:  h.cfg.Worker.DeadLetterList,
			Result:    "SUCCESS",
			Reason:    req.Reason,
			Details: map[string]interface{}{
				"items_deleted": dlqLen,
			},
			IP:        getClientIP(r),
			UserAgent: r.UserAgent(),
		}

		if claims, ok := r.Context().Value(contextKeyClaims).(*Claims); ok {
			entry.User = claims.Subject
		}

		h.auditLog.Log(entry)
	}

	response := PurgeResponse{
		Success:      true,
		ItemsDeleted: dlqLen,
		Message:      fmt.Sprintf("Successfully purged %d items from dead letter queue", dlqLen),
		Timestamp:    time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// PurgeAll handles DELETE /api/v1/queues/all
func (h *Handler) PurgeAll(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req PurgeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Require double confirmation for this dangerous operation
	expectedPhrase := h.apiCfg.ConfirmationPhrase + "_ALL"
	if req.Confirmation != expectedPhrase {
		writeError(w, http.StatusBadRequest, "CONFIRMATION_FAILED",
			fmt.Sprintf("Confirmation phrase must be '%s' for purging all queues", expectedPhrase))
		return
	}

	if req.Reason == "" || len(req.Reason) < 10 {
		writeError(w, http.StatusBadRequest, "REASON_REQUIRED", "A detailed reason (min 10 chars) is required for this operation")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Perform purge
	deleted, err := admin.PurgeAll(ctx, h.cfg, h.rdb)
	if err != nil {
		h.logger.Error("Failed to purge all", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "PURGE_ERROR", "Failed to purge all queues")
		return
	}

	// Log audit entry
	if h.auditLog != nil {
		entry := AuditEntry{
			ID:        generateID(),
			Timestamp: time.Now(),
			Action:    "PURGE_ALL",
			Resource:  "ALL_QUEUES",
			Result:    "SUCCESS",
			Reason:    req.Reason,
			Details: map[string]interface{}{
				"keys_deleted": deleted,
			},
			IP:        getClientIP(r),
			UserAgent: r.UserAgent(),
		}

		if claims, ok := r.Context().Value(contextKeyClaims).(*Claims); ok {
			entry.User = claims.Subject
		}

		h.auditLog.Log(entry)
	}

	response := PurgeResponse{
		Success:      true,
		ItemsDeleted: deleted,
		Message:      fmt.Sprintf("Successfully purged %d keys from all queues", deleted),
		Timestamp:    time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// RunBenchmark handles POST /api/v1/bench
func (h *Handler) RunBenchmark(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req BenchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	// Validate parameters
	if req.Count <= 0 || req.Count > 10000 {
		writeError(w, http.StatusBadRequest, "INVALID_COUNT", "Count must be between 1 and 10000")
		return
	}

	if req.Priority != "high" && req.Priority != "low" {
		writeError(w, http.StatusBadRequest, "INVALID_PRIORITY", "Priority must be 'high' or 'low'")
		return
	}

	if req.Rate <= 0 {
		req.Rate = 100
	}

	timeout := 30 * time.Second
	if req.Timeout > 0 {
		timeout = time.Duration(req.Timeout) * time.Second
	}

	ctx, cancel := context.WithTimeout(r.Context(), timeout+10*time.Second)
	defer cancel()

	// Run benchmark
	result, err := admin.Bench(ctx, h.cfg, h.rdb, req.Priority, req.Count, req.Rate, timeout)
	if err != nil {
		h.logger.Error("Failed to run benchmark", zap.Error(err))
		writeError(w, http.StatusInternalServerError, "BENCH_ERROR", "Failed to run benchmark")
		return
	}

	// Log audit entry
	if h.auditLog != nil {
		entry := AuditEntry{
			ID:        generateID(),
			Timestamp: time.Now(),
			Action:    "RUN_BENCHMARK",
			Resource:  req.Priority,
			Result:    "SUCCESS",
			Details: map[string]interface{}{
				"count":      req.Count,
				"rate":       req.Rate,
				"throughput": result.Throughput,
			},
			IP:        getClientIP(r),
			UserAgent: r.UserAgent(),
		}

		if claims, ok := r.Context().Value(contextKeyClaims).(*Claims); ok {
			entry.User = claims.Subject
		}

		h.auditLog.Log(entry)
	}

	response := BenchResponse{
		Count:      result.Count,
		Duration:   result.Duration,
		Throughput: result.Throughput,
		P50:        result.P50,
		P95:        result.P95,
		Timestamp:  time.Now(),
	}

	writeJSON(w, http.StatusOK, response)
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	response := ErrorResponse{
		Error: message,
		Code:  code,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}