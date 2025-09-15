// Copyright 2025 James Ross
package chaosharness

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
)

// FaultInjectorManager manages fault injectors
type FaultInjectorManager struct {
	injectors map[string]*FaultInjector
	config    *InjectorConfig
	logger    *zap.Logger
	mu        sync.RWMutex
	random    *rand.Rand

	// Cleanup goroutine
	cleanupStop chan struct{}
	cleanupWG   sync.WaitGroup
}

// NewFaultInjectorManager creates a new fault injector manager
func NewFaultInjectorManager(logger *zap.Logger, config *InjectorConfig) *FaultInjectorManager {
	if config == nil {
		config = &InjectorConfig{
			Enabled:    true,
			DefaultTTL: 5 * time.Minute,
			MaxTTL:     1 * time.Hour,
		}
	}

	fim := &FaultInjectorManager{
		injectors:   make(map[string]*FaultInjector),
		config:      config,
		logger:      logger,
		random:      rand.New(rand.NewSource(time.Now().UnixNano())),
		cleanupStop: make(chan struct{}),
	}

	// Start cleanup goroutine
	fim.cleanupWG.Add(1)
	go fim.cleanupExpired()

	return fim
}

// AddInjector adds a new fault injector
func (fim *FaultInjectorManager) AddInjector(injector *FaultInjector) error {
	if !fim.config.Enabled {
		return fmt.Errorf("fault injection is disabled")
	}

	// Validate injector
	if err := fim.validateInjector(injector); err != nil {
		return fmt.Errorf("invalid injector: %w", err)
	}

	// Set defaults
	if injector.TTL == 0 {
		injector.TTL = fim.config.DefaultTTL
	}
	if injector.TTL > fim.config.MaxTTL {
		injector.TTL = fim.config.MaxTTL
	}

	// Calculate expiry
	if injector.TTL > 0 {
		expiresAt := time.Now().Add(injector.TTL)
		injector.ExpiresAt = &expiresAt
	}

	injector.CreatedAt = time.Now()

	fim.mu.Lock()
	defer fim.mu.Unlock()

	fim.injectors[injector.ID] = injector

	fim.logger.Info("Added fault injector",
		zap.String("id", injector.ID),
		zap.String("type", string(injector.Type)),
		zap.String("scope", string(injector.Scope)),
		zap.Float64("probability", injector.Probability),
		zap.Duration("ttl", injector.TTL))

	return nil
}

// RemoveInjector removes a fault injector
func (fim *FaultInjectorManager) RemoveInjector(id string) error {
	fim.mu.Lock()
	defer fim.mu.Unlock()

	if _, exists := fim.injectors[id]; !exists {
		return fmt.Errorf("injector not found: %s", id)
	}

	delete(fim.injectors, id)

	fim.logger.Info("Removed fault injector", zap.String("id", id))
	return nil
}

// ShouldInject determines if a fault should be injected
func (fim *FaultInjectorManager) ShouldInject(ctx context.Context, scope InjectorScope, scopeValue string, injectorType InjectorType) (bool, *FaultInjector) {
	if !fim.config.Enabled {
		return false, nil
	}

	fim.mu.RLock()
	defer fim.mu.RUnlock()

	for _, injector := range fim.injectors {
		if !injector.Enabled {
			continue
		}

		// Check if expired
		if injector.ExpiresAt != nil && time.Now().After(*injector.ExpiresAt) {
			continue
		}

		// Check type match
		if injector.Type != injectorType {
			continue
		}

		// Check scope match
		if !fim.matchesScope(injector, scope, scopeValue) {
			continue
		}

		// Check probability
		if fim.random.Float64() <= injector.Probability {
			return true, injector
		}
	}

	return false, nil
}

// InjectLatency injects latency if configured
func (fim *FaultInjectorManager) InjectLatency(ctx context.Context, scope InjectorScope, scopeValue string) time.Duration {
	should, injector := fim.ShouldInject(ctx, scope, scopeValue, InjectorLatency)
	if !should || injector == nil {
		return 0
	}

	// Get latency from parameters
	latencyMs, ok := injector.Parameters["latency_ms"].(float64)
	if !ok {
		latencyMs = 100 // Default 100ms
	}

	// Add jitter if configured
	jitterMs, hasJitter := injector.Parameters["jitter_ms"].(float64)
	if hasJitter && jitterMs > 0 {
		jitter := (fim.random.Float64() - 0.5) * 2 * jitterMs
		latencyMs += jitter
	}

	if latencyMs < 0 {
		latencyMs = 0
	}

	latency := time.Duration(latencyMs) * time.Millisecond

	fim.logger.Debug("Injecting latency",
		zap.String("injector_id", injector.ID),
		zap.String("scope", string(scope)),
		zap.String("scope_value", scopeValue),
		zap.Duration("latency", latency))

	time.Sleep(latency)
	return latency
}

// InjectError injects an error if configured
func (fim *FaultInjectorManager) InjectError(ctx context.Context, scope InjectorScope, scopeValue string) error {
	should, injector := fim.ShouldInject(ctx, scope, scopeValue, InjectorError)
	if !should || injector == nil {
		return nil
	}

	// Get error message from parameters
	errorMsg, ok := injector.Parameters["error_message"].(string)
	if !ok {
		errorMsg = "injected fault error"
	}

	fim.logger.Debug("Injecting error",
		zap.String("injector_id", injector.ID),
		zap.String("scope", string(scope)),
		zap.String("scope_value", scopeValue),
		zap.String("error", errorMsg))

	return &InjectedError{
		Message:    errorMsg,
		InjectorID: injector.ID,
		Scope:      scope,
		ScopeValue: scopeValue,
	}
}

