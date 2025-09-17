// Copyright 2025 James Ross
package multicluster

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/admin"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// ManagerImpl implements the Manager interface
type ManagerImpl struct {
	config          *Config
	connections     map[string]*ClusterConnection
	cache           *ClusterCache
	logger          *zap.Logger
	events          chan Event
	subscribers     []chan Event
	activeTab       string
	compareMode     bool
	compareClusters []string

	mu     sync.RWMutex
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewManager creates a new multi-cluster manager
func NewManager(cfg *Config, logger *zap.Logger) (*ManagerImpl, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	if logger == nil {
		logger = zap.NewNop()
	}

	m := &ManagerImpl{
		config:      cfg,
		connections: make(map[string]*ClusterConnection),
		cache:       NewClusterCache(),
		logger:      logger,
		events:      make(chan Event, 100),
		subscribers: make([]chan Event, 0),
		stopCh:      make(chan struct{}),
		activeTab:   cfg.DefaultCluster,
	}

	// Initialize connections for enabled clusters
	for _, clusterCfg := range cfg.GetEnabledClusters() {
		if err := m.initConnection(clusterCfg); err != nil {
			m.logger.Warn("Failed to initialize cluster connection",
				zap.String("cluster", clusterCfg.Name),
				zap.Error(err))
		}
	}

	// Start background tasks
	m.startBackgroundTasks()

	return m, nil
}

// initConnection initializes a connection to a cluster
func (m *ManagerImpl) initConnection(cfg ClusterConfig) error {
	opts := &redis.Options{
		Addr:     cfg.Endpoint,
		Password: cfg.Password,
		DB:       cfg.DB,
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	start := time.Now()
	if err := client.Ping(ctx).Err(); err != nil {
		return NewConnectionError(cfg.Name, cfg.Endpoint, err)
	}
	latency := time.Since(start).Milliseconds()

	conn := &ClusterConnection{
		Config:   cfg,
		Client:   client,
		LastPing: time.Now(),
		Status: ConnectionStatus{
			Connected:   true,
			LastChecked: time.Now(),
			Latency:     float64(latency),
		},
	}

	m.mu.Lock()
	m.connections[cfg.Name] = conn
	m.mu.Unlock()

	// Emit event
	m.emitEvent(Event{
		Type:      EventTypeClusterConnected,
		Cluster:   cfg.Name,
		Message:   fmt.Sprintf("Connected to cluster %s", cfg.Name),
		Timestamp: time.Now(),
	})

	return nil
}

// startBackgroundTasks starts background tasks for polling and cleanup
func (m *ManagerImpl) startBackgroundTasks() {
	// Start polling task
	if m.config.Polling.Enabled {
		m.wg.Add(1)
		go m.pollClusters()
	}

	// Start cache cleanup task
	if m.config.Cache.Enabled {
		m.wg.Add(1)
		go m.cleanupCache()
	}

	// Start health check task
	m.wg.Add(1)
	go m.healthCheck()
}

// pollClusters polls clusters for stats
func (m *ManagerImpl) pollClusters() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.Polling.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.mu.RLock()
			connections := make([]*ClusterConnection, 0, len(m.connections))
			for _, conn := range m.connections {
				connections = append(connections, conn)
			}
			m.mu.RUnlock()

			for _, conn := range connections {
				// Add jitter to avoid thundering herd
				jitter := time.Duration(float64(m.config.Polling.Jitter) * (0.5 - float64(time.Now().UnixNano()%1000)/1000))
				time.Sleep(jitter)

				ctx, cancel := context.WithTimeout(context.Background(), m.config.Polling.Timeout)
				stats, err := m.collectStats(ctx, conn)
				cancel()

				if err != nil {
					m.logger.Warn("Failed to collect stats",
						zap.String("cluster", conn.Config.Name),
						zap.Error(err))
					continue
				}

				// Cache stats
				if m.config.Cache.Enabled {
					m.cache.Set(fmt.Sprintf("stats:%s", conn.Config.Name), stats, m.config.Cache.TTL)
				}
			}
		}
	}
}

// cleanupCache periodically cleans up expired cache entries
func (m *ManagerImpl) cleanupCache() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.Cache.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.cache.Cleanup()
		}
	}
}

