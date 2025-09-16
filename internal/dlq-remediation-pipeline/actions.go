// Copyright 2025 James Ross
package dlqremediation

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// ActionExecutor executes remediation actions on DLQ jobs
type ActionExecutor struct {
	redis  *redis.Client
	logger *zap.Logger
	config *PipelineConfig
}

// NewActionExecutor creates a new action executor
func NewActionExecutor(redisClient *redis.Client, config *PipelineConfig, logger *zap.Logger) *ActionExecutor {
	return &ActionExecutor{
		redis:  redisClient,
		logger: logger,
		config: config,
	}
}

// Execute executes a list of actions on a DLQ job
func (ae *ActionExecutor) Execute(ctx context.Context, job *DLQJob, actions []Action, dryRun bool) (*ProcessingResult, error) {
	startTime := time.Now()
	result := &ProcessingResult{
		JobID:       job.JobID,
		Actions:     make([]ActionType, len(actions)),
		DryRun:      dryRun,
		BeforeState: ae.copyJob(job),
	}

	// Copy job for modification
	modifiedJob := ae.copyJob(job)

	// Execute each action in sequence
	for i, action := range actions {
		result.Actions[i] = action.Type

		if !ae.shouldExecuteAction(modifiedJob, action) {
			ae.logger.Debug("Skipping action due to conditions",
				zap.String("job_id", job.JobID),
				zap.String("action_type", string(action.Type)))
			continue
		}

		err := ae.executeAction(ctx, modifiedJob, action, dryRun)
		if err != nil {
			result.Success = false
			result.Error = err.Error()
			result.Duration = time.Since(startTime)
			return result, err
		}
	}

	result.Success = true
	result.AfterState = modifiedJob
	result.Duration = time.Since(startTime)

	return result, nil
}

// shouldExecuteAction checks if action conditions are met
func (ae *ActionExecutor) shouldExecuteAction(job *DLQJob, action Action) bool {
	for _, condition := range action.Conditions {
		if !ae.evaluateCondition(job, condition) {
			return false
		}
	}
	return true
}

// evaluateCondition evaluates a single action condition
func (ae *ActionExecutor) evaluateCondition(job *DLQJob, condition ActionCondition) bool {
	var fieldValue interface{}

	switch condition.Field {
	case "job_id":
		fieldValue = job.JobID
	case "queue":
		fieldValue = job.Queue
	case "job_type":
		fieldValue = job.JobType
	case "error":
		fieldValue = job.Error
	case "error_type":
		fieldValue = job.ErrorType
	case "retry_count":
		fieldValue = job.RetryCount
	case "payload_size":
		fieldValue = job.PayloadSize
	case "worker_id":
		fieldValue = job.WorkerID
	case "trace_id":
		fieldValue = job.TraceID
	default:
		// Try to extract from metadata or payload
		if strings.HasPrefix(condition.Field, "metadata.") {
			key := strings.TrimPrefix(condition.Field, "metadata.")
			fieldValue = job.Metadata[key]
		} else if strings.HasPrefix(condition.Field, "payload.") {
			path := strings.TrimPrefix(condition.Field, "payload.")
			var data interface{}
			if err := json.Unmarshal(job.Payload, &data); err == nil {
				if value, err := jsonpath.Get(path, data); err == nil {
					fieldValue = value
				}
			}
		}
	}

	return ae.compareConditionValue(fieldValue, condition.Operator, condition.Value)
}

