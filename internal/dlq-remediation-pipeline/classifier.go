// Copyright 2025 James Ross
package dlqremediation

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PaesslerAG/jsonpath"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ClassificationEngine classifies DLQ jobs using rules and optional external service
type ClassificationEngine struct {
	redis      *redis.Client
	logger     *zap.Logger
	config     *PipelineConfig
	httpClient *http.Client
	cache      map[string]*Classification
}

// NewClassificationEngine creates a new classification engine
func NewClassificationEngine(redisClient *redis.Client, config *PipelineConfig, logger *zap.Logger) *ClassificationEngine {
	return &ClassificationEngine{
		redis:  redisClient,
		logger: logger,
		config: config,
		httpClient: &http.Client{
			Timeout: config.ExternalClassifier.Timeout,
		},
		cache: make(map[string]*Classification),
	}
}

// Classify classifies a DLQ job against all enabled rules
func (ce *ClassificationEngine) Classify(ctx context.Context, job *DLQJob, rules []RemediationRule) (*Classification, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s:%d", job.JobID, job.Error, job.RetryCount)
	if cached, exists := ce.cache[cacheKey]; exists {
		ce.logger.Debug("Using cached classification", zap.String("job_id", job.JobID))
		return cached, nil
	}

	// Try external classifier first if enabled
	if ce.config.ExternalClassifier.Enabled {
		if classification, err := ce.classifyExternal(ctx, job); err == nil {
			ce.cacheClassification(cacheKey, classification)
			return classification, nil
		} else {
			ce.logger.Warn("External classification failed, falling back to rules",
				zap.String("job_id", job.JobID), zap.Error(err))
		}
	}

	// Classify using rules
	classification := ce.classifyWithRules(ctx, job, rules)
	ce.cacheClassification(cacheKey, classification)

	return classification, nil
}

// classifyExternal uses external classification service
func (ce *ClassificationEngine) classifyExternal(ctx context.Context, job *DLQJob) (*Classification, error) {
	request := ClassificationRequest{
		JobID:      job.JobID,
		Error:      job.Error,
		ErrorType:  job.ErrorType,
		Payload:    job.Payload,
		Queue:      job.Queue,
		JobType:    job.JobType,
		RetryCount: job.RetryCount,
		Metadata:   job.Metadata,
		FailedAt:   job.FailedAt,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal classification request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ce.config.ExternalClassifier.Endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range ce.config.ExternalClassifier.Headers {
		req.Header.Set(key, value)
	}

	resp, err := ce.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("external classifier request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("external classifier returned status %d", resp.StatusCode)
	}

	var response ClassificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode classification response: %w", err)
	}

	classification := &Classification{
		JobID:      job.JobID,
		Category:   response.Category,
		Confidence: response.Confidence,
		Actions:    response.Actions,
		Reason:     response.Reason,
		Metadata:   response.Metadata,
		Timestamp:  time.Now(),
	}

	return classification, nil
}

// classifyWithRules classifies job using configured rules
func (ce *ClassificationEngine) classifyWithRules(ctx context.Context, job *DLQJob, rules []RemediationRule) *Classification {
	var bestMatch *RemediationRule
	var bestConfidence float64

	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		confidence := ce.calculateMatchConfidence(job, &rule)
		if confidence > bestConfidence {
			bestConfidence = confidence
			bestMatch = &rule
		}
	}

	if bestMatch == nil {
		return &Classification{
			JobID:      job.JobID,
			Category:   "unclassified",
			Confidence: 0.0,
			Actions:    []string{},
			Reason:     "No matching rules found",
			Timestamp:  time.Now(),
		}
	}

	actions := make([]string, len(bestMatch.Actions))
	for i, action := range bestMatch.Actions {
		actions[i] = string(action.Type)
	}

	return &Classification{
		JobID:      job.JobID,
		Category:   bestMatch.Name,
		Confidence: bestConfidence,
		RuleID:     bestMatch.ID,
		Actions:    actions,
		Reason:     fmt.Sprintf("Matched rule '%s' with confidence %.2f", bestMatch.Name, bestConfidence),
		Timestamp:  time.Now(),
	}
}

