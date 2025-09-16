package canary_deployments

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// HTTPHandler provides REST API endpoints for canary deployment management
type HTTPHandler struct {
	manager CanaryManager
	logger  *slog.Logger
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(manager CanaryManager, logger *slog.Logger) *HTTPHandler {
	return &HTTPHandler{
		manager: manager,
		logger:  logger,
	}
}

// RegisterRoutes registers all HTTP routes
func (h *HTTPHandler) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/v1/canary").Subrouter()

	// Deployment management
	api.HandleFunc("/deployments", h.listDeployments).Methods("GET")
	api.HandleFunc("/deployments", h.createDeployment).Methods("POST")
	api.HandleFunc("/deployments/{id}", h.getDeployment).Methods("GET")
	api.HandleFunc("/deployments/{id}", h.deleteDeployment).Methods("DELETE")
	api.HandleFunc("/deployments/{id}/percentage", h.updatePercentage).Methods("PUT")
	api.HandleFunc("/deployments/{id}/promote", h.promoteDeployment).Methods("POST")
	api.HandleFunc("/deployments/{id}/rollback", h.rollbackDeployment).Methods("POST")

	// Health and monitoring
	api.HandleFunc("/deployments/{id}/health", h.getDeploymentHealth).Methods("GET")
	api.HandleFunc("/deployments/{id}/metrics", h.getDeploymentMetrics).Methods("GET")
	api.HandleFunc("/deployments/{id}/events", h.getDeploymentEvents).Methods("GET")

	// Worker management
	api.HandleFunc("/workers", h.listWorkers).Methods("GET")
	api.HandleFunc("/workers", h.registerWorker).Methods("POST")
	api.HandleFunc("/workers/{id}", h.getWorker).Methods("GET")
	api.HandleFunc("/workers/{id}/status", h.updateWorkerStatus).Methods("PUT")

	// Configuration
	api.HandleFunc("/config/profiles", h.getConfigProfiles).Methods("GET")
}

// Deployment endpoints

