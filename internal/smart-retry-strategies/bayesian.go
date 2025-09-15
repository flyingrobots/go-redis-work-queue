// Copyright 2025 James Ross
package smartretry

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"

	"go.uber.org/zap"
)

// getBayesianModel retrieves or creates a Bayesian model for job type and error class
func (m *manager) getBayesianModel(jobType, errorClass string) (*BayesianModel, error) {
	m.bayesianMu.RLock()
	defer m.bayesianMu.RUnlock()

	cacheKey := fmt.Sprintf("bayesian:%s:%s", jobType, errorClass)

	// Check cache first
	if cached, ok := m.cache.get(cacheKey); ok {
		return cached.(*BayesianModel), nil
	}

	ctx := context.Background()
	modelKey := fmt.Sprintf("retry:bayesian:%s:%s", jobType, errorClass)

	// Try to load from Redis
	modelData, err := m.redis.Get(ctx, modelKey).Result()
	if err == nil {
		var model BayesianModel
		if err := json.Unmarshal([]byte(modelData), &model); err == nil {
			m.cache.set(cacheKey, &model, m.cache.ttl)
			return &model, nil
		}
	}

	// Model doesn't exist, return nil (will be created when UpdateBayesianModel is called)
	return nil, fmt.Errorf("no Bayesian model found for %s:%s", jobType, errorClass)
}

// UpdateBayesianModel updates the Bayesian model using recent attempt data
func (m *manager) UpdateBayesianModel(jobType, errorClass string) error {
	m.logger.Debug("Updating Bayesian model",
		zap.String("job_type", jobType),
		zap.String("error_class", errorClass))

	// Collect recent attempt data
	attempts, err := m.getRecentAttempts(jobType, errorClass, 30*24*time.Hour) // Last 30 days
	if err != nil {
		return fmt.Errorf("failed to get recent attempts: %w", err)
	}

	if len(attempts) < 10 {
		m.logger.Debug("Insufficient data for Bayesian model",
			zap.String("job_type", jobType),
			zap.String("error_class", errorClass),
			zap.Int("attempts", len(attempts)))
		return fmt.Errorf("insufficient data: only %d attempts", len(attempts))
	}

	// Create delay buckets
	buckets := m.createDelayBuckets(attempts)

	// Calculate Beta distribution parameters for each bucket
	for i := range buckets {
		bucket := &buckets[i]

		// Beta distribution: alpha = successes + 1, beta = failures + 1
		alpha := float64(bucket.Successes + 1)
		beta := float64(bucket.Failures + 1)

		// Mean of Beta distribution
		bucket.Probability = alpha / (alpha + beta)

		// Calculate confidence intervals (95%)
		bucket.LowerBound, bucket.UpperBound = m.betaConfidenceInterval(alpha, beta, 0.95)
	}

	// Create the model
	model := &BayesianModel{
		JobType:     jobType,
		ErrorClass:  errorClass,
		Buckets:     buckets,
		LastUpdated: time.Now(),
		SampleCount: len(attempts),
		Confidence:  m.calculateModelConfidence(buckets),
	}

	// Store the model
	if err := m.storeBayesianModel(model); err != nil {
		return fmt.Errorf("failed to store Bayesian model: %w", err)
	}

	m.logger.Info("Updated Bayesian model",
		zap.String("job_type", jobType),
		zap.String("error_class", errorClass),
		zap.Int("samples", len(attempts)),
		zap.Int("buckets", len(buckets)),
		zap.Float64("confidence", model.Confidence))

	return nil
}

// getRecentAttempts retrieves recent attempt data for analysis
func (m *manager) getRecentAttempts(jobType, errorClass string, window time.Duration) ([]AttemptHistory, error) {
	ctx := context.Background()

	// Use a pattern to find all attempt keys for this job type and error class
	pattern := fmt.Sprintf("retry:attempt:*")

	var attempts []AttemptHistory
	cutoff := time.Now().Add(-window)

	// Scan for keys (in production, you'd want to use a more efficient approach)
	iter := m.redis.Scan(ctx, 0, pattern, 1000).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()

		// Get attempt data
		data, err := m.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var attempt AttemptHistory
		if err := json.Unmarshal([]byte(data), &attempt); err != nil {
			continue
		}

		// Filter by job type, error class, and time window
		if attempt.JobType == jobType &&
		   attempt.ErrorClass == errorClass &&
		   attempt.Timestamp.After(cutoff) {
			attempts = append(attempts, attempt)
		}
	}

	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("Redis scan error: %w", err)
	}

	// Sort by timestamp
	sort.Slice(attempts, func(i, j int) bool {
		return attempts[i].Timestamp.Before(attempts[j].Timestamp)
	})

	return attempts, nil
}

