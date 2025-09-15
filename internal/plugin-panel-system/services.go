// Copyright 2025 James Ross
package pluginpanel

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
)

// DefaultEventBus implements the EventBus interface
type DefaultEventBus struct {
	subscribers map[string]map[EventType]bool // pluginID -> eventTypes
	eventQueue  chan *Event
	logger      *zap.Logger
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	running     bool
}

// NewDefaultEventBus creates a new event bus
func NewDefaultEventBus(queueSize int, logger *zap.Logger) *DefaultEventBus {
	ctx, cancel := context.WithCancel(context.Background())

	bus := &DefaultEventBus{
		subscribers: make(map[string]map[EventType]bool),
		eventQueue:  make(chan *Event, queueSize),
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
	}

	go bus.eventProcessor()
	return bus
}

// Subscribe subscribes a plugin to event types
func (eb *DefaultEventBus) Subscribe(pluginID string, eventTypes []EventType) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.subscribers[pluginID] == nil {
		eb.subscribers[pluginID] = make(map[EventType]bool)
	}

	for _, eventType := range eventTypes {
		eb.subscribers[pluginID][eventType] = true
	}

	eb.logger.Debug("Plugin subscribed to events",
		zap.String("plugin_id", pluginID),
		zap.Int("event_types", len(eventTypes)))

	return nil
}

// Unsubscribe unsubscribes a plugin from event types
func (eb *DefaultEventBus) Unsubscribe(pluginID string, eventTypes []EventType) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if len(eventTypes) == 0 {
		// Unsubscribe from all events
		delete(eb.subscribers, pluginID)
	} else {
		if subs, exists := eb.subscribers[pluginID]; exists {
			for _, eventType := range eventTypes {
				delete(subs, eventType)
			}
			if len(subs) == 0 {
				delete(eb.subscribers, pluginID)
			}
		}
	}

	return nil
}

// Publish publishes an event to all subscribers
func (eb *DefaultEventBus) Publish(ctx context.Context, event *Event) error {
	select {
	case eb.eventQueue <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("event queue full")
	}
}

// SendToPlugin sends an event to a specific plugin
func (eb *DefaultEventBus) SendToPlugin(ctx context.Context, pluginID string, event *Event) error {
	event.PluginID = pluginID
	return eb.Publish(ctx, event)
}

// GetSubscriptions returns the event types a plugin is subscribed to
func (eb *DefaultEventBus) GetSubscriptions(pluginID string) []EventType {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if subs, exists := eb.subscribers[pluginID]; exists {
		eventTypes := make([]EventType, 0, len(subs))
		for eventType := range subs {
			eventTypes = append(eventTypes, eventType)
		}
		return eventTypes
	}

	return []EventType{}
}

// Close shuts down the event bus
func (eb *DefaultEventBus) Close() error {
	eb.cancel()
	close(eb.eventQueue)
	return nil
}

// eventProcessor processes events from the queue
func (eb *DefaultEventBus) eventProcessor() {
	for {
		select {
		case event, ok := <-eb.eventQueue:
			if !ok {
				// Channel is closed, exit
				return
			}
			if event != nil {
				eb.distributeEvent(event)
			}
		case <-eb.ctx.Done():
			return
		}
	}
}

// distributeEvent sends an event to all interested subscribers
func (eb *DefaultEventBus) distributeEvent(event *Event) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// If event is targeted to a specific plugin
	if event.PluginID != "" {
		if subs, exists := eb.subscribers[event.PluginID]; exists {
			if subs[event.Type] {
				eb.logger.Debug("Sending targeted event",
					zap.String("plugin_id", event.PluginID),
					zap.String("event_type", string(event.Type)))
				// In a real implementation, you'd send this to the plugin manager
			}
		}
		return
	}

	// Broadcast to all interested subscribers
	for pluginID, subs := range eb.subscribers {
		if subs[event.Type] {
			eb.logger.Debug("Broadcasting event",
				zap.String("plugin_id", pluginID),
				zap.String("event_type", string(event.Type)))
			// In a real implementation, you'd send this to the plugin manager
		}
	}
}

