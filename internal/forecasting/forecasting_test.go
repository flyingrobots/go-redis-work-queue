//go:build forecasting_tests
// +build forecasting_tests

// Copyright 2025 James Ross
package forecasting

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestEWMAForecaster(t *testing.T) {
	t.Run("basic forecasting", func(t *testing.T) {
		config := &EWMAConfig{
			Alpha:              0.3,
			AutoAdjust:         false,
			MinObservations:    3,
			ConfidenceInterval: 0.95,
		}

		ewma := NewEWMAForecaster(config)

		// Add observations
		values := []float64{100, 110, 105, 115, 108}
		for _, v := range values {
			err := ewma.Update(v)
			require.NoError(t, err)
		}

		// Generate forecast
		forecast, err := ewma.Forecast(10)
		require.NoError(t, err)
		assert.NotNil(t, forecast)
		assert.Len(t, forecast.Points, 10)
		assert.Len(t, forecast.UpperBounds, 10)
		assert.Len(t, forecast.LowerBounds, 10)

		// Check forecast is reasonable
		assert.InDelta(t, 110, forecast.Points[0], 10)

		// Confidence bounds should widen over time
		assert.Less(t,
			forecast.UpperBounds[0]-forecast.LowerBounds[0],
			forecast.UpperBounds[9]-forecast.LowerBounds[9])
	})

	t.Run("insufficient observations", func(t *testing.T) {
		config := &EWMAConfig{
			Alpha:           0.3,
			MinObservations: 5,
		}

		ewma := NewEWMAForecaster(config)

		// Add only 2 observations
		ewma.Update(100)
		ewma.Update(110)

		// Should fail with insufficient observations
		_, err := ewma.Forecast(10)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insufficient observations")
	})

	t.Run("auto adjust alpha", func(t *testing.T) {
		config := &EWMAConfig{
			Alpha:           0.3,
			AutoAdjust:      true,
			MinObservations: 3,
		}

		ewma := NewEWMAForecaster(config)

		// Add observations with large jumps
		for i := 0; i < 25; i++ {
			if i%5 == 0 {
				ewma.Update(200) // Large jump
			} else {
				ewma.Update(100)
			}
		}

		// Alpha should have adjusted
		config2 := ewma.GetConfiguration()
		alpha := config2.Parameters["alpha"].(float64)
		assert.NotEqual(t, 0.3, alpha) // Should have changed from initial
	})
}

func TestHoltWintersForecaster(t *testing.T) {
	t.Run("seasonal pattern detection", func(t *testing.T) {
		config := &HoltWintersConfig{
			Alpha:            0.3,
			Beta:             0.1,
			Gamma:            0.1,
			SeasonLength:     24,
			SeasonalMethod:   "additive",
			AutoDetectSeason: false,
		}

		hw := NewHoltWintersForecaster(config)

		// Generate seasonal data (daily pattern)
		for day := 0; day < 3; day++ {
			for hour := 0; hour < 24; hour++ {
				// Peak at noon, low at midnight
				value := 100 + 50*math.Sin(2*math.Pi*float64(hour)/24)
				err := hw.Update(value)
				require.NoError(t, err)
			}
		}

		// Generate forecast
		forecast, err := hw.Forecast(24)
		require.NoError(t, err)
		assert.NotNil(t, forecast)

		// Should capture seasonal pattern
		// Peak should be around index 12 (noon)
		maxVal := forecast.Points[0]
		maxIdx := 0
		for i, v := range forecast.Points {
			if v > maxVal {
				maxVal = v
				maxIdx = i
			}
		}

		assert.InDelta(t, 12, maxIdx, 3) // Peak near noon
	})

	t.Run("multiplicative seasonality", func(t *testing.T) {
		config := &HoltWintersConfig{
			Alpha:          0.3,
			Beta:           0.1,
			Gamma:          0.1,
			SeasonLength:   12,
			SeasonalMethod: "multiplicative",
		}

		hw := NewHoltWintersForecaster(config)

		// Generate data with multiplicative seasonality
		for i := 0; i < 36; i++ {
			base := 100 + float64(i) // Trend
			seasonal := 1 + 0.5*math.Sin(2*math.Pi*float64(i)/12)
			value := base * seasonal
			hw.Update(value)
		}

		forecast, err := hw.Forecast(12)
		require.NoError(t, err)
		assert.NotNil(t, forecast)

		// Should show increasing trend
		assert.Greater(t, forecast.Points[11], forecast.Points[0])
	})
}

