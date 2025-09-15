// Copyright 2025 James Ross
package patternedloadgenerator

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// LoadGenerator generates load with various patterns
type LoadGenerator struct {
	config        *GeneratorConfig
	redis         *redis.Client
	logger        *zap.Logger
	jobGenerator  JobGenerator

	// State
	status        *GeneratorStatus
	currentProfile *LoadProfile
	metrics       []MetricsSnapshot
	chartData     *ChartData

	// Control
	ctx           context.Context
	cancel        context.CancelFunc
	pauseCh       chan bool
	controlCh     chan ControlMessage
	eventCh       chan GeneratorEvent

	// Synchronization
	mu            sync.RWMutex
	wg            sync.WaitGroup
	running       atomic.Bool
	paused        atomic.Bool

	// Rate limiting
	rateLimiter   *RateLimiter
	jobsGenerated atomic.Int64
	errors        atomic.Int64
}

// NewLoadGenerator creates a new load generator
func NewLoadGenerator(config *GeneratorConfig, redis *redis.Client, logger *zap.Logger) *LoadGenerator {
	if config == nil {
		config = &GeneratorConfig{
			DefaultGuardrails: Guardrails{
				MaxRate:         1000,
				MaxTotal:        1000000,
				MaxDuration:     1 * time.Hour,
				MaxQueueDepth:   10000,
				RateLimitWindow: 1 * time.Second,
			},
			MetricsInterval:  1 * time.Second,
			ProfilesPath:     "./profiles",
			EnableCharts:     true,
			ChartUpdateRate:  500 * time.Millisecond,
			MaxHistoryPoints: 1000,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	lg := &LoadGenerator{
		config:      config,
		redis:       redis,
		logger:      logger,
		status:      &GeneratorStatus{},
		metrics:     make([]MetricsSnapshot, 0),
		chartData:   &ChartData{},
		ctx:         ctx,
		cancel:      cancel,
		pauseCh:     make(chan bool, 1),
		controlCh:   make(chan ControlMessage, 10),
		eventCh:     make(chan GeneratorEvent, 100),
		rateLimiter: NewRateLimiter(config.DefaultGuardrails.MaxRate, config.DefaultGuardrails.RateLimitWindow),
	}

	// Create default job generator
	lg.jobGenerator = &SimpleJobGenerator{
		Template: map[string]interface{}{
			"type": "load_test",
		},
	}

	// Start control loop
	lg.wg.Add(1)
	go lg.controlLoop()

	// Start metrics collector
	lg.wg.Add(1)
	go lg.metricsLoop()

	return lg
}

// Start starts load generation with a profile
func (lg *LoadGenerator) Start(profileID string) error {
	if lg.running.Load() {
		return fmt.Errorf("generator already running")
	}

	// Load profile
	profile, err := lg.LoadProfile(profileID)
	if err != nil {
		return fmt.Errorf("failed to load profile: %w", err)
	}

	lg.mu.Lock()
	lg.currentProfile = profile
	lg.status = &GeneratorStatus{
		Running:   true,
		StartedAt: time.Now(),
	}
	lg.mu.Unlock()

	lg.running.Store(true)
	lg.sendEvent(EventStarted, "Load generation started", map[string]interface{}{
		"profile": profileID,
	})

	// Start pattern execution
	lg.wg.Add(1)
	go lg.executeProfile(profile)

	return nil
}

// StartPattern starts load generation with a specific pattern
func (lg *LoadGenerator) StartPattern(pattern *LoadPattern, guardrails *Guardrails) error {
	if lg.running.Load() {
		return fmt.Errorf("generator already running")
	}

	if guardrails == nil {
		guardrails = &lg.config.DefaultGuardrails
	}

	profile := &LoadProfile{
		ID:          uuid.New().String(),
		Name:        fmt.Sprintf("Pattern: %s", pattern.Type),
		Description: "Ad-hoc pattern execution",
		Patterns:    []LoadPattern{*pattern},
		Guardrails:  *guardrails,
		QueueName:   "default",
		CreatedAt:   time.Now(),
	}

	lg.mu.Lock()
	lg.currentProfile = profile
	lg.status = &GeneratorStatus{
		Running:   true,
		Pattern:   pattern.Type,
		StartedAt: time.Now(),
	}
	lg.mu.Unlock()

	lg.running.Store(true)
	lg.sendEvent(EventStarted, "Pattern generation started", map[string]interface{}{
		"pattern": pattern.Type,
	})

	// Start pattern execution
	lg.wg.Add(1)
	go lg.executeProfile(profile)

	return nil
}

// Stop stops load generation
func (lg *LoadGenerator) Stop() error {
	if !lg.running.Load() {
		return fmt.Errorf("generator not running")
	}

	lg.running.Store(false)
	lg.cancel()

	lg.sendEvent(EventStopped, "Load generation stopped", map[string]interface{}{
		"jobs_generated": lg.jobsGenerated.Load(),
	})

	return nil
}

// Pause pauses load generation
func (lg *LoadGenerator) Pause() error {
	if !lg.running.Load() {
		return fmt.Errorf("generator not running")
	}

	if lg.paused.Load() {
		return fmt.Errorf("generator already paused")
	}

	lg.paused.Store(true)
	lg.pauseCh <- true

	lg.sendEvent(EventPaused, "Load generation paused", nil)
	return nil
}

// Resume resumes load generation
func (lg *LoadGenerator) Resume() error {
	if !lg.running.Load() {
		return fmt.Errorf("generator not running")
	}

	if !lg.paused.Load() {
		return fmt.Errorf("generator not paused")
	}

	lg.paused.Store(false)
	lg.pauseCh <- false

	lg.sendEvent(EventResumed, "Load generation resumed", nil)
	return nil
}

// GetStatus returns current generator status
func (lg *LoadGenerator) GetStatus() *GeneratorStatus {
	lg.mu.RLock()
	defer lg.mu.RUnlock()

	status := *lg.status
	status.JobsGenerated = lg.jobsGenerated.Load()
	status.Errors = lg.errors.Load()
	status.Running = lg.running.Load()

	if lg.running.Load() {
		status.Duration = time.Since(status.StartedAt)
	}

	return &status
}

// GetChartData returns chart visualization data
func (lg *LoadGenerator) GetChartData() *ChartData {
	lg.mu.RLock()
	defer lg.mu.RUnlock()

	return lg.chartData
}

// GetMetrics returns recent metrics
func (lg *LoadGenerator) GetMetrics(duration time.Duration) []MetricsSnapshot {
	lg.mu.RLock()
	defer lg.mu.RUnlock()

	cutoff := time.Now().Add(-duration)
	var result []MetricsSnapshot

	for _, m := range lg.metrics {
		if m.Timestamp.After(cutoff) {
			result = append(result, m)
		}
	}

	return result
}

// SaveProfile saves a load profile
func (lg *LoadGenerator) SaveProfile(profile *LoadProfile) error {
	if profile.ID == "" {
		profile.ID = uuid.New().String()
	}

	profile.UpdatedAt = time.Now()
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = time.Now()
	}

	// Create profiles directory
	if err := os.MkdirAll(lg.config.ProfilesPath, 0755); err != nil {
		return fmt.Errorf("failed to create profiles directory: %w", err)
	}

	// Save profile
	filename := filepath.Join(lg.config.ProfilesPath, fmt.Sprintf("%s.json", profile.ID))
	data, err := json.MarshalIndent(profile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal profile: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write profile: %w", err)
	}

	lg.logger.Info("Profile saved", zap.String("id", profile.ID), zap.String("name", profile.Name))
	return nil
}

// LoadProfile loads a saved profile
func (lg *LoadGenerator) LoadProfile(profileID string) (*LoadProfile, error) {
	filename := filepath.Join(lg.config.ProfilesPath, fmt.Sprintf("%s.json", profileID))

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile: %w", err)
	}

	var profile LoadProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile: %w", err)
	}

	return &profile, nil
}

