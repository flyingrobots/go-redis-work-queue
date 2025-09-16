// Copyright 2025 James Ross
package exactlyonce

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisIdempotencyStorage implements IdempotencyStorage using Redis
type RedisIdempotencyStorage struct {
	rdb *redis.Client
	cfg *Config
	log *zap.Logger
}

// NewRedisIdempotencyStorage creates a new Redis-based idempotency storage
func NewRedisIdempotencyStorage(rdb *redis.Client, cfg *Config, log *zap.Logger) *RedisIdempotencyStorage {
	return &RedisIdempotencyStorage{
		rdb: rdb,
		cfg: cfg,
		log: log,
	}
}

// Check verifies if an idempotency key has been processed before
func (r *RedisIdempotencyStorage) Check(ctx context.Context, key IdempotencyKey) (*IdempotencyResult, error) {
	redisKey := r.buildRedisKey(key)

	if r.cfg.Idempotency.Storage.Redis.UseHashes {
		return r.checkWithHash(ctx, key, redisKey)
	}

	return r.checkWithKey(ctx, key, redisKey)
}

// Set marks an idempotency key as processed with optional result value
func (r *RedisIdempotencyStorage) Set(ctx context.Context, key IdempotencyKey, value interface{}) error {
	redisKey := r.buildRedisKey(key)

	// Serialize the value
	valueBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if r.cfg.Idempotency.Storage.Redis.UseHashes {
		return r.setWithHash(ctx, key, redisKey, valueBytes)
	}

	return r.setWithKey(ctx, key, redisKey, valueBytes)
}

// Delete removes an idempotency key from storage
func (r *RedisIdempotencyStorage) Delete(ctx context.Context, key IdempotencyKey) error {
	redisKey := r.buildRedisKey(key)

	if r.cfg.Idempotency.Storage.Redis.UseHashes {
		hashKey := r.buildHashKey(key)
		return r.rdb.HDel(ctx, hashKey, key.ID).Err()
	}

	return r.rdb.Del(ctx, redisKey).Err()
}

