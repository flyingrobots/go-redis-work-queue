package workerfleetcontrols

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisAuditLogger struct {
	redis  *redis.Client
	config Config
	logger *slog.Logger
}

func NewRedisAuditLogger(redisClient *redis.Client, config Config, logger *slog.Logger) *RedisAuditLogger {
	return &RedisAuditLogger{
		redis:  redisClient,
		config: config,
		logger: logger,
	}
}

func (a *RedisAuditLogger) LogAction(log AuditLog) error {
	ctx := context.Background()

	data, err := json.Marshal(log)
	if err != nil {
		return fmt.Errorf("failed to marshal audit log: %w", err)
	}

	key := "worker:audit_logs"
	err = a.redis.LPush(ctx, key, data).Err()
	if err != nil {
		return fmt.Errorf("failed to store audit log: %w", err)
	}

	err = a.redis.LTrim(ctx, key, 0, 10000).Err()
	if err != nil {
		a.logger.Warn("Failed to trim audit logs", "error", err)
	}

	a.logger.Info("Audit log recorded",
		"id", log.ID,
		"action", log.Action,
		"worker_count", len(log.WorkerIDs),
		"success", log.Success,
		"duration", log.Duration)

	return nil
}

func (a *RedisAuditLogger) GetAuditLogs(filter AuditLogFilter) ([]AuditLog, error) {
	ctx := context.Background()

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}

	key := "worker:audit_logs"
	start := int64(filter.Offset)
	end := start + int64(limit) - 1

	results, err := a.redis.LRange(ctx, key, start, end).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	logs := make([]AuditLog, 0)
	for _, result := range results {
		var log AuditLog
		err := json.Unmarshal([]byte(result), &log)
		if err != nil {
			a.logger.Warn("Failed to unmarshal audit log", "error", err)
			continue
		}

		if a.matchesFilter(log, filter) {
			logs = append(logs, log)
		}
	}

	return logs, nil
}

func (a *RedisAuditLogger) GetAuditLogsByWorker(workerID string, limit int) ([]AuditLog, error) {
	filter := AuditLogFilter{
		WorkerIDs: []string{workerID},
		Limit:     limit,
	}

	return a.GetAuditLogs(filter)
}

func (a *RedisAuditLogger) GetAuditLogsByUser(userID string, limit int) ([]AuditLog, error) {
	filter := AuditLogFilter{
		UserIDs: []string{userID},
		Limit:   limit,
	}

	return a.GetAuditLogs(filter)
}

func (a *RedisAuditLogger) matchesFilter(log AuditLog, filter AuditLogFilter) bool {
	if filter.StartTime != nil && log.Timestamp.Before(*filter.StartTime) {
		return false
	}

	if filter.EndTime != nil && log.Timestamp.After(*filter.EndTime) {
		return false
	}

	if len(filter.Actions) > 0 {
		found := false
		for _, action := range filter.Actions {
			if log.Action == action {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(filter.WorkerIDs) > 0 {
		found := false
		for _, filterWorkerID := range filter.WorkerIDs {
			for _, logWorkerID := range log.WorkerIDs {
				if filterWorkerID == logWorkerID {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			return false
		}
	}

	if len(filter.UserIDs) > 0 {
		found := false
		for _, userID := range filter.UserIDs {
			if log.UserID == userID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	if filter.Success != nil && log.Success != *filter.Success {
		return false
	}

	return true
}
