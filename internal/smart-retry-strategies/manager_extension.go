// Copyright 2025 James Ross
package smartretry

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"go.uber.org/zap"
)

// AddPolicy adds a new retry policy
func (m *manager) AddPolicy(policy RetryPolicy) error {
	if err := policy.Validate(); err != nil {
		return NewPolicyError("Invalid policy", err)
	}

	m.policyMu.Lock()
	defer m.policyMu.Unlock()

	// Check for duplicate names
	for _, existing := range m.strategy.Policies {
		if existing.Name == policy.Name {
			return ErrPolicyAlreadyExists(policy.Name)
		}
	}

	m.strategy.Policies = append(m.strategy.Policies, policy)

	// Sort by priority
	sort.Slice(m.strategy.Policies, func(i, j int) bool {
		return m.strategy.Policies[i].Priority > m.strategy.Policies[j].Priority
	})

	m.logger.Info("Policy added",
		zap.String("name", policy.Name),
		zap.Int("priority", policy.Priority))

	return m.persistStrategy()
}

// RemovePolicy removes a retry policy by name
func (m *manager) RemovePolicy(name string) error {
	m.policyMu.Lock()
	defer m.policyMu.Unlock()

	for i, policy := range m.strategy.Policies {
		if policy.Name == name {
			// Remove the policy
			m.strategy.Policies = append(m.strategy.Policies[:i], m.strategy.Policies[i+1:]...)

			m.logger.Info("Policy removed", zap.String("name", name))
			return m.persistStrategy()
		}
	}

	return NewPolicyError(fmt.Sprintf("Policy '%s' not found", name), ErrPolicyNotFound)
}

// UpdateGuardrails updates the policy guardrails
func (m *manager) UpdateGuardrails(guardrails PolicyGuardrails) error {
	m.policyMu.Lock()
	defer m.policyMu.Unlock()

	// Validate guardrails
	if guardrails.MaxAttempts <= 0 {
		return NewConfigError("MaxAttempts must be positive", ErrInvalidConfig)
	}

	if guardrails.MaxDelayMs <= 0 {
		return NewConfigError("MaxDelayMs must be positive", ErrInvalidConfig)
	}

	if guardrails.MaxBudgetPercent < 0 || guardrails.MaxBudgetPercent > 100 {
		return NewConfigError("MaxBudgetPercent must be between 0 and 100", ErrInvalidConfig)
	}

	m.strategy.Guardrails = guardrails

	m.logger.Info("Guardrails updated",
		zap.Int("max_attempts", guardrails.MaxAttempts),
		zap.Int64("max_delay_ms", guardrails.MaxDelayMs),
		zap.Float64("max_budget_percent", guardrails.MaxBudgetPercent))

	return m.persistStrategy()
}

// GetStrategy returns the current retry strategy
func (m *manager) GetStrategy() (*RetryStrategy, error) {
	m.policyMu.RLock()
	defer m.policyMu.RUnlock()

	// Return a copy to prevent external modification
	strategyData, err := json.Marshal(m.strategy)
	if err != nil {
		return nil, NewDataError("Failed to serialize strategy", err)
	}

	var strategyCopy RetryStrategy
	if err := json.Unmarshal(strategyData, &strategyCopy); err != nil {
		return nil, NewDataError("Failed to deserialize strategy", err)
	}

	return &strategyCopy, nil
}

// UpdateStrategy updates the entire retry strategy
func (m *manager) UpdateStrategy(strategy *RetryStrategy) error {
	if strategy == nil {
		return NewConfigError("Strategy cannot be nil", ErrInvalidConfig)
	}

	// Validate the strategy
	if strategy.BayesianThreshold < 0 || strategy.BayesianThreshold > 1 {
		return NewConfigError("BayesianThreshold must be between 0 and 1", ErrInvalidThreshold)
	}

	// Validate all policies
	for i, policy := range strategy.Policies {
		if err := policy.Validate(); err != nil {
			return NewPolicyError(fmt.Sprintf("Policy %d invalid", i), err)
		}
	}

	m.policyMu.Lock()
	defer m.policyMu.Unlock()

	m.strategy = strategy

	m.logger.Info("Strategy updated",
		zap.String("name", strategy.Name),
		zap.Bool("enabled", strategy.Enabled),
		zap.Int("policies", len(strategy.Policies)),
		zap.Float64("bayesian_threshold", strategy.BayesianThreshold),
		zap.Bool("ml_enabled", strategy.MLEnabled))

	return m.persistStrategy()
}

// persistStrategy saves the current strategy to Redis
func (m *manager) persistStrategy() error {
	ctx := context.Background()
	strategyKey := "retry:strategy"

	strategyData, err := json.Marshal(m.strategy)
	if err != nil {
		return NewDataError("Failed to marshal strategy", err)
	}

	err = m.redis.Set(ctx, strategyKey, strategyData, 0).Err() // No expiration
	if err != nil {
		return NewDataError("Failed to persist strategy", err)
	}

	return nil
}

// loadStrategy loads the strategy from Redis
func (m *manager) loadStrategy() error {
	ctx := context.Background()
	strategyKey := "retry:strategy"

	strategyData, err := m.redis.Get(ctx, strategyKey).Result()
	if err != nil {
		// If not found, keep default strategy
		return nil
	}

	var strategy RetryStrategy
	if err := json.Unmarshal([]byte(strategyData), &strategy); err != nil {
		m.logger.Warn("Failed to load strategy from Redis, using default", zap.Error(err))
		return nil
	}

	m.strategy = &strategy
	m.logger.Info("Strategy loaded from Redis", zap.String("name", strategy.Name))

	return nil
}

