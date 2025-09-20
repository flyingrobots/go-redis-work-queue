//go:build capacity_planning_tests
// +build capacity_planning_tests

// Copyright 2025 James Ross
package capacityplanning

import (
	"context"
	"math"
	"testing"
	"time"
)

func TestNewForecaster(t *testing.T) {
	config := PlannerConfig{
		ForecastModel: "ewma",
	}

	forecaster := NewForecaster(config)
	if forecaster == nil {
		t.Fatal("NewForecaster returned nil")
	}

	// Test interface compliance
	var _ Forecaster = forecaster
}

func TestPredictEWMA(t *testing.T) {
	config := PlannerConfig{
		ForecastModel: "ewma",
	}

	forecaster := NewForecaster(config)

	// Create historical data with trend
	history := make([]Metrics, 10)
	baseTime := time.Now().Add(-10 * time.Hour)
	for i := range history {
		history[i] = Metrics{
			Timestamp:   baseTime.Add(time.Duration(i) * time.Hour),
			ArrivalRate: float64(10 + i), // Increasing trend
		}
	}

	ctx := context.Background()
	horizon := 2 * time.Hour

	forecasts, err := forecaster.Predict(ctx, history, horizon)
	if err != nil {
		t.Fatalf("Predict() failed: %v", err)
	}

	if len(forecasts) == 0 {
		t.Fatal("Predict() returned empty forecasts")
	}

	// EWMA should give constant prediction
	expectedRate := history[len(history)-1].ArrivalRate // Roughly the last value
	tolerance := 5.0                                    // Allow some smoothing difference

	for i, forecast := range forecasts {
		if forecast.Model != "EWMA" {
			t.Errorf("Forecast[%d] model = %v, want EWMA", i, forecast.Model)
		}

		if math.Abs(forecast.ArrivalRate-expectedRate) > tolerance {
			t.Errorf("Forecast[%d] rate = %v, want ~%v", i, forecast.ArrivalRate, expectedRate)
		}

		if forecast.Confidence <= 0 || forecast.Confidence > 1 {
			t.Errorf("Forecast[%d] confidence = %v, want 0 < conf <= 1", i, forecast.Confidence)
		}

		if forecast.Lower >= forecast.ArrivalRate || forecast.Upper <= forecast.ArrivalRate {
			t.Errorf("Forecast[%d] confidence bands invalid: %v <= %v <= %v",
				forecast.Lower, forecast.ArrivalRate, forecast.Upper)
		}

		// Check timestamp progression
		expectedTime := baseTime.Add(time.Duration(len(history)+i+1) * 5 * time.Minute)
		if math.Abs(forecast.Timestamp.Sub(expectedTime).Seconds()) > 60 {
			t.Errorf("Forecast[%d] timestamp = %v, want ~%v", i, forecast.Timestamp, expectedTime)
		}
	}
}

func TestPredictLinear(t *testing.T) {
	config := PlannerConfig{
		ForecastModel: "linear",
	}

	forecaster := NewForecaster(config)

	// Create data with clear linear trend
	history := make([]Metrics, 5)
	baseTime := time.Now().Add(-5 * time.Hour)
	for i := range history {
		history[i] = Metrics{
			Timestamp:   baseTime.Add(time.Duration(i) * time.Hour),
			ArrivalRate: float64(10 + 2*i), // Linear: 10, 12, 14, 16, 18
		}
	}

	ctx := context.Background()
	horizon := 1 * time.Hour

	forecasts, err := forecaster.Predict(ctx, history, horizon)
	if err != nil {
		t.Fatalf("Predict() failed: %v", err)
	}

	if len(forecasts) == 0 {
		t.Fatal("Predict() returned empty forecasts")
	}

	for i, forecast := range forecasts {
		if forecast.Model != "Linear" {
			t.Errorf("Forecast[%d] model = %v, want Linear", i, forecast.Model)
		}

		// Linear trend should continue: next value should be around 20, 22, etc.
		expectedRate := 20.0 + float64(i)*2.0 // Continuing the trend
		tolerance := 1.0

		if math.Abs(forecast.ArrivalRate-expectedRate) > tolerance {
			t.Errorf("Forecast[%d] rate = %v, want ~%v", i, forecast.ArrivalRate, expectedRate)
		}

		if forecast.ArrivalRate < 0 {
			t.Errorf("Forecast[%d] rate = %v, want non-negative", i, forecast.ArrivalRate)
		}
	}
}

