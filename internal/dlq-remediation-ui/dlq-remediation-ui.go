package dlqremediationui

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type DLQManagerImpl struct {
	redis     *redis.Client
	storage   DLQStorage
	analyzer  PatternAnalyzer
	engine    RemediationEngine
	logger    *slog.Logger
	config    Config
}

type Config struct {
	MaxPageSize         int           `json:"max_page_size"`
	DefaultPageSize     int           `json:"default_page_size"`
	PeekTimeout         time.Duration `json:"peek_timeout"`
	BulkOperationLimit  int           `json:"bulk_operation_limit"`
	PatternCacheSize    int           `json:"pattern_cache_size"`
	EnableMetrics       bool          `json:"enable_metrics"`
	MetricsPrefix       string        `json:"metrics_prefix"`
}

func NewDLQManager(redisClient *redis.Client, config Config, logger *slog.Logger) *DLQManagerImpl {
	if config.MaxPageSize == 0 {
		config.MaxPageSize = 1000
	}
	if config.DefaultPageSize == 0 {
		config.DefaultPageSize = 50
	}
	if config.PeekTimeout == 0 {
		config.PeekTimeout = 5 * time.Second
	}
	if config.BulkOperationLimit == 0 {
		config.BulkOperationLimit = 100
	}

	storage := NewRedisStorage(redisClient, logger)
	analyzer := NewPatternAnalyzer(config.PatternCacheSize, logger)
	engine := NewRemediationEngine(redisClient, logger)

	return &DLQManagerImpl{
		redis:    redisClient,
		storage:  storage,
		analyzer: analyzer,
		engine:   engine,
		logger:   logger,
		config:   config,
	}
}

func (d *DLQManagerImpl) ListEntries(ctx context.Context, filter DLQFilter, pagination PaginationRequest) (*DLQListResponse, error) {
	d.logger.Debug("Listing DLQ entries",
		"filter", filter,
		"pagination", pagination)

	if pagination.PageSize <= 0 {
		pagination.PageSize = d.config.DefaultPageSize
	}
	if pagination.PageSize > d.config.MaxPageSize {
		pagination.PageSize = d.config.MaxPageSize
	}

	entries, totalCount, err := d.storage.List(ctx, filter, pagination)
	if err != nil {
		d.logger.Error("Failed to list DLQ entries", "error", err)
		return nil, fmt.Errorf("failed to list DLQ entries: %w", err)
	}

	patterns := make([]ErrorPattern, 0)
	if filter.IncludePatterns {
		patterns, err = d.analyzer.AnalyzeEntries(ctx, entries)
		if err != nil {
			d.logger.Warn("Failed to analyze patterns", "error", err)
		}
	}

	totalPages := (totalCount + pagination.PageSize - 1) / pagination.PageSize
	hasNext := pagination.Page < totalPages
	hasPrev := pagination.Page > 1

	response := &DLQListResponse{
		Entries:     entries,
		TotalCount:  totalCount,
		Page:        pagination.Page,
		PageSize:    pagination.PageSize,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrev,
		Patterns:    patterns,
		Filter:      filter,
	}

	d.logger.Debug("DLQ entries listed successfully",
		"count", len(entries),
		"total", totalCount,
		"page", pagination.Page)

	return response, nil
}

func (d *DLQManagerImpl) PeekEntry(ctx context.Context, id string) (*DLQEntry, error) {
	d.logger.Debug("Peeking DLQ entry", "id", id)

	entry, err := d.storage.Get(ctx, id)
	if err != nil {
		d.logger.Error("Failed to peek DLQ entry", "id", id, "error", err)
		return nil, fmt.Errorf("failed to peek DLQ entry %s: %w", id, err)
	}

	d.logger.Debug("DLQ entry peeked successfully", "id", id)
	return entry, nil
}

