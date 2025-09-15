// Copyright 2025 James Ross
package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	multicluster "github.com/flyingrobots/go-redis-work-queue/internal/multi-cluster-control"
	"github.com/alicebob/miniredis/v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockTUIRenderer simulates TUI interactions for testing
type MockTUIRenderer struct {
	app           *tview.Application
	pages         *tview.Pages
	clusterTabs   *tview.Pages
	compareView   *tview.Flex
	actionPanel   *tview.Form
	statusBar     *tview.TextView
	manager       multicluster.Manager
	currentView   string
	keyEvents     []tcell.Event
	renderHistory []string
}

// NewMockTUIRenderer creates a new mock TUI renderer for testing
func NewMockTUIRenderer(manager multicluster.Manager) *MockTUIRenderer {
	app := tview.NewApplication()
	pages := tview.NewPages()

	renderer := &MockTUIRenderer{
		app:           app,
		pages:         pages,
		manager:       manager,
		keyEvents:     make([]tcell.Event, 0),
		renderHistory: make([]string, 0),
	}

	renderer.setupTUI()
	return renderer
}

func (r *MockTUIRenderer) setupTUI() {
	// Setup cluster tabs
	r.clusterTabs = tview.NewPages()
	r.clusterTabs.SetBorder(true).SetTitle("Clusters")

	// Setup compare view
	r.compareView = tview.NewFlex().SetDirection(tview.FlexColumn)
	r.compareView.SetBorder(true).SetTitle("Compare Mode")

	// Setup action panel
	r.actionPanel = tview.NewForm()
	r.actionPanel.SetBorder(true).SetTitle("Actions")

	// Setup status bar
	r.statusBar = tview.NewTextView()
	r.statusBar.SetText("Ready")

	// Main layout
	mainFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(r.clusterTabs, 0, 3, true).
		AddItem(r.actionPanel, 0, 1, false).
		AddItem(r.statusBar, 1, 0, false)

	r.pages.AddPage("main", mainFlex, true, true)
	r.pages.AddPage("compare", r.compareView, true, false)

	// Setup key bindings
	r.setupKeyBindings()

	r.app.SetRoot(r.pages, true)
	r.currentView = "main"
}

func (r *MockTUIRenderer) setupKeyBindings() {
	r.app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		r.keyEvents = append(r.keyEvents, event)

		switch event.Key() {
		case tcell.KeyRune:
			switch event.Rune() {
			case '1', '2', '3', '4', '5', '6', '7', '8', '9':
				r.handleClusterSwitch(string(event.Rune()))
				return nil
			case 'c', 'C':
				r.handleCompareMode()
				return nil
			case 'q', 'Q':
				r.app.Stop()
				return nil
			}
		case tcell.KeyEscape:
			if r.currentView == "compare" {
				r.exitCompareMode()
				return nil
			}
		case tcell.KeyTab:
			r.handleTabNavigation()
			return nil
		}

		return event
	})
}

func (r *MockTUIRenderer) handleClusterSwitch(clusterNum string) {
	ctx := context.Background()

	// Get available clusters
	clusters, err := r.manager.ListClusters(ctx)
	if err != nil {
		r.setStatus("Error: " + err.Error())
		return
	}

	// Convert to 0-based index
	index := int(clusterNum[0] - '1')
	if index >= 0 && index < len(clusters) {
		clusterName := clusters[index].Name
		err := r.manager.SwitchCluster(ctx, clusterName)
		if err != nil {
			r.setStatus("Failed to switch to " + clusterName + ": " + err.Error())
		} else {
			r.setStatus("Switched to cluster: " + clusterName)
			r.refreshClusterView(clusterName)
		}
	}
}

func (r *MockTUIRenderer) handleCompareMode() {
	ctx := context.Background()

	if r.currentView == "compare" {
		r.exitCompareMode()
		return
	}

	// Get available clusters for comparison
	clusters, err := r.manager.ListClusters(ctx)
	if err != nil {
		r.setStatus("Error getting clusters: " + err.Error())
		return
	}

	if len(clusters) < 2 {
		r.setStatus("Need at least 2 clusters for comparison")
		return
	}

	// Enable compare mode with first two clusters
	clusterNames := []string{clusters[0].Name, clusters[1].Name}
	err = r.manager.SetCompareMode(ctx, true, clusterNames)
	if err != nil {
		r.setStatus("Failed to enable compare mode: " + err.Error())
		return
	}

	r.currentView = "compare"
	r.pages.SwitchToPage("compare")
	r.setStatus("Compare mode enabled")
	r.refreshCompareView(clusterNames)
}

