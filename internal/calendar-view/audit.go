package calendarview

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"
)

// AuditLogger provides audit logging for calendar operations
type AuditLogger struct {
	entries []AuditEntry
	mutex   sync.RWMutex
	config  *AuditConfig
}

// AuditConfig represents audit logging configuration
type AuditConfig struct {
	Enabled        bool          `json:"enabled"`
	MaxEntries     int           `json:"max_entries"`
	RetentionDays  int           `json:"retention_days"`
	LogToConsole   bool          `json:"log_to_console"`
	LogToFile      bool          `json:"log_to_file"`
	LogFile        string        `json:"log_file"`
	DetailLevel    AuditLevel    `json:"detail_level"`
	RotationSize   int64         `json:"rotation_size"`
	RotationBackups int          `json:"rotation_backups"`
}

// AuditLevel represents the level of audit detail
type AuditLevel int

const (
	AuditLevelBasic AuditLevel = iota
	AuditLevelDetailed
	AuditLevelVerbose
)

// AuditAction represents different types of audit actions
type AuditAction string

const (
	ActionViewCalendar   AuditAction = "view_calendar"
	ActionRescheduleEvent AuditAction = "reschedule_event"
	ActionCreateRule     AuditAction = "create_rule"
	ActionUpdateRule     AuditAction = "update_rule"
	ActionDeleteRule     AuditAction = "delete_rule"
	ActionPauseRule      AuditAction = "pause_rule"
	ActionResumeRule     AuditAction = "resume_rule"
	ActionFilterEvents   AuditAction = "filter_events"
	ActionNavigate       AuditAction = "navigate"
	ActionExportData     AuditAction = "export_data"
	ActionBulkOperation  AuditAction = "bulk_operation"
)

// AuditResult represents the result of an audited action
type AuditResult string

const (
	ResultSuccess AuditResult = "success"
	ResultFailure AuditResult = "failure"
	ResultPartial AuditResult = "partial"
	ResultDenied  AuditResult = "denied"
)

// NewAuditLogger creates a new audit logger
func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		entries: make([]AuditEntry, 0),
		config:  DefaultAuditConfig(),
	}
}

// NewAuditLoggerWithConfig creates a new audit logger with custom configuration
func NewAuditLoggerWithConfig(config *AuditConfig) *AuditLogger {
	return &AuditLogger{
		entries: make([]AuditEntry, 0),
		config:  config,
	}
}

// DefaultAuditConfig returns default audit configuration
func DefaultAuditConfig() *AuditConfig {
	return &AuditConfig{
		Enabled:         true,
		MaxEntries:      10000,
		RetentionDays:   90,
		LogToConsole:    true,
		LogToFile:       false,
		LogFile:         "calendar-audit.log",
		DetailLevel:     AuditLevelDetailed,
		RotationSize:    10 * 1024 * 1024, // 10MB
		RotationBackups: 5,
	}
}

// LogSuccess logs a successful operation
func (al *AuditLogger) LogSuccess(action string, userID string, details map[string]string) {
	al.logEntry(AuditAction(action), userID, ResultSuccess, "", details)
}

// LogFailure logs a failed operation
func (al *AuditLogger) LogFailure(action string, userID string, err error, details map[string]string) {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	al.logEntry(AuditAction(action), userID, ResultFailure, errorMsg, details)
}

// LogError logs an error operation
func (al *AuditLogger) LogError(action string, userID string, err error) {
	errorMsg := ""
	if err != nil {
		errorMsg = err.Error()
	}
	al.logEntry(AuditAction(action), userID, ResultFailure, errorMsg, nil)
}

// LogPartial logs a partially successful operation
func (al *AuditLogger) LogPartial(action string, userID string, details map[string]string) {
	al.logEntry(AuditAction(action), userID, ResultPartial, "", details)
}

// LogDenied logs a denied operation
func (al *AuditLogger) LogDenied(action string, userID string, reason string) {
	details := map[string]string{"denial_reason": reason}
	al.logEntry(AuditAction(action), userID, ResultDenied, reason, details)
}

// LogCalendarView logs calendar view access
func (al *AuditLogger) LogCalendarView(userID string, viewType ViewType, date time.Time, filter *EventFilter) {
	details := map[string]string{
		"view_type": fmt.Sprintf("%d", viewType),
		"date":      date.Format("2006-01-02"),
	}

	if filter != nil {
		if len(filter.QueueNames) > 0 {
			details["filtered_queues"] = fmt.Sprintf("%v", filter.QueueNames)
		}
		if len(filter.JobTypes) > 0 {
			details["filtered_job_types"] = fmt.Sprintf("%v", filter.JobTypes)
		}
		if filter.SearchText != "" {
			details["search_text"] = filter.SearchText
		}
	}

	al.logEntry(ActionViewCalendar, userID, ResultSuccess, "", details)
}