// ListProfiles lists available profiles
func (lg *LoadGenerator) ListProfiles() ([]*LoadProfile, error) {
	pattern := filepath.Join(lg.config.ProfilesPath, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var profiles []*LoadProfile
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var profile LoadProfile
		if err := json.Unmarshal(data, &profile); err != nil {
			continue
		}

		profiles = append(profiles, &profile)
	}

	return profiles, nil
}

// DeleteProfile deletes a saved profile
func (lg *LoadGenerator) DeleteProfile(profileID string) error {
	filename := filepath.Join(lg.config.ProfilesPath, fmt.Sprintf("%s.json", profileID))
	return os.Remove(filename)
}

// Pattern execution

func (lg *LoadGenerator) executeProfile(profile *LoadProfile) {
	defer lg.wg.Done()

	startTime := time.Now()
	ticker := time.NewTicker(10 * time.Millisecond) // High frequency for accurate pattern generation
	defer ticker.Stop()

	for _, pattern := range profile.Patterns {
		if !lg.running.Load() {
			break
		}

		lg.mu.Lock()
		lg.status.Pattern = pattern.Type
		lg.mu.Unlock()

		lg.sendEvent(EventPatternChanged, fmt.Sprintf("Starting pattern: %s", pattern.Type), map[string]interface{}{
			"pattern": pattern,
		})

		endTime := time.Now().Add(pattern.Duration)

		for time.Now().Before(endTime) && lg.running.Load() {
			select {
			case <-lg.ctx.Done():
				return

			case isPaused := <-lg.pauseCh:
				if isPaused {
					// Wait for resume
					<-lg.pauseCh
				}

			case <-ticker.C:
				// Check guardrails
				if err := lg.checkGuardrails(&profile.Guardrails); err != nil {
					lg.logger.Warn("Guardrail hit", zap.Error(err))
					lg.sendEvent(EventGuardrailHit, err.Error(), nil)
					lg.Stop()
					return
				}

				// Calculate target rate for this moment
				elapsed := time.Since(startTime)
				targetRate := lg.calculateRate(pattern, elapsed)

				// Update status
				lg.mu.Lock()
				lg.status.TargetRate = targetRate
				lg.mu.Unlock()

				// Generate jobs at target rate
				lg.generateJobs(targetRate, profile.QueueName)
			}
		}
	}

	lg.Stop()
}

