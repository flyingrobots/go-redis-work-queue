// Copyright 2025 James Ross
package multicluster

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// HTTPHandler provides HTTP handlers for multi-cluster operations
type HTTPHandler struct {
	manager Manager
	logger  *zap.Logger
}

// NewHTTPHandler creates a new HTTP handler
func NewHTTPHandler(manager Manager, logger *zap.Logger) *HTTPHandler {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &HTTPHandler{
		manager: manager,
		logger:  logger,
	}
}

// RegisterRoutes registers HTTP routes
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	// Cluster management
	mux.HandleFunc("/api/v1/clusters", h.handleClusters)
	mux.HandleFunc("/api/v1/clusters/", h.handleCluster)

	// Stats and monitoring
	mux.HandleFunc("/api/v1/stats", h.handleStats)
	mux.HandleFunc("/api/v1/stats/compare", h.handleCompare)
	mux.HandleFunc("/api/v1/health", h.handleHealth)

	// Actions
	mux.HandleFunc("/api/v1/actions", h.handleActions)
	mux.HandleFunc("/api/v1/actions/", h.handleAction)

	// Events
	mux.HandleFunc("/api/v1/events", h.handleEvents)

	// UI support
	mux.HandleFunc("/api/v1/ui/tabs", h.handleTabs)
	mux.HandleFunc("/api/v1/ui/compare", h.handleCompareMode)
}

// handleClusters handles cluster list and creation
func (h *HTTPHandler) handleClusters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		h.listClusters(w, r, ctx)
	case http.MethodPost:
		h.createCluster(w, r, ctx)
	default:
		h.methodNotAllowed(w, r)
	}
}

// handleCluster handles individual cluster operations
func (h *HTTPHandler) handleCluster(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract cluster name from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/clusters/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		h.badRequest(w, "cluster name required")
		return
	}

	clusterName := parts[0]

	switch r.Method {
	case http.MethodGet:
		h.getCluster(w, r, ctx, clusterName)
	case http.MethodPut:
		h.updateCluster(w, r, ctx, clusterName)
	case http.MethodDelete:
		h.deleteCluster(w, r, ctx, clusterName)
	default:
		h.methodNotAllowed(w, r)
	}
}

// listClusters returns the list of clusters
func (h *HTTPHandler) listClusters(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	clusters, err := h.manager.ListClusters(ctx)
	if err != nil {
		h.internalError(w, err)
		return
	}

	h.jsonResponse(w, http.StatusOK, clusters)
}