// healthCheck periodically checks cluster health
func (m *ManagerImpl) healthCheck() {
	defer m.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			m.mu.RLock()
			connections := make([]*ClusterConnection, 0, len(m.connections))
			for _, conn := range m.connections {
				connections = append(connections, conn)
			}
			m.mu.RUnlock()

			for _, conn := range connections {
				ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

				start := time.Now()
				err := conn.Client.Ping(ctx).Err()
				latency := time.Since(start).Milliseconds()

				conn.mu.Lock()
				conn.LastPing = time.Now()
				conn.Status.LastChecked = time.Now()
				conn.Status.Latency = float64(latency)

				if err != nil {
					conn.Status.Connected = false
					conn.Status.LastError = err.Error()

					m.emitEvent(Event{
						Type:      EventTypeClusterDisconnected,
						Cluster:   conn.Config.Name,
						Message:   fmt.Sprintf("Lost connection to cluster %s: %v", conn.Config.Name, err),
						Timestamp: time.Now(),
					})
				} else {
					wasDisconnected := !conn.Status.Connected
					conn.Status.Connected = true
					conn.Status.LastError = ""

					if wasDisconnected {
						m.emitEvent(Event{
							Type:      EventTypeClusterConnected,
							Cluster:   conn.Config.Name,
							Message:   fmt.Sprintf("Reconnected to cluster %s", conn.Config.Name),
							Timestamp: time.Now(),
						})
					}
				}
				conn.mu.Unlock()

				cancel()
			}
		}
	}
}

// collectStats collects statistics from a cluster
func (m *ManagerImpl) collectStats(ctx context.Context, conn *ClusterConnection) (*ClusterStats, error) {
	// Create a temporary config for the admin.Stats call
	tempCfg := &config.Config{
		Worker: config.Worker{
			Queues: map[string]string{
				"default": "jobqueue:queue:default",
				"high":    "jobqueue:queue:high",
				"low":     "jobqueue:queue:low",
			},
			CompletedList:  "jobqueue:completed",
			DeadLetterList: "jobqueue:dead_letter",
		},
	}

	statsResult, err := admin.Stats(ctx, tempCfg, conn.Client)
	if err != nil {
		return nil, err
	}

	stats := &ClusterStats{
		ClusterName:     conn.Config.Name,
		QueueSizes:      statsResult.Queues,
		ProcessingCount: 0,
		DeadLetterCount: 0,
		WorkerCount:     int(statsResult.Heartbeats),
		Timestamp:       time.Now(),
	}

	// Calculate processing count
	for _, count := range statsResult.ProcessingLists {
		stats.ProcessingCount += count
	}

	// Get dead letter count
	if dlCount, ok := statsResult.Queues["dead_letter(jobqueue:dead_letter)"]; ok {
		stats.DeadLetterCount = dlCount
	}

	// Calculate job rate (simplified - would need historical data for accurate rate)
	// This is a placeholder implementation
	stats.JobRate = 0.0
	stats.ErrorRate = 0.0

	return stats, nil
}

// AddCluster adds a new cluster
func (m *ManagerImpl) AddCluster(ctx context.Context, cfg ClusterConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.connections[cfg.Name]; exists {
		return ErrClusterAlreadyExists
	}

	// Add to config
	if err := m.config.AddCluster(cfg); err != nil {
		return err
	}

	// Initialize connection if enabled
	if cfg.Enabled {
		if err := m.initConnection(cfg); err != nil {
			// Remove from config if connection fails
			m.config.RemoveCluster(cfg.Name)
			return err
		}
	}

	return nil
}

// RemoveCluster removes a cluster
func (m *ManagerImpl) RemoveCluster(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, exists := m.connections[name]
	if !exists {
		return NewClusterError(name, "remove", ErrClusterNotFound)
	}

	// Close connection
	if conn.Client != nil {
		conn.Client.Close()
	}

	delete(m.connections, name)

	// Remove from config
	return m.config.RemoveCluster(name)
}

// ListClusters lists all configured clusters
func (m *ManagerImpl) ListClusters(ctx context.Context) ([]ClusterConfig, error) {
	return m.config.Clusters, nil
}

// GetCluster gets a specific cluster connection
func (m *ManagerImpl) GetCluster(ctx context.Context, name string) (*ClusterConnection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conn, exists := m.connections[name]
	if !exists {
		return nil, NewClusterError(name, "get", ErrClusterNotFound)
	}

	return conn, nil
}

// SwitchCluster switches the active cluster
func (m *ManagerImpl) SwitchCluster(ctx context.Context, name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.connections[name]; !exists {
		return NewClusterError(name, "switch", ErrClusterNotFound)
	}

	m.activeTab = name
	return nil
}

