package budgeting

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"sync"
	"time"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// CostAggregator handles aggregation of job costs into daily summaries
type CostAggregator struct {
	db           *sql.DB
	buffer       map[string][]JobCost // tenant:queue -> costs
	bufferMutex  sync.RWMutex
	flushSize    int
	flushInterval time.Duration
	stopChan     chan struct{}
}

// NewCostAggregator creates a new cost aggregator
func NewCostAggregator(db *sql.DB) *CostAggregator {
	return &CostAggregator{
		db:            db,
		buffer:        make(map[string][]JobCost),
		flushSize:     1000,
		flushInterval: 5 * time.Minute,
		stopChan:      make(chan struct{}),
	}
}

// Start begins the aggregation background process
func (a *CostAggregator) Start(ctx context.Context) {
	ticker := time.NewTicker(a.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.flush() // Final flush before shutdown
			return
		case <-a.stopChan:
			a.flush()
			return
		case <-ticker.C:
			a.flush()
		}
	}
}

// Stop stops the aggregation process
func (a *CostAggregator) Stop() {
	close(a.stopChan)
}

// AddJobCost adds a job cost to the aggregation buffer
func (a *CostAggregator) AddJobCost(cost JobCost) error {
	key := fmt.Sprintf("%s:%s", cost.TenantID, cost.QueueName)

	a.bufferMutex.Lock()
	defer a.bufferMutex.Unlock()

	a.buffer[key] = append(a.buffer[key], cost)

	// Check if we need to flush early
	if len(a.buffer[key]) >= a.flushSize {
		go a.flushKey(key)
	}

	return nil
}

// flush processes all buffered costs and updates daily aggregates
func (a *CostAggregator) flush() {
	a.bufferMutex.Lock()
	defer a.bufferMutex.Unlock()

	for key := range a.buffer {
		a.flushKeyLocked(key)
	}
}

// flushKey processes buffered costs for a specific tenant:queue key
func (a *CostAggregator) flushKey(key string) {
	a.bufferMutex.Lock()
	defer a.bufferMutex.Unlock()
	a.flushKeyLocked(key)
}

// flushKeyLocked processes costs for a key (assumes lock is held)
func (a *CostAggregator) flushKeyLocked(key string) {
	costs := a.buffer[key]
	if len(costs) == 0 {
		return
	}

	// Group costs by date
	dailyCosts := a.groupCostsByDate(costs)

	// Update database for each date
	for date, dateCosts := range dailyCosts {
		aggregate := a.calculateDailyAggregate(dateCosts)
		aggregate.Date = date

		if err := a.upsertDailyAggregate(aggregate); err != nil {
			// Log error but continue processing
			fmt.Printf("Error updating daily aggregate: %v\n", err)
		}
	}

	// Clear the buffer for this key
	a.buffer[key] = nil
}

// groupCostsByDate groups job costs by date
func (a *CostAggregator) groupCostsByDate(costs []JobCost) map[time.Time][]JobCost {
	dailyCosts := make(map[time.Time][]JobCost)

	for _, cost := range costs {
		date := time.Date(
			cost.Timestamp.Year(),
			cost.Timestamp.Month(),
			cost.Timestamp.Day(),
			0, 0, 0, 0,
			cost.Timestamp.Location(),
		)
		dailyCosts[date] = append(dailyCosts[date], cost)
	}

	return dailyCosts
}

// calculateDailyAggregate computes aggregate statistics for a day's costs
func (a *CostAggregator) calculateDailyAggregate(costs []JobCost) DailyCostAggregate {
	if len(costs) == 0 {
		return DailyCostAggregate{}
	}

	// Extract basic info from first cost
	first := costs[0]
	aggregate := DailyCostAggregate{
		TenantID:    first.TenantID,
		QueueName:   first.QueueName,
		TotalJobs:   len(costs),
		UpdatedAt:   time.Now(),
	}

	// Calculate totals and collect individual costs for percentiles
	totalCosts := make([]float64, len(costs))

	for i, cost := range costs {
		aggregate.TotalCost += cost.TotalCost
		aggregate.CPUCost += cost.CostBreakdown.CPUCost
		aggregate.MemoryCost += cost.CostBreakdown.MemoryCost
		aggregate.PayloadCost += cost.CostBreakdown.PayloadCost
		aggregate.RedisCost += cost.CostBreakdown.RedisCost
		aggregate.NetworkCost += cost.CostBreakdown.NetworkCost

		totalCosts[i] = cost.TotalCost
	}

	// Calculate derived metrics
	aggregate.AvgJobCost = aggregate.TotalCost / float64(len(costs))

	// Sort costs for min/max/percentile calculations
	sort.Float64s(totalCosts)
	aggregate.MinJobCost = totalCosts[0]
	aggregate.MaxJobCost = totalCosts[len(totalCosts)-1]

	// Calculate 95th percentile
	p95Index := int(float64(len(totalCosts)) * 0.95)
	if p95Index >= len(totalCosts) {
		p95Index = len(totalCosts) - 1
	}
	aggregate.P95JobCost = totalCosts[p95Index]

	return aggregate
}

