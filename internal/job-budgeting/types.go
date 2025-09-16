package budgeting

import (
	"time"
)

// CostModel defines the weights and rates for calculating job costs
type CostModel struct {
	CPUTimeWeight         float64 `json:"cpu_time_weight"`     // $/second
	MemoryWeight          float64 `json:"memory_weight"`       // $/MB·second
	PayloadWeight         float64 `json:"payload_weight"`      // $/KB
	RedisOpsWeight        float64 `json:"redis_ops_weight"`    // $/operation
	NetworkWeight         float64 `json:"network_weight"`      // $/MB transferred
	BaseJobWeight         float64 `json:"base_job_weight"`     // Fixed cost per job
	EnvironmentMultiplier float64 `json:"env_multiplier"`      // Production vs staging
}

// JobMetrics contains the raw resource usage data for a job
type JobMetrics struct {
	JobID           string    `json:"job_id"`
	TenantID        string    `json:"tenant_id"`
	QueueName       string    `json:"queue_name"`
	CPUTime         float64   `json:"cpu_time_seconds"`
	MemoryMBSeconds float64   `json:"memory_mb_seconds"`
	PayloadBytes    int       `json:"payload_bytes"`
	RedisOps        int       `json:"redis_operations"`
	NetworkBytes    int       `json:"network_bytes"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Priority        int       `json:"priority"`
	JobType         string    `json:"job_type"`
}

// JobCost represents the calculated cost of a single job
type JobCost struct {
	JobID         string    `json:"job_id"`
	TenantID      string    `json:"tenant_id"`
	QueueName     string    `json:"queue_name"`
	CPUTime       float64   `json:"cpu_time_seconds"`
	MemorySeconds float64   `json:"memory_mb_seconds"`
	PayloadSize   int       `json:"payload_bytes"`
	RedisOps      int       `json:"redis_operations"`
	NetworkBytes  int       `json:"network_bytes"`
	TotalCost     float64   `json:"total_cost"`
	CostBreakdown CostBreakdown `json:"cost_breakdown"`
	Timestamp     time.Time `json:"timestamp"`
	JobType       string    `json:"job_type"`
	Priority      int       `json:"priority"`
}

// CostBreakdown shows the individual components of job cost
type CostBreakdown struct {
	BaseCost    float64 `json:"base_cost"`
	CPUCost     float64 `json:"cpu_cost"`
	MemoryCost  float64 `json:"memory_cost"`
	PayloadCost float64 `json:"payload_cost"`
	RedisCost   float64 `json:"redis_cost"`
	NetworkCost float64 `json:"network_cost"`
}

// DailyCostAggregate represents aggregated cost data for a tenant/queue/day
type DailyCostAggregate struct {
	TenantID     string    `json:"tenant_id"`
	QueueName    string    `json:"queue_name"`
	Date         time.Time `json:"date"`
	TotalJobs    int       `json:"total_jobs"`
	TotalCost    float64   `json:"total_cost"`
	CPUCost      float64   `json:"cpu_cost"`
	MemoryCost   float64   `json:"memory_cost"`
	PayloadCost  float64   `json:"payload_cost"`
	RedisCost    float64   `json:"redis_cost"`
	NetworkCost  float64   `json:"network_cost"`
	AvgJobCost   float64   `json:"avg_job_cost"`
	MaxJobCost   float64   `json:"max_job_cost"`
	MinJobCost   float64   `json:"min_job_cost"`
	P95JobCost   float64   `json:"p95_job_cost"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Budget defines spending limits and enforcement policies
type Budget struct {
	ID                string                 `json:"id"`
	TenantID          string                 `json:"tenant_id"`
	QueueName         string                 `json:"queue_name,omitempty"` // Empty = tenant budget
	Period            BudgetPeriod           `json:"period"`
	Amount            float64                `json:"amount"`
	Currency          string                 `json:"currency"`
	WarningThreshold  float64                `json:"warning_threshold"`   // 0.75 = 75%
	ThrottleThreshold float64                `json:"throttle_threshold"`  // 0.90 = 90%
	BlockThreshold    float64                `json:"block_threshold"`     // 1.00 = 100%
	EnforcementPolicy EnforcementPolicy      `json:"enforcement_policy"`
	Notifications     []NotificationChannel  `json:"notifications"`
	Tags              map[string]string      `json:"tags"`
	Active            bool                   `json:"active"`
	CreatedAt         time.Time              `json:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at"`
	CreatedBy         string                 `json:"created_by"`
}

// BudgetPeriod defines the time window for budget enforcement
type BudgetPeriod struct {
	Type      string    `json:"type"`       // "monthly", "weekly", "daily", "custom"
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Timezone  string    `json:"timezone"`
}

// EnforcementPolicy controls how budget violations are handled
type EnforcementPolicy struct {
	WarnOnly        bool    `json:"warn_only"`
	ThrottleFactor  float64 `json:"throttle_factor"`  // 0.5 = 50% slower
	BlockNewJobs    bool    `json:"block_new_jobs"`
	AllowEmergency  bool    `json:"allow_emergency"`   // High priority bypass
	GracePeriodHours int    `json:"grace_period_hours"` // Hours before enforcement
}

// NotificationChannel defines how budget alerts are delivered
type NotificationChannel struct {
	Type       string            `json:"type"`       // "email", "slack", "webhook", "pagerduty"
	Target     string            `json:"target"`     // email address, webhook URL, etc.
	Events     []string          `json:"events"`     // "warning", "throttle", "block", "reset"
	Enabled    bool              `json:"enabled"`
	Metadata   map[string]string `json:"metadata"`   // Additional channel-specific config
}

// BudgetStatus represents the current state of a budget
type BudgetStatus struct {
	BudgetID          string    `json:"budget_id"`
	TenantID          string    `json:"tenant_id"`
	QueueName         string    `json:"queue_name,omitempty"`
	CurrentSpend      float64   `json:"current_spend"`
	BudgetAmount      float64   `json:"budget_amount"`
	Utilization       float64   `json:"utilization"`       // Current spend / budget
	DaysInPeriod      int       `json:"days_in_period"`
	DaysRemaining     int       `json:"days_remaining"`
	DailyBurnRate     float64   `json:"daily_burn_rate"`
	ProjectedSpend    float64   `json:"projected_spend"`
	IsOverBudget      bool      `json:"is_over_budget"`
	CurrentThreshold  string    `json:"current_threshold"` // "none", "warning", "throttle", "block"
	LastAlertSent     *time.Time `json:"last_alert_sent,omitempty"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Forecast represents budget spending predictions
type Forecast struct {
	TenantID           string     `json:"tenant_id"`
	QueueName          string     `json:"queue_name,omitempty"`
	PeriodEnd          time.Time  `json:"period_end"`
	PredictedSpend     float64    `json:"predicted_spend"`
	ConfidenceInterval float64    `json:"confidence_interval"`
	BudgetUtilization  float64    `json:"budget_utilization"`
	DaysUntilOverrun   *int       `json:"days_until_overrun,omitempty"`
	Recommendation     string     `json:"recommendation"`
	TrendDirection     string     `json:"trend_direction"` // "increasing", "decreasing", "stable"
	SeasonalFactor     float64    `json:"seasonal_factor"`
	GeneratedAt        time.Time  `json:"generated_at"`
}

// EnforcementAction represents the result of budget enforcement check
type EnforcementAction struct {
	Type         string  `json:"type"`         // "allow", "warn", "throttle", "block"
	Factor       float64 `json:"factor"`       // For throttle actions
	Message      string  `json:"message"`
	BlockReason  string  `json:"block_reason,omitempty"`
	BypassAllowed bool   `json:"bypass_allowed"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
}

// CostDriver represents a significant contributor to costs
type CostDriver struct {
	TenantID      string  `json:"tenant_id"`
	QueueName     string  `json:"queue_name"`
	JobType       string  `json:"job_type"`
	Component     string  `json:"component"`     // "cpu", "memory", "payload", "redis", "network"
	TotalCost     float64 `json:"total_cost"`
	JobCount      int     `json:"job_count"`
	AvgCostPerJob float64 `json:"avg_cost_per_job"`
	Percentage    float64 `json:"percentage"`    // % of total spend
	Trend         string  `json:"trend"`        // "increasing", "decreasing", "stable"
}

// BudgetAlert represents a budget threshold violation
type BudgetAlert struct {
	ID          string    `json:"id"`
	BudgetID    string    `json:"budget_id"`
	TenantID    string    `json:"tenant_id"`
	QueueName   string    `json:"queue_name,omitempty"`
	AlertType   string    `json:"alert_type"`   // "warning", "throttle", "block", "reset"
	Message     string    `json:"message"`
	CurrentSpend float64  `json:"current_spend"`
	BudgetAmount float64  `json:"budget_amount"`
	Utilization float64   `json:"utilization"`
	Acknowledged bool     `json:"acknowledged"`
	CreatedAt   time.Time `json:"created_at"`
	AckedAt     *time.Time `json:"acked_at,omitempty"`
	AckedBy     string    `json:"acked_by,omitempty"`
}

// BudgetReport represents a comprehensive budget analysis
type BudgetReport struct {
	TenantID        string               `json:"tenant_id"`
	Period          BudgetPeriod         `json:"period"`
	TotalSpend      float64              `json:"total_spend"`
	BudgetAmount    float64              `json:"budget_amount"`
	Utilization     float64              `json:"utilization"`
	TopDrivers      []CostDriver         `json:"top_drivers"`
	DailyBreakdown  []DailyCostAggregate `json:"daily_breakdown"`
	QueueBreakdown  map[string]float64   `json:"queue_breakdown"`
	Alerts          []BudgetAlert        `json:"alerts"`
	Recommendations []string             `json:"recommendations"`
	GeneratedAt     time.Time            `json:"generated_at"`
	GeneratedBy     string               `json:"generated_by"`
}

// DefaultCostModel returns a standard cost model with realistic pricing
func DefaultCostModel() *CostModel {
	return &CostModel{
		CPUTimeWeight:         0.0001,  // $0.0001 per CPU second
		MemoryWeight:          0.00001, // $0.00001 per MB·second
		PayloadWeight:         0.00002, // $0.00002 per KB
		RedisOpsWeight:        0.000001, // $0.000001 per operation
		NetworkWeight:         0.01,    // $0.01 per MB transferred
		BaseJobWeight:         0.001,   // $0.001 base cost per job
		EnvironmentMultiplier: 1.0,     // No multiplier by default
	}
}

// ProductionCostModel returns a cost model optimized for production environments
func ProductionCostModel() *CostModel {
	model := DefaultCostModel()
	model.EnvironmentMultiplier = 2.0 // 2x cost for production
	model.CPUTimeWeight = 0.0002
	model.MemoryWeight = 0.00002
	return model
}

// StagingCostModel returns a cost model for staging environments
func StagingCostModel() *CostModel {
	model := DefaultCostModel()
	model.EnvironmentMultiplier = 0.5 // Half cost for staging
	return model
}