// GetStats gets statistics for a cluster
func (m *ManagerImpl) GetStats(ctx context.Context, clusterName string) (*ClusterStats, error) {
	// Check cache first
	if m.config.Cache.Enabled {
		if cached, ok := m.cache.Get(fmt.Sprintf("stats:%s", clusterName)); ok {
			if stats, ok := cached.(*ClusterStats); ok {
				return stats, nil
			}
		}
	}

	conn, err := m.GetCluster(ctx, clusterName)
	if err != nil {
		return nil, err
	}

	stats, err := m.collectStats(ctx, conn)
	if err != nil {
		return nil, NewClusterError(clusterName, "stats", err)
	}

	// Cache the stats
	if m.config.Cache.Enabled {
		m.cache.Set(fmt.Sprintf("stats:%s", clusterName), stats, m.config.Cache.TTL)
	}

	return stats, nil
}

// GetAllStats gets statistics for all clusters
func (m *ManagerImpl) GetAllStats(ctx context.Context) (map[string]*ClusterStats, error) {
	m.mu.RLock()
	clusterNames := make([]string, 0, len(m.connections))
	for name := range m.connections {
		clusterNames = append(clusterNames, name)
	}
	m.mu.RUnlock()

	results := make(map[string]*ClusterStats)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, name := range clusterNames {
		wg.Add(1)
		go func(clusterName string) {
			defer wg.Done()

			stats, err := m.GetStats(ctx, clusterName)
			if err != nil {
				m.logger.Warn("Failed to get stats",
					zap.String("cluster", clusterName),
					zap.Error(err))
				return
			}

			mu.Lock()
			results[clusterName] = stats
			mu.Unlock()
		}(name)
	}

	wg.Wait()
	return results, nil
}

// CompareClusters compares statistics across clusters
func (m *ManagerImpl) CompareClusters(ctx context.Context, clusters []string) (*CompareResult, error) {
	if len(clusters) < 2 {
		return nil, ErrInsufficientClusters
	}

	// Get stats for all requested clusters
	allStats := make(map[string]*ClusterStats)
	for _, cluster := range clusters {
		stats, err := m.GetStats(ctx, cluster)
		if err != nil {
			return nil, err
		}
		allStats[cluster] = stats
	}

	// Build comparison
	result := &CompareResult{
		Clusters:  clusters,
		Metrics:   make(map[string]MetricCompare),
		Anomalies: []Anomaly{},
		Timestamp: time.Now(),
	}

	// Compare key metrics
	metrics := []string{"queue_size", "processing_count", "dead_letter_count", "worker_count", "job_rate", "error_rate"}

	for _, metric := range metrics {
		compare := MetricCompare{
			Name:   metric,
			Values: make(map[string]float64),
		}

		var min, max float64
		first := true

		for cluster, stats := range allStats {
			var value float64
			switch metric {
			case "queue_size":
				for _, size := range stats.QueueSizes {
					value += float64(size)
				}
			case "processing_count":
				value = float64(stats.ProcessingCount)
			case "dead_letter_count":
				value = float64(stats.DeadLetterCount)
			case "worker_count":
				value = float64(stats.WorkerCount)
			case "job_rate":
				value = stats.JobRate
			case "error_rate":
				value = stats.ErrorRate
			}

			compare.Values[cluster] = value

			if first {
				min = value
				max = value
				first = false
			} else {
				if value < min {
					min = value
				}
				if value > max {
					max = value
				}
			}
		}

		compare.Delta = max - min
		result.Metrics[metric] = compare

		// Detect anomalies (simplified)
		if compare.Delta > 0 {
			avg := 0.0
			for _, v := range compare.Values {
				avg += v
			}
			avg /= float64(len(compare.Values))

			for cluster, value := range compare.Values {
				deviation := math.Abs(value - avg)
				if deviation > avg*m.config.CompareMode.DeltaThreshold/100 {
					result.Anomalies = append(result.Anomalies, Anomaly{
						Type:        "deviation",
						Cluster:     cluster,
						Description: fmt.Sprintf("%s deviates significantly from average", metric),
						Value:       value,
						Expected:    avg,
						Severity:    "warning",
						Timestamp:   time.Now(),
					})
				}
			}
		}
	}

	// Emit anomaly events
	for _, anomaly := range result.Anomalies {
		m.emitEvent(Event{
			Type:    EventTypeAnomalyDetected,
			Cluster: anomaly.Cluster,
			Message: anomaly.Description,
			Data: map[string]interface{}{
				"anomaly": anomaly,
			},
			Timestamp: time.Now(),
		})
	}

	return result, nil
}

