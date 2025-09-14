// Copyright 2025 James Ross
package exactly_once

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// IdempotencyManager provides exactly-once semantics
type IdempotencyManager interface {
	// CheckAndReserve atomically checks for duplicates and reserves the key
	CheckAndReserve(ctx context.Context, key string, ttl time.Duration) (bool, error)

	// Release removes the reservation (for failed jobs)
	Release(ctx context.Context, key string) error

	// Confirm marks the key as successfully processed
	Confirm(ctx context.Context, key string) error

	// Stats returns deduplication statistics
	Stats(ctx context.Context) (*DedupStats, error)
}

// KeyGenerator creates idempotency keys from jobs
type KeyGenerator interface {
	Generate(payload interface{}) string
	Validate(key string) error
}

// DedupStats holds deduplication statistics
type DedupStats struct {
	Processed    int64   `json:"processed"`
	Duplicates   int64   `json:"duplicates"`
	HitRate      float64 `json:"hit_rate"`
	StorageSize  int64   `json:"storage_size"`
	ActiveKeys   int64   `json:"active_keys"`
}

// RedisIdempotencyManager implements IdempotencyManager using Redis
type RedisIdempotencyManager struct {
	client     *redis.Client
	namespace  string
	defaultTTL time.Duration
}

// NewRedisIdempotencyManager creates a new Redis-backed idempotency manager
func NewRedisIdempotencyManager(client *redis.Client, namespace string, defaultTTL time.Duration) *RedisIdempotencyManager {
	if namespace == "" {
		namespace = "idempotency"
	}
	if defaultTTL == 0 {
		defaultTTL = 24 * time.Hour
	}
	return &RedisIdempotencyManager{
		client:     client,
		namespace:  namespace,
		defaultTTL: defaultTTL,
	}
}

func (r *RedisIdempotencyManager) keyName(key string) string {
	return fmt.Sprintf("%s:key:%s", r.namespace, key)
}

func (r *RedisIdempotencyManager) statsKey() string {
	return fmt.Sprintf("%s:stats", r.namespace)
}

// CheckAndReserve atomically checks for duplicates and reserves the key
func (r *RedisIdempotencyManager) CheckAndReserve(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	if ttl == 0 {
		ttl = r.defaultTTL
	}

	script := `
		local key = KEYS[1]
		local stats_key = KEYS[2]
		local ttl = ARGV[1]
		local timestamp = ARGV[2]

		-- Check if key already exists
		if redis.call('EXISTS', key) == 1 then
			-- Increment duplicate counter
			redis.call('HINCRBY', stats_key, 'duplicates', 1)
			return 1 -- Duplicate found
		else
			-- Reserve the key with TTL
			redis.call('SETEX', key, ttl, timestamp)
			-- Increment processed counter
			redis.call('HINCRBY', stats_key, 'processed', 1)
			return 0 -- Successfully reserved
		end
	`

	result, err := r.client.Eval(
		ctx,
		script,
		[]string{r.keyName(key), r.statsKey()},
		int(ttl.Seconds()),
		time.Now().Unix(),
	).Int()

	if err != nil {
		return false, fmt.Errorf("failed to check and reserve key: %w", err)
	}

	return result == 1, nil
}

// Release removes the reservation (for failed jobs)
func (r *RedisIdempotencyManager) Release(ctx context.Context, key string) error {
	return r.client.Del(ctx, r.keyName(key)).Err()
}

// Confirm marks the key as successfully processed (extends TTL)
func (r *RedisIdempotencyManager) Confirm(ctx context.Context, key string) error {
	// Extend TTL to keep record of successful processing
	return r.client.Expire(ctx, r.keyName(key), r.defaultTTL).Err()
}

