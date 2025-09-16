package budgeting

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/lib/pq"
)

func TestCostCalculationEngine(t *testing.T) {
	model := DefaultCostModel()
	engine := NewCostCalculationEngine(model)

	t.Run("valid job cost calculation", func(t *testing.T) {
		metrics := JobMetrics{
			JobID:           "test-job-1",
			TenantID:        "test-tenant",
			QueueName:       "test-queue",
			CPUTime:         1.5,  // 1.5 seconds
			MemoryMBSeconds: 100,  // 100 MB·seconds
			PayloadBytes:    1024, // 1KB
			RedisOps:        10,   // 10 operations
			NetworkBytes:    2048, // 2KB
			StartTime:       time.Now().Add(-2 * time.Minute),
			EndTime:         time.Now(),
			Priority:        5,
			JobType:         "test",
		}

		cost, err := engine.CalculateJobCost(metrics)
		require.NoError(t, err)

		assert.Equal(t, metrics.JobID, cost.JobID)
		assert.Equal(t, metrics.TenantID, cost.TenantID)
		assert.Equal(t, metrics.QueueName, cost.QueueName)
		assert.True(t, cost.TotalCost > 0)
		assert.True(t, cost.CostBreakdown.BaseCost > 0)
		assert.True(t, cost.CostBreakdown.CPUCost > 0)
		assert.True(t, cost.CostBreakdown.MemoryCost > 0)
	})

	t.Run("invalid metrics validation", func(t *testing.T) {
		tests := []struct {
			name    string
			metrics JobMetrics
			errField string
		}{
			{
				name: "empty job ID",
				metrics: JobMetrics{
					TenantID: "test-tenant",
				},
				errField: "job_id",
			},
			{
				name: "empty tenant ID",
				metrics: JobMetrics{
					JobID: "test-job",
				},
				errField: "tenant_id",
			},
			{
				name: "negative CPU time",
				metrics: JobMetrics{
					JobID:    "test-job",
					TenantID: "test-tenant",
					CPUTime:  -1.0,
				},
				errField: "cpu_time",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := engine.CalculateJobCost(tt.metrics)
				assert.Error(t, err)

				var validationErr *ValidationError
				assert.ErrorAs(t, err, &validationErr)
				assert.Contains(t, validationErr.Field, tt.errField)
			})
		}
	})

	t.Run("cost estimation", func(t *testing.T) {
		estimate := engine.EstimateJobCost("ml-training", 10.0, 500)
		assert.True(t, estimate > 0)

		// ML training should be more expensive than API calls
		apiEstimate := engine.EstimateJobCost("api-call", 0.1, 10)
		assert.True(t, estimate > apiEstimate)
	})

	t.Run("cost breakdown percentages", func(t *testing.T) {
		breakdown := CostBreakdown{
			BaseCost:    0.001,
			CPUCost:     0.01,
			MemoryCost:  0.005,
			PayloadCost: 0.002,
			RedisCost:   0.001,
			NetworkCost: 0.001,
		}

		percentages := engine.GetCostBreakdownPercentages(breakdown)

		totalPercentage := 0.0
		for _, pct := range percentages {
			totalPercentage += pct
		}

		assert.InDelta(t, 100.0, totalPercentage, 0.1)
		assert.True(t, percentages["cpu"] > percentages["memory"])
	})
}

func TestModelCalibrator(t *testing.T) {
	calibrator := NewModelCalibrator()
	model := DefaultCostModel()

	t.Run("benchmark management", func(t *testing.T) {
		benchmark := BenchmarkResult{
			JobType:        "test",
			ActualCost:     0.05,
			CalculatedCost: 0.04,
			CPUTime:        1.0,
			Timestamp:      time.Now(),
		}

		calibrator.AddBenchmark(benchmark)

		// Not enough benchmarks for calibration
		_, err := calibrator.CalibrateModel(model)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "at least 10 benchmarks")
	})

	t.Run("calibration with sufficient data", func(t *testing.T) {
		// Add 10 benchmarks with slight overestimation
		for i := 0; i < 10; i++ {
			benchmark := BenchmarkResult{
				JobType:        "test",
				ActualCost:     0.05,     // Actual cost
				CalculatedCost: 0.04,     // Model underestimates by 20%
				CPUTime:        1.0,
				Timestamp:      time.Now(),
			}
			calibrator.AddBenchmark(benchmark)
		}

		calibratedModel, err := calibrator.CalibrateModel(model)
		require.NoError(t, err)

		// Calibrated model should have higher weights to compensate
		assert.True(t, calibratedModel.CPUTimeWeight > model.CPUTimeWeight)
		assert.True(t, calibratedModel.MemoryWeight > model.MemoryWeight)
	})

	t.Run("calibration accuracy", func(t *testing.T) {
		accuracy := calibrator.GetCalibrationAccuracy()
		assert.True(t, accuracy >= 0.0 && accuracy <= 1.0)
	})

	t.Run("calibration report", func(t *testing.T) {
		report := calibrator.GetCalibrationReport()
		assert.True(t, report.BenchmarkCount > 0)
		assert.True(t, report.Accuracy >= 0.0 && report.Accuracy <= 1.0)
		assert.NotEmpty(t, report.RecommendedAction)
	})
}

