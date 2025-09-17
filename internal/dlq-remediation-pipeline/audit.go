// Copyright 2025 James Ross
package dlqremediation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// AuditLogger handles audit logging for the remediation pipeline
type AuditLogger struct {
	redis  *redis.Client
	logger *zap.Logger
	config *PipelineConfig
}

// NewAuditLogger creates a new audit logger
func NewAuditLogger(redisClient *redis.Client, config *PipelineConfig, logger *zap.Logger) *AuditLogger {
	return &AuditLogger{
		redis:  redisClient,
		logger: logger,
		config: config,
	}
}

// LogRemediation logs a remediation action
func (al *AuditLogger) LogRemediation(ctx context.Context, job *DLQJob, rule *RemediationRule, classification *Classification, result *ProcessingResult, userID string) error {
	if !al.config.AuditEnabled {
		return nil
	}

	entry := &AuditLogEntry{
		ID:          uuid.New().String(),
		Timestamp:   time.Now(),
		JobID:       job.JobID,
		RuleID:      rule.ID,
		RuleName:    rule.Name,
		Action:      ActionType("classify_and_remediate"),
		Parameters:  al.buildParameters(rule, classification),
		Result:      al.buildResult(result),
		DryRun:      result.DryRun,
		UserID:      userID,
		Duration:    result.Duration,
		BeforeState: al.serializeJob(result.BeforeState),
		AfterState:  al.serializeJob(result.AfterState),
	}

	if !result.Success {
		entry.Error = result.Error
	}

	return al.storeAuditEntry(ctx, entry)
}

// LogRuleChange logs changes to remediation rules
func (al *AuditLogger) LogRuleChange(ctx context.Context, ruleID, operation, userID string, before, after *RemediationRule) error {
	if !al.config.AuditEnabled {
		return nil
	}

	entry := &AuditLogEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		RuleID:    ruleID,
		Action:    ActionType(operation),
		UserID:    userID,
		Result:    "success",
	}

	if before != nil {
		entry.BeforeState = al.serializeRule(before)
	}
	if after != nil {
		entry.AfterState = al.serializeRule(after)
	}

	return al.storeAuditEntry(ctx, entry)
}

// LogPipelineState logs pipeline state changes
func (al *AuditLogger) LogPipelineState(ctx context.Context, operation string, state *PipelineState, userID string) error {
	if !al.config.AuditEnabled {
		return nil
	}

	entry := &AuditLogEntry{
		ID:         uuid.New().String(),
		Timestamp:  time.Now(),
		Action:     ActionType(operation),
		UserID:     userID,
		Result:     "success",
		AfterState: al.serializeState(state),
	}

	return al.storeAuditEntry(ctx, entry)
}

// LogBulkOperation logs bulk operations
func (al *AuditLogger) LogBulkOperation(ctx context.Context, operation string, results *BatchResult, userID string) error {
	if !al.config.AuditEnabled {
		return nil
	}

	entry := &AuditLogEntry{
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Action:    ActionType(operation),
		UserID:    userID,
		Duration:  results.CompletedAt.Sub(results.StartedAt),
		Parameters: map[string]interface{}{
			"total_jobs":      results.TotalJobs,
			"processed_jobs":  results.ProcessedJobs,
			"successful_jobs": results.SuccessfulJobs,
			"failed_jobs":     results.FailedJobs,
			"skipped_jobs":    results.SkippedJobs,
		},
	}

	if results.SuccessfulJobs == results.ProcessedJobs {
		entry.Result = "success"
	} else if results.SuccessfulJobs > 0 {
		entry.Result = "partial_success"
	} else {
		entry.Result = "failure"
	}

	if len(results.Errors) > 0 {
		entry.Error = fmt.Sprintf("Errors: %v", results.Errors)
	}

	return al.storeAuditEntry(ctx, entry)
}

// GetAuditLog retrieves audit log entries
func (al *AuditLogger) GetAuditLog(ctx context.Context, filter AuditFilter) ([]*AuditLogEntry, error) {
	// Get all audit entries from Redis
	entries, err := al.redis.LRange(ctx, "audit:log", 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve audit log: %w", err)
	}

	var auditEntries []*AuditLogEntry
	for _, entryData := range entries {
		var entry AuditLogEntry
		if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
			al.logger.Warn("Failed to unmarshal audit entry", zap.Error(err))
			continue
		}

		if al.matchesFilter(&entry, filter) {
			auditEntries = append(auditEntries, &entry)
		}
	}

	// Apply sorting and limiting
	return al.applySortingAndLimiting(auditEntries, filter), nil
}

