// Copyright 2025 James Ross
package pluginpanel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// StarlarkRuntime implements RuntimeEngine for Starlark scripts
type StarlarkRuntime struct {
	logger *zap.Logger
}

// NewStarlarkRuntime creates a new Starlark runtime engine
func NewStarlarkRuntime(logger *zap.Logger) *StarlarkRuntime {
	return &StarlarkRuntime{
		logger: logger,
	}
}

// LoadPlugin loads a Starlark plugin
func (r *StarlarkRuntime) LoadPlugin(ctx context.Context, manifest *PluginManifest, code []byte) (PluginInstance, error) {
	instance := &StarlarkInstance{
		id:       manifest.Name,
		runtime:  r,
		manifest: manifest,
		code:     string(code),
		logger:   r.logger,
		status:   StatusUnloaded,
		events:   make(chan *Event, 100),
		methods:  make(map[string]starlark.Value),
	}

	// Validate and load the Starlark code
	if err := instance.load(ctx); err != nil {
		return nil, err
	}

	instance.status = StatusReady
	return instance, nil
}

// SupportedRuntime returns the runtime type
func (r *StarlarkRuntime) SupportedRuntime() Runtime {
	return RuntimeStarlark
}

// ValidateCode validates Starlark code syntax (simplified implementation)
func (r *StarlarkRuntime) ValidateCode(code []byte) error {
	// Simplified validation - just check if it's not empty
	if len(code) == 0 {
		return fmt.Errorf("empty code")
	}
	// In a real implementation, this would use go.starlark.net to parse and validate
	return nil
}

// GetCapabilities returns the capabilities this runtime supports
func (r *StarlarkRuntime) GetCapabilities() []Capability {
	return []Capability{
		CapabilityReadStats,
		CapabilityReadKeys,
		CapabilityReadSelection,
		CapabilityReadTimers,
		CapabilityReadQueues,
		CapabilityReadJobs,
		CapabilityRenderPanel,
		CapabilityKeyEvents,
		CapabilityMouseEvents,
	}
}

// StarlarkInstance represents a running Starlark plugin
type StarlarkInstance struct {
	id       string
	runtime  *StarlarkRuntime
	manifest *PluginManifest
	code     string
	logger   *zap.Logger
	status   PluginStatus
	events   chan *Event
	methods  map[string]starlark.Value
	globals  starlark.StringDict
	thread   *starlark.Thread
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
	metrics  map[string]interface{}
}

// ID returns the plugin ID
func (i *StarlarkInstance) ID() string {
	return i.id
}

// Status returns the current status
func (i *StarlarkInstance) Status() PluginStatus {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.status
}

// Start starts the plugin instance
func (i *StarlarkInstance) Start(ctx context.Context) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.status == StatusRunning {
		return nil
	}

	i.ctx, i.cancel = context.WithCancel(ctx)
	i.status = StatusRunning

	// Start event processing goroutine
	go i.processEvents()

	// Call plugin's start method if it exists
	if startFn, exists := i.methods["start"]; exists {
		thread := &starlark.Thread{Name: "start"}
		_, err := starlark.Call(thread, startFn, nil, nil)
		if err != nil {
			i.status = StatusError
			return fmt.Errorf("plugin start failed: %w", err)
		}
	}

	i.logger.Debug("Starlark plugin started", zap.String("plugin_id", i.id))
	return nil
}

// Stop stops the plugin instance
func (i *StarlarkInstance) Stop(ctx context.Context) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.status != StatusRunning {
		return nil
	}

	// Call plugin's stop method if it exists
	if stopFn, exists := i.methods["stop"]; exists {
		thread := &starlark.Thread{Name: "stop"}
		starlark.Call(thread, stopFn, nil, nil)
	}

	if i.cancel != nil {
		i.cancel()
	}

	i.status = StatusTerminated
	i.logger.Debug("Starlark plugin stopped", zap.String("plugin_id", i.id))
	return nil
}

