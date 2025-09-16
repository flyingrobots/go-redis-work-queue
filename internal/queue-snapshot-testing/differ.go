// Copyright 2025 James Ross
package queuesnapshotesting

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
)

// Differ compares snapshots with smart diffing
type Differ struct {
	config *SnapshotConfig
}

// NewDiffer creates a new differ
func NewDiffer(config *SnapshotConfig) *Differ {
	return &Differ{
		config: config,
	}
}

// Compare compares two snapshots
func (d *Differ) Compare(left, right *Snapshot) (*DiffResult, error) {
	result := &DiffResult{
		LeftID:          left.ID,
		RightID:         right.ID,
		Timestamp:       time.Now(),
		QueueChanges:    []Change{},
		JobChanges:      []Change{},
		WorkerChanges:   []Change{},
		MetricChanges:   []Change{},
		SemanticChanges: []SemanticChange{},
	}

	// Compare queues
	d.compareQueues(left.Queues, right.Queues, result)

	// Compare jobs
	d.compareJobs(left.Jobs, right.Jobs, result)

	// Compare workers
	d.compareWorkers(left.Workers, right.Workers, result)

	// Compare metrics
	d.compareMetrics(left.Metrics, right.Metrics, result)

	// Analyze semantic changes
	d.analyzeSemanticChanges(result)

	// Calculate totals
	result.TotalChanges = len(result.QueueChanges) + len(result.JobChanges) +
		len(result.WorkerChanges) + len(result.MetricChanges)

	for _, change := range append(append(append(result.QueueChanges, result.JobChanges...),
		result.WorkerChanges...), result.MetricChanges...) {
		switch change.Type {
		case ChangeAdded:
			result.Added++
		case ChangeRemoved:
			result.Removed++
		case ChangeModified:
			result.Modified++
		}
	}

	return result, nil
}

func (d *Differ) compareQueues(left, right []QueueState, result *DiffResult) {
	leftMap := make(map[string]*QueueState)
	rightMap := make(map[string]*QueueState)

	for i := range left {
		leftMap[left[i].Name] = &left[i]
	}

	for i := range right {
		rightMap[right[i].Name] = &right[i]
	}

	// Check for removed queues
	for name, lq := range leftMap {
		if _, exists := rightMap[name]; !exists {
			result.QueueChanges = append(result.QueueChanges, Change{
				Type:        ChangeRemoved,
				Path:        fmt.Sprintf("queue.%s", name),
				OldValue:    lq,
				Description: fmt.Sprintf("Queue '%s' was removed", name),
				Impact:      "high",
			})
		}
	}

	// Check for added/modified queues
	for name, rq := range rightMap {
		lq, exists := leftMap[name]
		if !exists {
			result.QueueChanges = append(result.QueueChanges, Change{
				Type:        ChangeAdded,
				Path:        fmt.Sprintf("queue.%s", name),
				NewValue:    rq,
				Description: fmt.Sprintf("Queue '%s' was added", name),
				Impact:      "high",
			})
		} else {
			// Compare queue properties
			if lq.Length != rq.Length {
				change := Change{
					Type:        ChangeModified,
					Path:        fmt.Sprintf("queue.%s.length", name),
					OldValue:    lq.Length,
					NewValue:    rq.Length,
					Description: fmt.Sprintf("Queue '%s' length changed from %d to %d", name, lq.Length, rq.Length),
				}

				// Determine impact
				diff := rq.Length - lq.Length
				if diff > 100 || diff < -100 {
					change.Impact = "high"
				} else if diff > 10 || diff < -10 {
					change.Impact = "medium"
				} else {
					change.Impact = "low"
				}

				result.QueueChanges = append(result.QueueChanges, change)
			}

			// Compare configurations
			d.compareConfigs(lq.Config, rq.Config, name, result)
		}
	}
}

