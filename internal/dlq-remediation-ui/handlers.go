package dlqremediationui

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type APIHandlers struct {
	manager DLQManager
	logger  *slog.Logger
}

func NewAPIHandlers(manager DLQManager, logger *slog.Logger) *APIHandlers {
	return &APIHandlers{
		manager: manager,
		logger:  logger,
	}
}

func (h *APIHandlers) RegisterRoutes(router *mux.Router) {
	api := router.PathPrefix("/api/dlq").Subrouter()

	api.HandleFunc("/entries", h.ListEntries).Methods("GET")
	api.HandleFunc("/entries/{id}", h.GetEntry).Methods("GET")
	api.HandleFunc("/entries/requeue", h.RequeueEntries).Methods("POST")
	api.HandleFunc("/entries/purge", h.PurgeEntries).Methods("POST")
	api.HandleFunc("/entries/purge-all", h.PurgeAll).Methods("POST")
	api.HandleFunc("/stats", h.GetStats).Methods("GET")

	api.Use(h.loggingMiddleware)
	api.Use(h.corsMiddleware)
}

func (h *APIHandlers) ListEntries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter, err := h.parseFilter(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	pagination, err := h.parsePagination(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid pagination parameters", err)
		return
	}

	response, err := h.manager.ListEntries(ctx, filter, pagination)
	if err != nil {
		h.logger.Error("Failed to list DLQ entries", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to list entries", err)
		return
	}

	h.writeJSON(w, http.StatusOK, response)
}

func (h *APIHandlers) GetEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Entry ID is required", nil)
		return
	}

	entry, err := h.manager.PeekEntry(ctx, id)
	if err != nil {
		h.logger.Error("Failed to get DLQ entry", "id", id, "error", err)
		h.writeError(w, http.StatusNotFound, "Entry not found", err)
		return
	}

	h.writeJSON(w, http.StatusOK, entry)
}

func (h *APIHandlers) RequeueEntries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		IDs []string `json:"ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if len(request.IDs) == 0 {
		h.writeError(w, http.StatusBadRequest, "No entry IDs provided", nil)
		return
	}

	result, err := h.manager.RequeueEntries(ctx, request.IDs)
	if err != nil {
		h.logger.Error("Failed to requeue entries", "ids", request.IDs, "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to requeue entries", err)
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

func (h *APIHandlers) PurgeEntries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var request struct {
		IDs []string `json:"ids"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	if len(request.IDs) == 0 {
		h.writeError(w, http.StatusBadRequest, "No entry IDs provided", nil)
		return
	}

	result, err := h.manager.PurgeEntries(ctx, request.IDs)
	if err != nil {
		h.logger.Error("Failed to purge entries", "ids", request.IDs, "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to purge entries", err)
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

func (h *APIHandlers) PurgeAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	filter, err := h.parseFilter(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	confirm := r.URL.Query().Get("confirm")
	if confirm != "true" {
		h.writeError(w, http.StatusBadRequest, "Purge all requires explicit confirmation", nil)
		return
	}

	result, err := h.manager.PurgeAll(ctx, filter)
	if err != nil {
		h.logger.Error("Failed to purge all entries", "filter", filter, "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to purge all entries", err)
		return
	}

	h.writeJSON(w, http.StatusOK, result)
}

func (h *APIHandlers) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	stats, err := h.manager.GetStats(ctx)
	if err != nil {
		h.logger.Error("Failed to get DLQ stats", "error", err)
		h.writeError(w, http.StatusInternalServerError, "Failed to get stats", err)
		return
	}

	h.writeJSON(w, http.StatusOK, stats)
}

func (h *APIHandlers) parseFilter(r *http.Request) (DLQFilter, error) {
	query := r.URL.Query()

	filter := DLQFilter{
		Queue:           query.Get("queue"),
		Type:            query.Get("type"),
		ErrorPattern:    query.Get("error_pattern"),
		IncludePatterns: query.Get("include_patterns") == "true",
	}

	if startTime := query.Get("start_time"); startTime != "" {
		t, err := time.Parse(time.RFC3339, startTime)
		if err != nil {
			return filter, fmt.Errorf("invalid start_time format: %w", err)
		}
		filter.StartTime = t
	}

	if endTime := query.Get("end_time"); endTime != "" {
		t, err := time.Parse(time.RFC3339, endTime)
		if err != nil {
			return filter, fmt.Errorf("invalid end_time format: %w", err)
		}
		filter.EndTime = t
	}

	if minAttempts := query.Get("min_attempts"); minAttempts != "" {
		n, err := strconv.Atoi(minAttempts)
		if err != nil {
			return filter, fmt.Errorf("invalid min_attempts: %w", err)
		}
		filter.MinAttempts = n
	}

	if maxAttempts := query.Get("max_attempts"); maxAttempts != "" {
		n, err := strconv.Atoi(maxAttempts)
		if err != nil {
			return filter, fmt.Errorf("invalid max_attempts: %w", err)
		}
		filter.MaxAttempts = n
	}

	return filter, nil
}

func (h *APIHandlers) parsePagination(r *http.Request) (PaginationRequest, error) {
	query := r.URL.Query()

	pagination := PaginationRequest{
		Page:      1,
		PageSize:  50,
		SortBy:    "failed_at",
		SortOrder: SortOrderDesc,
	}

	if page := query.Get("page"); page != "" {
		n, err := strconv.Atoi(page)
		if err != nil || n < 1 {
			return pagination, fmt.Errorf("invalid page number")
		}
		pagination.Page = n
	}

	if pageSize := query.Get("page_size"); pageSize != "" {
		n, err := strconv.Atoi(pageSize)
		if err != nil || n < 1 || n > 1000 {
			return pagination, fmt.Errorf("invalid page size")
		}
		pagination.PageSize = n
	}

	if sortBy := query.Get("sort_by"); sortBy != "" {
		validSorts := map[string]bool{
			"failed_at":   true,
			"created_at":  true,
			"queue":       true,
			"type":        true,
			"attempts":    true,
		}
		if !validSorts[sortBy] {
			return pagination, fmt.Errorf("invalid sort field: %s", sortBy)
		}
		pagination.SortBy = sortBy
	}

	if sortOrder := query.Get("sort_order"); sortOrder != "" {
		switch sortOrder {
		case "asc":
			pagination.SortOrder = SortOrderAsc
		case "desc":
			pagination.SortOrder = SortOrderDesc
		default:
			return pagination, fmt.Errorf("invalid sort order: %s", sortOrder)
		}
	}

	return pagination, nil
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