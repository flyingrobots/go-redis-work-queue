package budgeting

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// BudgetService is the main service that orchestrates job budgeting functionality
type BudgetService struct {
	db            *sql.DB
	costEngine    *CostCalculationEngine
	aggregator    *CostAggregator
	budgetManager *BudgetManager
	enforcer      *BudgetEnforcer
	forecaster    *BudgetForecaster
	notifier      *NotificationService
	tui           *BudgetTUI
}

// Config contains configuration for the budget service
type Config struct {
	DatabaseURL      string        `json:"database_url"`
	FlushInterval    time.Duration `json:"flush_interval"`
	RetentionDays    int           `json:"retention_days"`
	DefaultTenant    string        `json:"default_tenant"`
	CostModel        *CostModel    `json:"cost_model"`
	EnableTUI        bool          `json:"enable_tui"`
	NotificationChannels []NotificationChannel `json:"notification_channels"`
}

// NewBudgetService creates a new budget service instance
func NewBudgetService(config Config) (*BudgetService, error) {
	// Initialize database connection
	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test database connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Initialize cost model
	costModel := config.CostModel
	if costModel == nil {
		costModel = DefaultCostModel()
	}

	// Initialize components
	costEngine := NewCostCalculationEngine(costModel)
	aggregator := NewCostAggregator(db)
	notifier := NewNotificationService()
	budgetManager := NewBudgetManager(db, aggregator, notifier)
	forecaster := NewBudgetForecaster(aggregator)

	service := &BudgetService{
		db:            db,
		costEngine:    costEngine,
		aggregator:    aggregator,
		budgetManager: budgetManager,
		enforcer:      budgetManager.enforcer,
		forecaster:    forecaster,
		notifier:      notifier,
	}

	// Initialize TUI if enabled
	if config.EnableTUI {
		service.tui = NewBudgetTUI(budgetManager, config.DefaultTenant)
	}

	return service, nil
}

// Start starts the budget service background processes
func (s *BudgetService) Start(ctx context.Context) error {
	// Initialize database schema
	if err := s.initializeSchema(); err != nil {
		return fmt.Errorf("failed to initialize database schema: %w", err)
	}

	// Start aggregator
	go s.aggregator.Start(ctx)

	// Start periodic cleanup
	go s.runPeriodicCleanup(ctx)

	return nil
}

// Stop stops the budget service
func (s *BudgetService) Stop() error {
	s.aggregator.Stop()
	return s.db.Close()
}

// ProcessJobCost processes a job's cost metrics
func (s *BudgetService) ProcessJobCost(metrics JobMetrics) error {
	// Calculate job cost
	jobCost, err := s.costEngine.CalculateJobCost(metrics)
	if err != nil {
		return fmt.Errorf("failed to calculate job cost: %w", err)
	}

	// Add to aggregation pipeline
	if err := s.aggregator.AddJobCost(jobCost); err != nil {
		return fmt.Errorf("failed to add job cost to aggregator: %w", err)
	}

	return nil
}

// CheckBudgetCompliance checks if a job is compliant with budget limits
func (s *BudgetService) CheckBudgetCompliance(tenantID, queueName string, jobPriority int) (EnforcementAction, error) {
	return s.enforcer.CheckJobAllowed(tenantID, queueName, jobPriority)
}

// CreateBudget creates a new budget
func (s *BudgetService) CreateBudget(budget *Budget) error {
	return s.budgetManager.CreateBudget(budget)
}

// GetBudgetStatus returns the current status of a budget
func (s *BudgetService) GetBudgetStatus(budgetID string) (*BudgetStatus, error) {
	return s.budgetManager.GetBudgetStatus(budgetID)
}

// GetForecast generates a spending forecast for a tenant
func (s *BudgetService) GetForecast(tenantID, queueName string) (*Forecast, error) {
	budget, err := s.budgetManager.GetBudgetForTenant(tenantID, queueName)
	if err != nil {
		return nil, err
	}

	return s.forecaster.GenerateForecast(tenantID, queueName, budget)
}

