// Copyright 2025 James Ross
package forecasting

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Forecaster interface for all forecasting models
type Forecaster interface {
	Update(observation float64) error
	Forecast(horizonMinutes int) (*ForecastResult, error)
	GetAccuracy() *AccuracyMetrics
	GetConfiguration() ModelConfig
}

// ForecastingEngine orchestrates forecasting and recommendations
type ForecastingEngine struct {
	config    *ForecastConfig
	storage   *MetricsStorage
	models    map[string]Forecaster
	recommender *RecommendationEngine
	logger    *zap.Logger

	// Current state
	latestMetrics   *QueueMetrics
	latestForecasts map[MetricType]*ForecastResult
	latestRecs      []Recommendation

	// Lifecycle
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	mu       sync.RWMutex
}

// NewForecastingEngine creates a new forecasting engine
func NewForecastingEngine(config *ForecastConfig, logger *zap.Logger) *ForecastingEngine {
	if config == nil {
		config = &ForecastConfig{
			EWMAConfig:        &EWMAConfig{Alpha: 0.3},
			HoltWintersConfig: &HoltWintersConfig{Alpha: 0.3, Beta: 0.1, Gamma: 0.1, SeasonLength: 24},
			StorageConfig:     &StorageConfig{RetentionDuration: 7 * 24 * time.Hour},
			EngineConfig:      &EngineConfig{Enabled: true, UpdateInterval: 1 * time.Minute},
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	fe := &ForecastingEngine{
		config:          config,
		storage:         NewMetricsStorage(config.StorageConfig, logger),
		models:          make(map[string]Forecaster),
		recommender:     NewRecommendationEngine(config.EngineConfig, logger),
		logger:          logger,
		latestForecasts: make(map[MetricType]*ForecastResult),
		ctx:             ctx,
		cancel:          cancel,
	}

	// Initialize models
	fe.initializeModels()

	// Start background processing
	fe.wg.Add(1)
	go fe.updateLoop()

	return fe
}

// UpdateMetrics updates the engine with new metrics
func (fe *ForecastingEngine) UpdateMetrics(metrics *QueueMetrics) error {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	fe.latestMetrics = metrics

	// Store metrics
	fe.storage.Store(MetricBacklog, metrics.QueueName, float64(metrics.Backlog))
	fe.storage.Store(MetricThroughput, metrics.QueueName, metrics.Throughput)
	fe.storage.Store(MetricErrorRate, metrics.QueueName, metrics.ErrorRate)
	fe.storage.Store(MetricLatency, metrics.QueueName, metrics.LatencyP99)
	fe.storage.Store(MetricWorkers, metrics.QueueName, float64(metrics.ActiveWorkers))

	// Update models
	for name, model := range fe.models {
		var value float64
		switch name {
		case "ewma-backlog":
			value = float64(metrics.Backlog)
		case "ewma-throughput":
			value = metrics.Throughput
		case "ewma-error-rate":
			value = metrics.ErrorRate
		case "hw-backlog":
			value = float64(metrics.Backlog)
		case "hw-throughput":
			value = metrics.Throughput
		}

		if err := model.Update(value); err != nil {
			fe.logger.Warn("Failed to update model",
				zap.String("model", name),
				zap.Error(err))
		}
	}

	return nil
}

// GetForecasts returns current forecasts
func (fe *ForecastingEngine) GetForecasts(horizonMinutes int) (map[MetricType]*ForecastResult, error) {
	fe.mu.Lock()
	defer fe.mu.Unlock()

	forecasts := make(map[MetricType]*ForecastResult)

	// Generate forecasts from each model
	for name, model := range fe.models {
		forecast, err := model.Forecast(horizonMinutes)
		if err != nil {
			fe.logger.Debug("Model forecast failed",
				zap.String("model", name),
				zap.Error(err))
			continue
		}

		// Determine metric type from model name
		var metricType MetricType
		switch {
		case contains(name, "backlog"):
			metricType = MetricBacklog
		case contains(name, "throughput"):
			metricType = MetricThroughput
		case contains(name, "error"):
			metricType = MetricErrorRate
		}

		if metricType != "" {
			forecast.MetricType = metricType

			// Use best forecast (highest confidence)
			if existing, ok := forecasts[metricType]; !ok || forecast.Confidence > existing.Confidence {
				forecasts[metricType] = forecast
			}
		}
	}

	fe.latestForecasts = forecasts
	return forecasts, nil
}

// GetRecommendations returns current recommendations
func (fe *ForecastingEngine) GetRecommendations() []Recommendation {
	fe.mu.RLock()
	defer fe.mu.RUnlock()

	return fe.latestRecs
}

// GetHistoricalData returns historical metric data
func (fe *ForecastingEngine) GetHistoricalData(metricType MetricType, queueName string, duration time.Duration) []DataPoint {
	return fe.storage.GetTimeSeries(metricType, queueName, duration)
}

// GetModelAccuracy returns accuracy metrics for all models
func (fe *ForecastingEngine) GetModelAccuracy() map[string]*AccuracyMetrics {
	fe.mu.RLock()
	defer fe.mu.RUnlock()

	accuracy := make(map[string]*AccuracyMetrics)
	for name, model := range fe.models {
		if acc := model.GetAccuracy(); acc != nil {
			accuracy[name] = acc
		}
	}

	return accuracy
}

// Stop stops the forecasting engine
func (fe *ForecastingEngine) Stop() {
	fe.cancel()
	fe.wg.Wait()
	fe.storage.Stop()
}

// Helper methods

func (fe *ForecastingEngine) initializeModels() {
	// EWMA models
	fe.models["ewma-backlog"] = NewEWMAForecaster(fe.config.EWMAConfig)
	fe.models["ewma-throughput"] = NewEWMAForecaster(fe.config.EWMAConfig)
	fe.models["ewma-error-rate"] = NewEWMAForecaster(&EWMAConfig{
		Alpha:              0.1, // Slower adaptation for error rates
		AutoAdjust:         true,
		MinObservations:    5,
		ConfidenceInterval: 0.95,
	})

	// Holt-Winters models
	fe.models["hw-backlog"] = NewHoltWintersForecaster(fe.config.HoltWintersConfig)
	fe.models["hw-throughput"] = NewHoltWintersForecaster(fe.config.HoltWintersConfig)
}

func (fe *ForecastingEngine) updateLoop() {
	defer fe.wg.Done()

	ticker := time.NewTicker(fe.config.EngineConfig.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-fe.ctx.Done():
			return
		case <-ticker.C:
			fe.update()
		}
	}
}

func (fe *ForecastingEngine) update() {
	// Generate new forecasts
	forecasts, err := fe.GetForecasts(120) // 2-hour horizon
	if err != nil {
		fe.logger.Warn("Failed to generate forecasts", zap.Error(err))
		return
	}

	fe.mu.Lock()
	metrics := fe.latestMetrics
	fe.mu.Unlock()

	if metrics == nil {
		return
	}

	// Generate recommendations
	recs := fe.recommender.GenerateRecommendations(forecasts, metrics)

	fe.mu.Lock()
	fe.latestRecs = recs
	fe.mu.Unlock()

	// Evaluate model accuracy
	fe.evaluateAccuracy()
}

func (fe *ForecastingEngine) evaluateAccuracy() {
	// Compare previous forecasts with actual values
	for metricType, forecast := range fe.latestForecasts {
		if forecast == nil || len(forecast.Points) == 0 {
			continue
		}

		// Get actual value
		var actual float64
		switch metricType {
		case MetricBacklog:
			if fe.latestMetrics != nil {
				actual = float64(fe.latestMetrics.Backlog)
			}
		case MetricThroughput:
			if fe.latestMetrics != nil {
				actual = fe.latestMetrics.Throughput
			}
		case MetricErrorRate:
			if fe.latestMetrics != nil {
				actual = fe.latestMetrics.ErrorRate
			}
		}

		// Record prediction accuracy
		if model, ok := fe.models[fmt.Sprintf("ewma-%s", metricType)]; ok {
			if ewma, ok := model.(*EWMAForecaster); ok {
				ewma.RecordPrediction(forecast.Points[0], actual, time.Minute)
			}
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[0:len(substr)] == substr || len(s) > len(substr) && s[len(s)-len(substr):] == substr || len(s) > len(substr) && containsMiddle(s, substr)
}

func containsMiddle(s, substr string) bool {
	for i := 1; i < len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// GetStatus returns current engine status
func (fe *ForecastingEngine) GetStatus() map[string]interface{} {
	fe.mu.RLock()
	defer fe.mu.RUnlock()

	status := map[string]interface{}{
		"enabled":          fe.config.EngineConfig.Enabled,
		"models_count":     len(fe.models),
		"forecasts_count":  len(fe.latestForecasts),
		"recommendations":  len(fe.latestRecs),
		"update_interval":  fe.config.EngineConfig.UpdateInterval.String(),
	}

	if fe.latestMetrics != nil {
		status["latest_metrics"] = fe.latestMetrics
	}

	return status
}