func (d *DLQManagerImpl) RequeueEntries(ctx context.Context, ids []string) (*BulkOperationResult, error) {
	d.logger.Info("Requeuing DLQ entries", "count", len(ids))

	if len(ids) > d.config.BulkOperationLimit {
		return nil, fmt.Errorf("bulk operation limit exceeded: %d > %d", len(ids), d.config.BulkOperationLimit)
	}

	result := &BulkOperationResult{
		TotalRequested: len(ids),
		Successful:     make([]string, 0),
		Failed:         make([]OperationError, 0),
		StartedAt:      time.Now(),
	}

	for _, id := range ids {
		err := d.engine.Requeue(ctx, id)
		if err != nil {
			result.Failed = append(result.Failed, OperationError{
				ID:    id,
				Error: err.Error(),
			})
			d.logger.Warn("Failed to requeue entry", "id", id, "error", err)
		} else {
			result.Successful = append(result.Successful, id)
			d.logger.Debug("Entry requeued successfully", "id", id)
		}
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	d.logger.Info("Requeue operation completed",
		"successful", len(result.Successful),
		"failed", len(result.Failed),
		"duration", result.Duration)

	return result, nil
}

func (d *DLQManagerImpl) PurgeEntries(ctx context.Context, ids []string) (*BulkOperationResult, error) {
	d.logger.Info("Purging DLQ entries", "count", len(ids))

	if len(ids) > d.config.BulkOperationLimit {
		return nil, fmt.Errorf("bulk operation limit exceeded: %d > %d", len(ids), d.config.BulkOperationLimit)
	}

	result := &BulkOperationResult{
		TotalRequested: len(ids),
		Successful:     make([]string, 0),
		Failed:         make([]OperationError, 0),
		StartedAt:      time.Now(),
	}

	for _, id := range ids {
		err := d.engine.Purge(ctx, id)
		if err != nil {
			result.Failed = append(result.Failed, OperationError{
				ID:    id,
				Error: err.Error(),
			})
			d.logger.Warn("Failed to purge entry", "id", id, "error", err)
		} else {
			result.Successful = append(result.Successful, id)
			d.logger.Debug("Entry purged successfully", "id", id)
		}
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	d.logger.Info("Purge operation completed",
		"successful", len(result.Successful),
		"failed", len(result.Failed),
		"duration", result.Duration)

	return result, nil
}

func (d *DLQManagerImpl) PurgeAll(ctx context.Context, filter DLQFilter) (*BulkOperationResult, error) {
	d.logger.Warn("Purging all DLQ entries with filter", "filter", filter)

	allIDs, err := d.storage.ListIDs(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list entry IDs for purge all: %w", err)
	}

	if len(allIDs) == 0 {
		return &BulkOperationResult{
			TotalRequested: 0,
			Successful:     make([]string, 0),
			Failed:         make([]OperationError, 0),
			StartedAt:      time.Now(),
			CompletedAt:    time.Now(),
		}, nil
	}

	return d.PurgeEntries(ctx, allIDs)
}

func (d *DLQManagerImpl) GetStats(ctx context.Context) (*DLQStats, error) {
	d.logger.Debug("Getting DLQ statistics")

	stats, err := d.storage.GetStats(ctx)
	if err != nil {
		d.logger.Error("Failed to get DLQ stats", "error", err)
		return nil, fmt.Errorf("failed to get DLQ stats: %w", err)
	}

	d.logger.Debug("DLQ statistics retrieved", "total_entries", stats.TotalEntries)
	return stats, nil
}

type RedisStorage struct {
	redis  *redis.Client
	logger *slog.Logger
}

func NewRedisStorage(redisClient *redis.Client, logger *slog.Logger) *RedisStorage {
	return &RedisStorage{
		redis:  redisClient,
		logger: logger,
	}
}

func (r *RedisStorage) List(ctx context.Context, filter DLQFilter, pagination PaginationRequest) ([]DLQEntry, int, error) {
	dlqKey := "dlq:entries"

	allEntries, err := r.redis.HGetAll(ctx, dlqKey).Result()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get DLQ entries: %w", err)
	}

	entries := make([]DLQEntry, 0)
	for id, data := range allEntries {
		var entry DLQEntry
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			r.logger.Warn("Failed to unmarshal DLQ entry", "id", id, "error", err)
			continue
		}
		entries = append(entries, entry)
	}

	filteredEntries := r.applyFilter(entries, filter)
	sortedEntries := r.applySorting(filteredEntries, pagination.SortBy, pagination.SortOrder)

	totalCount := len(sortedEntries)

	start := (pagination.Page - 1) * pagination.PageSize
	end := start + pagination.PageSize

	if start >= totalCount {
		return []DLQEntry{}, totalCount, nil
	}

	if end > totalCount {
		end = totalCount
	}

	pagedEntries := sortedEntries[start:end]

	return pagedEntries, totalCount, nil
}

