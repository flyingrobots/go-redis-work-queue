// Copyright 2025 James Ross
package forecasting

import (
	"sync"
	"time"
)

// SLOTracker tracks SLO budget consumption
type SLOTracker struct {
	target         float64 // SLO target (e.g., 0.999 for 99.9%)
	weeklyBudget   float64 // Weekly error budget
	monthlyBudget  float64 // Monthly error budget

	// Historical data
	weeklyErrors  []float64
	monthlyErrors []float64

	// Current state
	currentWeekBurn  float64
	currentMonthBurn float64
	lastUpdated      time.Time

	mu sync.RWMutex
}

// NewSLOTracker creates a new SLO tracker
func NewSLOTracker() *SLOTracker {
	return &SLOTracker{
		target:        0.999, // 99.9% availability
		weeklyBudget:  1 - 0.999,
		monthlyBudget: 1 - 0.999,
		weeklyErrors:  make([]float64, 0, 7*24*60),   // 7 days of minute data
		monthlyErrors: make([]float64, 0, 30*24*60), // 30 days of minute data
		lastUpdated:   time.Now(),
	}
}

// SetTarget sets the SLO target
func (st *SLOTracker) SetTarget(target float64) {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.target = target
	st.weeklyBudget = 1 - target
	st.monthlyBudget = 1 - target
}

// Update updates the tracker with current error rate
func (st *SLOTracker) Update(errorRate float64) {
	st.mu.Lock()
	defer st.mu.Unlock()

	now := time.Now()

	// Add to historical data
	st.weeklyErrors = append(st.weeklyErrors, errorRate)
	st.monthlyErrors = append(st.monthlyErrors, errorRate)

	// Trim old data
	weekCutoff := 7 * 24 * 60 // 7 days in minutes
	if len(st.weeklyErrors) > weekCutoff {
		st.weeklyErrors = st.weeklyErrors[len(st.weeklyErrors)-weekCutoff:]
	}

	monthCutoff := 30 * 24 * 60 // 30 days in minutes
	if len(st.monthlyErrors) > monthCutoff {
		st.monthlyErrors = st.monthlyErrors[len(st.monthlyErrors)-monthCutoff:]
	}

	// Calculate current burn rates
	st.currentWeekBurn = st.calculateBurnRate(st.weeklyErrors)
	st.currentMonthBurn = st.calculateBurnRate(st.monthlyErrors)
	st.lastUpdated = now
}

// ProjectBudgetBurn projects budget consumption based on forecast
func (st *SLOTracker) ProjectBudgetBurn(errorForecast []float64) *SLOBudget {
	st.mu.RLock()
	defer st.mu.RUnlock()

	// Calculate projected burn
	projectedErrors := append(st.weeklyErrors, errorForecast...)
	projectedBurn := st.calculateBurnRate(projectedErrors)

	// Calculate time to exhaustion
	remainingBudget := st.weeklyBudget - st.currentWeekBurn
	if remainingBudget <= 0 {
		remainingBudget = 0
	}

	var timeToExhaustion time.Duration
	if projectedBurn > 0 && remainingBudget > 0 {
		// Calculate how long until budget exhausted at current rate
		minutesRemaining := (remainingBudget / projectedBurn) * float64(len(st.weeklyErrors))
		timeToExhaustion = time.Duration(minutesRemaining) * time.Minute
	}

	return &SLOBudget{
		Target:           st.target,
		CurrentBurn:      st.currentWeekBurn,
		WeeklyBurnRate:   st.currentWeekBurn / st.weeklyBudget,
		MonthlyBurnRate:  st.currentMonthBurn / st.monthlyBudget,
		RemainingBudget:  remainingBudget,
		ProjectedBurn:    projectedBurn,
		TimeToExhaustion: timeToExhaustion,
		LastUpdated:      st.lastUpdated,
	}
}

// GetCurrentBudget returns current budget status
func (st *SLOTracker) GetCurrentBudget() *SLOBudget {
	st.mu.RLock()
	defer st.mu.RUnlock()

	remainingWeekly := st.weeklyBudget - st.currentWeekBurn
	if remainingWeekly < 0 {
		remainingWeekly = 0
	}

	return &SLOBudget{
		Target:           st.target,
		CurrentBurn:      st.currentWeekBurn,
		WeeklyBurnRate:   st.currentWeekBurn / st.weeklyBudget,
		MonthlyBurnRate:  st.currentMonthBurn / st.monthlyBudget,
		RemainingBudget:  remainingWeekly,
		ProjectedBurn:    st.currentWeekBurn,
		TimeToExhaustion: 0,
		LastUpdated:      st.lastUpdated,
	}
}

// calculateBurnRate calculates the error budget burn rate
func (st *SLOTracker) calculateBurnRate(errors []float64) float64 {
	if len(errors) == 0 {
		return 0
	}

	sum := 0.0
	for _, e := range errors {
		sum += e
	}

	return sum / float64(len(errors))
}

// Reset resets the SLO tracker
func (st *SLOTracker) Reset() {
	st.mu.Lock()
	defer st.mu.Unlock()

	st.weeklyErrors = make([]float64, 0, 7*24*60)
	st.monthlyErrors = make([]float64, 0, 30*24*60)
	st.currentWeekBurn = 0
	st.currentMonthBurn = 0
	st.lastUpdated = time.Now()
}