func (d *Differ) compareJobs(left, right []JobState, result *DiffResult) {
	// Build maps by queue and status
	leftByQueue := d.groupJobsByQueue(left)
	rightByQueue := d.groupJobsByQueue(right)

	for queueName, leftJobs := range leftByQueue {
		rightJobs := rightByQueue[queueName]

		// Count by status
		leftStatus := d.countJobsByStatus(leftJobs)
		rightStatus := d.countJobsByStatus(rightJobs)

		for status, leftCount := range leftStatus {
			rightCount := rightStatus[status]
			if leftCount != rightCount {
				change := Change{
					Type:        ChangeModified,
					Path:        fmt.Sprintf("jobs.%s.%s", queueName, status),
					OldValue:    leftCount,
					NewValue:    rightCount,
					Description: fmt.Sprintf("Queue '%s' %s jobs changed from %d to %d",
						queueName, status, leftCount, rightCount),
				}

				diff := rightCount - leftCount
				if diff > 50 || diff < -50 {
					change.Impact = "high"
				} else if diff > 10 || diff < -10 {
					change.Impact = "medium"
				} else {
					change.Impact = "low"
				}

				result.JobChanges = append(result.JobChanges, change)
			}
		}
	}

	// Check for job movements between queues
	if !d.config.IgnoreIDs {
		d.detectJobMovements(left, right, result)
	}
}

func (d *Differ) compareWorkers(left, right []WorkerState, result *DiffResult) {
	if d.config.IgnoreWorkerIDs {
		// Compare only worker counts and aggregate stats
		leftActive := d.countActiveWorkers(left)
		rightActive := d.countActiveWorkers(right)

		if leftActive != rightActive {
			result.WorkerChanges = append(result.WorkerChanges, Change{
				Type:        ChangeModified,
				Path:        "workers.active_count",
				OldValue:    leftActive,
				NewValue:    rightActive,
				Description: fmt.Sprintf("Active workers changed from %d to %d", leftActive, rightActive),
				Impact:      d.determineWorkerImpact(leftActive, rightActive),
			})
		}
	} else {
		// Detailed worker comparison
		leftMap := make(map[string]*WorkerState)
		rightMap := make(map[string]*WorkerState)

		for i := range left {
			leftMap[left[i].ID] = &left[i]
		}

		for i := range right {
			rightMap[right[i].ID] = &right[i]
		}

		// Check for removed workers
		for id, lw := range leftMap {
			if _, exists := rightMap[id]; !exists {
				result.WorkerChanges = append(result.WorkerChanges, Change{
					Type:        ChangeRemoved,
					Path:        fmt.Sprintf("worker.%s", id),
					OldValue:    lw,
					Description: fmt.Sprintf("Worker '%s' was removed", id),
					Impact:      "medium",
				})
			}
		}

		// Check for added/modified workers
		for id, rw := range rightMap {
			lw, exists := leftMap[id]
			if !exists {
				result.WorkerChanges = append(result.WorkerChanges, Change{
					Type:        ChangeAdded,
					Path:        fmt.Sprintf("worker.%s", id),
					NewValue:    rw,
					Description: fmt.Sprintf("Worker '%s' was added", id),
					Impact:      "medium",
				})
			} else if lw.Status != rw.Status {
				result.WorkerChanges = append(result.WorkerChanges, Change{
					Type:        ChangeModified,
					Path:        fmt.Sprintf("worker.%s.status", id),
					OldValue:    lw.Status,
					NewValue:    rw.Status,
					Description: fmt.Sprintf("Worker '%s' status changed from '%s' to '%s'",
						id, lw.Status, rw.Status),
					Impact:      "low",
				})
			}
		}
	}
}

