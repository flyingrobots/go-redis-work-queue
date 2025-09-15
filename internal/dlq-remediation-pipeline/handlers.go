// Copyright 2025 James Ross
package dlqremediation

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// HTTPHandler provides HTTP endpoints for DLQ remediation pipeline
type HTTPHandler struct {
	pipeline *RemediationPipeline
	logger   *zap.Logger
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(pipeline *RemediationPipeline, logger *zap.Logger) *HTTPHandler {
	return &HTTPHandler{
		pipeline: pipeline,
		logger:   logger,
	}
}

// RegisterRoutes registers HTTP routes
func (h *HTTPHandler) RegisterRoutes(router *mux.Router) {
	// Pipeline management
	api := router.PathPrefix("/api/v1/dlq-remediation").Subrouter()

	// Pipeline control
	api.HandleFunc("/pipeline/start", h.startPipeline).Methods("POST")
	api.HandleFunc("/pipeline/stop", h.stopPipeline).Methods("POST")
	api.HandleFunc("/pipeline/pause", h.pausePipeline).Methods("POST")
	api.HandleFunc("/pipeline/resume", h.resumePipeline).Methods("POST")
	api.HandleFunc("/pipeline/status", h.getPipelineStatus).Methods("GET")
	api.HandleFunc("/pipeline/metrics", h.getPipelineMetrics).Methods("GET")

	// Batch processing
	api.HandleFunc("/pipeline/process-batch", h.processBatch).Methods("POST")
	api.HandleFunc("/pipeline/dry-run", h.dryRunBatch).Methods("POST")

	// Rule management
	api.HandleFunc("/rules", h.getRules).Methods("GET")
	api.HandleFunc("/rules", h.createRule).Methods("POST")
	api.HandleFunc("/rules/{ruleID}", h.getRule).Methods("GET")
	api.HandleFunc("/rules/{ruleID}", h.updateRule).Methods("PUT")
	api.HandleFunc("/rules/{ruleID}", h.deleteRule).Methods("DELETE")
	api.HandleFunc("/rules/{ruleID}/enable", h.enableRule).Methods("POST")
	api.HandleFunc("/rules/{ruleID}/disable", h.disableRule).Methods("POST")
	api.HandleFunc("/rules/{ruleID}/test", h.testRule).Methods("POST")

	// Classification
	api.HandleFunc("/classify", h.classifyJob).Methods("POST")
	api.HandleFunc("/classify/batch", h.classifyBatch).Methods("POST")

	// Audit and monitoring
	api.HandleFunc("/audit", h.getAuditLog).Methods("GET")
	api.HandleFunc("/audit/stats", h.getAuditStats).Methods("GET")

	// Health check
	api.HandleFunc("/health", h.getHealth).Methods("GET")
}

// Pipeline control handlers

func (h *HTTPHandler) startPipeline(w http.ResponseWriter, r *http.Request) {
	err := h.pipeline.Start(r.Context())
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to start pipeline", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "started"})
}

func (h *HTTPHandler) stopPipeline(w http.ResponseWriter, r *http.Request) {
	err := h.pipeline.Stop(r.Context())
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to stop pipeline", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "stopped"})
}

func (h *HTTPHandler) pausePipeline(w http.ResponseWriter, r *http.Request) {
	err := h.pipeline.Pause(r.Context())
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to pause pipeline", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "paused"})
}

func (h *HTTPHandler) resumePipeline(w http.ResponseWriter, r *http.Request) {
	err := h.pipeline.Resume(r.Context())
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Failed to resume pipeline", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "resumed"})
}

func (h *HTTPHandler) getPipelineStatus(w http.ResponseWriter, r *http.Request) {
	state := h.pipeline.GetState()
	h.writeJSON(w, state)
}

func (h *HTTPHandler) getPipelineMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := h.pipeline.GetMetrics()
	h.writeJSON(w, metrics)
}

// Batch processing handlers

func (h *HTTPHandler) processBatch(w http.ResponseWriter, r *http.Request) {
	result, err := h.pipeline.ProcessBatch(r.Context(), false)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Batch processing failed", err)
		return
	}

	h.writeJSON(w, result)
}

func (h *HTTPHandler) dryRunBatch(w http.ResponseWriter, r *http.Request) {
	result, err := h.pipeline.ProcessBatch(r.Context(), true)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Dry run failed", err)
		return
	}

	h.writeJSON(w, result)
}

// Rule management handlers

