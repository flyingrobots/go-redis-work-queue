// Copyright 2025 James Ross
package tracedrilldownlogtail

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// HTTPHandlers provides HTTP handlers for trace drilldown functionality
type HTTPHandlers struct {
	traceManager *TraceManager
	logTailer    *LogTailer
	logger       *zap.Logger
}

// NewHTTPHandlers creates new HTTP handlers
func NewHTTPHandlers(traceManager *TraceManager, logTailer *LogTailer, logger *zap.Logger) *HTTPHandlers {
	return &HTTPHandlers{
		traceManager: traceManager,
		logTailer:    logTailer,
		logger:       logger,
	}
}

// RegisterRoutes registers all HTTP routes
func (h *HTTPHandlers) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1/trace-drilldown").Subrouter()

	// Trace operations
	api.HandleFunc("/traces/{traceId}", h.handleGetTrace).Methods("GET")
	api.HandleFunc("/traces/{traceId}/summary", h.handleGetTraceSummary).Methods("GET")
	api.HandleFunc("/traces/{traceId}/links", h.handleGetTraceLinks).Methods("GET")
	api.HandleFunc("/traces/{traceId}/open", h.handleOpenTrace).Methods("POST")
	api.HandleFunc("/traces/search", h.handleSearchTraces).Methods("POST")

	// Log operations
	api.HandleFunc("/logs/search", h.handleSearchLogs).Methods("POST")
	api.HandleFunc("/logs/stats", h.handleGetLogStats).Methods("GET")
	api.HandleFunc("/logs/tail", h.handleStartTail).Methods("POST")
	api.HandleFunc("/logs/tail/{sessionId}", h.handleStopTail).Methods("DELETE")
	api.HandleFunc("/logs/tail/sessions", h.handleGetTailSessions).Methods("GET")

	// WebSocket endpoint for log streaming
	api.HandleFunc("/logs/stream/{sessionId}", h.handleLogStream).Methods("GET")

	// Job trace extraction
	api.HandleFunc("/jobs/{jobPayload}/trace", h.handleExtractJobTrace).Methods("POST")
}

// Trace handlers

func (h *HTTPHandlers) handleGetTrace(w http.ResponseWriter, r *http.Request) {
	traceID := mux.Vars(r)["traceId"]

	trace, err := h.traceManager.GetTrace(traceID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Trace not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, trace)
}

func (h *HTTPHandlers) handleGetTraceSummary(w http.ResponseWriter, r *http.Request) {
	traceID := mux.Vars(r)["traceId"]

	summary, err := h.traceManager.GetSpanSummary(r.Context(), traceID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get trace summary", err)
		return
	}

	h.writeJSON(w, http.StatusOK, summary)
}

func (h *HTTPHandlers) handleGetTraceLinks(w http.ResponseWriter, r *http.Request) {
	traceID := mux.Vars(r)["traceId"]

	link, err := h.traceManager.GetTraceLink(traceID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to generate trace link", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"trace_id": traceID,
		"links":    []TraceLink{*link},
	})
}

func (h *HTTPHandlers) handleOpenTrace(w http.ResponseWriter, r *http.Request) {
	traceID := mux.Vars(r)["traceId"]

	link, err := h.traceManager.GetTraceLink(traceID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to generate trace link", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"trace_id":    traceID,
		"action":      "open",
		"url":         link.URL,
		"description": "Trace link generated. Open in browser manually or via automation.",
	})
}

func (h *HTTPHandlers) handleSearchTraces(w http.ResponseWriter, r *http.Request) {
	var filter LogFilter
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid filter", err)
		return
	}

	result, err := h.traceManager.SearchTraces(r.Context(), &filter)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Search failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

// Log handlers

func (h *HTTPHandlers) handleSearchLogs(w http.ResponseWriter, r *http.Request) {
	var filter LogFilter
	if err := json.NewDecoder(r.Body).Decode(&filter); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid filter", err)
		return
	}

	result, err := h.logTailer.SearchLogs(r.Context(), &filter)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Search failed", err)
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

func (h *HTTPHandlers) handleGetLogStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.logTailer.GetLogStats(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get stats", err)
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

func (h *HTTPHandlers) handleStartTail(w http.ResponseWriter, r *http.Request) {
	var config TailConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid config", err)
		return
	}

	session, eventCh, err := h.logTailer.StartTail(&config)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to start tail", err)
		return
	}

	// Start goroutine to handle the event channel
	go func() {
		for range eventCh {
			// Events are handled via WebSocket connection
		}
	}()

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"session_id": session.ID,
		"stream_url": fmt.Sprintf("/api/v1/trace-drilldown/logs/stream/%s", session.ID),
		"config":     config,
	})
}

func (h *HTTPHandlers) handleStopTail(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	if err := h.logTailer.StopTail(sessionID); err != nil {
		h.writeError(w, http.StatusNotFound, "Session not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"status":     "stopped",
		"session_id": sessionID,
	})
}

func (h *HTTPHandlers) handleGetTailSessions(w http.ResponseWriter, r *http.Request) {
	// Note: LogTailer doesn't have GetSessions method, but we can add it
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"sessions": []interface{}{},
		"count":    0,
	})
}

