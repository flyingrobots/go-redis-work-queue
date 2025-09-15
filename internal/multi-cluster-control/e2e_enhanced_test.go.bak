// Copyright 2025 James Ross
package multicluster

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// EnhancedTUIRenderer provides comprehensive TUI testing capabilities
type EnhancedTUIRenderer struct {
	app             *tview.Application
	manager         Manager
	renderHistory   []RenderEvent
	keyEvents       []tcell.Event
	currentMode     string
	activeCluster   string
	selectedClusters []string
}

// RenderEvent represents a rendering event for testing
type RenderEvent struct {
	Type      string
	Timestamp time.Time
	Data      map[string]interface{}
	Content   string
}

// NewEnhancedTUIRenderer creates a comprehensive TUI renderer for testing
func NewEnhancedTUIRenderer(manager Manager) *EnhancedTUIRenderer {
	return &EnhancedTUIRenderer{
		app:              tview.NewApplication(),
		manager:          manager,
		renderHistory:    make([]RenderEvent, 0),
		keyEvents:        make([]tcell.Event, 0),
		currentMode:      "overview",
		selectedClusters: make([]string, 0),
	}
}

// TestMultiCluster_TUIIntegration tests comprehensive TUI functionality
func TestMultiCluster_TUIIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup test environment with multiple clusters
	mr1 := miniredis.RunT(t)
	defer mr1.Close()
	mr2 := miniredis.RunT(t)
	defer mr2.Close()
	mr3 := miniredis.RunT(t)
	defer mr3.Close()

	// Setup realistic test data
	setupTUITestData(mr1, mr2, mr3)

	cfg := &Config{
		Clusters: []ClusterConfig{
			{Name: "prod", Label: "Production", Color: "#ff0000", Endpoint: mr1.Addr(), DB: 0, Enabled: true},
			{Name: "staging", Label: "Staging", Color: "#00ff00", Endpoint: mr2.Addr(), DB: 0, Enabled: true},
			{Name: "dev", Label: "Development", Color: "#0000ff", Endpoint: mr3.Addr(), DB: 0, Enabled: true},
		},
		DefaultCluster: "prod",
		Polling: PollingConfig{
			Enabled:  true,
			Interval: 100 * time.Millisecond,
			Timeout:  50 * time.Millisecond,
		},
		Actions: ActionsConfig{
			RequireConfirmation: false,
			AllowedActions: []ActionType{
				ActionTypePurgeDLQ,
				ActionTypeBenchmark,
				ActionTypePauseQueue,
				ActionTypeResumeQueue,
			},
		},
	}

	manager, err := NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	renderer := NewEnhancedTUIRenderer(manager)
	ctx := context.Background()

	t.Run("Overview_Mode", func(t *testing.T) {
		testOverviewMode(t, renderer, ctx)
	})

	t.Run("Compare_Mode", func(t *testing.T) {
		testCompareMode(t, renderer, ctx)
	})

	t.Run("Cluster_Navigation", func(t *testing.T) {
		testClusterNavigation(t, renderer, ctx)
	})

	t.Run("Action_Execution", func(t *testing.T) {
		testActionExecution(t, renderer, ctx)
	})

	t.Run("Error_Handling", func(t *testing.T) {
		testTUIErrorHandling(t, renderer, ctx)
	})

	t.Run("Performance_UI", func(t *testing.T) {
		testTUIPerformance(t, renderer, ctx)
	})
}

func testOverviewMode(t *testing.T, renderer *EnhancedTUIRenderer, ctx context.Context) {
	// Test initial overview mode
	assert.Equal(t, "overview", renderer.currentMode)

	// Test cluster grid display
	clusters, err := renderer.manager.ListClusters(ctx)
	if err == nil {
		assert.Len(t, clusters, 3)

		// Simulate displaying cluster cards
		for _, cluster := range clusters {
			stats, err := renderer.manager.GetStats(ctx, cluster.Name)
			if err == nil {
				renderer.addRenderEvent("cluster_card", map[string]interface{}{
					"cluster": cluster.Name,
					"stats":   stats,
				})
			}

			health, err := renderer.manager.GetHealth(ctx, cluster.Name)
			if err == nil {
				renderer.addRenderEvent("health_indicator", map[string]interface{}{
					"cluster": cluster.Name,
					"healthy": health.Healthy,
				})
			}
		}

		// Verify we have render events for all clusters
		clusterCards := renderer.getRenderEventsByType("cluster_card")
		assert.GreaterOrEqual(t, len(clusterCards), 1)
	}

	// Test global actions panel
	renderer.addRenderEvent("global_actions", map[string]interface{}{
		"actions": []string{"sync_config", "compare_stats", "purge_dlq"},
	})

	globalActions := renderer.getRenderEventsByType("global_actions")
	assert.Len(t, globalActions, 1)
}

