# Plugin Panel System (F005) - Design Document

**Version:** 1.0
**Date:** 2025-09-14
**Status:** Draft
**Author:** Claude (Worker 6)
**Reviewers:** TBD

## Executive Summary

The Plugin Panel System transforms the go-redis-work-queue TUI from a static monitoring tool into a dynamic, extensible platform. This revolutionary feature enables organizations to create custom panels for proprietary metrics, specialized actions, and organizational workflows without modifying the core codebase.

Built on WebAssembly (WASM) and Starlark runtimes, the system provides a secure sandbox environment where plugins can render custom UI components, subscribe to queue events, and execute administrative actions through a capability-gated API. This architecture ensures safety, portability, and performance while enabling unlimited customization.

The system represents a strategic platform play that positions the queue system as an extensible infrastructure tool rather than a static application. By enabling third-party development and custom integrations, it creates network effects that increase adoption and reduce churn.

### Key Benefits

- **Infinite Customization**: Organizations can build panels for any metric, visualization, or workflow
- **Enterprise Lock-in**: Custom plugins create switching costs and organizational dependencies
- **Developer Ecosystem**: Third-party plugins expand functionality without core development overhead
- **Zero-Risk Extension**: Sandboxed plugins cannot crash or compromise the host application
- **Hot Development**: Real-time plugin reloading enables rapid iteration and deployment
- **Production Safety**: Comprehensive resource limits and permission controls ensure stability

### Architecture Overview

```mermaid
graph TB
    subgraph "Host Application"
        A[TUI Framework]
        B[Event Bus]
        C[Admin API Gateway]
        D[Permission Manager]
    end

    subgraph "Plugin Runtime"
        E[WASM Runtime]
        F[Starlark Runtime]
        G[Plugin Loader]
        H[Resource Monitor]
    end

    subgraph "Plugin Ecosystem"
        I[Plugin A<br/>Tenant SLA Monitor]
        J[Plugin B<br/>Bulk Operations]
        K[Plugin C<br/>Custom Metrics]
        L[Plugin D<br/>Alert Manager]
    end

    subgraph "Plugin Infrastructure"
        M[Plugin Registry]
        N[Manifest Validator]
        O[Capability Store]
        P[Hot Reload Engine]
    end

    A --> B
    B --> E
    B --> F
    C --> D
    D --> E
    D --> F
    G --> E
    G --> F
    H --> E
    H --> F
    E --> I
    E --> J
    F --> K
    F --> L
    M --> G
    N --> G
    O --> D
    P --> G

    style A fill:#e1f5fe
    style B fill:#e1f5fe
    style C fill:#e1f5fe
    style D fill:#e1f5fe
    style E fill:#f3e5f5
    style F fill:#f3e5f5
    style G fill:#f3e5f5
    style H fill:#f3e5f5
    style I fill:#fff3e0
    style J fill:#fff3e0
    style K fill:#fff3e0
    style L fill:#fff3e0
    style M fill:#e8f5e8
    style N fill:#e8f5e8
    style O fill:#e8f5e8
    style P fill:#e8f5e8
```

## System Architecture

### Core Components

#### 1. Plugin Runtime Engine

The runtime engine provides secure execution environments for plugins using multiple runtime options to balance performance, security, and developer experience.

**WASM Runtime (Primary)**
- **Technology**: Wasmtime with WASI support
- **Languages**: TinyGo, Rust, C/C++ compiled to WASM
- **Advantages**: Near-native performance, strong sandboxing, wide language support
- **Use Cases**: Performance-critical plugins, complex data processing

**Starlark Runtime (Alternative)**
- **Technology**: Go Starlark interpreter
- **Language**: Python-like syntax with Go integration
- **Advantages**: Simple deployment, no compilation step, excellent for scripts
- **Use Cases**: Configuration-driven panels, simple data transformations

```go
type PluginRuntime interface {
    LoadPlugin(manifest *PluginManifest, source []byte) (*Plugin, error)
    ExecuteFunction(plugin *Plugin, function string, args []interface{}) (interface{}, error)
    RegisterHostFunction(name string, handler HostFunction) error
    SetResourceLimits(limits ResourceLimits) error
    Destroy() error
}

type ResourceLimits struct {
    MaxMemoryBytes    int64
    MaxExecutionTime  time.Duration
    MaxCPUPercent     float64
    AllowedOperations []string
}
```

