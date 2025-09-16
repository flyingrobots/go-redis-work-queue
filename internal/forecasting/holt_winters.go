// Copyright 2025 James Ross
package forecasting

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// HoltWintersForecaster implements Holt-Winters triple exponential smoothing
type HoltWintersForecaster struct {
	level        float64              // Current level component
	trend        float64              // Current trend component
	seasonal     []float64            // Seasonal components
	alpha        float64              // Level smoothing parameter
	beta         float64              // Trend smoothing parameter
	gamma        float64              // Seasonal smoothing parameter
	seasonLength int                  // Length of seasonal cycle
	method       string               // "additive" or "multiplicative"
	observations int                  // Number of observations
	config       *HoltWintersConfig
	mu           sync.RWMutex

	// Historical data for initialization
	history []float64

	// Accuracy tracking
	predictions []PredictionRecord
	accuracy    *AccuracyMetrics
}

// NewHoltWintersForecaster creates a new Holt-Winters forecaster
func NewHoltWintersForecaster(config *HoltWintersConfig) *HoltWintersForecaster {
	if config == nil {
		config = &HoltWintersConfig{
			Alpha:            0.3,
			Beta:             0.1,
			Gamma:            0.1,
			SeasonLength:     24, // Default to daily seasonality (hourly data)
			SeasonalMethod:   "additive",
			AutoDetectSeason: true,
		}
	}

	hw := &HoltWintersForecaster{
		alpha:        config.Alpha,
		beta:         config.Beta,
		gamma:        config.Gamma,
		seasonLength: config.SeasonLength,
		method:       config.SeasonalMethod,
		config:       config,
		seasonal:     make([]float64, config.SeasonLength),
		history:      make([]float64, 0, config.SeasonLength*2),
		predictions:  make([]PredictionRecord, 0, 100),
		accuracy:     &AccuracyMetrics{},
	}

	// Initialize seasonal components
	for i := range hw.seasonal {
		if config.SeasonalMethod == "multiplicative" {
			hw.seasonal[i] = 1.0
		} else {
			hw.seasonal[i] = 0.0
		}
	}

	return hw
}

// Update updates the model with a new observation
func (hw *HoltWintersForecaster) Update(observation float64) error {
	hw.mu.Lock()
	defer hw.mu.Unlock()

	// Store history for initialization
	hw.history = append(hw.history, observation)

	// Need at least 2 seasons of data for proper initialization
	if len(hw.history) < hw.seasonLength*2 {
		return nil // Still collecting initial data
	}

	// Initialize on first complete dataset
	if hw.observations == 0 && len(hw.history) >= hw.seasonLength*2 {
		hw.initialize()
	}

	if hw.observations > 0 {
		period := hw.observations % hw.seasonLength

		if hw.method == "multiplicative" {
			hw.updateMultiplicative(observation, period)
		} else {
			hw.updateAdditive(observation, period)
		}
	}

	hw.observations++
	return nil
}

// Forecast generates predictions for the specified horizon
func (hw *HoltWintersForecaster) Forecast(horizonMinutes int) (*ForecastResult, error) {
	hw.mu.RLock()
	defer hw.mu.RUnlock()

	if hw.observations < hw.seasonLength*2 {
		return nil, fmt.Errorf("insufficient observations for Holt-Winters: %d < %d",
			hw.observations, hw.seasonLength*2)
	}

	forecasts := make([]float64, horizonMinutes)
	upperBounds := make([]float64, horizonMinutes)
	lowerBounds := make([]float64, horizonMinutes)

	// Calculate forecast variance for confidence intervals
	variance := hw.calculateVariance()
	stdDev := math.Sqrt(variance)
	zScore := 1.96 // 95% confidence interval

	for i := 0; i < horizonMinutes; i++ {
		seasonalIndex := i % hw.seasonLength
		levelWithTrend := hw.level + float64(i+1)*hw.trend

		if hw.method == "multiplicative" {
			forecasts[i] = levelWithTrend * hw.seasonal[seasonalIndex]
		} else {
			forecasts[i] = levelWithTrend + hw.seasonal[seasonalIndex]
		}

		// Confidence bounds widen over time
		confidenceMultiplier := zScore * stdDev * math.Sqrt(float64(i+1))
		upperBounds[i] = forecasts[i] + confidenceMultiplier
		lowerBounds[i] = forecasts[i] - confidenceMultiplier

		// Ensure non-negative for count metrics
		if lowerBounds[i] < 0 {
			lowerBounds[i] = 0
		}
		if forecasts[i] < 0 {
			forecasts[i] = 0
		}
	}

	confidence := hw.calculateConfidence()

	return &ForecastResult{
		Points:         forecasts,
		UpperBounds:    upperBounds,
		LowerBounds:    lowerBounds,
		Confidence:     confidence,
		ModelUsed:      "Holt-Winters",
		GeneratedAt:    time.Now(),
		HorizonMinutes: horizonMinutes,
	}, nil
}