// GetHealth gets health status for a cluster
func (m *ManagerImpl) GetHealth(ctx context.Context, clusterName string) (*HealthStatus, error) {
	conn, err := m.GetCluster(ctx, clusterName)
	if err != nil {
		return nil, err
	}

	health := &HealthStatus{
		Healthy:     true,
		Issues:      []string{},
		Metrics:     make(map[string]float64),
		LastChecked: time.Now(),
	}

	// Check connection status
	conn.mu.RLock()
	if !conn.Status.Connected {
		health.Healthy = false
		health.Issues = append(health.Issues, "Cluster is disconnected")
	}
	health.Metrics["latency_ms"] = conn.Status.Latency
	conn.mu.RUnlock()

	// Get stats and check for issues
	stats, err := m.GetStats(ctx, clusterName)
	if err != nil {
		health.Healthy = false
		health.Issues = append(health.Issues, fmt.Sprintf("Failed to get stats: %v", err))
	} else {
		// Check for high dead letter count
		if stats.DeadLetterCount > 100 {
			health.Issues = append(health.Issues, fmt.Sprintf("High dead letter count: %d", stats.DeadLetterCount))
		}

		// Check for no workers
		if stats.WorkerCount == 0 {
			health.Healthy = false
			health.Issues = append(health.Issues, "No active workers")
		}

		// Add metrics
		health.Metrics["worker_count"] = float64(stats.WorkerCount)
		health.Metrics["dead_letter_count"] = float64(stats.DeadLetterCount)
		health.Metrics["processing_count"] = float64(stats.ProcessingCount)
	}

	return health, nil
}

// ExecuteAction executes a multi-cluster action
func (m *ManagerImpl) ExecuteAction(ctx context.Context, action *MultiAction) error {
	// Validate action
	if !m.config.IsActionAllowed(action.Type) {
		return NewActionError(action.ID, action.Type, "", "validation", ErrActionNotAllowed)
	}

	// Check confirmation if required
	if m.config.Actions.RequireConfirmation && action.Status != ActionStatusConfirmed {
		return NewActionError(action.ID, action.Type, "", "confirmation", ErrConfirmationRequired)
	}

	action.Status = ActionStatusExecuting
	action.ExecutedAt = &[]time.Time{time.Now()}[0]

	// Execute on each target cluster
	var wg sync.WaitGroup
	results := make(map[string]ActionResult)
	var mu sync.Mutex

	for _, target := range action.Targets {
		wg.Add(1)
		go func(clusterName string) {
			defer wg.Done()

			start := time.Now()
			err := m.executeActionOnCluster(ctx, action, clusterName)
			duration := time.Since(start).Milliseconds()

			result := ActionResult{
				Success:   err == nil,
				Duration:  float64(duration),
				Timestamp: time.Now(),
			}

			if err != nil {
				result.Error = err.Error()
			} else {
				result.Message = "Action completed successfully"
			}

			mu.Lock()
			results[clusterName] = result
			mu.Unlock()
		}(target)
	}

	wg.Wait()

	action.Results = results
	action.Status = ActionStatusCompleted

	// Check if any failed
	for _, result := range results {
		if !result.Success {
			action.Status = ActionStatusFailed
			break
		}
	}

	// Emit event
	m.emitEvent(Event{
		Type:    EventTypeActionExecuted,
		Message: fmt.Sprintf("Action %s (%s) executed on %d clusters", action.ID, action.Type, len(action.Targets)),
		Data: map[string]interface{}{
			"action": action,
		},
		Timestamp: time.Now(),
	})

	return nil
}

// executeActionOnCluster executes an action on a specific cluster
func (m *ManagerImpl) executeActionOnCluster(ctx context.Context, action *MultiAction, clusterName string) error {
	conn, err := m.GetCluster(ctx, clusterName)
	if err != nil {
		return err
	}

	// Apply timeout
	timeout := m.config.GetActionTimeout(action.Type)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	switch action.Type {
	case ActionTypePurgeDLQ:
		return conn.Client.Del(ctx, "jobqueue:dead_letter").Err()

	case ActionTypePauseQueue:
		queueName, _ := action.Parameters["queue"].(string)
		if queueName == "" {
			return fmt.Errorf("queue name required")
		}
		// Implementation would depend on your pause mechanism
		return conn.Client.Set(ctx, fmt.Sprintf("jobqueue:paused:%s", queueName), "1", 0).Err()

	case ActionTypeResumeQueue:
		queueName, _ := action.Parameters["queue"].(string)
		if queueName == "" {
			return fmt.Errorf("queue name required")
		}
		return conn.Client.Del(ctx, fmt.Sprintf("jobqueue:paused:%s", queueName)).Err()

	case ActionTypeBenchmark:
		// Simple benchmark - ping the server multiple times
		iterations := 10
		if n, ok := action.Parameters["iterations"].(float64); ok {
			iterations = int(n)
		}

		for i := 0; i < iterations; i++ {
			if err := conn.Client.Ping(ctx).Err(); err != nil {
				return fmt.Errorf("benchmark failed at iteration %d: %w", i, err)
			}
		}
		return nil

	default:
		return fmt.Errorf("unsupported action type: %s", action.Type)
	}
}