// compareConditionValue compares field value against condition
func (ae *ActionExecutor) compareConditionValue(fieldValue interface{}, operator string, expectedValue interface{}) bool {
	switch operator {
	case "equals", "=", "==":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", expectedValue)
	case "not_equals", "!=", "<>":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", expectedValue)
	case "contains":
		fieldStr := fmt.Sprintf("%v", fieldValue)
		expectedStr := fmt.Sprintf("%v", expectedValue)
		return strings.Contains(fieldStr, expectedStr)
	case "not_contains":
		fieldStr := fmt.Sprintf("%v", fieldValue)
		expectedStr := fmt.Sprintf("%v", expectedValue)
		return !strings.Contains(fieldStr, expectedStr)
	case "regex":
		fieldStr := fmt.Sprintf("%v", fieldValue)
		patternStr := fmt.Sprintf("%v", expectedValue)
		if regex, err := regexp.Compile(patternStr); err == nil {
			return regex.MatchString(fieldStr)
		}
		return false
	case "gt", ">":
		return ae.compareNumeric(fieldValue, expectedValue, ">")
	case "gte", ">=":
		return ae.compareNumeric(fieldValue, expectedValue, ">=")
	case "lt", "<":
		return ae.compareNumeric(fieldValue, expectedValue, "<")
	case "lte", "<=":
		return ae.compareNumeric(fieldValue, expectedValue, "<=")
	default:
		return false
	}
}

// compareNumeric compares numeric values
func (ae *ActionExecutor) compareNumeric(a, b interface{}, operator string) bool {
	aFloat, aOk := ae.toFloat64(a)
	bFloat, bOk := ae.toFloat64(b)

	if !aOk || !bOk {
		return false
	}

	switch operator {
	case ">":
		return aFloat > bFloat
	case ">=":
		return aFloat >= bFloat
	case "<":
		return aFloat < bFloat
	case "<=":
		return aFloat <= bFloat
	default:
		return false
	}
}

// toFloat64 converts interface{} to float64
func (ae *ActionExecutor) toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
	}
	return 0, false
}

// executeAction executes a single action
func (ae *ActionExecutor) executeAction(ctx context.Context, job *DLQJob, action Action, dryRun bool) error {
	ae.logger.Debug("Executing action",
		zap.String("job_id", job.JobID),
		zap.String("action_type", string(action.Type)),
		zap.Bool("dry_run", dryRun))

	switch action.Type {
	case ActionRequeue:
		return ae.executeRequeue(ctx, job, action.Parameters, dryRun)
	case ActionTransform:
		return ae.executeTransform(ctx, job, action.Parameters, dryRun)
	case ActionRedact:
		return ae.executeRedact(ctx, job, action.Parameters, dryRun)
	case ActionDrop:
		return ae.executeDrop(ctx, job, action.Parameters, dryRun)
	case ActionRoute:
		return ae.executeRoute(ctx, job, action.Parameters, dryRun)
	case ActionDelay:
		return ae.executeDelay(ctx, job, action.Parameters, dryRun)
	case ActionTag:
		return ae.executeTag(ctx, job, action.Parameters, dryRun)
	case ActionNotify:
		return ae.executeNotify(ctx, job, action.Parameters, dryRun)
	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// executeRequeue requeues the job to a target queue
func (ae *ActionExecutor) executeRequeue(ctx context.Context, job *DLQJob, params map[string]interface{}, dryRun bool) error {
	targetQueue := ae.getStringParam(params, "target_queue", job.Queue)
	priority := ae.getIntParam(params, "priority", 5)
	delay := ae.getDurationParam(params, "delay", 0)
	resetRetryCount := ae.getBoolParam(params, "reset_retry_count", false)

	if dryRun {
		ae.logger.Info("DRY RUN: Would requeue job",
			zap.String("job_id", job.JobID),
			zap.String("target_queue", targetQueue),
			zap.Int("priority", priority),
			zap.Duration("delay", delay))
		return nil
	}

	// Reset retry count if requested
	if resetRetryCount {
		job.RetryCount = 0
	}

	// Prepare job data for requeue
	jobData := map[string]interface{}{
		"id":          job.JobID,
		"type":        job.JobType,
		"payload":     string(job.Payload),
		"priority":    priority,
		"retry_count": job.RetryCount,
		"queue":       targetQueue,
		"scheduled_at": time.Now().Add(delay).Unix(),
	}

	// Add to target queue
	pipe := ae.redis.Pipeline()

	// Add to queue
	jobJSON, _ := json.Marshal(jobData)
	pipe.LPush(ctx, targetQueue, string(jobJSON))

	// Remove from DLQ
	pipe.LRem(ctx, ae.config.RedisStreamKey, 1, job.ID)

	// Execute pipeline
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to requeue job: %w", err)
	}

	ae.logger.Info("Job requeued successfully",
		zap.String("job_id", job.JobID),
		zap.String("target_queue", targetQueue))

	return nil
}

// executeTransform transforms the job payload
func (ae *ActionExecutor) executeTransform(ctx context.Context, job *DLQJob, params map[string]interface{}, dryRun bool) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload for transformation: %w", err)
	}

	originalPayload := make(map[string]interface{})
	for k, v := range payload {
		originalPayload[k] = v
	}

	// Apply transformations
	if setFields, ok := params["set"].(map[string]interface{}); ok {
		for path, value := range setFields {
			ae.setPayloadField(payload, path, value)
		}
	}

	if removeFields, ok := params["remove"].([]interface{}); ok {
		for _, field := range removeFields {
			if fieldStr, ok := field.(string); ok {
				ae.removePayloadField(payload, fieldStr)
			}
		}
	}

	if addFields, ok := params["add_if_missing"].(map[string]interface{}); ok {
		for path, value := range addFields {
			if !ae.payloadFieldExists(payload, path) {
				ae.setPayloadField(payload, path, value)
			}
		}
	}

	if dryRun {
		transformedJSON, _ := json.Marshal(payload)
		ae.logger.Info("DRY RUN: Would transform payload",
			zap.String("job_id", job.JobID),
			zap.String("original", string(job.Payload)),
			zap.String("transformed", string(transformedJSON)))
		return nil
	}

	// Update job payload
	transformedJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal transformed payload: %w", err)
	}

	job.Payload = transformedJSON
	job.PayloadSize = int64(len(transformedJSON))

	ae.logger.Info("Payload transformed successfully",
		zap.String("job_id", job.JobID))

	return nil
}

