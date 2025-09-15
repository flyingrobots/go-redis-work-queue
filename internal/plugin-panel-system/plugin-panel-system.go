// Copyright 2025 James Ross
package pluginpanel

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// Manager is the main plugin management system
type Manager struct {
	config            PluginConfig
	plugins           map[string]*Plugin
	runtimes          map[Runtime]RuntimeEngine
	eventBus          EventBus
	permissionManager PermissionManager
	panelRenderer     PanelRenderer
	sandbox           Sandbox
	registry          PluginRegistry
	hotReloader       *HotReloader
	logger            *zap.Logger
	mu                sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	running           bool
}

// NewManager creates a new plugin manager
func NewManager(config PluginConfig, logger *zap.Logger) *Manager {
	if logger == nil {
		logger = zap.NewNop()
	}

	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		config:   config,
		plugins:  make(map[string]*Plugin),
		runtimes: make(map[Runtime]RuntimeEngine),
		logger:   logger,
		ctx:      ctx,
		cancel:   cancel,
	}

	// Initialize default components
	manager.eventBus = NewDefaultEventBus(config.EventQueueSize, logger)
	manager.permissionManager = NewPermissionManager(logger)
	manager.panelRenderer = NewPanelRenderer(logger)
	manager.registry = NewPluginRegistry(logger)

	if config.SandboxEnabled {
		manager.sandbox = NewDefaultSandbox(logger)
	}

	if config.HotReload {
		manager.hotReloader = NewHotReloader(logger)
	}

	return manager
}

// Start initializes the plugin manager and loads plugins
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return fmt.Errorf("plugin manager already running")
	}

	m.logger.Info("Starting plugin manager",
		zap.String("plugin_dir", m.config.PluginDir),
		zap.Int("max_plugins", m.config.MaxPlugins))

	// Register default runtimes
	if err := m.registerRuntimes(); err != nil {
		return fmt.Errorf("failed to register runtimes: %w", err)
	}

	// Create plugin directory if it doesn't exist
	if err := os.MkdirAll(m.config.PluginDir, 0755); err != nil {
		return fmt.Errorf("failed to create plugin directory: %w", err)
	}

	// Load existing plugins
	if err := m.discoverAndLoadPlugins(); err != nil {
		m.logger.Warn("Failed to load some plugins", zap.Error(err))
	}

	// Start hot reloader if enabled
	if m.hotReloader != nil {
		if err := m.hotReloader.Watch(m.config.PluginDir); err != nil {
			m.logger.Warn("Failed to start hot reloader", zap.Error(err))
		} else {
			m.hotReloader.OnChange(m.handleFileChange)
		}
	}

	// Start background maintenance
	go m.maintenanceLoop()

	m.running = true
	m.logger.Info("Plugin manager started")
	return nil
}

// Stop shuts down the plugin manager and all plugins
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return nil
	}

	m.logger.Info("Stopping plugin manager")

	// Stop all plugins
	for _, plugin := range m.plugins {
		if err := m.stopPlugin(ctx, plugin); err != nil {
			m.logger.Warn("Failed to stop plugin",
				zap.String("plugin_id", plugin.ID),
				zap.Error(err))
		}
	}

	// Stop components
	if m.hotReloader != nil {
		m.hotReloader.Stop()
	}

	if m.eventBus != nil {
		m.eventBus.Close()
	}

	// Cancel background operations
	m.cancel()

	m.running = false
	m.logger.Info("Plugin manager stopped")
	return nil
}