func TestPredictHoltWinters(t *testing.T) {
	config := PlannerConfig{
		ForecastModel: "holt_winters",
	}

	forecaster := NewForecaster(config)

	// Create sufficient data for Holt-Winters (need 24+ points)
	history := make([]Metrics, 30)
	baseTime := time.Now().Add(-30 * time.Hour)
	for i := range history {
		// Create data with trend and seasonality
		trend := float64(i) * 0.5
		seasonality := 5 * math.Sin(2*math.Pi*float64(i)/24) // 24-hour cycle
		history[i] = Metrics{
			Timestamp:   baseTime.Add(time.Duration(i) * time.Hour),
			ArrivalRate: 10 + trend + seasonality,
		}
	}

	ctx := context.Background()
	horizon := 6 * time.Hour

	forecasts, err := forecaster.Predict(ctx, history, horizon)
	if err != nil {
		t.Fatalf("Predict() failed: %v", err)
	}

	if len(forecasts) == 0 {
		t.Fatal("Predict() returned empty forecasts")
	}

	for i, forecast := range forecasts {
		if forecast.Model != "Holt-Winters" {
			t.Errorf("Forecast[%d] model = %v, want Holt-Winters", i, forecast.Model)
		}

		if forecast.ArrivalRate < 0 {
			t.Errorf("Forecast[%d] rate = %v, want non-negative", i, forecast.ArrivalRate)
		}

		// Should have wider confidence bands than EWMA
		bandWidth := forecast.Upper - forecast.Lower
		if bandWidth < 1.0 {
			t.Errorf("Forecast[%d] confidence band too narrow: %v", i, bandWidth)
		}
	}
}

func TestPredictSeasonal(t *testing.T) {
	config := PlannerConfig{
		ForecastModel: "seasonal",
	}

	forecaster := NewForecaster(config)

	// Create data with clear daily pattern (need 48+ points for seasonal)
	history := make([]Metrics, 72) // 3 days of hourly data
	baseTime := time.Now().Add(-72 * time.Hour)
	for i := range history {
		hour := i % 24
		dailyMultiplier := 1.0
		if hour >= 9 && hour <= 17 {
			dailyMultiplier = 2.0 // Business hours peak
		} else if hour >= 0 && hour <= 6 {
			dailyMultiplier = 0.3 // Night low
		}

		history[i] = Metrics{
			Timestamp:   baseTime.Add(time.Duration(i) * time.Hour),
			ArrivalRate: 10 * dailyMultiplier,
		}
	}

	ctx := context.Background()
	horizon := 24 * time.Hour // One day ahead

	forecasts, err := forecaster.Predict(ctx, history, horizon)
	if err != nil {
		t.Fatalf("Predict() failed: %v", err)
	}

	if len(forecasts) == 0 {
		t.Fatal("Predict() returned empty forecasts")
	}

	// Should detect the daily pattern
	businessHourForecasts := make([]Forecast, 0)
	nightHourForecasts := make([]Forecast, 0)

	for _, forecast := range forecasts {
		if forecast.Model != "Seasonal" {
			t.Errorf("Forecast model = %v, want Seasonal", forecast.Model)
		}

		hour := forecast.Timestamp.Hour()
		if hour >= 9 && hour <= 17 {
			businessHourForecasts = append(businessHourForecasts, forecast)
		} else if hour >= 0 && hour <= 6 {
			nightHourForecasts = append(nightHourForecasts, forecast)
		}
	}

	// Business hours should have higher rates than night hours
	if len(businessHourForecasts) > 0 && len(nightHourForecasts) > 0 {
		avgBusinessRate := 0.0
		for _, f := range businessHourForecasts {
			avgBusinessRate += f.ArrivalRate
		}
		avgBusinessRate /= float64(len(businessHourForecasts))

		avgNightRate := 0.0
		for _, f := range nightHourForecasts {
			avgNightRate += f.ArrivalRate
		}
		avgNightRate /= float64(len(nightHourForecasts))

		if avgBusinessRate <= avgNightRate {
			t.Errorf("Business hour rate %v should be higher than night rate %v",
				avgBusinessRate, avgNightRate)
		}
	}
}

