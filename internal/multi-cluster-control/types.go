// Copyright 2025 James Ross
package multicluster

import (
	"context"
	"sync"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/go-redis/redis/v8"
)

// ClusterConfig represents configuration for a single cluster
type ClusterConfig struct {
	Name     string `json:"name"`
	Label    string `json:"label"`
	Color    string `json:"color"`
	Endpoint string `json:"endpoint"`
	Password string `json:"password,omitempty"`
	DB       int    `json:"db"`
	Enabled  bool   `json:"enabled"`
}

// ClusterConnection represents an active connection to a Redis cluster
type ClusterConnection struct {
	Config   ClusterConfig
	Client   *redis.Client
	Status   ConnectionStatus
	LastPing time.Time
	mu       sync.RWMutex
}

// ConnectionStatus represents the status of a cluster connection
type ConnectionStatus struct {
	Connected   bool      `json:"connected"`
	LastError   string    `json:"last_error,omitempty"`
	LastChecked time.Time `json:"last_checked"`
	Latency     float64   `json:"latency_ms"`
}

// ClusterStats represents statistics for a cluster
type ClusterStats struct {
	ClusterName     string           `json:"cluster_name"`
	QueueSizes      map[string]int64 `json:"queue_sizes"`
	ProcessingCount int64            `json:"processing_count"`
	DeadLetterCount int64            `json:"dead_letter_count"`
	WorkerCount     int              `json:"worker_count"`
	JobRate         float64          `json:"job_rate"`
	ErrorRate       float64          `json:"error_rate"`
	Timestamp       time.Time        `json:"timestamp"`
}

// CompareResult represents the result of comparing multiple clusters
type CompareResult struct {
	Clusters   []string                 `json:"clusters"`
	Metrics    map[string]MetricCompare `json:"metrics"`
	Anomalies  []Anomaly                `json:"anomalies"`
	Timestamp  time.Time                `json:"timestamp"`
}

// MetricCompare represents a comparison of a metric across clusters
type MetricCompare struct {
	Name   string             `json:"name"`
	Values map[string]float64 `json:"values"`
	Delta  float64            `json:"delta"`
	Unit   string             `json:"unit"`
}

