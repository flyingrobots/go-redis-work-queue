package workerfleetcontrols

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type APIHandlers struct {
	manager WorkerFleetManager
	logger  *slog.Logger
}

func NewAPIHandlers(manager WorkerFleetManager, logger *slog.Logger) *APIHandlers {
	return &APIHandlers{
		manager: manager,
		logger:  logger,
	}
}

func (h *APIHandlers) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/workers").Subrouter()

	api.HandleFunc("", h.ListWorkers).Methods("GET")
	api.HandleFunc("/{id}", h.GetWorker).Methods("GET")
	api.HandleFunc("/{id}/heartbeat", h.UpdateHeartbeat).Methods("POST")
	api.HandleFunc("/register", h.RegisterWorker).Methods("POST")
	api.HandleFunc("/actions/pause", h.PauseWorkers).Methods("POST")
	api.HandleFunc("/actions/resume", h.ResumeWorkers).Methods("POST")
	api.HandleFunc("/actions/drain", h.DrainWorkers).Methods("POST")
	api.HandleFunc("/actions/stop", h.StopWorkers).Methods("POST")
	api.HandleFunc("/actions/restart", h.RestartWorkers).Methods("POST")
	api.HandleFunc("/actions/rolling-restart", h.RollingRestart).Methods("POST")
	api.HandleFunc("/actions/{request_id}", h.GetActionStatus).Methods("GET")
	api.HandleFunc("/actions/{request_id}/cancel", h.CancelAction).Methods("POST")
	api.HandleFunc("/summary", h.GetFleetSummary).Methods("GET")
	api.HandleFunc("/audit", h.GetAuditLogs).Methods("GET")

	api.Use(h.loggingMiddleware)
	api.Use(h.corsMiddleware)
}

