// Copyright 2025 James Ross
package pluginpanel

import (
	"context"
	"time"
)

// Plugin represents a loaded plugin with its metadata and runtime
type Plugin struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Author      string            `json:"author"`
	Description string            `json:"description"`
	Runtime     Runtime           `json:"runtime"`
	Status      PluginStatus      `json:"status"`
	Manifest    *PluginManifest   `json:"manifest"`
	Instance    PluginInstance    `json:"-"`
	LoadedAt    time.Time         `json:"loaded_at"`
	LastUpdate  time.Time         `json:"last_update"`
	ErrorCount  int               `json:"error_count"`
	Metadata    map[string]string `json:"metadata"`
}

// PluginManifest describes a plugin's requirements and capabilities
type PluginManifest struct {
	Name         string              `yaml:"name" json:"name"`
	Version      string              `yaml:"version" json:"version"`
	Author       string              `yaml:"author" json:"author"`
	Description  string              `yaml:"description" json:"description"`
	Runtime      Runtime             `yaml:"runtime" json:"runtime"`
	EntryPoint   string              `yaml:"entry_point" json:"entry_point"`
	Capabilities []Capability        `yaml:"capabilities" json:"capabilities"`
	Resources    ResourceRequirement `yaml:"resources" json:"resources"`
	Dependencies map[string]string   `yaml:"dependencies" json:"dependencies"`
	Assets       []string            `yaml:"assets" json:"assets"`
	Schema       string              `yaml:"schema" json:"schema"`
	Metadata     map[string]string   `yaml:"metadata" json:"metadata"`
}

// Runtime defines the plugin execution environment
type Runtime string

const (
	RuntimeWASM     Runtime = "wasm"
	RuntimeStarlark Runtime = "starlark"
	RuntimeLua      Runtime = "lua"
	RuntimeNative   Runtime = "native" // For development only
)

// PluginStatus represents the current state of a plugin
type PluginStatus string

const (
	StatusUnloaded   PluginStatus = "unloaded"
	StatusLoading    PluginStatus = "loading"
	StatusReady      PluginStatus = "ready"
	StatusRunning    PluginStatus = "running"
	StatusError      PluginStatus = "error"
	StatusDisabled   PluginStatus = "disabled"
	StatusSandboxed  PluginStatus = "sandboxed"
	StatusTerminated PluginStatus = "terminated"
)

// Capability represents a permission that plugins can request
type Capability string

const (
	// Read-only capabilities
	CapabilityReadStats     Capability = "read_stats"
	CapabilityReadKeys      Capability = "read_keys"
	CapabilityReadSelection Capability = "read_selection"
	CapabilityReadTimers    Capability = "read_timers"
	CapabilityReadQueues    Capability = "read_queues"
	CapabilityReadJobs      Capability = "read_jobs"

	// Action capabilities (require explicit user grants)
	CapabilityEnqueue    Capability = "enqueue"
	CapabilityPeek       Capability = "peek"
	CapabilityRequeue    Capability = "requeue"
	CapabilityPurge      Capability = "purge"
	CapabilityModifyJob  Capability = "modify_job"
	CapabilityDeleteJob  Capability = "delete_job"

	// UI capabilities
	CapabilityRenderPanel Capability = "render_panel"
	CapabilityKeyEvents   Capability = "key_events"
	CapabilityMouseEvents Capability = "mouse_events"
	CapabilityDialogs     Capability = "dialogs"

	// System capabilities (highly restricted)
	CapabilityFileAccess    Capability = "file_access"
	CapabilityNetworkAccess Capability = "network_access"
	CapabilitySystemExec    Capability = "system_exec"
)

// ResourceRequirement defines the resource limits for a plugin
type ResourceRequirement struct {
	MaxMemoryMB     int           `yaml:"max_memory_mb" json:"max_memory_mb"`
	MaxCPUPercent   int           `yaml:"max_cpu_percent" json:"max_cpu_percent"`
	MaxExecutionMs  int           `yaml:"max_execution_ms" json:"max_execution_ms"`
	MaxGoroutines   int           `yaml:"max_goroutines" json:"max_goroutines"`
	Timeout         time.Duration `yaml:"timeout" json:"timeout"`
	IdleTimeout     time.Duration `yaml:"idle_timeout" json:"idle_timeout"`
	HeartbeatPeriod time.Duration `yaml:"heartbeat_period" json:"heartbeat_period"`
}