func (r *MockTUIRenderer) exitCompareMode() {
	ctx := context.Background()

	err := r.manager.SetCompareMode(ctx, false, nil)
	if err != nil {
		r.setStatus("Failed to exit compare mode: " + err.Error())
		return
	}

	r.currentView = "main"
	r.pages.SwitchToPage("main")
	r.setStatus("Compare mode disabled")
}

func (r *MockTUIRenderer) handleTabNavigation() {
	// Simulate tab navigation between UI elements
	r.setStatus("Tab navigation")
}

func (r *MockTUIRenderer) refreshClusterView(clusterName string) {
	ctx := context.Background()

	// Get cluster stats
	stats, err := r.manager.GetStats(ctx, clusterName)
	if err != nil {
		r.setStatus("Failed to get stats: " + err.Error())
		return
	}

	// Create cluster view
	view := tview.NewTextView().SetDynamicColors(true)
	content := r.formatClusterStats(stats)
	view.SetText(content)

	r.clusterTabs.RemovePage(clusterName)
	r.clusterTabs.AddPage(clusterName, view, true, true)

	r.renderHistory = append(r.renderHistory, "cluster_view:"+clusterName)
}

func (r *MockTUIRenderer) refreshCompareView(clusters []string) {
	ctx := context.Background()

	// Get comparison results
	compareResult, err := r.manager.CompareClusters(ctx, clusters)
	if err != nil {
		r.setStatus("Failed to compare clusters: " + err.Error())
		return
	}

	// Clear compare view
	r.compareView.Clear()

	// Create left panel
	leftPanel := tview.NewTextView().SetDynamicColors(true)
	leftPanel.SetBorder(true).SetTitle(clusters[0])
	leftContent := r.formatComparePanel(clusters[0], compareResult)
	leftPanel.SetText(leftContent)

	// Create right panel
	rightPanel := tview.NewTextView().SetDynamicColors(true)
	rightPanel.SetBorder(true).SetTitle(clusters[1])
	rightContent := r.formatComparePanel(clusters[1], compareResult)
	rightPanel.SetText(rightContent)

	r.compareView.AddItem(leftPanel, 0, 1, false)
	r.compareView.AddItem(rightPanel, 0, 1, false)

	r.renderHistory = append(r.renderHistory, "compare_view:"+strings.Join(clusters, ","))
}

func (r *MockTUIRenderer) formatClusterStats(stats *multicluster.ClusterStats) string {
	var content strings.Builder

	content.WriteString("[yellow]Cluster: [white]" + stats.ClusterName + "\n\n")
	content.WriteString("[yellow]Queues:\n")
	for queueName, size := range stats.QueueSizes {
		content.WriteString("  " + queueName + ": " + string(rune(size)) + "\n")
	}
	content.WriteString("[yellow]Processing: [white]" + string(rune(stats.ProcessingCount)) + "\n")
	content.WriteString("[yellow]Dead Letter: [white]" + string(rune(stats.DeadLetterCount)) + "\n")
	content.WriteString("[yellow]Workers: [white]" + string(rune(stats.WorkerCount)) + "\n")

	return content.String()
}

func (r *MockTUIRenderer) formatComparePanel(clusterName string, compareResult *multicluster.CompareResult) string {
	var content strings.Builder

	content.WriteString("[yellow]Metrics:\n")
	for metricName, metric := range compareResult.Metrics {
		if value, exists := metric.Values[clusterName]; exists {
			content.WriteString("  " + metricName + ": " + string(rune(int(value))) + "\n")
		}
	}

	content.WriteString("\n[yellow]Anomalies:\n")
	for _, anomaly := range compareResult.Anomalies {
		if anomaly.Cluster == clusterName {
			content.WriteString("  [red]" + anomaly.Type + ": " + anomaly.Description + "\n")
		}
	}

	return content.String()
}

func (r *MockTUIRenderer) setStatus(message string) {
	r.statusBar.SetText(message)
	r.renderHistory = append(r.renderHistory, "status:"+message)
}

func (r *MockTUIRenderer) simulateKeyPress(key tcell.Key, rune rune) {
	event := tcell.NewEventKey(key, rune, tcell.ModNone)
	r.app.QueueEvent(event)
	time.Sleep(10 * time.Millisecond) // Small delay for event processing
}

func (r *MockTUIRenderer) getLastRender() string {
	if len(r.renderHistory) == 0 {
		return ""
	}
	return r.renderHistory[len(r.renderHistory)-1]
}