func (lg *LoadGenerator) calculateRate(pattern LoadPattern, elapsed time.Duration) float64 {
	switch pattern.Type {
	case PatternSine:
		return lg.calculateSineRate(pattern, elapsed)
	case PatternBurst:
		return lg.calculateBurstRate(pattern, elapsed)
	case PatternRamp:
		return lg.calculateRampRate(pattern, elapsed)
	case PatternStep:
		return lg.calculateStepRate(pattern, elapsed)
	case PatternConstant:
		return lg.calculateConstantRate(pattern)
	case PatternCustom:
		return lg.calculateCustomRate(pattern, elapsed)
	default:
		return 0
	}
}

func (lg *LoadGenerator) calculateSineRate(pattern LoadPattern, elapsed time.Duration) float64 {
	var params SineParameters
	if data, _ := json.Marshal(pattern.Parameters); data != nil {
		json.Unmarshal(data, &params)
	}

	// Default values
	if params.Period == 0 {
		params.Period = 60 * time.Second
	}
	if params.Amplitude == 0 {
		params.Amplitude = 50
	}
	if params.Baseline == 0 {
		params.Baseline = 100
	}

	// Calculate sine wave
	t := elapsed.Seconds()
	period := params.Period.Seconds()
	rate := params.Baseline + params.Amplitude*math.Sin(2*math.Pi*t/period+params.Phase)

	return math.Max(0, rate)
}

func (lg *LoadGenerator) calculateBurstRate(pattern LoadPattern, elapsed time.Duration) float64 {
	var params BurstParameters
	if data, _ := json.Marshal(pattern.Parameters); data != nil {
		json.Unmarshal(data, &params)
	}

	// Default values
	if params.BurstDuration == 0 {
		params.BurstDuration = 5 * time.Second
	}
	if params.IdleDuration == 0 {
		params.IdleDuration = 10 * time.Second
	}
	if params.BurstRate == 0 {
		params.BurstRate = 200
	}

	// Calculate cycle position
	cycleTime := params.BurstDuration + params.IdleDuration
	cyclePosition := time.Duration(elapsed.Nanoseconds() % cycleTime.Nanoseconds())

	if cyclePosition < params.BurstDuration {
		return params.BurstRate
	}
	return 0
}