// InjectPanic injects a panic if configured
func (fim *FaultInjectorManager) InjectPanic(ctx context.Context, scope InjectorScope, scopeValue string) {
	should, injector := fim.ShouldInject(ctx, scope, scopeValue, InjectorPanic)
	if !should || injector == nil {
		return
	}

	panicMsg, ok := injector.Parameters["panic_message"].(string)
	if !ok {
		panicMsg = "injected panic"
	}

	fim.logger.Warn("Injecting panic",
		zap.String("injector_id", injector.ID),
		zap.String("scope", string(scope)),
		zap.String("scope_value", scopeValue),
		zap.String("message", panicMsg))

	panic(fmt.Sprintf("[CHAOS] %s (injector: %s)", panicMsg, injector.ID))
}

// InjectPartialFail injects partial failure
func (fim *FaultInjectorManager) InjectPartialFail(ctx context.Context, scope InjectorScope, scopeValue string, items []interface{}) []interface{} {
	should, injector := fim.ShouldInject(ctx, scope, scopeValue, InjectorPartialFail)
	if !should || injector == nil {
		return items
	}

	// Get failure rate from parameters
	failRate, ok := injector.Parameters["fail_rate"].(float64)
	if !ok {
		failRate = 0.5 // Default 50%
	}

	// Filter items based on failure rate
	var processed []interface{}
	failed := 0
	for _, item := range items {
		if fim.random.Float64() > failRate {
			processed = append(processed, item)
		} else {
			failed++
		}
	}

	fim.logger.Debug("Injected partial failure",
		zap.String("injector_id", injector.ID),
		zap.String("scope", string(scope)),
		zap.String("scope_value", scopeValue),
		zap.Int("total_items", len(items)),
		zap.Int("failed_items", failed))

	return processed
}

// GetActiveInjectors returns currently active injectors
func (fim *FaultInjectorManager) GetActiveInjectors() []*FaultInjector {
	fim.mu.RLock()
	defer fim.mu.RUnlock()

	var active []*FaultInjector
	now := time.Now()

	for _, injector := range fim.injectors {
		if injector.Enabled {
			if injector.ExpiresAt == nil || now.Before(*injector.ExpiresAt) {
				active = append(active, injector)
			}
		}
	}

	return active
}

// ClearAll removes all injectors
func (fim *FaultInjectorManager) ClearAll() {
	fim.mu.Lock()
	defer fim.mu.Unlock()

	count := len(fim.injectors)
	fim.injectors = make(map[string]*FaultInjector)

	fim.logger.Info("Cleared all fault injectors", zap.Int("count", count))
}

// Stop stops the injector manager
func (fim *FaultInjectorManager) Stop() {
	close(fim.cleanupStop)
	fim.cleanupWG.Wait()
	fim.ClearAll()
}

// Helper methods

func (fim *FaultInjectorManager) validateInjector(injector *FaultInjector) error {
	if injector.ID == "" {
		return fmt.Errorf("injector ID is required")
	}

	if injector.Probability < 0 || injector.Probability > 1 {
		return fmt.Errorf("probability must be between 0 and 1")
	}

	// Validate scope
	switch injector.Scope {
	case ScopeGlobal:
		// Global scope doesn't need scope value
	case ScopeWorker, ScopeQueue, ScopeTenant:
		if injector.ScopeValue == "" {
			return fmt.Errorf("scope value required for scope %s", injector.Scope)
		}
	default:
		return fmt.Errorf("invalid scope: %s", injector.Scope)
	}

	// Validate type-specific parameters
	switch injector.Type {
	case InjectorLatency:
		if injector.Parameters == nil {
			injector.Parameters = make(map[string]interface{})
		}
		if _, ok := injector.Parameters["latency_ms"]; !ok {
			injector.Parameters["latency_ms"] = 100.0 // Default 100ms
		}
	case InjectorPartialFail:
		if injector.Parameters == nil {
			injector.Parameters = make(map[string]interface{})
		}
		if _, ok := injector.Parameters["fail_rate"]; !ok {
			injector.Parameters["fail_rate"] = 0.5 // Default 50%
		}
	}

	return nil
}

func (fim *FaultInjectorManager) matchesScope(injector *FaultInjector, scope InjectorScope, scopeValue string) bool {
	// Global scope matches everything
	if injector.Scope == ScopeGlobal {
		return true
	}

	// Check exact scope match
	if injector.Scope != scope {
		return false
	}

	// For non-global scopes, check value match
	if injector.Scope != ScopeGlobal {
		return injector.ScopeValue == "" || injector.ScopeValue == scopeValue
	}

	return false
}

func (fim *FaultInjectorManager) cleanupExpired() {
	defer fim.cleanupWG.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fim.removeExpired()
		case <-fim.cleanupStop:
			return
		}
	}
}

func (fim *FaultInjectorManager) removeExpired() {
	fim.mu.Lock()
	defer fim.mu.Unlock()

	now := time.Now()
	var expired []string

	for id, injector := range fim.injectors {
		if injector.ExpiresAt != nil && now.After(*injector.ExpiresAt) {
			expired = append(expired, id)
		}
	}

	for _, id := range expired {
		delete(fim.injectors, id)
		fim.logger.Info("Removed expired injector", zap.String("id", id))
	}
}
