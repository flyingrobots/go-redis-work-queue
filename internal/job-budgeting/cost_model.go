package budgeting

import (
	"fmt"
	"math"
	"time"
)

// CostCalculationEngine handles job cost calculations using configurable models
type CostCalculationEngine struct {
	model      *CostModel
	calibrator *ModelCalibrator
}

// NewCostCalculationEngine creates a new cost calculation engine
func NewCostCalculationEngine(model *CostModel) *CostCalculationEngine {
	return &CostCalculationEngine{
		model:      model,
		calibrator: NewModelCalibrator(),
	}
}

// CalculateJobCost calculates the total cost of a job based on its metrics
func (c *CostCalculationEngine) CalculateJobCost(metrics JobMetrics) (JobCost, error) {
	if err := c.validateMetrics(metrics); err != nil {
		return JobCost{}, NewCostCalculationError(
			metrics.JobID,
			metrics.TenantID,
			"validation",
			"invalid metrics",
			err,
		)
	}

	breakdown := c.calculateCostBreakdown(metrics)
	totalCost := c.sumCostBreakdown(breakdown)

	// Apply environment multiplier
	totalCost *= c.model.EnvironmentMultiplier

	return JobCost{
		JobID:         metrics.JobID,
		TenantID:      metrics.TenantID,
		QueueName:     metrics.QueueName,
		CPUTime:       metrics.CPUTime,
		MemorySeconds: metrics.MemoryMBSeconds,
		PayloadSize:   metrics.PayloadBytes,
		RedisOps:      metrics.RedisOps,
		NetworkBytes:  metrics.NetworkBytes,
		TotalCost:     totalCost,
		CostBreakdown: breakdown,
		Timestamp:     time.Now(),
		JobType:       metrics.JobType,
		Priority:      metrics.Priority,
	}, nil
}

// calculateCostBreakdown breaks down the cost into individual components
func (c *CostCalculationEngine) calculateCostBreakdown(metrics JobMetrics) CostBreakdown {
	return CostBreakdown{
		BaseCost:    c.model.BaseJobWeight,
		CPUCost:     metrics.CPUTime * c.model.CPUTimeWeight,
		MemoryCost:  metrics.MemoryMBSeconds * c.model.MemoryWeight,
		PayloadCost: float64(metrics.PayloadBytes) / 1024 * c.model.PayloadWeight,
		RedisCost:   float64(metrics.RedisOps) * c.model.RedisOpsWeight,
		NetworkCost: float64(metrics.NetworkBytes) / 1024 / 1024 * c.model.NetworkWeight,
	}
}

// sumCostBreakdown calculates the total cost from breakdown components
func (c *CostCalculationEngine) sumCostBreakdown(breakdown CostBreakdown) float64 {
	return breakdown.BaseCost +
		breakdown.CPUCost +
		breakdown.MemoryCost +
		breakdown.PayloadCost +
		breakdown.RedisCost +
		breakdown.NetworkCost
}

// validateMetrics ensures job metrics are valid for cost calculation
func (c *CostCalculationEngine) validateMetrics(metrics JobMetrics) error {
	if metrics.JobID == "" {
		return NewValidationError("job_id", metrics.JobID, "required", "job ID cannot be empty")
	}

	if metrics.TenantID == "" {
		return NewValidationError("tenant_id", metrics.TenantID, "required", "tenant ID cannot be empty")
	}

	if metrics.CPUTime < 0 {
		return NewValidationError("cpu_time", metrics.CPUTime, "non_negative", "CPU time cannot be negative")
	}

	if metrics.MemoryMBSeconds < 0 {
		return NewValidationError("memory_mb_seconds", metrics.MemoryMBSeconds, "non_negative", "memory usage cannot be negative")
	}

	if metrics.PayloadBytes < 0 {
		return NewValidationError("payload_bytes", metrics.PayloadBytes, "non_negative", "payload size cannot be negative")
	}

	if metrics.RedisOps < 0 {
		return NewValidationError("redis_ops", metrics.RedisOps, "non_negative", "Redis operations cannot be negative")
	}

	if metrics.NetworkBytes < 0 {
		return NewValidationError("network_bytes", metrics.NetworkBytes, "non_negative", "network bytes cannot be negative")
	}

	return nil
}

// UpdateModel updates the cost model with new weights
func (c *CostCalculationEngine) UpdateModel(newModel *CostModel) error {
	if err := c.validateCostModel(newModel); err != nil {
		return err
	}
	c.model = newModel
	return nil
}