// GetAccuracy returns current model accuracy metrics
func (hw *HoltWintersForecaster) GetAccuracy() *AccuracyMetrics {
	hw.mu.RLock()
	defer hw.mu.RUnlock()

	if len(hw.predictions) < 10 {
		return nil
	}

	return hw.calculateAccuracy()
}

// GetConfiguration returns the model configuration
func (hw *HoltWintersForecaster) GetConfiguration() ModelConfig {
	hw.mu.RLock()
	defer hw.mu.RUnlock()

	return ModelConfig{
		ModelType: "Holt-Winters",
		Parameters: map[string]interface{}{
			"alpha":         hw.alpha,
			"beta":          hw.beta,
			"gamma":         hw.gamma,
			"season_length": hw.seasonLength,
			"method":        hw.method,
			"observations":  hw.observations,
		},
		Enabled: true,
	}
}

// Helper methods

func (hw *HoltWintersForecaster) initialize() {
	// Initialize level as average of first season
	sum := 0.0
	for i := 0; i < hw.seasonLength; i++ {
		sum += hw.history[i]
	}
	hw.level = sum / float64(hw.seasonLength)

	// Initialize trend as average change between seasons
	trendSum := 0.0
	for i := 0; i < hw.seasonLength; i++ {
		trendSum += (hw.history[hw.seasonLength+i] - hw.history[i]) / float64(hw.seasonLength)
	}
	hw.trend = trendSum / float64(hw.seasonLength)

	// Initialize seasonal components
	if hw.method == "multiplicative" {
		hw.initializeMultiplicativeSeasonal()
	} else {
		hw.initializeAdditiveSeasonal()
	}

	// Auto-detect seasonality if enabled
	if hw.config.AutoDetectSeason {
		hw.detectSeasonality()
	}
}

func (hw *HoltWintersForecaster) initializeAdditiveSeasonal() {
	for i := 0; i < hw.seasonLength; i++ {
		seasonSum := 0.0
		count := 0

		// Average deviation from trend for each seasonal period
		for j := 0; j*hw.seasonLength+i < len(hw.history); j++ {
			index := j*hw.seasonLength + i
			expected := hw.level + float64(index)*hw.trend
			seasonSum += hw.history[index] - expected
			count++
		}

		if count > 0 {
			hw.seasonal[i] = seasonSum / float64(count)
		}
	}
}

func (hw *HoltWintersForecaster) initializeMultiplicativeSeasonal() {
	for i := 0; i < hw.seasonLength; i++ {
		seasonSum := 0.0
		count := 0

		// Average ratio to trend for each seasonal period
		for j := 0; j*hw.seasonLength+i < len(hw.history); j++ {
			index := j*hw.seasonLength + i
			expected := hw.level + float64(index)*hw.trend
			if expected > 0 {
				seasonSum += hw.history[index] / expected
				count++
			}
		}

		if count > 0 {
			hw.seasonal[i] = seasonSum / float64(count)
		} else {
			hw.seasonal[i] = 1.0
		}
	}
}

func (hw *HoltWintersForecaster) updateAdditive(observation float64, period int) {
	// Deseasonalize observation
	deseasonalized := observation - hw.seasonal[period]

	// Update level and trend
	previousLevel := hw.level
	hw.level = hw.alpha*deseasonalized + (1-hw.alpha)*(hw.level+hw.trend)
	hw.trend = hw.beta*(hw.level-previousLevel) + (1-hw.beta)*hw.trend

	// Update seasonal component
	hw.seasonal[period] = hw.gamma*(observation-hw.level) + (1-hw.gamma)*hw.seasonal[period]
}

func (hw *HoltWintersForecaster) updateMultiplicative(observation float64, period int) {
	// Deseasonalize observation
	deseasonalized := observation
	if hw.seasonal[period] != 0 {
		deseasonalized = observation / hw.seasonal[period]
	}

	// Update level and trend
	previousLevel := hw.level
	hw.level = hw.alpha*deseasonalized + (1-hw.alpha)*(hw.level+hw.trend)
	hw.trend = hw.beta*(hw.level-previousLevel) + (1-hw.beta)*hw.trend

	// Update seasonal component
	if hw.level != 0 {
		hw.seasonal[period] = hw.gamma*(observation/hw.level) + (1-hw.gamma)*hw.seasonal[period]
	}
}