// calculateMatchConfidence calculates how well a job matches a rule
func (ce *ClassificationEngine) calculateMatchConfidence(job *DLQJob, rule *RemediationRule) float64 {
	var totalWeight float64
	var matchedWeight float64

	matcher := rule.Matcher

	// Error pattern matching
	if matcher.ErrorPattern != "" {
		totalWeight += 3.0
		if ce.matchErrorPattern(job.Error, matcher.ErrorPattern) {
			matchedWeight += 3.0
		}
	}

	// Error type matching
	if matcher.ErrorType != "" {
		totalWeight += 2.0
		if job.ErrorType == matcher.ErrorType {
			matchedWeight += 2.0
		}
	}

	// Job type matching
	if matcher.JobType != "" {
		totalWeight += 2.0
		if ce.matchPattern(job.JobType, matcher.JobType) {
			matchedWeight += 2.0
		}
	}

	// Source queue matching
	if matcher.SourceQueue != "" {
		totalWeight += 2.0
		if ce.matchPattern(job.Queue, matcher.SourceQueue) {
			matchedWeight += 2.0
		}
	}

	// Retry count matching
	if matcher.RetryCount != "" {
		totalWeight += 1.5
		if ce.matchNumericCondition(job.RetryCount, matcher.RetryCount) {
			matchedWeight += 1.5
		}
	}

	// Payload size matching
	if matcher.PayloadSize != "" {
		totalWeight += 1.0
		if ce.matchSizeCondition(job.PayloadSize, matcher.PayloadSize) {
			matchedWeight += 1.0
		}
	}

	// Age threshold matching
	if matcher.AgeThreshold != "" {
		totalWeight += 1.0
		age := time.Since(job.FailedAt)
		if ce.matchDurationCondition(age, matcher.AgeThreshold) {
			matchedWeight += 1.0
		}
	}

	// Payload matchers
	for _, payloadMatcher := range matcher.PayloadMatchers {
		totalWeight += 1.5
		if ce.matchPayloadField(job.Payload, payloadMatcher) {
			matchedWeight += 1.5
		}
	}

	// Metadata filters
	for key, value := range matcher.MetadataFilters {
		totalWeight += 0.5
		if metaValue, exists := job.Metadata[key]; exists {
			if fmt.Sprintf("%v", metaValue) == value {
				matchedWeight += 0.5
			}
		}
	}

	// Time pattern matching
	if matcher.TimePattern != "" {
		totalWeight += 0.5
		if ce.matchTimePattern(job.FailedAt, matcher.TimePattern) {
			matchedWeight += 0.5
		}
	}

	if totalWeight == 0 {
		return 0.0
	}

	return matchedWeight / totalWeight
}

// matchErrorPattern matches error message against pattern (regex or substring)
func (ce *ClassificationEngine) matchErrorPattern(errorMsg, pattern string) bool {
	// Try as regex first
	if regex, err := regexp.Compile(pattern); err == nil {
		return regex.MatchString(errorMsg)
	}

	// Fall back to substring matching
	return strings.Contains(strings.ToLower(errorMsg), strings.ToLower(pattern))
}

// matchPattern matches string against pattern (supports wildcards)
func (ce *ClassificationEngine) matchPattern(value, pattern string) bool {
	if strings.Contains(pattern, "*") {
		// Convert wildcard pattern to regex
		regexPattern := strings.ReplaceAll(regexp.QuoteMeta(pattern), `\*`, `.*`)
		if regex, err := regexp.Compile("^" + regexPattern + "$"); err == nil {
			return regex.MatchString(value)
		}
	}
	return value == pattern
}