func TestRecommendationEngine(t *testing.T) {
	logger := zap.NewNop()
	config := &EngineConfig{
		Enabled:        true,
		UpdateInterval: 1 * time.Minute,
		Thresholds: map[string]float64{
			"critical_backlog":    1000,
			"high_backlog":        500,
			"critical_error_rate": 0.1,
			"high_error_rate":     0.05,
		},
		ScalingPolicy: ScalingPolicy{
			MinWorkers:     1,
			MaxWorkers:     10,
			CooldownPeriod: 5 * time.Minute,
		},
	}

	re := NewRecommendationEngine(config, logger)

	t.Run("critical backlog recommendation", func(t *testing.T) {
		forecasts := map[MetricType]*ForecastResult{
			MetricBacklog: {
				Points:      []float64{1500, 1600, 1700, 1800},
				Confidence:  0.85,
				MetricType:  MetricBacklog,
				GeneratedAt: time.Now(),
			},
		}

		metrics := &QueueMetrics{
			Backlog:       800,
			Throughput:    10,
			ActiveWorkers: 2,
		}

		recs := re.GenerateRecommendations(forecasts, metrics)
		assert.NotEmpty(t, recs)

		// Should have critical scaling recommendation
		var foundCritical bool
		for _, rec := range recs {
			if rec.Priority == PriorityCritical && rec.Category == CategoryCapacityScaling {
				foundCritical = true
				assert.Contains(t, rec.Title, "CRITICAL")
				assert.Contains(t, rec.Action, "kubectl scale")
			}
		}
		assert.True(t, foundCritical)
	})

	t.Run("SLO budget warning", func(t *testing.T) {
		forecasts := map[MetricType]*ForecastResult{
			MetricErrorRate: {
				Points:     []float64{0.08, 0.09, 0.10, 0.11},
				Confidence: 0.90,
				MetricType: MetricErrorRate,
			},
		}

		metrics := &QueueMetrics{
			ErrorRate: 0.07,
		}

		// Update SLO tracker
		for i := 0; i < 100; i++ {
			re.sloTracker.Update(0.08) // High error rate
		}

		recs := re.GenerateRecommendations(forecasts, metrics)

		// Should have SLO recommendation
		var foundSLO bool
		for _, rec := range recs {
			if rec.Category == CategorySLOManagement {
				foundSLO = true
				assert.Contains(t, rec.Title, "SLO")
			}
		}
		assert.True(t, foundSLO)
	})

	t.Run("maintenance window recommendation", func(t *testing.T) {
		// Low forecast for maintenance window
		forecasts := map[MetricType]*ForecastResult{
			MetricBacklog: {
				Points:     make([]float64, 10080), // 7 days
				Confidence: 0.75,
				MetricType: MetricBacklog,
			},
		}

		// Set low values during preferred maintenance hours
		for i := range forecasts[MetricBacklog].Points {
			forecasts[MetricBacklog].Points[i] = 100 // Low baseline
		}

		metrics := &QueueMetrics{}

		recs := re.GenerateRecommendations(forecasts, metrics)

		// Should have maintenance recommendation
		var foundMaintenance bool
		for _, rec := range recs {
			if rec.Category == CategoryMaintenanceScheduling {
				foundMaintenance = true
				assert.Contains(t, rec.Title, "Maintenance")
			}
		}
		assert.True(t, foundMaintenance)
	})
}