// LogReschedule logs event reschedule operations
func (al *AuditLogger) LogReschedule(userID string, eventID string, oldTime, newTime time.Time, reason string, success bool) {
	details := map[string]string{
		"event_id": eventID,
		"old_time": oldTime.Format(time.RFC3339),
		"new_time": newTime.Format(time.RFC3339),
		"reason":   reason,
	}

	result := ResultSuccess
	if !success {
		result = ResultFailure
	}

	al.logEntry(ActionRescheduleEvent, userID, result, "", details)
}

// LogRuleOperation logs recurring rule operations
func (al *AuditLogger) LogRuleOperation(action AuditAction, userID string, ruleID string, ruleName string, success bool, error string) {
	details := map[string]string{
		"rule_id":   ruleID,
		"rule_name": ruleName,
	}

	result := ResultSuccess
	if !success {
		result = ResultFailure
	}

	al.logEntry(action, userID, result, error, details)
}

// LogBulkOperation logs bulk operations
func (al *AuditLogger) LogBulkOperation(userID string, operation string, totalItems, successCount, failureCount int) {
	details := map[string]string{
		"operation":      operation,
		"total_items":    fmt.Sprintf("%d", totalItems),
		"success_count":  fmt.Sprintf("%d", successCount),
		"failure_count":  fmt.Sprintf("%d", failureCount),
	}

	result := ResultSuccess
	if failureCount > 0 {
		if successCount > 0 {
			result = ResultPartial
		} else {
			result = ResultFailure
		}
	}

	al.logEntry(ActionBulkOperation, userID, result, "", details)
}

// LogNavigation logs navigation actions
func (al *AuditLogger) LogNavigation(userID string, action NavigationAction, from, to time.Time) {
	details := map[string]string{
		"navigation_action": fmt.Sprintf("%d", action),
		"from_date":         from.Format("2006-01-02"),
		"to_date":           to.Format("2006-01-02"),
	}

	al.logEntry(ActionNavigate, userID, ResultSuccess, "", details)
}

// logEntry creates and stores an audit entry
func (al *AuditLogger) logEntry(action AuditAction, userID string, result AuditResult, errorMsg string, details map[string]string) {
	if !al.config.Enabled {
		return
	}

	al.mutex.Lock()
	defer al.mutex.Unlock()

	entry := AuditEntry{
		ID:        al.generateEntryID(),
		Action:    string(action),
		UserID:    userID,
		Timestamp: time.Now().UTC(),
		Details:   details,
		Success:   result == ResultSuccess,
		Error:     errorMsg,
	}

	// Add additional context based on detail level
	if al.config.DetailLevel >= AuditLevelDetailed {
		if entry.Details == nil {
			entry.Details = make(map[string]string)
		}
		entry.Details["result"] = string(result)
		entry.Details["timestamp_unix"] = fmt.Sprintf("%d", entry.Timestamp.Unix())
	}

	if al.config.DetailLevel >= AuditLevelVerbose {
		entry.Details["entry_id"] = entry.ID
	}

	// Store entry
	al.entries = append(al.entries, entry)

	// Trim entries if we exceed the maximum
	if len(al.entries) > al.config.MaxEntries {
		al.entries = al.entries[len(al.entries)-al.config.MaxEntries:]
	}

	// Log to console if enabled
	if al.config.LogToConsole {
		al.logToConsole(entry)
	}

	// Log to file if enabled
	if al.config.LogToFile {
		al.logToFile(entry)
	}
}

// GetEntries returns audit entries with optional filtering
func (al *AuditLogger) GetEntries(filter *AuditFilter) []AuditEntry {
	al.mutex.RLock()
	defer al.mutex.RUnlock()

	if filter == nil {
		// Return a copy of all entries
		result := make([]AuditEntry, len(al.entries))
		copy(result, al.entries)
		return result
	}

	var filtered []AuditEntry
	for _, entry := range al.entries {
		if al.matchesFilter(entry, filter) {
			filtered = append(filtered, entry)
		}
	}

	// Apply limit
	if filter.Limit > 0 && len(filtered) > filter.Limit {
		filtered = filtered[:filter.Limit]
	}

	return filtered
}

// GetEntriesCount returns the count of audit entries
func (al *AuditLogger) GetEntriesCount() int {
	al.mutex.RLock()
	defer al.mutex.RUnlock()
	return len(al.entries)
}

// ClearEntries removes all audit entries
func (al *AuditLogger) ClearEntries() {
	al.mutex.Lock()
	defer al.mutex.Unlock()
	al.entries = make([]AuditEntry, 0)
}

// ClearOldEntries removes entries older than the retention period
func (al *AuditLogger) ClearOldEntries() {
	if al.config.RetentionDays <= 0 {
		return
	}

	al.mutex.Lock()
	defer al.mutex.Unlock()

	cutoff := time.Now().AddDate(0, 0, -al.config.RetentionDays)
	var retained []AuditEntry

	for _, entry := range al.entries {
		if entry.Timestamp.After(cutoff) {
			retained = append(retained, entry)
		}
	}

	al.entries = retained
}