// PermissionGrant represents a user's decision about a capability
type PermissionGrant struct {
	PluginID     string     `json:"plugin_id"`
	Capability   Capability `json:"capability"`
	Granted      bool       `json:"granted"`
	GrantedAt    time.Time  `json:"granted_at"`
	GrantedBy    string     `json:"granted_by"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	Conditions   []string   `json:"conditions,omitempty"`
	Reason       string     `json:"reason,omitempty"`
	LastUsed     *time.Time `json:"last_used,omitempty"`
	UsageCount   int        `json:"usage_count"`
	Revoked      bool       `json:"revoked"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	RevokedBy    string     `json:"revoked_by,omitempty"`
	RevokeReason string     `json:"revoke_reason,omitempty"`
}

// PanelZone represents a rendering area for plugins
type PanelZone struct {
	ID       string `json:"id"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	PluginID string `json:"plugin_id"`
	Title    string `json:"title"`
	Border   bool   `json:"border"`
	Focused  bool   `json:"focused"`
	Visible  bool   `json:"visible"`
	ZIndex   int    `json:"z_index"`
}

// Event represents system events that plugins can subscribe to
type Event struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
	PluginID  string                 `json:"plugin_id,omitempty"`
	SessionID string                 `json:"session_id,omitempty"`
}

// EventType defines the types of events plugins can receive
type EventType string

const (
	EventTypeStats       EventType = "stats"
	EventTypeSelection   EventType = "selection"
	EventTypeTimer       EventType = "timer"
	EventTypeKeypress    EventType = "keypress"
	EventTypeMouseClick  EventType = "mouse_click"
	EventTypeJobEnqueued EventType = "job_enqueued"
	EventTypeJobComplete EventType = "job_complete"
	EventTypeJobFailed   EventType = "job_failed"
	EventTypeQueueEmpty  EventType = "queue_empty"
	EventTypePluginLoad  EventType = "plugin_load"
	EventTypePluginError EventType = "plugin_error"
	EventTypeShutdown    EventType = "shutdown"
)

// RenderCommand represents a drawing instruction from a plugin
type RenderCommand struct {
	Type   RenderType             `json:"type"`
	Zone   string                 `json:"zone"`
	X      int                    `json:"x"`
	Y      int                    `json:"y"`
	Text   string                 `json:"text,omitempty"`
	Style  RenderStyle            `json:"style,omitempty"`
	Props  map[string]interface{} `json:"props,omitempty"`
	Buffer [][]rune               `json:"buffer,omitempty"`
}

// RenderType defines the types of rendering operations
type RenderType string

const (
	RenderTypeText      RenderType = "text"
	RenderTypeLine      RenderType = "line"
	RenderTypeRect      RenderType = "rect"
	RenderTypeTable     RenderType = "table"
	RenderTypeChart     RenderType = "chart"
	RenderTypeProgress  RenderType = "progress"
	RenderTypeClear     RenderType = "clear"
	RenderTypeBuffer    RenderType = "buffer"
	RenderTypeComponent RenderType = "component"
)

// RenderStyle defines text styling options
type RenderStyle struct {
	Foreground string `json:"foreground,omitempty"`
	Background string `json:"background,omitempty"`
	Bold       bool   `json:"bold,omitempty"`
	Italic     bool   `json:"italic,omitempty"`
	Underline  bool   `json:"underline,omitempty"`
	Blink      bool   `json:"blink,omitempty"`
	Reverse    bool   `json:"reverse,omitempty"`
}

// PluginConfig holds configuration for the plugin system
type PluginConfig struct {
	PluginDir          string        `yaml:"plugin_dir" json:"plugin_dir"`
	MaxPlugins         int           `yaml:"max_plugins" json:"max_plugins"`
	HotReload          bool          `yaml:"hot_reload" json:"hot_reload"`
	SandboxEnabled     bool          `yaml:"sandbox_enabled" json:"sandbox_enabled"`
	DefaultResources   ResourceRequirement `yaml:"default_resources" json:"default_resources"`
	PermissionTimeout  time.Duration `yaml:"permission_timeout" json:"permission_timeout"`
	EventQueueSize     int           `yaml:"event_queue_size" json:"event_queue_size"`
	PluginTimeout      time.Duration `yaml:"plugin_timeout" json:"plugin_timeout"`
	GCInterval         time.Duration `yaml:"gc_interval" json:"gc_interval"`
	LogLevel           string        `yaml:"log_level" json:"log_level"`
	TrustedPlugins     []string      `yaml:"trusted_plugins" json:"trusted_plugins"`
	DefaultPermissions []Capability  `yaml:"default_permissions" json:"default_permissions"`
}

// Interfaces for plugin system components

// PluginManager manages the lifecycle of plugins
type PluginManager interface {
	LoadPlugin(ctx context.Context, path string) (*Plugin, error)
	UnloadPlugin(ctx context.Context, pluginID string) error
	GetPlugin(pluginID string) (*Plugin, error)
	ListPlugins() []*Plugin
	ReloadPlugin(ctx context.Context, pluginID string) error
	EnablePlugin(ctx context.Context, pluginID string) error
	DisablePlugin(ctx context.Context, pluginID string) error
	RegisterPlugin(plugin *Plugin) error
	UnregisterPlugin(pluginID string) error
}

// PluginInstance represents a running plugin
type PluginInstance interface {
	ID() string
	Status() PluginStatus
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Call(ctx context.Context, method string, args interface{}) (interface{}, error)
	SendEvent(ctx context.Context, event *Event) error
	GetMetrics() map[string]interface{}
	Cleanup() error
}

// RuntimeEngine provides the execution environment for plugins
type RuntimeEngine interface {
	LoadPlugin(ctx context.Context, manifest *PluginManifest, code []byte) (PluginInstance, error)
	SupportedRuntime() Runtime
	ValidateCode(code []byte) error
	GetCapabilities() []Capability
}

// EventBus handles event distribution to plugins
type EventBus interface {
	Subscribe(pluginID string, eventTypes []EventType) error
	Unsubscribe(pluginID string, eventTypes []EventType) error
	Publish(ctx context.Context, event *Event) error
	SendToPlugin(ctx context.Context, pluginID string, event *Event) error
	GetSubscriptions(pluginID string) []EventType
	Close() error
}

// PermissionManager handles capability grants and security
type PermissionManager interface {
	RequestPermission(ctx context.Context, pluginID string, capability Capability) (*PermissionGrant, error)
	CheckPermission(pluginID string, capability Capability) bool
	GrantPermission(ctx context.Context, grant *PermissionGrant) error
	RevokePermission(ctx context.Context, pluginID string, capability Capability) error
	ListPermissions(pluginID string) []*PermissionGrant
	GetPendingPermissions() []*PermissionGrant
	PersistPermissions() error
}

// PanelRenderer handles the visual representation of plugin panels
type PanelRenderer interface {
	CreateZone(pluginID string, width, height int) (*PanelZone, error)
	ResizeZone(zoneID string, width, height int) error
	FocusZone(zoneID string) error
	RenderCommand(ctx context.Context, cmd *RenderCommand) error
	ClearZone(zoneID string) error
	DestroyZone(zoneID string) error
	GetZone(zoneID string) (*PanelZone, error)
	ListZones() []*PanelZone
}

// Sandbox provides isolation and resource limiting for plugins
type Sandbox interface {
	CreateContainer(pluginID string, resources *ResourceRequirement) error
	StartPlugin(ctx context.Context, pluginID string, instance PluginInstance) error
	StopPlugin(ctx context.Context, pluginID string) error
	MonitorResources(pluginID string) (*ResourceUsage, error)
	EnforceTimeout(pluginID string, timeout time.Duration) error
	KillPlugin(pluginID string) error
}

// ResourceUsage tracks plugin resource consumption
type ResourceUsage struct {
	PluginID      string        `json:"plugin_id"`
	MemoryUsageMB float64       `json:"memory_usage_mb"`
	CPUPercent    float64       `json:"cpu_percent"`
	ExecutionTime time.Duration `json:"execution_time"`
	Goroutines    int           `json:"goroutines"`
	LastUpdate    time.Time     `json:"last_update"`
	Violations    []string      `json:"violations,omitempty"`
}

// HostAPI provides the capability-gated API surface for plugins
type HostAPI interface {
	// Stats and monitoring
	GetStats(ctx context.Context) (map[string]interface{}, error)
	GetQueues(ctx context.Context) ([]string, error)
	GetJobs(ctx context.Context, queue string, limit int) ([]interface{}, error)

	// Actions (capability-gated)
	EnqueueJob(ctx context.Context, queue string, payload interface{}) error
	PeekJob(ctx context.Context, queue string) (interface{}, error)
	RequeueJob(ctx context.Context, jobID string) error
	PurgeQueue(ctx context.Context, queue string) error

	// UI operations
	RenderPanel(ctx context.Context, commands []*RenderCommand) error
	ShowDialog(ctx context.Context, title, message string) error
	GetSelection(ctx context.Context) (interface{}, error)

	// Event handling
	Subscribe(ctx context.Context, eventTypes []EventType) error
	Unsubscribe(ctx context.Context, eventTypes []EventType) error
}

// PluginRegistry maintains information about available and installed plugins
type PluginRegistry interface {
	Register(plugin *Plugin) error
	Unregister(pluginID string) error
	Find(name, version string) (*Plugin, error)
	List() []*Plugin
	Search(query string) []*Plugin
	Update(plugin *Plugin) error
	GetDependencies(pluginID string) []*Plugin
	ValidateManifest(manifest *PluginManifest) error
}

// HotReloader watches for plugin file changes and reloads them
type HotReloader interface {
	Watch(pluginDir string) error
	Stop() error
	OnChange(callback func(path string, event string)) error
}

// PluginError represents errors that occur during plugin operations
type PluginError struct {
	PluginID  string    `json:"plugin_id"`
	Operation string    `json:"operation"`
	Message   string    `json:"message"`
	Code      ErrorCode `json:"code"`
	Timestamp time.Time `json:"timestamp"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Err       error     `json:"-"`
}

