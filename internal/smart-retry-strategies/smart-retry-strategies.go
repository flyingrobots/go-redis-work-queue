// Copyright 2025 James Ross
package smartretry

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"sort"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// manager implements the Manager interface
type manager struct {
	config      *Config
	logger      *zap.Logger
	redis       *redis.Client
	strategy    *RetryStrategy
	cache       *retryCache
	bayesianMu  sync.RWMutex
	policyMu    sync.RWMutex
	mlModel     *MLModel
	mlMu        sync.RWMutex
}

// retryCache provides caching for retry data
type retryCache struct {
	enabled    bool
	ttl        time.Duration
	maxEntries int
	entries    map[string]*cacheEntry
	mu         sync.RWMutex
}

type cacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

// NewManager creates a new smart retry strategies manager
func NewManager(config *Config, logger *zap.Logger) (Manager, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       config.RedisDB,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	// Initialize cache
	cache := &retryCache{
		enabled:    config.Cache.Enabled,
		ttl:        config.Cache.TTL,
		maxEntries: config.Cache.MaxEntries,
		entries:    make(map[string]*cacheEntry),
	}

	m := &manager{
		config:   config,
		logger:   logger,
		redis:    rdb,
		strategy: &config.Strategy,
		cache:    cache,
	}

	// Initialize default policies if none exist
	if len(m.strategy.Policies) == 0 {
		m.strategy.Policies = defaultPolicies()
	}

	logger.Info("Smart retry strategies manager initialized",
		zap.String("redis_addr", config.RedisAddr),
		zap.Bool("enabled", config.Enabled),
		zap.Int("policies", len(m.strategy.Policies)))

	return m, nil
}

// GetRecommendation returns a retry recommendation based on features
func (m *manager) GetRecommendation(features RetryFeatures) (*RetryRecommendation, error) {
	m.logger.Debug("Generating retry recommendation",
		zap.String("job_type", features.JobType),
		zap.String("error_class", features.ErrorClass),
		zap.Int("attempt", features.AttemptNumber))

	// Check guardrails first
	if features.AttemptNumber >= m.strategy.Guardrails.MaxAttempts {
		return &RetryRecommendation{
			ShouldRetry:      false,
			DelayMs:          0,
			MaxAttempts:      m.strategy.Guardrails.MaxAttempts,
			Confidence:       1.0,
			Rationale:        fmt.Sprintf("Max attempts (%d) reached", m.strategy.Guardrails.MaxAttempts),
			Method:           "guardrails",
			EstimatedSuccess: 0.0,
			NextEvaluation:   time.Now(),
			PolicyGuardrails: []string{"max_attempts"},
		}, nil
	}

	// Try ML model first if enabled
	if m.strategy.MLEnabled && m.mlModel != nil && m.mlModel.Enabled {
		if rec, err := m.getMLRecommendation(features); err == nil {
			return rec, nil
		}
		m.logger.Warn("ML recommendation failed, falling back to Bayesian", zap.Error(err))
	}

	// Try Bayesian model
	if rec, err := m.getBayesianRecommendation(features); err == nil && rec.Confidence >= m.strategy.BayesianThreshold {
		return rec, nil
	}

	// Fall back to rule-based policies
	return m.getRuleBasedRecommendation(features)
}

// getRuleBasedRecommendation generates recommendation using rule-based policies
func (m *manager) getRuleBasedRecommendation(features RetryFeatures) (*RetryRecommendation, error) {
	m.policyMu.RLock()
	defer m.policyMu.RUnlock()

	// Sort policies by priority (higher first)
	policies := make([]RetryPolicy, len(m.strategy.Policies))
	copy(policies, m.strategy.Policies)
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority > policies[j].Priority
	})

	for _, policy := range policies {
		if m.policyMatches(policy, features) {
			delayMs := m.calculateDelay(policy, features.AttemptNumber)

			// Apply jitter
			if policy.JitterPercent > 0 {
				jitter := float64(delayMs) * policy.JitterPercent / 100.0
				delayMs += int64(rand.Float64()*jitter - jitter/2)
			}

			// Apply guardrails
			if delayMs > m.strategy.Guardrails.MaxDelayMs {
				delayMs = m.strategy.Guardrails.MaxDelayMs
			}

			shouldRetry := features.AttemptNumber < policy.MaxAttempts
			if policy.StopOnValidation && m.isValidationError(features.ErrorClass) {
				shouldRetry = false
			}

			return &RetryRecommendation{
				ShouldRetry:      shouldRetry,
				DelayMs:          delayMs,
				MaxAttempts:      policy.MaxAttempts,
				Confidence:       0.8, // Rule-based confidence
				Rationale:        fmt.Sprintf("Policy '%s' matched", policy.Name),
				Method:           "rules",
				EstimatedSuccess: m.estimateSuccessFromHistory(features),
				NextEvaluation:   time.Now().Add(time.Duration(delayMs) * time.Millisecond),
			}, nil
		}
	}

	// Default fallback policy
	return &RetryRecommendation{
		ShouldRetry:      features.AttemptNumber < 3,
		DelayMs:          int64(math.Min(float64(1000*math.Pow(2, float64(features.AttemptNumber))), 30000)),
		MaxAttempts:      3,
		Confidence:       0.5,
		Rationale:        "Default exponential backoff",
		Method:           "default",
		EstimatedSuccess: 0.5,
		NextEvaluation:   time.Now().Add(time.Duration(1000*math.Pow(2, float64(features.AttemptNumber))) * time.Millisecond),
	}, nil
}