#### 2. Event Bus and Subscription System

The event bus delivers real-time queue events to plugins through a typed subscription system that ensures plugins receive only relevant data.

**Event Types**:
- Queue statistics (depth, throughput, error rates)
- Job lifecycle events (enqueued, started, completed, failed)
- Worker status updates (active, idle, error states)
- Admin operations (configuration changes, manual interventions)
- System health metrics (memory usage, performance indicators)

```go
type EventBus struct {
    subscriptions map[string][]Subscription
    filters       map[string]EventFilter
    rateLimits    map[string]RateLimit
}

type Subscription struct {
    PluginID     string
    EventTypes   []string
    QueueFilters []string
    Callback     EventCallback
    Priority     int
}

type Event struct {
    Type      string      `json:"type"`
    Timestamp time.Time   `json:"timestamp"`
    Source    string      `json:"source"`
    Data      interface{} `json:"data"`
    Metadata  EventMeta   `json:"metadata"`
}
```

#### 3. Capability-Gated Admin API

The admin API provides controlled access to queue operations through a fine-grained permission system that plugins must explicitly request and users must approve.

**Permission Levels**:
- **Read-Only**: Subscribe to events, query statistics, inspect job metadata
- **Queue Operations**: Enqueue jobs, peek at job contents, query job status
- **Administrative**: Requeue failed jobs, purge queues, modify worker settings
- **System**: Access configuration, restart services, modify system settings

```go
type CapabilityGrant struct {
    PluginID    string      `json:"plugin_id"`
    Permissions []string    `json:"permissions"`
    Scope       PermScope   `json:"scope"`
    ExpiresAt   *time.Time  `json:"expires_at,omitempty"`
    GrantedBy   string      `json:"granted_by"`
    GrantedAt   time.Time   `json:"granted_at"`
}

type PermScope struct {
    Queues     []string `json:"queues,omitempty"`     // Specific queues or "*"
    Operations []string `json:"operations,omitempty"` // Specific ops or "*"
    Resources  []string `json:"resources,omitempty"`  // Resource limitations
}
```

#### 4. Plugin Lifecycle Manager

The lifecycle manager handles plugin discovery, loading, validation, and hot-reloading with zero downtime and automatic rollback on failures.

**Discovery Process**:
1. Scan `plugins/` directory for plugin bundles
2. Validate plugin manifests and security requirements
3. Check compatibility with host API version
4. Load plugins in dependency order
5. Initialize plugin panels and register event subscriptions

```go
type PluginManager struct {
    plugins     map[string]*LoadedPlugin
    manifests   map[string]*PluginManifest
    watcher     *fsnotify.Watcher
    runtime     PluginRuntime
    validator   ManifestValidator
    permissions PermissionStore
}

type LoadedPlugin struct {
    ID          string
    Manifest    *PluginManifest
    Instance    PluginInstance
    Permissions []CapabilityGrant
    Status      PluginStatus
    LoadedAt    time.Time
    LastActive  time.Time
    ErrorCount  int
}
```

### Plugin Architecture

#### Plugin Manifest Structure

Every plugin includes a YAML manifest that declares its requirements, capabilities, and metadata.

```yaml
# manifest.yaml
name: "tenant-sla-monitor"
version: "1.2.3"
api_version: "v1"
description: "Real-time SLA monitoring and alerting for tenant queues"

# Plugin metadata
author: "Platform Team"
license: "MIT"
homepage: "https://github.com/company/tenant-sla-plugin"
tags: ["monitoring", "sla", "alerts"]

# Runtime configuration
runtime: "wasm"  # or "starlark"
entry_point: "main.wasm"  # or "main.star"

# Requested capabilities (user must approve)
permissions:
  - "events.subscribe"      # Read queue events
  - "admin.queue.peek"      # Peek at job contents
  - "admin.alerts.send"     # Send alerts (requires approval)

# Resource limits
resources:
  max_memory: "10MB"
  max_cpu: "5%"
  timeout: "30s"

# UI configuration
ui:
  panel_title: "SLA Monitor"
  min_width: 40
  min_height: 10
  resizable: true

# Event subscriptions
events:
  - type: "job.completed"
    queues: ["tenant.*"]
  - type: "queue.stats"
    interval: "5s"

# Dependencies (other plugins)
dependencies:
  - name: "alert-manager"
    version: ">=1.0.0"
    optional: false
```