// CleanupExpiredEntries removes old audit log entries
func (al *AuditLogger) CleanupExpiredEntries(ctx context.Context) error {
	if !al.config.AuditEnabled {
		return nil
	}

	cutoff := time.Now().Add(-al.config.RetentionPolicy.AuditLogTTL)

	// Get all entries
	entries, err := al.redis.LRange(ctx, "audit:log", 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to retrieve audit log for cleanup: %w", err)
	}

	var validEntries []string
	expiredCount := 0

	for _, entryData := range entries {
		var entry AuditLogEntry
		if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
			continue
		}

		if entry.Timestamp.After(cutoff) {
			validEntries = append(validEntries, entryData)
		} else {
			expiredCount++
		}
	}

	if expiredCount > 0 {
		// Replace the list with valid entries
		pipe := al.redis.Pipeline()
		pipe.Del(ctx, "audit:log")
		if len(validEntries) > 0 {
			pipe.LPush(ctx, "audit:log", validEntries)
		}
		_, err := pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to cleanup expired audit entries: %w", err)
		}

		al.logger.Info("Cleaned up expired audit entries",
			zap.Int("expired_count", expiredCount),
			zap.Int("remaining_count", len(validEntries)))
	}

	return nil
}

// storeAuditEntry stores an audit entry in Redis
func (al *AuditLogger) storeAuditEntry(ctx context.Context, entry *AuditLogEntry) error {
	entryData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal audit entry: %w", err)
	}

	// Store in Redis list (most recent first)
	err = al.redis.LPush(ctx, "audit:log", string(entryData)).Err()
	if err != nil {
		return fmt.Errorf("failed to store audit entry: %w", err)
	}

	// Also store in time-series for analytics
	timeKey := fmt.Sprintf("audit:by_day:%s", entry.Timestamp.Format("2006-01-02"))
	al.redis.Incr(ctx, timeKey)
	al.redis.Expire(ctx, timeKey, al.config.RetentionPolicy.AuditLogTTL)

	// Store action-specific metrics
	actionKey := fmt.Sprintf("audit:by_action:%s", entry.Action)
	al.redis.Incr(ctx, actionKey)
	al.redis.Expire(ctx, actionKey, al.config.RetentionPolicy.AuditLogTTL)

	al.logger.Debug("Audit entry stored",
		zap.String("entry_id", entry.ID),
		zap.String("job_id", entry.JobID),
		zap.String("action", string(entry.Action)))

	return nil
}

// buildParameters builds parameters map for audit log
func (al *AuditLogger) buildParameters(rule *RemediationRule, classification *Classification) map[string]interface{} {
	params := map[string]interface{}{
		"rule_priority":             rule.Priority,
		"classification_confidence": classification.Confidence,
		"classification_category":   classification.Category,
		"classification_reason":     classification.Reason,
		"actions_count":             len(rule.Actions),
	}

	if len(rule.Actions) > 0 {
		actionTypes := make([]string, len(rule.Actions))
		for i, action := range rule.Actions {
			actionTypes[i] = string(action.Type)
		}
		params["action_types"] = actionTypes
	}

	return params
}

// buildResult builds result string for audit log
func (al *AuditLogger) buildResult(result *ProcessingResult) string {
	if result.Success {
		return "success"
	}
	return "failure"
}

// serializeJob serializes a job for audit storage
func (al *AuditLogger) serializeJob(job *DLQJob) json.RawMessage {
	if job == nil {
		return nil
	}

	// Create a copy without the full payload for audit efficiency
	auditJob := map[string]interface{}{
		"id":           job.ID,
		"job_id":       job.JobID,
		"queue":        job.Queue,
		"job_type":     job.JobType,
		"error":        job.Error,
		"error_type":   job.ErrorType,
		"retry_count":  job.RetryCount,
		"failed_at":    job.FailedAt,
		"payload_size": job.PayloadSize,
		"worker_id":    job.WorkerID,
		"trace_id":     job.TraceID,
	}

	// Include metadata if present
	if job.Metadata != nil {
		auditJob["metadata"] = job.Metadata
	}

	// Include payload hash for verification (but not full payload for privacy)
	if len(job.Payload) > 0 {
		// Could implement payload hashing here
		auditJob["payload_hash"] = fmt.Sprintf("sha256:%x", len(job.Payload)) // Simplified
	}

	data, _ := json.Marshal(auditJob)
	return data
}