// getBayesianRecommendation generates recommendation using Bayesian model
func (m *manager) getBayesianRecommendation(features RetryFeatures) (*RetryRecommendation, error) {
	model, err := m.getBayesianModel(features.JobType, features.ErrorClass)
	if err != nil {
		return nil, fmt.Errorf("failed to get Bayesian model: %w", err)
	}

	if model == nil || len(model.Buckets) == 0 {
		return nil, fmt.Errorf("no Bayesian model available")
	}

	// Find the bucket with highest success probability above threshold
	var bestBucket *BayesianBucket
	var bestDelay int64

	for _, bucket := range model.Buckets {
		if bucket.Probability >= m.strategy.BayesianThreshold &&
		   bucket.LowerBound >= m.strategy.BayesianThreshold * 0.8 { // Confidence check
			if bestBucket == nil || bucket.Probability > bestBucket.Probability {
				bestBucket = &bucket
				// Use middle of bucket range
				bestDelay = (bucket.DelayMinMs + bucket.DelayMaxMs) / 2
			}
		}
	}

	if bestBucket == nil {
		return nil, fmt.Errorf("no suitable Bayesian recommendation found")
	}

	// Apply guardrails
	if bestDelay > m.strategy.Guardrails.MaxDelayMs {
		bestDelay = m.strategy.Guardrails.MaxDelayMs
	}

	confidence := (bestBucket.Probability + bestBucket.LowerBound) / 2

	return &RetryRecommendation{
		ShouldRetry:      true,
		DelayMs:          bestDelay,
		MaxAttempts:      m.strategy.Guardrails.MaxAttempts,
		Confidence:       confidence,
		Rationale:        fmt.Sprintf("Bayesian model predicts %.1f%% success", bestBucket.Probability*100),
		Method:           "bayesian",
		EstimatedSuccess: bestBucket.Probability,
		NextEvaluation:   time.Now().Add(time.Duration(bestDelay) * time.Millisecond),
	}, nil
}

// getMLRecommendation generates recommendation using ML model
func (m *manager) getMLRecommendation(features RetryFeatures) (*RetryRecommendation, error) {
	m.mlMu.RLock()
	model := m.mlModel
	m.mlMu.RUnlock()

	if model == nil || !model.Enabled {
		return nil, fmt.Errorf("ML model not available")
	}

	// Extract features for ML model
	featureVector, err := m.extractMLFeatures(features, model.Features)
	if err != nil {
		return nil, fmt.Errorf("failed to extract ML features: %w", err)
	}

	// Run inference (simplified - would use actual ML library)
	prediction, confidence := m.runMLInference(model, featureVector)

	// Convert prediction to delay recommendation
	delayMs := m.predictionToDelay(prediction, features.AttemptNumber)

	// Apply guardrails
	if delayMs > m.strategy.Guardrails.MaxDelayMs {
		delayMs = m.strategy.Guardrails.MaxDelayMs
	}

	return &RetryRecommendation{
		ShouldRetry:      prediction > 0.3, // Threshold for retry
		DelayMs:          delayMs,
		MaxAttempts:      m.strategy.Guardrails.MaxAttempts,
		Confidence:       confidence,
		Rationale:        fmt.Sprintf("ML model (%s) prediction: %.3f", model.ModelType, prediction),
		Method:           "ml",
		EstimatedSuccess: prediction,
		NextEvaluation:   time.Now().Add(time.Duration(delayMs) * time.Millisecond),
	}, nil
}

