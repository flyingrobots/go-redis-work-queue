// Copyright 2025 James Ross
package capacityplanning

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"
)

// Forecaster interface defines traffic prediction capabilities
type Forecaster interface {
	Predict(ctx context.Context, history []Metrics, horizon time.Duration) ([]Forecast, error)
	SetModel(model string) error
	GetAccuracy() float64
}

// forecaster implements time-series forecasting
type forecaster struct {
	config   PlannerConfig
	model    string
	accuracy float64
}

// NewForecaster creates a new forecasting engine
func NewForecaster(config PlannerConfig) Forecaster {
	return &forecaster{
		config:   config,
		model:    config.ForecastModel,
		accuracy: 0.0, // Will be calculated based on historical performance
	}
}

// Predict generates arrival rate forecasts for the specified horizon
func (f *forecaster) Predict(ctx context.Context, history []Metrics, horizon time.Duration) ([]Forecast, error) {
	if len(history) < 2 {
		return nil, fmt.Errorf("insufficient history for forecasting: need at least 2 points, got %d", len(history))
	}

	// Sort history by timestamp
	sort.Slice(history, func(i, j int) bool {
		return history[i].Timestamp.Before(history[j].Timestamp)
	})

	switch f.model {
	case "ewma":
		return f.predictEWMA(history, horizon)
	case "holt_winters":
		return f.predictHoltWinters(history, horizon)
	case "linear":
		return f.predictLinear(history, horizon)
	case "seasonal":
		return f.predictSeasonal(history, horizon)
	default:
		// Default to EWMA
		return f.predictEWMA(history, horizon)
	}
}

// predictEWMA implements Exponential Weighted Moving Average
func (f *forecaster) predictEWMA(history []Metrics, horizon time.Duration) ([]Forecast, error) {
	alpha := 0.3 // Smoothing parameter

	if len(history) == 0 {
		return nil, fmt.Errorf("no history provided")
	}

	// Calculate EWMA
	var ewma float64
	for i, metrics := range history {
		if i == 0 {
			ewma = metrics.ArrivalRate
		} else {
			ewma = alpha*metrics.ArrivalRate + (1-alpha)*ewma
		}
	}

	// Calculate prediction error for confidence estimation
	var errors []float64
	predictedEWMA := history[0].ArrivalRate

	for i := 1; i < len(history); i++ {
		actual := history[i].ArrivalRate
		error := math.Abs(actual - predictedEWMA)
		errors = append(errors, error)

		// Update EWMA for next prediction
		predictedEWMA = alpha*actual + (1-alpha)*predictedEWMA
	}

	// Calculate confidence based on prediction errors
	confidence := f.calculateConfidence(errors, ewma)

	// Generate forecast points
	var forecasts []Forecast
	granularity := 5 * time.Minute // 5-minute intervals
	points := int(horizon / granularity)

	baseTime := history[len(history)-1].Timestamp
	for i := 1; i <= points; i++ {
		timestamp := baseTime.Add(time.Duration(i) * granularity)

		// For EWMA, prediction is constant
		forecast := Forecast{
			Timestamp:   timestamp,
			ArrivalRate: ewma,
			Confidence:  confidence,
			Lower:       ewma * 0.8, // 20% confidence band
			Upper:       ewma * 1.2,
			Model:       "EWMA",
		}

		forecasts = append(forecasts, forecast)
	}

	return forecasts, nil
}

