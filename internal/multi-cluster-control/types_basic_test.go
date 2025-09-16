package multicluster

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnectionStatus(t *testing.T) {
	status := &ConnectionStatus{
		Connected:   true,
		LastChecked: time.Now(),
		Latency:     45.2,
		LastError:   "",
	}

	assert.True(t, status.Connected)
	assert.Equal(t, 45.2, status.Latency)
	assert.Empty(t, status.LastError)

	// Test unhealthy states
	status.Connected = false
	assert.False(t, status.Connected)

	status.Connected = true
	status.Latency = 2000 // High latency
	assert.Equal(t, 2000.0, status.Latency)

	status.Latency = 45.2
	status.LastError = "connection timeout"
	assert.Equal(t, "connection timeout", status.LastError)
}

func TestClusterConnection(t *testing.T) {
	conn := &ClusterConnection{
		Config: ClusterConfig{
			Name:     "test",
			Endpoint: "localhost:6379",
		},
		Client:   nil, // Would be a real Redis client in practice
		LastPing: time.Now(),
		Status: ConnectionStatus{
			Connected:   true,
			LastChecked: time.Now(),
			Latency:     30.0,
		},
	}

	assert.Equal(t, "test", conn.Config.Name)
	assert.Equal(t, "localhost:6379", conn.Config.Endpoint)
	assert.True(t, conn.Status.Connected)
	assert.Equal(t, 30.0, conn.Status.Latency)
}

func TestClusterStats(t *testing.T) {
	stats := &ClusterStats{
		ClusterName:     "test",
		QueueSizes:      map[string]int64{"high": 10, "normal": 20},
		ProcessingCount: 5,
		DeadLetterCount: 2,
		WorkerCount:     3,
		JobRate:         100.5,
		ErrorRate:       0.5,
		Timestamp:       time.Now(),
	}

	assert.Equal(t, "test", stats.ClusterName)
	assert.Equal(t, int64(10), stats.QueueSizes["high"])
	assert.Equal(t, int64(20), stats.QueueSizes["normal"])
	assert.Equal(t, int64(5), stats.ProcessingCount)
	assert.Equal(t, int64(2), stats.DeadLetterCount)
	assert.Equal(t, 3, stats.WorkerCount)
	assert.Equal(t, 100.5, stats.JobRate)
	assert.Equal(t, 0.5, stats.ErrorRate)
}

func TestMetricCompare(t *testing.T) {
	compare := MetricCompare{
		Name: "queue_size",
		Values: map[string]float64{
			"cluster1": 100.0,
			"cluster2": 150.0,
			"cluster3": 80.0,
		},
		Delta: 70.0,
		Unit:  "jobs",
	}

	assert.Equal(t, "queue_size", compare.Name)
	assert.Equal(t, 100.0, compare.Values["cluster1"])
	assert.Equal(t, 150.0, compare.Values["cluster2"])
	assert.Equal(t, 80.0, compare.Values["cluster3"])
	assert.Equal(t, 70.0, compare.Delta)
	assert.Equal(t, "jobs", compare.Unit)
}

func TestCompareResult(t *testing.T) {
	result := &CompareResult{
		Clusters: []string{"cluster1", "cluster2"},
		Metrics: map[string]MetricCompare{
			"queue_size": {
				Values: map[string]float64{
					"cluster1": 100.0,
					"cluster2": 120.0,
				},
				Delta: 20.0,
			},
		},
		Anomalies: []Anomaly{
			{
				Type:        "deviation",
				Cluster:     "cluster2",
				Description: "High queue size",
				Severity:    "warning",
			},
		},
		Timestamp: time.Now(),
	}

	assert.Len(t, result.Clusters, 2)
	assert.Contains(t, result.Clusters, "cluster1")
	assert.Contains(t, result.Clusters, "cluster2")
	assert.Len(t, result.Metrics, 1)
	assert.Len(t, result.Anomalies, 1)
	assert.Equal(t, "deviation", result.Anomalies[0].Type)
	assert.Equal(t, "cluster2", result.Anomalies[0].Cluster)
}