// matchNumericCondition matches numeric values against conditions like "> 3", "= 0", "< 5"
func (ce *ClassificationEngine) matchNumericCondition(value int, condition string) bool {
	condition = strings.TrimSpace(condition)

	if strings.HasPrefix(condition, ">=") {
		if threshold, err := strconv.Atoi(strings.TrimSpace(condition[2:])); err == nil {
			return value >= threshold
		}
	} else if strings.HasPrefix(condition, "<=") {
		if threshold, err := strconv.Atoi(strings.TrimSpace(condition[2:])); err == nil {
			return value <= threshold
		}
	} else if strings.HasPrefix(condition, ">") {
		if threshold, err := strconv.Atoi(strings.TrimSpace(condition[1:])); err == nil {
			return value > threshold
		}
	} else if strings.HasPrefix(condition, "<") {
		if threshold, err := strconv.Atoi(strings.TrimSpace(condition[1:])); err == nil {
			return value < threshold
		}
	} else if strings.HasPrefix(condition, "=") {
		if threshold, err := strconv.Atoi(strings.TrimSpace(condition[1:])); err == nil {
			return value == threshold
		}
	} else {
		// Try direct equality
		if threshold, err := strconv.Atoi(condition); err == nil {
			return value == threshold
		}
	}

	return false
}

// matchSizeCondition matches size values against conditions like "> 1MB", "< 100KB"
func (ce *ClassificationEngine) matchSizeCondition(value int64, condition string) bool {
	condition = strings.TrimSpace(condition)

	var operator string
	var thresholdStr string

	if strings.HasPrefix(condition, ">=") {
		operator = ">="
		thresholdStr = strings.TrimSpace(condition[2:])
	} else if strings.HasPrefix(condition, "<=") {
		operator = "<="
		thresholdStr = strings.TrimSpace(condition[2:])
	} else if strings.HasPrefix(condition, ">") {
		operator = ">"
		thresholdStr = strings.TrimSpace(condition[1:])
	} else if strings.HasPrefix(condition, "<") {
		operator = "<"
		thresholdStr = strings.TrimSpace(condition[1:])
	} else if strings.HasPrefix(condition, "=") {
		operator = "="
		thresholdStr = strings.TrimSpace(condition[1:])
	} else {
		operator = "="
		thresholdStr = condition
	}

	threshold := ce.parseSize(thresholdStr)
	if threshold < 0 {
		return false
	}

	switch operator {
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case "=":
		return value == threshold
	default:
		return false
	}
}

// parseSize parses size strings like "1MB", "100KB", "1GB"
func (ce *ClassificationEngine) parseSize(sizeStr string) int64 {
	sizeStr = strings.ToUpper(strings.TrimSpace(sizeStr))

	multipliers := map[string]int64{
		"B":  1,
		"KB": 1024,
		"MB": 1024 * 1024,
		"GB": 1024 * 1024 * 1024,
	}

	for suffix, multiplier := range multipliers {
		if strings.HasSuffix(sizeStr, suffix) {
			numStr := strings.TrimSuffix(sizeStr, suffix)
			if num, err := strconv.ParseInt(numStr, 10, 64); err == nil {
				return num * multiplier
			}
		}
	}

	// Try parsing as raw number
	if num, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
		return num
	}

	return -1
}

// matchDurationCondition matches duration against conditions like "> 1h", "< 30m"
func (ce *ClassificationEngine) matchDurationCondition(value time.Duration, condition string) bool {
	condition = strings.TrimSpace(condition)

	var operator string
	var thresholdStr string

	if strings.HasPrefix(condition, ">=") {
		operator = ">="
		thresholdStr = strings.TrimSpace(condition[2:])
	} else if strings.HasPrefix(condition, "<=") {
		operator = "<="
		thresholdStr = strings.TrimSpace(condition[2:])
	} else if strings.HasPrefix(condition, ">") {
		operator = ">"
		thresholdStr = strings.TrimSpace(condition[1:])
	} else if strings.HasPrefix(condition, "<") {
		operator = "<"
		thresholdStr = strings.TrimSpace(condition[1:])
	} else if strings.HasPrefix(condition, "=") {
		operator = "="
		thresholdStr = strings.TrimSpace(condition[1:])
	} else {
		operator = "="
		thresholdStr = condition
	}

	threshold, err := time.ParseDuration(thresholdStr)
	if err != nil {
		return false
	}

	switch operator {
	case ">=":
		return value >= threshold
	case "<=":
		return value <= threshold
	case ">":
		return value > threshold
	case "<":
		return value < threshold
	case "=":
		return value == threshold
	default:
		return false
	}
}

