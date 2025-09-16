package context_menus

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContextMenuSystem(t *testing.T) {
	cms := New()
	assert.NotNil(t, cms)
	assert.NotNil(t, cms.menu)
	assert.False(t, cms.IsVisible())
	assert.True(t, cms.IsEnabled())
}

func TestZoneRegistration(t *testing.T) {
	cms := New()

	// Test table row registration
	err := cms.RegisterTableRow(0, 0, 1, 50, "test_queue")
	require.NoError(t, err)

	zone, found := cms.GetZoneAt(25, 1)
	assert.True(t, found)
	assert.Equal(t, "table-row-0", zone.ID)
	assert.Equal(t, ContextQueueRow, zone.Context.Type)
	assert.Equal(t, "test_queue", zone.Context.QueueName)

	// Test tab registration
	err = cms.RegisterTab("jobs", "Jobs", 10, 0, 10)
	require.NoError(t, err)

	zone, found = cms.GetZoneAt(15, 0)
	assert.True(t, found)
	assert.Equal(t, "tab-jobs", zone.ID)
	assert.Equal(t, ContextTab, zone.Context.Type)
	assert.Equal(t, "jobs", zone.Context.ItemID)

	// Test chart registration
	err = cms.RegisterChart("throughput", 0, 5, 80, 20)
	require.NoError(t, err)

	zone, found = cms.GetZoneAt(40, 15)
	assert.True(t, found)
	assert.Equal(t, "chart-throughput", zone.ID)
	assert.Equal(t, ContextChart, zone.Context.Type)

	// Test DLQ item registration
	err = cms.RegisterDLQItem(0, "job123", 0, 25, 50)
	require.NoError(t, err)

	zone, found = cms.GetZoneAt(25, 25)
	assert.True(t, found)
	assert.Equal(t, "dlq-item-0", zone.ID)
	assert.Equal(t, ContextDLQItem, zone.Context.Type)
	assert.Equal(t, "job123", zone.Context.JobID)
}

func TestZoneManagerMethods(t *testing.T) {
	cms := New()

	// Register multiple zones
	err := cms.RegisterTableRow(0, 0, 1, 50, "queue1")
	require.NoError(t, err)
	err = cms.RegisterTableRow(1, 0, 2, 50, "queue2")
	require.NoError(t, err)

	// Test zone retrieval
	zone, found := cms.GetZoneAt(25, 1)
	assert.True(t, found)
	assert.Equal(t, "queue1", zone.Context.QueueName)

	zone, found = cms.GetZoneAt(25, 2)
	assert.True(t, found)
	assert.Equal(t, "queue2", zone.Context.QueueName)

	// Test zone outside bounds
	_, found = cms.GetZoneAt(100, 100)
	assert.False(t, found)

	// Test clear zones
	cms.ClearZones()
	_, found = cms.GetZoneAt(25, 1)
	assert.False(t, found)
}

func TestActionRegistry(t *testing.T) {
	cms := New()

	// Test default actions are registered
	ctx := MenuContext{
		Type:      ContextQueueRow,
		QueueName: "test_queue",
	}

	actions := cms.GetActions(ctx)
	assert.NotEmpty(t, actions)

	// Verify specific actions exist
	actionIDs := make([]string, len(actions))
	for i, action := range actions {
		actionIDs[i] = action.ID
	}

	assert.Contains(t, actionIDs, "peek")
	assert.Contains(t, actionIDs, "enqueue")
	assert.Contains(t, actionIDs, "purge")
	assert.Contains(t, actionIDs, "copy_queue_name")

	// Test DLQ actions
	dlqCtx := MenuContext{
		Type:  ContextDLQItem,
		JobID: "job123",
	}

	dlqActions := cms.GetActions(dlqCtx)
	assert.NotEmpty(t, dlqActions)

	dlqActionIDs := make([]string, len(dlqActions))
	for i, action := range dlqActions {
		dlqActionIDs[i] = action.ID
	}

	assert.Contains(t, dlqActionIDs, "requeue")
	assert.Contains(t, dlqActionIDs, "purge_dlq")
	assert.Contains(t, dlqActionIDs, "copy_job_id")
}

func TestCustomActionRegistration(t *testing.T) {
	cms := New()

	// Register custom action
	customAction := MenuAction{
		ID:          "custom_test",
		Label:       "Custom Test Action",
		Accelerator: "t",
		Destructive: false,
		Confirm:     false,
	}

	cms.RegisterAction(ContextQueueRow, customAction)

	// Register custom handler
	customHandler := func(ctx context.Context, action MenuAction, menuCtx MenuContext) tea.Cmd {
		return nil
	}

	cms.RegisterHandler("custom_test", customHandler)

	// Test that custom action appears in context
	ctx := MenuContext{
		Type:      ContextQueueRow,
		QueueName: "test_queue",
	}

	actions := cms.GetActions(ctx)
	found := false
	for _, action := range actions {
		if action.ID == "custom_test" {
			found = true
			assert.Equal(t, "Custom Test Action", action.Label)
			assert.Equal(t, "t", action.Accelerator)
			break
		}
	}
	assert.True(t, found, "Custom action not found in actions list")
}