func (lg *LoadGenerator) calculateRampRate(pattern LoadPattern, elapsed time.Duration) float64 {
	var params RampParameters
	if data, _ := json.Marshal(pattern.Parameters); data != nil {
		json.Unmarshal(data, &params)
	}

	// Default values
	if params.RampDuration == 0 {
		params.RampDuration = 30 * time.Second
	}

	if elapsed < params.RampDuration {
		// Ramping up
		progress := elapsed.Seconds() / params.RampDuration.Seconds()
		return params.StartRate + (params.EndRate-params.StartRate)*progress
	} else if elapsed < params.RampDuration+params.HoldDuration {
		// Holding
		return params.EndRate
	} else if params.RampDown && elapsed < 2*params.RampDuration+params.HoldDuration {
		// Ramping down
		downElapsed := elapsed - params.RampDuration - params.HoldDuration
		progress := downElapsed.Seconds() / params.RampDuration.Seconds()
		return params.EndRate - (params.EndRate-params.StartRate)*progress
	}

	return params.StartRate
}

func (lg *LoadGenerator) calculateStepRate(pattern LoadPattern, elapsed time.Duration) float64 {
	var params StepParameters
	if data, _ := json.Marshal(pattern.Parameters); data != nil {
		json.Unmarshal(data, &params)
	}

	if len(params.Steps) == 0 {
		return 0
	}

	// Calculate which step we're in
	totalDuration := time.Duration(0)
	for _, step := range params.Steps {
		duration := params.StepDuration
		if step.Duration > 0 {
			duration = step.Duration
		}

		if elapsed < totalDuration+duration {
			return step.Rate
		}
		totalDuration += duration
	}

	// If repeat is enabled, loop back
	if params.Repeat && totalDuration > 0 {
		modElapsed := time.Duration(elapsed.Nanoseconds() % totalDuration.Nanoseconds())
		return lg.calculateStepRate(pattern, modElapsed)
	}

	return 0
}

func (lg *LoadGenerator) calculateConstantRate(pattern LoadPattern) float64 {
	if rate, ok := pattern.Parameters["rate"].(float64); ok {
		return rate
	}
	return 100 // Default constant rate
}

func (lg *LoadGenerator) calculateCustomRate(pattern LoadPattern, elapsed time.Duration) float64 {
	var params CustomParameters
	if data, _ := json.Marshal(pattern.Parameters); data != nil {
		json.Unmarshal(data, &params)
	}

	if len(params.Points) == 0 {
		return 0
	}

	// Find the appropriate rate for current time
	for i := len(params.Points) - 1; i >= 0; i-- {
		if elapsed >= params.Points[i].Time {
			if i == len(params.Points)-1 {
				return params.Points[i].Rate
			}

			// Interpolate between points
			nextPoint := params.Points[i+1]
			progress := (elapsed - params.Points[i].Time).Seconds() / (nextPoint.Time - params.Points[i].Time).Seconds()
			return params.Points[i].Rate + (nextPoint.Rate-params.Points[i].Rate)*progress
		}
	}

	return params.Points[0].Rate
}

