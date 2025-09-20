// Copyright 2025 James Ross
package archives

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// HTTPHandler provides HTTP endpoints for long-term archives
type HTTPHandler struct {
	manager Manager
	logger  *zap.Logger
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(manager Manager, logger *zap.Logger) *HTTPHandler {
	return &HTTPHandler{
		manager: manager,
		logger:  logger,
	}
}

// RegisterRoutes registers HTTP routes
func (h *HTTPHandler) RegisterRoutes(router *mux.Router) {
	archive := router.PathPrefix("/api/v1/archive").Subrouter()

	// Archive operations
	archive.HandleFunc("/jobs", h.archiveJob).Methods("POST")
	archive.HandleFunc("/jobs/{jobId}", h.getArchivedJob).Methods("GET")
	archive.HandleFunc("/jobs/search", h.searchJobs).Methods("POST")

	// Export operations
	archive.HandleFunc("/export", h.exportJobs).Methods("POST")
	archive.HandleFunc("/export/{exportId}", h.getExportStatus).Methods("GET")
	archive.HandleFunc("/export/{exportId}/cancel", h.cancelExport).Methods("POST")
	archive.HandleFunc("/exports", h.listExports).Methods("GET")

	// Statistics
	archive.HandleFunc("/stats", h.getStats).Methods("GET")

	// Schema management
	archive.HandleFunc("/schema/version", h.getSchemaVersion).Methods("GET")
	archive.HandleFunc("/schema/upgrade", h.upgradeSchema).Methods("POST")
	archive.HandleFunc("/schema/evolution", h.getSchemaEvolution).Methods("GET")

	// Retention management
	archive.HandleFunc("/retention/cleanup", h.cleanupExpired).Methods("POST")
	archive.HandleFunc("/retention/policy", h.getRetentionPolicy).Methods("GET")
	archive.HandleFunc("/retention/policy", h.updateRetentionPolicy).Methods("PUT")
	archive.HandleFunc("/retention/gdpr", h.processGDPRDelete).Methods("POST")

	// Query templates
	archive.HandleFunc("/templates", h.getQueryTemplates).Methods("GET")
	archive.HandleFunc("/templates", h.addQueryTemplate).Methods("POST")
	archive.HandleFunc("/templates/{templateName}/execute", h.executeQuery).Methods("POST")

	// Health and monitoring
	archive.HandleFunc("/health", h.getHealth).Methods("GET")
}

// archiveJob handles POST /api/v1/archive/jobs
func (h *HTTPHandler) archiveJob(w http.ResponseWriter, r *http.Request) {
	var job ArchiveJob
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err := h.manager.ArchiveJob(r.Context(), job)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to archive job", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "archived"})
}

// getArchivedJob handles GET /api/v1/archive/jobs/{jobId}
func (h *HTTPHandler) getArchivedJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID := vars["jobId"]

	job, err := h.manager.GetArchivedJob(r.Context(), jobID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Job not found", err)
		return
	}

	h.writeJSON(w, job)
}

// searchJobs handles POST /api/v1/archive/jobs/search
func (h *HTTPHandler) searchJobs(w http.ResponseWriter, r *http.Request) {
	var query SearchQuery
	if err := json.NewDecoder(r.Body).Decode(&query); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Set defaults
	if query.Limit <= 0 {
		query.Limit = 100
	}
	if query.OrderBy == "" {
		query.OrderBy = "completed_at"
	}
	if query.OrderDir == "" {
		query.OrderDir = "DESC"
	}

	jobs, err := h.manager.SearchJobs(r.Context(), query)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Search failed", err)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"jobs":  jobs,
		"count": len(jobs),
		"query": query,
	})
}

// exportJobs handles POST /api/v1/archive/export
func (h *HTTPHandler) exportJobs(w http.ResponseWriter, r *http.Request) {
	var request ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Search for jobs to export
	jobs, err := h.manager.SearchJobs(r.Context(), request.Query)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to search jobs", err)
		return
	}

	if len(jobs) == 0 {
		h.writeError(w, http.StatusBadRequest, "No jobs found matching query", nil)
		return
	}

	// Export jobs
	status, err := h.manager.ExportJobs(r.Context(), jobs, request.Type)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Export failed", err)
		return
	}

	h.writeJSON(w, status)
}

// getExportStatus handles GET /api/v1/archive/export/{exportId}
func (h *HTTPHandler) getExportStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exportID := vars["exportId"]

	status, err := h.manager.GetExportStatus(r.Context(), exportID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "Export not found", err)
		return
	}

	h.writeJSON(w, status)
}

// cancelExport handles POST /api/v1/archive/export/{exportId}/cancel
func (h *HTTPHandler) cancelExport(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exportID := vars["exportId"]

	err := h.manager.CancelExport(r.Context(), exportID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to cancel export", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "canceled"})
}

// listExports handles GET /api/v1/archive/exports
func (h *HTTPHandler) listExports(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50 // default
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	offset := 0 // default
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	exports, err := h.manager.ListExports(r.Context(), limit, offset)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to list exports", err)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"exports": exports,
		"count":   len(exports),
		"limit":   limit,
		"offset":  offset,
	})
}

