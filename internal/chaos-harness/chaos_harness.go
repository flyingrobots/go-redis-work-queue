// Copyright 2025 James Ross
package chaosharness

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

// ChaosHarness manages chaos testing functionality
type ChaosHarness struct {
	injectorManager *FaultInjectorManager
	scenarioRunner  *ScenarioRunner
	apiHandler      *APIHandler
	logger          *zap.Logger
	config          *Config

	// Lifecycle
	stopOnce sync.Once
	stopped  chan struct{}
}

// Config defines chaos harness configuration
type Config struct {
	Enabled         bool          `json:"enabled"`
	AllowProduction bool          `json:"allow_production"`
	DefaultTTL      time.Duration `json:"default_ttl"`
	MaxTTL          time.Duration `json:"max_ttl"`
	APIPrefix       string        `json:"api_prefix"`
}

// DefaultConfig returns default chaos harness configuration
func DefaultConfig() *Config {
	return &Config{
		Enabled:         true,
		AllowProduction: false,
		DefaultTTL:      5 * time.Minute,
		MaxTTL:          1 * time.Hour,
		APIPrefix:       "/api/v1",
	}
}

// NewChaosHarness creates a new chaos harness
func NewChaosHarness(logger *zap.Logger, config *Config) *ChaosHarness {
	if config == nil {
		config = DefaultConfig()
	}

	injectorConfig := &InjectorConfig{
		Enabled:         config.Enabled,
		DefaultTTL:      config.DefaultTTL,
		MaxTTL:          config.MaxTTL,
		AllowProduction: config.AllowProduction,
	}

	injectorManager := NewFaultInjectorManager(logger, injectorConfig)
	scenarioRunner := NewScenarioRunner(injectorManager, logger)
	apiHandler := NewAPIHandler(injectorManager, scenarioRunner, logger)

	return &ChaosHarness{
		injectorManager: injectorManager,
		scenarioRunner:  scenarioRunner,
		apiHandler:      apiHandler,
		logger:          logger,
		config:          config,
		stopped:         make(chan struct{}),
	}
}

// RegisterRoutes registers chaos harness API routes
func (ch *ChaosHarness) RegisterRoutes(router *mux.Router) {
	if !ch.config.Enabled {
		ch.logger.Info("Chaos harness is disabled")
		return
	}

	subrouter := router.PathPrefix(ch.config.APIPrefix).Subrouter()
	ch.apiHandler.RegisterRoutes(subrouter)

	ch.logger.Info("Chaos harness API registered",
		zap.String("prefix", ch.config.APIPrefix))
}

// InjectLatency injects latency for a scope
func (ch *ChaosHarness) InjectLatency(ctx context.Context, scope InjectorScope, scopeValue string) time.Duration {
	if !ch.config.Enabled {
		return 0
	}
	return ch.injectorManager.InjectLatency(ctx, scope, scopeValue)
}

// InjectError injects an error for a scope
func (ch *ChaosHarness) InjectError(ctx context.Context, scope InjectorScope, scopeValue string) error {
	if !ch.config.Enabled {
		return nil
	}
	return ch.injectorManager.InjectError(ctx, scope, scopeValue)
}

// InjectPanic injects a panic for a scope
func (ch *ChaosHarness) InjectPanic(ctx context.Context, scope InjectorScope, scopeValue string) {
	if !ch.config.Enabled {
		return
	}
	ch.injectorManager.InjectPanic(ctx, scope, scopeValue)
}

// InjectPartialFail injects partial failure for a scope
func (ch *ChaosHarness) InjectPartialFail(ctx context.Context, scope InjectorScope, scopeValue string, items []interface{}) []interface{} {
	if !ch.config.Enabled {
		return items
	}
	return ch.injectorManager.InjectPartialFail(ctx, scope, scopeValue, items)
}

// ShouldInject checks if a fault should be injected
func (ch *ChaosHarness) ShouldInject(ctx context.Context, scope InjectorScope, scopeValue string, injectorType InjectorType) (bool, *FaultInjector) {
	if !ch.config.Enabled {
		return false, nil
	}
	return ch.injectorManager.ShouldInject(ctx, scope, scopeValue, injectorType)
}

// RunScenario executes a chaos scenario
func (ch *ChaosHarness) RunScenario(ctx context.Context, scenario *ChaosScenario) error {
	if !ch.config.Enabled {
		return fmt.Errorf("chaos harness is disabled")
	}
	return ch.scenarioRunner.RunScenario(ctx, scenario)
}

// GetStatus returns chaos harness status
func (ch *ChaosHarness) GetStatus() map[string]interface{} {
	injectors := ch.injectorManager.GetActiveInjectors()
	scenarios := ch.scenarioRunner.GetRunningScenarios()

	return map[string]interface{}{
		"enabled":           ch.config.Enabled,
		"allow_production":  ch.config.AllowProduction,
		"active_injectors":  len(injectors),
		"running_scenarios": len(scenarios),
		"injectors":         injectors,
		"scenarios":         scenarios,
	}
}

// Stop stops the chaos harness
func (ch *ChaosHarness) Stop() error {
	ch.stopOnce.Do(func() {
		ch.logger.Info("Stopping chaos harness")

		// Abort all running scenarios
		for _, scenario := range ch.scenarioRunner.GetRunningScenarios() {
			ch.scenarioRunner.AbortScenario(scenario.ID)
		}

		// Stop injector manager
		ch.injectorManager.Stop()

		close(ch.stopped)
	})

	return nil
}

// Wait waits for chaos harness to stop
func (ch *ChaosHarness) Wait() {
	<-ch.stopped
}
