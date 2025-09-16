// Copyright 2025 James Ross
package smartretry

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// HTTPHandler provides HTTP endpoints for smart retry strategies
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
	retry := router.PathPrefix("/api/v1/retry").Subrouter()

	// Recommendation endpoints
	retry.HandleFunc("/recommendation", h.getRecommendation).Methods("POST")
	retry.HandleFunc("/preview", h.previewSchedule).Methods("POST")

	// Stats endpoints
	retry.HandleFunc("/stats", h.getStats).Methods("GET")
	retry.HandleFunc("/attempt", h.recordAttempt).Methods("POST")

	// Policy management
	retry.HandleFunc("/policies", h.getPolicies).Methods("GET")
	retry.HandleFunc("/policies", h.addPolicy).Methods("POST")
	retry.HandleFunc("/policies/{name}", h.removePolicy).Methods("DELETE")

	// Model management
	retry.HandleFunc("/bayesian/update", h.updateBayesianModel).Methods("POST")
	retry.HandleFunc("/ml/train", h.trainMLModel).Methods("POST")
	retry.HandleFunc("/ml/deploy", h.deployMLModel).Methods("POST")
	retry.HandleFunc("/ml/rollback", h.rollbackMLModel).Methods("POST")

	// Configuration
	retry.HandleFunc("/strategy", h.getStrategy).Methods("GET")
	retry.HandleFunc("/strategy", h.updateStrategy).Methods("PUT")
	retry.HandleFunc("/guardrails", h.updateGuardrails).Methods("PUT")
}

// getRecommendation handles POST /api/v1/retry/recommendation
func (h *HTTPHandler) getRecommendation(w http.ResponseWriter, r *http.Request) {
	var features RetryFeatures
	if err := json.NewDecoder(r.Body).Decode(&features); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	recommendation, err := h.manager.GetRecommendation(features)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get recommendation", err)
		return
	}

	h.writeJSON(w, recommendation)
}

// previewSchedule handles POST /api/v1/retry/preview
func (h *HTTPHandler) previewSchedule(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Features    RetryFeatures `json:"features"`
		MaxAttempts int           `json:"max_attempts"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	preview, err := h.manager.PreviewRetrySchedule(request.Features, request.MaxAttempts)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to generate preview", err)
		return
	}

	h.writeJSON(w, preview)
}

// getStats handles GET /api/v1/retry/stats
func (h *HTTPHandler) getStats(w http.ResponseWriter, r *http.Request) {
	jobType := r.URL.Query().Get("job_type")
	errorClass := r.URL.Query().Get("error_class")
	windowStr := r.URL.Query().Get("window")

	if jobType == "" || errorClass == "" {
		h.writeError(w, http.StatusBadRequest, "job_type and error_class are required", nil)
		return
	}

	window := 24 * time.Hour // Default window
	if windowStr != "" {
		if parsed, err := time.ParseDuration(windowStr); err == nil {
			window = parsed
		}
	}

	stats, err := h.manager.GetStats(jobType, errorClass, window)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get stats", err)
		return
	}

	h.writeJSON(w, stats)
}

// recordAttempt handles POST /api/v1/retry/attempt
func (h *HTTPHandler) recordAttempt(w http.ResponseWriter, r *http.Request) {
	var attempt AttemptHistory
	if err := json.NewDecoder(r.Body).Decode(&attempt); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.manager.RecordAttempt(attempt); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to record attempt", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "recorded"})
}

// getPolicies handles GET /api/v1/retry/policies
func (h *HTTPHandler) getPolicies(w http.ResponseWriter, r *http.Request) {
	strategy, err := h.manager.GetStrategy()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get strategy", err)
		return
	}

	h.writeJSON(w, map[string]interface{}{
		"policies": strategy.Policies,
	})
}

// addPolicy handles POST /api/v1/retry/policies
func (h *HTTPHandler) addPolicy(w http.ResponseWriter, r *http.Request) {
	var policy RetryPolicy
	if err := json.NewDecoder(r.Body).Decode(&policy); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.manager.AddPolicy(policy); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to add policy", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "added"})
}

// removePolicy handles DELETE /api/v1/retry/policies/{name}
func (h *HTTPHandler) removePolicy(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	name := vars["name"]

	if name == "" {
		h.writeError(w, http.StatusBadRequest, "Policy name is required", nil)
		return
	}

	if err := h.manager.RemovePolicy(name); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to remove policy", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "removed"})
}

// updateBayesianModel handles POST /api/v1/retry/bayesian/update
func (h *HTTPHandler) updateBayesianModel(w http.ResponseWriter, r *http.Request) {
	var request struct {
		JobType    string `json:"job_type"`
		ErrorClass string `json:"error_class"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.manager.UpdateBayesianModel(request.JobType, request.ErrorClass); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update Bayesian model", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "updated"})
}