func (r *RedisStorage) Get(ctx context.Context, id string) (*DLQEntry, error) {
	dlqKey := "dlq:entries"

	data, err := r.redis.HGet(ctx, dlqKey, id).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("DLQ entry not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get DLQ entry: %w", err)
	}

	var entry DLQEntry
	if err := json.Unmarshal([]byte(data), &entry); err != nil {
		return nil, fmt.Errorf("failed to unmarshal DLQ entry: %w", err)
	}

	return &entry, nil
}

func (r *RedisStorage) ListIDs(ctx context.Context, filter DLQFilter) ([]string, error) {
	entries, _, err := r.List(ctx, filter, PaginationRequest{
		Page:     1,
		PageSize: 10000, // Large page to get all IDs
	})
	if err != nil {
		return nil, err
	}

	ids := make([]string, len(entries))
	for i, entry := range entries {
		ids[i] = entry.ID
	}

	return ids, nil
}

func (r *RedisStorage) GetStats(ctx context.Context) (*DLQStats, error) {
	dlqKey := "dlq:entries"

	totalEntries, err := r.redis.HLen(ctx, dlqKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count DLQ entries: %w", err)
	}

	allEntries, err := r.redis.HGetAll(ctx, dlqKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get DLQ entries for stats: %w", err)
	}

	queueCounts := make(map[string]int)
	typeCounts := make(map[string]int)

	for _, data := range allEntries {
		var entry DLQEntry
		if err := json.Unmarshal([]byte(data), &entry); err != nil {
			continue
		}
		queueCounts[entry.Queue]++
		typeCounts[entry.Type]++
	}

	return &DLQStats{
		TotalEntries: int(totalEntries),
		ByQueue:      queueCounts,
		ByType:       typeCounts,
		UpdatedAt:    time.Now(),
	}, nil
}

func (r *RedisStorage) applyFilter(entries []DLQEntry, filter DLQFilter) []DLQEntry {
	if filter.IsEmpty() {
		return entries
	}

	filtered := make([]DLQEntry, 0)
	for _, entry := range entries {
		if r.matchesFilter(entry, filter) {
			filtered = append(filtered, entry)
		}
	}
	return filtered
}

func (r *RedisStorage) matchesFilter(entry DLQEntry, filter DLQFilter) bool {
	if filter.Queue != "" && entry.Queue != filter.Queue {
		return false
	}

	if filter.Type != "" && entry.Type != filter.Type {
		return false
	}

	if filter.ErrorPattern != "" && !strings.Contains(entry.Error.Message, filter.ErrorPattern) {
		return false
	}

	if !filter.StartTime.IsZero() && entry.FailedAt.Before(filter.StartTime) {
		return false
	}

	if !filter.EndTime.IsZero() && entry.FailedAt.After(filter.EndTime) {
		return false
	}

	if filter.MinAttempts > 0 && len(entry.Attempts) < filter.MinAttempts {
		return false
	}

	if filter.MaxAttempts > 0 && len(entry.Attempts) > filter.MaxAttempts {
		return false
	}

	return true
}

func (r *RedisStorage) applySorting(entries []DLQEntry, sortBy string, order SortOrder) []DLQEntry {
	if sortBy == "" {
		sortBy = "failed_at"
	}

	sort.Slice(entries, func(i, j int) bool {
		var less bool

		switch sortBy {
		case "failed_at":
			less = entries[i].FailedAt.Before(entries[j].FailedAt)
		case "created_at":
			less = entries[i].CreatedAt.Before(entries[j].CreatedAt)
		case "queue":
			less = entries[i].Queue < entries[j].Queue
		case "type":
			less = entries[i].Type < entries[j].Type
		case "attempts":
			less = len(entries[i].Attempts) < len(entries[j].Attempts)
		default:
			less = entries[i].FailedAt.Before(entries[j].FailedAt)
		}

		if order == SortOrderDesc {
			return !less
		}
		return less
	})

	return entries
}