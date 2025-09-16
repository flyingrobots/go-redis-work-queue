package budgeting

import (
	"math"
	"time"
)

// BudgetForecaster generates spending forecasts and predictions
type BudgetForecaster struct {
	aggregator      *CostAggregator
	seasonalPattern *SeasonalPattern
	trendAnalyzer   *TrendAnalyzer
}

// NewBudgetForecaster creates a new budget forecaster
func NewBudgetForecaster(aggregator *CostAggregator) *BudgetForecaster {
	return &BudgetForecaster{
		aggregator:      aggregator,
		seasonalPattern: NewSeasonalPattern(),
		trendAnalyzer:   NewTrendAnalyzer(),
	}
}

// GenerateForecast generates a spending forecast for a tenant/queue
func (f *BudgetForecaster) GenerateForecast(tenantID, queueName string, budget *Budget) (*Forecast, error) {
	// Get historical data (last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	history, err := f.aggregator.GetDailySpend(tenantID, queueName, startDate, endDate)
	if err != nil {
		return nil, NewForecastError(tenantID, queueName, budget.Period.Type, "failed to get historical data", err)
	}

	if len(history) < 7 {
		return nil, NewForecastError(tenantID, queueName, budget.Period.Type, "insufficient historical data", ErrInsufficientData)
	}

	// Calculate linear trend
	trend := f.trendAnalyzer.CalculateLinearTrend(history)

	// Apply seasonal adjustments
	seasonalFactor := f.seasonalPattern.GetSeasonalFactor(time.Now(), budget.Period.Type)

	// Project to period end
	now := time.Now()
	daysRemaining := int(budget.Period.EndDate.Sub(now).Hours() / 24)
	if daysRemaining <= 0 {
		daysRemaining = 1 // Avoid division by zero
	}

	projectedSpend := trend.DailyRate * float64(daysRemaining) * seasonalFactor

	// Get current spend
	currentSpend, err := f.aggregator.GetCurrentSpend(tenantID, queueName, budget.Period)
	if err != nil {
		return nil, err
	}

	totalProjectedSpend := currentSpend + projectedSpend

	// Calculate confidence interval based on historical variance
	variance := f.calculateVariance(history, trend.DailyRate)
	confidenceInterval := 1.96 * math.Sqrt(variance/float64(len(history))) // 95% CI

	forecast := &Forecast{
		TenantID:           tenantID,
		QueueName:          queueName,
		PeriodEnd:          budget.Period.EndDate,
		PredictedSpend:     totalProjectedSpend,
		ConfidenceInterval: confidenceInterval,
		BudgetUtilization:  totalProjectedSpend / budget.Amount,
		TrendDirection:     f.getTrendDirection(trend.Slope),
		SeasonalFactor:     seasonalFactor,
		GeneratedAt:        time.Now(),
	}

	// Calculate days until overrun
	if totalProjectedSpend > budget.Amount && trend.DailyRate > 0 {
		remainingBudget := budget.Amount - currentSpend
		daysUntilOverrun := int(remainingBudget / trend.DailyRate)
		if daysUntilOverrun > 0 {
			forecast.DaysUntilOverrun = &daysUntilOverrun
		}
	}

	// Generate recommendations
	forecast.Recommendation = f.generateRecommendation(forecast, budget, trend)

	return forecast, nil
}

// TrendAnalyzer analyzes spending trends
type TrendAnalyzer struct{}

// NewTrendAnalyzer creates a new trend analyzer
func NewTrendAnalyzer() *TrendAnalyzer {
	return &TrendAnalyzer{}
}

// LinearTrend represents a linear trend analysis
type LinearTrend struct {
	DailyRate float64 `json:"daily_rate"`
	Slope     float64 `json:"slope"`
	R2        float64 `json:"r2"`        // Coefficient of determination
	Intercept float64 `json:"intercept"`
}