// trainMLModel handles POST /api/v1/retry/ml/train
func (h *HTTPHandler) trainMLModel(w http.ResponseWriter, r *http.Request) {
	var config MLTrainingConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	model, err := h.manager.TrainMLModel(config)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to train ML model", err)
		return
	}

	h.writeJSON(w, model)
}

// deployMLModel handles POST /api/v1/retry/ml/deploy
func (h *HTTPHandler) deployMLModel(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Model         *MLModel `json:"model"`
		CanaryPercent float64  `json:"canary_percent"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.manager.DeployMLModel(request.Model, request.CanaryPercent); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to deploy ML model", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "deployed"})
}

// rollbackMLModel handles POST /api/v1/retry/ml/rollback
func (h *HTTPHandler) rollbackMLModel(w http.ResponseWriter, r *http.Request) {
	if err := h.manager.RollbackMLModel(); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to rollback ML model", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "rolled_back"})
}

// getStrategy handles GET /api/v1/retry/strategy
func (h *HTTPHandler) getStrategy(w http.ResponseWriter, r *http.Request) {
	strategy, err := h.manager.GetStrategy()
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to get strategy", err)
		return
	}

	h.writeJSON(w, strategy)
}

// updateStrategy handles PUT /api/v1/retry/strategy
func (h *HTTPHandler) updateStrategy(w http.ResponseWriter, r *http.Request) {
	var strategy RetryStrategy
	if err := json.NewDecoder(r.Body).Decode(&strategy); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.manager.UpdateStrategy(&strategy); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update strategy", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "updated"})
}

// updateGuardrails handles PUT /api/v1/retry/guardrails
func (h *HTTPHandler) updateGuardrails(w http.ResponseWriter, r *http.Request) {
	var guardrails PolicyGuardrails
	if err := json.NewDecoder(r.Body).Decode(&guardrails); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if err := h.manager.UpdateGuardrails(guardrails); err != nil {
		h.writeError(w, http.StatusInternalServerError, "Failed to update guardrails", err)
		return
	}

	h.writeJSON(w, map[string]string{"status": "updated"})
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
		"error":   message,
		"status":  status,
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

		h.logger.Info("HTTP request started",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.String("remote_addr", r.RemoteAddr))

		next.ServeHTTP(w, r)

		h.logger.Info("HTTP request completed",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Duration("duration", time.Since(start)))
	})
}

// Middleware for request validation
func (h *HTTPHandler) ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add request validation logic here
		// For example, check content type for POST/PUT requests
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

// Health check endpoint
func (h *HTTPHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now(),
		"service":   "smart-retry-strategies",
	}

	// Add health checks here (Redis connectivity, etc.)
	h.writeJSON(w, health)
}

// Metrics endpoint (placeholder)
func (h *HTTPHandler) Metrics(w http.ResponseWriter, r *http.Request) {
	// This would typically expose Prometheus metrics
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("# Smart Retry Strategies Metrics\n# TODO: Implement Prometheus metrics\n"))
}