func TestPredictInsufficientHistory(t *testing.T) {
	config := PlannerConfig{
		ForecastModel: "ewma",
	}

	forecaster := NewForecaster(config)

	// Test with insufficient history
	history := []Metrics{
		{
			Timestamp:   time.Now().Add(-1 * time.Hour),
			ArrivalRate: 10.0,
		},
	}

	ctx := context.Background()
	horizon := 1 * time.Hour

	_, err := forecaster.Predict(ctx, history, horizon)
	if err == nil {
		t.Error("Expected error with insufficient history")
	}

	if !containsString(err.Error(), "insufficient history") {
		t.Errorf("Error should mention insufficient history: %v", err)
	}
}

func TestSetModel(t *testing.T) {
	config := PlannerConfig{
		ForecastModel: "ewma",
	}

	forecaster := NewForecaster(config)

	validModels := []string{"ewma", "holt_winters", "linear", "seasonal"}

	for _, model := range validModels {
		err := forecaster.SetModel(model)
		if err != nil {
			t.Errorf("SetModel(%s) failed: %v", model, err)
		}
	}

	// Test invalid model
	err := forecaster.SetModel("invalid_model")
	if err == nil {
		t.Error("Expected error for invalid model")
	}
}

func TestGetAccuracy(t *testing.T) {
	config := PlannerConfig{
		ForecastModel: "ewma",
	}

	forecaster := NewForecaster(config)

	accuracy := forecaster.GetAccuracy()
	if accuracy < 0 || accuracy > 1 {
		t.Errorf("Accuracy = %v, want 0 <= accuracy <= 1", accuracy)
	}
}

func TestDefaultModel(t *testing.T) {
	config := PlannerConfig{
		ForecastModel: "invalid_model",
	}

	forecaster := NewForecaster(config)

	history := []Metrics{
		{Timestamp: time.Now().Add(-2 * time.Hour), ArrivalRate: 10.0},
		{Timestamp: time.Now().Add(-1 * time.Hour), ArrivalRate: 12.0},
	}

	ctx := context.Background()
	forecasts, err := forecaster.Predict(ctx, history, 1*time.Hour)

	if err != nil {
		t.Fatalf("Predict() with invalid model failed: %v", err)
	}

	// Should default to EWMA
	if len(forecasts) > 0 && forecasts[0].Model != "EWMA" {
		t.Errorf("Default model = %v, want EWMA", forecasts[0].Model)
	}
}

func TestExtractDailyPattern(t *testing.T) {
	config := PlannerConfig{ForecastModel: "seasonal"}
	forecaster := NewForecaster(config).(*forecaster)

	// Create data spanning multiple days with clear hourly pattern
	history := make([]Metrics, 72) // 3 days
	baseTime := time.Now().Add(-72 * time.Hour).Truncate(time.Hour)

	for i := range history {
		hour := i % 24
		// Create pattern: low at night (0-6), high during day (9-17), medium otherwise
		rate := 10.0
		if hour >= 0 && hour <= 6 {
			rate = 5.0 // Night
		} else if hour >= 9 && hour <= 17 {
			rate = 20.0 // Business hours
		}

		history[i] = Metrics{
			Timestamp:   baseTime.Add(time.Duration(i) * time.Hour),
			ArrivalRate: rate,
		}
	}

	pattern := forecaster.extractDailyPattern(history)

	if len(pattern) != 24 {
		t.Fatalf("Daily pattern length = %d, want 24", len(pattern))
	}

	// Verify pattern captures the structure
	// Night hours should have lower multipliers
	nightAvg := (pattern[0] + pattern[1] + pattern[2] + pattern[3] + pattern[4] + pattern[5] + pattern[6]) / 7
	businessAvg := (pattern[9] + pattern[10] + pattern[11] + pattern[12] + pattern[13] + pattern[14] + pattern[15] + pattern[16] + pattern[17]) / 9

	if businessAvg <= nightAvg {
		t.Errorf("Business hours pattern %v should be higher than night pattern %v",
			businessAvg, nightAvg)
	}

	// Pattern should be normalized (average ~1.0)
	sum := 0.0
	for _, value := range pattern {
		sum += value
	}
	avg := sum / 24.0

	if math.Abs(avg-1.0) > 0.1 {
		t.Errorf("Daily pattern average = %v, want ~1.0", avg)
	}
}