// createCluster creates a new cluster
func (h *HTTPHandler) createCluster(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var cfg ClusterConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		h.badRequest(w, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if err := h.manager.AddCluster(ctx, cfg); err != nil {
		if err == ErrClusterAlreadyExists {
			h.conflict(w, err.Error())
			return
		}
		h.internalError(w, err)
		return
	}

	h.jsonResponse(w, http.StatusCreated, map[string]string{
		"message": fmt.Sprintf("cluster %s created", cfg.Name),
	})
}

// getCluster returns details for a specific cluster
func (h *HTTPHandler) getCluster(w http.ResponseWriter, r *http.Request, ctx context.Context, name string) {
	conn, err := h.manager.GetCluster(ctx, name)
	if err != nil {
		if err == ErrClusterNotFound {
			h.notFound(w, err.Error())
			return
		}
		h.internalError(w, err)
		return
	}

	response := map[string]interface{}{
		"config": conn.Config,
		"status": conn.Status,
	}

	h.jsonResponse(w, http.StatusOK, response)
}

// updateCluster updates a cluster configuration
func (h *HTTPHandler) updateCluster(w http.ResponseWriter, r *http.Request, ctx context.Context, name string) {
	var cfg ClusterConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		h.badRequest(w, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Remove and re-add the cluster with new config
	if err := h.manager.RemoveCluster(ctx, name); err != nil {
		if err == ErrClusterNotFound {
			h.notFound(w, err.Error())
			return
		}
		h.internalError(w, err)
		return
	}

	if err := h.manager.AddCluster(ctx, cfg); err != nil {
		h.internalError(w, err)
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("cluster %s updated", name),
	})
}

// deleteCluster deletes a cluster
func (h *HTTPHandler) deleteCluster(w http.ResponseWriter, r *http.Request, ctx context.Context, name string) {
	if err := h.manager.RemoveCluster(ctx, name); err != nil {
		if err == ErrClusterNotFound {
			h.notFound(w, err.Error())
			return
		}
		h.internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleStats handles statistics requests
func (h *HTTPHandler) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.methodNotAllowed(w, r)
		return
	}

	ctx := r.Context()
	clusterName := r.URL.Query().Get("cluster")

	if clusterName != "" {
		// Get stats for specific cluster
		stats, err := h.manager.GetStats(ctx, clusterName)
		if err != nil {
			if err == ErrClusterNotFound {
				h.notFound(w, err.Error())
				return
			}
			h.internalError(w, err)
			return
		}
		h.jsonResponse(w, http.StatusOK, stats)
	} else {
		// Get stats for all clusters
		stats, err := h.manager.GetAllStats(ctx)
		if err != nil {
			h.internalError(w, err)
			return
		}
		h.jsonResponse(w, http.StatusOK, stats)
	}
}

// handleCompare handles cluster comparison requests
func (h *HTTPHandler) handleCompare(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.methodNotAllowed(w, r)
		return
	}

	ctx := r.Context()

	var request struct {
		Clusters []string `json:"clusters"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.badRequest(w, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	result, err := h.manager.CompareClusters(ctx, request.Clusters)
	if err != nil {
		if err == ErrInsufficientClusters {
			h.badRequest(w, err.Error())
			return
		}
		h.internalError(w, err)
		return
	}

	h.jsonResponse(w, http.StatusOK, result)
}

// handleHealth handles health check requests
func (h *HTTPHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.methodNotAllowed(w, r)
		return
	}

	ctx := r.Context()
	clusterName := r.URL.Query().Get("cluster")

	if clusterName == "" {
		// Overall health check
		clusters, err := h.manager.ListClusters(ctx)
		if err != nil {
			h.internalError(w, err)
			return
		}

		healthMap := make(map[string]*HealthStatus)
		allHealthy := true

		for _, cluster := range clusters {
			if !cluster.Enabled {
				continue
			}
			health, err := h.manager.GetHealth(ctx, cluster.Name)
			if err != nil {
				h.logger.Warn("Failed to get health",
					zap.String("cluster", cluster.Name),
					zap.Error(err))
				allHealthy = false
				continue
			}
			healthMap[cluster.Name] = health
			if !health.Healthy {
				allHealthy = false
			}
		}

		response := map[string]interface{}{
			"healthy":  allHealthy,
			"clusters": healthMap,
		}

		status := http.StatusOK
		if !allHealthy {
			status = http.StatusServiceUnavailable
		}

		h.jsonResponse(w, status, response)
	} else {
		// Health check for specific cluster
		health, err := h.manager.GetHealth(ctx, clusterName)
		if err != nil {
			if err == ErrClusterNotFound {
				h.notFound(w, err.Error())
				return
			}
			h.internalError(w, err)
			return
		}

		status := http.StatusOK
		if !health.Healthy {
			status = http.StatusServiceUnavailable
		}

		h.jsonResponse(w, status, health)
	}
}

// handleActions handles action requests
func (h *HTTPHandler) handleActions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodPost:
		h.createAction(w, r, ctx)
	default:
		h.methodNotAllowed(w, r)
	}
}

// handleAction handles individual action operations
func (h *HTTPHandler) handleAction(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract action ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/actions/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		h.badRequest(w, "action ID required")
		return
	}

	actionID := parts[0]

	// Check for sub-resources
	if len(parts) > 1 {
		switch parts[1] {
		case "confirm":
			if r.Method == http.MethodPost {
				h.confirmAction(w, r, ctx, actionID)
				return
			}
		case "cancel":
			if r.Method == http.MethodPost {
				h.cancelAction(w, r, ctx, actionID)
				return
			}
		}
	}

	switch r.Method {
	case http.MethodGet:
		h.getAction(w, r, ctx, actionID)
	default:
		h.methodNotAllowed(w, r)
	}
}

// createAction creates and executes a new action
func (h *HTTPHandler) createAction(w http.ResponseWriter, r *http.Request, ctx context.Context) {
	var action MultiAction
	if err := json.NewDecoder(r.Body).Decode(&action); err != nil {
		h.badRequest(w, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	// Generate ID if not provided
	if action.ID == "" {
		action.ID = fmt.Sprintf("action-%d", time.Now().UnixNano())
	}

	action.CreatedAt = time.Now()
	action.Status = ActionStatusPending

	// Execute the action
	if err := h.manager.ExecuteAction(ctx, &action); err != nil {
		if err == ErrActionNotAllowed {
			h.forbidden(w, err.Error())
			return
		}
		if err == ErrConfirmationRequired {
			// Return action for confirmation
			h.jsonResponse(w, http.StatusAccepted, action)
			return
		}
		h.internalError(w, err)
		return
	}

	h.jsonResponse(w, http.StatusCreated, action)
}

// getAction gets action status
func (h *HTTPHandler) getAction(w http.ResponseWriter, r *http.Request, ctx context.Context, actionID string) {
	action, err := h.manager.GetActionStatus(ctx, actionID)
	if err != nil {
		h.notFound(w, err.Error())
		return
	}

	h.jsonResponse(w, http.StatusOK, action)
}

// confirmAction confirms a pending action
func (h *HTTPHandler) confirmAction(w http.ResponseWriter, r *http.Request, ctx context.Context, actionID string) {
	var request struct {
		ConfirmedBy string `json:"confirmed_by"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.badRequest(w, fmt.Sprintf("invalid request body: %v", err))
		return
	}

	if err := h.manager.ConfirmAction(ctx, actionID, request.ConfirmedBy); err != nil {
		h.internalError(w, err)
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("action %s confirmed", actionID),
	})
}