// DefaultPermissionManager implements the PermissionManager interface
type DefaultPermissionManager struct {
	grants       map[string]map[Capability]*PermissionGrant // pluginID -> capability -> grant
	pending      []*PermissionGrant
	persistPath  string
	logger       *zap.Logger
	mu           sync.RWMutex
}

// NewPermissionManager creates a new permission manager
func NewPermissionManager(logger *zap.Logger) *DefaultPermissionManager {
	pm := &DefaultPermissionManager{
		grants:      make(map[string]map[Capability]*PermissionGrant),
		pending:     make([]*PermissionGrant, 0),
		persistPath: "./plugins/permissions.json",
		logger:      logger,
	}

	// Load existing permissions
	pm.loadPermissions()
	return pm
}

// RequestPermission requests permission for a capability
func (pm *DefaultPermissionManager) RequestPermission(ctx context.Context, pluginID string, capability Capability) (*PermissionGrant, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Check if permission already exists
	if grants, exists := pm.grants[pluginID]; exists {
		if grant, exists := grants[capability]; exists {
			if !grant.Revoked {
				return grant, nil
			}
		}
	}

	// Create new permission request
	grant := &PermissionGrant{
		PluginID:   pluginID,
		Capability: capability,
		Granted:    false, // Will be set when user approves
		GrantedAt:  time.Now(),
	}

	// For read-only capabilities, auto-grant
	if IsReadOnlyCapability(capability) {
		grant.Granted = true
		grant.GrantedBy = "system"
		pm.storeGrant(grant)
	} else {
		// Add to pending list for user approval
		pm.pending = append(pm.pending, grant)
		pm.logger.Info("Permission request pending user approval",
			zap.String("plugin_id", pluginID),
			zap.String("capability", string(capability)))
	}

	return grant, nil
}

// CheckPermission checks if a plugin has permission for a capability
func (pm *DefaultPermissionManager) CheckPermission(pluginID string, capability Capability) bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if grants, exists := pm.grants[pluginID]; exists {
		if grant, exists := grants[capability]; exists {
			if grant.Granted && !grant.Revoked {
				// Check expiration
				if grant.ExpiresAt != nil && time.Now().After(*grant.ExpiresAt) {
					return false
				}
				return true
			}
		}
	}

	return false
}

// GrantPermission grants permission for a capability
func (pm *DefaultPermissionManager) GrantPermission(ctx context.Context, grant *PermissionGrant) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	grant.Granted = true
	grant.GrantedAt = time.Now()
	pm.storeGrant(grant)

	// Remove from pending list
	for i, pending := range pm.pending {
		if pending.PluginID == grant.PluginID && pending.Capability == grant.Capability {
			pm.pending = append(pm.pending[:i], pm.pending[i+1:]...)
			break
		}
	}

	pm.logger.Info("Permission granted",
		zap.String("plugin_id", grant.PluginID),
		zap.String("capability", string(grant.Capability)))

	return pm.PersistPermissions()
}

// RevokePermission revokes permission for a capability
func (pm *DefaultPermissionManager) RevokePermission(ctx context.Context, pluginID string, capability Capability) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if grants, exists := pm.grants[pluginID]; exists {
		if grant, exists := grants[capability]; exists {
			now := time.Now()
			grant.Revoked = true
			grant.RevokedAt = &now

			pm.logger.Info("Permission revoked",
				zap.String("plugin_id", pluginID),
				zap.String("capability", string(capability)))

			return pm.PersistPermissions()
		}
	}

	return fmt.Errorf("permission not found")
}

// ListPermissions returns all permissions for a plugin
func (pm *DefaultPermissionManager) ListPermissions(pluginID string) []*PermissionGrant {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	if grants, exists := pm.grants[pluginID]; exists {
		permissions := make([]*PermissionGrant, 0, len(grants))
		for _, grant := range grants {
			permissions = append(permissions, grant)
		}
		return permissions
	}

	return []*PermissionGrant{}
}

// GetPendingPermissions returns all pending permission requests
func (pm *DefaultPermissionManager) GetPendingPermissions() []*PermissionGrant {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pending := make([]*PermissionGrant, len(pm.pending))
	copy(pending, pm.pending)
	return pending
}