func (e *PluginError) Error() string {
	if e.PluginID != "" {
		return "plugin " + e.PluginID + " " + e.Operation + " failed: " + e.Message
	}
	return "plugin " + e.Operation + " failed: " + e.Message
}

func (e *PluginError) Unwrap() error {
	return e.Err
}

// ErrorCode represents different types of plugin errors
type ErrorCode string

const (
	ErrorCodeLoadFailed       ErrorCode = "LOAD_FAILED"
	ErrorCodeRuntimeError     ErrorCode = "RUNTIME_ERROR"
	ErrorCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	ErrorCodeResourceExceeded ErrorCode = "RESOURCE_EXCEEDED"
	ErrorCodeTimeout          ErrorCode = "TIMEOUT"
	ErrorCodeSandboxViolation ErrorCode = "SANDBOX_VIOLATION"
	ErrorCodeInvalidManifest  ErrorCode = "INVALID_MANIFEST"
	ErrorCodeUnsupportedAPI   ErrorCode = "UNSUPPORTED_API"
	ErrorCodeDependencyError  ErrorCode = "DEPENDENCY_ERROR"
	ErrorCodeValidationFailed ErrorCode = "VALIDATION_FAILED"
)

// Common validation functions

// IsValidCapability checks if a capability is recognized
func IsValidCapability(cap Capability) bool {
	switch cap {
	case CapabilityReadStats, CapabilityReadKeys, CapabilityReadSelection, CapabilityReadTimers,
		 CapabilityReadQueues, CapabilityReadJobs, CapabilityEnqueue, CapabilityPeek,
		 CapabilityRequeue, CapabilityPurge, CapabilityModifyJob, CapabilityDeleteJob,
		 CapabilityRenderPanel, CapabilityKeyEvents, CapabilityMouseEvents, CapabilityDialogs,
		 CapabilityFileAccess, CapabilityNetworkAccess, CapabilitySystemExec:
		return true
	default:
		return false
	}
}