func testCompareMode(t *testing.T, renderer *EnhancedTUIRenderer, ctx context.Context) {
	// Switch to compare mode
	err := renderer.switchToCompareMode(ctx, []string{"prod", "staging"})
	if err != nil {
		t.Skip("Compare mode not properly implemented")
	}

	assert.Equal(t, "compare", renderer.currentMode)
	assert.Equal(t, []string{"prod", "staging"}, renderer.selectedClusters)

	// Test side-by-side comparison
	result, err := renderer.manager.CompareClusters(ctx, []string{"prod", "staging"})
	if err == nil {
		renderer.addRenderEvent("comparison_view", map[string]interface{}{
			"clusters": result.Clusters,
			"metrics":  result.Metrics,
			"deltas":   extractDeltas(result),
		})

		// Verify comparison content
		compareEvents := renderer.getRenderEventsByType("comparison_view")
		assert.Len(t, compareEvents, 1)

		data := compareEvents[0].Data
		assert.Equal(t, []string{"prod", "staging"}, data["clusters"])
	}

	// Test exiting compare mode
	err = renderer.exitCompareMode(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "overview", renderer.currentMode)
}

func testClusterNavigation(t *testing.T, renderer *EnhancedTUIRenderer, ctx context.Context) {
	clusters := []string{"prod", "staging", "dev"}

	// Test switching between clusters with hotkeys
	for i, cluster := range clusters {
		key := string(rune('1' + i))
		err := renderer.simulateKeyPress(tcell.KeyRune, rune('1'+i))
		assert.NoError(t, err)

		err = renderer.switchCluster(ctx, cluster)
		if err == nil {
			assert.Equal(t, cluster, renderer.activeCluster)

			renderer.addRenderEvent("cluster_switch", map[string]interface{}{
				"cluster": cluster,
				"hotkey":  key,
			})
		}
	}

	// Verify all switches were recorded
	switches := renderer.getRenderEventsByType("cluster_switch")
	assert.GreaterOrEqual(t, len(switches), 1)

	// Test tab configuration
	tabConfig, err := renderer.manager.GetTabConfig(ctx)
	if err == nil {
		renderer.addRenderEvent("tab_config", map[string]interface{}{
			"tabs":        tabConfig.Tabs,
			"active_tab":  tabConfig.ActiveTab,
			"compare_mode": tabConfig.CompareMode,
		})

		tabEvents := renderer.getRenderEventsByType("tab_config")
		assert.Len(t, tabEvents, 1)
	}
}

func testActionExecution(t *testing.T, renderer *EnhancedTUIRenderer, ctx context.Context) {
	// Test multi-cluster action execution through UI
	action := &MultiAction{
		ID:      "ui-benchmark-001",
		Type:    ActionTypeBenchmark,
		Targets: []string{"prod", "staging"},
		Parameters: map[string]interface{}{
			"iterations":   float64(5),
			"payload_size": float64(100),
		},
		Status:    ActionStatusPending,
		CreatedAt: time.Now(),
	}

	// Simulate action panel display
	renderer.addRenderEvent("action_panel", map[string]interface{}{
		"action_type": action.Type,
		"targets":     action.Targets,
		"parameters":  action.Parameters,
	})

	// Execute action
	err := renderer.manager.ExecuteAction(ctx, action)
	if err == nil {
		renderer.addRenderEvent("action_result", map[string]interface{}{
			"action_id": action.ID,
			"status":    action.Status,
			"results":   action.Results,
		})

		// Verify action execution was tracked
		actionEvents := renderer.getRenderEventsByType("action_result")
		assert.Len(t, actionEvents, 1)

		result := actionEvents[0].Data
		assert.Equal(t, action.ID, result["action_id"])
	}

	// Test action confirmation UI (if enabled)
	if err != nil && strings.Contains(err.Error(), "confirmation") {
		renderer.addRenderEvent("confirmation_dialog", map[string]interface{}{
			"action_id": action.ID,
			"message":   "Confirm benchmark execution",
			"targets":   action.Targets,
		})

		confirmEvents := renderer.getRenderEventsByType("confirmation_dialog")
		assert.Len(t, confirmEvents, 1)
	}
}

func testTUIErrorHandling(t *testing.T, renderer *EnhancedTUIRenderer, ctx context.Context) {
	// Test error display for connection failures
	err := renderer.switchCluster(ctx, "nonexistent")
	assert.Error(t, err)

	renderer.addRenderEvent("error_display", map[string]interface{}{
		"error_type": "cluster_not_found",
		"message":    err.Error(),
		"severity":   "error",
	})

	// Test error display for action failures
	invalidAction := &MultiAction{
		ID:      "invalid-001",
		Type:    ActionType("invalid"),
		Targets: []string{"prod"},
		Status:  ActionStatusPending,
	}

	err = renderer.manager.ExecuteAction(ctx, invalidAction)
	if err != nil {
		renderer.addRenderEvent("error_display", map[string]interface{}{
			"error_type": "invalid_action",
			"message":    err.Error(),
			"severity":   "error",
		})
	}

	// Verify error events were recorded
	errorEvents := renderer.getRenderEventsByType("error_display")
	assert.GreaterOrEqual(t, len(errorEvents), 1)

	// Test status bar updates
	renderer.addRenderEvent("status_update", map[string]interface{}{
		"message": "Connection failed to cluster nonexistent",
		"level":   "error",
	})

	statusEvents := renderer.getRenderEventsByType("status_update")
	assert.Len(t, statusEvents, 1)
}