// PersistPermissions saves permissions to disk
func (pm *DefaultPermissionManager) PersistPermissions() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(pm.persistPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Flatten grants for serialization
	allGrants := make([]*PermissionGrant, 0)
	for _, grants := range pm.grants {
		for _, grant := range grants {
			allGrants = append(allGrants, grant)
		}
	}

	data, err := json.MarshalIndent(allGrants, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(pm.persistPath, data, 0644)
}

// loadPermissions loads permissions from disk
func (pm *DefaultPermissionManager) loadPermissions() {
	data, err := ioutil.ReadFile(pm.persistPath)
	if err != nil {
		return // File doesn't exist yet
	}

	var allGrants []*PermissionGrant
	if err := json.Unmarshal(data, &allGrants); err != nil {
		pm.logger.Warn("Failed to load permissions", zap.Error(err))
		return
	}

	for _, grant := range allGrants {
		pm.storeGrant(grant)
	}

	pm.logger.Info("Loaded permissions", zap.Int("count", len(allGrants)))
}

// storeGrant stores a grant in memory
func (pm *DefaultPermissionManager) storeGrant(grant *PermissionGrant) {
	if pm.grants[grant.PluginID] == nil {
		pm.grants[grant.PluginID] = make(map[Capability]*PermissionGrant)
	}
	pm.grants[grant.PluginID][grant.Capability] = grant
}

// DefaultPanelRenderer implements the PanelRenderer interface
type DefaultPanelRenderer struct {
	zones  map[string]*PanelZone
	logger *zap.Logger
	mu     sync.RWMutex
}

// NewPanelRenderer creates a new panel renderer
func NewPanelRenderer(logger *zap.Logger) *DefaultPanelRenderer {
	return &DefaultPanelRenderer{
		zones:  make(map[string]*PanelZone),
		logger: logger,
	}
}

// CreateZone creates a new panel zone for a plugin
func (pr *DefaultPanelRenderer) CreateZone(pluginID string, width, height int) (*PanelZone, error) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	zoneID := fmt.Sprintf("%s-zone", pluginID)
	zone := &PanelZone{
		ID:       zoneID,
		X:        0, // Will be positioned by layout manager
		Y:        0,
		Width:    width,
		Height:   height,
		PluginID: pluginID,
		Title:    pluginID,
		Border:   true,
		Focused:  false,
		Visible:  true,
		ZIndex:   1,
	}

	pr.zones[zoneID] = zone
	pr.logger.Debug("Created panel zone",
		zap.String("zone_id", zoneID),
		zap.String("plugin_id", pluginID))

	return zone, nil
}

// ResizeZone resizes a panel zone
func (pr *DefaultPanelRenderer) ResizeZone(zoneID string, width, height int) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	if zone, exists := pr.zones[zoneID]; exists {
		zone.Width = width
		zone.Height = height
		return nil
	}

	return fmt.Errorf("zone not found: %s", zoneID)
}

// FocusZone sets focus to a panel zone
func (pr *DefaultPanelRenderer) FocusZone(zoneID string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	// Unfocus all zones
	for _, zone := range pr.zones {
		zone.Focused = false
	}

	// Focus the specified zone
	if zone, exists := pr.zones[zoneID]; exists {
		zone.Focused = true
		return nil
	}

	return fmt.Errorf("zone not found: %s", zoneID)
}

// RenderCommand processes a render command from a plugin
func (pr *DefaultPanelRenderer) RenderCommand(ctx context.Context, cmd *RenderCommand) error {
	pr.logger.Debug("Processing render command",
		zap.String("type", string(cmd.Type)),
		zap.String("zone", cmd.Zone))

	// In a real implementation, this would integrate with the TUI framework
	// For now, just log the command
	switch cmd.Type {
	case RenderTypeText:
		pr.logger.Debug("Rendering text",
			zap.String("text", cmd.Text),
			zap.Int("x", cmd.X),
			zap.Int("y", cmd.Y))
	case RenderTypeClear:
		pr.logger.Debug("Clearing zone", zap.String("zone", cmd.Zone))
	default:
		pr.logger.Debug("Unsupported render command", zap.String("type", string(cmd.Type)))
	}

	return nil
}