// IsReadOnlyCapability determines if a capability is read-only (safe)
func IsReadOnlyCapability(cap Capability) bool {
	switch cap {
	case CapabilityReadStats, CapabilityReadKeys, CapabilityReadSelection,
		 CapabilityReadTimers, CapabilityReadQueues, CapabilityReadJobs:
		return true
	default:
		return false
	}
}

// IsActionCapability determines if a capability performs actions
func IsActionCapability(cap Capability) bool {
	switch cap {
	case CapabilityEnqueue, CapabilityPeek, CapabilityRequeue,
		 CapabilityPurge, CapabilityModifyJob, CapabilityDeleteJob:
		return true
	default:
		return false
	}
}

// IsSystemCapability determines if a capability is system-level
func IsSystemCapability(cap Capability) bool {
	switch cap {
	case CapabilityFileAccess, CapabilityNetworkAccess, CapabilitySystemExec:
		return true
	default:
		return false
	}
}

// DefaultResourceRequirements returns safe default resource limits
func DefaultResourceRequirements() ResourceRequirement {
	return ResourceRequirement{
		MaxMemoryMB:     64,
		MaxCPUPercent:   25,
		MaxExecutionMs:  5000,
		MaxGoroutines:   10,
		Timeout:         30 * time.Second,
		IdleTimeout:     5 * time.Minute,
		HeartbeatPeriod: 30 * time.Second,
	}
}

// DefaultPluginConfig returns a safe default plugin configuration
func DefaultPluginConfig() PluginConfig {
	return PluginConfig{
		PluginDir:         "./plugins",
		MaxPlugins:        20,
		HotReload:         true,
		SandboxEnabled:    true,
		DefaultResources:  DefaultResourceRequirements(),
		PermissionTimeout: 30 * time.Second,
		EventQueueSize:    1000,
		PluginTimeout:     60 * time.Second,
		GCInterval:        5 * time.Minute,
		LogLevel:          "info",
		TrustedPlugins:    []string{},
		DefaultPermissions: []Capability{
			CapabilityReadStats,
			CapabilityReadKeys,
			CapabilityRenderPanel,
		},
	}
}