// Copyright 2025 James Ross
package pluginpanel

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestPluginManager_New(t *testing.T) {
	config := DefaultPluginConfig()
	logger := zaptest.NewLogger(t)

	manager := NewManager(config, logger)

	assert.NotNil(t, manager)
	assert.Equal(t, config, manager.config)
	assert.NotNil(t, manager.eventBus)
	assert.NotNil(t, manager.permissionManager)
	assert.NotNil(t, manager.panelRenderer)
}

func TestPluginManager_StartStop(t *testing.T) {
	config := DefaultPluginConfig()
	config.PluginDir = t.TempDir()
	logger := zaptest.NewLogger(t)

	manager := NewManager(config, logger)

	ctx := context.Background()

	// Test start
	err := manager.Start(ctx)
	require.NoError(t, err)
	assert.True(t, manager.running)

	// Test stop
	err = manager.Stop(ctx)
	require.NoError(t, err)
	assert.False(t, manager.running)
}

func TestEventBus_SubscribeUnsubscribe(t *testing.T) {
	logger := zaptest.NewLogger(t)
	bus := NewDefaultEventBus(100, logger)
	defer bus.Close()

	pluginID := "test-plugin"
	eventTypes := []EventType{EventTypeStats, EventTypeSelection}

	// Test subscribe
	err := bus.Subscribe(pluginID, eventTypes)
	assert.NoError(t, err)

	// Check subscriptions
	subs := bus.GetSubscriptions(pluginID)
	assert.Len(t, subs, 2)
	assert.Contains(t, subs, EventTypeStats)
	assert.Contains(t, subs, EventTypeSelection)

	// Test unsubscribe
	err = bus.Unsubscribe(pluginID, []EventType{EventTypeStats})
	assert.NoError(t, err)

	subs = bus.GetSubscriptions(pluginID)
	assert.Len(t, subs, 1)
	assert.Contains(t, subs, EventTypeSelection)

	// Test unsubscribe all
	err = bus.Unsubscribe(pluginID, []EventType{})
	assert.NoError(t, err)

	subs = bus.GetSubscriptions(pluginID)
	assert.Len(t, subs, 0)
}

func TestEventBus_PublishEvent(t *testing.T) {
	logger := zaptest.NewLogger(t)
	bus := NewDefaultEventBus(100, logger)
	defer bus.Close()

	ctx := context.Background()
	event := &Event{
		Type:      EventTypeStats,
		Timestamp: time.Now(),
		Source:    "test",
		Data: map[string]interface{}{
			"test": "data",
		},
	}

	err := bus.Publish(ctx, event)
	assert.NoError(t, err)
}

func TestPermissionManager_RequestGrant(t *testing.T) {
	logger := zaptest.NewLogger(t)
	pm := NewPermissionManager(logger)

	ctx := context.Background()
	pluginID := "test-plugin"

	// Test read-only capability (auto-granted)
	grant, err := pm.RequestPermission(ctx, pluginID, CapabilityReadStats)
	require.NoError(t, err)
	assert.True(t, grant.Granted)
	assert.Equal(t, "system", grant.GrantedBy)

	// Test action capability (pending approval)
	grant, err = pm.RequestPermission(ctx, pluginID, CapabilityEnqueue)
	require.NoError(t, err)
	assert.False(t, grant.Granted)

	// Check pending permissions
	pending := pm.GetPendingPermissions()
	assert.Len(t, pending, 1)
	assert.Equal(t, CapabilityEnqueue, pending[0].Capability)

	// Grant the permission
	grant.Granted = true
	grant.GrantedBy = "user"
	err = pm.GrantPermission(ctx, grant)
	assert.NoError(t, err)

	// Check permission
	hasPermission := pm.CheckPermission(pluginID, CapabilityEnqueue)
	assert.True(t, hasPermission)

	// Test revoke permission
	err = pm.RevokePermission(ctx, pluginID, CapabilityEnqueue)
	assert.NoError(t, err)

	hasPermission = pm.CheckPermission(pluginID, CapabilityEnqueue)
	assert.False(t, hasPermission)
}