#### Plugin Bundle Structure

```
tenant-sla-monitor.zip
├── manifest.yaml           # Plugin metadata and requirements
├── main.wasm              # Compiled plugin binary (or main.star)
├── assets/
│   ├── styles.css         # Optional styling
│   └── config.json        # Default configuration
├── docs/
│   ├── README.md          # Plugin documentation
│   └── examples/          # Usage examples
└── tests/
    └── integration/       # Plugin tests
```

#### Plugin API Interface

Plugins implement a standard interface that provides lifecycle hooks and event handlers.

```go
// Plugin interface (exposed to WASM/Starlark)
type PluginInterface interface {
    // Lifecycle hooks
    Initialize(config map[string]interface{}) error
    Activate() error
    Deactivate() error
    Cleanup() error

    // Event handling
    OnEvent(event Event) error
    OnTimer(timer Timer) error

    // UI rendering
    Render(context RenderContext) ([]UIElement, error)
    OnInput(input InputEvent) error

    // Configuration
    GetConfig() map[string]interface{}
    SetConfig(config map[string]interface{}) error

    // Health monitoring
    HealthCheck() HealthStatus
}

// Host API (available to plugins)
type HostAPI interface {
    // Event system
    Subscribe(eventType string, callback EventCallback) error
    Unsubscribe(subscriptionID string) error

    // Admin operations (capability-gated)
    EnqueueJob(queue string, job JobData) error
    PeekJob(queue string) (*Job, error)
    RequeueJob(jobID string) error
    PurgeQueue(queue string) error

    // Statistics
    GetQueueStats(queue string) (*QueueStats, error)
    GetSystemStats() (*SystemStats, error)

    // UI helpers
    Log(level LogLevel, message string) error
    ShowNotification(message string, level NotificationLevel) error

    // Storage (plugin-scoped key-value store)
    Get(key string) (string, error)
    Set(key string, value string) error
    Delete(key string) error
}
```

### Data Flow Architecture

```mermaid
sequenceDiagram
    participant U as User
    participant H as Host App
    participant PM as Plugin Manager
    participant R as Runtime
    participant P as Plugin
    participant API as Admin API

    Note over U,API: Plugin Installation
    U->>H: Install plugin bundle
    H->>PM: Load plugin
    PM->>PM: Validate manifest
    PM->>U: Request permission approval
    U->>PM: Grant capabilities
    PM->>R: Initialize runtime
    R->>P: Load plugin code
    P->>R: Register event handlers
    R->>PM: Plugin ready
    PM->>H: Plugin activated

    Note over U,API: Runtime Operation
    H->>P: OnEvent(queue.stats)
    P->>P: Process statistics
    P->>R: Render(panel_data)
    R->>H: Return UI elements
    H->>U: Display plugin panel

    Note over U,API: Plugin Action
    U->>P: OnInput(button_click)
    P->>API: RequeueJob(job_id)
    API->>API: Check permissions
    API->>H: Execute operation
    H->>P: OnEvent(job.requeued)
```

### Security Model

#### Sandboxing and Isolation

**WASM Sandbox**:
- Memory isolation with configurable limits
- No direct system call access
- Host function whitelist enforcement
- CPU time limiting with preemption
- Deterministic execution for testing

**Starlark Sandbox**:
- No file system access
- No network operations
- No import of external modules
- Limited standard library access
- Configurable execution timeouts

```go
type SecurityPolicy struct {
    AllowedHostFunctions []string
    ResourceLimits       ResourceLimits
    NetworkAccess        bool
    FileSystemAccess     bool
    SystemCallAccess     bool
    EnvironmentAccess    []string
}

type SandboxManager struct {
    policies map[string]SecurityPolicy
    monitors map[string]*ResourceMonitor
    isolator RuntimeIsolator
}
```

#### Permission System

**Capability Model**:
- Explicit permission requests in manifest
- User approval required for sensitive operations
- Granular scope controls (queues, operations, resources)
- Permission inheritance and delegation
- Audit logging for all permission grants and usage

**Permission Categories**:
- **Observer**: Read-only access to events and statistics
- **Operator**: Basic queue operations (enqueue, peek, requeue)
- **Administrator**: Advanced operations (purge, configuration)
- **System**: Host-level operations (restart, system configuration)