func (lg *LoadGenerator) generateJobs(targetRate float64, queueName string) {
	// Calculate jobs to generate in this tick
	tickDuration := 10 * time.Millisecond
	jobsPerTick := targetRate * tickDuration.Seconds()

	// Use fractional accumulation for accurate generation
	lg.mu.Lock()
	fractionalJobs := jobsPerTick
	jobsToGenerate := int(fractionalJobs)
	lg.mu.Unlock()

	ctx := context.Background()

	for i := 0; i < jobsToGenerate; i++ {
		// Rate limiting
		if !lg.rateLimiter.Allow() {
			break
		}

		// Generate job
		job, err := lg.jobGenerator.GenerateJob()
		if err != nil {
			lg.errors.Add(1)
			lg.logger.Warn("Failed to generate job", zap.Error(err))
			continue
		}

		// Enqueue job
		jobData, _ := json.Marshal(job)
		if err := lg.redis.RPush(ctx, fmt.Sprintf("queue:%s", queueName), string(jobData)).Err(); err != nil {
			lg.errors.Add(1)
			lg.logger.Warn("Failed to enqueue job", zap.Error(err))
			continue
		}

		lg.jobsGenerated.Add(1)
	}

	// Update actual rate
	lg.mu.Lock()
	lg.status.CurrentRate = float64(jobsToGenerate) / tickDuration.Seconds()
	lg.mu.Unlock()
}

func (lg *LoadGenerator) checkGuardrails(guardrails *Guardrails) error {
	// Check max total
	if guardrails.MaxTotal > 0 && lg.jobsGenerated.Load() >= guardrails.MaxTotal {
		return fmt.Errorf("max total jobs reached: %d", guardrails.MaxTotal)
	}

	// Check max duration
	if guardrails.MaxDuration > 0 {
		lg.mu.RLock()
		elapsed := time.Since(lg.status.StartedAt)
		lg.mu.RUnlock()

		if elapsed >= guardrails.MaxDuration {
			return fmt.Errorf("max duration reached: %v", guardrails.MaxDuration)
		}
	}

	// Check queue depth
	if guardrails.MaxQueueDepth > 0 && lg.currentProfile != nil {
		ctx := context.Background()
		length, err := lg.redis.LLen(ctx, fmt.Sprintf("queue:%s", lg.currentProfile.QueueName)).Result()
		if err == nil && length > guardrails.MaxQueueDepth {
			return fmt.Errorf("max queue depth exceeded: %d > %d", length, guardrails.MaxQueueDepth)
		}
	}

	// Check emergency stop file
	if guardrails.EmergencyStopFile != "" {
		if _, err := os.Stat(guardrails.EmergencyStopFile); err == nil {
			return fmt.Errorf("emergency stop file detected: %s", guardrails.EmergencyStopFile)
		}
	}

	return nil
}

func (lg *LoadGenerator) controlLoop() {
	defer lg.wg.Done()

	for {
		select {
		case <-lg.ctx.Done():
			return

		case msg := <-lg.controlCh:
			lg.handleControlMessage(msg)
		}
	}
}

func (lg *LoadGenerator) handleControlMessage(msg ControlMessage) {
	switch msg.Command {
	case CommandStart:
		if msg.ProfileID != "" {
			lg.Start(msg.ProfileID)
		} else if msg.Pattern != nil {
			lg.StartPattern(msg.Pattern, nil)
		}

	case CommandStop:
		lg.Stop()

	case CommandPause:
		lg.Pause()

	case CommandResume:
		lg.Resume()

	case CommandReset:
		lg.jobsGenerated.Store(0)
		lg.errors.Store(0)
		lg.mu.Lock()
		lg.metrics = make([]MetricsSnapshot, 0)
		lg.chartData = &ChartData{}
		lg.mu.Unlock()
	}
}

func (lg *LoadGenerator) metricsLoop() {
	defer lg.wg.Done()

	ticker := time.NewTicker(lg.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-lg.ctx.Done():
			return

		case <-ticker.C:
			if lg.running.Load() {
				lg.collectMetrics()
			}
		}
	}
}

