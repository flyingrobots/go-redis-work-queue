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

	assert.True(t, status.IsHealthy())

	// Test unhealthy states
	status.Connected = false
	assert.False(t, status.IsHealthy())

	status.Connected = true
	status.Latency = 2000 // High latency
	assert.False(t, status.IsHealthy())

	status.Latency = 45.2
	status.LastError = "connection timeout"
	assert.False(t, status.IsHealthy())
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

	assert.Equal(t, "test", conn.GetName())
	assert.Equal(t, "localhost:6379", conn.GetEndpoint())
	assert.True(t, conn.IsConnected())
	assert.Equal(t, 30.0, conn.GetLatency())
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

	assert.Equal(t, 30, stats.GetTotalJobs())
	assert.Equal(t, 100.5, stats.GetJobRate())
	assert.Equal(t, 0.5, stats.GetErrorRate())
	assert.True(t, stats.IsHealthy())

	// Test unhealthy state
	stats.ErrorRate = 10.0 // High error rate
	assert.False(t, stats.IsHealthy())

	stats.ErrorRate = 0.5
	stats.WorkerCount = 0 // No workers
	assert.False(t, stats.IsHealthy())
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
	}

	min, max := compare.GetMinMax()
	assert.Equal(t, 80.0, min)
	assert.Equal(t, 150.0, max)

	avg := compare.GetAverage()
	assert.InDelta(t, 110.0, avg, 0.001)

	deviation := compare.GetStandardDeviation()
	assert.Greater(t, deviation, 0.0)
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

	assert.Len(t, result.GetAnomaliesBySeverity("warning"), 1)
	assert.Len(t, result.GetAnomaliesBySeverity("critical"), 0)

	topAnomalies := result.GetTopAnomalies(5)
	assert.Len(t, topAnomalies, 1)

	summary := result.GetSummary()
	assert.Contains(t, summary, "2 clusters")
	assert.Contains(t, summary, "1 metrics")
	assert.Contains(t, summary, "1 anomalies")
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

	assert.False(t, action.IsCompleted())
	assert.False(t, action.IsSuccessful())
	assert.Equal(t, time.Duration(0), action.GetDuration())

	// Simulate completion
	action.Status = ActionStatusCompleted
	executed := time.Now()
	action.ExecutedAt = &executed
	action.Results = map[string]ActionResult{
		"cluster1": {Success: true, Duration: 1000},
		"cluster2": {Success: true, Duration: 1500},
	}

	assert.True(t, action.IsCompleted())
	assert.True(t, action.IsSuccessful())
	assert.Greater(t, action.GetDuration(), time.Duration(0))

	summary := action.GetSummary()
	assert.Contains(t, summary, "test-action")
	assert.Contains(t, summary, "completed")
	assert.Contains(t, summary, "2/2 successful")
}

func TestActionResult(t *testing.T) {
	result := ActionResult{
		Success:   true,
		Message:   "Operation completed",
		Duration:  1500.0,
		Timestamp: time.Now(),
	}

	assert.True(t, result.IsSuccess())
	assert.Equal(t, "Operation completed", result.GetMessage())
	assert.Equal(t, 1500.0, result.GetDuration())

	// Test failed result
	result.Success = false
	result.Error = "Connection failed"
	assert.False(t, result.IsSuccess())
	assert.Equal(t, "Connection failed", result.GetMessage())
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

	assert.True(t, health.IsHealthy())
	assert.Equal(t, "healthy", health.GetStatus())
	assert.Len(t, health.GetCriticalIssues(), 0)

	// Add issues
	health.Issues = []string{
		"High latency detected",
		"Critical: No workers available",
		"Warning: Dead letter queue growing",
	}
	health.Healthy = false

	assert.False(t, health.IsHealthy())
	assert.Equal(t, "unhealthy", health.GetStatus())
	criticalIssues := health.GetCriticalIssues()
	assert.Len(t, criticalIssues, 1)
	assert.Contains(t, criticalIssues[0], "Critical:")

	score := health.GetHealthScore()
	assert.Less(t, score, 1.0) // Should be less than perfect
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

	assert.True(t, anomaly.IsWarning())
	assert.False(t, anomaly.IsCritical())
	assert.Equal(t, 50.0, anomaly.GetDeviation())
	assert.InDelta(t, 50.0, anomaly.GetDeviationPercent(), 0.001)

	details := anomaly.GetDetails()
	assert.Contains(t, details, "deviation")
	assert.Contains(t, details, "cluster1")
	assert.Contains(t, details, "150")
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

	assert.Len(t, config.GetEnabledTabs(), 2)
	assert.Equal(t, "cluster1", config.GetActiveCluster())
	assert.False(t, config.IsCompareMode())

	// Enable compare mode
	config.CompareMode = true
	config.CompareWith = []string{"cluster1", "cluster2"}
	assert.True(t, config.IsCompareMode())
	assert.Len(t, config.GetCompareClusters(), 2)
}