func (d *Differ) compareMetrics(left, right map[string]interface{}, result *DiffResult) {
	// Check for removed metrics
	for key, lval := range left {
		if _, exists := right[key]; !exists {
			result.MetricChanges = append(result.MetricChanges, Change{
				Type:        ChangeRemoved,
				Path:        fmt.Sprintf("metrics.%s", key),
				OldValue:    lval,
				Description: fmt.Sprintf("Metric '%s' was removed", key),
				Impact:      "low",
			})
		}
	}

	// Check for added/modified metrics
	for key, rval := range right {
		lval, exists := left[key]
		if !exists {
			result.MetricChanges = append(result.MetricChanges, Change{
				Type:        ChangeAdded,
				Path:        fmt.Sprintf("metrics.%s", key),
				NewValue:    rval,
				Description: fmt.Sprintf("Metric '%s' was added", key),
				Impact:      "low",
			})
		} else if !d.valuesEqual(lval, rval) {
			result.MetricChanges = append(result.MetricChanges, Change{
				Type:        ChangeModified,
				Path:        fmt.Sprintf("metrics.%s", key),
				OldValue:    lval,
				NewValue:    rval,
				Description: fmt.Sprintf("Metric '%s' changed from %v to %v", key, lval, rval),
				Impact:      d.determineMetricImpact(key, lval, rval),
			})
		}
	}
}

func (d *Differ) compareConfigs(left, right map[string]interface{}, queueName string, result *DiffResult) {
	for key, rval := range right {
		lval, exists := left[key]
		if !exists || !d.valuesEqual(lval, rval) {
			var changeType ChangeType
			if !exists {
				changeType = ChangeAdded
			} else {
				changeType = ChangeModified
			}

			result.QueueChanges = append(result.QueueChanges, Change{
				Type:        changeType,
				Path:        fmt.Sprintf("queue.%s.config.%s", queueName, key),
				OldValue:    lval,
				NewValue:    rval,
				Description: fmt.Sprintf("Queue '%s' config '%s' changed", queueName, key),
				Impact:      "medium",
			})
		}
	}

	for key, lval := range left {
		if _, exists := right[key]; !exists {
			result.QueueChanges = append(result.QueueChanges, Change{
				Type:        ChangeRemoved,
				Path:        fmt.Sprintf("queue.%s.config.%s", queueName, key),
				OldValue:    lval,
				Description: fmt.Sprintf("Queue '%s' config '%s' was removed", queueName, key),
				Impact:      "medium",
			})
		}
	}
}

func (d *Differ) analyzeSemanticChanges(result *DiffResult) {
	// Analyze queue load changes
	var totalQueueGrowth int64
	for _, change := range result.QueueChanges {
		if strings.HasSuffix(change.Path, ".length") {
			if old, ok := change.OldValue.(int64); ok {
				if new, ok := change.NewValue.(int64); ok {
					totalQueueGrowth += new - old
				}
			}
		}
	}

	if totalQueueGrowth > 100 {
		result.SemanticChanges = append(result.SemanticChanges, SemanticChange{
			Type:        "queue_overload",
			Description: fmt.Sprintf("Significant queue growth detected: +%d jobs", totalQueueGrowth),
			Severity:    "high",
			Components:  d.getAffectedQueues(result.QueueChanges),
		})
	} else if totalQueueGrowth < -100 {
		result.SemanticChanges = append(result.SemanticChanges, SemanticChange{
			Type:        "queue_drain",
			Description: fmt.Sprintf("Significant queue drain detected: %d jobs", totalQueueGrowth),
			Severity:    "medium",
			Components:  d.getAffectedQueues(result.QueueChanges),
		})
	}

	// Analyze worker changes
	workerAdded := 0
	workerRemoved := 0
	for _, change := range result.WorkerChanges {
		switch change.Type {
		case ChangeAdded:
			workerAdded++
		case ChangeRemoved:
			workerRemoved++
		}
	}

	if workerAdded > 0 || workerRemoved > 0 {
		result.SemanticChanges = append(result.SemanticChanges, SemanticChange{
			Type:        "worker_scaling",
			Description: fmt.Sprintf("Worker pool changed: +%d/-%d workers", workerAdded, workerRemoved),
			Severity:    "medium",
			Components:  []string{"worker_pool"},
		})
	}

	// Analyze error rate changes
	for _, change := range result.MetricChanges {
		if strings.Contains(change.Path, "error") || strings.Contains(change.Path, "failed") {
			result.SemanticChanges = append(result.SemanticChanges, SemanticChange{
				Type:        "error_rate_change",
				Description: change.Description,
				Severity:    "high",
				Components:  []string{change.Path},
			})
		}
	}
}