// upsertDailyAggregate inserts or updates a daily cost aggregate in the database
func (a *CostAggregator) upsertDailyAggregate(aggregate DailyCostAggregate) error {
	query := `
		INSERT INTO daily_costs (
			tenant_id, queue_name, date, total_jobs, total_cost,
			cpu_cost, memory_cost, payload_cost, redis_cost, network_cost,
			avg_job_cost, max_job_cost, min_job_cost, p95_job_cost, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
		ON CONFLICT (tenant_id, queue_name, date)
		DO UPDATE SET
			total_jobs = daily_costs.total_jobs + EXCLUDED.total_jobs,
			total_cost = daily_costs.total_cost + EXCLUDED.total_cost,
			cpu_cost = daily_costs.cpu_cost + EXCLUDED.cpu_cost,
			memory_cost = daily_costs.memory_cost + EXCLUDED.memory_cost,
			payload_cost = daily_costs.payload_cost + EXCLUDED.payload_cost,
			redis_cost = daily_costs.redis_cost + EXCLUDED.redis_cost,
			network_cost = daily_costs.network_cost + EXCLUDED.network_cost,
			avg_job_cost = (daily_costs.total_cost + EXCLUDED.total_cost) /
			              (daily_costs.total_jobs + EXCLUDED.total_jobs),
			max_job_cost = GREATEST(daily_costs.max_job_cost, EXCLUDED.max_job_cost),
			min_job_cost = LEAST(daily_costs.min_job_cost, EXCLUDED.min_job_cost),
			p95_job_cost = EXCLUDED.p95_job_cost,
			updated_at = EXCLUDED.updated_at
	`

	_, err := a.db.Exec(query,
		aggregate.TenantID,
		aggregate.QueueName,
		aggregate.Date,
		aggregate.TotalJobs,
		aggregate.TotalCost,
		aggregate.CPUCost,
		aggregate.MemoryCost,
		aggregate.PayloadCost,
		aggregate.RedisCost,
		aggregate.NetworkCost,
		aggregate.AvgJobCost,
		aggregate.MaxJobCost,
		aggregate.MinJobCost,
		aggregate.P95JobCost,
		aggregate.UpdatedAt,
	)

	return err
}