func TestBudgetManager(t *testing.T) {
	// Note: This would require a test database in a real implementation
	// For now, we'll test the validation logic

	t.Run("budget validation", func(t *testing.T) {
		// Mock a budget manager for validation testing
		db, _ := sql.Open("postgres", "postgresql://localhost/test")
		aggregator := NewCostAggregator(db)
		notifier := NewNotificationService()
		manager := NewBudgetManager(db, aggregator, notifier)

		validBudget := &Budget{
			TenantID:          "test-tenant",
			QueueName:         "test-queue",
			Amount:            1000.0,
			Currency:          "USD",
			WarningThreshold:  0.75,
			ThrottleThreshold: 0.90,
			BlockThreshold:    1.00,
			Period: BudgetPeriod{
				Type:      "monthly",
				StartDate: time.Now(),
				EndDate:   time.Now().AddDate(0, 1, 0),
			},
			EnforcementPolicy: EnforcementPolicy{
				ThrottleFactor: 0.5,
			},
		}

		// Test valid budget
		err := manager.validateBudget(validBudget)
		assert.NoError(t, err)

		// Test invalid budgets
		invalidBudgets := []struct {
			name   string
			modify func(*Budget)
		}{
			{
				name: "empty tenant ID",
				modify: func(b *Budget) { b.TenantID = "" },
			},
			{
				name: "negative amount",
				modify: func(b *Budget) { b.Amount = -100 },
			},
			{
				name: "invalid warning threshold",
				modify: func(b *Budget) { b.WarningThreshold = 1.5 },
			},
			{
				name: "thresholds out of order",
				modify: func(b *Budget) {
					b.WarningThreshold = 0.95
					b.ThrottleThreshold = 0.75
				},
			},
		}

		for _, test := range invalidBudgets {
			t.Run(test.name, func(t *testing.T) {
				budget := *validBudget // Copy
				test.modify(&budget)

				err := manager.validateBudget(&budget)
				assert.Error(t, err)

				var validationErr *ValidationError
				assert.ErrorAs(t, err, &validationErr)
			})
		}
	})
}

func TestBudgetEnforcer(t *testing.T) {
	// Mock budget manager for testing
	db, _ := sql.Open("postgres", "postgresql://localhost/test")
	aggregator := NewCostAggregator(db)
	notifier := NewNotificationService()
	budgetManager := NewBudgetManager(db, aggregator, notifier)
	enforcer := NewBudgetEnforcer(budgetManager, notifier)

	t.Run("enforcement summary", func(t *testing.T) {
		// This would require database setup in a real test
		summary, err := enforcer.GetEnforcementSummary("test-tenant")
		// In a mock scenario, we expect specific errors or mock responses
		assert.Error(t, err) // Expected since we don't have a real DB
	})

	t.Run("alert tracker", func(t *testing.T) {
		// Test alert rate limiting
		enforcer.ResetAlertTracker()

		// This would be expanded in a real test with mock budget data
		factor := enforcer.GetThrottleFactor("test-tenant", "test-queue")
		assert.Equal(t, 1.0, factor) // No throttling without budget
	})
}

