// Copyright 2025 James Ross
package ratelimiting

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// PriorityFairness implements weighted fair queuing for rate limiting
type PriorityFairness struct {
	redis  *redis.Client
	logger *zap.Logger
	config *FairnessConfig

	// Cached weights for performance
	weightCache map[string]float64
	mu          sync.RWMutex
}

// FairnessConfig defines priority fairness configuration
type FairnessConfig struct {
	// Base weights for each priority level
	Weights map[string]float64

	// Starvation prevention
	MinGuaranteedShare float64 // Minimum share per priority (e.g., 0.05 = 5%)
	MaxWaitTime        time.Duration

	// Adaptive fairness
	EnableAdaptive bool
	AdaptiveWindow time.Duration

	// Burst allowance
	BurstMultiplier float64 // Allow bursts up to this multiple of fair share
}

// DefaultFairnessConfig returns default fairness configuration
func DefaultFairnessConfig() *FairnessConfig {
	return &FairnessConfig{
		Weights: map[string]float64{
			"critical": 4.0,
			"high":     2.0,
			"normal":   1.0,
			"low":      0.5,
		},
		MinGuaranteedShare: 0.05,
		MaxWaitTime:        30 * time.Second,
		EnableAdaptive:     true,
		AdaptiveWindow:     1 * time.Minute,
		BurstMultiplier:    2.0,
	}
}

// FairnessState tracks the state for fair scheduling
type FairnessState struct {
	Priority         string
	AllocatedTokens  int64
	ConsumedTokens   int64
	QueuedRequests   int64
	AverageWaitTime  time.Duration
	LastScheduled    time.Time
	StarvationRisk   bool
}

// NewPriorityFairness creates a new priority fairness scheduler
func NewPriorityFairness(redis *redis.Client, logger *zap.Logger, config *FairnessConfig) *PriorityFairness {
	if config == nil {
		config = DefaultFairnessConfig()
	}

	return &PriorityFairness{
		redis:       redis,
		logger:      logger,
		config:      config,
		weightCache: make(map[string]float64),
	}
}

// AllocateTokens distributes available tokens among priorities fairly
func (pf *PriorityFairness) AllocateTokens(ctx context.Context, availableTokens int64, demands map[string]int64) (map[string]int64, error) {
	allocations := make(map[string]int64)

	// Calculate total weighted demand
	totalWeightedDemand := 0.0
	normalizedWeights := pf.normalizeWeights(demands)

	for priority, demand := range demands {
		if demand > 0 {
			weight := normalizedWeights[priority]
			totalWeightedDemand += float64(demand) * weight
		}
	}

	if totalWeightedDemand == 0 {
		return allocations, nil
	}

	// Phase 1: Allocate minimum guaranteed share
	remainingTokens := availableTokens
	minShare := int64(float64(availableTokens) * pf.config.MinGuaranteedShare)

	for priority, demand := range demands {
		if demand > 0 {
			guaranteed := minShare
			if guaranteed > demand {
				guaranteed = demand
			}
			allocations[priority] = guaranteed
			remainingTokens -= guaranteed
		}
	}

	// Phase 2: Distribute remaining tokens by weighted fair share
	if remainingTokens > 0 {
		for priority, demand := range demands {
			if demand > allocations[priority] {
				weight := normalizedWeights[priority]
				additionalDemand := demand - allocations[priority]

				fairShare := int64(float64(remainingTokens) * weight)
				if fairShare > additionalDemand {
					fairShare = additionalDemand
				}

				allocations[priority] += fairShare
			}
		}
	}

	// Phase 3: Handle starvation prevention
	if pf.config.EnableAdaptive {
		allocations = pf.preventStarvation(ctx, allocations, demands)
	}

	// Record allocation metrics
	pf.recordAllocationMetrics(allocations, demands)

	return allocations, nil
}

// CheckFairness evaluates if current consumption is fair
func (pf *PriorityFairness) CheckFairness(ctx context.Context, priority string, requestedTokens int64) (*FairnessDecision, error) {
	// Get current state from Redis
	state, err := pf.getState(ctx, priority)
	if err != nil {
		return nil, err
	}

	// Calculate fair share based on weight
	weight := pf.getWeight(priority)
	fairShare := pf.calculateFairShare(ctx, weight)

	// Check if within fair allocation
	withinFairShare := state.ConsumedTokens + requestedTokens <= fairShare

	// Check burst allowance
	burstLimit := int64(float64(fairShare) * pf.config.BurstMultiplier)
	withinBurstLimit := state.ConsumedTokens + requestedTokens <= burstLimit

	// Check starvation risk
	timeSinceScheduled := time.Since(state.LastScheduled)
	isStarving := timeSinceScheduled > pf.config.MaxWaitTime

	decision := &FairnessDecision{
		Priority:         priority,
		Allowed:          withinBurstLimit || isStarving,
		FairShare:        fairShare,
		CurrentUsage:     state.ConsumedTokens,
		BurstLimit:       burstLimit,
		IsWithinFairShare: withinFairShare,
		IsStarving:       isStarving,
		SuggestedDelay:   pf.calculateDelay(state, fairShare, requestedTokens),
	}

	// Update state if allowed
	if decision.Allowed {
		err = pf.updateState(ctx, priority, requestedTokens)
		if err != nil {
			pf.logger.Warn("Failed to update fairness state",
				zap.String("priority", priority),
				zap.Error(err))
		}
	}

	return decision, nil
}