func TestMultiAction(t *testing.T) {
	action := &MultiAction{
		ID:      "test-action",
		Type:    ActionTypePurgeDLQ,
		Targets: []string{"cluster1", "cluster2"},
		Parameters: map[string]interface{}{
			"confirm": true,
		},
		Status:    ActionStatusPending,
		CreatedAt: time.Now(),
		Results:   make(map[string]ActionResult),
	}

	assert.Equal(t, "test-action", action.ID)
	assert.Equal(t, ActionTypePurgeDLQ, action.Type)
	assert.Len(t, action.Targets, 2)
	assert.Equal(t, ActionStatusPending, action.Status)
	assert.True(t, action.Parameters["confirm"].(bool))

	// Simulate completion
	action.Status = ActionStatusCompleted
	executed := time.Now()
	action.ExecutedAt = &executed
	action.Results = map[string]ActionResult{
		"cluster1": {Success: true, Duration: 1000},
		"cluster2": {Success: true, Duration: 1500},
	}

	assert.Equal(t, ActionStatusCompleted, action.Status)
	assert.NotNil(t, action.ExecutedAt)
	assert.Len(t, action.Results, 2)
	assert.True(t, action.Results["cluster1"].Success)
	assert.True(t, action.Results["cluster2"].Success)
}

func TestActionResult(t *testing.T) {
	result := ActionResult{
		Success:   true,
		Message:   "Operation completed",
		Duration:  1500.0,
		Timestamp: time.Now(),
	}

	assert.True(t, result.Success)
	assert.Equal(t, "Operation completed", result.Message)
	assert.Equal(t, 1500.0, result.Duration)

	// Test failed result
	result.Success = false
	result.Error = "Connection failed"
	assert.False(t, result.Success)
	assert.Equal(t, "Connection failed", result.Error)
}

func TestHealthStatus(t *testing.T) {
	health := &HealthStatus{
		Healthy: true,
		Issues:  []string{},
		Metrics: map[string]float64{
			"latency_ms":        30.0,
			"worker_count":      5.0,
			"dead_letter_count": 0.0,
		},
		LastChecked: time.Now(),
	}

	assert.True(t, health.Healthy)
	assert.Empty(t, health.Issues)
	assert.Equal(t, 30.0, health.Metrics["latency_ms"])
	assert.Equal(t, 5.0, health.Metrics["worker_count"])
	assert.Equal(t, 0.0, health.Metrics["dead_letter_count"])

	// Add issues
	health.Issues = []string{
		"High latency detected",
		"Critical: No workers available",
		"Warning: Dead letter queue growing",
	}
	health.Healthy = false

	assert.False(t, health.Healthy)
	assert.Len(t, health.Issues, 3)
	assert.Contains(t, health.Issues, "High latency detected")
}

func TestAnomaly(t *testing.T) {
	anomaly := Anomaly{
		Type:        "deviation",
		Cluster:     "cluster1",
		Description: "Queue size deviation",
		Value:       150.0,
		Expected:    100.0,
		Severity:    "warning",
		Timestamp:   time.Now(),
	}

	assert.Equal(t, "deviation", anomaly.Type)
	assert.Equal(t, "cluster1", anomaly.Cluster)
	assert.Equal(t, "Queue size deviation", anomaly.Description)
	assert.Equal(t, 150.0, anomaly.Value)
	assert.Equal(t, 100.0, anomaly.Expected)
	assert.Equal(t, "warning", anomaly.Severity)
}

func TestTabConfig(t *testing.T) {
	config := &TabConfig{
		Tabs: []TabInfo{
			{Index: 1, ClusterName: "cluster1", Label: "Prod", Shortcut: "1"},
			{Index: 2, ClusterName: "cluster2", Label: "Stage", Shortcut: "2"},
		},
		ActiveTab:   0,
		CompareMode: false,
		CompareWith: []string{},
	}

	assert.Len(t, config.Tabs, 2)
	assert.Equal(t, 0, config.ActiveTab)
	assert.False(t, config.CompareMode)
	assert.Empty(t, config.CompareWith)

	// Enable compare mode
	config.CompareMode = true
	config.CompareWith = []string{"cluster1", "cluster2"}
	assert.True(t, config.CompareMode)
	assert.Len(t, config.CompareWith, 2)
}

func TestTabInfo(t *testing.T) {
	tab := TabInfo{
		Index:       1,
		ClusterName: "production",
		Label:       "Production",
		Color:       "#ff0000",
		Shortcut:    "1",
	}

	assert.Equal(t, 1, tab.Index)
	assert.Equal(t, "production", tab.ClusterName)
	assert.Equal(t, "Production", tab.Label)
	assert.Equal(t, "#ff0000", tab.Color)
	assert.Equal(t, "1", tab.Shortcut)
}