func TestTabInfo(t *testing.T) {
	tab := TabInfo{
		Index:       1,
		ClusterName: "production",
		Label:       "Production",
		Color:       "#ff0000",
		Shortcut:    "1",
		Health:      "healthy",
		Stats: map[string]interface{}{
			"workers": 5,
			"queues":  3,
		},
	}

	assert.Equal(t, "Production", tab.GetDisplayName())
	assert.Equal(t, "#ff0000", tab.GetColor())
	assert.True(t, tab.HasShortcut())
	assert.Equal(t, "1", tab.GetShortcut())
	assert.True(t, tab.IsHealthy())

	formatted := tab.FormatLabel()
	assert.Contains(t, formatted, "Production")
	assert.Contains(t, formatted, "1")
}

func TestEvent(t *testing.T) {
	event := Event{
		Type:      EventTypeClusterConnected,
		Cluster:   "cluster1",
		Message:   "Connected successfully",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"latency": 30.0,
		},
	}

	assert.True(t, event.IsConnection())
	assert.False(t, event.IsAction())
	assert.False(t, event.IsAnomaly())
	assert.Equal(t, "cluster1", event.GetCluster())
	assert.Contains(t, event.GetMessage(), "Connected")

	formatted := event.Format()
	assert.Contains(t, formatted, "cluster1")
	assert.Contains(t, formatted, "Connected")

	// Test different event types
	event.Type = EventTypeActionExecuted
	assert.True(t, event.IsAction())

	event.Type = EventTypeAnomalyDetected
	assert.True(t, event.IsAnomaly())
}

func TestCacheEntry(t *testing.T) {
	now := time.Now()
	entry := &CacheEntry{
		Value:     "test-value",
		ExpiresAt: now.Add(5 * time.Minute),
	}

	assert.False(t, entry.IsExpired())
	assert.True(t, entry.IsValid())

	// Test expired entry
	entry.ExpiresAt = now.Add(-1 * time.Minute)
	assert.True(t, entry.IsExpired())
	assert.False(t, entry.IsValid())

	remaining := entry.GetRemainingTTL()
	assert.Less(t, remaining, time.Duration(0))
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
			assert.True(t, tt.action.IsValid())
		})
	}

	// Test invalid action type
	invalid := ActionType("invalid")
	assert.False(t, invalid.IsValid())
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
			assert.True(t, tt.status.IsValid())
		})
	}

	// Test status transitions
	assert.True(t, ActionStatusPending.CanTransitionTo(ActionStatusConfirmed))
	assert.True(t, ActionStatusConfirmed.CanTransitionTo(ActionStatusExecuting))
	assert.True(t, ActionStatusExecuting.CanTransitionTo(ActionStatusCompleted))
	assert.False(t, ActionStatusCompleted.CanTransitionTo(ActionStatusPending))
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
			assert.True(t, tt.event.IsValid())
		})
	}

	// Test event categorization
	assert.True(t, EventTypeClusterConnected.IsConnectionEvent())
	assert.True(t, EventTypeActionExecuted.IsActionEvent())
	assert.True(t, EventTypeAnomalyDetected.IsAnomalyEvent())
	assert.False(t, EventTypeClusterConnected.IsActionEvent())
}