// predictHoltWinters implements Holt-Winters exponential smoothing
func (f *forecaster) predictHoltWinters(history []Metrics, horizon time.Duration) ([]Forecast, error) {
	if len(history) < 24 { // Need enough data for trend and seasonality
		return f.predictEWMA(history, horizon) // Fall back to EWMA
	}

	// Holt-Winters parameters
	alpha := 0.3 // Level smoothing
	beta := 0.1  // Trend smoothing
	gamma := 0.1 // Seasonal smoothing

	// Determine seasonal period (assume hourly data, 24-hour seasonality)
	seasonalPeriod := 24

	// Initialize level, trend, and seasonal components
	level, trend, seasonal := f.initializeHoltWinters(history, seasonalPeriod)

	// Apply Holt-Winters smoothing to historical data
	for i := seasonalPeriod; i < len(history); i++ {
		actual := history[i].ArrivalRate
		seasonalIndex := i % seasonalPeriod

		// Update level
		newLevel := alpha*(actual/seasonal[seasonalIndex]) + (1-alpha)*(level+trend)

		// Update trend
		newTrend := beta*(newLevel-level) + (1-beta)*trend

		// Update seasonal
		seasonal[seasonalIndex] = gamma*(actual/newLevel) + (1-gamma)*seasonal[seasonalIndex]

		level = newLevel
		trend = newTrend
	}

	// Calculate prediction errors for confidence
	errors := f.calculateHoltWintersErrors(history, seasonalPeriod, alpha, beta, gamma)
	confidence := f.calculateConfidence(errors, level)

	// Generate forecasts
	var forecasts []Forecast
	granularity := time.Hour // Hourly forecasts for Holt-Winters
	points := int(horizon / granularity)

	baseTime := history[len(history)-1].Timestamp
	for i := 1; i <= points; i++ {
		timestamp := baseTime.Add(time.Duration(i) * granularity)
		seasonalIndex := (len(history) + i - 1) % seasonalPeriod

		// Holt-Winters forecast: (L + h*T) * S
		forecastValue := (level + float64(i)*trend) * seasonal[seasonalIndex]

		// Ensure non-negative
		if forecastValue < 0 {
			forecastValue = 0
		}

		forecast := Forecast{
			Timestamp:   timestamp,
			ArrivalRate: forecastValue,
			Confidence:  confidence,
			Lower:       forecastValue * 0.7, // Wider confidence bands for Holt-Winters
			Upper:       forecastValue * 1.3,
			Model:       "Holt-Winters",
		}

		forecasts = append(forecasts, forecast)
	}

	return forecasts, nil
}

