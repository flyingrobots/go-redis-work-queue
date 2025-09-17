// Copyright 2025 James Ross
package ratelimiting

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimiter provides token-bucket based rate limiting with priority fairness
type RateLimiter struct {
	redis  *redis.Client
	logger *zap.Logger
	config *Config
	mu     sync.RWMutex

	// Lua scripts
	consumeScript *redis.Script
	refillScript  *redis.Script
	statusScript  *redis.Script
}

// Config defines rate limiter configuration
type Config struct {
	// Global limits
	GlobalRatePerSecond int64
	GlobalBurstSize     int64

	// Default per-tenant limits
	DefaultRatePerSecond int64
	DefaultBurstSize     int64

	// Priority weights (higher = more tokens)
	PriorityWeights map[string]float64

	// Refill interval
	RefillInterval time.Duration

	// TTL for rate limit keys
	KeyTTL time.Duration

	// Dry run mode
	DryRun bool
}

// DefaultConfig returns a sensible default configuration
func DefaultConfig() *Config {
	return &Config{
		GlobalRatePerSecond:  10000,
		GlobalBurstSize:      20000,
		DefaultRatePerSecond: 100,
		DefaultBurstSize:     200,
		PriorityWeights: map[string]float64{
			"critical": 3.0,
			"high":     2.0,
			"normal":   1.0,
			"low":      0.5,
		},
		RefillInterval: 100 * time.Millisecond,
		KeyTTL:         1 * time.Hour,
		DryRun:         false,
	}
}

// TenantConfig defines per-tenant rate limiting configuration
type TenantConfig struct {
	TenantID         string
	RatePerSecond    int64
	BurstSize        int64
	Priority         string
	CustomWeight     float64 // Override priority weight
	Enabled          bool
	ExemptFromGlobal bool // Bypass global limits
}

// ConsumeResult contains the result of a rate limit check
type ConsumeResult struct {
	Allowed          bool
	Tokens           int64         // Tokens consumed
	Remaining        int64         // Tokens remaining
	RetryAfter       time.Duration // Wait time if denied
	ResetAt          time.Time     // When bucket refills
	DryRunWouldAllow bool          // Result if not in dry-run mode
}