// validateCostModel ensures the cost model configuration is valid
func (c *CostCalculationEngine) validateCostModel(model *CostModel) error {
	if model.CPUTimeWeight < 0 {
		return NewValidationError("cpu_time_weight", model.CPUTimeWeight, "non_negative", "CPU time weight cannot be negative")
	}

	if model.MemoryWeight < 0 {
		return NewValidationError("memory_weight", model.MemoryWeight, "non_negative", "memory weight cannot be negative")
	}

	if model.PayloadWeight < 0 {
		return NewValidationError("payload_weight", model.PayloadWeight, "non_negative", "payload weight cannot be negative")
	}

	if model.RedisOpsWeight < 0 {
		return NewValidationError("redis_ops_weight", model.RedisOpsWeight, "non_negative", "Redis ops weight cannot be negative")
	}

	if model.NetworkWeight < 0 {
		return NewValidationError("network_weight", model.NetworkWeight, "non_negative", "network weight cannot be negative")
	}

	if model.BaseJobWeight < 0 {
		return NewValidationError("base_job_weight", model.BaseJobWeight, "non_negative", "base job weight cannot be negative")
	}

	if model.EnvironmentMultiplier <= 0 {
		return NewValidationError("environment_multiplier", model.EnvironmentMultiplier, "positive", "environment multiplier must be positive")
	}

	return nil
}

// GetModel returns the current cost model
func (c *CostCalculationEngine) GetModel() *CostModel {
	return c.model
}

// ModelCalibrator handles cost model calibration against real infrastructure costs
type ModelCalibrator struct {
	benchmarks []BenchmarkResult
}

// BenchmarkResult represents a calibration benchmark result
type BenchmarkResult struct {
	JobType          string    `json:"job_type"`
	ActualCost       float64   `json:"actual_cost"`
	CalculatedCost   float64   `json:"calculated_cost"`
	CPUTime          float64   `json:"cpu_time"`
	MemoryMBSeconds  float64   `json:"memory_mb_seconds"`
	PayloadBytes     int       `json:"payload_bytes"`
	RedisOps         int       `json:"redis_ops"`
	NetworkBytes     int       `json:"network_bytes"`
	InfrastructureCost float64 `json:"infrastructure_cost"`
	Timestamp        time.Time `json:"timestamp"`
}

// NewModelCalibrator creates a new model calibrator
func NewModelCalibrator() *ModelCalibrator {
	return &ModelCalibrator{
		benchmarks: make([]BenchmarkResult, 0),
	}
}

// AddBenchmark adds a benchmark result for model calibration
func (m *ModelCalibrator) AddBenchmark(result BenchmarkResult) {
	m.benchmarks = append(m.benchmarks, result)
}

// CalibrateModel adjusts model weights based on benchmark data
func (m *ModelCalibrator) CalibrateModel(currentModel *CostModel) (*CostModel, error) {
	if len(m.benchmarks) < 10 {
		return nil, NewValidationError("benchmarks", len(m.benchmarks), "min_count", "need at least 10 benchmarks for calibration")
	}

	// Calculate adjustment factors based on actual vs calculated costs
	totalActual := 0.0
	totalCalculated := 0.0

	for _, benchmark := range m.benchmarks {
		totalActual += benchmark.ActualCost
		totalCalculated += benchmark.CalculatedCost
	}

	if totalCalculated == 0 {
		return nil, fmt.Errorf("calculated costs sum to zero, cannot calibrate")
	}

	adjustmentFactor := totalActual / totalCalculated

	// Apply calibration to create new model
	calibratedModel := &CostModel{
		CPUTimeWeight:         currentModel.CPUTimeWeight * adjustmentFactor,
		MemoryWeight:          currentModel.MemoryWeight * adjustmentFactor,
		PayloadWeight:         currentModel.PayloadWeight * adjustmentFactor,
		RedisOpsWeight:        currentModel.RedisOpsWeight * adjustmentFactor,
		NetworkWeight:         currentModel.NetworkWeight * adjustmentFactor,
		BaseJobWeight:         currentModel.BaseJobWeight * adjustmentFactor,
		EnvironmentMultiplier: currentModel.EnvironmentMultiplier,
	}

	return calibratedModel, nil
}

// GetCalibrationAccuracy returns the accuracy of the current model
func (m *ModelCalibrator) GetCalibrationAccuracy() float64 {
	if len(m.benchmarks) == 0 {
		return 0.0
	}

	totalError := 0.0
	for _, benchmark := range m.benchmarks {
		error := math.Abs(benchmark.ActualCost - benchmark.CalculatedCost)
		totalError += error / benchmark.ActualCost // Relative error
	}

	return 1.0 - (totalError / float64(len(m.benchmarks)))
}