// Stats returns deduplication statistics
func (r *RedisIdempotencyManager) Stats(ctx context.Context) (*DedupStats, error) {
	// Get stats from Redis
	stats, err := r.client.HGetAll(ctx, r.statsKey()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	// Parse stats
	var processed, duplicates int64
	if val, ok := stats["processed"]; ok {
		fmt.Sscanf(val, "%d", &processed)
	}
	if val, ok := stats["duplicates"]; ok {
		fmt.Sscanf(val, "%d", &duplicates)
	}

	// Calculate hit rate
	var hitRate float64
	if processed > 0 {
		hitRate = float64(duplicates) / float64(processed+duplicates) * 100
	}

	// Count active keys
	pattern := fmt.Sprintf("%s:key:*", r.namespace)
	keys, err := r.client.Keys(ctx, pattern).Result()
	activeKeys := int64(len(keys))

	// Estimate storage size (rough estimate)
	storageSize := activeKeys * 100 // ~100 bytes per key

	return &DedupStats{
		Processed:   processed,
		Duplicates:  duplicates,
		HitRate:     hitRate,
		StorageSize: storageSize,
		ActiveKeys:  activeKeys,
	}, nil
}

// UUIDKeyGenerator generates UUID-based idempotency keys
type UUIDKeyGenerator struct {
	namespace string
	prefix    string
}

// NewUUIDKeyGenerator creates a new UUID-based key generator
func NewUUIDKeyGenerator(namespace, prefix string) *UUIDKeyGenerator {
	return &UUIDKeyGenerator{
		namespace: namespace,
		prefix:    prefix,
	}
}

// Generate creates a new UUID-based key
func (g *UUIDKeyGenerator) Generate(payload interface{}) string {
	timestamp := time.Now().Format("2006-01-02")
	id := uuid.New().String()[:8]
	if g.prefix != "" {
		return fmt.Sprintf("%s_%s_%s_uuid_%s", g.namespace, g.prefix, timestamp, id)
	}
	return fmt.Sprintf("%s_%s_uuid_%s", g.namespace, timestamp, id)
}

// Validate checks if a key is valid
func (g *UUIDKeyGenerator) Validate(key string) error {
	if len(key) < 10 {
		return fmt.Errorf("key too short: %s", key)
	}
	return nil
}

// ContentHashGenerator generates content-based idempotency keys
type ContentHashGenerator struct {
	namespace string
	fields    []string // fields to include in hash
}

// NewContentHashGenerator creates a new content-based key generator
func NewContentHashGenerator(namespace string, fields []string) *ContentHashGenerator {
	return &ContentHashGenerator{
		namespace: namespace,
		fields:    fields,
	}
}

// Generate creates a content-based hash key
func (g *ContentHashGenerator) Generate(payload interface{}) string {
	hasher := sha256.New()

	// For simplicity, we'll just hash the entire payload as string
	hasher.Write([]byte(fmt.Sprintf("%v", payload)))

	hash := hex.EncodeToString(hasher.Sum(nil))[:16]
	return fmt.Sprintf("%s_hash_%s", g.namespace, hash)
}

// Validate checks if a key is valid
func (g *ContentHashGenerator) Validate(key string) error {
	if len(key) < 10 {
		return fmt.Errorf("key too short: %s", key)
	}
	return nil
}

// HybridKeyGenerator combines UUID and content hash strategies
type HybridKeyGenerator struct {
	uuidGen *UUIDKeyGenerator
	hashGen *ContentHashGenerator
}

// NewHybridKeyGenerator creates a new hybrid key generator
func NewHybridKeyGenerator(namespace string) *HybridKeyGenerator {
	return &HybridKeyGenerator{
		uuidGen: NewUUIDKeyGenerator(namespace, ""),
		hashGen: NewContentHashGenerator(namespace, nil),
	}
}

// Generate creates a hybrid key combining content hash with UUID
func (g *HybridKeyGenerator) Generate(payload interface{}) string {
	contentKey := g.hashGen.Generate(payload)
	uuid := uuid.New().String()[:8]
	return fmt.Sprintf("%s_%s", contentKey, uuid)
}

// Validate checks if a key is valid
func (g *HybridKeyGenerator) Validate(key string) error {
	if len(key) < 20 {
		return fmt.Errorf("key too short for hybrid: %s", key)
	}
	return nil
}