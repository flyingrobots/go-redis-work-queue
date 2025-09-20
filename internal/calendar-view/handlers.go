package calendarview

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// CalendarHandler handles HTTP requests for calendar operations
type CalendarHandler struct {
	manager *CalendarManager
	config  *CalendarConfig
}

// NewCalendarHandler creates a new calendar handler
func NewCalendarHandler(manager *CalendarManager, config *CalendarConfig) *CalendarHandler {
	return &CalendarHandler{
		manager: manager,
		config:  config,
	}
}

// RegisterRoutes registers calendar routes with the given router
func (h *CalendarHandler) RegisterRoutes(r *mux.Router) {
	// Calendar data endpoints
	r.HandleFunc("/calendar/data", h.GetCalendarData).Methods("GET")
	r.HandleFunc("/calendar/events", h.GetEvents).Methods("GET")

	// Reschedule operations
	r.HandleFunc("/calendar/reschedule", h.RescheduleEvent).Methods("POST")
	r.HandleFunc("/calendar/reschedule/bulk", h.BulkReschedule).Methods("POST")

	// Recurring rules CRUD
	r.HandleFunc("/calendar/rules", h.CreateRule).Methods("POST")
	r.HandleFunc("/calendar/rules", h.GetRules).Methods("GET")
	r.HandleFunc("/calendar/rules/{id}", h.GetRule).Methods("GET")
	r.HandleFunc("/calendar/rules/{id}", h.UpdateRule).Methods("PUT")
	r.HandleFunc("/calendar/rules/{id}", h.DeleteRule).Methods("DELETE")
	r.HandleFunc("/calendar/rules/{id}/pause", h.PauseRule).Methods("POST")
	r.HandleFunc("/calendar/rules/{id}/resume", h.ResumeRule).Methods("POST")

	// Configuration endpoints
	r.HandleFunc("/calendar/config", h.GetConfig).Methods("GET")
	r.HandleFunc("/calendar/timezones", h.GetTimezones).Methods("GET")
}

// GetCalendarData returns calendar data for the specified view
func (h *CalendarHandler) GetCalendarData(w http.ResponseWriter, r *http.Request) {
	view, err := h.parseCalendarViewFromQuery(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	data, err := h.manager.GetCalendarData(view)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, data)
}

// GetEvents returns events for a specified time window
func (h *CalendarHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	window, err := h.parseScheduleWindowFromQuery(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	response, err := h.manager.dataSource.GetEvents(*window)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// RescheduleEvent reschedules a single event
func (h *CalendarHandler) RescheduleEvent(w http.ResponseWriter, r *http.Request) {
	var req RescheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, NewCalendarError(
			ErrorCodeRuleValidation, "invalid request body", err.Error()))
		return
	}

	// TODO: Extract user ID from request context/auth
	userID := h.getUserIDFromRequest(r)

	result, err := h.manager.RescheduleEvent(req.EventID, req.NewTime, userID, req.Reason)
	if err != nil {
		status := http.StatusInternalServerError
		if calErr, ok := err.(*CalendarError); ok {
			switch calErr.Code {
			case ErrorCodeEventNotFound:
				status = http.StatusNotFound
			case ErrorCodeRescheduleConflict, ErrorCodeRuleValidation:
				status = http.StatusBadRequest
			case ErrorCodePermissionDenied:
				status = http.StatusForbidden
			}
		}
		h.writeErrorResponse(w, status, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, result)
}

// BulkReschedule reschedules multiple events
func (h *CalendarHandler) BulkReschedule(w http.ResponseWriter, r *http.Request) {
	var requests []RescheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, NewCalendarError(
			ErrorCodeRuleValidation, "invalid request body", err.Error()))
		return
	}

	userID := h.getUserIDFromRequest(r)
	results := make([]RescheduleResult, len(requests))
	errors := make([]error, len(requests))

	// Process each reschedule request
	for i, req := range requests {
		result, err := h.manager.RescheduleEvent(req.EventID, req.NewTime, userID, req.Reason)
		if result != nil {
			results[i] = *result
		}
		errors[i] = err
	}

	response := map[string]interface{}{
		"results": results,
		"errors":  errors,
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// CreateRule creates a new recurring rule
func (h *CalendarHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var req RuleCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, NewCalendarError(
			ErrorCodeRuleValidation, "invalid request body", err.Error()))
		return
	}

	rule, err := h.manager.dataSource.CreateRule(&req)
	if err != nil {
		status := http.StatusInternalServerError
		if calErr, ok := err.(*CalendarError); ok {
			switch calErr.Code {
			case ErrorCodeRuleValidation, ErrorCodeInvalidCronSpec:
				status = http.StatusBadRequest
			case ErrorCodePermissionDenied:
				status = http.StatusForbidden
			}
		}
		h.writeErrorResponse(w, status, err)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, rule)
}