// cancelAction cancels a pending action
func (h *HTTPHandler) cancelAction(w http.ResponseWriter, r *http.Request, ctx context.Context, actionID string) {
	if err := h.manager.CancelAction(ctx, actionID); err != nil {
		h.internalError(w, err)
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]string{
		"message": fmt.Sprintf("action %s cancelled", actionID),
	})
}

// handleEvents handles event streaming
func (h *HTTPHandler) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.methodNotAllowed(w, r)
		return
	}

	ctx := r.Context()

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Subscribe to events
	events, err := h.manager.SubscribeEvents(ctx)
	if err != nil {
		h.internalError(w, err)
		return
	}

	// Ensure cleanup on disconnect
	defer h.manager.UnsubscribeEvents(ctx, events)

	// Create a flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.internalError(w, fmt.Errorf("streaming not supported"))
		return
	}

	// Send events
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}

			data, err := json.Marshal(event)
			if err != nil {
				h.logger.Error("Failed to marshal event", zap.Error(err))
				continue
			}

			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// handleTabs handles tab configuration requests
func (h *HTTPHandler) handleTabs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.methodNotAllowed(w, r)
		return
	}

	ctx := r.Context()

	tabConfig, err := h.manager.GetTabConfig(ctx)
	if err != nil {
		h.internalError(w, err)
		return
	}

	h.jsonResponse(w, http.StatusOK, tabConfig)
}

// handleCompareMode handles compare mode configuration
func (h *HTTPHandler) handleCompareMode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	switch r.Method {
	case http.MethodGet:
		tabConfig, err := h.manager.GetTabConfig(ctx)
		if err != nil {
			h.internalError(w, err)
			return
		}

		response := map[string]interface{}{
			"enabled":  tabConfig.CompareMode,
			"clusters": tabConfig.CompareWith,
		}
		h.jsonResponse(w, http.StatusOK, response)

	case http.MethodPut:
		var request struct {
			Enabled  bool     `json:"enabled"`
			Clusters []string `json:"clusters"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			h.badRequest(w, fmt.Sprintf("invalid request body: %v", err))
			return
		}

		if err := h.manager.SetCompareMode(ctx, request.Enabled, request.Clusters); err != nil {
			if err == ErrInsufficientClusters {
				h.badRequest(w, err.Error())
				return
			}
			h.internalError(w, err)
			return
		}

		h.jsonResponse(w, http.StatusOK, map[string]string{
			"message": "compare mode updated",
		})

	default:
		h.methodNotAllowed(w, r)
	}
}

// Helper methods for HTTP responses

func (h *HTTPHandler) jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode response", zap.Error(err))
	}
}

func (h *HTTPHandler) errorResponse(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{
		"error": message,
	})
}

func (h *HTTPHandler) badRequest(w http.ResponseWriter, message string) {
	h.errorResponse(w, http.StatusBadRequest, message)
}

func (h *HTTPHandler) notFound(w http.ResponseWriter, message string) {
	h.errorResponse(w, http.StatusNotFound, message)
}

func (h *HTTPHandler) conflict(w http.ResponseWriter, message string) {
	h.errorResponse(w, http.StatusConflict, message)
}

func (h *HTTPHandler) forbidden(w http.ResponseWriter, message string) {
	h.errorResponse(w, http.StatusForbidden, message)
}

func (h *HTTPHandler) methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	h.errorResponse(w, http.StatusMethodNotAllowed, fmt.Sprintf("method %s not allowed", r.Method))
}

func (h *HTTPHandler) internalError(w http.ResponseWriter, err error) {
	h.logger.Error("Internal server error", zap.Error(err))
	h.errorResponse(w, http.StatusInternalServerError, "internal server error")
}