// CalculateLinearTrend calculates linear trend from daily cost data
func (t *TrendAnalyzer) CalculateLinearTrend(history []DailyCostAggregate) LinearTrend {
	if len(history) == 0 {
		return LinearTrend{}
	}

	n := float64(len(history))
	var sumX, sumY, sumXY, sumX2 float64

	for i, data := range history {
		x := float64(i) // Day index
		y := data.TotalCost

		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope and intercept using least squares
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / n

	// Calculate R-squared
	meanY := sumY / n
	var ssRes, ssTot float64

	for i, data := range history {
		x := float64(i)
		y := data.TotalCost
		predicted := slope*x + intercept

		ssRes += (y - predicted) * (y - predicted)
		ssTot += (y - meanY) * (y - meanY)
	}

	r2 := 0.0
	if ssTot != 0 {
		r2 = 1 - (ssRes / ssTot)
	}

	// Calculate average daily rate from recent data
	recentDays := int(math.Min(7, float64(len(history))))
	totalRecent := 0.0
	for i := len(history) - recentDays; i < len(history); i++ {
		totalRecent += history[i].TotalCost
	}
	dailyRate := totalRecent / float64(recentDays)

	return LinearTrend{
		DailyRate: dailyRate,
		Slope:     slope,
		R2:        r2,
		Intercept: intercept,
	}
}

// SeasonalPattern handles seasonal adjustments for forecasting
type SeasonalPattern struct {
	weekdayFactors map[time.Weekday]float64
	monthFactors   map[time.Month]float64
}

// NewSeasonalPattern creates a new seasonal pattern analyzer
func NewSeasonalPattern() *SeasonalPattern {
	return &SeasonalPattern{
		// Default weekday patterns (assuming business workloads)
		weekdayFactors: map[time.Weekday]float64{
			time.Sunday:    0.3, // Low weekend activity
			time.Monday:    1.2, // Monday surge
			time.Tuesday:   1.1, // High weekday activity
			time.Wednesday: 1.1, // High weekday activity
			time.Thursday:  1.1, // High weekday activity
			time.Friday:    1.0, // Normal Friday
			time.Saturday:  0.4, // Low weekend activity
		},
		// Default monthly patterns
		monthFactors: map[time.Month]float64{
			time.January:   0.9, // Post-holiday slowdown
			time.February:  1.0, // Normal
			time.March:     1.1, // Q1 end push
			time.April:     1.0, // Normal
			time.May:       1.0, // Normal
			time.June:      1.1, // Q2 end push
			time.July:      0.9, // Summer slowdown
			time.August:    0.9, // Summer slowdown
			time.September: 1.1, // Back to work surge
			time.October:   1.0, // Normal
			time.November:  1.0, // Normal
			time.December:  1.2, // Year-end push
		},
	}
}

// GetSeasonalFactor returns the seasonal adjustment factor for a given time and period
func (s *SeasonalPattern) GetSeasonalFactor(t time.Time, periodType string) float64 {
	switch periodType {
	case "daily":
		return s.weekdayFactors[t.Weekday()]
	case "weekly":
		// For weekly periods, use a blend of weekday patterns
		return 1.0 // Simplified for weekly
	case "monthly":
		return s.monthFactors[t.Month()]
	default:
		return 1.0 // No adjustment for unknown periods
	}
}

// calculateVariance calculates the variance in daily spending around the trend
func (f *BudgetForecaster) calculateVariance(history []DailyCostAggregate, expectedDaily float64) float64 {
	if len(history) == 0 {
		return 0.0
	}

	var variance float64
	for _, data := range history {
		diff := data.TotalCost - expectedDaily
		variance += diff * diff
	}

	return variance / float64(len(history))
}

// getTrendDirection returns a human-readable trend direction
func (f *BudgetForecaster) getTrendDirection(slope float64) string {
	switch {
	case slope > 0.1:
		return "increasing"
	case slope < -0.1:
		return "decreasing"
	default:
		return "stable"
	}
}

// generateRecommendation generates forecast-based recommendations
func (f *BudgetForecaster) generateRecommendation(forecast *Forecast, budget *Budget, trend LinearTrend) string {
	switch {
	case forecast.BudgetUtilization > 1.2:
		return "Projected to exceed budget by >20% - immediate cost optimization or budget increase required"

	case forecast.BudgetUtilization > 1.1:
		return "Projected to exceed budget by >10% - cost optimization recommended or consider budget adjustment"

	case forecast.BudgetUtilization > 1.0:
		return "Projected to slightly exceed budget - monitor closely and optimize where possible"

	case forecast.BudgetUtilization > 0.9:
		return "On track to use 90%+ of budget - good utilization, monitor for unexpected spikes"

	case forecast.BudgetUtilization < 0.5:
		return "Projected to use <50% of budget - consider reducing budget or increasing workload"

	default:
		if trend.Slope > 0.1 {
			return "Spending is increasing - monitor trend and be prepared for budget adjustments"
		}
		return "Spending forecast looks healthy - continue monitoring"
	}
}

// GetForecastAccuracy compares previous forecasts with actual spending
func (f *BudgetForecaster) GetForecastAccuracy(tenantID, queueName string, forecastPeriod time.Duration) (float64, error) {
	// This would compare historical forecasts with actual outcomes
	// For now, return a placeholder accuracy value
	return 0.85, nil // 85% accuracy
}

// GenerateMultipleForecast generates forecasts for multiple scenarios
func (f *BudgetForecaster) GenerateMultipleForecast(tenantID, queueName string, budget *Budget) (*MultipleForecast, error) {
	baseForecast, err := f.GenerateForecast(tenantID, queueName, budget)
	if err != nil {
		return nil, err
	}

	// Generate optimistic scenario (20% lower spending)
	optimisticForecast := *baseForecast
	optimisticForecast.PredictedSpend *= 0.8
	optimisticForecast.BudgetUtilization = optimisticForecast.PredictedSpend / budget.Amount

	// Generate pessimistic scenario (20% higher spending)
	pessimisticForecast := *baseForecast
	pessimisticForecast.PredictedSpend *= 1.2
	pessimisticForecast.BudgetUtilization = pessimisticForecast.PredictedSpend / budget.Amount

	return &MultipleForecast{
		Base:        baseForecast,
		Optimistic:  &optimisticForecast,
		Pessimistic: &pessimisticForecast,
		GeneratedAt: time.Now(),
	}, nil
}

// MultipleForecast contains multiple forecast scenarios
type MultipleForecast struct {
	Base        *Forecast `json:"base"`
	Optimistic  *Forecast `json:"optimistic"`
	Pessimistic *Forecast `json:"pessimistic"`
	GeneratedAt time.Time `json:"generated_at"`
}