// createDelayBuckets creates delay buckets and populates them with attempt data
func (m *manager) createDelayBuckets(attempts []AttemptHistory) []BayesianBucket {
	// Define delay buckets (in milliseconds)
	bucketRanges := []struct {
		min, max int64
	}{
		{0, 1000},        // 0-1s
		{1001, 5000},     // 1-5s
		{5001, 15000},    // 5-15s
		{15001, 30000},   // 15-30s
		{30001, 60000},   // 30-60s
		{60001, 300000},  // 1-5m
		{300001, 900000}, // 5-15m
		{900001, math.MaxInt64}, // 15m+
	}

	buckets := make([]BayesianBucket, len(bucketRanges))

	// Initialize buckets
	for i, r := range bucketRanges {
		buckets[i] = BayesianBucket{
			DelayMinMs: r.min,
			DelayMaxMs: r.max,
			Successes:  0,
			Failures:   0,
		}
	}

	// Populate buckets with attempt data
	for _, attempt := range attempts {
		bucketIndex := m.findBucketIndex(attempt.DelayMs, bucketRanges)
		if bucketIndex >= 0 && bucketIndex < len(buckets) {
			if attempt.Success {
				buckets[bucketIndex].Successes++
			} else {
				buckets[bucketIndex].Failures++
			}
		}
	}

	// Filter out empty buckets
	var nonEmptyBuckets []BayesianBucket
	for _, bucket := range buckets {
		if bucket.Successes > 0 || bucket.Failures > 0 {
			nonEmptyBuckets = append(nonEmptyBuckets, bucket)
		}
	}

	return nonEmptyBuckets
}

// findBucketIndex finds the appropriate bucket index for a delay value
func (m *manager) findBucketIndex(delayMs int64, ranges []struct{ min, max int64 }) int {
	for i, r := range ranges {
		if delayMs >= r.min && delayMs <= r.max {
			return i
		}
	}
	return -1
}

// betaConfidenceInterval calculates confidence interval for Beta distribution
func (m *manager) betaConfidenceInterval(alpha, beta float64, confidence float64) (float64, float64) {
	// Simplified confidence interval calculation
	// In practice, you'd use proper statistical functions

	mean := alpha / (alpha + beta)
	variance := (alpha * beta) / ((alpha + beta) * (alpha + beta) * (alpha + beta + 1))
	stddev := math.Sqrt(variance)

	// Approximate using normal distribution for large samples
	z := 1.96 // 95% confidence
	if confidence == 0.90 {
		z = 1.645
	} else if confidence == 0.99 {
		z = 2.576
	}

	margin := z * stddev
	lower := math.Max(0, mean-margin)
	upper := math.Min(1, mean+margin)

	return lower, upper
}

// calculateModelConfidence calculates overall confidence in the model
func (m *manager) calculateModelConfidence(buckets []BayesianBucket) float64 {
	if len(buckets) == 0 {
		return 0.0
	}

	totalSamples := 0
	weightedConfidence := 0.0

	for _, bucket := range buckets {
		samples := bucket.Successes + bucket.Failures
		totalSamples += samples

		// Confidence based on sample size and interval width
		intervalWidth := bucket.UpperBound - bucket.LowerBound
		sampleConfidence := math.Min(1.0, float64(samples)/20.0) * (1.0 - intervalWidth)

		weightedConfidence += sampleConfidence * float64(samples)
	}

	if totalSamples == 0 {
		return 0.0
	}

	confidence := weightedConfidence / float64(totalSamples)

	// Apply penalty for insufficient data
	if totalSamples < 50 {
		confidence *= float64(totalSamples) / 50.0
	}

	return confidence
}

// storeBayesianModel stores the Bayesian model in Redis and cache
func (m *manager) storeBayesianModel(model *BayesianModel) error {
	ctx := context.Background()
	modelKey := fmt.Sprintf("retry:bayesian:%s:%s", model.JobType, model.ErrorClass)

	modelData, err := json.Marshal(model)
	if err != nil {
		return fmt.Errorf("failed to marshal model: %w", err)
	}

	// Store in Redis with expiration
	err = m.redis.Set(ctx, modelKey, modelData, 7*24*time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to store model in Redis: %w", err)
	}

	// Update cache
	cacheKey := fmt.Sprintf("bayesian:%s:%s", model.JobType, model.ErrorClass)
	m.cache.set(cacheKey, model, m.cache.ttl)

	// Emit event
	m.emitEvent(RetryEvent{
		ID:        fmt.Sprintf("bayesian_%d", time.Now().UnixNano()),
		Type:      EventTypeBayesianUpdated,
		Message:   fmt.Sprintf("Bayesian model updated for %s:%s", model.JobType, model.ErrorClass),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"job_type":     model.JobType,
			"error_class":  model.ErrorClass,
			"sample_count": model.SampleCount,
			"confidence":   model.Confidence,
			"buckets":      len(model.Buckets),
		},
	})

	return nil
}

// emitEvent emits a retry event (placeholder - would integrate with event system)
func (m *manager) emitEvent(event RetryEvent) {
	m.logger.Debug("Retry event emitted",
		zap.String("type", string(event.Type)),
		zap.String("message", event.Message))

	// In a real implementation, this would publish to an event bus or notification system
}