func (r *MockTUIRenderer) getRenderHistory() []string {
	return r.renderHistory
}

// E2E Test Cases

func TestMultiClusterTUI_BasicNavigation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup test clusters
	mr1 := miniredis.RunT(t)
	defer mr1.Close()
	mr2 := miniredis.RunT(t)
	defer mr2.Close()

	cfg := &multicluster.Config{
		Clusters: []multicluster.ClusterConfig{
			{Name: "test1", Label: "Test 1", Color: "#ff0000", Endpoint: mr1.Addr(), DB: 0, Enabled: true},
			{Name: "test2", Label: "Test 2", Color: "#00ff00", Endpoint: mr2.Addr(), DB: 0, Enabled: true},
		},
		DefaultCluster: "test1",
		Polling:        multicluster.PollingConfig{Enabled: false},
	}

	manager, err := multicluster.NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	renderer := NewMockTUIRenderer(manager)

	// Test initial state
	assert.Equal(t, "main", renderer.currentView)

	// Test cluster switching with hotkey '1'
	renderer.simulateKeyPress(tcell.KeyRune, '1')
	time.Sleep(50 * time.Millisecond)

	assert.Contains(t, renderer.getLastRender(), "status:Switched to cluster: test1")

	// Test switching to second cluster with hotkey '2'
	renderer.simulateKeyPress(tcell.KeyRune, '2')
	time.Sleep(50 * time.Millisecond)

	assert.Contains(t, renderer.getLastRender(), "status:Switched to cluster: test2")
}

func TestMultiClusterTUI_CompareMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup test clusters with different data
	mr1 := miniredis.RunT(t)
	defer mr1.Close()
	mr2 := miniredis.RunT(t)
	defer mr2.Close()

	// Add different data to each cluster
	mr1.Lpush("jobqueue:queue:high", "job1")
	mr1.Lpush("jobqueue:queue:high", "job2")
	mr1.Lpush("jobqueue:queue:high", "job3")
	mr1.Lpush("jobqueue:dead_letter", "dead1")
	mr2.Lpush("jobqueue:queue:high", "job4")
	mr2.Lpush("jobqueue:queue:high", "job5")
	mr2.Lpush("jobqueue:dead_letter", "dead2")
	mr2.Lpush("jobqueue:dead_letter", "dead3")

	cfg := &multicluster.Config{
		Clusters: []multicluster.ClusterConfig{
			{Name: "cluster1", Label: "Cluster 1", Color: "#ff0000", Endpoint: mr1.Addr(), DB: 0, Enabled: true},
			{Name: "cluster2", Label: "Cluster 2", Color: "#00ff00", Endpoint: mr2.Addr(), DB: 0, Enabled: true},
		},
		DefaultCluster: "cluster1",
		Polling:        multicluster.PollingConfig{Enabled: false},
	}

	manager, err := multicluster.NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	renderer := NewMockTUIRenderer(manager)

	// Test entering compare mode with 'c' key
	renderer.simulateKeyPress(tcell.KeyRune, 'c')
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, "compare", renderer.currentView)
	assert.Contains(t, renderer.getLastRender(), "compare_view:cluster1,cluster2")

	// Test exiting compare mode with Escape key
	renderer.simulateKeyPress(tcell.KeyEscape, 0)
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, "main", renderer.currentView)
	assert.Contains(t, renderer.getLastRender(), "status:Compare mode disabled")
}

func TestMultiClusterTUI_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup with one working cluster and one failing cluster
	mr := miniredis.RunT(t)
	defer mr.Close()

	cfg := &multicluster.Config{
		Clusters: []multicluster.ClusterConfig{
			{Name: "working", Label: "Working", Color: "#00ff00", Endpoint: mr.Addr(), DB: 0, Enabled: true},
			{Name: "failing", Label: "Failing", Color: "#ff0000", Endpoint: "localhost:9999", DB: 0, Enabled: true},
		},
		DefaultCluster: "working",
		Polling:        multicluster.PollingConfig{Enabled: false},
	}

	manager, err := multicluster.NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	renderer := NewMockTUIRenderer(manager)

	// Test switching to working cluster (should succeed)
	renderer.simulateKeyPress(tcell.KeyRune, '1')
	time.Sleep(50 * time.Millisecond)

	assert.Contains(t, renderer.getLastRender(), "status:Switched to cluster: working")

	// Test switching to failing cluster (should show error in status)
	renderer.simulateKeyPress(tcell.KeyRune, '2')
	time.Sleep(50 * time.Millisecond)

	lastRender := renderer.getLastRender()
	// Should either succeed in switching or show an error - depends on connection handling
	assert.True(t, strings.Contains(lastRender, "Switched to cluster: failing") ||
		strings.Contains(lastRender, "Failed to switch"))
}