// executeRedact redacts sensitive fields from the job payload
func (ae *ActionExecutor) executeRedact(ctx context.Context, job *DLQJob, params map[string]interface{}, dryRun bool) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(job.Payload, &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload for redaction: %w", err)
	}

	fieldsToRedact, ok := params["fields"].([]interface{})
	if !ok {
		return fmt.Errorf("redact action requires 'fields' parameter")
	}

	replacement := ae.getStringParam(params, "replacement", "[REDACTED]")

	if dryRun {
		ae.logger.Info("DRY RUN: Would redact fields",
			zap.String("job_id", job.JobID),
			zap.Any("fields", fieldsToRedact),
			zap.String("replacement", replacement))
		return nil
	}

	// Redact specified fields
	for _, field := range fieldsToRedact {
		if fieldStr, ok := field.(string); ok {
			ae.redactPayloadField(payload, fieldStr, replacement)
		}
	}

	// Update job payload
	redactedJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal redacted payload: %w", err)
	}

	job.Payload = redactedJSON
	job.PayloadSize = int64(len(redactedJSON))

	ae.logger.Info("Payload redacted successfully",
		zap.String("job_id", job.JobID),
		zap.Any("redacted_fields", fieldsToRedact))

	return nil
}

// executeDrop drops the job permanently
func (ae *ActionExecutor) executeDrop(ctx context.Context, job *DLQJob, params map[string]interface{}, dryRun bool) error {
	reason := ae.getStringParam(params, "reason", "Dropped by remediation rule")
	retainForAudit := ae.getBoolParam(params, "retain_for_audit", true)

	if dryRun {
		ae.logger.Info("DRY RUN: Would drop job",
			zap.String("job_id", job.JobID),
			zap.String("reason", reason))
		return nil
	}

	// Store in dropped jobs archive if retention is enabled
	if retainForAudit {
		droppedJob := map[string]interface{}{
			"job_id":     job.JobID,
			"queue":      job.Queue,
			"job_type":   job.JobType,
			"error":      job.Error,
			"payload":    string(job.Payload),
			"reason":     reason,
			"dropped_at": time.Now(),
		}

		droppedJSON, _ := json.Marshal(droppedJob)
		ae.redis.LPush(ctx, "dlq:dropped", string(droppedJSON))
		ae.redis.Expire(ctx, "dlq:dropped", 30*24*time.Hour) // 30 days retention
	}

	// Remove from DLQ
	err := ae.redis.LRem(ctx, ae.config.RedisStreamKey, 1, job.ID).Err()
	if err != nil {
		return fmt.Errorf("failed to remove job from DLQ: %w", err)
	}

	ae.logger.Info("Job dropped successfully",
		zap.String("job_id", job.JobID),
		zap.String("reason", reason))

	return nil
}