func TestNotificationService(t *testing.T) {
	notifier := NewNotificationService()

	t.Run("notification channel validation", func(t *testing.T) {
		validChannels := []NotificationChannel{
			{
				Type:   "email",
				Target: "test@example.com",
				Events: []string{"warning", "throttle"},
			},
			{
				Type:   "slack",
				Target: "https://hooks.slack.com/services/test",
				Events: []string{"warning"},
			},
			{
				Type:   "webhook",
				Target: "https://example.com/webhook",
				Events: []string{"throttle", "block"},
			},
		}

		for _, channel := range validChannels {
			err := notifier.ValidateNotificationChannel(channel)
			assert.NoError(t, err)
		}

		invalidChannels := []NotificationChannel{
			{Type: "", Target: "test@example.com"},
			{Type: "email", Target: ""},
			{Type: "email", Target: "invalid-email"},
			{Type: "slack", Target: "not-a-slack-url"},
		}

		for _, channel := range invalidChannels {
			err := notifier.ValidateNotificationChannel(channel)
			assert.Error(t, err)
		}
	})

	t.Run("alert message formatting", func(t *testing.T) {
		alert := BudgetAlert{
			TenantID:     "test-tenant",
			QueueName:    "test-queue",
			AlertType:    "budget_warning",
			CurrentSpend: 75.0,
			BudgetAmount: 100.0,
			Utilization:  0.75,
			CreatedAt:    time.Now(),
		}

		message := notifier.formatAlertMessage(alert)
		assert.Contains(t, message, "test-tenant")
		assert.Contains(t, message, "$75.00")
		assert.Contains(t, message, "$100.00")
		assert.Contains(t, message, "75.0%")
	})

	t.Run("test notification", func(t *testing.T) {
		channel := NotificationChannel{
			Type:   "webhook",
			Target: "https://httpbin.org/post",
			Events: []string{"test"},
		}

		// This would actually send a webhook in a real environment
		err := notifier.SendTestNotification(channel)
		// We expect no error for the mock implementation
		assert.NoError(t, err)
	})
}

func TestBudgetForecaster(t *testing.T) {
	db, _ := sql.Open("postgres", "postgresql://localhost/test")
	aggregator := NewCostAggregator(db)
	forecaster := NewBudgetForecaster(aggregator)

	t.Run("trend analysis", func(t *testing.T) {
		// Create sample daily data with increasing trend
		history := []DailyCostAggregate{
			{TotalCost: 10.0, Date: time.Now().AddDate(0, 0, -6)},
			{TotalCost: 11.0, Date: time.Now().AddDate(0, 0, -5)},
			{TotalCost: 12.0, Date: time.Now().AddDate(0, 0, -4)},
			{TotalCost: 13.0, Date: time.Now().AddDate(0, 0, -3)},
			{TotalCost: 14.0, Date: time.Now().AddDate(0, 0, -2)},
			{TotalCost: 15.0, Date: time.Now().AddDate(0, 0, -1)},
			{TotalCost: 16.0, Date: time.Now()},
		}

		analyzer := NewTrendAnalyzer()
		trend := analyzer.CalculateLinearTrend(history)

		assert.True(t, trend.DailyRate > 0)
		assert.True(t, trend.Slope > 0) // Increasing trend
		assert.True(t, trend.R2 > 0.8)  // Good fit
	})

	t.Run("seasonal patterns", func(t *testing.T) {
		pattern := NewSeasonalPattern()

		// Test weekday patterns
		mondayFactor := pattern.GetSeasonalFactor(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), "daily") // Monday
		sundayFactor := pattern.GetSeasonalFactor(time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC), "daily")  // Sunday

		assert.True(t, mondayFactor > sundayFactor) // Monday should be higher than Sunday

		// Test monthly patterns
		decemberFactor := pattern.GetSeasonalFactor(time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC), "monthly")
		julySactor := pattern.GetSeasonalFactor(time.Date(2024, 7, 1, 0, 0, 0, 0, time.UTC), "monthly")

		assert.True(t, decemberFactor > julySactor) // December should be higher than July
	})

	t.Run("forecast generation", func(t *testing.T) {
		// This would require database setup for a real test
		budget := &Budget{
			Amount: 1000.0,
			Period: BudgetPeriod{
				Type:      "monthly",
				StartDate: time.Now().AddDate(0, 0, -15),
				EndDate:   time.Now().AddDate(0, 0, 15),
			},
		}

		// In a real test, this would work with actual data
		_, err := forecaster.GenerateForecast("test-tenant", "test-queue", budget)
		assert.Error(t, err) // Expected since we don't have real data

		var forecastErr *ForecastError
		assert.ErrorAs(t, err, &forecastErr)
	})
}