func (h *HTTPHandler) listDeployments(w http.ResponseWriter, r *http.Request) {
	deployments, err := h.manager.ListDeployments(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}

	response := ListDeploymentsResponse{
		Deployments: deployments,
		Count:       len(deployments),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *HTTPHandler) createDeployment(w http.ResponseWriter, r *http.Request) {
	var req CreateDeploymentRequest
	if err := h.readJSON(r, &req); err != nil {
		h.writeError(w, NewValidationError("request_body", err.Error()))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.writeError(w, err)
		return
	}

	// Convert request to config
	config := req.ToCanaryConfig()

	// Create deployment
	deployment, err := h.manager.CreateDeployment(r.Context(), config)
	if err != nil {
		h.writeError(w, err)
		return
	}

	// Set additional fields from request
	deployment.QueueName = req.QueueName
	deployment.TenantID = req.TenantID
	deployment.StableVersion = req.StableVersion
	deployment.CanaryVersion = req.CanaryVersion
	deployment.CreatedBy = req.CreatedBy

	h.writeJSON(w, http.StatusCreated, deployment)
}

func (h *HTTPHandler) getDeployment(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	deployment, err := h.manager.GetDeployment(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, deployment)
}

func (h *HTTPHandler) deleteDeployment(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := h.manager.DeleteDeployment(r.Context(), id); err != nil {
		h.writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *HTTPHandler) updatePercentage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req UpdatePercentageRequest
	if err := h.readJSON(r, &req); err != nil {
		h.writeError(w, NewValidationError("request_body", err.Error()))
		return
	}

	if req.Percentage < 0 || req.Percentage > 100 {
		h.writeError(w, NewInvalidPercentageError(req.Percentage))
		return
	}

	if err := h.manager.UpdateDeploymentPercentage(r.Context(), id, req.Percentage); err != nil {
		h.writeError(w, err)
		return
	}

	response := UpdatePercentageResponse{
		Success:    true,
		Percentage: req.Percentage,
		UpdatedAt:  time.Now(),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *HTTPHandler) promoteDeployment(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if err := h.manager.PromoteDeployment(r.Context(), id); err != nil {
		h.writeError(w, err)
		return
	}

	response := PromoteResponse{
		Success:   true,
		Message:   "Deployment promoted to 100%",
		Timestamp: time.Now(),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *HTTPHandler) rollbackDeployment(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req RollbackRequest
	if err := h.readJSON(r, &req); err != nil {
		req.Reason = "Manual rollback" // Default reason
	}

	if err := h.manager.RollbackDeployment(r.Context(), id, req.Reason); err != nil {
		h.writeError(w, err)
		return
	}

	response := RollbackResponse{
		Success:   true,
		Message:   fmt.Sprintf("Deployment rolled back: %s", req.Reason),
		Reason:    req.Reason,
		Timestamp: time.Now(),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// Health and monitoring endpoints

func (h *HTTPHandler) getDeploymentHealth(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	health, err := h.manager.GetDeploymentHealth(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusOK, health)
}

func (h *HTTPHandler) getDeploymentMetrics(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	stableMetrics, canaryMetrics, err := h.manager.GetDeploymentMetrics(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response := MetricsResponse{
		Stable: stableMetrics,
		Canary: canaryMetrics,
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *HTTPHandler) getDeploymentEvents(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	events, err := h.manager.GetDeploymentEvents(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	response := EventsResponse{
		Events: events,
		Count:  len(events),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// Worker endpoints

func (h *HTTPHandler) listWorkers(w http.ResponseWriter, r *http.Request) {
	lane := r.URL.Query().Get("lane")

	var workers []*WorkerInfo
	var err error

	if lane != "" {
		workers, err = h.manager.GetWorkers(r.Context(), lane)
	} else {
		// Get all workers
		stableWorkers, _ := h.manager.GetWorkers(r.Context(), "stable")
		canaryWorkers, _ := h.manager.GetWorkers(r.Context(), "canary")
		workers = append(stableWorkers, canaryWorkers...)
	}

	if err != nil {
		h.writeError(w, err)
		return
	}

	response := WorkersResponse{
		Workers: workers,
		Count:   len(workers),
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *HTTPHandler) registerWorker(w http.ResponseWriter, r *http.Request) {
	var req RegisterWorkerRequest
	if err := h.readJSON(r, &req); err != nil {
		h.writeError(w, NewValidationError("request_body", err.Error()))
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		h.writeError(w, err)
		return
	}

	// Convert to WorkerInfo
	worker := req.ToWorkerInfo()

	if err := h.manager.RegisterWorker(r.Context(), worker); err != nil {
		h.writeError(w, err)
		return
	}

	h.writeJSON(w, http.StatusCreated, worker)
}

func (h *HTTPHandler) getWorker(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	// This would need to be implemented in the manager interface
	h.writeError(w, NewCanaryError(CodeSystemNotReady, "worker lookup not implemented"))
}

func (h *HTTPHandler) updateWorkerStatus(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req UpdateWorkerStatusRequest
	if err := h.readJSON(r, &req); err != nil {
		h.writeError(w, NewValidationError("request_body", err.Error()))
		return
	}

	if err := h.manager.UpdateWorkerStatus(r.Context(), id, req.Status); err != nil {
		h.writeError(w, err)
		return
	}

	response := UpdateWorkerStatusResponse{
		Success:   true,
		Status:    req.Status,
		UpdatedAt: time.Now(),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// Configuration endpoints

func (h *HTTPHandler) getConfigProfiles(w http.ResponseWriter, r *http.Request) {
	profiles := map[string]*CanaryConfig{
		"default":      DefaultCanaryConfig(),
		"conservative": ConservativeCanaryConfig(),
		"aggressive":   AggressiveCanaryConfig(),
	}

	h.writeJSON(w, http.StatusOK, profiles)
}

// Helper methods

func (h *HTTPHandler) readJSON(r *http.Request, v interface{}) error {
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func (h *HTTPHandler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
	}
}

func (h *HTTPHandler) writeError(w http.ResponseWriter, err error) {
	var status int
	var response *ErrorResponse

	if canaryErr := GetCanaryError(err); canaryErr != nil {
		switch canaryErr.Code {
		case CodeDeploymentNotFound, CodeWorkerNotFound, CodeQueueNotFound:
			status = http.StatusNotFound
		case CodeValidationFailed, CodeInvalidPercentage, CodeInvalidConfiguration:
			status = http.StatusBadRequest
		case CodeDeploymentExists:
			status = http.StatusConflict
		case CodeConcurrencyLimit:
			status = http.StatusTooManyRequests
		default:
			status = http.StatusInternalServerError
		}
		response = canaryErr.ToErrorResponse("")
	} else {
		status = http.StatusInternalServerError
		response = &ErrorResponse{
			Error: err.Error(),
			Code:  "INTERNAL_ERROR",
		}
	}

	h.writeJSON(w, status, response)
}

// Request/Response types

type CreateDeploymentRequest struct {
	QueueName       string                `json:"queue_name"`
	TenantID        string                `json:"tenant_id,omitempty"`
	StableVersion   string                `json:"stable_version"`
	CanaryVersion   string                `json:"canary_version"`
	RoutingStrategy RoutingStrategy       `json:"routing_strategy,omitempty"`
	StickyRouting   bool                  `json:"sticky_routing,omitempty"`
	AutoPromotion   bool                  `json:"auto_promotion,omitempty"`
	MaxDuration     string                `json:"max_duration,omitempty"`
	MinDuration     string                `json:"min_duration,omitempty"`
	DrainTimeout    string                `json:"drain_timeout,omitempty"`
	MetricsWindow   string                `json:"metrics_window,omitempty"`
	CreatedBy       string                `json:"created_by,omitempty"`
	Profile         string                `json:"profile,omitempty"` // "default", "conservative", "aggressive"
}

func (req *CreateDeploymentRequest) Validate() error {
	if req.QueueName == "" {
		return NewValidationError("queue_name", "queue name is required")
	}
	if req.StableVersion == "" {
		return NewValidationError("stable_version", "stable version is required")
	}
	if req.CanaryVersion == "" {
		return NewValidationError("canary_version", "canary version is required")
	}
	return nil
}

func (req *CreateDeploymentRequest) ToCanaryConfig() *CanaryConfig {
	var config *CanaryConfig

	// Use profile if specified
	if req.Profile != "" {
		config = GetConfigByProfile(req.Profile)
	} else {
		config = DefaultCanaryConfig()
	}

	// Override with request values
	if req.RoutingStrategy != "" {
		config.RoutingStrategy = req.RoutingStrategy
	}
	if req.StickyRouting {
		config.StickyRouting = req.StickyRouting
	}
	if req.AutoPromotion {
		config.AutoPromotion = req.AutoPromotion
	}

	// Parse durations
	if req.MaxDuration != "" {
		if duration, err := time.ParseDuration(req.MaxDuration); err == nil {
			config.MaxCanaryDuration = duration
		}
	}
	if req.MinDuration != "" {
		if duration, err := time.ParseDuration(req.MinDuration); err == nil {
			config.MinCanaryDuration = duration
		}
	}
	if req.DrainTimeout != "" {
		if duration, err := time.ParseDuration(req.DrainTimeout); err == nil {
			config.DrainTimeout = duration
		}
	}
	if req.MetricsWindow != "" {
		if duration, err := time.ParseDuration(req.MetricsWindow); err == nil {
			config.MetricsWindow = duration
		}
	}

	return config
}

type UpdatePercentageRequest struct {
	Percentage int `json:"percentage"`
}

type RollbackRequest struct {
	Reason string `json:"reason"`
}

type RegisterWorkerRequest struct {
	ID       string            `json:"id"`
	Version  string            `json:"version"`
	Lane     string            `json:"lane,omitempty"`
	Queues   []string          `json:"queues"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

func (req *RegisterWorkerRequest) Validate() error {
	if req.ID == "" {
		return NewValidationError("id", "worker ID is required")
	}
	if req.Version == "" {
		return NewValidationError("version", "worker version is required")
	}
	if len(req.Queues) == 0 {
		return NewValidationError("queues", "at least one queue is required")
	}
	return nil
}

func (req *RegisterWorkerRequest) ToWorkerInfo() *WorkerInfo {
	return &WorkerInfo{
		ID:       req.ID,
		Version:  req.Version,
		Lane:     req.Lane,
		Queues:   req.Queues,
		Metadata: req.Metadata,
	}
}

type UpdateWorkerStatusRequest struct {
	Status WorkerStatus `json:"status"`
}

// Response types

type ListDeploymentsResponse struct {
	Deployments []*CanaryDeployment `json:"deployments"`
	Count       int                 `json:"count"`
}

type UpdatePercentageResponse struct {
	Success    bool      `json:"success"`
	Percentage int       `json:"percentage"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type PromoteResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

type RollbackResponse struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message"`
	Reason    string    `json:"reason"`
	Timestamp time.Time `json:"timestamp"`
}

type MetricsResponse struct {
	Stable *MetricsSnapshot `json:"stable"`
	Canary *MetricsSnapshot `json:"canary"`
}

type EventsResponse struct {
	Events []*DeploymentEvent `json:"events"`
	Count  int                `json:"count"`
}

type WorkersResponse struct {
	Workers []*WorkerInfo `json:"workers"`
	Count   int           `json:"count"`
}

type UpdateWorkerStatusResponse struct {
	Success   bool         `json:"success"`
	Status    WorkerStatus `json:"status"`
	UpdatedAt time.Time    `json:"updated_at"`
}

// Middleware for authentication and logging

func (h *HTTPHandler) WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Simple token-based auth
		token := r.Header.Get("Authorization")
		if token == "" {
			h.writeError(w, NewCanaryError("UNAUTHORIZED", "missing authorization header"))
			return
		}

		// Remove "Bearer " prefix if present
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		// Here you would validate the token
		// For this implementation, we'll skip actual validation

		next(w, r)
	}
}

func (h *HTTPHandler) WithLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(wrapper, r)

		duration := time.Since(start)

		h.logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapper.statusCode,
			"duration", duration,
			"remote_addr", r.RemoteAddr)
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// StartHTTPServer starts the HTTP server with all routes configured
func StartHTTPServer(manager CanaryManager, config *Config, logger *slog.Logger) error {
	if !config.EnableAPI {
		logger.Info("API server disabled")
		return nil
	}

	handler := NewHTTPHandler(manager, logger)
	router := mux.NewRouter()

	// Register routes
	handler.RegisterRoutes(router)

	// Add middleware
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	server := &http.Server{
		Addr:         config.APIListenAddr,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Info("Starting HTTP API server", "addr", config.APIListenAddr)
	return server.ListenAndServe()
}