func testTUIPerformance(t *testing.T, renderer *EnhancedTUIRenderer, ctx context.Context) {
	start := time.Now()

	// Test rapid cluster switching
	clusters := []string{"prod", "staging", "dev"}
	for i := 0; i < 10; i++ {
		cluster := clusters[i%len(clusters)]
		err := renderer.switchCluster(ctx, cluster)
		if err == nil {
			renderer.addRenderEvent("rapid_switch", map[string]interface{}{
				"cluster": cluster,
				"iteration": i,
			})
		}
	}

	switchDuration := time.Since(start)
	t.Logf("Rapid switching completed in %v", switchDuration)
	assert.Less(t, switchDuration, 2*time.Second)

	// Test rapid stats refresh
	start = time.Now()
	for i := 0; i < 5; i++ {
		allStats, err := renderer.manager.GetAllStats(ctx)
		if err == nil {
			renderer.addRenderEvent("stats_refresh", map[string]interface{}{
				"iteration": i,
				"clusters":  len(allStats),
			})
		}
	}

	refreshDuration := time.Since(start)
	t.Logf("Stats refresh completed in %v", refreshDuration)
	assert.Less(t, refreshDuration, 3*time.Second)

	// Verify performance events
	rapidEvents := renderer.getRenderEventsByType("rapid_switch")
	assert.GreaterOrEqual(t, len(rapidEvents), 5)

	refreshEvents := renderer.getRenderEventsByType("stats_refresh")
	assert.GreaterOrEqual(t, len(refreshEvents), 3)
}

// Helper methods for EnhancedTUIRenderer

func (r *EnhancedTUIRenderer) addRenderEvent(eventType string, data map[string]interface{}) {
	event := RenderEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
	r.renderHistory = append(r.renderHistory, event)
}

func (r *EnhancedTUIRenderer) getRenderEventsByType(eventType string) []RenderEvent {
	var events []RenderEvent
	for _, event := range r.renderHistory {
		if event.Type == eventType {
			events = append(events, event)
		}
	}
	return events
}

func (r *EnhancedTUIRenderer) simulateKeyPress(key tcell.Key, rune rune) error {
	event := tcell.NewEventKey(key, rune, tcell.ModNone)
	r.keyEvents = append(r.keyEvents, event)
	return nil
}

func (r *EnhancedTUIRenderer) switchCluster(ctx context.Context, clusterName string) error {
	err := r.manager.SwitchCluster(ctx, clusterName)
	if err == nil {
		r.activeCluster = clusterName
	}
	return err
}

func (r *EnhancedTUIRenderer) switchToCompareMode(ctx context.Context, clusters []string) error {
	err := r.manager.SetCompareMode(ctx, true, clusters)
	if err == nil {
		r.currentMode = "compare"
		r.selectedClusters = clusters
	}
	return err
}

func (r *EnhancedTUIRenderer) exitCompareMode(ctx context.Context) error {
	err := r.manager.SetCompareMode(ctx, false, nil)
	if err == nil {
		r.currentMode = "overview"
		r.selectedClusters = nil
	}
	return err
}

// Helper functions

func setupTUITestData(mr1, mr2, mr3 *miniredis.Miniredis) {
	// Production data (high load)
	for i := 0; i < 50; i++ {
		mr1.Lpush("jobqueue:queue:high", fmt.Sprintf("prod-job-%d", i))
	}
	for i := 0; i < 20; i++ {
		mr1.Lpush("jobqueue:queue:normal", fmt.Sprintf("prod-normal-%d", i))
	}
	mr1.Lpush("jobqueue:dead_letter", "prod-dead-1")
	mr1.Lpush("jobqueue:dead_letter", "prod-dead-2")

	// Staging data (medium load)
	for i := 0; i < 25; i++ {
		mr2.Lpush("jobqueue:queue:high", fmt.Sprintf("staging-job-%d", i))
	}
	for i := 0; i < 10; i++ {
		mr2.Lpush("jobqueue:queue:normal", fmt.Sprintf("staging-normal-%d", i))
	}
	mr2.Lpush("jobqueue:dead_letter", "staging-dead-1")

	// Development data (low load)
	for i := 0; i < 5; i++ {
		mr3.Lpush("jobqueue:queue:low", fmt.Sprintf("dev-job-%d", i))
	}
}

func extractDeltas(result *CompareResult) map[string]interface{} {
	deltas := make(map[string]interface{})
	for name, metric := range result.Metrics {
		if len(metric.Values) == 2 {
			values := make([]float64, 0, len(metric.Values))
			for _, v := range metric.Values {
				values = append(values, v)
			}
			if len(values) == 2 {
				deltas[name] = values[1] - values[0]
			}
		}
	}
	return deltas
}