// GetCalibrationReport generates a detailed calibration report
func (m *ModelCalibrator) GetCalibrationReport() CalibrationReport {
	if len(m.benchmarks) == 0 {
		return CalibrationReport{
			BenchmarkCount: 0,
			Accuracy:       0.0,
			MeanError:      0.0,
		}
	}

	totalError := 0.0
	totalRelativeError := 0.0
	maxError := 0.0

	for _, benchmark := range m.benchmarks {
		absoluteError := math.Abs(benchmark.ActualCost - benchmark.CalculatedCost)
		relativeError := absoluteError / benchmark.ActualCost

		totalError += absoluteError
		totalRelativeError += relativeError

		if absoluteError > maxError {
			maxError = absoluteError
		}
	}

	meanError := totalError / float64(len(m.benchmarks))
	meanRelativeError := totalRelativeError / float64(len(m.benchmarks))
	accuracy := 1.0 - meanRelativeError

	return CalibrationReport{
		BenchmarkCount:     len(m.benchmarks),
		Accuracy:           accuracy,
		MeanError:          meanError,
		MeanRelativeError:  meanRelativeError,
		MaxError:           maxError,
		RecommendedAction:  m.getRecommendedAction(accuracy),
		LastCalibration:    time.Now(),
	}
}

// CalibrationReport provides insights into model accuracy
type CalibrationReport struct {
	BenchmarkCount     int       `json:"benchmark_count"`
	Accuracy           float64   `json:"accuracy"`           // 0.0 to 1.0
	MeanError          float64   `json:"mean_error"`         // Average absolute error
	MeanRelativeError  float64   `json:"mean_relative_error"` // Average relative error
	MaxError           float64   `json:"max_error"`          // Maximum absolute error
	RecommendedAction  string    `json:"recommended_action"`
	LastCalibration    time.Time `json:"last_calibration"`
}

// getRecommendedAction returns calibration recommendations based on accuracy
func (m *ModelCalibrator) getRecommendedAction(accuracy float64) string {
	switch {
	case accuracy > 0.95:
		return "Model is highly accurate, no action needed"
	case accuracy > 0.85:
		return "Model is reasonably accurate, consider periodic recalibration"
	case accuracy > 0.70:
		return "Model accuracy is moderate, recalibration recommended"
	case accuracy > 0.50:
		return "Model accuracy is poor, immediate recalibration required"
	default:
		return "Model accuracy is very poor, review benchmarking methodology"
	}
}

// EstimateJobCost provides a quick cost estimate for job planning
func (c *CostCalculationEngine) EstimateJobCost(jobType string, estimatedCPUTime float64, estimatedPayloadKB int) float64 {
	// Use simplified estimation based on job type patterns
	var memoryMultiplier float64 = 1.0
	var redisOpsMultiplier float64 = 1.0

	// Job type-specific adjustments
	switch jobType {
	case "ml-training":
		memoryMultiplier = 5.0
		redisOpsMultiplier = 0.5
	case "data-processing":
		memoryMultiplier = 2.0
		redisOpsMultiplier = 2.0
	case "image-processing":
		memoryMultiplier = 3.0
		redisOpsMultiplier = 1.0
	case "api-call":
		memoryMultiplier = 0.5
		redisOpsMultiplier = 1.5
	default:
		// Use default multipliers
	}

	estimatedMemoryMBSeconds := estimatedCPUTime * 100 * memoryMultiplier // Assume 100MB avg
	estimatedRedisOps := int(estimatedCPUTime * 10 * redisOpsMultiplier)  // Assume 10 ops per second
	estimatedNetworkMB := float64(estimatedPayloadKB) / 1024 * 2          // Assume 2x payload for network

	totalCost := c.model.BaseJobWeight +
		(estimatedCPUTime * c.model.CPUTimeWeight) +
		(estimatedMemoryMBSeconds * c.model.MemoryWeight) +
		(float64(estimatedPayloadKB) * c.model.PayloadWeight) +
		(float64(estimatedRedisOps) * c.model.RedisOpsWeight) +
		(estimatedNetworkMB * c.model.NetworkWeight)

	return totalCost * c.model.EnvironmentMultiplier
}

// GetCostBreakdownPercentages returns the percentage contribution of each cost component
func (c *CostCalculationEngine) GetCostBreakdownPercentages(breakdown CostBreakdown) map[string]float64 {
	total := c.sumCostBreakdown(breakdown)
	if total == 0 {
		return map[string]float64{
			"base":    0,
			"cpu":     0,
			"memory":  0,
			"payload": 0,
			"redis":   0,
			"network": 0,
		}
	}

	return map[string]float64{
		"base":    breakdown.BaseCost / total * 100,
		"cpu":     breakdown.CPUCost / total * 100,
		"memory":  breakdown.MemoryCost / total * 100,
		"payload": breakdown.PayloadCost / total * 100,
		"redis":   breakdown.RedisCost / total * 100,
		"network": breakdown.NetworkCost / total * 100,
	}
}