func (h *HTTPHandler) getRules(w http.ResponseWriter, r *http.Request) {
	rules := h.pipeline.GetRules()

	// Apply filters
	enabled := r.URL.Query().Get("enabled")
	if enabled != "" {
		enabledBool, _ := strconv.ParseBool(enabled)
		var filteredRules []RemediationRule
		for _, rule := range rules {
			if rule.Enabled == enabledBool {
				filteredRules = append(filteredRules, rule)
			}
		}
		rules = filteredRules
	}

	tag := r.URL.Query().Get("tag")
	if tag != "" {
		var filteredRules []RemediationRule
		for _, rule := range rules {
			for _, ruleTag := range rule.Tags {
				if ruleTag == tag {
					filteredRules = append(filteredRules, rule)
					break
				}
			}
		}
		rules = filteredRules
	}

	h.writeJSON(w, map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}

func (h *HTTPHandler) createRule(w http.ResponseWriter, r *http.Request) {
	var rule RemediationRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userID := h.getUserID(r)
	err := h.pipeline.AddRule(r.Context(), rule, userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to create rule", err)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"status":  "created",
		"rule_id": rule.ID,
	})
}

func (h *HTTPHandler) getRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["ruleID"]

	rules := h.pipeline.GetRules()
	for _, rule := range rules {
		if rule.ID == ruleID {
			h.writeJSON(w, rule)
			return
		}
	}

	h.writeError(w, http.StatusNotFound, "Rule not found", nil)
}

func (h *HTTPHandler) updateRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["ruleID"]

	var rule RemediationRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userID := h.getUserID(r)
	err := h.pipeline.UpdateRule(r.Context(), ruleID, rule, userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update rule", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "updated"})
}

func (h *HTTPHandler) deleteRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["ruleID"]

	userID := h.getUserID(r)
	err := h.pipeline.DeleteRule(r.Context(), ruleID, userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to delete rule", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "deleted"})
}

func (h *HTTPHandler) enableRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["ruleID"]

	// Get existing rule
	rules := h.pipeline.GetRules()
	var rule *RemediationRule
	for _, r := range rules {
		if r.ID == ruleID {
			rule = &r
			break
		}
	}

	if rule == nil {
		h.writeError(w, http.StatusNotFound, "Rule not found", nil)
		return
	}

	rule.Enabled = true
	userID := h.getUserID(r)
	err := h.pipeline.UpdateRule(r.Context(), ruleID, *rule, userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to enable rule", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "enabled"})
}

func (h *HTTPHandler) disableRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["ruleID"]

	// Get existing rule
	rules := h.pipeline.GetRules()
	var rule *RemediationRule
	for _, r := range rules {
		if r.ID == ruleID {
			rule = &r
			break
		}
	}

	if rule == nil {
		h.writeError(w, http.StatusNotFound, "Rule not found", nil)
		return
	}

	rule.Enabled = false
	userID := h.getUserID(r)
	err := h.pipeline.UpdateRule(r.Context(), ruleID, *rule, userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to disable rule", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "disabled"})
}

func (h *HTTPHandler) testRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["ruleID"]

	var testJob DLQJob
	if err := json.NewDecoder(r.Body).Decode(&testJob); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Get the rule
	rules := h.pipeline.GetRules()
	var rule *RemediationRule
	for _, r := range rules {
		if r.ID == ruleID {
			rule = &r
			break
		}
	}

	if rule == nil {
		h.writeError(w, http.StatusNotFound, "Rule not found", nil)
		return
	}

	// Test classification
	classification, err := h.pipeline.classifier.Classify(r.Context(), &testJob, []RemediationRule{*rule})
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Classification failed", err)
		return
	}

	// Test actions (dry run)
	result, err := h.pipeline.actionExecutor.Execute(r.Context(), &testJob, rule.Actions, true)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Action test failed", err)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"rule_id":        ruleID,
		"classification": classification,
		"execution":      result,
		"would_match":    classification.RuleID == ruleID,
	})
}

// Classification handlers

func (h *HTTPHandler) classifyJob(w http.ResponseWriter, r *http.Request) {
	var job DLQJob
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	rules := h.pipeline.GetRules()
	classification, err := h.pipeline.classifier.Classify(r.Context(), &job, rules)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Classification failed", err)
		return
	}

	h.writeJSON(w, classification)
}

func (h *HTTPHandler) classifyBatch(w http.ResponseWriter, r *http.Request) {
	var jobs []DLQJob
	if err := json.NewDecoder(r.Body).Decode(&jobs); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	rules := h.pipeline.GetRules()
	results := make([]*Classification, len(jobs))

	for i, job := range jobs {
		classification, err := h.pipeline.classifier.Classify(r.Context(), &job, rules)
		if err != nil {
			h.writeError(w, http.StatusInternalServerError, "Batch classification failed", err)
			return
		}
		results[i] = classification
	}

	h.writeJSON(w, map[string]interface{}{
		"classifications": results,
		"count":          len(results),
	})
}

// Audit handlers

