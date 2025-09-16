// Copyright 2025 James Ross
package forecasting

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// EWMAForecaster implements Exponentially Weighted Moving Average forecasting
type EWMAForecaster struct {
	alpha        float64       // Smoothing parameter (0 < α < 1)
	lastValue    float64       // Previous smoothed value
	variance     float64       // Estimate of forecast variance
	observations int           // Number of observations
	config       *EWMAConfig
	mu           sync.RWMutex

	// Accuracy tracking
	predictions []PredictionRecord
	accuracy    *AccuracyMetrics
}

// NewEWMAForecaster creates a new EWMA forecaster
func NewEWMAForecaster(config *EWMAConfig) *EWMAForecaster {
	if config == nil {
		config = &EWMAConfig{
			Alpha:              0.3,
			AutoAdjust:         true,
			MinObservations:    5,
			ConfidenceInterval: 0.95,
		}
	}

	return &EWMAForecaster{
		alpha:       config.Alpha,
		config:      config,
		predictions: make([]PredictionRecord, 0, 100),
		accuracy:    &AccuracyMetrics{},
	}
}

// Update updates the model with a new observation
func (e *EWMAForecaster) Update(observation float64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.observations == 0 {
		e.lastValue = observation
		e.variance = 0
	} else {
		// Update smoothed value
		prevValue := e.lastValue
		e.lastValue = e.alpha*observation + (1-e.alpha)*e.lastValue

		// Update variance estimate for confidence bounds
		error := observation - prevValue
		e.variance = e.alpha*error*error + (1-e.alpha)*e.variance

		// Auto-adjust alpha if enabled
		if e.config.AutoAdjust && e.observations > 20 {
			e.adjustAlpha(error)
		}
	}

	e.observations++
	return nil
}

// Forecast generates predictions for the specified horizon
func (e *EWMAForecaster) Forecast(horizonMinutes int) (*ForecastResult, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if e.observations < e.config.MinObservations {
		return nil, fmt.Errorf("insufficient observations: %d < %d",
			e.observations, e.config.MinObservations)
	}

	forecasts := make([]float64, horizonMinutes)
	upperBounds := make([]float64, horizonMinutes)
	lowerBounds := make([]float64, horizonMinutes)

	stdDev := math.Sqrt(e.variance)

	// Calculate Z-score for confidence interval
	zScore := e.getZScore(e.config.ConfidenceInterval)

	for i := 0; i < horizonMinutes; i++ {
		// Point forecast remains constant for EWMA
		forecasts[i] = e.lastValue

		// Confidence bounds widen over time
		// Uncertainty grows with sqrt of time
		confidenceMultiplier := zScore * stdDev * math.Sqrt(float64(i+1))
		upperBounds[i] = e.lastValue + confidenceMultiplier
		lowerBounds[i] = e.lastValue - confidenceMultiplier

		// Ensure lower bound doesn't go negative for count metrics
		if lowerBounds[i] < 0 {
			lowerBounds[i] = 0
		}
	}

	// Calculate confidence based on model accuracy
	confidence := e.calculateConfidence()

	return &ForecastResult{
		Points:         forecasts,
		UpperBounds:    upperBounds,
		LowerBounds:    lowerBounds,
		Confidence:     confidence,
		ModelUsed:      "EWMA",
		GeneratedAt:    time.Now(),
		HorizonMinutes: horizonMinutes,
	}, nil
}

// GetAccuracy returns current model accuracy metrics
func (e *EWMAForecaster) GetAccuracy() *AccuracyMetrics {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if len(e.predictions) < 10 {
		return nil
	}

	return e.calculateAccuracy()
}

// GetConfiguration returns the model configuration
func (e *EWMAForecaster) GetConfiguration() ModelConfig {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return ModelConfig{
		ModelType: "EWMA",
		Parameters: map[string]interface{}{
			"alpha":               e.alpha,
			"auto_adjust":         e.config.AutoAdjust,
			"confidence_interval": e.config.ConfidenceInterval,
			"observations":        e.observations,
		},
		Enabled: true,
	}
}

// RecordPrediction records a prediction for accuracy tracking
func (e *EWMAForecaster) RecordPrediction(predicted, actual float64, horizon time.Duration) {
	e.mu.Lock()
	defer e.mu.Unlock()

	error := predicted - actual
	errorPercent := 0.0
	if actual != 0 {
		errorPercent = math.Abs(error) / math.Abs(actual) * 100
	}

	record := PredictionRecord{
		Timestamp:    time.Now(),
		Predicted:    predicted,
		Actual:       actual,
		ModelUsed:    "EWMA",
		Horizon:      horizon,
		Error:        error,
		ErrorPercent: errorPercent,
	}

	e.predictions = append(e.predictions, record)

	// Keep only recent predictions
	if len(e.predictions) > 1000 {
		e.predictions = e.predictions[len(e.predictions)-1000:]
	}

	// Update accuracy metrics
	e.accuracy = e.calculateAccuracy()
}

// Helper methods

func (e *EWMAForecaster) adjustAlpha(error float64) {
	// Simple adaptive alpha adjustment based on error magnitude
	// Increase alpha for larger errors (faster adaptation)
	// Decrease alpha for smaller errors (more smoothing)
	normalizedError := math.Abs(error) / (e.lastValue + 1) // Avoid division by zero

	if normalizedError > 0.2 {
		// Large error - increase responsiveness
		e.alpha = math.Min(0.5, e.alpha*1.1)
	} else if normalizedError < 0.05 {
		// Small error - increase smoothing
		e.alpha = math.Max(0.1, e.alpha*0.95)
	}
}

func (e *EWMAForecaster) getZScore(confidenceLevel float64) float64 {
	// Common confidence levels
	switch confidenceLevel {
	case 0.90:
		return 1.645
	case 0.95:
		return 1.96
	case 0.99:
		return 2.576
	default:
		return 1.96 // Default to 95%
	}
}

func (e *EWMAForecaster) calculateConfidence() float64 {
	if e.accuracy == nil || e.accuracy.SampleSize < 10 {
		// Not enough data for confidence calculation
		return 0.5
	}

	// Base confidence on MAPE (Mean Absolute Percentage Error)
	// Lower MAPE = higher confidence
	if e.accuracy.MAPE < 5 {
		return 0.95
	} else if e.accuracy.MAPE < 10 {
		return 0.85
	} else if e.accuracy.MAPE < 20 {
		return 0.70
	} else if e.accuracy.MAPE < 30 {
		return 0.55
	}
	return 0.40
}

func (e *EWMAForecaster) calculateAccuracy() *AccuracyMetrics {
	if len(e.predictions) == 0 {
		return nil
	}

	var sumAbsError, sumSquaredError, sumPercentError, sumBias float64
	var sumActual, sumActualSquared float64
	n := float64(len(e.predictions))

	for _, pred := range e.predictions {
		absError := math.Abs(pred.Error)
		sumAbsError += absError
		sumSquaredError += pred.Error * pred.Error
		sumPercentError += pred.ErrorPercent
		sumBias += pred.Error
		sumActual += pred.Actual
		sumActualSquared += pred.Actual * pred.Actual
	}

	// Calculate R² score
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
		SampleSize:     len(e.predictions),
		LastUpdated:    time.Now(),
	}
}

// Reset resets the forecaster state
func (e *EWMAForecaster) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.lastValue = 0
	e.variance = 0
	e.observations = 0
	e.alpha = e.config.Alpha
	e.predictions = make([]PredictionRecord, 0, 100)
	e.accuracy = &AccuracyMetrics{}
}