// Anomaly represents an anomaly detected across clusters
type Anomaly struct {
	Type        string    `json:"type"`
	Cluster     string    `json:"cluster"`
	Description string    `json:"description"`
	Value       float64   `json:"value"`
	Expected    float64   `json:"expected"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
}

// MultiAction represents an action to be applied across multiple clusters
type MultiAction struct {
	ID          string                `json:"id"`
	Type        ActionType            `json:"type"`
	Targets     []string              `json:"targets"`
	Parameters  map[string]interface{} `json:"parameters"`
	Confirmations []ActionConfirmation `json:"confirmations"`
	Status      ActionStatus          `json:"status"`
	Results     map[string]ActionResult `json:"results"`
	CreatedAt   time.Time             `json:"created_at"`
	ExecutedAt  *time.Time            `json:"executed_at,omitempty"`
}

// ActionType represents the type of multi-cluster action
type ActionType string

const (
	ActionTypePurgeDLQ     ActionType = "purge_dlq"
	ActionTypePauseQueue   ActionType = "pause_queue"
	ActionTypeResumeQueue  ActionType = "resume_queue"
	ActionTypeBenchmark    ActionType = "benchmark"
	ActionTypeRebalance    ActionType = "rebalance"
	ActionTypeFailover     ActionType = "failover"
)

// ActionStatus represents the status of a multi-cluster action
type ActionStatus string

const (
	ActionStatusPending   ActionStatus = "pending"
	ActionStatusConfirmed ActionStatus = "confirmed"
	ActionStatusExecuting ActionStatus = "executing"
	ActionStatusCompleted ActionStatus = "completed"
	ActionStatusFailed    ActionStatus = "failed"
	ActionStatusCancelled ActionStatus = "cancelled"
)

// ActionConfirmation represents a confirmation requirement for an action
type ActionConfirmation struct {
	Required    bool      `json:"required"`
	Message     string    `json:"message"`
	ConfirmedBy string    `json:"confirmed_by,omitempty"`
	ConfirmedAt time.Time `json:"confirmed_at,omitempty"`
}

// ActionResult represents the result of an action on a specific cluster
type ActionResult struct {
	Success   bool      `json:"success"`
	Message   string    `json:"message,omitempty"`
	Error     string    `json:"error,omitempty"`
	Duration  float64   `json:"duration_ms"`
	Timestamp time.Time `json:"timestamp"`
}

// CacheEntry represents a cached value with TTL
type CacheEntry struct {
	Value     interface{}
	ExpiresAt time.Time
}

// ClusterCache represents a cache for cluster data
type ClusterCache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
}

// HealthStatus represents the health status of a cluster
type HealthStatus struct {
	Healthy     bool              `json:"healthy"`
	Issues      []string          `json:"issues,omitempty"`
	Metrics     map[string]float64 `json:"metrics"`
	LastChecked time.Time         `json:"last_checked"`
}

// TabConfig represents the configuration for cluster tabs in the TUI
type TabConfig struct {
	Tabs         []TabInfo `json:"tabs"`
	ActiveTab    int       `json:"active_tab"`
	CompareMode  bool      `json:"compare_mode"`
	CompareWith  []string  `json:"compare_with,omitempty"`
}

// TabInfo represents information about a single tab
type TabInfo struct {
	Index       int    `json:"index"`
	ClusterName string `json:"cluster_name"`
	Label       string `json:"label"`
	Color       string `json:"color"`
	Shortcut    string `json:"shortcut"`
}

// Event represents an event in the multi-cluster system
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Cluster   string                 `json:"cluster"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventType represents the type of event
type EventType string

const (
	EventTypeClusterConnected    EventType = "cluster_connected"
	EventTypeClusterDisconnected EventType = "cluster_disconnected"
	EventTypeActionExecuted      EventType = "action_executed"
	EventTypeAnomalyDetected    EventType = "anomaly_detected"
	EventTypeConfigChanged      EventType = "config_changed"
)

// Manager interface defines the multi-cluster control operations
type Manager interface {
	// Cluster management
	AddCluster(ctx context.Context, config ClusterConfig) error
	RemoveCluster(ctx context.Context, name string) error
	ListClusters(ctx context.Context) ([]ClusterConfig, error)
	GetCluster(ctx context.Context, name string) (*ClusterConnection, error)
	SwitchCluster(ctx context.Context, name string) error

	// Monitoring and stats
	GetStats(ctx context.Context, clusterName string) (*ClusterStats, error)
	GetAllStats(ctx context.Context) (map[string]*ClusterStats, error)
	CompareClusters(ctx context.Context, clusters []string) (*CompareResult, error)
	GetHealth(ctx context.Context, clusterName string) (*HealthStatus, error)

	// Multi-cluster actions
	ExecuteAction(ctx context.Context, action *MultiAction) error
	ConfirmAction(ctx context.Context, actionID string, confirmedBy string) error
	CancelAction(ctx context.Context, actionID string) error
	GetActionStatus(ctx context.Context, actionID string) (*MultiAction, error)

	// TUI integration
	GetTabConfig(ctx context.Context) (*TabConfig, error)
	SetCompareMode(ctx context.Context, enabled bool, clusters []string) error

	// Event streaming
	SubscribeEvents(ctx context.Context) (<-chan Event, error)
	UnsubscribeEvents(ctx context.Context, ch <-chan Event) error
}

// WorkerInfo represents information about a worker
type WorkerInfo struct {
	ID          string    `json:"id"`
	ClusterName string    `json:"cluster_name"`
	Status      string    `json:"status"`
	JobsProcessed int64   `json:"jobs_processed"`
	LastActivity time.Time `json:"last_activity"`
	Queues      []string  `json:"queues"`
}

// JobInfo represents information about a job across clusters
type JobInfo struct {
	ID          string                 `json:"id"`
	ClusterName string                 `json:"cluster_name"`
	Queue       string                 `json:"queue"`
	Status      string                 `json:"status"`
	Payload     map[string]interface{} `json:"payload"`
	CreatedAt   time.Time              `json:"created_at"`
	ProcessedAt *time.Time             `json:"processed_at,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// CompareView represents the side-by-side comparison view data
type CompareView struct {
	Left     ClusterViewData `json:"left"`
	Right    ClusterViewData `json:"right"`
	Deltas   map[string]Delta `json:"deltas"`
	Updated  time.Time       `json:"updated"`
}

// ClusterViewData represents data for a single cluster in compare view
type ClusterViewData struct {
	Name            string           `json:"name"`
	Stats           *ClusterStats    `json:"stats"`
	Health          *HealthStatus    `json:"health"`
	RecentJobs      []JobInfo        `json:"recent_jobs"`
	ActiveWorkers   []WorkerInfo     `json:"active_workers"`
}

// Delta represents the difference between two values
type Delta struct {
	Left       float64 `json:"left"`
	Right      float64 `json:"right"`
	Difference float64 `json:"difference"`
	Percentage float64 `json:"percentage"`
	Direction  string  `json:"direction"` // "up", "down", "equal"
}

// PollingConfig represents the configuration for cluster polling
type PollingConfig struct {
	Interval time.Duration `json:"interval"`
	Jitter   time.Duration `json:"jitter"`
	Timeout  time.Duration `json:"timeout"`
	Enabled  bool          `json:"enabled"`
}

// MetricsCollector represents a collector for cluster metrics
type MetricsCollector interface {
	CollectMetrics(ctx context.Context, cluster *ClusterConnection) (*ClusterStats, error)
	CollectHealth(ctx context.Context, cluster *ClusterConnection) (*HealthStatus, error)
}

// ActionExecutor represents an executor for multi-cluster actions
type ActionExecutor interface {
	Execute(ctx context.Context, action *MultiAction, clusters map[string]*ClusterConnection) error
	Validate(ctx context.Context, action *MultiAction) error
}

// AnomalyDetector represents a detector for cluster anomalies
type AnomalyDetector interface {
	Detect(ctx context.Context, stats map[string]*ClusterStats) ([]Anomaly, error)
	Configure(thresholds map[string]float64) error
}