// ConfirmAction confirms a pending action
func (m *ManagerImpl) ConfirmAction(ctx context.Context, actionID string, confirmedBy string) error {
	// In a real implementation, you would store actions and retrieve them
	// For now, this is a placeholder
	return nil
}

// CancelAction cancels a pending action
func (m *ManagerImpl) CancelAction(ctx context.Context, actionID string) error {
	// Placeholder implementation
	return nil
}

// GetActionStatus gets the status of an action
func (m *ManagerImpl) GetActionStatus(ctx context.Context, actionID string) (*MultiAction, error) {
	// Placeholder implementation
	return nil, fmt.Errorf("action %s not found", actionID)
}

// GetTabConfig gets the tab configuration for the TUI
func (m *ManagerImpl) GetTabConfig(ctx context.Context) (*TabConfig, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tabs := make([]TabInfo, 0, len(m.connections))
	index := 1

	// Sort clusters by name for consistent ordering
	var names []string
	for name := range m.connections {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		conn := m.connections[name]
		shortcut := ""
		if index <= 9 {
			shortcut = fmt.Sprintf("%d", index)
		}

		tabs = append(tabs, TabInfo{
			Index:       index,
			ClusterName: name,
			Label:       conn.Config.Label,
			Color:       conn.Config.Color,
			Shortcut:    shortcut,
		})
		index++
	}

	activeIndex := 0
	for i, tab := range tabs {
		if tab.ClusterName == m.activeTab {
			activeIndex = i
			break
		}
	}

	return &TabConfig{
		Tabs:        tabs,
		ActiveTab:   activeIndex,
		CompareMode: m.compareMode,
		CompareWith: m.compareClusters,
	}, nil
}

// SetCompareMode sets the compare mode
func (m *ManagerImpl) SetCompareMode(ctx context.Context, enabled bool, clusters []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if enabled && len(clusters) < 2 {
		return ErrInsufficientClusters
	}

	m.compareMode = enabled
	m.compareClusters = clusters

	return nil
}

// SubscribeEvents subscribes to events
func (m *ManagerImpl) SubscribeEvents(ctx context.Context) (<-chan Event, error) {
	ch := make(chan Event, 100)

	m.mu.Lock()
	m.subscribers = append(m.subscribers, ch)
	m.mu.Unlock()

	return ch, nil
}

// UnsubscribeEvents unsubscribes from events
func (m *ManagerImpl) UnsubscribeEvents(ctx context.Context, ch <-chan Event) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, sub := range m.subscribers {
		if sub == ch {
			m.subscribers = append(m.subscribers[:i], m.subscribers[i+1:]...)
			close(sub)
			return nil
		}
	}

	return fmt.Errorf("subscriber not found")
}

// emitEvent emits an event to all subscribers
func (m *ManagerImpl) emitEvent(event Event) {
	m.mu.RLock()
	subscribers := make([]chan Event, len(m.subscribers))
	copy(subscribers, m.subscribers)
	m.mu.RUnlock()

	for _, ch := range subscribers {
		select {
		case ch <- event:
		default:
			// Channel full, skip
		}
	}
}

// Close closes the manager and all connections
func (m *ManagerImpl) Close() error {
	close(m.stopCh)
	m.wg.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Close all connections
	for _, conn := range m.connections {
		if conn.Client != nil {
			conn.Client.Close()
		}
	}

	// Close event channels
	for _, ch := range m.subscribers {
		close(ch)
	}

	return nil
}

// NewClusterCache creates a new cluster cache
func NewClusterCache() *ClusterCache {
	return &ClusterCache{
		entries: make(map[string]*CacheEntry),
	}
}

// Get gets a value from the cache
func (c *ClusterCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil, false
	}

	return entry.Value, true
}

// Set sets a value in the cache
func (c *ClusterCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Delete deletes a value from the cache
func (c *ClusterCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

// Cleanup removes expired entries
func (c *ClusterCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.ExpiresAt) {
			delete(c.entries, key)
		}
	}
}