// GenerateReport generates a comprehensive budget report
func (s *BudgetService) GenerateReport(tenantID string, period BudgetPeriod) (*BudgetReport, error) {
	// Get total spending
	totalSpend, err := s.aggregator.GetCurrentSpend(tenantID, "", period)
	if err != nil {
		return nil, err
	}

	// Get budget information
	budget, err := s.budgetManager.GetBudgetForTenant(tenantID, "")
	if err != nil {
		return nil, err
	}

	// Get top cost drivers
	drivers, err := s.aggregator.GetTopCostDrivers(tenantID, period.StartDate, period.EndDate, 10)
	if err != nil {
		return nil, err
	}

	// Get daily breakdown
	dailyBreakdown, err := s.aggregator.GetDailySpend(tenantID, "", period.StartDate, period.EndDate)
	if err != nil {
		return nil, err
	}

	// Get queue breakdown
	queueBreakdown, err := s.aggregator.GetQueueBreakdown(tenantID, period.StartDate, period.EndDate)
	if err != nil {
		return nil, err
	}

	// Get violations
	violations, err := s.enforcer.GetBudgetViolations(tenantID)
	if err != nil {
		return nil, err
	}

	// Convert violations to alerts
	var alerts []BudgetAlert
	for _, violation := range violations {
		alert := BudgetAlert{
			BudgetID:     violation.BudgetID,
			TenantID:     violation.TenantID,
			QueueName:    violation.QueueName,
			AlertType:    violation.ViolationType,
			CurrentSpend: violation.CurrentSpend,
			BudgetAmount: violation.BudgetAmount,
			Utilization:  violation.Utilization,
			CreatedAt:    violation.DetectedAt,
		}
		alerts = append(alerts, alert)
	}

	utilization := 0.0
	if budget.Amount > 0 {
		utilization = totalSpend / budget.Amount
	}

	report := &BudgetReport{
		TenantID:        tenantID,
		Period:          period,
		TotalSpend:      totalSpend,
		BudgetAmount:    budget.Amount,
		Utilization:     utilization,
		TopDrivers:      drivers,
		DailyBreakdown:  dailyBreakdown,
		QueueBreakdown:  queueBreakdown,
		Alerts:          alerts,
		Recommendations: s.generateRecommendations(utilization, drivers),
		GeneratedAt:     time.Now(),
		GeneratedBy:     "budget-service",
	}

	return report, nil
}

// StartTUI starts the terminal user interface
func (s *BudgetService) StartTUI() error {
	if s.tui == nil {
		return fmt.Errorf("TUI not enabled in configuration")
	}
	return s.tui.Start()
}

// EstimateJobCost provides a quick cost estimate for planning
func (s *BudgetService) EstimateJobCost(jobType string, estimatedCPUTime float64, estimatedPayloadKB int) float64 {
	return s.costEngine.EstimateJobCost(jobType, estimatedCPUTime, estimatedPayloadKB)
}

// UpdateCostModel updates the cost calculation model
func (s *BudgetService) UpdateCostModel(newModel *CostModel) error {
	return s.costEngine.UpdateModel(newModel)
}

// GetCostBreakdown returns cost breakdown for a specific tenant/queue/period
func (s *BudgetService) GetCostBreakdown(tenantID, queueName string, startDate, endDate time.Time) (map[string]float64, error) {
	return s.aggregator.GetComponentBreakdown(tenantID, queueName, startDate, endDate)
}

// GetTopCostDrivers returns the top cost drivers for a tenant
func (s *BudgetService) GetTopCostDrivers(tenantID string, startDate, endDate time.Time, limit int) ([]CostDriver, error) {
	return s.aggregator.GetTopCostDrivers(tenantID, startDate, endDate, limit)
}