// LoadPlugin loads a plugin from the specified path
func (m *Manager) LoadPlugin(ctx context.Context, path string) (*Plugin, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	manifest, err := m.loadManifest(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	// Validate manifest
	if err := m.registry.ValidateManifest(manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	// Check if plugin already loaded
	if existing, exists := m.plugins[manifest.Name]; exists {
		if existing.Version == manifest.Version {
			return existing, nil
		}
		// Unload old version
		if err := m.unloadPlugin(ctx, existing); err != nil {
			m.logger.Warn("Failed to unload old plugin version",
				zap.String("plugin_id", existing.ID),
				zap.Error(err))
		}
	}

	// Create plugin
	plugin := &Plugin{
		ID:          manifest.Name,
		Name:        manifest.Name,
		Version:     manifest.Version,
		Author:      manifest.Author,
		Description: manifest.Description,
		Runtime:     manifest.Runtime,
		Status:      StatusUnloaded,
		Manifest:    manifest,
		LoadedAt:    time.Now(),
		LastUpdate:  time.Now(),
		Metadata:    manifest.Metadata,
	}

	// Load plugin code
	codePath := filepath.Join(path, manifest.EntryPoint)
	code, err := ioutil.ReadFile(codePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read plugin code: %w", err)
	}

	// Get runtime engine
	runtime, exists := m.runtimes[manifest.Runtime]
	if !exists {
		return nil, fmt.Errorf("unsupported runtime: %s", manifest.Runtime)
	}

	// Validate code
	if err := runtime.ValidateCode(code); err != nil {
		return nil, fmt.Errorf("code validation failed: %w", err)
	}

	plugin.Status = StatusLoading

	// Load plugin in runtime
	instance, err := runtime.LoadPlugin(ctx, manifest, code)
	if err != nil {
		plugin.Status = StatusError
		return nil, fmt.Errorf("failed to load plugin in runtime: %w", err)
	}

	plugin.Instance = instance
	plugin.Status = StatusReady

	// Register plugin
	if err := m.registry.Register(plugin); err != nil {
		return nil, fmt.Errorf("failed to register plugin: %w", err)
	}

	// Request permissions for capabilities
	for _, capability := range manifest.Capabilities {
		if !IsReadOnlyCapability(capability) {
			// For non-read-only capabilities, request permission
			grant, err := m.permissionManager.RequestPermission(ctx, plugin.ID, capability)
			if err != nil {
				m.logger.Warn("Failed to request permission",
					zap.String("plugin_id", plugin.ID),
					zap.String("capability", string(capability)),
					zap.Error(err))
			} else if grant != nil && !grant.Granted {
				m.logger.Info("Permission denied for capability",
					zap.String("plugin_id", plugin.ID),
					zap.String("capability", string(capability)))
			}
		}
	}

	// Start plugin if sandbox is enabled
	if m.sandbox != nil {
		if err := m.sandbox.CreateContainer(plugin.ID, &manifest.Resources); err != nil {
			return nil, fmt.Errorf("failed to create sandbox: %w", err)
		}
		if err := m.sandbox.StartPlugin(ctx, plugin.ID, instance); err != nil {
			return nil, fmt.Errorf("failed to start plugin in sandbox: %w", err)
		}
	}

	m.plugins[plugin.ID] = plugin

	// Send plugin load event
	m.eventBus.Publish(ctx, &Event{
		Type:      EventTypePluginLoad,
		Timestamp: time.Now(),
		Source:    "plugin_manager",
		Data: map[string]interface{}{
			"plugin_id": plugin.ID,
			"version":   plugin.Version,
		},
	})

	m.logger.Info("Plugin loaded successfully",
		zap.String("plugin_id", plugin.ID),
		zap.String("version", plugin.Version),
		zap.String("runtime", string(plugin.Runtime)))

	return plugin, nil
}

// UnloadPlugin unloads a plugin by ID
func (m *Manager) UnloadPlugin(ctx context.Context, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	return m.unloadPlugin(ctx, plugin)
}

func (m *Manager) unloadPlugin(ctx context.Context, plugin *Plugin) error {
	m.logger.Info("Unloading plugin", zap.String("plugin_id", plugin.ID))

	// Stop plugin instance
	if err := m.stopPlugin(ctx, plugin); err != nil {
		m.logger.Warn("Failed to stop plugin",
			zap.String("plugin_id", plugin.ID),
			zap.Error(err))
	}

	// Unsubscribe from events
	m.eventBus.Unsubscribe(plugin.ID, []EventType{})

	// Clean up sandbox
	if m.sandbox != nil {
		m.sandbox.StopPlugin(ctx, plugin.ID)
	}

	// Unregister plugin
	m.registry.Unregister(plugin.ID)

	// Remove from plugins map
	delete(m.plugins, plugin.ID)

	m.logger.Info("Plugin unloaded", zap.String("plugin_id", plugin.ID))
	return nil
}

func (m *Manager) stopPlugin(ctx context.Context, plugin *Plugin) error {
	if plugin.Instance == nil {
		return nil
	}

	plugin.Status = StatusTerminated
	return plugin.Instance.Stop(ctx)
}

// GetPlugin retrieves a plugin by ID
func (m *Manager) GetPlugin(pluginID string) (*Plugin, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginID)
	}

	return plugin, nil
}