// GetDailySpend returns the daily spending for a tenant/queue over a period
func (a *CostAggregator) GetDailySpend(tenantID, queueName string, startDate, endDate time.Time) ([]DailyCostAggregate, error) {
	query := `
		SELECT tenant_id, queue_name, date, total_jobs, total_cost,
		       cpu_cost, memory_cost, payload_cost, redis_cost, network_cost,
		       avg_job_cost, max_job_cost, min_job_cost, p95_job_cost, updated_at
		FROM daily_costs
		WHERE tenant_id = $1 AND queue_name = $2 AND date >= $3 AND date <= $4
		ORDER BY date
	`

	rows, err := a.db.Query(query, tenantID, queueName, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var aggregates []DailyCostAggregate
	for rows.Next() {
		var aggregate DailyCostAggregate
		err := rows.Scan(
			&aggregate.TenantID,
			&aggregate.QueueName,
			&aggregate.Date,
			&aggregate.TotalJobs,
			&aggregate.TotalCost,
			&aggregate.CPUCost,
			&aggregate.MemoryCost,
			&aggregate.PayloadCost,
			&aggregate.RedisCost,
			&aggregate.NetworkCost,
			&aggregate.AvgJobCost,
			&aggregate.MaxJobCost,
			&aggregate.MinJobCost,
			&aggregate.P95JobCost,
			&aggregate.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		aggregates = append(aggregates, aggregate)
	}

	return aggregates, rows.Err()
}

// GetCurrentSpend returns the total spending for a tenant in the current period
func (a *CostAggregator) GetCurrentSpend(tenantID, queueName string, period BudgetPeriod) (float64, error) {
	query := `
		SELECT COALESCE(SUM(total_cost), 0)
		FROM daily_costs
		WHERE tenant_id = $1 AND date >= $2 AND date <= $3
	`

	args := []interface{}{tenantID, period.StartDate, period.EndDate}

	// Add queue filter if specified
	if queueName != "" {
		query += " AND queue_name = $4"
		args = append(args, queueName)
	}

	var totalSpend float64
	err := a.db.QueryRow(query, args...).Scan(&totalSpend)
	return totalSpend, err
}

// GetTopCostDrivers returns the highest cost drivers for a tenant
func (a *CostAggregator) GetTopCostDrivers(tenantID string, startDate, endDate time.Time, limit int) ([]CostDriver, error) {
	query := `
		SELECT
			tenant_id,
			queue_name,
			SUM(total_cost) as total_cost,
			SUM(total_jobs) as job_count,
			AVG(avg_job_cost) as avg_cost_per_job
		FROM daily_costs
		WHERE tenant_id = $1 AND date >= $2 AND date <= $3
		GROUP BY tenant_id, queue_name
		ORDER BY total_cost DESC
		LIMIT $4
	`

	rows, err := a.db.Query(query, tenantID, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Get total spend for percentage calculation
	totalSpend, err := a.getTotalSpend(tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var drivers []CostDriver
	for rows.Next() {
		var driver CostDriver
		err := rows.Scan(
			&driver.TenantID,
			&driver.QueueName,
			&driver.TotalCost,
			&driver.JobCount,
			&driver.AvgCostPerJob,
		)
		if err != nil {
			return nil, err
		}

		// Calculate percentage of total spend
		if totalSpend > 0 {
			driver.Percentage = driver.TotalCost / totalSpend * 100
		}

		// Set default values
		driver.JobType = "all"
		driver.Component = "total"
		driver.Trend = "stable" // Would need historical data to calculate

		drivers = append(drivers, driver)
	}

	return drivers, rows.Err()
}

// getTotalSpend calculates total spending for a tenant across all queues
func (a *CostAggregator) getTotalSpend(tenantID string, startDate, endDate time.Time) (float64, error) {
	query := `
		SELECT COALESCE(SUM(total_cost), 0)
		FROM daily_costs
		WHERE tenant_id = $1 AND date >= $2 AND date <= $3
	`

	var totalSpend float64
	err := a.db.QueryRow(query, tenantID, startDate, endDate).Scan(&totalSpend)
	return totalSpend, err
}

// GetQueueBreakdown returns cost breakdown by queue for a tenant
func (a *CostAggregator) GetQueueBreakdown(tenantID string, startDate, endDate time.Time) (map[string]float64, error) {
	query := `
		SELECT queue_name, SUM(total_cost) as total_cost
		FROM daily_costs
		WHERE tenant_id = $1 AND date >= $2 AND date <= $3
		GROUP BY queue_name
		ORDER BY total_cost DESC
	`

	rows, err := a.db.Query(query, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	breakdown := make(map[string]float64)
	for rows.Next() {
		var queueName string
		var totalCost float64

		err := rows.Scan(&queueName, &totalCost)
		if err != nil {
			return nil, err
		}

		breakdown[queueName] = totalCost
	}

	return breakdown, rows.Err()
}

// GetComponentBreakdown returns cost breakdown by component for a tenant/queue
func (a *CostAggregator) GetComponentBreakdown(tenantID, queueName string, startDate, endDate time.Time) (map[string]float64, error) {
	query := `
		SELECT
			SUM(cpu_cost) as cpu_cost,
			SUM(memory_cost) as memory_cost,
			SUM(payload_cost) as payload_cost,
			SUM(redis_cost) as redis_cost,
			SUM(network_cost) as network_cost
		FROM daily_costs
		WHERE tenant_id = $1 AND queue_name = $2 AND date >= $3 AND date <= $4
	`

	var cpuCost, memoryCost, payloadCost, redisCost, networkCost float64
	err := a.db.QueryRow(query, tenantID, queueName, startDate, endDate).Scan(
		&cpuCost, &memoryCost, &payloadCost, &redisCost, &networkCost,
	)
	if err != nil {
		return nil, err
	}

	breakdown := map[string]float64{
		"cpu":     cpuCost,
		"memory":  memoryCost,
		"payload": payloadCost,
		"redis":   redisCost,
		"network": networkCost,
	}

	return breakdown, nil
}

// PurgeOldData removes cost data older than the specified retention period
func (a *CostAggregator) PurgeOldData(retentionDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	query := `DELETE FROM daily_costs WHERE date < $1`
	result, err := a.db.Exec(query, cutoffDate)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	fmt.Printf("Purged %d old cost records older than %s\n", rowsAffected, cutoffDate.Format("2006-01-02"))
	return nil
}