func TestMultiClusterTUI_KeyboardShortcuts(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	mr := miniredis.RunT(t)
	defer mr.Close()

	cfg := &multicluster.Config{
		Clusters: []multicluster.ClusterConfig{
			{Name: "shortcuts", Label: "Shortcuts Test", Color: "#0000ff", Endpoint: mr.Addr(), DB: 0, Enabled: true},
		},
		DefaultCluster: "shortcuts",
		Polling:        multicluster.PollingConfig{Enabled: false},
	}

	manager, err := multicluster.NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	renderer := NewMockTUIRenderer(manager)

	// Test various keyboard shortcuts
	shortcuts := []struct {
		key          tcell.Key
		rune         rune
		expectedFunc string
	}{
		{tcell.KeyRune, '1', "cluster_switch"},
		{tcell.KeyRune, 'c', "compare_mode"},
		{tcell.KeyTab, 0, "tab_navigation"},
		{tcell.KeyEscape, 0, "escape"},
	}

	for _, shortcut := range shortcuts {
		initialEvents := len(renderer.keyEvents)
		renderer.simulateKeyPress(shortcut.key, shortcut.rune)
		time.Sleep(20 * time.Millisecond)

		// Verify key event was captured
		assert.Greater(t, len(renderer.keyEvents), initialEvents,
			"Key event should be captured for %c", shortcut.rune)
	}
}

func TestMultiClusterTUI_StatusUpdates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	mr := miniredis.RunT(t)
	defer mr.Close()

	cfg := &multicluster.Config{
		Clusters: []multicluster.ClusterConfig{
			{Name: "status-test", Label: "Status Test", Color: "#ff00ff", Endpoint: mr.Addr(), DB: 0, Enabled: true},
		},
		DefaultCluster: "status-test",
		Polling:        multicluster.PollingConfig{Enabled: false},
	}

	manager, err := multicluster.NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	renderer := NewMockTUIRenderer(manager)

	// Verify initial status
	initialHistory := renderer.getRenderHistory()
	assert.Empty(t, initialHistory)

	// Trigger status update by switching clusters
	renderer.simulateKeyPress(tcell.KeyRune, '1')
	time.Sleep(50 * time.Millisecond)

	history := renderer.getRenderHistory()
	assert.NotEmpty(t, history)

	// Should have multiple render events
	statusUpdates := 0
	clusterViews := 0
	for _, event := range history {
		if strings.HasPrefix(event, "status:") {
			statusUpdates++
		}
		if strings.HasPrefix(event, "cluster_view:") {
			clusterViews++
		}
	}

	assert.Greater(t, statusUpdates, 0, "Should have status updates")
	assert.Greater(t, clusterViews, 0, "Should have cluster view updates")
}

func TestMultiClusterTUI_DataRefresh(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	mr := miniredis.RunT(t)
	defer mr.Close()

	cfg := &multicluster.Config{
		Clusters: []multicluster.ClusterConfig{
			{Name: "refresh-test", Label: "Refresh Test", Color: "#ffff00", Endpoint: mr.Addr(), DB: 0, Enabled: true},
		},
		DefaultCluster: "refresh-test",
		Polling: multicluster.PollingConfig{
			Enabled:  true,
			Interval: 100 * time.Millisecond, // Fast polling for test
			Timeout:  50 * time.Millisecond,
		},
	}

	manager, err := multicluster.NewManager(cfg, zap.NewNop())
	require.NoError(t, err)
	defer manager.Close()

	renderer := NewMockTUIRenderer(manager)

	// Add initial data
	mr.Lpush("jobqueue:queue:test", "initial")

	// Switch to cluster to trigger initial view
	renderer.simulateKeyPress(tcell.KeyRune, '1')
	time.Sleep(100 * time.Millisecond)

	initialHistory := len(renderer.getRenderHistory())

	// Add more data to simulate changes
	mr.Lpush("jobqueue:queue:test", "new_job1")
	mr.Lpush("jobqueue:queue:test", "new_job2")

	// Wait for polling to pick up changes
	time.Sleep(200 * time.Millisecond)

	// Refresh view manually by switching again
	renderer.simulateKeyPress(tcell.KeyRune, '1')
	time.Sleep(50 * time.Millisecond)

	finalHistory := len(renderer.getRenderHistory())
	assert.Greater(t, finalHistory, initialHistory, "Should have more render events after data changes")
}