// serializeRule serializes a rule for audit storage
func (al *AuditLogger) serializeRule(rule *RemediationRule) json.RawMessage {
	if rule == nil {
		return nil
	}

	data, _ := json.Marshal(rule)
	return data
}

// serializeState serializes pipeline state for audit storage
func (al *AuditLogger) serializeState(state *PipelineState) json.RawMessage {
	if state == nil {
		return nil
	}

	data, _ := json.Marshal(state)
	return data
}

// AuditFilter defines filters for audit log queries
type AuditFilter struct {
	JobID     string     `json:"job_id,omitempty"`
	RuleID    string     `json:"rule_id,omitempty"`
	Action    ActionType `json:"action,omitempty"`
	UserID    string     `json:"user_id,omitempty"`
	Result    string     `json:"result,omitempty"`
	DryRun    *bool      `json:"dry_run,omitempty"`
	StartTime *time.Time `json:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty"`
	Limit     int        `json:"limit,omitempty"`
	Offset    int        `json:"offset,omitempty"`
	SortBy    string     `json:"sort_by,omitempty"`    // timestamp, duration
	SortOrder string     `json:"sort_order,omitempty"` // asc, desc
}

// matchesFilter checks if an audit entry matches the filter
func (al *AuditLogger) matchesFilter(entry *AuditLogEntry, filter AuditFilter) bool {
	if filter.JobID != "" && entry.JobID != filter.JobID {
		return false
	}

	if filter.RuleID != "" && entry.RuleID != filter.RuleID {
		return false
	}

	if filter.Action != "" && entry.Action != filter.Action {
		return false
	}

	if filter.UserID != "" && entry.UserID != filter.UserID {
		return false
	}

	if filter.Result != "" && entry.Result != filter.Result {
		return false
	}

	if filter.DryRun != nil && entry.DryRun != *filter.DryRun {
		return false
	}

	if filter.StartTime != nil && entry.Timestamp.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && entry.Timestamp.After(*filter.EndTime) {
		return false
	}

	return true
}

// applySortingAndLimiting applies sorting and limiting to audit entries
func (al *AuditLogger) applySortingAndLimiting(entries []*AuditLogEntry, filter AuditFilter) []*AuditLogEntry {
	// Apply sorting
	if filter.SortBy == "duration" {
		// Sort by duration
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if filter.SortOrder == "asc" {
					if entries[i].Duration > entries[j].Duration {
						entries[i], entries[j] = entries[j], entries[i]
					}
				} else {
					if entries[i].Duration < entries[j].Duration {
						entries[i], entries[j] = entries[j], entries[i]
					}
				}
			}
		}
	} else {
		// Sort by timestamp (default)
		for i := 0; i < len(entries)-1; i++ {
			for j := i + 1; j < len(entries); j++ {
				if filter.SortOrder == "asc" {
					if entries[i].Timestamp.After(entries[j].Timestamp) {
						entries[i], entries[j] = entries[j], entries[i]
					}
				} else {
					if entries[i].Timestamp.Before(entries[j].Timestamp) {
						entries[i], entries[j] = entries[j], entries[i]
					}
				}
			}
		}
	}

	// Apply offset and limit
	start := filter.Offset
	if start >= len(entries) {
		return []*AuditLogEntry{}
	}

	end := len(entries)
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}

	return entries[start:end]
}

// GetAuditStatistics returns statistics about audit log
func (al *AuditLogger) GetAuditStatistics(ctx context.Context, days int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Get total entries count
	totalEntries, err := al.redis.LLen(ctx, "audit:log").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get total entries count: %w", err)
	}
	stats["total_entries"] = totalEntries

	// Get daily counts for the last N days
	dailyCounts := make(map[string]int64)
	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i).Format("2006-01-02")
		key := fmt.Sprintf("audit:by_day:%s", date)
		count, _ := al.redis.Get(ctx, key).Int64()
		dailyCounts[date] = count
	}
	stats["daily_counts"] = dailyCounts

	// Get action type distribution
	actionTypes := []ActionType{
		ActionRequeue, ActionTransform, ActionRedact, ActionDrop,
		ActionRoute, ActionDelay, ActionTag, ActionNotify,
	}

	actionCounts := make(map[string]int64)
	for _, actionType := range actionTypes {
		key := fmt.Sprintf("audit:by_action:%s", actionType)
		count, _ := al.redis.Get(ctx, key).Int64()
		if count > 0 {
			actionCounts[string(actionType)] = count
		}
	}
	stats["action_counts"] = actionCounts

	return stats, nil
}