// initializeSchema creates the necessary database tables
func (s *BudgetService) initializeSchema() error {
	schema := `
		-- Daily cost aggregates table
		CREATE TABLE IF NOT EXISTS daily_costs (
			tenant_id VARCHAR(50) NOT NULL,
			queue_name VARCHAR(100) NOT NULL,
			date DATE NOT NULL,
			total_jobs INTEGER NOT NULL DEFAULT 0,
			total_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
			cpu_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
			memory_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
			payload_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
			redis_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
			network_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
			avg_job_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
			max_job_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
			min_job_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
			p95_job_cost DECIMAL(10,4) NOT NULL DEFAULT 0,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (tenant_id, queue_name, date)
		);

		-- Budget definitions table
		CREATE TABLE IF NOT EXISTS budgets (
			id UUID PRIMARY KEY,
			tenant_id VARCHAR(50) NOT NULL,
			queue_name VARCHAR(100) NOT NULL DEFAULT '',
			period_type VARCHAR(20) NOT NULL,
			period_start DATE NOT NULL,
			period_end DATE NOT NULL,
			amount DECIMAL(10,2) NOT NULL,
			currency VARCHAR(3) NOT NULL DEFAULT 'USD',
			warning_threshold DECIMAL(3,2) NOT NULL DEFAULT 0.75,
			throttle_threshold DECIMAL(3,2) NOT NULL DEFAULT 0.90,
			block_threshold DECIMAL(3,2) NOT NULL DEFAULT 1.00,
			enforcement_policy JSONB NOT NULL DEFAULT '{}',
			notifications JSONB NOT NULL DEFAULT '[]',
			tags JSONB NOT NULL DEFAULT '{}',
			active BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			created_by VARCHAR(100) NOT NULL DEFAULT 'system'
		);

		-- Indexes for fast lookups
		CREATE INDEX IF NOT EXISTS idx_daily_costs_tenant_date ON daily_costs (tenant_id, date DESC);
		CREATE INDEX IF NOT EXISTS idx_daily_costs_queue_date ON daily_costs (tenant_id, queue_name, date DESC);
		CREATE INDEX IF NOT EXISTS idx_budgets_tenant_period ON budgets (tenant_id, period_start, period_end);
		CREATE INDEX IF NOT EXISTS idx_budgets_active ON budgets (tenant_id, active) WHERE active = true;
	`

	_, err := s.db.Exec(schema)
	return err
}

// runPeriodicCleanup runs periodic data cleanup tasks
func (s *BudgetService) runPeriodicCleanup(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Clean up old cost data (default 365 days retention)
			if err := s.aggregator.PurgeOldData(365); err != nil {
				fmt.Printf("Error during data cleanup: %v\n", err)
			}
		}
	}
}

// generateRecommendations generates budget optimization recommendations
func (s *BudgetService) generateRecommendations(utilization float64, drivers []CostDriver) []string {
	var recommendations []string

	if utilization > 0.9 {
		recommendations = append(recommendations, "Budget utilization is very high - consider increasing budget or optimizing costs")
	} else if utilization > 0.75 {
		recommendations = append(recommendations, "Budget utilization is elevated - monitor closely and optimize where possible")
	}

	if len(drivers) > 0 {
		topDriver := drivers[0]
		if topDriver.Percentage > 50 {
			recommendations = append(recommendations,
				fmt.Sprintf("Queue '%s' accounts for %.1f%% of costs - focus optimization efforts here",
					topDriver.QueueName, topDriver.Percentage))
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Budget is healthy - continue monitoring spending patterns")
	}

	return recommendations
}

// GetMetrics returns service-level metrics for monitoring
func (s *BudgetService) GetMetrics() map[string]interface{} {
	return map[string]interface{}{
		"cost_calculations_total": "tracked by cost engine",
		"budget_checks_total":     "tracked by enforcer",
		"aggregation_errors":      "tracked by aggregator",
		"active_budgets":          "tracked by budget manager",
		"last_cleanup":            "tracked by cleanup process",
	}
}