// executeRoute routes the job to a different queue based on conditions
func (ae *ActionExecutor) executeRoute(ctx context.Context, job *DLQJob, params map[string]interface{}, dryRun bool) error {
	rules, ok := params["rules"].([]interface{})
	if !ok {
		return fmt.Errorf("route action requires 'rules' parameter")
	}

	defaultQueue := ae.getStringParam(params, "default_queue", job.Queue)

	for _, rule := range rules {
		if ruleMap, ok := rule.(map[string]interface{}); ok {
			condition := ae.getStringParam(ruleMap, "condition", "")
			targetQueue := ae.getStringParam(ruleMap, "target_queue", "")

			if condition != "" && targetQueue != "" && ae.evaluateRoutingCondition(job, condition) {
				return ae.executeRequeue(ctx, job, map[string]interface{}{
					"target_queue": targetQueue,
				}, dryRun)
			}
		}
	}

	// Use default queue if no rules match
	return ae.executeRequeue(ctx, job, map[string]interface{}{
		"target_queue": defaultQueue,
	}, dryRun)
}

// executeDelay delays the job by adding it to a delayed queue
func (ae *ActionExecutor) executeDelay(ctx context.Context, job *DLQJob, params map[string]interface{}, dryRun bool) error {
	delay := ae.getDurationParam(params, "delay", 5*time.Minute)

	return ae.executeRequeue(ctx, job, map[string]interface{}{
		"target_queue": job.Queue,
		"delay":        delay,
	}, dryRun)
}

// executeTag adds tags to the job metadata
func (ae *ActionExecutor) executeTag(ctx context.Context, job *DLQJob, params map[string]interface{}, dryRun bool) error {
	tags, ok := params["tags"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("tag action requires 'tags' parameter")
	}

	if dryRun {
		ae.logger.Info("DRY RUN: Would add tags to job",
			zap.String("job_id", job.JobID),
			zap.Any("tags", tags))
		return nil
	}

	// Add tags to metadata
	if job.Metadata == nil {
		job.Metadata = make(map[string]interface{})
	}

	for key, value := range tags {
		job.Metadata[key] = value
	}

	ae.logger.Info("Tags added successfully",
		zap.String("job_id", job.JobID),
		zap.Any("tags", tags))

	return nil
}

// executeNotify sends notifications about the job
func (ae *ActionExecutor) executeNotify(ctx context.Context, job *DLQJob, params map[string]interface{}, dryRun bool) error {
	channels, ok := params["channels"].([]interface{})
	if !ok {
		return fmt.Errorf("notify action requires 'channels' parameter")
	}

	message := ae.getStringParam(params, "message", fmt.Sprintf("Job %s processed by remediation pipeline", job.JobID))

	if dryRun {
		ae.logger.Info("DRY RUN: Would send notifications",
			zap.String("job_id", job.JobID),
			zap.Any("channels", channels),
			zap.String("message", message))
		return nil
	}

	// Store notification in Redis for processing by notification service
	notification := map[string]interface{}{
		"job_id":    job.JobID,
		"message":   message,
		"channels":  channels,
		"timestamp": time.Now(),
	}

	notificationJSON, _ := json.Marshal(notification)
	err := ae.redis.LPush(ctx, "notifications:queue", string(notificationJSON)).Err()
	if err != nil {
		return fmt.Errorf("failed to queue notification: %w", err)
	}

	ae.logger.Info("Notification queued successfully",
		zap.String("job_id", job.JobID),
		zap.Any("channels", channels))

	return nil
}

