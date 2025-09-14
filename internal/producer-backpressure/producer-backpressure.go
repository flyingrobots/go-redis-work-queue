// Copyright 2025 James Ross
package backpressure

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Controller implements the BackpressureController interface
type Controller struct {
	config       BackpressureConfig
	statsProvider StatsProvider
	metrics      *BackpressureMetrics
	logger       *zap.Logger

	// State management
	mu               sync.RWMutex
	started          bool
	stopped          bool
	manualOverride   bool
	emergencyMode    bool
	lastFallbackTime time.Time

	// Circuit breakers per queue
	circuitBreakers map[string]*CircuitBreaker
	cbMu           sync.RWMutex

	// Caching for throttle decisions
	throttleCache map[string]*CachedDecision
	cacheMu       sync.RWMutex
	cacheStats    struct {
		hits   int64
		misses int64
	}

	// Background goroutine management
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Random source for jitter
	rng *rand.Rand
}

// CachedDecision represents a cached throttle decision
type CachedDecision struct {
	Decision  *ThrottleDecision
	ExpiresAt time.Time
}

// NewController creates a new backpressure controller
func NewController(config BackpressureConfig, statsProvider StatsProvider, logger *zap.Logger) (*Controller, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	if statsProvider == nil {
		return nil, fmt.Errorf("stats provider is required")
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	controller := &Controller{
		config:          config,
		statsProvider:   statsProvider,
		metrics:         NewBackpressureMetrics(),
		logger:          logger,
		circuitBreakers: make(map[string]*CircuitBreaker),
		throttleCache:   make(map[string]*CachedDecision),
		rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	// Register metrics
	controller.metrics.Register()

	return controller, nil
}

// Start begins background operations
func (c *Controller) Start(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.started {
		return fmt.Errorf("controller already started")
	}

	if c.stopped {
		return fmt.Errorf("controller was stopped and cannot be restarted")
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.started = true

	// Start background polling if enabled
	if c.config.Polling.Enabled {
		c.wg.Add(1)
		go c.pollingLoop()
	}

	// Start cache cleanup routine
	c.wg.Add(1)
	go c.cacheCleanupLoop()

	c.logger.Info("Backpressure controller started",
		zap.Bool("polling_enabled", c.config.Polling.Enabled),
		zap.Duration("polling_interval", c.config.Polling.Interval))

	return nil
}

// Stop shuts down the controller
func (c *Controller) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.started {
		return fmt.Errorf("controller not started")
	}

	if c.stopped {
		return nil // Already stopped
	}

	c.cancel()
	c.stopped = true

	// Wait for background goroutines to finish
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	// Wait with timeout
	select {
	case <-done:
		c.logger.Info("Backpressure controller stopped gracefully")
	case <-time.After(10 * time.Second):
		c.logger.Warn("Timeout waiting for controller to stop")
	}

	return nil
}

// SuggestThrottle returns throttling recommendation for given priority and queue
func (c *Controller) SuggestThrottle(ctx context.Context, priority Priority, queueName string) (*ThrottleDecision, error) {
	// Input validation
	if !IsValidPriority(priority) {
		return nil, NewBackpressureError("suggest_throttle", queueName, priority, ErrInvalidPriority)
	}
	if !IsValidQueueName(queueName) {
		return nil, NewBackpressureError("suggest_throttle", queueName, priority, ErrInvalidQueue)
	}

	// Check if controller is running
	c.mu.RLock()
	if !c.started || c.stopped {
		c.mu.RUnlock()
		return nil, NewBackpressureError("suggest_throttle", queueName, priority, ErrControllerNotStarted)
	}

	// Check manual overrides
	if c.manualOverride || c.emergencyMode {
		c.mu.RUnlock()
		return &ThrottleDecision{
			Priority:   priority,
			QueueName:  queueName,
			Delay:      0,
			ShouldShed: false,
			Reason:     "manual_override_enabled",
			Timestamp:  time.Now(),
		}, nil
	}
	c.mu.RUnlock()

	// Check cache first
	if decision := c.getCachedDecision(priority, queueName); decision != nil {
		return decision, nil
	}

	// Check circuit breaker
	cb := c.getOrCreateCircuitBreaker(queueName)
	if !cb.ShouldAllow() {
		decision := &ThrottleDecision{
			Priority:   priority,
			QueueName:  queueName,
			Delay:      InfiniteDelay,
			ShouldShed: true,
			Reason:     fmt.Sprintf("circuit_breaker_%s", cb.State.String()),
			Timestamp:  time.Now(),
		}
		c.cacheDecision(priority, queueName, decision)
		c.metrics.ShedEventsTotal.WithLabelValues(priority.String(), queueName).Inc()
		return decision, nil
	}

	// Get current queue stats
	stats, err := c.statsProvider.GetQueueStats(ctx, queueName)
	if err != nil {
		// Handle stats unavailable - use fallback strategy
		decision := c.handleStatsUnavailable(priority, queueName, err)
		c.cacheDecision(priority, queueName, decision)
		return decision, nil
	}

	// Calculate throttle decision based on backlog
	decision := c.calculateThrottleDecision(priority, queueName, stats)
	c.cacheDecision(priority, queueName, decision)

	// Update metrics
	if decision.Delay > 0 {
		c.metrics.ThrottleEventsTotal.WithLabelValues(priority.String(), queueName).Inc()
		c.metrics.ThrottleDelayHistogram.WithLabelValues(priority.String(), queueName).
			Observe(decision.Delay.Seconds())
	}
	if decision.ShouldShed {
		c.metrics.ShedEventsTotal.WithLabelValues(priority.String(), queueName).Inc()
	}
	c.metrics.QueueBacklogGauge.WithLabelValues(queueName).Set(float64(decision.BacklogSize))

	return decision, nil
}

// Run executes work function with automatic throttling
func (c *Controller) Run(ctx context.Context, priority Priority, queueName string, work func() error) error {
	// Get throttle recommendation
	decision, err := c.SuggestThrottle(ctx, priority, queueName)
	if err != nil {
		return err
	}

	// Handle shedding
	if decision.ShouldShed {
		return NewBackpressureError("run", queueName, priority, ErrJobShed)
	}

	// Apply throttling delay
	if decision.Delay > 0 {
		c.logger.Debug("Throttling work execution",
			zap.String("queue", queueName),
			zap.String("priority", priority.String()),
			zap.Duration("delay", decision.Delay),
			zap.String("reason", decision.Reason))

		select {
		case <-time.After(decision.Delay):
			// Delay completed
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// Execute work
	startTime := time.Now()
	err = work()
	duration := time.Since(startTime)

	// Update circuit breaker based on result
	cb := c.getOrCreateCircuitBreaker(queueName)
	if err != nil {
		cb.RecordFailure()
		c.logger.Debug("Work execution failed",
			zap.String("queue", queueName),
			zap.String("priority", priority.String()),
			zap.Duration("duration", duration),
			zap.Error(err))
	} else {
		cb.RecordSuccess()
		c.logger.Debug("Work execution succeeded",
			zap.String("queue", queueName),
			zap.String("priority", priority.String()),
			zap.Duration("duration", duration))
	}

	// Update circuit breaker metrics
	c.metrics.CircuitBreakerState.WithLabelValues(queueName).Set(float64(cb.State))

	return err
}

// ProcessBatch processes multiple jobs with backpressure awareness
func (c *Controller) ProcessBatch(ctx context.Context, jobs []BatchJob) error {
	var errors []error

	for i, job := range jobs {
		err := c.Run(ctx, job.Priority, job.QueueName, job.Work)
		if err != nil {
			if IsShedError(err) {
				// Log shed but continue with other jobs
				c.logger.Debug("Job shed in batch processing",
					zap.Int("job_index", i),
					zap.String("queue", job.QueueName),
					zap.String("priority", job.Priority.String()))
				continue
			}
			errors = append(errors, fmt.Errorf("job %d failed: %w", i, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("batch processing had %d errors: %v", len(errors), errors[0])
	}

	return nil
}

// GetCircuitState returns current circuit breaker state for queue
func (c *Controller) GetCircuitState(queueName string) CircuitState {
	cb := c.getOrCreateCircuitBreaker(queueName)
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.State
}

// SetManualOverride enables/disables manual override mode
func (c *Controller) SetManualOverride(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.manualOverride = enabled

	c.logger.Info("Manual override changed",
		zap.Bool("enabled", enabled))
}

// Health returns controller health status
func (c *Controller) Health() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	c.cacheMu.RLock()
	cacheHits := c.cacheStats.hits
	cacheMisses := c.cacheStats.misses
	cacheSize := len(c.throttleCache)
	c.cacheMu.RUnlock()

	c.cbMu.RLock()
	circuitStates := make(map[string]string)
	for queue, cb := range c.circuitBreakers {
		cb.mu.RLock()
		circuitStates[queue] = cb.State.String()
		cb.mu.RUnlock()
	}
	c.cbMu.RUnlock()

	var cacheHitRate float64
	if cacheHits+cacheMisses > 0 {
		cacheHitRate = float64(cacheHits) / float64(cacheHits+cacheMisses)
	}

	return map[string]interface{}{
		"started":          c.started,
		"stopped":          c.stopped,
		"manual_override":  c.manualOverride,
		"emergency_mode":   c.emergencyMode,
		"cache_hit_rate":   cacheHitRate,
		"cache_size":       cacheSize,
		"circuit_states":   circuitStates,
		"polling_enabled":  c.config.Polling.Enabled,
		"last_fallback":    c.lastFallbackTime,
	}
}

// calculateThrottleDecision determines throttling based on queue stats
func (c *Controller) calculateThrottleDecision(priority Priority, queueName string, stats *QueueStats) *ThrottleDecision {
	window := c.config.Thresholds.GetBacklogWindow(priority)
	current := stats.BacklogCount

	var delay time.Duration
	var shouldShed bool
	var reason string

	switch {
	case current <= window.Green:
		delay = 0
		reason = "backlog_green"

	case current <= window.Yellow:
		// Light throttling: 10ms to 500ms based on position in window
		ratio := float64(current-window.Green) / float64(window.Yellow-window.Green)
		delay = time.Duration(10+490*ratio) * time.Millisecond
		reason = "backlog_yellow"

	default:
		// Heavy throttling: 500ms to 5s, with shedding for low priority
		ratio := math.Min(1.0, float64(current-window.Yellow)/float64(window.Red-window.Yellow))
		baseDelay := time.Duration(500+4500*ratio) * time.Millisecond

		// Apply priority scaling
		switch priority {
		case HighPriority:
			delay = time.Duration(float64(baseDelay) * 0.5) // 50% of base delay
			reason = "backlog_red_high_priority"
		case MediumPriority:
			delay = baseDelay // Full delay
			reason = "backlog_red_medium_priority"
		case LowPriority:
			if ratio > 0.8 {
				delay = InfiniteDelay
				shouldShed = true
				reason = "backlog_red_shed_low_priority"
			} else {
				delay = time.Duration(float64(baseDelay) * 1.5) // 150% of base delay
				reason = "backlog_red_low_priority"
			}
		}
	}

	return &ThrottleDecision{
		Priority:    priority,
		QueueName:   queueName,
		Delay:       delay,
		ShouldShed:  shouldShed,
		Reason:      reason,
		Timestamp:   time.Now(),
		BacklogSize: current,
	}
}

// handleStatsUnavailable implements fallback strategy when stats are unavailable
func (c *Controller) handleStatsUnavailable(priority Priority, queueName string, err error) *ThrottleDecision {
	c.mu.Lock()
	c.lastFallbackTime = time.Now()
	c.mu.Unlock()

	c.metrics.PollingErrors.WithLabelValues("stats_unavailable").Inc()

	if c.config.Recovery.FallbackMode {
		// Use conservative throttling when stats unavailable
		var delay time.Duration
		switch priority {
		case HighPriority:
			delay = 100 * time.Millisecond
		case MediumPriority:
			delay = 500 * time.Millisecond
		case LowPriority:
			delay = 1 * time.Second
		}

		return &ThrottleDecision{
			Priority:  priority,
			QueueName: queueName,
			Delay:     delay,
			ShouldShed: false,
			Reason:    "fallback_conservative",
			Timestamp: time.Now(),
		}
	}

	// No fallback - allow through without throttling
	return &ThrottleDecision{
		Priority:  priority,
		QueueName: queueName,
		Delay:     0,
		ShouldShed: false,
		Reason:    "stats_unavailable_no_fallback",
		Timestamp: time.Now(),
	}
}

// getCachedDecision retrieves a cached throttle decision if valid
func (c *Controller) getCachedDecision(priority Priority, queueName string) *ThrottleDecision {
	key := fmt.Sprintf("%s:%s", queueName, priority.String())

	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	cached, exists := c.throttleCache[key]
	if !exists {
		c.cacheStats.misses++
		return nil
	}

	if time.Now().After(cached.ExpiresAt) {
		c.cacheStats.misses++
		return nil
	}

	c.cacheStats.hits++

	// Update cache hit rate metric
	hitRate := float64(c.cacheStats.hits) / float64(c.cacheStats.hits+c.cacheStats.misses)
	c.metrics.CacheHitRate.WithLabelValues().Set(hitRate)

	return cached.Decision
}

// cacheDecision stores a throttle decision in the cache
func (c *Controller) cacheDecision(priority Priority, queueName string, decision *ThrottleDecision) {
	key := fmt.Sprintf("%s:%s", queueName, priority.String())

	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	c.throttleCache[key] = &CachedDecision{
		Decision:  decision,
		ExpiresAt: time.Now().Add(c.config.Polling.CacheTTL),
	}
}

// getOrCreateCircuitBreaker gets or creates a circuit breaker for the queue
func (c *Controller) getOrCreateCircuitBreaker(queueName string) *CircuitBreaker {
	c.cbMu.RLock()
	cb, exists := c.circuitBreakers[queueName]
	c.cbMu.RUnlock()

	if exists {
		return cb
	}

	c.cbMu.Lock()
	defer c.cbMu.Unlock()

	// Double-check after acquiring write lock
	if cb, exists := c.circuitBreakers[queueName]; exists {
		return cb
	}

	// Create new circuit breaker
	cb = NewCircuitBreaker(c.config.Circuit)
	c.circuitBreakers[queueName] = cb
	return cb
}

// pollingLoop runs background polling for queue statistics
func (c *Controller) pollingLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.Polling.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.pollAllQueues()
		}
	}
}

// pollAllQueues polls statistics for all known queues
func (c *Controller) pollAllQueues() {
	ctx, cancel := context.WithTimeout(c.ctx, c.config.Polling.Timeout)
	defer cancel()

	// Add jitter to prevent thundering herd
	jitter := time.Duration(c.rng.Int63n(int64(c.config.Polling.Jitter)))
	time.Sleep(jitter)

	stats, err := c.statsProvider.GetAllQueueStats(ctx)
	if err != nil {
		c.logger.Warn("Failed to poll queue statistics", zap.Error(err))
		c.metrics.PollingErrors.WithLabelValues("poll_failed").Inc()
		return
	}

	// Update queue backlog metrics
	for queueName, queueStats := range stats {
		c.metrics.QueueBacklogGauge.WithLabelValues(queueName).Set(float64(queueStats.BacklogCount))
	}

	c.logger.Debug("Polled queue statistics", zap.Int("queue_count", len(stats)))
}

// cacheCleanupLoop runs periodic cache cleanup
func (c *Controller) cacheCleanupLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(c.config.Polling.CacheTTL / 2) // Cleanup twice as often as TTL
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.cleanupExpiredCache()
		}
	}
}

// cleanupExpiredCache removes expired entries from the throttle cache
func (c *Controller) cleanupExpiredCache() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	now := time.Now()
	for key, cached := range c.throttleCache {
		if now.After(cached.ExpiresAt) {
			delete(c.throttleCache, key)
		}
	}
}