// predictLinear implements simple linear trend forecasting
func (f *forecaster) predictLinear(history []Metrics, horizon time.Duration) ([]Forecast, error) {
	if len(history) < 2 {
		return nil, fmt.Errorf("need at least 2 points for linear forecasting")
	}

	// Calculate linear regression parameters
	n := float64(len(history))
	var sumX, sumY, sumXY, sumX2 float64

	for i, metrics := range history {
		x := float64(i)
		y := metrics.ArrivalRate

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Linear regression: y = a + bx
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// Calculate R² for confidence
	var ssRes, ssTot float64
	yMean := sumY / n

	for i, metrics := range history {
		x := float64(i)
		y := metrics.ArrivalRate
		predicted := intercept + slope*x

		ssRes += math.Pow(y-predicted, 2)
		ssTot += math.Pow(y-yMean, 2)
	}

	r2 := 1.0 - (ssRes / ssTot)
	confidence := math.Max(r2, 0.1) // Minimum 10% confidence

	// Generate forecasts
	var forecasts []Forecast
	granularity := 5 * time.Minute
	points := int(horizon / granularity)

	baseTime := history[len(history)-1].Timestamp
	startIndex := len(history)

	for i := 1; i <= points; i++ {
		timestamp := baseTime.Add(time.Duration(i) * granularity)
		x := float64(startIndex + i)
		forecastValue := intercept + slope*x

		// Ensure non-negative
		if forecastValue < 0 {
			forecastValue = 0
		}

		forecast := Forecast{
			Timestamp:   timestamp,
			ArrivalRate: forecastValue,
			Confidence:  confidence,
			Lower:       forecastValue * 0.9,
			Upper:       forecastValue * 1.1,
			Model:       "Linear",
		}

		forecasts = append(forecasts, forecast)
	}

	return forecasts, nil
}

// predictSeasonal implements seasonal pattern recognition and forecasting
func (f *forecaster) predictSeasonal(history []Metrics, horizon time.Duration) ([]Forecast, error) {
	if len(history) < 48 { // Need at least 2 days of hourly data
		return f.predictEWMA(history, horizon)
	}

	// Detect seasonal patterns (daily, weekly)
	dailyPattern := f.extractDailyPattern(history)
	weeklyPattern := f.extractWeeklyPattern(history)

	// Choose the pattern with higher confidence
	pattern := dailyPattern
	if len(weeklyPattern) > 0 && f.evaluatePatternQuality(weeklyPattern) > f.evaluatePatternQuality(dailyPattern) {
		pattern = weeklyPattern
	}

	// Apply seasonal decomposition
	deseasonalized := f.removeSeasonality(history, pattern)

	// Forecast trend component
	trendForecasts, err := f.predictLinear(deseasonalized, horizon)
	if err != nil {
		return f.predictEWMA(history, horizon)
	}

	// Reapply seasonality to forecasts
	var forecasts []Forecast
	for i, trendForecast := range trendForecasts {
		seasonalIndex := i % len(pattern)
		seasonalFactor := pattern[seasonalIndex]

		forecastValue := trendForecast.ArrivalRate * seasonalFactor

		forecast := Forecast{
			Timestamp:   trendForecast.Timestamp,
			ArrivalRate: forecastValue,
			Confidence:  trendForecast.Confidence * 0.9, // Slightly lower confidence for seasonal
			Lower:       forecastValue * 0.8,
			Upper:       forecastValue * 1.2,
			Model:       "Seasonal",
		}

		forecasts = append(forecasts, forecast)
	}

	return forecasts, nil
}

// Helper methods for Holt-Winters

func (f *forecaster) initializeHoltWinters(history []Metrics, seasonalPeriod int) (float64, float64, []float64) {
	// Initialize level as mean of first seasonal period
	var sum float64
	for i := 0; i < seasonalPeriod && i < len(history); i++ {
		sum += history[i].ArrivalRate
	}
	level := sum / float64(min(seasonalPeriod, len(history)))

	// Initialize trend as average change over first two periods
	trend := 0.0
	if len(history) >= 2*seasonalPeriod {
		sum1 := 0.0
		sum2 := 0.0

		for i := 0; i < seasonalPeriod; i++ {
			sum1 += history[i].ArrivalRate
			sum2 += history[i+seasonalPeriod].ArrivalRate
		}

		trend = (sum2 - sum1) / float64(seasonalPeriod*seasonalPeriod)
	}

	// Initialize seasonal indices
	seasonal := make([]float64, seasonalPeriod)
	for i := 0; i < seasonalPeriod; i++ {
		if level > 0 && i < len(history) {
			seasonal[i] = history[i].ArrivalRate / level
		} else {
			seasonal[i] = 1.0
		}
	}

	return level, trend, seasonal
}

func (f *forecaster) calculateHoltWintersErrors(history []Metrics, seasonalPeriod int, alpha, beta, gamma float64) []float64 {
	if len(history) < seasonalPeriod*2 {
		return []float64{}
	}

	level, trend, seasonal := f.initializeHoltWinters(history, seasonalPeriod)
	var errors []float64

	for i := seasonalPeriod; i < len(history); i++ {
		seasonalIndex := i % seasonalPeriod

		// Forecast for this point
		forecast := (level + trend) * seasonal[seasonalIndex]
		actual := history[i].ArrivalRate
		error := math.Abs(actual - forecast)
		errors = append(errors, error)

		// Update components
		newLevel := alpha*(actual/seasonal[seasonalIndex]) + (1-alpha)*(level+trend)
		newTrend := beta*(newLevel-level) + (1-beta)*trend
		seasonal[seasonalIndex] = gamma*(actual/newLevel) + (1-gamma)*seasonal[seasonalIndex]

		level = newLevel
		trend = newTrend
	}

	return errors
}

// Helper methods for seasonal analysis

func (f *forecaster) extractDailyPattern(history []Metrics) []float64 {
	// Group by hour of day
	hourlyData := make(map[int][]float64)

	for _, metrics := range history {
		hour := metrics.Timestamp.Hour()
		hourlyData[hour] = append(hourlyData[hour], metrics.ArrivalRate)
	}

	// Calculate average for each hour
	pattern := make([]float64, 24)
	var totalAverage float64

	for hour := 0; hour < 24; hour++ {
		if data, exists := hourlyData[hour]; exists && len(data) > 0 {
			sum := 0.0
			for _, value := range data {
				sum += value
			}
			pattern[hour] = sum / float64(len(data))
			totalAverage += pattern[hour]
		}
	}

	// Normalize pattern (seasonal indices should average to 1.0)
	if totalAverage > 0 {
		avgValue := totalAverage / 24.0
		for i := range pattern {
			pattern[i] /= avgValue
		}
	} else {
		// If no data, use flat pattern
		for i := range pattern {
			pattern[i] = 1.0
		}
	}

	return pattern
}

func (f *forecaster) extractWeeklyPattern(history []Metrics) []float64 {
	// Group by day of week
	weeklyData := make(map[int][]float64)

	for _, metrics := range history {
		day := int(metrics.Timestamp.Weekday())
		weeklyData[day] = append(weeklyData[day], metrics.ArrivalRate)
	}

	// Calculate average for each day
	pattern := make([]float64, 7)
	var totalAverage float64

	for day := 0; day < 7; day++ {
		if data, exists := weeklyData[day]; exists && len(data) > 0 {
			sum := 0.0
			for _, value := range data {
				sum += value
			}
			pattern[day] = sum / float64(len(data))
			totalAverage += pattern[day]
		}
	}

	// Normalize pattern
	if totalAverage > 0 {
		avgValue := totalAverage / 7.0
		for i := range pattern {
			pattern[i] /= avgValue
		}
	} else {
		for i := range pattern {
			pattern[i] = 1.0
		}
	}

	return pattern
}

func (f *forecaster) evaluatePatternQuality(pattern []float64) float64 {
	if len(pattern) == 0 {
		return 0.0
	}

	// Calculate coefficient of variation as a measure of pattern strength
	mean := 0.0
	for _, value := range pattern {
		mean += value
	}
	mean /= float64(len(pattern))

	variance := 0.0
	for _, value := range pattern {
		variance += math.Pow(value-mean, 2)
	}
	variance /= float64(len(pattern))

	if mean == 0 {
		return 0.0
	}

	cv := math.Sqrt(variance) / mean
	return cv // Higher CV indicates stronger seasonal pattern
}

func (f *forecaster) removeSeasonality(history []Metrics, pattern []float64) []Metrics {
	if len(pattern) == 0 {
		return history
	}

	deseasonalized := make([]Metrics, len(history))

	for i, metrics := range history {
		seasonalIndex := i % len(pattern)
		seasonalFactor := pattern[seasonalIndex]

		deseasonalized[i] = metrics
		if seasonalFactor > 0 {
			deseasonalized[i].ArrivalRate = metrics.ArrivalRate / seasonalFactor
		}
	}

	return deseasonalized
}

// calculateConfidence estimates forecast confidence based on prediction errors
func (f *forecaster) calculateConfidence(errors []float64, forecast float64) float64 {
	if len(errors) == 0 {
		return 0.5 // Default confidence
	}

	// Calculate Mean Absolute Percentage Error (MAPE)
	var mape float64
	validErrors := 0

	for _, error := range errors {
		if forecast > 0 {
			mape += error / forecast
			validErrors++
		}
	}

	if validErrors == 0 {
		return 0.5
	}

	mape /= float64(validErrors)

	// Convert MAPE to confidence (lower error = higher confidence)
	confidence := math.Max(0.1, 1.0-mape)
	return math.Min(confidence, 0.95) // Cap at 95%
}

// SetModel changes the forecasting model
func (f *forecaster) SetModel(model string) error {
	validModels := []string{"ewma", "holt_winters", "linear", "seasonal"}

	for _, valid := range validModels {
		if model == valid {
			f.model = model
			return nil
		}
	}

	return fmt.Errorf("unsupported forecasting model: %s", model)
}

// GetAccuracy returns the current forecasting accuracy
func (f *forecaster) GetAccuracy() float64 {
	return f.accuracy
}

// Utility functions

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}