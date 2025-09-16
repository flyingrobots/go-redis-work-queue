package dlqremediationui

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

type PatternAnalyzerImpl struct {
	cache     map[string][]ErrorPattern
	cacheMu   sync.RWMutex
	cacheSize int
	logger    *slog.Logger
}

func NewPatternAnalyzer(cacheSize int, logger *slog.Logger) *PatternAnalyzerImpl {
	if cacheSize <= 0 {
		cacheSize = 100
	}

	return &PatternAnalyzerImpl{
		cache:     make(map[string][]ErrorPattern),
		cacheSize: cacheSize,
		logger:    logger,
	}
}

func (p *PatternAnalyzerImpl) AnalyzeEntries(ctx context.Context, entries []DLQEntry) ([]ErrorPattern, error) {
	if len(entries) == 0 {
		return []ErrorPattern{}, nil
	}

	cacheKey := p.generateCacheKey(entries)

	p.cacheMu.RLock()
	if patterns, exists := p.cache[cacheKey]; exists {
		p.cacheMu.RUnlock()
		p.logger.Debug("Returning cached patterns", "count", len(patterns))
		return patterns, nil
	}
	p.cacheMu.RUnlock()

	patterns := p.analyzePatterns(entries)

	p.cacheMu.Lock()
	if len(p.cache) >= p.cacheSize {
		p.evictOldestCacheEntry()
	}
	p.cache[cacheKey] = patterns
	p.cacheMu.Unlock()

	p.logger.Debug("Analyzed error patterns", "count", len(patterns))
	return patterns, nil
}

func (p *PatternAnalyzerImpl) GetSimilarEntries(ctx context.Context, entry DLQEntry, threshold float64) ([]DLQEntry, error) {
	return []DLQEntry{}, nil
}

func (p *PatternAnalyzerImpl) analyzePatterns(entries []DLQEntry) []ErrorPattern {
	errorGroups := make(map[string][]DLQEntry)

	for _, entry := range entries {
		signature := p.generateErrorSignature(entry)
		errorGroups[signature] = append(errorGroups[signature], entry)
	}

	patterns := make([]ErrorPattern, 0)
	for signature, groupEntries := range errorGroups {
		if len(groupEntries) < 2 {
			continue
		}

		pattern := p.createPatternFromGroup(signature, groupEntries)
		patterns = append(patterns, pattern)
	}

	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Count > patterns[j].Count
	})

	return patterns
}

func (p *PatternAnalyzerImpl) generateErrorSignature(entry DLQEntry) string {
	signature := fmt.Sprintf("%s|%s|%s",
		entry.Queue,
		entry.Type,
		p.normalizeErrorMessage(entry.Error.Message),
	)

	if entry.Error.Code != "" {
		signature += "|" + entry.Error.Code
	}

	return signature
}

func (p *PatternAnalyzerImpl) normalizeErrorMessage(message string) string {
	patterns := []struct {
		regex       *regexp.Regexp
		replacement string
	}{
		{regexp.MustCompile(`\d+`), "N"},
		{regexp.MustCompile(`0x[a-fA-F0-9]+`), "ADDR"},
		{regexp.MustCompile(`[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}`), "UUID"},
		{regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z?\b`), "TIMESTAMP"},
		{regexp.MustCompile(`"[^"]*"`), "STRING"},
		{regexp.MustCompile(`'[^']*'`), "STRING"},
	}

	normalized := strings.ToLower(message)

	for _, pattern := range patterns {
		normalized = pattern.regex.ReplaceAllString(normalized, pattern.replacement)
	}

	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")
	normalized = strings.TrimSpace(normalized)

	return normalized
}

func (p *PatternAnalyzerImpl) createPatternFromGroup(signature string, entries []DLQEntry) ErrorPattern {
	if len(entries) == 0 {
		return ErrorPattern{}
	}

	first := entries[0]
	last := entries[len(entries)-1]

	sampleIDs := make([]string, 0, min(5, len(entries)))
	for i, entry := range entries {
		if i < 5 {
			sampleIDs = append(sampleIDs, entry.ID)
		}
	}

	queues := make(map[string]bool)
	types := make(map[string]bool)
	for _, entry := range entries {
		queues[entry.Queue] = true
		types[entry.Type] = true
	}

	affectedQueues := make([]string, 0, len(queues))
	for queue := range queues {
		affectedQueues = append(affectedQueues, queue)
	}

	affectedTypes := make([]string, 0, len(types))
	for jobType := range types {
		affectedTypes = append(affectedTypes, jobType)
	}

	return ErrorPattern{
		ID:              signature,
		Pattern:         p.normalizeErrorMessage(first.Error.Message),
		Message:         first.Error.Message,
		Count:           len(entries),
		FirstSeen:       first.FailedAt,
		LastSeen:        last.FailedAt,
		AffectedQueues:  affectedQueues,
		AffectedTypes:   affectedTypes,
		SampleEntryIDs:  sampleIDs,
		Severity:        p.calculateSeverity(len(entries)),
		SuggestedAction: p.suggestAction(first, len(entries)),
	}
}

func (p *PatternAnalyzerImpl) calculateSeverity(count int) PatternSeverity {
	if count >= 100 {
		return SeverityCritical
	}
	if count >= 50 {
		return SeverityHigh
	}
	if count >= 10 {
		return SeverityMedium
	}
	return SeverityLow
}