// Stats returns statistics about the deduplication store
func (r *RedisIdempotencyStorage) Stats(ctx context.Context, queueName, tenantID string) (*DedupStats, error) {
	var pattern string
	if r.cfg.Idempotency.Storage.Redis.UseHashes {
		// For hash-based storage, scan hash keys
		pattern = r.buildHashKeyPattern(queueName, tenantID)
	} else {
		// For key-based storage, scan individual keys
		pattern = r.buildKeyPattern(queueName, tenantID)
	}

	// Get total keys count
	var totalKeys int64
	var cursor uint64
	for {
		keys, nextCursor, err := r.rdb.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return nil, fmt.Errorf("failed to scan keys: %w", err)
		}

		if r.cfg.Idempotency.Storage.Redis.UseHashes {
			// For hashes, count the fields in each hash
			for _, hashKey := range keys {
				count, err := r.rdb.HLen(ctx, hashKey).Result()
				if err != nil {
					r.log.Warn("Failed to get hash length", zap.String("hash", hashKey), zap.Error(err))
					continue
				}
				totalKeys += count
			}
		} else {
			totalKeys += int64(len(keys))
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	// TODO: Implement proper hit rate calculation
	// This would require tracking requests and hits over time
	hitRate := 0.0
	totalRequests := int64(0)
	duplicatesAvoided := int64(0)

	return &DedupStats{
		QueueName:         queueName,
		TenantID:          tenantID,
		TotalKeys:         totalKeys,
		HitRate:           hitRate,
		TotalRequests:     totalRequests,
		DuplicatesAvoided: duplicatesAvoided,
		LastUpdated:       time.Now().UTC(),
	}, nil
}

// checkWithKey checks idempotency using individual Redis keys
func (r *RedisIdempotencyStorage) checkWithKey(ctx context.Context, key IdempotencyKey, redisKey string) (*IdempotencyResult, error) {
	value, err := r.rdb.Get(ctx, redisKey).Result()
	if err != nil {
		if err == redis.Nil {
			// Key doesn't exist - first time processing
			return &IdempotencyResult{
				IsFirstTime: true,
				Key:         redisKey,
			}, nil
		}
		return nil, fmt.Errorf("failed to check Redis key: %w", err)
	}

	// Key exists - not first time
	var existingValue interface{}
	if err := json.Unmarshal([]byte(value), &existingValue); err != nil {
		r.log.Warn("Failed to unmarshal existing value", zap.String("key", redisKey), zap.Error(err))
	}

	return &IdempotencyResult{
		IsFirstTime:   false,
		ExistingValue: existingValue,
		Key:           redisKey,
	}, nil
}

// checkWithHash checks idempotency using Redis hashes
func (r *RedisIdempotencyStorage) checkWithHash(ctx context.Context, key IdempotencyKey, redisKey string) (*IdempotencyResult, error) {
	hashKey := r.buildHashKey(key)

	value, err := r.rdb.HGet(ctx, hashKey, key.ID).Result()
	if err != nil {
		if err == redis.Nil {
			// Field doesn't exist - first time processing
			return &IdempotencyResult{
				IsFirstTime: true,
				Key:         redisKey,
			}, nil
		}
		return nil, fmt.Errorf("failed to check Redis hash field: %w", err)
	}

	// Field exists - not first time
	var existingValue interface{}
	if err := json.Unmarshal([]byte(value), &existingValue); err != nil {
		r.log.Warn("Failed to unmarshal existing value", zap.String("key", redisKey), zap.Error(err))
	}

	return &IdempotencyResult{
		IsFirstTime:   false,
		ExistingValue: existingValue,
		Key:           redisKey,
	}, nil
}

// setWithKey stores value using individual Redis keys
func (r *RedisIdempotencyStorage) setWithKey(ctx context.Context, key IdempotencyKey, redisKey string, valueBytes []byte) error {
	return r.rdb.Set(ctx, redisKey, valueBytes, key.TTL).Err()
}

// setWithHash stores value using Redis hashes
func (r *RedisIdempotencyStorage) setWithHash(ctx context.Context, key IdempotencyKey, redisKey string, valueBytes []byte) error {
	hashKey := r.buildHashKey(key)

	pipe := r.rdb.Pipeline()
	pipe.HSet(ctx, hashKey, key.ID, valueBytes)
	pipe.Expire(ctx, hashKey, key.TTL)
	_, err := pipe.Exec(ctx)

	return err
}

// buildRedisKey constructs the Redis key for an idempotency key
func (r *RedisIdempotencyStorage) buildRedisKey(key IdempotencyKey) string {
	pattern := r.cfg.Idempotency.Storage.Redis.KeyPattern
	if pattern == "" {
		pattern = "{queue}:idempotency:{tenant}:{key}"
	}

	redisKey := strings.ReplaceAll(pattern, "{queue}", key.QueueName)
	redisKey = strings.ReplaceAll(redisKey, "{key}", key.ID)

	if key.TenantID != "" {
		redisKey = strings.ReplaceAll(redisKey, "{tenant}", key.TenantID)
	} else {
		// Remove tenant placeholder if no tenant
		redisKey = strings.ReplaceAll(redisKey, ":{tenant}", "")
		redisKey = strings.ReplaceAll(redisKey, "{tenant}:", "")
		redisKey = strings.ReplaceAll(redisKey, "{tenant}", "")
	}

	return fmt.Sprintf("%s%s", r.cfg.Idempotency.KeyPrefix, redisKey)
}

// buildHashKey constructs the Redis hash key for an idempotency key
func (r *RedisIdempotencyStorage) buildHashKey(key IdempotencyKey) string {
	pattern := r.cfg.Idempotency.Storage.Redis.HashKeyPattern
	if pattern == "" {
		pattern = "{queue}:idempotency:{tenant}"
	}

	hashKey := strings.ReplaceAll(pattern, "{queue}", key.QueueName)

	if key.TenantID != "" {
		hashKey = strings.ReplaceAll(hashKey, "{tenant}", key.TenantID)
	} else {
		// Remove tenant placeholder if no tenant
		hashKey = strings.ReplaceAll(hashKey, ":{tenant}", "")
		hashKey = strings.ReplaceAll(hashKey, "{tenant}:", "")
		hashKey = strings.ReplaceAll(hashKey, "{tenant}", "")
	}

	return fmt.Sprintf("%s%s", r.cfg.Idempotency.KeyPrefix, hashKey)
}

// buildKeyPattern constructs a pattern for scanning keys
func (r *RedisIdempotencyStorage) buildKeyPattern(queueName, tenantID string) string {
	pattern := r.cfg.Idempotency.Storage.Redis.KeyPattern
	if pattern == "" {
		pattern = "{queue}:idempotency:{tenant}:{key}"
	}

	keyPattern := strings.ReplaceAll(pattern, "{queue}", queueName)
	keyPattern = strings.ReplaceAll(keyPattern, "{key}", "*")

	if tenantID != "" {
		keyPattern = strings.ReplaceAll(keyPattern, "{tenant}", tenantID)
	} else {
		// Replace tenant placeholder with wildcard
		keyPattern = strings.ReplaceAll(keyPattern, ":{tenant}", ":*")
		keyPattern = strings.ReplaceAll(keyPattern, "{tenant}:", "*:")
		keyPattern = strings.ReplaceAll(keyPattern, "{tenant}", "*")
	}

	return fmt.Sprintf("%s%s", r.cfg.Idempotency.KeyPrefix, keyPattern)
}

// buildHashKeyPattern constructs a pattern for scanning hash keys
func (r *RedisIdempotencyStorage) buildHashKeyPattern(queueName, tenantID string) string {
	pattern := r.cfg.Idempotency.Storage.Redis.HashKeyPattern
	if pattern == "" {
		pattern = "{queue}:idempotency:{tenant}"
	}

	hashPattern := strings.ReplaceAll(pattern, "{queue}", queueName)

	if tenantID != "" {
		hashPattern = strings.ReplaceAll(hashPattern, "{tenant}", tenantID)
	} else {
		// Replace tenant placeholder with wildcard
		hashPattern = strings.ReplaceAll(hashPattern, ":{tenant}", ":*")
		hashPattern = strings.ReplaceAll(hashPattern, "{tenant}:", "*:")
		hashPattern = strings.ReplaceAll(hashPattern, "{tenant}", "*")
	}

	return fmt.Sprintf("%s%s", r.cfg.Idempotency.KeyPrefix, hashPattern)
}