func TestExtractWeeklyPattern(t *testing.T) {
	config := PlannerConfig{ForecastModel: "seasonal"}
	forecaster := NewForecaster(config).(*forecaster)

	// Create data spanning multiple weeks
	history := make([]Metrics, 21*24) // 3 weeks of hourly data
	baseTime := time.Now().Add(-21 * 24 * time.Hour).Truncate(time.Hour)

	for i := range history {
		dayOfWeek := (i / 24) % 7
		// Weekdays higher than weekends
		rate := 10.0
		if dayOfWeek == 0 || dayOfWeek == 6 { // Sunday (0) or Saturday (6)
			rate = 5.0
		} else {
			rate = 15.0
		}

		history[i] = Metrics{
			Timestamp:   baseTime.Add(time.Duration(i) * time.Hour),
			ArrivalRate: rate,
		}
	}

	pattern := forecaster.extractWeeklyPattern(history)

	if len(pattern) != 7 {
		t.Fatalf("Weekly pattern length = %d, want 7", len(pattern))
	}

	// Weekdays should have higher values than weekends
	weekdayAvg := (pattern[1] + pattern[2] + pattern[3] + pattern[4] + pattern[5]) / 5 // Mon-Fri
	weekendAvg := (pattern[0] + pattern[6]) / 2                                        // Sun, Sat

	if weekdayAvg <= weekendAvg {
		t.Errorf("Weekday pattern %v should be higher than weekend pattern %v",
			weekdayAvg, weekendAvg)
	}
}

func TestCalculateConfidence(t *testing.T) {
	config := PlannerConfig{ForecastModel: "ewma"}
	forecaster := NewForecaster(config).(*forecaster)

	tests := []struct {
		name        string
		errors      []float64
		forecast    float64
		wantMinConf float64
		wantMaxConf float64
	}{
		{
			name:        "no errors",
			errors:      []float64{},
			forecast:    10.0,
			wantMinConf: 0.4,
			wantMaxConf: 0.6,
		},
		{
			name:        "low errors",
			errors:      []float64{0.5, 0.3, 0.8, 0.2},
			forecast:    10.0,
			wantMinConf: 0.8,
			wantMaxConf: 1.0,
		},
		{
			name:        "high errors",
			errors:      []float64{5.0, 8.0, 6.0, 7.0},
			forecast:    10.0,
			wantMinConf: 0.1,
			wantMaxConf: 0.4,
		},
		{
			name:        "zero forecast",
			errors:      []float64{1.0, 2.0},
			forecast:    0.0,
			wantMinConf: 0.4,
			wantMaxConf: 0.6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := forecaster.calculateConfidence(tt.errors, tt.forecast)

			if conf < 0.1 || conf > 0.95 {
				t.Errorf("confidence = %v, want between 0.1 and 0.95", conf)
			}

			if conf < tt.wantMinConf || conf > tt.wantMaxConf {
				t.Errorf("confidence = %v, want between %v and %v", conf, tt.wantMinConf, tt.wantMaxConf)
			}
		})
	}
}