func (lg *LoadGenerator) collectMetrics() {
	lg.mu.Lock()
	defer lg.mu.Unlock()

	// Get queue depth
	var queueDepth int64
	if lg.currentProfile != nil {
		ctx := context.Background()
		depth, _ := lg.redis.LLen(ctx, fmt.Sprintf("queue:%s", lg.currentProfile.QueueName)).Result()
		queueDepth = depth
	}

	// Create snapshot
	snapshot := MetricsSnapshot{
		Timestamp:     time.Now(),
		TargetRate:    lg.status.TargetRate,
		ActualRate:    lg.status.CurrentRate,
		JobsGenerated: lg.jobsGenerated.Load(),
		QueueDepth:    queueDepth,
		Errors:        lg.errors.Load(),
	}

	// Add to metrics
	lg.metrics = append(lg.metrics, snapshot)

	// Trim old metrics
	if len(lg.metrics) > lg.config.MaxHistoryPoints {
		lg.metrics = lg.metrics[len(lg.metrics)-lg.config.MaxHistoryPoints:]
	}

	// Update chart data
	if lg.config.EnableCharts {
		lg.updateChartData()
	}

	// Send metrics event
	lg.sendEvent(EventMetricsUpdate, "Metrics updated", map[string]interface{}{
		"snapshot": snapshot,
	})
}

func (lg *LoadGenerator) updateChartData() {
	// Build chart data from metrics
	lg.chartData = &ChartData{
		TimePoints:  make([]time.Time, 0, len(lg.metrics)),
		TargetRates: make([]float64, 0, len(lg.metrics)),
		ActualRates: make([]float64, 0, len(lg.metrics)),
		QueueDepths: make([]int64, 0, len(lg.metrics)),
		ErrorCounts: make([]int64, 0, len(lg.metrics)),
	}

	for _, m := range lg.metrics {
		lg.chartData.TimePoints = append(lg.chartData.TimePoints, m.Timestamp)
		lg.chartData.TargetRates = append(lg.chartData.TargetRates, m.TargetRate)
		lg.chartData.ActualRates = append(lg.chartData.ActualRates, m.ActualRate)
		lg.chartData.QueueDepths = append(lg.chartData.QueueDepths, m.QueueDepth)
		lg.chartData.ErrorCounts = append(lg.chartData.ErrorCounts, m.Errors)
	}
}

func (lg *LoadGenerator) sendEvent(eventType string, message string, data map[string]interface{}) {
	event := GeneratorEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Message:   message,
		Data:      data,
	}

	select {
	case lg.eventCh <- event:
	default:
		// Event channel full, drop event
	}
}

// GetEventChannel returns the event channel for monitoring
func (lg *LoadGenerator) GetEventChannel() <-chan GeneratorEvent {
	return lg.eventCh
}

// Shutdown gracefully shuts down the generator
func (lg *LoadGenerator) Shutdown() {
	lg.Stop()
	lg.cancel()
	lg.wg.Wait()
	close(lg.eventCh)
}

// GenerateJob generates a test job
func (sg *SimpleJobGenerator) GenerateJob() (interface{}, error) {
	sg.Counter++

	job := make(map[string]interface{})
	for k, v := range sg.Template {
		job[k] = v
	}

	job["id"] = fmt.Sprintf("job-%d", sg.Counter)
	job["timestamp"] = time.Now().Unix()
	job["sequence"] = sg.Counter

	return job, nil
}

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	rate       float64
	capacity   float64
	tokens     float64
	lastUpdate time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate float64, window time.Duration) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		capacity:   rate * window.Seconds(),
		tokens:     rate * window.Seconds(),
		lastUpdate: time.Now(),
	}
}

// Allow checks if an operation is allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(rl.lastUpdate).Seconds()
	rl.tokens = math.Min(rl.capacity, rl.tokens+rl.rate*elapsed)
	rl.lastUpdate = now

	// Check if token available
	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}

	return false
}

// SetRate updates the rate limit
func (rl *RateLimiter) SetRate(rate float64) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.rate = rate
	rl.capacity = rate // 1 second capacity
}

// PreviewPattern generates preview points for a pattern
func (lg *LoadGenerator) PreviewPattern(pattern LoadPattern, duration time.Duration, resolution time.Duration) ([]DataPoint, error) {
	if duration == 0 {
		duration = pattern.Duration
	}
	if resolution == 0 {
		resolution = 1 * time.Second
	}

	var points []DataPoint
	for elapsed := time.Duration(0); elapsed <= duration; elapsed += resolution {
		rate := lg.calculateRate(pattern, elapsed)
		points = append(points, DataPoint{
			Time: elapsed,
			Rate: rate,
		})
	}

	return points, nil
}