func TestEvent(t *testing.T) {
	event := Event{
		ID:        "event-123",
		Type:      EventTypeClusterConnected,
		Cluster:   "cluster1",
		Message:   "Connected successfully",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"latency": 30.0,
		},
	}

	assert.Equal(t, "event-123", event.ID)
	assert.Equal(t, EventTypeClusterConnected, event.Type)
	assert.Equal(t, "cluster1", event.Cluster)
	assert.Equal(t, "Connected successfully", event.Message)
	assert.Equal(t, 30.0, event.Data["latency"])

	// Test different event types
	event.Type = EventTypeActionExecuted
	assert.Equal(t, EventTypeActionExecuted, event.Type)

	event.Type = EventTypeAnomalyDetected
	assert.Equal(t, EventTypeAnomalyDetected, event.Type)
}

func TestCacheEntry(t *testing.T) {
	now := time.Now()
	entry := &CacheEntry{
		Value:     "test-value",
		ExpiresAt: now.Add(5 * time.Minute),
	}

	assert.Equal(t, "test-value", entry.Value)
	assert.True(t, entry.ExpiresAt.After(now))

	// Test expired entry
	entry.ExpiresAt = now.Add(-1 * time.Minute)
	assert.True(t, entry.ExpiresAt.Before(now))
}

func TestActionType(t *testing.T) {
	tests := []struct {
		action   ActionType
		expected string
	}{
		{ActionTypePurgeDLQ, "purge_dlq"},
		{ActionTypePauseQueue, "pause_queue"},
		{ActionTypeResumeQueue, "resume_queue"},
		{ActionTypeBenchmark, "benchmark"},
		{ActionTypeRebalance, "rebalance"},
		{ActionTypeFailover, "failover"},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.action))
		})
	}
}

func TestActionStatus(t *testing.T) {
	tests := []struct {
		status   ActionStatus
		expected string
	}{
		{ActionStatusPending, "pending"},
		{ActionStatusConfirmed, "confirmed"},
		{ActionStatusExecuting, "executing"},
		{ActionStatusCompleted, "completed"},
		{ActionStatusFailed, "failed"},
		{ActionStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}

func TestEventType(t *testing.T) {
	tests := []struct {
		event    EventType
		expected string
	}{
		{EventTypeClusterConnected, "cluster_connected"},
		{EventTypeClusterDisconnected, "cluster_disconnected"},
		{EventTypeActionExecuted, "action_executed"},
		{EventTypeAnomalyDetected, "anomaly_detected"},
		{EventTypeConfigChanged, "config_changed"},
	}

	for _, tt := range tests {
		t.Run(string(tt.event), func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.event))
		})
	}
}

func TestClusterConfig(t *testing.T) {
	config := ClusterConfig{
		Name:     "test-cluster",
		Label:    "Test Cluster",
		Color:    "#ff0000",
		Endpoint: "localhost:6379",
		Password: "secret",
		DB:       1,
		Enabled:  true,
	}

	assert.Equal(t, "test-cluster", config.Name)
	assert.Equal(t, "Test Cluster", config.Label)
	assert.Equal(t, "#ff0000", config.Color)
	assert.Equal(t, "localhost:6379", config.Endpoint)
	assert.Equal(t, "secret", config.Password)
	assert.Equal(t, 1, config.DB)
	assert.True(t, config.Enabled)
}

func TestWorkerInfo(t *testing.T) {
	worker := WorkerInfo{
		ID:            "worker-123",
		ClusterName:   "prod",
		Status:        "active",
		JobsProcessed: 100,
		LastActivity:  time.Now(),
		Queues:        []string{"high", "normal"},
	}

	assert.Equal(t, "worker-123", worker.ID)
	assert.Equal(t, "prod", worker.ClusterName)
	assert.Equal(t, "active", worker.Status)
	assert.Equal(t, int64(100), worker.JobsProcessed)
	assert.Len(t, worker.Queues, 2)
	assert.Contains(t, worker.Queues, "high")
	assert.Contains(t, worker.Queues, "normal")
}

func TestJobInfo(t *testing.T) {
	now := time.Now()
	job := JobInfo{
		ID:          "job-456",
		ClusterName: "prod",
		Queue:       "high",
		Status:      "completed",
		Payload:     map[string]interface{}{"task": "process_data"},
		CreatedAt:   now,
		ProcessedAt: &now,
		Error:       "",
	}

	assert.Equal(t, "job-456", job.ID)
	assert.Equal(t, "prod", job.ClusterName)
	assert.Equal(t, "high", job.Queue)
	assert.Equal(t, "completed", job.Status)
	assert.Equal(t, "process_data", job.Payload["task"])
	assert.NotNil(t, job.ProcessedAt)
	assert.Empty(t, job.Error)
}