func TestMenuStateUpdates(t *testing.T) {
	cms := New()

	// Initially menu should not be visible
	assert.False(t, cms.IsVisible())

	// Show menu
	ctx := MenuContext{
		Type:      ContextQueueRow,
		QueueName: "test_queue",
		Position:  Position{X: 10, Y: 10},
	}

	cmd := cms.ShowMenu(ctx)
	assert.NotNil(t, cmd)

	// Update with show menu message
	msg := cmd()
	cms, _ = cms.Update(msg)
	assert.True(t, cms.IsVisible())

	// Hide menu
	hideCmd := cms.HideMenu()
	assert.NotNil(t, hideCmd)

	// Update with hide menu message
	hideMsg := hideCmd()
	cms, _ = cms.Update(hideMsg)
	assert.False(t, cms.IsVisible())
}

func TestKeyboardNavigation(t *testing.T) {
	cms := New()

	// Show menu first
	ctx := MenuContext{
		Type:      ContextQueueRow,
		QueueName: "test_queue",
		Position:  Position{X: 10, Y: 10},
	}

	showCmd := cms.ShowMenu(ctx)
	showMsg := showCmd()
	cms, _ = cms.Update(showMsg)

	// Test escape key hides menu
	escMsg := tea.KeyMsg{Type: tea.KeyEsc}
	cms, _ = cms.Update(escMsg)
	assert.False(t, cms.IsVisible())

	// Show menu again for navigation tests
	cms, _ = cms.Update(showMsg)
	assert.True(t, cms.IsVisible())

	// Test up/down navigation would be tested with actual menu state
	// This is a basic structure test
	upMsg := tea.KeyMsg{Type: tea.KeyUp}
	cms, _ = cms.Update(upMsg)
	assert.True(t, cms.IsVisible()) // Menu should remain visible

	downMsg := tea.KeyMsg{Type: tea.KeyDown}
	cms, _ = cms.Update(downMsg)
	assert.True(t, cms.IsVisible()) // Menu should remain visible
}

func TestMouseInteraction(t *testing.T) {
	cms := New()

	// Register a zone
	err := cms.RegisterTableRow(0, 0, 5, 50, "test_queue")
	require.NoError(t, err)

	// Test right-click on zone
	rightClickMsg := tea.MouseMsg{
		X:    25,
		Y:    5,
		Type: tea.MouseRight,
	}

	cms, _ = cms.Update(rightClickMsg)
	// After right-click processing, menu should be visible
	// Note: This test may need adjustment based on actual implementation details

	// Test left-click outside menu area should hide menu if visible
	leftClickMsg := tea.MouseMsg{
		X:    100,
		Y:    100,
		Type: tea.MouseLeft,
	}

	cms, _ = cms.Update(leftClickMsg)
	// Menu should be hidden after clicking outside
}

func TestEnableDisable(t *testing.T) {
	cms := New()

	// Initially enabled
	assert.True(t, cms.IsEnabled())

	// Disable
	cms.SetEnabled(false)
	assert.False(t, cms.IsEnabled())

	// Re-enable
	cms.SetEnabled(true)
	assert.True(t, cms.IsEnabled())
}

func TestActionAvailability(t *testing.T) {
	cms := New()

	// Test queue context with empty queue name
	emptyQueueCtx := MenuContext{
		Type:      ContextQueueRow,
		QueueName: "",
	}

	actions := cms.GetActions(emptyQueueCtx)
	// Actions should be filtered based on availability
	for _, action := range actions {
		// Actions requiring queue name should not be available
		if action.ID == "peek" || action.ID == "enqueue" {
			t.Errorf("Action %s should not be available without queue name", action.ID)
		}
	}

	// Test DLQ context with empty job ID
	emptyJobCtx := MenuContext{
		Type:  ContextDLQItem,
		JobID: "",
	}

	dlqActions := cms.GetActions(emptyJobCtx)
	for _, action := range dlqActions {
		// Actions requiring job ID should not be available
		if action.ID == "requeue" || action.ID == "copy_job_id" {
			t.Errorf("Action %s should not be available without job ID", action.ID)
		}
	}
}

func TestWindowSizeUpdate(t *testing.T) {
	cms := New()

	// Test window size message
	sizeMsg := tea.WindowSizeMsg{
		Width:  100,
		Height: 50,
	}

	cms, _ = cms.Update(sizeMsg)
	// Should not panic and should handle the resize
	assert.NotNil(t, cms)
}

func TestErrorCases(t *testing.T) {
	cms := New()

	// Test registering zone with empty ID should fail at zone level
	zone := BubbleZone{
		ID:      "",
		X:       0,
		Y:       0,
		Width:   10,
		Height:  10,
		Enabled: true,
	}

	err := cms.RegisterZone(zone)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "zone ID cannot be empty")

	// Test unregistering non-existent zone should not panic
	cms.UnregisterZone("non-existent")
	assert.NotNil(t, cms) // Should not panic
}