// PreviewRetrySchedule generates a preview of the retry schedule
func (m *manager) PreviewRetrySchedule(features RetryFeatures, maxAttempts int) (*RetryPreview, error) {
	preview := &RetryPreview{
		JobID:          features.JobType + "_preview",
		CurrentAttempt: features.AttemptNumber,
		Features:       features,
		Timeline:       make([]RetryTimelineEntry, 0, maxAttempts),
		GeneratedAt:    time.Now(),
	}

	currentTime := time.Now()
	currentFeatures := features

	for attempt := features.AttemptNumber; attempt < maxAttempts; attempt++ {
		currentFeatures.AttemptNumber = attempt

		rec, err := m.GetRecommendation(currentFeatures)
		if err != nil {
			m.logger.Warn("Failed to get recommendation for preview",
				zap.Int("attempt", attempt), zap.Error(err))
			continue
		}

		preview.Recommendations = append(preview.Recommendations, *rec)

		if !rec.ShouldRetry {
			break
		}

		currentTime = currentTime.Add(time.Duration(rec.DelayMs) * time.Millisecond)

		entry := RetryTimelineEntry{
			AttemptNumber:    attempt + 1,
			ScheduledTime:    currentTime,
			EstimatedSuccess: rec.EstimatedSuccess,
			DelayMs:          rec.DelayMs,
			Method:           rec.Method,
			Rationale:        rec.Rationale,
		}

		preview.Timeline = append(preview.Timeline, entry)
	}

	return preview, nil
}

// RecordAttempt records a retry attempt for learning
func (m *manager) RecordAttempt(attempt AttemptHistory) error {
	if !m.config.DataCollection.Enabled {
		return nil
	}

	// Sample rate check
	if rand.Float64() > m.config.DataCollection.SampleRate {
		return nil
	}

	ctx := context.Background()

	// Store attempt history
	attemptKey := fmt.Sprintf("retry:attempt:%s:%d", attempt.JobID, attempt.AttemptNumber)
	attemptData, err := json.Marshal(attempt)
	if err != nil {
		return fmt.Errorf("failed to marshal attempt: %w", err)
	}

	pipe := m.redis.Pipeline()
	pipe.Set(ctx, attemptKey, attemptData, time.Duration(m.config.DataCollection.RetentionDays)*24*time.Hour)

	// Update aggregated stats
	statsKey := fmt.Sprintf("retry:stats:%s:%s", attempt.JobType, attempt.ErrorClass)
	pipe.HIncrBy(ctx, statsKey, "total_attempts", 1)

	if attempt.Success {
		pipe.HIncrBy(ctx, statsKey, "successful_retries", 1)
	} else {
		pipe.HIncrBy(ctx, statsKey, "failed_retries", 1)
	}

	pipe.HSet(ctx, statsKey, "last_updated", time.Now().Unix())
	pipe.Expire(ctx, statsKey, time.Duration(m.config.DataCollection.RetentionDays)*24*time.Hour)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to record attempt: %w", err)
	}

	// Update Bayesian model if enough data
	go func() {
		if err := m.UpdateBayesianModel(attempt.JobType, attempt.ErrorClass); err != nil {
			m.logger.Warn("Failed to update Bayesian model",
				zap.String("job_type", attempt.JobType),
				zap.String("error_class", attempt.ErrorClass),
				zap.Error(err))
		}
	}()

	m.logger.Debug("Recorded retry attempt",
		zap.String("job_id", attempt.JobID),
		zap.Int("attempt", attempt.AttemptNumber),
		zap.Bool("success", attempt.Success))

	return nil
}

// GetStats returns retry statistics for a job type and error class
func (m *manager) GetStats(jobType, errorClass string, window time.Duration) (*RetryStats, error) {
	cacheKey := fmt.Sprintf("stats:%s:%s:%s", jobType, errorClass, window.String())

	// Check cache first
	if cached, ok := m.cache.get(cacheKey); ok {
		return cached.(*RetryStats), nil
	}

	ctx := context.Background()
	statsKey := fmt.Sprintf("retry:stats:%s:%s", jobType, errorClass)

	results, err := m.redis.HMGet(ctx, statsKey,
		"total_attempts", "successful_retries", "failed_retries", "last_updated").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}

	stats := &RetryStats{
		JobType:    jobType,
		ErrorClass: errorClass,
		WindowEnd:  time.Now(),
		WindowStart: time.Now().Add(-window),
	}

	if results[0] != nil {
		if total, ok := results[0].(string); ok {
			if val, err := parseInt64(total); err == nil {
				stats.TotalAttempts = val
			}
		}
	}

	if results[1] != nil {
		if success, ok := results[1].(string); ok {
			if val, err := parseInt64(success); err == nil {
				stats.SuccessfulRetries = val
			}
		}
	}

	if results[2] != nil {
		if failed, ok := results[2].(string); ok {
			if val, err := parseInt64(failed); err == nil {
				stats.FailedRetries = val
			}
		}
	}

	if stats.TotalAttempts > 0 {
		stats.SuccessRate = float64(stats.SuccessfulRetries) / float64(stats.TotalAttempts)
	}

	// Cache the result
	m.cache.set(cacheKey, stats, m.cache.ttl)

	return stats, nil
}