// ListPlugins returns all loaded plugins
func (m *Manager) ListPlugins() []*Plugin {
	m.mu.RLock()
	defer m.mu.RUnlock()

	plugins := make([]*Plugin, 0, len(m.plugins))
	for _, plugin := range m.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// ReloadPlugin reloads a plugin
func (m *Manager) ReloadPlugin(ctx context.Context, pluginID string) error {
	plugin, err := m.GetPlugin(pluginID)
	if err != nil {
		return err
	}

	// Get the plugin path (assuming it's stored or can be derived)
	pluginPath := filepath.Join(m.config.PluginDir, plugin.Name)

	// Unload current version
	if err := m.UnloadPlugin(ctx, pluginID); err != nil {
		return fmt.Errorf("failed to unload plugin: %w", err)
	}

	// Load new version
	_, err = m.LoadPlugin(ctx, pluginPath)
	return err
}

// EnablePlugin enables a disabled plugin
func (m *Manager) EnablePlugin(ctx context.Context, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	if plugin.Status == StatusDisabled {
		plugin.Status = StatusReady
		if plugin.Instance != nil {
			return plugin.Instance.Start(ctx)
		}
	}

	return nil
}

// DisablePlugin disables a plugin
func (m *Manager) DisablePlugin(ctx context.Context, pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	plugin, exists := m.plugins[pluginID]
	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginID)
	}

	if plugin.Status != StatusDisabled {
		plugin.Status = StatusDisabled
		if plugin.Instance != nil {
			return plugin.Instance.Stop(ctx)
		}
	}

	return nil
}

// RegisterPlugin registers a plugin (for external plugins)
func (m *Manager) RegisterPlugin(plugin *Plugin) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.plugins[plugin.ID] = plugin
	return m.registry.Register(plugin)
}

// UnregisterPlugin unregisters a plugin
func (m *Manager) UnregisterPlugin(pluginID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.plugins, pluginID)
	return m.registry.Unregister(pluginID)
}

// SendEventToPlugin sends an event to a specific plugin
func (m *Manager) SendEventToPlugin(ctx context.Context, pluginID string, event *Event) error {
	plugin, err := m.GetPlugin(pluginID)
	if err != nil {
		return err
	}

	if plugin.Status != StatusReady && plugin.Status != StatusRunning {
		return fmt.Errorf("plugin not ready: %s", pluginID)
	}

	return plugin.Instance.SendEvent(ctx, event)
}

// GetPluginMetrics returns metrics for a plugin
func (m *Manager) GetPluginMetrics(pluginID string) (map[string]interface{}, error) {
	plugin, err := m.GetPlugin(pluginID)
	if err != nil {
		return nil, err
	}

	metrics := plugin.Instance.GetMetrics()

	// Add resource usage if sandbox is enabled
	if m.sandbox != nil {
		usage, err := m.sandbox.MonitorResources(pluginID)
		if err == nil {
			metrics["resource_usage"] = usage
		}
	}

	return metrics, nil
}

// Helper methods