// ClearZone clears the contents of a panel zone
func (pr *DefaultPanelRenderer) ClearZone(zoneID string) error {
	pr.logger.Debug("Clearing zone", zap.String("zone_id", zoneID))
	return nil
}

// DestroyZone removes a panel zone
func (pr *DefaultPanelRenderer) DestroyZone(zoneID string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	delete(pr.zones, zoneID)
	pr.logger.Debug("Destroyed zone", zap.String("zone_id", zoneID))
	return nil
}

// GetZone retrieves a panel zone by ID
func (pr *DefaultPanelRenderer) GetZone(zoneID string) (*PanelZone, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	if zone, exists := pr.zones[zoneID]; exists {
		return zone, nil
	}

	return nil, fmt.Errorf("zone not found: %s", zoneID)
}

// ListZones returns all panel zones
func (pr *DefaultPanelRenderer) ListZones() []*PanelZone {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	zones := make([]*PanelZone, 0, len(pr.zones))
	for _, zone := range pr.zones {
		zones = append(zones, zone)
	}

	return zones
}

// DefaultPluginRegistry implements the PluginRegistry interface
type DefaultPluginRegistry struct {
	plugins map[string]*Plugin // pluginID -> plugin
	logger  *zap.Logger
	mu      sync.RWMutex
}

// NewPluginRegistry creates a new plugin registry
func NewPluginRegistry(logger *zap.Logger) *DefaultPluginRegistry {
	return &DefaultPluginRegistry{
		plugins: make(map[string]*Plugin),
		logger:  logger,
	}
}

// Register registers a plugin
func (pr *DefaultPluginRegistry) Register(plugin *Plugin) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.plugins[plugin.ID] = plugin
	pr.logger.Debug("Registered plugin", zap.String("plugin_id", plugin.ID))
	return nil
}

// Unregister removes a plugin from the registry
func (pr *DefaultPluginRegistry) Unregister(pluginID string) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	delete(pr.plugins, pluginID)
	pr.logger.Debug("Unregistered plugin", zap.String("plugin_id", pluginID))
	return nil
}

// Find finds a plugin by name and version
func (pr *DefaultPluginRegistry) Find(name, version string) (*Plugin, error) {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	for _, plugin := range pr.plugins {
		if plugin.Name == name && (version == "" || plugin.Version == version) {
			return plugin, nil
		}
	}

	return nil, fmt.Errorf("plugin not found: %s@%s", name, version)
}

// List returns all registered plugins
func (pr *DefaultPluginRegistry) List() []*Plugin {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	plugins := make([]*Plugin, 0, len(pr.plugins))
	for _, plugin := range pr.plugins {
		plugins = append(plugins, plugin)
	}

	return plugins
}

// Search searches for plugins by query
func (pr *DefaultPluginRegistry) Search(query string) []*Plugin {
	// Simple implementation - search in name and description
	plugins := pr.List()
	results := make([]*Plugin, 0)

	for _, plugin := range plugins {
		if contains(plugin.Name, query) || contains(plugin.Description, query) {
			results = append(results, plugin)
		}
	}

	return results
}

// Update updates plugin information
func (pr *DefaultPluginRegistry) Update(plugin *Plugin) error {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.plugins[plugin.ID] = plugin
	return nil
}

// GetDependencies returns plugin dependencies (placeholder)
func (pr *DefaultPluginRegistry) GetDependencies(pluginID string) []*Plugin {
	// Placeholder implementation
	return []*Plugin{}
}

// ValidateManifest validates a plugin manifest
func (pr *DefaultPluginRegistry) ValidateManifest(manifest *PluginManifest) error {
	if manifest.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if manifest.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	if manifest.EntryPoint == "" {
		return fmt.Errorf("plugin entry point is required")
	}
	if manifest.Runtime == "" {
		return fmt.Errorf("plugin runtime is required")
	}

	// Validate capabilities
	for _, capability := range manifest.Capabilities {
		if !IsValidCapability(capability) {
			return fmt.Errorf("invalid capability: %s", capability)
		}
	}

	return nil
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				(len(s) > len(substr)*2 && s[len(s)/2-len(substr)/2:len(s)/2+len(substr)/2+len(substr)%2] == substr))))
}