func (h *HTTPHandler) getAuditLog(w http.ResponseWriter, r *http.Request) {
	filter := AuditFilter{}

	// Parse query parameters
	if jobID := r.URL.Query().Get("job_id"); jobID != "" {
		filter.JobID = jobID
	}

	if ruleID := r.URL.Query().Get("rule_id"); ruleID != "" {
		filter.RuleID = ruleID
	}

	if action := r.URL.Query().Get("action"); action != "" {
		filter.Action = ActionType(action)
	}

	if userID := r.URL.Query().Get("user_id"); userID != "" {
		filter.UserID = userID
	}

	if result := r.URL.Query().Get("result"); result != "" {
		filter.Result = result
	}

	if dryRun := r.URL.Query().Get("dry_run"); dryRun != "" {
		if dryRunBool, err := strconv.ParseBool(dryRun); err == nil {
			filter.DryRun = &dryRunBool
		}
	}

	if startTime := r.URL.Query().Get("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = &t
		}
	}

	if endTime := r.URL.Query().Get("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = &t
		}
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if limitInt, err := strconv.Atoi(limit); err == nil {
			filter.Limit = limitInt
		}
	} else {
		filter.Limit = 100 // Default limit
	}

	if offset := r.URL.Query().Get("offset"); offset != "" {
		if offsetInt, err := strconv.Atoi(offset); err == nil {
			filter.Offset = offsetInt
		}
	}

	filter.SortBy = r.URL.Query().Get("sort_by")
	if filter.SortBy == "" {
		filter.SortBy = "timestamp"
	}

	filter.SortOrder = r.URL.Query().Get("sort_order")
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}

	entries, err := h.pipeline.auditLogger.GetAuditLog(r.Context(), filter)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get audit log", err)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
		"filter":  filter,
	})
}

func (h *HTTPHandler) getAuditStats(w http.ResponseWriter, r *http.Request) {
	days := 30 // Default
	if daysStr := r.URL.Query().Get("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	stats, err := h.pipeline.auditLogger.GetAuditStatistics(r.Context(), days)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get audit statistics", err)
		return
	}

	h.writeJSON(w, stats)
}

// Health check handler

func (h *HTTPHandler) getHealth(w http.ResponseWriter, r *http.Request) {
	state := h.pipeline.GetState()
	metrics := h.pipeline.GetMetrics()

	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"pipeline": map[string]interface{}{
			"status":           state.Status,
			"uptime":          time.Since(state.StartedAt),
			"last_run":        state.LastRunAt,
			"total_processed": state.TotalProcessed,
			"success_rate":    float64(state.TotalSuccessful) / float64(state.TotalProcessed+1), // +1 to avoid division by zero
		},
		"rules": map[string]interface{}{
			"enabled":  state.RulesEnabled,
			"disabled": state.RulesDisabled,
			"total":    state.RulesEnabled + state.RulesDisabled,
		},
		"metrics": map[string]interface{}{
			"jobs_processed":       metrics.JobsProcessed,
			"actions_successful":   metrics.ActionsSuccessful,
			"actions_failed":       metrics.ActionsFailed,
			"average_latency_ms":   metrics.EndToEndTime,
			"rate_limit_hits":      metrics.RateLimitHits,
			"circuit_breaker_trips": metrics.CircuitBreakerTrips,
		},
	}

	// Determine overall health status
	if state.Status == StatusError {
		health["status"] = "unhealthy"
	} else if state.Status == StatusPaused {
		health["status"] = "degraded"
	}

	// Check if there are recent errors
	if !state.LastErrorAt.IsZero() && time.Since(state.LastErrorAt) < 5*time.Minute {
		health["status"] = "degraded"
		health["last_error"] = state.LastError
		health["last_error_at"] = state.LastErrorAt
	}

	h.writeJSON(w, health)
}

// Helper methods

func (h *HTTPHandler) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to write JSON response", zap.Error(err))
	}
}

func (h *HTTPHandler) writeError(w http.ResponseWriter, status int, message string, err error) {
	h.logger.Error(message, zap.Error(err), zap.Int("status", status))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResponse := map[string]interface{}{
		"error":     message,
		"status":    status,
		"timestamp": time.Now(),
	}

	if err != nil {
		errorResponse["details"] = err.Error()
	}

	json.NewEncoder(w).Encode(errorResponse)
}

func (h *HTTPHandler) getUserID(r *http.Request) string {
	// Extract user ID from authentication context
	// This would typically come from JWT claims or session
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = "anonymous"
	}
	return userID
}

// Middleware for request logging
func (h *HTTPHandler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		h.logger.Info("DLQ remediation API request started",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr))

		next.ServeHTTP(w, r)

		h.logger.Info("DLQ remediation API request completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", time.Since(start)))
	})
}

// Middleware for authentication
func (h *HTTPHandler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple authentication check
		// In a real implementation, this would validate JWT tokens
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			h.writeError(w, http.StatusUnauthorized, "Authorization header required", nil)
			return
		}

		// For now, just check that some token is provided
		if !h.isValidToken(authHeader) {
			h.writeError(w, http.StatusUnauthorized, "Invalid or expired token", nil)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (h *HTTPHandler) isValidToken(token string) bool {
	// Simple token validation
	// In a real implementation, this would validate JWT/PASETO tokens
	return len(token) > 10 // Minimal check
}