// Helper methods for parameter extraction

func (ae *ActionExecutor) getStringParam(params map[string]interface{}, key, defaultValue string) string {
	if value, ok := params[key].(string); ok {
		return value
	}
	return defaultValue
}

func (ae *ActionExecutor) getIntParam(params map[string]interface{}, key string, defaultValue int) int {
	if value, ok := params[key].(float64); ok {
		return int(value)
	}
	if value, ok := params[key].(int); ok {
		return value
	}
	return defaultValue
}

func (ae *ActionExecutor) getBoolParam(params map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := params[key].(bool); ok {
		return value
	}
	return defaultValue
}

func (ae *ActionExecutor) getDurationParam(params map[string]interface{}, key string, defaultValue time.Duration) time.Duration {
	if value, ok := params[key].(string); ok {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// Helper methods for payload manipulation

func (ae *ActionExecutor) setPayloadField(payload map[string]interface{}, path string, value interface{}) {
	keys := strings.Split(path, ".")
	current := payload

	for i, key := range keys[:len(keys)-1] {
		if _, exists := current[key]; !exists {
			current[key] = make(map[string]interface{})
		}
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			// Cannot traverse further, create new map
			newMap := make(map[string]interface{})
			current[key] = newMap
			current = newMap
		}
	}

	current[keys[len(keys)-1]] = value
}

func (ae *ActionExecutor) removePayloadField(payload map[string]interface{}, path string) {
	keys := strings.Split(path, ".")
	current := payload

	for _, key := range keys[:len(keys)-1] {
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return // Path doesn't exist
		}
	}

	delete(current, keys[len(keys)-1])
}

func (ae *ActionExecutor) payloadFieldExists(payload map[string]interface{}, path string) bool {
	keys := strings.Split(path, ".")
	current := payload

	for _, key := range keys {
		if value, exists := current[key]; exists {
			if next, ok := value.(map[string]interface{}); ok {
				current = next
			} else {
				// Last key should exist as a value
				return len(keys) == 1
			}
		} else {
			return false
		}
	}

	return true
}

func (ae *ActionExecutor) redactPayloadField(payload map[string]interface{}, path, replacement string) {
	keys := strings.Split(path, ".")
	current := payload

	for _, key := range keys[:len(keys)-1] {
		if next, ok := current[key].(map[string]interface{}); ok {
			current = next
		} else {
			return // Path doesn't exist
		}
	}

	if _, exists := current[keys[len(keys)-1]]; exists {
		current[keys[len(keys)-1]] = replacement
	}
}

func (ae *ActionExecutor) evaluateRoutingCondition(job *DLQJob, condition string) bool {
	// Simple condition evaluation for routing
	// Could be extended with more complex logic
	if strings.Contains(condition, "error_type") {
		parts := strings.Split(condition, "=")
		if len(parts) == 2 {
			expectedType := strings.TrimSpace(strings.Trim(parts[1], "\"'"))
			return job.ErrorType == expectedType
		}
	}

	if strings.Contains(condition, "retry_count") {
		if strings.Contains(condition, ">") {
			parts := strings.Split(condition, ">")
			if len(parts) == 2 {
				if threshold, err := strconv.Atoi(strings.TrimSpace(parts[1])); err == nil {
					return job.RetryCount > threshold
				}
			}
		}
	}

	return false
}

func (ae *ActionExecutor) copyJob(job *DLQJob) *DLQJob {
	copied := *job

	// Deep copy metadata
	if job.Metadata != nil {
		copied.Metadata = make(map[string]interface{})
		for k, v := range job.Metadata {
			copied.Metadata[k] = v
		}
	}

	// Copy payload
	if job.Payload != nil {
		copied.Payload = make([]byte, len(job.Payload))
		copy(copied.Payload, job.Payload)
	}

	return &copied
}