// GetRules returns a list of recurring rules
func (h *CalendarHandler) GetRules(w http.ResponseWriter, r *http.Request) {
	filter := h.parseRuleFilterFromQuery(r)

	rules, err := h.manager.dataSource.GetRecurringRules(filter)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"rules": rules,
		"count": len(rules),
	})
}

// GetRule returns a specific recurring rule
func (h *CalendarHandler) GetRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["id"]

	// For now, return rule from rules list
	// In a real implementation, this would be a dedicated method
	filter := RuleFilter{IDs: []string{ruleID}}
	rules, err := h.manager.dataSource.GetRecurringRules(filter)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	if len(rules) == 0 {
		h.writeErrorResponse(w, http.StatusNotFound, ErrRuleNotFound(ruleID))
		return
	}

	h.writeJSONResponse(w, http.StatusOK, rules[0])
}

// UpdateRule updates an existing recurring rule
func (h *CalendarHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["id"]

	var req RuleUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, NewCalendarError(
			ErrorCodeRuleValidation, "invalid request body", err.Error()))
		return
	}

	req.ID = ruleID

	rule, err := h.manager.dataSource.UpdateRule(&req)
	if err != nil {
		status := http.StatusInternalServerError
		if calErr, ok := err.(*CalendarError); ok {
			switch calErr.Code {
			case ErrorCodeRuleNotFound:
				status = http.StatusNotFound
			case ErrorCodeRuleValidation, ErrorCodeInvalidCronSpec:
				status = http.StatusBadRequest
			case ErrorCodePermissionDenied:
				status = http.StatusForbidden
			}
		}
		h.writeErrorResponse(w, status, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, rule)
}

// DeleteRule deletes a recurring rule
func (h *CalendarHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["id"]

	if err := h.manager.dataSource.DeleteRule(ruleID); err != nil {
		status := http.StatusInternalServerError
		if calErr, ok := err.(*CalendarError); ok {
			switch calErr.Code {
			case ErrorCodeRuleNotFound:
				status = http.StatusNotFound
			case ErrorCodePermissionDenied:
				status = http.StatusForbidden
			}
		}
		h.writeErrorResponse(w, status, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PauseRule pauses a recurring rule
func (h *CalendarHandler) PauseRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["id"]

	if err := h.manager.dataSource.PauseRule(ruleID); err != nil {
		status := http.StatusInternalServerError
		if calErr, ok := err.(*CalendarError); ok {
			switch calErr.Code {
			case ErrorCodeRuleNotFound:
				status = http.StatusNotFound
			case ErrorCodePermissionDenied:
				status = http.StatusForbidden
			}
		}
		h.writeErrorResponse(w, status, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Rule paused successfully",
	})
}

// ResumeRule resumes a paused recurring rule
func (h *CalendarHandler) ResumeRule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ruleID := vars["id"]

	if err := h.manager.dataSource.ResumeRule(ruleID); err != nil {
		status := http.StatusInternalServerError
		if calErr, ok := err.(*CalendarError); ok {
			switch calErr.Code {
			case ErrorCodeRuleNotFound:
				status = http.StatusNotFound
			case ErrorCodePermissionDenied:
				status = http.StatusForbidden
			}
		}
		h.writeErrorResponse(w, status, err)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Rule resumed successfully",
	})
}

// GetConfig returns the calendar configuration
func (h *CalendarHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	h.writeJSONResponse(w, http.StatusOK, h.config)
}

// GetTimezones returns the list of supported timezones
func (h *CalendarHandler) GetTimezones(w http.ResponseWriter, r *http.Request) {
	timezones := DefaultTimezoneConfig().AllowedTimezones
	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"timezones": timezones,
	})
}

// Helper methods

// parseCalendarViewFromQuery parses calendar view parameters from query string
func (h *CalendarHandler) parseCalendarViewFromQuery(r *http.Request) (*CalendarView, error) {
	query := r.URL.Query()

	// Parse view type
	viewType := ViewTypeMonth // default
	if v := query.Get("view"); v != "" {
		switch strings.ToLower(v) {
		case "month":
			viewType = ViewTypeMonth
		case "week":
			viewType = ViewTypeWeek
		case "day":
			viewType = ViewTypeDay
		default:
			return nil, NewCalendarError(ErrorCodeInvalidFilter, "invalid view type", v)
		}
	}

	// Parse current date
	currentDate := time.Now()
	if d := query.Get("date"); d != "" {
		parsed, err := time.Parse("2006-01-02", d)
		if err != nil {
			return nil, NewCalendarError(ErrorCodeInvalidTimeRange, "invalid date format", d)
		}
		currentDate = parsed
	}

	// Parse timezone
	timezone := time.UTC
	if tz := query.Get("timezone"); tz != "" {
		location, err := time.LoadLocation(tz)
		if err != nil {
			return nil, ErrTimezoneNotFound(tz)
		}
		timezone = location
	}

	// Parse filters
	filter := EventFilter{
		QueueNames: parseStringSlice(query.Get("queues")),
		JobTypes:   parseStringSlice(query.Get("job_types")),
		Tags:       parseStringSlice(query.Get("tags")),
		SearchText: query.Get("search"),
	}

	// Parse status filter
	if statuses := query.Get("statuses"); statuses != "" {
		for _, s := range strings.Split(statuses, ",") {
			status := parseEventStatus(strings.TrimSpace(s))
			if status != StatusScheduled || s != "" { // Only add if valid
				filter.Statuses = append(filter.Statuses, status)
			}
		}
	}

	return &CalendarView{
		ViewType:    viewType,
		CurrentDate: currentDate.In(timezone),
		Timezone:    timezone,
		Filter:      filter,
	}, nil
}