// Call invokes a plugin method
func (i *StarlarkInstance) Call(ctx context.Context, method string, args interface{}) (interface{}, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if i.status != StatusRunning {
		return nil, fmt.Errorf("plugin not running")
	}

	fn, exists := i.methods[method]
	if !exists {
		return nil, fmt.Errorf("method not found: %s", method)
	}

	thread := &starlark.Thread{Name: method}

	// Convert args to Starlark values
	var starlarkArgs []starlark.Value
	if args != nil {
		starlarkValue, err := i.goToStarlark(args)
		if err != nil {
			return nil, fmt.Errorf("failed to convert args: %w", err)
		}
		starlarkArgs = []starlark.Value{starlarkValue}
	}

	result, err := starlark.Call(thread, fn, starlarkArgs, nil)
	if err != nil {
		return nil, fmt.Errorf("plugin method call failed: %w", err)
	}

	// Convert result back to Go
	return i.starlarkToGo(result)
}

// SendEvent sends an event to the plugin
func (i *StarlarkInstance) SendEvent(ctx context.Context, event *Event) error {
	select {
	case i.events <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("event queue full")
	}
}

// GetMetrics returns plugin metrics
func (i *StarlarkInstance) GetMetrics() map[string]interface{} {
	i.mu.RLock()
	defer i.mu.RUnlock()

	metrics := make(map[string]interface{})
	for k, v := range i.metrics {
		metrics[k] = v
	}
	metrics["status"] = string(i.status)
	metrics["last_heartbeat"] = time.Now()
	return metrics
}

// Cleanup performs cleanup operations
func (i *StarlarkInstance) Cleanup() error {
	if i.cancel != nil {
		i.cancel()
	}
	close(i.events)
	return nil
}

// load initializes the Starlark environment and loads the plugin code
func (i *StarlarkInstance) load(ctx context.Context) error {
	i.thread = &starlark.Thread{Name: "main"}
	i.metrics = make(map[string]interface{})

	// Create built-in functions for the plugin
	builtins := starlark.StringDict{
		"log":         starlark.NewBuiltin("log", i.builtinLog),
		"render":      starlark.NewBuiltin("render", i.builtinRender),
		"subscribe":   starlark.NewBuiltin("subscribe", i.builtinSubscribe),
		"get_stats":   starlark.NewBuiltin("get_stats", i.builtinGetStats),
		"enqueue":     starlark.NewBuiltin("enqueue", i.builtinEnqueue),
		"struct":      starlarkstruct.Module,
	}

	// Execute the plugin code
	globals, err := starlark.ExecFile(i.thread, i.id+".star", i.code, builtins)
	if err != nil {
		return fmt.Errorf("failed to execute plugin code: %w", err)
	}

	i.globals = globals

	// Extract plugin methods
	for name, value := range globals {
		if fn, ok := value.(*starlark.Function); ok {
			i.methods[name] = fn
		}
	}

	// Validate required methods
	requiredMethods := []string{"on_event", "render"}
	for _, method := range requiredMethods {
		if _, exists := i.methods[method]; !exists {
			return fmt.Errorf("required method not found: %s", method)
		}
	}

	return nil
}

// processEvents handles incoming events
func (i *StarlarkInstance) processEvents() {
	for {
		select {
		case event := <-i.events:
			i.handleEvent(event)
		case <-i.ctx.Done():
			return
		}
	}
}

// handleEvent processes a single event
func (i *StarlarkInstance) handleEvent(event *Event) {
	i.mu.RLock()
	onEventFn, exists := i.methods["on_event"]
	i.mu.RUnlock()

	if !exists {
		return
	}

	thread := &starlark.Thread{Name: "on_event"}

	// Convert event to Starlark dict
	eventDict := starlark.NewDict(len(event.Data) + 3)
	eventDict.SetKey(starlark.String("type"), starlark.String(string(event.Type)))
	eventDict.SetKey(starlark.String("timestamp"), starlark.String(event.Timestamp.Format(time.RFC3339)))
	eventDict.SetKey(starlark.String("source"), starlark.String(event.Source))

	for key, value := range event.Data {
		starlarkValue, err := i.goToStarlark(value)
		if err != nil {
			i.logger.Warn("Failed to convert event data",
				zap.String("plugin_id", i.id),
				zap.String("key", key),
				zap.Error(err))
			continue
		}
		eventDict.SetKey(starlark.String(key), starlarkValue)
	}

	_, err := starlark.Call(thread, onEventFn, []starlark.Value{eventDict}, nil)
	if err != nil {
		i.logger.Warn("Plugin event handler failed",
			zap.String("plugin_id", i.id),
			zap.String("event_type", string(event.Type)),
			zap.Error(err))
	}
}