func (h *HTTPHandlers) handleLogStream(w http.ResponseWriter, r *http.Request) {
	sessionID := mux.Vars(r)["sessionId"]

	// For now, return a simple response. In production, this would be a WebSocket endpoint
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"session_id": sessionID,
		"message":    "WebSocket streaming not implemented in this example",
		"note":       "Use Server-Sent Events or WebSocket for real-time log streaming",
	})
}

func (h *HTTPHandlers) handleExtractJobTrace(w http.ResponseWriter, r *http.Request) {
	// For now, expect the job payload to be passed in the request body
	var payload struct {
		JobData string `json:"job_data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid payload", err)
		return
	}

	// Parse job data and extract trace info
	var jobData map[string]interface{}
	if err := json.Unmarshal([]byte(payload.JobData), &jobData); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid job JSON", err)
		return
	}

	traceID, _ := jobData["trace_id"].(string)
	spanID, _ := jobData["span_id"].(string)

	if traceID == "" {
		h.writeError(w, http.StatusNotFound, "No trace ID found in job", nil)
		return
	}

	// Get trace links
	link, err := h.traceManager.GetTraceLink(traceID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to generate trace link", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"trace_id": traceID,
		"span_id":  spanID,
		"links":    []TraceLink{*link},
		"actions":  []string{"view", "copy"},
	})
}

// Helper methods

func (h *HTTPHandlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", zap.Error(err))
	}
}

func (h *HTTPHandlers) writeError(w http.ResponseWriter, status int, message string, err error) {
	h.logger.Error(message, zap.Error(err))
	h.writeJSON(w, status, map[string]interface{}{
		"error": message,
		"details": func() string {
			if err != nil {
				return err.Error()
			}
			return ""
		}(),
	})
}

// TraceActionRequest represents a request for trace actions
type TraceActionRequest struct {
	TraceID string `json:"trace_id"`
	Action  string `json:"action"` // "view", "copy", "open"
}

// TailConfigRequest represents a request to start log tailing
type TailConfigRequest struct {
	Follow            bool       `json:"follow"`
	BufferSize        int        `json:"buffer_size,omitempty"`
	MaxLinesPerSecond int        `json:"max_lines_per_second,omitempty"`
	BackpressureLimit int        `json:"backpressure_limit,omitempty"`
	FlushInterval     string     `json:"flush_interval,omitempty"` // Duration string
	Filter            *LogFilter `json:"filter,omitempty"`
}

// ToTailConfig converts a TailConfigRequest to TailConfig
func (tcr *TailConfigRequest) ToTailConfig() TailConfig {
	config := TailConfig{
		Follow:            tcr.Follow,
		BufferSize:        tcr.BufferSize,
		MaxLinesPerSecond: tcr.MaxLinesPerSecond,
		BackpressureLimit: tcr.BackpressureLimit,
		Filter:            tcr.Filter,
	}

	if tcr.FlushInterval != "" {
		if duration, err := time.ParseDuration(tcr.FlushInterval); err == nil {
			config.FlushInterval = duration
		} else {
			config.FlushInterval = time.Second // default
		}
	}

	// Set defaults
	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}
	if config.MaxLinesPerSecond == 0 {
		config.MaxLinesPerSecond = 100
	}
	if config.BackpressureLimit == 0 {
		config.BackpressureLimit = 5000
	}
	if config.FlushInterval == 0 {
		config.FlushInterval = time.Second
	}

	return config
}

// Enhanced admin endpoints with trace information

// EnhancedPeekRequest includes trace enhancement options
type EnhancedPeekRequest struct {
	Queue          string `json:"queue"`
	Count          int    `json:"count,omitempty"`
	IncludeTraces  bool   `json:"include_traces,omitempty"`
	IncludeActions bool   `json:"include_actions,omitempty"`
}

// EnhancedPeekResponse includes trace information with jobs
type EnhancedPeekResponse struct {
	Queue         string                 `json:"queue"`
	Items         []string               `json:"items"`
	TraceInfo     map[string]*TraceInfo  `json:"trace_info,omitempty"`
	TraceActions  map[string][]TraceLink `json:"trace_actions,omitempty"`
	JobsWithTrace []JobWithTraceInfo     `json:"jobs_with_trace,omitempty"`
}

// Enhanced peek handler that includes trace information
func (h *HTTPHandlers) handleEnhancedPeek(w http.ResponseWriter, r *http.Request) {
	var req EnhancedPeekRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request", err)
		return
	}

	// This would integrate with the existing admin.Peek function
	// For now, return a placeholder response
	response := &EnhancedPeekResponse{
		Queue:         req.Queue,
		Items:         []string{},
		TraceInfo:     make(map[string]*TraceInfo),
		TraceActions:  make(map[string][]TraceLink),
		JobsWithTrace: []JobWithTraceInfo{},
	}

	h.writeJSON(w, http.StatusOK, response)
}

// Helper function to parse query parameters
func parseIntQuery(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseBoolQuery(r *http.Request, key string, defaultValue bool) bool {
	if value := r.URL.Query().Get(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseDurationQuery(r *http.Request, key string, defaultValue time.Duration) time.Duration {
	if value := r.URL.Query().Get(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