func (hw *HoltWintersForecaster) detectSeasonality() {
	// Simple autocorrelation-based seasonality detection
	if len(hw.history) < hw.seasonLength*3 {
		return
	}

	bestCorrelation := 0.0
	bestPeriod := hw.seasonLength

	// Test different seasonal periods
	for period := 2; period <= 48; period++ {
		if len(hw.history) < period*2 {
			continue
		}

		correlation := hw.calculateAutocorrelation(period)
		if correlation > bestCorrelation {
			bestCorrelation = correlation
			bestPeriod = period
		}
	}

	// Update season length if significantly better correlation found
	if bestCorrelation > 0.7 && bestPeriod != hw.seasonLength {
		hw.seasonLength = bestPeriod
		hw.seasonal = make([]float64, bestPeriod)
		hw.initialize() // Re-initialize with new season length
	}
}

func (hw *HoltWintersForecaster) calculateAutocorrelation(lag int) float64 {
	n := len(hw.history)
	if n < lag*2 {
		return 0
	}

	// Calculate mean
	mean := 0.0
	for _, v := range hw.history {
		mean += v
	}
	mean /= float64(n)

	// Calculate autocorrelation
	numerator := 0.0
	denominator := 0.0

	for i := lag; i < n; i++ {
		numerator += (hw.history[i] - mean) * (hw.history[i-lag] - mean)
	}

	for i := 0; i < n; i++ {
		denominator += (hw.history[i] - mean) * (hw.history[i] - mean)
	}

	if denominator == 0 {
		return 0
	}

	return numerator / denominator
}

func (hw *HoltWintersForecaster) calculateVariance() float64 {
	if len(hw.history) < 2 {
		return 0
	}

	// Calculate residual variance
	var sumSquaredError float64
	count := 0

	for i := hw.seasonLength; i < len(hw.history); i++ {
		period := (i - hw.seasonLength) % hw.seasonLength
		var predicted float64

		if hw.method == "multiplicative" {
			predicted = (hw.level + hw.trend) * hw.seasonal[period]
		} else {
			predicted = hw.level + hw.trend + hw.seasonal[period]
		}

		error := hw.history[i] - predicted
		sumSquaredError += error * error
		count++
	}

	if count > 0 {
		return sumSquaredError / float64(count)
	}
	return 0
}

func (hw *HoltWintersForecaster) calculateConfidence() float64 {
	if hw.accuracy == nil || hw.accuracy.SampleSize < 10 {
		return 0.5
	}

	// Base confidence on MAPE
	if hw.accuracy.MAPE < 5 {
		return 0.92
	} else if hw.accuracy.MAPE < 10 {
		return 0.82
	} else if hw.accuracy.MAPE < 20 {
		return 0.68
	} else if hw.accuracy.MAPE < 30 {
		return 0.52
	}
	return 0.35
}

func (hw *HoltWintersForecaster) calculateAccuracy() *AccuracyMetrics {
	if len(hw.predictions) == 0 {
		return nil
	}

	var sumAbsError, sumSquaredError, sumPercentError, sumBias float64
	var sumActual, sumActualSquared float64
	n := float64(len(hw.predictions))

	for _, pred := range hw.predictions {
		absError := math.Abs(pred.Error)
		sumAbsError += absError
		sumSquaredError += pred.Error * pred.Error
		sumPercentError += pred.ErrorPercent
		sumBias += pred.Error
		sumActual += pred.Actual
		sumActualSquared += pred.Actual * pred.Actual
	}

	// Calculate R²
	meanActual := sumActual / n
	totalSumSquares := sumActualSquared - n*meanActual*meanActual
	residualSumSquares := sumSquaredError
	r2Score := 1.0
	if totalSumSquares > 0 {
		r2Score = 1.0 - (residualSumSquares / totalSumSquares)
	}

	return &AccuracyMetrics{
		MAE:            sumAbsError / n,
		RMSE:           math.Sqrt(sumSquaredError / n),
		MAPE:           sumPercentError / n,
		PredictionBias: sumBias / n,
		R2Score:        r2Score,
		SampleSize:     len(hw.predictions),
		LastUpdated:    time.Now(),
	}
}