```go
type PermissionManager struct {
    grants    map[string][]CapabilityGrant
    auditor   AuditLogger
    validator PermissionValidator
    store     PermissionStore
}

func (pm *PermissionManager) CheckPermission(
    pluginID string,
    operation string,
    resource string,
) (bool, error) {
    grants := pm.grants[pluginID]
    for _, grant := range grants {
        if pm.matchesGrant(grant, operation, resource) {
            pm.auditor.LogAccess(pluginID, operation, resource, "granted")
            return true, nil
        }
    }
    pm.auditor.LogAccess(pluginID, operation, resource, "denied")
    return false, ErrPermissionDenied
}
```

## Performance Requirements

### Latency Requirements

- **Plugin Loading**: <500ms for typical plugins (WASM), <100ms (Starlark)
- **Event Delivery**: <10ms from host event to plugin delivery
- **UI Rendering**: <16ms for 60fps UI updates (target 33ms for 30fps minimum)
- **API Calls**: <5ms for simple operations, <50ms for complex operations

### Throughput Requirements

- **Event Processing**: 1,000+ events/second per plugin
- **Concurrent Plugins**: Support 50+ active plugins simultaneously
- **API Operations**: 500+ operations/second across all plugins
- **UI Updates**: 30+ fps for responsive user interface

### Resource Requirements

- **Memory Overhead**: <50MB total for plugin system, <5MB per plugin
- **CPU Overhead**: <10% of system CPU for plugin management
- **Storage**: <100MB for plugin system, configurable per-plugin limits
- **Network**: Minimal - plugins communicate through host API only

### Scalability Requirements

- **Plugin Count**: Support 100+ installed plugins
- **Event Subscriptions**: 1,000+ concurrent event subscriptions
- **Panel Rendering**: 20+ active panels simultaneously
- **Hot Reload**: <1 second for plugin updates

## Testing Strategy

### Unit Testing

- Plugin runtime execution and sandboxing
- Permission system and capability validation
- Event bus routing and filtering
- Manifest parsing and validation
- Resource limit enforcement

### Integration Testing

- End-to-end plugin lifecycle management
- Hot-reload functionality under load
- Multi-plugin concurrent execution
- Event delivery and UI rendering pipeline
- Security boundary enforcement

### Security Testing

- Sandbox escape attempt prevention
- Resource exhaustion protection
- Permission bypass testing
- Malicious plugin behavior isolation
- API security validation

### Performance Testing

- Plugin execution benchmarks
- Event processing throughput
- Memory usage under load
- UI responsiveness with multiple plugins
- Hot-reload performance impact

## Deployment Plan

### Migration Strategy

#### Phase 1: Core Infrastructure (Weeks 1-2)
- Implement WASM runtime with basic sandboxing
- Create plugin manifest parser and validator
- Build permission system with audit logging
- Develop hot-reload mechanism
- Set up monitoring and metrics collection

#### Phase 2: API Development (Weeks 3-4)
- Implement host API for event subscription
- Create admin API with capability gating
- Build plugin lifecycle management
- Add resource monitoring and limits
- Develop plugin storage and configuration system

#### Phase 3: Integration (Weeks 5-6)
- Integrate with existing TUI framework
- Implement panel rendering system
- Create event bus integration
- Add plugin discovery and loading
- Build administrative interface for plugin management

#### Phase 4: Sample Plugins (Weeks 7-8)
- Develop "Tenant SLA Monitor" sample plugin
- Create "Bulk Operations" administrative plugin
- Build plugin development toolkit and templates
- Write comprehensive documentation and tutorials
- Conduct security review and testing

### Rollout Plan

**Alpha Release** (Internal Testing):
- Basic plugin loading and execution
- Simple event subscription
- Read-only sample plugins
- Internal dogfooding with platform team

**Beta Release** (Limited External):
- Full API surface with permission system
- Administrative action capabilities
- Hot-reload and plugin management
- Selected enterprise customers

**Production Release** (General Availability):
- Complete feature set with security hardening
- Plugin marketplace integration
- Comprehensive documentation
- Full support and SLA commitments

---

This design document establishes the foundation for implementing the Plugin Panel System as a strategic platform feature that transforms the go-redis-work-queue from a static tool into an extensible ecosystem. The focus on security, performance, and developer experience ensures that the system can support enterprise adoption while enabling innovation through third-party development.