// ExportEntries exports audit entries as JSON
func (al *AuditLogger) ExportEntries(filter *AuditFilter) ([]byte, error) {
	entries := al.GetEntries(filter)
	return json.MarshalIndent(entries, "", "  ")
}

// Helper methods

// generateEntryID generates a unique ID for audit entries
func (al *AuditLogger) generateEntryID() string {
	return fmt.Sprintf("audit_%d_%d", time.Now().Unix(), len(al.entries))
}

// matchesFilter checks if an entry matches the filter criteria
func (al *AuditLogger) matchesFilter(entry AuditEntry, filter *AuditFilter) bool {
	// Time range filter
	if !filter.From.IsZero() && entry.Timestamp.Before(filter.From) {
		return false
	}
	if !filter.To.IsZero() && entry.Timestamp.After(filter.To) {
		return false
	}

	// User ID filter
	if filter.UserID != "" && entry.UserID != filter.UserID {
		return false
	}

	// Action filter
	if len(filter.Actions) > 0 {
		found := false
		for _, action := range filter.Actions {
			if entry.Action == action {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Success filter
	if filter.SuccessOnly != nil && entry.Success != *filter.SuccessOnly {
		return false
	}

	return true
}

// logToConsole outputs the audit entry to console
func (al *AuditLogger) logToConsole(entry AuditEntry) {
	level := "INFO"
	if !entry.Success {
		level = "ERROR"
	}

	log.Printf("[AUDIT] [%s] Action: %s, User: %s, Success: %v, Details: %v",
		level, entry.Action, entry.UserID, entry.Success, entry.Details)

	if entry.Error != "" {
		log.Printf("[AUDIT] [ERROR] %s", entry.Error)
	}
}

// logToFile outputs the audit entry to file
func (al *AuditLogger) logToFile(entry AuditEntry) {
	// TODO: Implement file logging with rotation
	// This would require additional file handling logic
	// For now, we'll skip this implementation
}

// AuditFilter represents filter criteria for audit entries
type AuditFilter struct {
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
	UserID      string    `json:"user_id"`
	Actions     []string  `json:"actions"`
	SuccessOnly *bool     `json:"success_only"`
	Limit       int       `json:"limit"`
}

// AuditSummary represents a summary of audit entries
type AuditSummary struct {
	TotalEntries   int                    `json:"total_entries"`
	SuccessCount   int                    `json:"success_count"`
	FailureCount   int                    `json:"failure_count"`
	ActionCounts   map[string]int         `json:"action_counts"`
	UserCounts     map[string]int         `json:"user_counts"`
	TimeRange      TimeRange              `json:"time_range"`
	TopActions     []ActionSummary        `json:"top_actions"`
	TopUsers       []UserSummary          `json:"top_users"`
	ErrorSummary   map[string]int         `json:"error_summary"`
}

// ActionSummary represents summary data for an action
type ActionSummary struct {
	Action       string  `json:"action"`
	Count        int     `json:"count"`
	SuccessRate  float64 `json:"success_rate"`
	LastOccurred time.Time `json:"last_occurred"`
}

// UserSummary represents summary data for a user
type UserSummary struct {
	UserID       string    `json:"user_id"`
	Count        int       `json:"count"`
	SuccessRate  float64   `json:"success_rate"`
	LastActivity time.Time `json:"last_activity"`
}

// GetSummary returns a summary of audit entries
func (al *AuditLogger) GetSummary(filter *AuditFilter) *AuditSummary {
	entries := al.GetEntries(filter)

	summary := &AuditSummary{
		TotalEntries: len(entries),
		ActionCounts: make(map[string]int),
		UserCounts:   make(map[string]int),
		ErrorSummary: make(map[string]int),
	}

	if len(entries) == 0 {
		return summary
	}

	// Initialize time range
	summary.TimeRange.Start = entries[0].Timestamp
	summary.TimeRange.End = entries[0].Timestamp

	// Collect statistics
	for _, entry := range entries {
		// Update time range
		if entry.Timestamp.Before(summary.TimeRange.Start) {
			summary.TimeRange.Start = entry.Timestamp
		}
		if entry.Timestamp.After(summary.TimeRange.End) {
			summary.TimeRange.End = entry.Timestamp
		}

		// Count success/failure
		if entry.Success {
			summary.SuccessCount++
		} else {
			summary.FailureCount++
			if entry.Error != "" {
				summary.ErrorSummary[entry.Error]++
			}
		}

		// Count actions and users
		summary.ActionCounts[entry.Action]++
		summary.UserCounts[entry.UserID]++
	}

	return summary
}