// matchPayloadField matches specific fields in job payload using JSONPath
func (ce *ClassificationEngine) matchPayloadField(payload json.RawMessage, matcher PayloadMatcher) bool {
	var data interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return false
	}

	value, err := jsonpath.Get(matcher.JSONPath, data)
	if err != nil {
		return matcher.Operator == "not_exists"
	}

	switch matcher.Operator {
	case "exists":
		return true
	case "not_exists":
		return false
	case "equals":
		return ce.compareValues(value, matcher.Value, matcher.CaseInsensitive)
	case "contains":
		valueStr := fmt.Sprintf("%v", value)
		searchStr := fmt.Sprintf("%v", matcher.Value)
		if matcher.CaseInsensitive {
			valueStr = strings.ToLower(valueStr)
			searchStr = strings.ToLower(searchStr)
		}
		return strings.Contains(valueStr, searchStr)
	case "regex":
		valueStr := fmt.Sprintf("%v", value)
		patternStr := fmt.Sprintf("%v", matcher.Value)
		if regex, err := regexp.Compile(patternStr); err == nil {
			return regex.MatchString(valueStr)
		}
		return false
	case "gt", "lt":
		return ce.compareNumeric(value, matcher.Value, matcher.Operator)
	default:
		return false
	}
}

// compareValues compares two values for equality
func (ce *ClassificationEngine) compareValues(a, b interface{}, caseInsensitive bool) bool {
	if caseInsensitive {
		aStr := strings.ToLower(fmt.Sprintf("%v", a))
		bStr := strings.ToLower(fmt.Sprintf("%v", b))
		return aStr == bStr
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// compareNumeric compares numeric values
func (ce *ClassificationEngine) compareNumeric(a, b interface{}, operator string) bool {
	aFloat, aOk := ce.toFloat64(a)
	bFloat, bOk := ce.toFloat64(b)

	if !aOk || !bOk {
		return false
	}

	switch operator {
	case "gt":
		return aFloat > bFloat
	case "lt":
		return aFloat < bFloat
	default:
		return false
	}
}

// toFloat64 converts interface{} to float64
func (ce *ClassificationEngine) toFloat64(val interface{}) (float64, bool) {
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

// matchTimePattern matches time against patterns like "business_hours", "weekends"
func (ce *ClassificationEngine) matchTimePattern(t time.Time, pattern string) bool {
	switch strings.ToLower(pattern) {
	case "business_hours":
		weekday := t.Weekday()
		hour := t.Hour()
		return weekday >= time.Monday && weekday <= time.Friday && hour >= 9 && hour <= 17
	case "weekends":
		weekday := t.Weekday()
		return weekday == time.Saturday || weekday == time.Sunday
	case "nights":
		hour := t.Hour()
		return hour >= 22 || hour <= 6
	case "peak_hours":
		hour := t.Hour()
		return hour >= 9 && hour <= 12 || hour >= 14 && hour <= 17
	default:
		return false
	}
}

// cacheClassification stores classification in cache
func (ce *ClassificationEngine) cacheClassification(key string, classification *Classification) {
	ce.cache[key] = classification

	// Store in Redis if enabled
	if ce.config.ExternalClassifier.CacheTTL > 0 {
		data, err := json.Marshal(classification)
		if err == nil {
			cacheKey := fmt.Sprintf("classification:%s", key)
			ce.redis.Set(context.Background(), cacheKey, data, ce.config.ExternalClassifier.CacheTTL)
		}
	}
}

// GetCachedClassification retrieves classification from cache
func (ce *ClassificationEngine) GetCachedClassification(ctx context.Context, key string) (*Classification, bool) {
	// Check local cache first
	if classification, exists := ce.cache[key]; exists {
		return classification, true
	}

	// Check Redis cache
	cacheKey := fmt.Sprintf("classification:%s", key)
	data, err := ce.redis.Get(ctx, cacheKey).Result()
	if err != nil {
		return nil, false
	}

	var classification Classification
	if err := json.Unmarshal([]byte(data), &classification); err != nil {
		return nil, false
	}

	// Store in local cache
	ce.cache[key] = &classification
	return &classification, true
}

// ClearCache clears the classification cache
func (ce *ClassificationEngine) ClearCache() {
	ce.cache = make(map[string]*Classification)
}