// Helper functions

func (m *manager) policyMatches(policy RetryPolicy, features RetryFeatures) bool {
	// Check error patterns
	for _, pattern := range policy.ErrorPatterns {
		if matched, _ := regexp.MatchString(pattern, features.ErrorClass); matched {
			return true
		}
		if matched, _ := regexp.MatchString(pattern, features.ErrorCode); matched {
			return true
		}
	}

	// Check job type patterns
	if len(policy.JobTypePatterns) > 0 {
		for _, pattern := range policy.JobTypePatterns {
			if matched, _ := regexp.MatchString(pattern, features.JobType); matched {
				return true
			}
		}
		return false
	}

	return len(policy.ErrorPatterns) == 0 // Match all if no patterns
}

func (m *manager) calculateDelay(policy RetryPolicy, attempt int) int64 {
	delay := float64(policy.BaseDelayMs)
	for i := 1; i < attempt; i++ {
		delay *= policy.BackoffMultiplier
	}

	if int64(delay) > policy.MaxDelayMs {
		return policy.MaxDelayMs
	}

	return int64(delay)
}

func (m *manager) isValidationError(errorClass string) bool {
	validationPatterns := []string{
		"validation",
		"invalid_input",
		"malformed",
		"schema_error",
	}

	for _, pattern := range validationPatterns {
		if matched, _ := regexp.MatchString(pattern, errorClass); matched {
			return true
		}
	}
	return false
}

func (m *manager) estimateSuccessFromHistory(features RetryFeatures) float64 {
	// Simplified estimation based on attempt number
	switch {
	case features.AttemptNumber <= 1:
		return 0.8
	case features.AttemptNumber <= 3:
		return 0.6
	case features.AttemptNumber <= 5:
		return 0.4
	default:
		return 0.2
	}
}

func defaultPolicies() []RetryPolicy {
	return []RetryPolicy{
		{
			Name:              "rate_limit",
			ErrorPatterns:     []string{"429", "rate_limit", "too_many_requests"},
			MaxAttempts:       5,
			BaseDelayMs:       5000,
			MaxDelayMs:        60000,
			BackoffMultiplier: 2.0,
			JitterPercent:     25.0,
			Priority:          100,
		},
		{
			Name:              "service_unavailable",
			ErrorPatterns:     []string{"503", "service_unavailable", "timeout"},
			MaxAttempts:       3,
			BaseDelayMs:       2000,
			MaxDelayMs:        30000,
			BackoffMultiplier: 2.0,
			JitterPercent:     20.0,
			Priority:          90,
		},
		{
			Name:              "validation_errors",
			ErrorPatterns:     []string{"400", "validation", "invalid_input"},
			MaxAttempts:       1,
			BaseDelayMs:       0,
			MaxDelayMs:        0,
			BackoffMultiplier: 1.0,
			StopOnValidation:  true,
			Priority:          80,
		},
		{
			Name:              "default",
			ErrorPatterns:     []string{},
			MaxAttempts:       3,
			BaseDelayMs:       1000,
			MaxDelayMs:        30000,
			BackoffMultiplier: 2.0,
			JitterPercent:     15.0,
			Priority:          1,
		},
	}
}

func parseInt64(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

// Cache methods
func (c *retryCache) get(key string) (interface{}, bool) {
	if !c.enabled {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists || time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.value, true
}

func (c *retryCache) set(key string, value interface{}, ttl time.Duration) {
	if !c.enabled {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Clean up expired entries if at capacity
	if len(c.entries) >= c.maxEntries {
		c.cleanup()
	}

	c.entries[key] = &cacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
}

func (c *retryCache) cleanup() {
	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expiresAt) {
			delete(c.entries, key)
		}
	}
}