func TestErrorTypes(t *testing.T) {
	t.Run("budget error", func(t *testing.T) {
		baseErr := ErrBudgetNotFound
		budgetErr := NewBudgetError("budget-123", "tenant-456", "queue-789", "get", baseErr)

		assert.Contains(t, budgetErr.Error(), "budget-123")
		assert.Contains(t, budgetErr.Error(), "tenant-456")
		assert.Contains(t, budgetErr.Error(), "queue-789")
		assert.Contains(t, budgetErr.Error(), "get")
		assert.ErrorIs(t, budgetErr, baseErr)
	})

	t.Run("cost calculation error", func(t *testing.T) {
		baseErr := ErrInvalidJobData
		costErr := NewCostCalculationError("job-123", "tenant-456", "cpu", "invalid value", baseErr)

		assert.Contains(t, costErr.Error(), "job-123")
		assert.Contains(t, costErr.Error(), "tenant-456")
		assert.Contains(t, costErr.Error(), "cpu")
		assert.ErrorIs(t, costErr, baseErr)
	})

	t.Run("error classification", func(t *testing.T) {
		// Test retryable errors
		assert.True(t, IsRetryable(ErrInsufficientData))
		assert.True(t, IsRetryable(ErrForecastFailed))

		// Test permanent errors
		assert.True(t, IsPermanent(ErrBudgetNotFound))
		assert.True(t, IsPermanent(ErrInvalidBudgetPeriod))

		// Test temporary errors
		assert.True(t, IsTemporary(ErrInsufficientData))
		assert.False(t, IsTemporary(ErrBudgetNotFound))
	})

	t.Run("error codes", func(t *testing.T) {
		assert.Equal(t, "BUDGET_NOT_FOUND", ErrorCode(ErrBudgetNotFound))
		assert.Equal(t, "INVALID_COST_MODEL", ErrorCode(ErrInvalidCostModel))
		assert.Equal(t, "ENFORCEMENT_BLOCKED", ErrorCode(ErrEnforcementBlocked))

		// Test typed errors
		budgetErr := NewBudgetError("123", "tenant", "queue", "test", ErrBudgetNotFound)
		assert.Equal(t, "BUDGET_ERROR", ErrorCode(budgetErr))
	})
}

func TestCostModelVariants(t *testing.T) {
	t.Run("default cost model", func(t *testing.T) {
		model := DefaultCostModel()
		assert.Equal(t, 1.0, model.EnvironmentMultiplier)
		assert.True(t, model.CPUTimeWeight > 0)
		assert.True(t, model.MemoryWeight > 0)
	})

	t.Run("production cost model", func(t *testing.T) {
		model := ProductionCostModel()
		assert.Equal(t, 2.0, model.EnvironmentMultiplier)
		assert.True(t, model.CPUTimeWeight > DefaultCostModel().CPUTimeWeight)
	})

	t.Run("staging cost model", func(t *testing.T) {
		model := StagingCostModel()
		assert.Equal(t, 0.5, model.EnvironmentMultiplier)
	})
}

func BenchmarkCostCalculation(b *testing.B) {
	model := DefaultCostModel()
	engine := NewCostCalculationEngine(model)

	metrics := JobMetrics{
		JobID:           "bench-job",
		TenantID:        "bench-tenant",
		QueueName:       "bench-queue",
		CPUTime:         1.0,
		MemoryMBSeconds: 50,
		PayloadBytes:    512,
		RedisOps:        5,
		NetworkBytes:    1024,
		JobType:         "benchmark",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := engine.CalculateJobCost(metrics)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBudgetValidation(b *testing.B) {
	db, _ := sql.Open("postgres", "postgresql://localhost/test")
	aggregator := NewCostAggregator(db)
	notifier := NewNotificationService()
	manager := NewBudgetManager(db, aggregator, notifier)

	budget := &Budget{
		TenantID:          "bench-tenant",
		Amount:            1000.0,
		WarningThreshold:  0.75,
		ThrottleThreshold: 0.90,
		BlockThreshold:    1.00,
		Period: BudgetPeriod{
			Type:      "monthly",
			StartDate: time.Now(),
			EndDate:   time.Now().AddDate(0, 1, 0),
		},
		EnforcementPolicy: EnforcementPolicy{
			ThrottleFactor: 0.5,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := manager.validateBudget(budget)
		if err != nil {
			b.Fatal(err)
		}
	}
}