func (p *PatternAnalyzerImpl) suggestAction(entry DLQEntry, count int) string {
	errorMsg := strings.ToLower(entry.Error.Message)

	if strings.Contains(errorMsg, "timeout") || strings.Contains(errorMsg, "deadline") {
		return "Consider increasing timeout values or investigating network latency"
	}

	if strings.Contains(errorMsg, "connection") || strings.Contains(errorMsg, "refused") {
		return "Check service connectivity and health"
	}

	if strings.Contains(errorMsg, "unauthorized") || strings.Contains(errorMsg, "forbidden") {
		return "Verify authentication credentials and permissions"
	}

	if strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "404") {
		return "Check if required resources exist"
	}

	if strings.Contains(errorMsg, "validation") || strings.Contains(errorMsg, "invalid") {
		return "Review input validation and data format"
	}

	if count >= 50 {
		return "High frequency error - investigate root cause immediately"
	}

	if count >= 10 {
		return "Recurring error - consider pattern analysis"
	}

	return "Monitor for recurrence"
}

func (p *PatternAnalyzerImpl) generateCacheKey(entries []DLQEntry) string {
	if len(entries) == 0 {
		return ""
	}

	hasher := sha256.New()
	for _, entry := range entries {
		hasher.Write([]byte(entry.ID))
		hasher.Write([]byte(entry.Error.Message))
	}

	return fmt.Sprintf("%x", hasher.Sum(nil))[:16]
}

func (p *PatternAnalyzerImpl) evictOldestCacheEntry() {
	if len(p.cache) == 0 {
		return
	}

	var oldestKey string
	for key := range p.cache {
		oldestKey = key
		break
	}

	delete(p.cache, oldestKey)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type RemediationEngineImpl struct {
	redis  *redis.Client
	logger *slog.Logger
}

func NewRemediationEngine(redisClient *redis.Client, logger *slog.Logger) *RemediationEngineImpl {
	return &RemediationEngineImpl{
		redis:  redisClient,
		logger: logger,
	}
}

func (r *RemediationEngineImpl) Requeue(ctx context.Context, entryID string) error {
	r.logger.Debug("Requeuing DLQ entry", "id", entryID)

	dlqKey := "dlq:entries"

	entryData, err := r.redis.HGet(ctx, dlqKey, entryID).Result()
	if err != nil {
		return fmt.Errorf("failed to get DLQ entry for requeue: %w", err)
	}

	var entry DLQEntry
	if err := json.Unmarshal([]byte(entryData), &entry); err != nil {
		return fmt.Errorf("failed to unmarshal DLQ entry: %w", err)
	}

	queueKey := fmt.Sprintf("queue:%s", entry.Queue)

	jobData := map[string]interface{}{
		"id":      entry.JobID,
		"type":    entry.Type,
		"payload": entry.Payload,
		"metadata": map[string]interface{}{
			"requeued_from_dlq": true,
			"original_dlq_id":   entry.ID,
			"requeued_at":       time.Now().Format(time.RFC3339),
			"attempt_count":     len(entry.Attempts),
		},
	}

	jobJSON, err := json.Marshal(jobData)
	if err != nil {
		return fmt.Errorf("failed to marshal job for requeue: %w", err)
	}

	err = r.redis.LPush(ctx, queueKey, jobJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to push job to queue: %w", err)
	}

	err = r.redis.HDel(ctx, dlqKey, entryID).Err()
	if err != nil {
		r.logger.Warn("Failed to remove entry from DLQ after requeue", "id", entryID, "error", err)
	}

	r.logger.Info("DLQ entry requeued successfully", "id", entryID, "queue", entry.Queue)
	return nil
}

func (r *RemediationEngineImpl) Purge(ctx context.Context, entryID string) error {
	r.logger.Debug("Purging DLQ entry", "id", entryID)

	dlqKey := "dlq:entries"

	err := r.redis.HDel(ctx, dlqKey, entryID).Err()
	if err != nil {
		return fmt.Errorf("failed to purge DLQ entry: %w", err)
	}

	r.logger.Info("DLQ entry purged successfully", "id", entryID)
	return nil
}

func (r *RemediationEngineImpl) BulkRequeue(ctx context.Context, entryIDs []string) (*BulkOperationResult, error) {
	result := &BulkOperationResult{
		TotalRequested: len(entryIDs),
		Successful:     make([]string, 0),
		Failed:         make([]OperationError, 0),
		StartedAt:      time.Now(),
	}

	for _, id := range entryIDs {
		err := r.Requeue(ctx, id)
		if err != nil {
			result.Failed = append(result.Failed, OperationError{
				ID:    id,
				Error: err.Error(),
			})
		} else {
			result.Successful = append(result.Successful, id)
		}
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	return result, nil
}

func (r *RemediationEngineImpl) BulkPurge(ctx context.Context, entryIDs []string) (*BulkOperationResult, error) {
	result := &BulkOperationResult{
		TotalRequested: len(entryIDs),
		Successful:     make([]string, 0),
		Failed:         make([]OperationError, 0),
		StartedAt:      time.Now(),
	}

	for _, id := range entryIDs {
		err := r.Purge(ctx, id)
		if err != nil {
			result.Failed = append(result.Failed, OperationError{
				ID:    id,
				Error: err.Error(),
			})
		} else {
			result.Successful = append(result.Successful, id)
		}
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	return result, nil
}