// Close closes the manager and cleans up resources
func (m *manager) Close() error {
	if m.redis != nil {
		return m.redis.Close()
	}
	return nil
}

// GetHealthStatus returns the health status of the retry system
func (m *manager) GetHealthStatus() map[string]interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health := map[string]interface{}{
		"service":   "smart-retry-strategies",
		"timestamp": time.Now(),
		"status":    "ok",
	}

	// Check Redis connectivity
	if err := m.redis.Ping(ctx).Err(); err != nil {
		health["status"] = "degraded"
		health["redis_error"] = err.Error()
	} else {
		health["redis"] = "connected"
	}

	// Add strategy info
	m.policyMu.RLock()
	health["strategy_name"] = m.strategy.Name
	health["strategy_enabled"] = m.strategy.Enabled
	health["policy_count"] = len(m.strategy.Policies)
	health["ml_enabled"] = m.strategy.MLEnabled
	m.policyMu.RUnlock()

	// Add ML model info if available
	m.mlMu.RLock()
	if m.mlModel != nil {
		health["ml_model"] = map[string]interface{}{
			"version":   m.mlModel.Version,
			"type":      m.mlModel.ModelType,
			"enabled":   m.mlModel.Enabled,
			"accuracy":  m.mlModel.Accuracy,
			"trained_at": m.mlModel.TrainedAt,
		}
	}
	m.mlMu.RUnlock()

	return health
}

// GetMetrics returns system metrics
func (m *manager) GetMetrics() map[string]interface{} {
	metrics := map[string]interface{}{
		"timestamp": time.Now(),
	}

	// Cache metrics
	m.cache.mu.RLock()
	metrics["cache_entries"] = len(m.cache.entries)
	metrics["cache_enabled"] = m.cache.enabled
	metrics["cache_max_entries"] = m.cache.maxEntries
	m.cache.mu.RUnlock()

	// Strategy metrics
	m.policyMu.RLock()
	metrics["policy_count"] = len(m.strategy.Policies)
	metrics["bayesian_threshold"] = m.strategy.BayesianThreshold
	metrics["max_attempts"] = m.strategy.Guardrails.MaxAttempts
	metrics["max_delay_ms"] = m.strategy.Guardrails.MaxDelayMs
	m.policyMu.RUnlock()

	return metrics
}

// EmergencyStop activates emergency stop mode
func (m *manager) EmergencyStop(reason string) error {
	m.policyMu.Lock()
	defer m.policyMu.Unlock()

	m.strategy.Guardrails.EmergencyStop = true

	m.logger.Warn("Emergency stop activated", zap.String("reason", reason))

	// Emit event
	m.emitEvent(RetryEvent{
		ID:        fmt.Sprintf("emergency_stop_%d", time.Now().UnixNano()),
		Type:      EventTypeGuardrailTriggered,
		Message:   fmt.Sprintf("Emergency stop activated: %s", reason),
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"reason": reason,
			"action": "emergency_stop",
		},
	})

	return m.persistStrategy()
}

// ResetEmergencyStop deactivates emergency stop mode
func (m *manager) ResetEmergencyStop() error {
	m.policyMu.Lock()
	defer m.policyMu.Unlock()

	m.strategy.Guardrails.EmergencyStop = false

	m.logger.Info("Emergency stop reset")

	return m.persistStrategy()
}

// GetRecommendationWithExplanation returns a recommendation with detailed explanation
func (m *manager) GetRecommendationWithExplanation(features RetryFeatures) (*RetryRecommendation, map[string]interface{}, error) {
	recommendation, err := m.GetRecommendation(features)
	if err != nil {
		return nil, nil, err
	}

	explanation := map[string]interface{}{
		"timestamp":     time.Now(),
		"features_used": features,
		"method":        recommendation.Method,
		"rationale":     recommendation.Rationale,
		"confidence":    recommendation.Confidence,
	}

	// Add method-specific explanation
	switch recommendation.Method {
	case "rules":
		explanation["matching_policy"] = m.getMatchingPolicyName(features)
	case "bayesian":
		explanation["bayesian_model"] = m.getBayesianModelInfo(features.JobType, features.ErrorClass)
	case "ml":
		if m.mlModel != nil {
			explanation["ml_model"] = map[string]interface{}{
				"version":  m.mlModel.Version,
				"type":     m.mlModel.ModelType,
				"accuracy": m.mlModel.Accuracy,
			}
		}
	case "guardrails":
		explanation["guardrail_triggered"] = true
		explanation["guardrails"] = m.strategy.Guardrails
	}

	return recommendation, explanation, nil
}

// Helper methods for explanation

func (m *manager) getMatchingPolicyName(features RetryFeatures) string {
	m.policyMu.RLock()
	defer m.policyMu.RUnlock()

	// Sort policies by priority
	policies := make([]RetryPolicy, len(m.strategy.Policies))
	copy(policies, m.strategy.Policies)
	sort.Slice(policies, func(i, j int) bool {
		return policies[i].Priority > policies[j].Priority
	})

	for _, policy := range policies {
		if m.policyMatches(policy, features) {
			return policy.Name
		}
	}

	return "default"
}

func (m *manager) getBayesianModelInfo(jobType, errorClass string) map[string]interface{} {
	model, err := m.getBayesianModel(jobType, errorClass)
	if err != nil || model == nil {
		return map[string]interface{}{
			"available": false,
		}
	}

	return map[string]interface{}{
		"available":    true,
		"sample_count": model.SampleCount,
		"confidence":   model.Confidence,
		"last_updated": model.LastUpdated,
		"buckets":      len(model.Buckets),
	}
}