func TestPanelRenderer_ZoneOperations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	renderer := NewPanelRenderer(logger)

	pluginID := "test-plugin"

	// Test create zone
	zone, err := renderer.CreateZone(pluginID, 80, 24)
	require.NoError(t, err)
	assert.Equal(t, pluginID, zone.PluginID)
	assert.Equal(t, 80, zone.Width)
	assert.Equal(t, 24, zone.Height)

	// Test get zone
	retrievedZone, err := renderer.GetZone(zone.ID)
	require.NoError(t, err)
	assert.Equal(t, zone.ID, retrievedZone.ID)

	// Test list zones
	zones := renderer.ListZones()
	assert.Len(t, zones, 1)

	// Test focus zone
	err = renderer.FocusZone(zone.ID)
	assert.NoError(t, err)

	retrievedZone, _ = renderer.GetZone(zone.ID)
	assert.True(t, retrievedZone.Focused)

	// Test resize zone
	err = renderer.ResizeZone(zone.ID, 100, 30)
	assert.NoError(t, err)

	retrievedZone, _ = renderer.GetZone(zone.ID)
	assert.Equal(t, 100, retrievedZone.Width)
	assert.Equal(t, 30, retrievedZone.Height)

	// Test render command
	cmd := &RenderCommand{
		Type: RenderTypeText,
		Zone: zone.ID,
		X:    10,
		Y:    5,
		Text: "Hello, Plugin!",
	}

	ctx := context.Background()
	err = renderer.RenderCommand(ctx, cmd)
	assert.NoError(t, err)

	// Test destroy zone
	err = renderer.DestroyZone(zone.ID)
	assert.NoError(t, err)

	zones = renderer.ListZones()
	assert.Len(t, zones, 0)
}

func TestPluginRegistry_Operations(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := NewPluginRegistry(logger)

	plugin := &Plugin{
		ID:          "test-plugin",
		Name:        "test-plugin",
		Version:     "1.0.0",
		Author:      "Test Author",
		Description: "Test plugin description",
		Status:      StatusReady,
	}

	// Test register
	err := registry.Register(plugin)
	assert.NoError(t, err)

	// Test find
	found, err := registry.Find("test-plugin", "1.0.0")
	require.NoError(t, err)
	assert.Equal(t, plugin.ID, found.ID)

	// Test find without version
	found, err = registry.Find("test-plugin", "")
	require.NoError(t, err)
	assert.Equal(t, plugin.ID, found.ID)

	// Test list
	plugins := registry.List()
	assert.Len(t, plugins, 1)
	assert.Equal(t, plugin.ID, plugins[0].ID)

	// Test search
	results := registry.Search("test")
	assert.Len(t, results, 1)
	assert.Equal(t, plugin.ID, results[0].ID)

	// Test search with no matches
	results = registry.Search("nonexistent")
	assert.Len(t, results, 0)

	// Test unregister
	err = registry.Unregister(plugin.ID)
	assert.NoError(t, err)

	plugins = registry.List()
	assert.Len(t, plugins, 0)
}

func TestValidationFunctions(t *testing.T) {
	// Test capability validation
	assert.True(t, IsValidCapability(CapabilityReadStats))
	assert.True(t, IsValidCapability(CapabilityEnqueue))
	assert.False(t, IsValidCapability("invalid_capability"))

	// Test capability categorization
	assert.True(t, IsReadOnlyCapability(CapabilityReadStats))
	assert.False(t, IsReadOnlyCapability(CapabilityEnqueue))

	assert.True(t, IsActionCapability(CapabilityEnqueue))
	assert.False(t, IsActionCapability(CapabilityReadStats))

	assert.True(t, IsSystemCapability(CapabilityFileAccess))
	assert.False(t, IsSystemCapability(CapabilityReadStats))
}