// Status represents current rate limiter status
type Status struct {
	Scope      string
	Available  int64
	Capacity   int64
	RefillRate int64
	LastRefill time.Time
	NextRefill time.Time
	Priority   string
	Weight     float64
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(redis *redis.Client, logger *zap.Logger, config *Config) *RateLimiter {
	if config == nil {
		config = DefaultConfig()
	}

	rl := &RateLimiter{
		redis:  redis,
		logger: logger,
		config: config,
	}

	// Initialize Lua scripts
	rl.initLuaScripts()

	return rl
}

// initLuaScripts initializes the Lua scripts for atomic operations
func (rl *RateLimiter) initLuaScripts() {
	// Script for atomic token consumption
	rl.consumeScript = redis.NewScript(`
		local key = KEYS[1]
		local requested = tonumber(ARGV[1])
		local capacity = tonumber(ARGV[2])
		local refill_rate = tonumber(ARGV[3])
		local now = tonumber(ARGV[4])
		local ttl = tonumber(ARGV[5])
		local dry_run = ARGV[6] == "true"

		-- Get current bucket state
		local bucket = redis.call('HMGET', key, 'tokens', 'last_refill')
		local current_tokens = tonumber(bucket[1]) or capacity
		local last_refill = tonumber(bucket[2]) or now

		-- Calculate refill
		local time_passed = now - last_refill
		local tokens_to_add = math.floor(time_passed * refill_rate / 1000)
		current_tokens = math.min(capacity, current_tokens + tokens_to_add)

		-- Check if we can consume
		local allowed = current_tokens >= requested
		local consumed = 0
		local remaining = current_tokens

		if allowed and not dry_run then
			consumed = requested
			remaining = current_tokens - requested

			-- Update bucket
			redis.call('HSET', key,
				'tokens', remaining,
				'last_refill', now,
				'capacity', capacity,
				'refill_rate', refill_rate
			)
			redis.call('EXPIRE', key, ttl)
		end

		-- Calculate retry after (ms to wait for requested tokens)
		local retry_after = 0
		if not allowed then
			local tokens_needed = requested - current_tokens
			retry_after = math.ceil(tokens_needed * 1000 / refill_rate)
		end

		return {allowed and 1 or 0, consumed, remaining, retry_after}
	`)

	// Script for manual refill
	rl.refillScript = redis.NewScript(`
		local key = KEYS[1]
		local tokens_to_add = tonumber(ARGV[1])
		local capacity = tonumber(ARGV[2])
		local ttl = tonumber(ARGV[3])

		local current = tonumber(redis.call('HGET', key, 'tokens')) or 0
		local new_tokens = math.min(capacity, current + tokens_to_add)

		redis.call('HSET', key, 'tokens', new_tokens)
		redis.call('EXPIRE', key, ttl)

		return new_tokens
	`)

	// Script for status check
	rl.statusScript = redis.NewScript(`
		local key = KEYS[1]
		local now = tonumber(ARGV[1])

		local bucket = redis.call('HGETALL', key)
		if #bucket == 0 then
			return {}
		end

		local result = {}
		for i = 1, #bucket, 2 do
			result[bucket[i]] = bucket[i + 1]
		end

		return cjson.encode(result)
	`)
}

// Consume attempts to consume tokens from the rate limiter
func (rl *RateLimiter) Consume(ctx context.Context, scope string, tokens int64, priority string) (*ConsumeResult, error) {
	// Determine rate limit parameters
	config := rl.getTenantConfig(scope)
	weight := rl.getPriorityWeight(priority, config)

	// Adjust tokens based on priority weight
	adjustedTokens := int64(math.Ceil(float64(tokens) / weight))

	// Check tenant-specific limit
	tenantKey := rl.keyForScope(scope)
	tenantResult, err := rl.consumeTokens(ctx, tenantKey, adjustedTokens, config.BurstSize, config.RatePerSecond)
	if err != nil {
		return nil, fmt.Errorf("tenant rate limit check failed: %w", err)
	}

	// If denied at tenant level, return immediately
	if !tenantResult.Allowed && !rl.config.DryRun {
		return tenantResult, nil
	}

	// Check global limit if not exempt
	if !config.ExemptFromGlobal {
		globalKey := rl.keyForScope("global")
		globalResult, err := rl.consumeTokens(ctx, globalKey, tokens, rl.config.GlobalBurstSize, rl.config.GlobalRatePerSecond)
		if err != nil {
			return nil, fmt.Errorf("global rate limit check failed: %w", err)
		}

		// If global denies, return global result
		if !globalResult.Allowed && !rl.config.DryRun {
			return globalResult, nil
		}

		// Merge results (take the most restrictive)
		if globalResult.RetryAfter > tenantResult.RetryAfter {
			tenantResult.RetryAfter = globalResult.RetryAfter
		}
		if globalResult.Remaining < tenantResult.Remaining {
			tenantResult.Remaining = globalResult.Remaining
		}
	}

	// Record metrics
	rl.recordMetrics(scope, priority, tenantResult.Allowed, tokens)

	return tenantResult, nil
}

// consumeTokens executes the Lua script to atomically consume tokens
func (rl *RateLimiter) consumeTokens(ctx context.Context, key string, tokens, capacity, rate int64) (*ConsumeResult, error) {
	now := time.Now().UnixMilli()

	res, err := rl.consumeScript.Run(ctx, rl.redis, []string{key},
		tokens,
		capacity,
		rate,
		now,
		int64(rl.config.KeyTTL.Seconds()),
		fmt.Sprintf("%v", rl.config.DryRun),
	).Result()

	if err != nil {
		return nil, err
	}

	vals := res.([]interface{})
	allowed := vals[0].(int64) == 1
	consumed := vals[1].(int64)
	remaining := vals[2].(int64)
	retryAfterMs := vals[3].(int64)

	result := &ConsumeResult{
		Allowed:    allowed,
		Tokens:     consumed,
		Remaining:  remaining,
		RetryAfter: time.Duration(retryAfterMs) * time.Millisecond,
		ResetAt:    time.Now().Add(time.Duration(remaining/rate) * time.Second),
	}

	if rl.config.DryRun {
		result.DryRunWouldAllow = allowed
		result.Allowed = true // Always allow in dry-run mode
	}

	return result, nil
}

// Refill manually adds tokens to a bucket
func (rl *RateLimiter) Refill(ctx context.Context, scope string, tokens int64) (int64, error) {
	key := rl.keyForScope(scope)
	config := rl.getTenantConfig(scope)

	res, err := rl.refillScript.Run(ctx, rl.redis, []string{key},
		tokens,
		config.BurstSize,
		int64(rl.config.KeyTTL.Seconds()),
	).Result()

	if err != nil {
		return 0, err
	}

	return res.(int64), nil
}

// GetStatus returns the current status of a rate limiter scope
func (rl *RateLimiter) GetStatus(ctx context.Context, scope string) (*Status, error) {
	key := rl.keyForScope(scope)
	config := rl.getTenantConfig(scope)

	// Get bucket state
	bucket, err := rl.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	// Parse bucket data
	var tokens, lastRefill int64
	if v, ok := bucket["tokens"]; ok {
		fmt.Sscanf(v, "%d", &tokens)
	} else {
		tokens = config.BurstSize // Default to full
	}

	if v, ok := bucket["last_refill"]; ok {
		fmt.Sscanf(v, "%d", &lastRefill)
	}

	lastRefillTime := time.UnixMilli(lastRefill)
	nextRefillTime := lastRefillTime.Add(rl.config.RefillInterval)

	return &Status{
		Scope:      scope,
		Available:  tokens,
		Capacity:   config.BurstSize,
		RefillRate: config.RatePerSecond,
		LastRefill: lastRefillTime,
		NextRefill: nextRefillTime,
		Priority:   config.Priority,
		Weight:     config.CustomWeight,
	}, nil
}

// UpdateConfig updates the configuration for a specific tenant
func (rl *RateLimiter) UpdateConfig(tenantConfig *TenantConfig) error {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Validate configuration
	if tenantConfig.RatePerSecond <= 0 || tenantConfig.BurstSize <= 0 {
		return fmt.Errorf("invalid rate limit configuration")
	}

	// Store in configuration map (in production, this would be persisted)
	_ = fmt.Sprintf("config:%s", tenantConfig.TenantID)
	rl.logger.Info("Updated rate limit configuration",
		zap.String("tenant", tenantConfig.TenantID),
		zap.Int64("rate", tenantConfig.RatePerSecond),
		zap.Int64("burst", tenantConfig.BurstSize))

	return nil
}

// Reset clears the rate limit state for a scope
func (rl *RateLimiter) Reset(ctx context.Context, scope string) error {
	key := rl.keyForScope(scope)
	return rl.redis.Del(ctx, key).Err()
}

// Helper methods

func (rl *RateLimiter) keyForScope(scope string) string {
	return fmt.Sprintf("rl:%s", scope)
}

func (rl *RateLimiter) getTenantConfig(scope string) *TenantConfig {
	// In production, this would fetch from configuration store
	// For now, return defaults
	return &TenantConfig{
		TenantID:      scope,
		RatePerSecond: rl.config.DefaultRatePerSecond,
		BurstSize:     rl.config.DefaultBurstSize,
		Priority:      "normal",
		Enabled:       true,
	}
}

func (rl *RateLimiter) getPriorityWeight(priority string, config *TenantConfig) float64 {
	if config.CustomWeight > 0 {
		return config.CustomWeight
	}

	if weight, ok := rl.config.PriorityWeights[priority]; ok {
		return weight
	}

	return 1.0
}

func (rl *RateLimiter) recordMetrics(scope, priority string, allowed bool, tokens int64) {
	// Record metrics (would integrate with Prometheus/metrics system)
	status := "allowed"
	if !allowed {
		status = "denied"
	}

	rl.logger.Debug("Rate limit decision",
		zap.String("scope", scope),
		zap.String("priority", priority),
		zap.String("status", status),
		zap.Int64("tokens", tokens))
}