// parseScheduleWindowFromQuery parses schedule window from query parameters
func (h *CalendarHandler) parseScheduleWindowFromQuery(r *http.Request) (*ScheduleWindow, error) {
	query := r.URL.Query()

	// Parse from time (required)
	fromStr := query.Get("from")
	if fromStr == "" {
		return nil, NewCalendarError(ErrorCodeInvalidTimeRange, "from parameter is required")
	}
	from, err := time.Parse(time.RFC3339, fromStr)
	if err != nil {
		return nil, NewCalendarError(ErrorCodeInvalidTimeRange, "invalid from time format", fromStr)
	}

	// Parse till time (required)
	tillStr := query.Get("till")
	if tillStr == "" {
		return nil, NewCalendarError(ErrorCodeInvalidTimeRange, "till parameter is required")
	}
	till, err := time.Parse(time.RFC3339, tillStr)
	if err != nil {
		return nil, NewCalendarError(ErrorCodeInvalidTimeRange, "invalid till time format", tillStr)
	}

	if !from.Before(till) {
		return nil, ErrInvalidTimeRange(from, till)
	}

	// Parse optional parameters
	window := &ScheduleWindow{
		From:      from,
		Till:      till,
		QueueName: query.Get("queue"),
	}

	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			return nil, NewCalendarError(ErrorCodeInvalidFilter, "invalid limit value", limitStr)
		}
		window.Limit = limit
	}

	return window, nil
}

// parseRuleFilterFromQuery parses rule filter from query parameters
func (h *CalendarHandler) parseRuleFilterFromQuery(r *http.Request) RuleFilter {
	query := r.URL.Query()

	filter := RuleFilter{
		QueueNames: parseStringSlice(query.Get("queues")),
		JobTypes:   parseStringSlice(query.Get("job_types")),
	}

	if active := query.Get("active"); active != "" {
		isActive := active == "true"
		filter.IsActive = &isActive
	}

	if paused := query.Get("paused"); paused != "" {
		isPaused := paused == "true"
		filter.IsPaused = &isPaused
	}

	return filter
}

// parseStringSlice parses a comma-separated string into a slice
func parseStringSlice(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// parseEventStatus parses string to EventStatus
func parseEventStatus(s string) EventStatus {
	switch strings.ToLower(s) {
	case "scheduled":
		return StatusScheduled
	case "running":
		return StatusRunning
	case "completed":
		return StatusCompleted
	case "failed":
		return StatusFailed
	case "canceled":
		return StatusCanceled
	default:
		return StatusScheduled
	}
}

// getUserIDFromRequest extracts user ID from request (placeholder)
func (h *CalendarHandler) getUserIDFromRequest(r *http.Request) string {
	// TODO: Implement proper user authentication/authorization
	// This could come from JWT token, session, or other auth mechanism
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return userID
	}
	return "anonymous"
}

// writeJSONResponse writes a JSON response
func (h *CalendarHandler) writeJSONResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// writeErrorResponse writes an error response
func (h *CalendarHandler) writeErrorResponse(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	errorResponse := map[string]interface{}{
		"error":   err.Error(),
		"message": err.Error(),
	}

	if calErr, ok := err.(*CalendarError); ok {
		errorResponse["code"] = calErr.Code
		if calErr.Details != "" {
			errorResponse["details"] = calErr.Details
		}
	}

	json.NewEncoder(w).Encode(errorResponse)
}

// RuleFilter represents filter options for recurring rules
type RuleFilter struct {
	IDs        []string `json:"ids,omitempty"`
	QueueNames []string `json:"queue_names,omitempty"`
	JobTypes   []string `json:"job_types,omitempty"`
	IsActive   *bool    `json:"is_active,omitempty"`
	IsPaused   *bool    `json:"is_paused,omitempty"`
}