func TestManifestValidation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	registry := NewPluginRegistry(logger)

	// Valid manifest
	validManifest := &PluginManifest{
		Name:         "test-plugin",
		Version:      "1.0.0",
		Author:       "Test",
		Description:  "Test plugin",
		Runtime:      RuntimeStarlark,
		EntryPoint:   "main.star",
		Capabilities: []Capability{CapabilityReadStats},
	}

	err := registry.ValidateManifest(validManifest)
	assert.NoError(t, err)

	// Invalid manifest - missing name
	invalidManifest := &PluginManifest{
		Version:    "1.0.0",
		Runtime:    RuntimeStarlark,
		EntryPoint: "main.star",
	}

	err = registry.ValidateManifest(invalidManifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")

	// Invalid manifest - invalid capability
	invalidCapabilityManifest := &PluginManifest{
		Name:         "test-plugin",
		Version:      "1.0.0",
		Runtime:      RuntimeStarlark,
		EntryPoint:   "main.star",
		Capabilities: []Capability{"invalid_capability"},
	}

	err = registry.ValidateManifest(invalidCapabilityManifest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid capability")
}

func TestDefaultConfigurations(t *testing.T) {
	// Test default resource requirements
	defaultResources := DefaultResourceRequirements()
	assert.Equal(t, 64, defaultResources.MaxMemoryMB)
	assert.Equal(t, 25, defaultResources.MaxCPUPercent)
	assert.Equal(t, 30*time.Second, defaultResources.Timeout)

	// Test default plugin config
	defaultConfig := DefaultPluginConfig()
	assert.Equal(t, "./plugins", defaultConfig.PluginDir)
	assert.Equal(t, 20, defaultConfig.MaxPlugins)
	assert.True(t, defaultConfig.HotReload)
	assert.True(t, defaultConfig.SandboxEnabled)
	assert.Contains(t, defaultConfig.DefaultPermissions, CapabilityReadStats)
}

func TestSandbox_ResourceMonitoring(t *testing.T) {
	logger := zaptest.NewLogger(t)
	sandbox := NewDefaultSandbox(logger)

	pluginID := "test-plugin"
	resources := DefaultResourceRequirements()

	// Test create container
	err := sandbox.CreateContainer(pluginID, &resources)
	assert.NoError(t, err)

	// Test resource monitoring
	usage, err := sandbox.MonitorResources(pluginID)
	require.NoError(t, err)
	assert.Equal(t, pluginID, usage.PluginID)

	// Test timeout enforcement
	err = sandbox.EnforceTimeout(pluginID, 1*time.Nanosecond)
	assert.Error(t, err) // Should fail immediately due to very short timeout

	// Test kill plugin
	err = sandbox.KillPlugin(pluginID)
	assert.NoError(t, err)
}

func TestHostAPI_CapabilityGating(t *testing.T) {
	config := DefaultPluginConfig()
	logger := zaptest.NewLogger(t)
	manager := NewManager(config, logger)

	pluginID := "test-plugin"
	hostAPI := NewHostAPI(manager, pluginID, logger)

	ctx := context.Background()

	// Test without permission (should fail)
	_, err := hostAPI.GetStats(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")

	// Grant permission
	pm := manager.GetPermissionManager()
	grant := &PermissionGrant{
		PluginID:   pluginID,
		Capability: CapabilityReadStats,
		Granted:    true,
		GrantedBy:  "test",
		GrantedAt:  time.Now(),
	}
	err = pm.GrantPermission(ctx, grant)
	require.NoError(t, err)

	// Test with permission (should succeed)
	stats, err := hostAPI.GetStats(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, stats)

	// Test action capability without permission
	err = hostAPI.EnqueueJob(ctx, "test-queue", map[string]interface{}{"test": "data"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestPluginError(t *testing.T) {
	err := &PluginError{
		PluginID:  "test-plugin",
		Operation: "load",
		Message:   "test error",
		Code:      ErrorCodeLoadFailed,
		Timestamp: time.Now(),
	}

	assert.Contains(t, err.Error(), "test-plugin")
	assert.Contains(t, err.Error(), "load")
	assert.Contains(t, err.Error(), "test error")
}

// Benchmark tests

func BenchmarkEventBus_Publish(b *testing.B) {
	logger := zaptest.NewLogger(b)
	bus := NewDefaultEventBus(10000, logger)
	defer bus.Close()

	ctx := context.Background()
	event := &Event{
		Type:      EventTypeStats,
		Timestamp: time.Now(),
		Source:    "bench",
		Data:      map[string]interface{}{"test": "data"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bus.Publish(ctx, event)
	}
}

func BenchmarkPermissionManager_CheckPermission(b *testing.B) {
	logger := zaptest.NewLogger(b)
	pm := NewPermissionManager(logger)

	ctx := context.Background()
	pluginID := "bench-plugin"

	// Grant a permission
	grant := &PermissionGrant{
		PluginID:   pluginID,
		Capability: CapabilityReadStats,
		Granted:    true,
		GrantedBy:  "bench",
		GrantedAt:  time.Now(),
	}
	pm.GrantPermission(ctx, grant)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.CheckPermission(pluginID, CapabilityReadStats)
	}
}