// Built-in functions for plugins

func (i *StarlarkInstance) builtinLog(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var level, message string
	if err := starlark.UnpackArgs("log", args, kwargs, "level", &level, "message", &message); err != nil {
		return nil, err
	}

	switch level {
	case "debug":
		i.logger.Debug(message, zap.String("plugin_id", i.id))
	case "info":
		i.logger.Info(message, zap.String("plugin_id", i.id))
	case "warn":
		i.logger.Warn(message, zap.String("plugin_id", i.id))
	case "error":
		i.logger.Error(message, zap.String("plugin_id", i.id))
	default:
		i.logger.Info(message, zap.String("plugin_id", i.id))
	}

	return nil, nil
}

func (i *StarlarkInstance) builtinRender(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var text string
	if err := starlark.UnpackArgs("render", args, kwargs, "text", &text); err != nil {
		return nil, err
	}

	// This would integrate with the panel renderer
	i.logger.Debug("Plugin render request",
		zap.String("plugin_id", i.id),
		zap.String("text", text))

	return nil, nil
}

func (i *StarlarkInstance) builtinSubscribe(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var eventType string
	if err := starlark.UnpackArgs("subscribe", args, kwargs, "event_type", &eventType); err != nil {
		return nil, err
	}

	i.logger.Debug("Plugin subscribing to event",
		zap.String("plugin_id", i.id),
		zap.String("event_type", eventType))

	return nil, nil
}

func (i *StarlarkInstance) builtinGetStats(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	// This would integrate with the stats system
	stats := starlark.NewDict(2)
	stats.SetKey(starlark.String("queues"), starlark.MakeInt(5))
	stats.SetKey(starlark.String("jobs"), starlark.MakeInt(150))

	return stats, nil
}

func (i *StarlarkInstance) builtinEnqueue(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var queue string
	var payload starlark.Value
	if err := starlark.UnpackArgs("enqueue", args, kwargs, "queue", &queue, "payload", &payload); err != nil {
		return nil, err
	}

	// This would require capability checking and integration with the job system
	i.logger.Debug("Plugin enqueue request",
		zap.String("plugin_id", i.id),
		zap.String("queue", queue))

	return nil, nil
}

// Helper methods for type conversion

func (i *StarlarkInstance) goToStarlark(value interface{}) (starlark.Value, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case bool:
		return starlark.Bool(v), nil
	case int:
		return starlark.MakeInt(v), nil
	case int64:
		return starlark.MakeInt64(v), nil
	case float64:
		return starlark.Float(v), nil
	case string:
		return starlark.String(v), nil
	case []interface{}:
		list := starlark.NewList(make([]starlark.Value, len(v)))
		for i, item := range v {
			starlarkItem, err := i.goToStarlark(item)
			if err != nil {
				return nil, err
			}
			list.SetIndex(i, starlarkItem)
		}
		return list, nil
	case map[string]interface{}:
		dict := starlark.NewDict(len(v))
		for key, value := range v {
			starlarkValue, err := i.goToStarlark(value)
			if err != nil {
				return nil, err
			}
			dict.SetKey(starlark.String(key), starlarkValue)
		}
		return dict, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", v)
	}
}

func (i *StarlarkInstance) starlarkToGo(value starlark.Value) (interface{}, error) {
	switch v := value.(type) {
	case starlark.NoneType:
		return nil, nil
	case starlark.Bool:
		return bool(v), nil
	case starlark.Int:
		val, ok := v.Int64()
		if !ok {
			return nil, fmt.Errorf("integer too large")
		}
		return val, nil
	case starlark.Float:
		return float64(v), nil
	case starlark.String:
		return string(v), nil
	case *starlark.List:
		result := make([]interface{}, v.Len())
		for i := 0; i < v.Len(); i++ {
			item, err := i.starlarkToGo(v.Index(i))
			if err != nil {
				return nil, err
			}
			result[i] = item
		}
		return result, nil
	case *starlark.Dict:
		result := make(map[string]interface{})
		for _, item := range v.Items() {
			key, ok := item[0].(starlark.String)
			if !ok {
				continue
			}
			value, err := i.starlarkToGo(item[1])
			if err != nil {
				return nil, err
			}
			result[string(key)] = value
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported Starlark type: %T", v)
	}
}