func (d *Differ) detectJobMovements(left, right []JobState, result *DiffResult) {
	// Build job maps by ID
	leftMap := make(map[string]*JobState)
	rightMap := make(map[string]*JobState)

	for i := range left {
		leftMap[left[i].ID] = &left[i]
	}

	for i := range right {
		rightMap[right[i].ID] = &right[i]
	}

	// Check for moved jobs
	for id, rj := range rightMap {
		if lj, exists := leftMap[id]; exists {
			if lj.QueueName != rj.QueueName {
				result.JobChanges = append(result.JobChanges, Change{
					Type:        ChangeMoved,
					Path:        fmt.Sprintf("job.%s", id),
					OldValue:    lj.QueueName,
					NewValue:    rj.QueueName,
					Description: fmt.Sprintf("Job '%s' moved from queue '%s' to '%s'",
						id, lj.QueueName, rj.QueueName),
					Impact:      "medium",
				})
			}
		}
	}
}

// Helper methods

func (d *Differ) groupJobsByQueue(jobs []JobState) map[string][]JobState {
	grouped := make(map[string][]JobState)
	for _, job := range jobs {
		grouped[job.QueueName] = append(grouped[job.QueueName], job)
	}
	return grouped
}

func (d *Differ) countJobsByStatus(jobs []JobState) map[string]int {
	counts := make(map[string]int)
	for _, job := range jobs {
		counts[job.Status]++
	}
	return counts
}

func (d *Differ) countActiveWorkers(workers []WorkerState) int {
	count := 0
	for _, worker := range workers {
		if worker.Status == "active" || worker.Status == "busy" {
			count++
		}
	}
	return count
}

func (d *Differ) valuesEqual(a, b interface{}) bool {
	// Handle timestamp ignoring
	if d.config.IgnoreTimestamps {
		if _, ok := a.(time.Time); ok {
			return true
		}
		if aStr, ok := a.(string); ok {
			if _, err := time.Parse(time.RFC3339, aStr); err == nil {
				return true
			}
		}
	}

	// Handle custom ignores
	for _, pattern := range d.config.CustomIgnores {
		if aStr, ok := a.(string); ok {
			if strings.Contains(aStr, pattern) {
				return true
			}
		}
	}

	return reflect.DeepEqual(a, b)
}

func (d *Differ) determineWorkerImpact(old, new int) string {
	diff := new - old
	if diff > 5 || diff < -5 {
		return "high"
	} else if diff > 2 || diff < -2 {
		return "medium"
	}
	return "low"
}

func (d *Differ) determineMetricImpact(key string, old, new interface{}) string {
	// Critical metrics
	criticalMetrics := []string{"error_rate", "failed", "timeout"}
	for _, critical := range criticalMetrics {
		if strings.Contains(key, critical) {
			return "high"
		}
	}

	// Performance metrics
	perfMetrics := []string{"latency", "throughput", "processed"}
	for _, perf := range perfMetrics {
		if strings.Contains(key, perf) {
			return "medium"
		}
	}

	return "low"
}

func (d *Differ) getAffectedQueues(changes []Change) []string {
	queues := make(map[string]bool)
	for _, change := range changes {
		parts := strings.Split(change.Path, ".")
		if len(parts) >= 2 && parts[0] == "queue" {
			queues[parts[1]] = true
		}
	}

	result := make([]string, 0, len(queues))
	for queue := range queues {
		result = append(result, queue)
	}

	sort.Strings(result)
	return result
}