// getStats handles GET /api/v1/archive/stats
func (h *HTTPHandler) getStats(w http.ResponseWriter, r *http.Request) {
	windowStr := r.URL.Query().Get("window")
	window := 24 * time.Hour // default

	if windowStr != "" {
		if parsed, err := time.ParseDuration(windowStr); err == nil {
			window = parsed
		}
	}

	stats, err := h.manager.GetStats(r.Context(), window)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get stats", err)
		return
	}

	h.writeJSON(w, stats)
}

// getSchemaVersion handles GET /api/v1/archive/schema/version
func (h *HTTPHandler) getSchemaVersion(w http.ResponseWriter, r *http.Request) {
	version, err := h.manager.GetSchemaVersion(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get schema version", err)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"version": version,
	})
}

// upgradeSchema handles POST /api/v1/archive/schema/upgrade
func (h *HTTPHandler) upgradeSchema(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Version int `json:"version"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err := h.manager.UpgradeSchema(r.Context(), request.Version)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Schema upgrade failed", err)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"status":  "upgraded",
		"version": request.Version,
	})
}

// getSchemaEvolution handles GET /api/v1/archive/schema/evolution
func (h *HTTPHandler) getSchemaEvolution(w http.ResponseWriter, r *http.Request) {
	evolution, err := h.manager.GetSchemaEvolution(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get schema evolution", err)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"evolution": evolution,
	})
}

// cleanupExpired handles POST /api/v1/archive/retention/cleanup
func (h *HTTPHandler) cleanupExpired(w http.ResponseWriter, r *http.Request) {
	cleaned, err := h.manager.CleanupExpired(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Cleanup failed", err)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"records_cleaned": cleaned,
		"timestamp":       time.Now(),
	})
}

// getRetentionPolicy handles GET /api/v1/archive/retention/policy
func (h *HTTPHandler) getRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	// This would need to be implemented in the manager interface
	h.writeJSON(w, map[string]string{
		"message": "Retention policy endpoint not fully implemented",
	})
}

// updateRetentionPolicy handles PUT /api/v1/archive/retention/policy
func (h *HTTPHandler) updateRetentionPolicy(w http.ResponseWriter, r *http.Request) {
	var policy RetentionConfig
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// This would need to be implemented in the manager interface
	h.writeJSON(w, map[string]string{
		"status": "updated",
	})
}

// processGDPRDelete handles POST /api/v1/archive/retention/gdpr
func (h *HTTPHandler) processGDPRDelete(w http.ResponseWriter, r *http.Request) {
	var request GDPRDeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Set request metadata
	if request.ID == "" {
		request.ID = generateGDPRRequestID()
	}
	request.RequestedAt = time.Now()
	request.Status = "processing"

	err := h.manager.ProcessGDPRDelete(r.Context(), request)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "GDPR delete failed", err)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"request_id": request.ID,
		"status":     "processing",
	})
}

// getQueryTemplates handles GET /api/v1/archive/templates
func (h *HTTPHandler) getQueryTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.manager.GetQueryTemplates(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get templates", err)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

// addQueryTemplate handles POST /api/v1/archive/templates
func (h *HTTPHandler) addQueryTemplate(w http.ResponseWriter, r *http.Request) {
	var template QueryTemplate
	if err := json.NewDecoder(r.Body).Decode(&template); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	template.CreatedAt = time.Now()

	err := h.manager.AddQueryTemplate(r.Context(), template)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to add template", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "added"})
}

// executeQuery handles POST /api/v1/archive/templates/{templateName}/execute
func (h *HTTPHandler) executeQuery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	templateName := vars["templateName"]

	var request struct {
		Parameters map[string]interface{} `json:"parameters"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	result, err := h.manager.ExecuteQuery(r.Context(), templateName, request.Parameters)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Query execution failed", err)
		return
	}

	h.writeJSON(w, result)
}

// getHealth handles GET /api/v1/archive/health
func (h *HTTPHandler) getHealth(w http.ResponseWriter, r *http.Request) {
	health, err := h.manager.GetHealth(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Health check failed", err)
		return
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

// Middleware for request logging
func (h *HTTPHandler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		h.logger.Info("Archive API request started",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr))

		next.ServeHTTP(w, r)

		h.logger.Info("Archive API request completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", time.Since(start)))
	})
}

// Middleware for request validation
func (h *HTTPHandler) ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" || r.Method == "PUT" {
			contentType := r.Header.Get("Content-Type")
			if contentType != "application/json" {
				h.writeError(w, http.StatusBadRequest, "Content-Type must be application/json", nil)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// generateGDPRRequestID generates a unique ID for GDPR requests
func generateGDPRRequestID() string {
	return fmt.Sprintf("gdpr_%d", time.Now().UnixNano())
}