func (h *APIHandlers) ListWorkers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	request, err := h.parseWorkerListRequest(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request parameters", err)
		return
	}

	response, err := h.manager.Registry().ListWorkers(request)
	if err != nil {
		h.logger.Error("Failed to list workers", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to list workers", err)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) GetWorker(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workerID := vars["id"]

	if workerID == "" {
		h.writeError(w, http.StatusBadRequest, "Worker ID is required", nil)
		return
	}

	worker, err := h.manager.Registry().GetWorker(workerID)
	if err != nil {
		h.logger.Error("Failed to get worker", "worker_id", workerID, "error", err)
		h.writeError(w, http.StatusNotFound, "Worker not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, worker)
}

func (h *APIHandlers) RegisterWorker(w http.ResponseWriter, r *http.Request) {
	var worker Worker
	if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if worker.ID == "" {
		h.writeError(w, http.StatusBadRequest, "Worker ID is required", nil)
		return
	}

	err := h.manager.Registry().RegisterWorker(&worker)
	if err != nil {
		h.logger.Error("Failed to register worker", "worker_id", worker.ID, "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to register worker", err)
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success":   true,
		"worker_id": worker.ID,
		"message":   "Worker registered successfully",
	})
}

func (h *APIHandlers) UpdateHeartbeat(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	workerID := vars["id"]

	var request struct {
		Timestamp  time.Time  `json:"timestamp"`
		CurrentJob *ActiveJob `json:"current_job,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if request.Timestamp.IsZero() {
		request.Timestamp = time.Now()
	}

	err := h.manager.Registry().UpdateHeartbeat(workerID, request.Timestamp, request.CurrentJob)
	if err != nil {
		h.logger.Error("Failed to update heartbeat", "worker_id", workerID, "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to update heartbeat", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Heartbeat updated successfully",
	})
}

func (h *APIHandlers) PauseWorkers(w http.ResponseWriter, r *http.Request) {
	var request struct {
		WorkerIDs []string `json:"worker_ids"`
		Reason    string   `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if len(request.WorkerIDs) == 0 {
		h.writeError(w, http.StatusBadRequest, "No worker IDs provided", nil)
		return
	}

	response, err := h.manager.Controller().PauseWorkers(request.WorkerIDs, request.Reason)
	if err != nil {
		h.logger.Error("Failed to pause workers", "worker_ids", request.WorkerIDs, "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to pause workers", err)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) ResumeWorkers(w http.ResponseWriter, r *http.Request) {
	var request struct {
		WorkerIDs []string `json:"worker_ids"`
		Reason    string   `json:"reason,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if len(request.WorkerIDs) == 0 {
		h.writeError(w, http.StatusBadRequest, "No worker IDs provided", nil)
		return
	}

	response, err := h.manager.Controller().ResumeWorkers(request.WorkerIDs, request.Reason)
	if err != nil {
		h.logger.Error("Failed to resume workers", "worker_ids", request.WorkerIDs, "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to resume workers", err)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) DrainWorkers(w http.ResponseWriter, r *http.Request) {
	var request struct {
		WorkerIDs    []string `json:"worker_ids"`
		Reason       string   `json:"reason,omitempty"`
		TimeoutSecs  int      `json:"timeout_seconds,omitempty"`
		Confirmation string   `json:"confirmation,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if len(request.WorkerIDs) == 0 {
		h.writeError(w, http.StatusBadRequest, "No worker IDs provided", nil)
		return
	}

	if h.manager.SafetyChecker().RequiresConfirmation(WorkerActionDrain, request.WorkerIDs) {
		if err := h.manager.SafetyChecker().ValidateConfirmation(WorkerActionDrain, request.WorkerIDs, request.Confirmation); err != nil {
			prompt := h.manager.SafetyChecker().GenerateConfirmationPrompt(WorkerActionDrain, request.WorkerIDs)
			h.writeJSON(w, http.StatusPreconditionRequired, map[string]interface{}{
				"confirmation_required": true,
				"prompt":               prompt,
				"error":                err.Error(),
			})
			return
		}
	}

	timeout := time.Duration(request.TimeoutSecs) * time.Second
	if timeout == 0 {
		timeout = 5 * time.Minute
	}

	response, err := h.manager.Controller().DrainWorkers(request.WorkerIDs, timeout, request.Reason)
	if err != nil {
		h.logger.Error("Failed to drain workers", "worker_ids", request.WorkerIDs, "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to drain workers", err)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) StopWorkers(w http.ResponseWriter, r *http.Request) {
	var request struct {
		WorkerIDs    []string `json:"worker_ids"`
		Reason       string   `json:"reason,omitempty"`
		Force        bool     `json:"force,omitempty"`
		Confirmation string   `json:"confirmation,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if len(request.WorkerIDs) == 0 {
		h.writeError(w, http.StatusBadRequest, "No worker IDs provided", nil)
		return
	}

	if h.manager.SafetyChecker().RequiresConfirmation(WorkerActionStop, request.WorkerIDs) {
		if err := h.manager.SafetyChecker().ValidateConfirmation(WorkerActionStop, request.WorkerIDs, request.Confirmation); err != nil {
			prompt := h.manager.SafetyChecker().GenerateConfirmationPrompt(WorkerActionStop, request.WorkerIDs)
			h.writeJSON(w, http.StatusPreconditionRequired, map[string]interface{}{
				"confirmation_required": true,
				"prompt":               prompt,
				"error":                err.Error(),
			})
			return
		}
	}

	response, err := h.manager.Controller().StopWorkers(request.WorkerIDs, request.Force, request.Reason)
	if err != nil {
		h.logger.Error("Failed to stop workers", "worker_ids", request.WorkerIDs, "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to stop workers", err)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) RestartWorkers(w http.ResponseWriter, r *http.Request) {
	var request struct {
		WorkerIDs    []string `json:"worker_ids"`
		Reason       string   `json:"reason,omitempty"`
		Confirmation string   `json:"confirmation,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if len(request.WorkerIDs) == 0 {
		h.writeError(w, http.StatusBadRequest, "No worker IDs provided", nil)
		return
	}

	if h.manager.SafetyChecker().RequiresConfirmation(WorkerActionRestart, request.WorkerIDs) {
		if err := h.manager.SafetyChecker().ValidateConfirmation(WorkerActionRestart, request.WorkerIDs, request.Confirmation); err != nil {
			prompt := h.manager.SafetyChecker().GenerateConfirmationPrompt(WorkerActionRestart, request.WorkerIDs)
			h.writeJSON(w, http.StatusPreconditionRequired, map[string]interface{}{
				"confirmation_required": true,
				"prompt":               prompt,
				"error":                err.Error(),
			})
			return
		}
	}

	response, err := h.manager.Controller().RestartWorkers(request.WorkerIDs, request.Reason)
	if err != nil {
		h.logger.Error("Failed to restart workers", "worker_ids", request.WorkerIDs, "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to restart workers", err)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) RollingRestart(w http.ResponseWriter, r *http.Request) {
	var request RollingRestartRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if request.Concurrency <= 0 {
		request.Concurrency = 1
	}

	if request.DrainTimeout == 0 {
		request.DrainTimeout = 5 * time.Minute
	}

	if request.RestartTimeout == 0 {
		request.RestartTimeout = 2 * time.Minute
	}

	response, err := h.manager.Controller().RollingRestart(request)
	if err != nil {
		h.logger.Error("Failed to start rolling restart", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to start rolling restart", err)
		return
	}

	h.writeJSON(w, http.StatusAccepted, response)
}

func (h *APIHandlers) GetActionStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["request_id"]

	if requestID == "" {
		h.writeError(w, http.StatusBadRequest, "Request ID is required", nil)
		return
	}

	response, err := h.manager.Controller().GetActionStatus(requestID)
	if err != nil {
		h.logger.Error("Failed to get action status", "request_id", requestID, "error", err)
		h.writeError(w, http.StatusNotFound, "Action not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) CancelAction(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["request_id"]

	if requestID == "" {
		h.writeError(w, http.StatusBadRequest, "Request ID is required", nil)
		return
	}

	err := h.manager.Controller().CancelAction(requestID)
	if err != nil {
		h.logger.Error("Failed to cancel action", "request_id", requestID, "error", err)
		h.writeError(w, http.StatusNotFound, "Action not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Action cancelled successfully",
	})
}

func (h *APIHandlers) GetFleetSummary(w http.ResponseWriter, r *http.Request) {
	summary, err := h.manager.Registry().GetFleetSummary()
	if err != nil {
		h.logger.Error("Failed to get fleet summary", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get fleet summary", err)
		return
	}

	h.writeJSON(w, http.StatusOK, summary)
}

func (h *APIHandlers) GetAuditLogs(w http.ResponseWriter, r *http.Request) {
	filter, err := h.parseAuditLogFilter(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	logs, err := h.manager.AuditLogger().GetAuditLogs(filter)
	if err != nil {
		h.logger.Error("Failed to get audit logs", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get audit logs", err)
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"audit_logs": logs,
		"filter":     filter,
		"count":      len(logs),
	})
}

func (h *APIHandlers) parseWorkerListRequest(r *http.Request) (WorkerListRequest, error) {
	query := r.URL.Query()

	request := WorkerListRequest{
		Pagination: Pagination{
			Page:     1,
			PageSize: 50,
		},
		SortBy:    "last_heartbeat",
		SortOrder: SortOrderDesc,
	}

	if page := query.Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			request.Pagination.Page = p
		}
	}

	if pageSize := query.Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 1000 {
			request.Pagination.PageSize = ps
		}
	}

	if sortBy := query.Get("sort_by"); sortBy != "" {
		request.SortBy = sortBy
	}

	if sortOrder := query.Get("sort_order"); sortOrder != "" {
		if sortOrder == "asc" {
			request.SortOrder = SortOrderAsc
		} else {
			request.SortOrder = SortOrderDesc
		}
	}

	filter := WorkerFilter{}

	if states := query.Get("states"); states != "" {
		stateStrings := strings.Split(states, ",")
		for _, state := range stateStrings {
			filter.States = append(filter.States, WorkerState(strings.TrimSpace(state)))
		}
	}

	if hostname := query.Get("hostname"); hostname != "" {
		filter.Hostname = hostname
	}

	if version := query.Get("version"); version != "" {
		filter.Version = version
	}

	request.Filter = filter
	return request, nil
}

func (h *APIHandlers) parseAuditLogFilter(r *http.Request) (AuditLogFilter, error) {
	query := r.URL.Query()

	filter := AuditLogFilter{
		Limit: 100,
	}

	if limit := query.Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 && l <= 1000 {
			filter.Limit = l
		}
	}

	if offset := query.Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			filter.Offset = o
		}
	}

	if startTime := query.Get("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			filter.StartTime = &t
		}
	}

	if endTime := query.Get("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			filter.EndTime = &t
		}
	}

	if actions := query.Get("actions"); actions != "" {
		actionStrings := strings.Split(actions, ",")
		for _, action := range actionStrings {
			filter.Actions = append(filter.Actions, WorkerAction(strings.TrimSpace(action)))
		}
	}

	if workerIDs := query.Get("worker_ids"); workerIDs != "" {
		filter.WorkerIDs = strings.Split(workerIDs, ",")
	}

	return filter, nil
}

func (h *APIHandlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
	}
}

func (h *APIHandlers) writeError(w http.ResponseWriter, status int, message string, err error) {
	response := map[string]interface{}{
		"error":     message,
		"timestamp": time.Now().Format(time.RFC3339),
	}

	if err != nil {
		response["details"] = err.Error()
	}

	h.writeJSON(w, status, response)
}

func (h *APIHandlers) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		h.logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration", duration,
			"remote_addr", r.RemoteAddr,
		)
	})
}

func (h *APIHandlers) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}