func TestEvaluatePatternQuality(t *testing.T) {
	config := PlannerConfig{ForecastModel: "seasonal"}
	forecaster := NewForecaster(config).(*forecaster)

	tests := []struct {
		name    string
		pattern []float64
		wantCV  float64 // Expected coefficient of variation range
	}{
		{
			name:    "flat pattern",
			pattern: []float64{1.0, 1.0, 1.0, 1.0, 1.0},
			wantCV:  0.0,
		},
		{
			name:    "variable pattern",
			pattern: []float64{0.5, 1.5, 0.8, 1.2, 0.9},
			wantCV:  0.3, // Should have some variation
		},
		{
			name:    "highly variable pattern",
			pattern: []float64{0.1, 2.0, 0.5, 1.8, 0.3},
			wantCV:  0.7, // High variation
		},
		{
			name:    "empty pattern",
			pattern: []float64{},
			wantCV:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			quality := forecaster.evaluatePatternQuality(tt.pattern)

			if tt.name == "empty pattern" {
				if quality != 0.0 {
					t.Errorf("Empty pattern quality = %v, want 0.0", quality)
				}
				return
			}

			if tt.name == "flat pattern" {
				if quality != 0.0 {
					t.Errorf("Flat pattern quality = %v, want 0.0", quality)
				}
				return
			}

			if quality < 0 {
				t.Errorf("Pattern quality = %v, want non-negative", quality)
			}

			// Higher variation should give higher quality scores
			if tt.wantCV > 0.5 && quality < 0.3 {
				t.Errorf("High variation pattern quality = %v, want > 0.3", quality)
			}
		})
	}
}

func TestRemoveSeasonality(t *testing.T) {
	config := PlannerConfig{ForecastModel: "seasonal"}
	forecaster := NewForecaster(config).(*forecaster)

	history := []Metrics{
		{ArrivalRate: 10.0}, // Affected by pattern[0]
		{ArrivalRate: 20.0}, // Affected by pattern[1]
		{ArrivalRate: 15.0}, // Affected by pattern[2]
		{ArrivalRate: 12.0}, // Affected by pattern[0] again
	}

	pattern := []float64{0.5, 2.0, 1.5} // Seasonal factors

	deseasonalized := forecaster.removeSeasonality(history, pattern)

	if len(deseasonalized) != len(history) {
		t.Fatalf("Deseasonalized length = %d, want %d", len(deseasonalized), len(history))
	}

	// Check deseasonalization: original / seasonal_factor
	expectedRates := []float64{
		10.0 / 0.5, // 20.0
		20.0 / 2.0, // 10.0
		15.0 / 1.5, // 10.0
		12.0 / 0.5, // 24.0
	}

	for i, expected := range expectedRates {
		if math.Abs(deseasonalized[i].ArrivalRate-expected) > 0.001 {
			t.Errorf("Deseasonalized[%d] = %v, want %v", i, deseasonalized[i].ArrivalRate, expected)
		}
	}

	// Test with empty pattern
	emptyPattern := []float64{}
	result := forecaster.removeSeasonality(history, emptyPattern)

	for i := range result {
		if result[i].ArrivalRate != history[i].ArrivalRate {
			t.Errorf("Empty pattern should not change data: got %v, want %v",
				result[i].ArrivalRate, history[i].ArrivalRate)
		}
	}
}

func TestHoltWintersInsufficientData(t *testing.T) {
	config := PlannerConfig{ForecastModel: "holt_winters"}
	forecaster := NewForecaster(config)

	// Provide insufficient data for Holt-Winters (< 24 points)
	history := make([]Metrics, 10)
	baseTime := time.Now().Add(-10 * time.Hour)
	for i := range history {
		history[i] = Metrics{
			Timestamp:   baseTime.Add(time.Duration(i) * time.Hour),
			ArrivalRate: float64(10 + i),
		}
	}

	ctx := context.Background()
	forecasts, err := forecaster.Predict(ctx, history, 2*time.Hour)

	if err != nil {
		t.Fatalf("Predict() should fall back to EWMA: %v", err)
	}

	// Should fall back to EWMA
	if len(forecasts) > 0 && forecasts[0].Model != "EWMA" {
		t.Errorf("Insufficient data should fall back to EWMA, got %v", forecasts[0].Model)
	}
}
