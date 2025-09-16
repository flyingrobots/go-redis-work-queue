// Copyright 2025 James Ross
package chaosharness

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// LoadGenerator generates load for chaos testing
type LoadGenerator struct {
	logger *zap.Logger
	random *rand.Rand

	// Current load generation
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.Mutex

	// Metrics
	totalRequests atomic.Int64
	successful    atomic.Int64
	failed        atomic.Int64
}

// NewLoadGenerator creates a new load generator
func NewLoadGenerator(logger *zap.Logger) *LoadGenerator {
	return &LoadGenerator{
		logger: logger,
		random: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Start starts load generation
func (lg *LoadGenerator) Start(ctx context.Context, config *LoadConfig, metrics *ScenarioMetrics) {
	lg.Stop() // Stop any existing generation

	lg.mu.Lock()
	ctx, lg.cancel = context.WithCancel(ctx)
	lg.mu.Unlock()

	lg.logger.Info("Starting load generation",
		zap.Int("rps", config.RPS),
		zap.String("pattern", string(config.Pattern)))

	lg.wg.Add(1)
	go lg.generate(ctx, config, metrics)
}

// Stop stops load generation
func (lg *LoadGenerator) Stop() {
	lg.mu.Lock()
	if lg.cancel != nil {
		lg.cancel()
		lg.cancel = nil
	}
	lg.mu.Unlock()

	lg.wg.Wait()
}

// generate generates load based on pattern
func (lg *LoadGenerator) generate(ctx context.Context, config *LoadConfig, metrics *ScenarioMetrics) {
	defer lg.wg.Done()

	ticker := time.NewTicker(time.Second / time.Duration(config.RPS))
	defer ticker.Stop()

	startTime := time.Now()
	var requestCount int64

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Calculate current RPS based on pattern
			elapsed := time.Since(startTime).Seconds()
			currentRPS := lg.calculateRPS(config, elapsed)

			// Adjust ticker if needed
			if currentRPS > 0 {
				ticker.Reset(time.Second / time.Duration(currentRPS))
				
				// Generate request
				lg.generateRequest(ctx, metrics)
				requestCount++
			}

			// Handle burst patterns
			if config.Pattern == LoadSpike && requestCount%(int64(config.RPS)*10) == 0 {
				lg.generateBurst(ctx, config, metrics)
			}
		}
	}
}

// calculateRPS calculates RPS based on pattern
func (lg *LoadGenerator) calculateRPS(config *LoadConfig, elapsed float64) int {
	baseRPS := float64(config.RPS)

	switch config.Pattern {
	case LoadConstant:
		return config.RPS

	case LoadLinear:
		// Linear ramp up over 60 seconds
		rampTime := 60.0
		if elapsed < rampTime {
			return int(baseRPS * (elapsed / rampTime))
		}
		return config.RPS

	case LoadSine:
		// Sine wave with 60 second period
		period := 60.0
		amplitude := baseRPS * 0.5
		return int(baseRPS + amplitude*math.Sin(2*math.Pi*elapsed/period))

	case LoadRandom:
		// Random variation ±50%
		variation := 0.5
		return int(baseRPS * (1 + (lg.random.Float64()-0.5)*variation*2))

	case LoadSpike:
		// Normal load with periodic spikes
		return config.RPS

	default:
		return config.RPS
	}
}

// generateRequest simulates a single request
func (lg *LoadGenerator) generateRequest(ctx context.Context, metrics *ScenarioMetrics) {
	lg.totalRequests.Add(1)

	// Simulate request processing
	// In a real implementation, this would make actual requests
	success := lg.random.Float64() > 0.1 // 90% success rate baseline

	if success {
		lg.successful.Add(1)
	} else {
		lg.failed.Add(1)
	}

	// Update metrics
	if metrics != nil {
		metrics.mu.Lock()
		metrics.TotalRequests = lg.totalRequests.Load()
		metrics.SuccessfulRequests = lg.successful.Load()
		metrics.FailedRequests = lg.failed.Load()
		metrics.mu.Unlock()
	}
}

// generateBurst generates a burst of requests
func (lg *LoadGenerator) generateBurst(ctx context.Context, config *LoadConfig, metrics *ScenarioMetrics) {
	burstSize := config.BurstSize
	if burstSize == 0 {
		burstSize = config.RPS * 5 // Default 5x normal rate
	}

	lg.logger.Debug("Generating burst",
		zap.Int("size", burstSize))

	for i := 0; i < burstSize; i++ {
		select {
		case <-ctx.Done():
			return
		default:
			lg.generateRequest(ctx, metrics)
		}
	}
}

// GetStats returns current load generator statistics
func (lg *LoadGenerator) GetStats() map[string]int64 {
	return map[string]int64{
		"total_requests": lg.totalRequests.Load(),
		"successful":     lg.successful.Load(),
		"failed":         lg.failed.Load(),
	}
}

// Reset resets load generator statistics
func (lg *LoadGenerator) Reset() {
	lg.totalRequests.Store(0)
	lg.successful.Store(0)
	lg.failed.Store(0)
}