func (m *Manager) loadManifest(pluginPath string) (*PluginManifest, error) {
	manifestPath := filepath.Join(pluginPath, "manifest.yaml")
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		manifestPath = filepath.Join(pluginPath, "manifest.yml")
	}

	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest PluginManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	// Set defaults
	if manifest.Resources.MaxMemoryMB == 0 {
		manifest.Resources = m.config.DefaultResources
	}

	return &manifest, nil
}

func (m *Manager) discoverAndLoadPlugins() error {
	entries, err := ioutil.ReadDir(m.config.PluginDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		pluginPath := filepath.Join(m.config.PluginDir, entry.Name())
		if _, err := m.LoadPlugin(m.ctx, pluginPath); err != nil {
			m.logger.Warn("Failed to load plugin",
				zap.String("path", pluginPath),
				zap.Error(err))
		}
	}

	return nil
}

func (m *Manager) registerRuntimes() error {
	// Register WASM runtime (temporarily disabled due to build issues)
	// wasmRuntime := NewWASMRuntime(m.logger)
	// m.runtimes[RuntimeWASM] = wasmRuntime

	// Register Starlark runtime
	starlarkRuntime := NewStarlarkRuntime(m.logger)
	m.runtimes[RuntimeStarlark] = starlarkRuntime

	// Register Lua runtime (optional)
	// luaRuntime := NewLuaRuntime(m.logger)
	// m.runtimes[RuntimeLua] = luaRuntime

	return nil
}

func (m *Manager) handleFileChange(path string, event string) {
	m.logger.Info("Plugin file changed",
		zap.String("path", path),
		zap.String("event", event))

	// Determine which plugin this affects
	pluginDir := filepath.Dir(path)
	pluginName := filepath.Base(pluginDir)

	// Reload the plugin
	if err := m.ReloadPlugin(m.ctx, pluginName); err != nil {
		m.logger.Warn("Failed to reload plugin",
			zap.String("plugin", pluginName),
			zap.Error(err))
	}
}

func (m *Manager) maintenanceLoop() {
	ticker := time.NewTicker(m.config.GCInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.performMaintenance()
		case <-m.ctx.Done():
			return
		}
	}
}

func (m *Manager) performMaintenance() {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check plugin health and resource usage
	for _, plugin := range m.plugins {
		if plugin.Status == StatusError {
			continue
		}

		// Check if plugin is responsive
		if plugin.Instance != nil {
			metrics := plugin.Instance.GetMetrics()
			if lastHeartbeat, ok := metrics["last_heartbeat"].(time.Time); ok {
				if time.Since(lastHeartbeat) > m.config.PluginTimeout {
					m.logger.Warn("Plugin unresponsive",
						zap.String("plugin_id", plugin.ID))
					plugin.Status = StatusError
				}
			}
		}

		// Monitor resource usage
		if m.sandbox != nil {
			usage, err := m.sandbox.MonitorResources(plugin.ID)
			if err != nil {
				continue
			}

			// Check for resource violations
			if len(usage.Violations) > 0 {
				m.logger.Warn("Plugin resource violations",
					zap.String("plugin_id", plugin.ID),
					zap.Strings("violations", usage.Violations))
				plugin.ErrorCount++
			}
		}
	}
}

// GetEventBus returns the event bus for external use
func (m *Manager) GetEventBus() EventBus {
	return m.eventBus
}

// GetPermissionManager returns the permission manager
func (m *Manager) GetPermissionManager() PermissionManager {
	return m.permissionManager
}

// GetPanelRenderer returns the panel renderer
func (m *Manager) GetPanelRenderer() PanelRenderer {
	return m.panelRenderer
}

// GetRegistry returns the plugin registry
func (m *Manager) GetRegistry() PluginRegistry {
	return m.registry
}

// CreateHostAPI creates a capability-gated host API for a plugin
func (m *Manager) CreateHostAPI(pluginID string) HostAPI {
	return NewHostAPI(m, pluginID, m.logger)
}