func TestMetricsStorage(t *testing.T) {
	logger := zap.NewNop()
	config := &StorageConfig{
		RetentionDuration: 1 * time.Hour,
		SamplingInterval:  1 * time.Second,
		MaxDataPoints:     100,
		PersistToDisk:     false,
	}

	storage := NewMetricsStorage(config, logger)
	defer storage.Stop()

	t.Run("store and retrieve", func(t *testing.T) {
		// Store data points
		for i := 0; i < 10; i++ {
			storage.Store(MetricBacklog, "test-queue", float64(100+i*10))
			time.Sleep(10 * time.Millisecond)
		}

		// Retrieve data
		data := storage.GetTimeSeries(MetricBacklog, "test-queue", 1*time.Minute)
		assert.Len(t, data, 10)

		// Check values
		for i, point := range data {
			assert.Equal(t, float64(100+i*10), point.Value)
		}
	})

	t.Run("get latest", func(t *testing.T) {
		storage.Store(MetricThroughput, "test-queue", 50.5)
		storage.Store(MetricThroughput, "test-queue", 60.5)
		storage.Store(MetricThroughput, "test-queue", 70.5)

		latest, exists := storage.GetLatest(MetricThroughput, "test-queue")
		assert.True(t, exists)
		assert.Equal(t, 70.5, latest)
	})

	t.Run("aggregation", func(t *testing.T) {
		// Store minute data
		now := time.Now()
		for i := 0; i < 60; i++ {
			storage.data["test"] = &TimeSeries{
				Name:       "test",
				MetricType: MetricBacklog,
				Points: append(storage.data["test"].Points, DataPoint{
					Timestamp: now.Add(time.Duration(i) * time.Second),
					Value:     float64(i),
				}),
			}
		}

		// Get aggregated by minute
		aggregated := storage.aggregateByMinute(storage.data["test"].Points)
		assert.LessOrEqual(t, len(aggregated), 2) // Should be 1-2 minutes of data
	})
}

func TestForecastingEngine(t *testing.T) {
	logger := zap.NewNop()
	config := &ForecastConfig{
		EWMAConfig: &EWMAConfig{
			Alpha:           0.3,
			MinObservations: 3,
		},
		HoltWintersConfig: &HoltWintersConfig{
			Alpha:        0.3,
			Beta:         0.1,
			Gamma:        0.1,
			SeasonLength: 24,
		},
		StorageConfig: &StorageConfig{
			RetentionDuration: 1 * time.Hour,
			PersistToDisk:     false,
		},
		EngineConfig: &EngineConfig{
			Enabled:        true,
			UpdateInterval: 100 * time.Millisecond,
		},
	}

	engine := NewForecastingEngine(config, logger)
	defer engine.Stop()

	t.Run("update metrics and forecast", func(t *testing.T) {
		// Update with metrics
		for i := 0; i < 10; i++ {
			metrics := &QueueMetrics{
				Timestamp:     time.Now(),
				Backlog:       int64(100 + i*10),
				Throughput:    10.0 + float64(i),
				ErrorRate:     0.01,
				ActiveWorkers: 5,
				QueueName:     "test-queue",
			}

			err := engine.UpdateMetrics(metrics)
			require.NoError(t, err)
			time.Sleep(10 * time.Millisecond)
		}

		// Get forecasts
		forecasts, err := engine.GetForecasts(60)
		require.NoError(t, err)
		assert.NotEmpty(t, forecasts)

		// Should have forecasts for different metrics
		assert.Contains(t, forecasts, MetricBacklog)
		assert.Contains(t, forecasts, MetricThroughput)
	})

	t.Run("get recommendations", func(t *testing.T) {
		// Wait for update loop
		time.Sleep(200 * time.Millisecond)

		recs := engine.GetRecommendations()
		// May or may not have recommendations depending on metrics
		assert.NotNil(t, recs)
	})

	t.Run("get status", func(t *testing.T) {
		status := engine.GetStatus()
		assert.True(t, status["enabled"].(bool))
		assert.Greater(t, status["models_count"].(int), 0)
	})
}

func TestSLOTracker(t *testing.T) {
	tracker := NewSLOTracker()
	tracker.SetTarget(0.99) // 99% SLO

	t.Run("budget tracking", func(t *testing.T) {
		// Simulate error rates
		for i := 0; i < 100; i++ {
			if i%10 == 0 {
				tracker.Update(0.05) // 5% error spike
			} else {
				tracker.Update(0.005) // 0.5% baseline
			}
		}

		budget := tracker.GetCurrentBudget()
		assert.NotNil(t, budget)
		assert.Equal(t, 0.99, budget.Target)
		assert.Greater(t, budget.WeeklyBurnRate, 0.0)
	})

	t.Run("project budget burn", func(t *testing.T) {
		// Project future errors
		forecast := []float64{0.02, 0.03, 0.04, 0.05} // Increasing errors

		budget := tracker.ProjectBudgetBurn(forecast)
		assert.NotNil(t, budget)
		assert.Greater(t, budget.ProjectedBurn, budget.CurrentBurn)
	})
}