// FairnessDecision represents the fairness check result
type FairnessDecision struct {
	Priority          string
	Allowed           bool
	FairShare         int64
	CurrentUsage      int64
	BurstLimit        int64
	IsWithinFairShare bool
	IsStarving        bool
	SuggestedDelay    time.Duration
}

// ResetFairness resets fairness state for a time window
func (pf *PriorityFairness) ResetFairness(ctx context.Context) error {
	// Reset all priority states
	pattern := "fairness:state:*"
	keys, err := pf.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return pf.redis.Del(ctx, keys...).Err()
	}

	return nil
}

// GetFairnessStats returns current fairness statistics
func (pf *PriorityFairness) GetFairnessStats(ctx context.Context) (map[string]*FairnessState, error) {
	stats := make(map[string]*FairnessState)

	// Get all priority states
	for priority := range pf.config.Weights {
		state, err := pf.getState(ctx, priority)
		if err != nil {
			continue
		}
		stats[priority] = state
	}

	return stats, nil
}

// Helper methods

func (pf *PriorityFairness) normalizeWeights(demands map[string]int64) map[string]float64 {
	normalized := make(map[string]float64)
	totalWeight := 0.0

	// Calculate total weight for active priorities
	for priority, demand := range demands {
		if demand > 0 {
			weight := pf.getWeight(priority)
			normalized[priority] = weight
			totalWeight += weight
		}
	}

	// Normalize to sum to 1
	if totalWeight > 0 {
		for priority := range normalized {
			normalized[priority] /= totalWeight
		}
	}

	return normalized
}

func (pf *PriorityFairness) getWeight(priority string) float64 {
	pf.mu.RLock()
	if weight, ok := pf.weightCache[priority]; ok {
		pf.mu.RUnlock()
		return weight
	}
	pf.mu.RUnlock()

	// Check config
	if weight, ok := pf.config.Weights[priority]; ok {
		pf.mu.Lock()
		pf.weightCache[priority] = weight
		pf.mu.Unlock()
		return weight
	}

	return 1.0 // Default weight
}

func (pf *PriorityFairness) getState(ctx context.Context, priority string) (*FairnessState, error) {
	key := fmt.Sprintf("fairness:state:%s", priority)

	data, err := pf.redis.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	state := &FairnessState{
		Priority: priority,
	}

	// Parse Redis hash data
	if v, ok := data["allocated"]; ok {
		fmt.Sscanf(v, "%d", &state.AllocatedTokens)
	}
	if v, ok := data["consumed"]; ok {
		fmt.Sscanf(v, "%d", &state.ConsumedTokens)
	}
	if v, ok := data["queued"]; ok {
		fmt.Sscanf(v, "%d", &state.QueuedRequests)
	}
	if v, ok := data["last_scheduled"]; ok {
		var ts int64
		fmt.Sscanf(v, "%d", &ts)
		state.LastScheduled = time.UnixMilli(ts)
	}

	// Check starvation
	if time.Since(state.LastScheduled) > pf.config.MaxWaitTime {
		state.StarvationRisk = true
	}

	return state, nil
}

func (pf *PriorityFairness) updateState(ctx context.Context, priority string, consumedTokens int64) error {
	key := fmt.Sprintf("fairness:state:%s", priority)

	pipe := pf.redis.Pipeline()
	pipe.HIncrBy(ctx, key, "consumed", consumedTokens)
	pipe.HSet(ctx, key, "last_scheduled", time.Now().UnixMilli())
	pipe.Expire(ctx, key, pf.config.AdaptiveWindow)

	_, err := pipe.Exec(ctx)
	return err
}

func (pf *PriorityFairness) calculateFairShare(ctx context.Context, weight float64) int64 {
	// In production, this would calculate based on total system capacity
	// and current demand across all priorities
	baseShare := int64(1000) // Base tokens per window
	return int64(float64(baseShare) * weight)
}

func (pf *PriorityFairness) calculateDelay(state *FairnessState, fairShare, requested int64) time.Duration {
	if state.ConsumedTokens + requested <= fairShare {
		return 0
	}

	// Calculate how long to wait for fair share to replenish
	tokensOverFairShare := (state.ConsumedTokens + requested) - fairShare
	refillRate := float64(fairShare) / float64(pf.config.AdaptiveWindow.Seconds())
	waitSeconds := float64(tokensOverFairShare) / refillRate

	return time.Duration(waitSeconds) * time.Second
}

func (pf *PriorityFairness) preventStarvation(ctx context.Context, allocations map[string]int64, demands map[string]int64) map[string]int64 {
	// Check for starving priorities
	for priority, demand := range demands {
		if demand > 0 && allocations[priority] == 0 {
			state, err := pf.getState(ctx, priority)
			if err != nil {
				continue
			}

			if state.StarvationRisk {
				// Give emergency allocation to prevent starvation
				emergencyTokens := int64(math.Min(float64(demand), 10))
				allocations[priority] = emergencyTokens

				pf.logger.Warn("Preventing starvation",
					zap.String("priority", priority),
					zap.Int64("emergency_tokens", emergencyTokens))
			}
		}
	}

	return allocations
}

func (pf *PriorityFairness) recordAllocationMetrics(allocations map[string]int64, demands map[string]int64) {
	for priority, allocated := range allocations {
		efficiency := float64(allocated) / float64(demands[priority])

		pf.logger.Debug("Token allocation",
			zap.String("priority", priority),
			zap.Int64("allocated", allocated),
			zap.Int64("demanded", demands[priority]),
			zap.Float64("efficiency", efficiency))
	}
}