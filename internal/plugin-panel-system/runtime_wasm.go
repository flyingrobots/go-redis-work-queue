// Copyright 2025 James Ross
package pluginpanel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wasmerio/wasmer-go/wasmer"
	"go.uber.org/zap"
)

// WASMRuntime implements RuntimeEngine for WASM plugins
type WASMRuntime struct {
	logger *zap.Logger
	engine *wasmer.Engine
	store  *wasmer.Store
}

// NewWASMRuntime creates a new WASM runtime engine
func NewWASMRuntime(logger *zap.Logger) *WASMRuntime {
	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	return &WASMRuntime{
		logger: logger,
		engine: engine,
		store:  store,
	}
}

// LoadPlugin loads a WASM plugin
func (r *WASMRuntime) LoadPlugin(ctx context.Context, manifest *PluginManifest, code []byte) (PluginInstance, error) {
	// Compile WASM module
	module, err := wasmer.NewModule(r.store, code)
	if err != nil {
		return nil, fmt.Errorf("failed to compile WASM module: %w", err)
	}

	// Create imports for the plugin to use
	imports, err := r.createImports()
	if err != nil {
		return nil, fmt.Errorf("failed to create imports: %w", err)
	}

	// Instantiate the module
	instance, err := wasmer.NewInstance(module, imports)
	if err != nil {
		return nil, fmt.Errorf("failed to instantiate WASM module: %w", err)
	}

	wasmInstance := &WASMInstance{
		id:       manifest.Name,
		runtime:  r,
		manifest: manifest,
		module:   module,
		instance: instance,
		logger:   r.logger,
		status:   StatusReady,
		events:   make(chan *Event, 100),
		metrics:  make(map[string]interface{}),
	}

	// Initialize the plugin
	if err := wasmInstance.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize WASM plugin: %w", err)
	}

	return wasmInstance, nil
}

// SupportedRuntime returns the runtime type
func (r *WASMRuntime) SupportedRuntime() Runtime {
	return RuntimeWASM
}

// ValidateCode validates WASM bytecode
func (r *WASMRuntime) ValidateCode(code []byte) error {
	_, err := wasmer.NewModule(r.store, code)
	return err
}

// GetCapabilities returns the capabilities this runtime supports
func (r *WASMRuntime) GetCapabilities() []Capability {
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
		CapabilityEnqueue,
		CapabilityPeek,
		CapabilityRequeue,
	}
}

// createImports creates the host functions available to WASM plugins
func (r *WASMRuntime) createImports() (*wasmer.ImportObject, error) {
	imports := wasmer.NewImportObject()

	// Add host functions that plugins can call
	logFunc := wasmer.NewFunction(
		r.store,
		wasmer.NewFunctionType(
			wasmer.NewValueTypes(wasmer.I32, wasmer.I32, wasmer.I32, wasmer.I32),
			wasmer.NewValueTypes(),
		),
		func(args []wasmer.Value) ([]wasmer.Value, error) {
			// Implementation for log function
			// args[0] = log level, args[1] = message pointer, args[2] = message length
			r.logger.Info("WASM plugin log call")
			return []wasmer.Value{}, nil
		},
	)

	renderFunc := wasmer.NewFunction(
		r.store,
		wasmer.NewFunctionType(
			wasmer.NewValueTypes(wasmer.I32, wasmer.I32),
			wasmer.NewValueTypes(),
		),
		func(args []wasmer.Value) ([]wasmer.Value, error) {
			// Implementation for render function
			r.logger.Info("WASM plugin render call")
			return []wasmer.Value{}, nil
		},
	)

	imports.Register("env", map[string]wasmer.IntoExtern{
		"log":    logFunc,
		"render": renderFunc,
	})

	return imports, nil
}

// WASMInstance represents a running WASM plugin
type WASMInstance struct {
	id       string
	runtime  *WASMRuntime
	manifest *PluginManifest
	module   *wasmer.Module
	instance *wasmer.Instance
	logger   *zap.Logger
	status   PluginStatus
	events   chan *Event
	metrics  map[string]interface{}
	mu       sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// ID returns the plugin ID
func (i *WASMInstance) ID() string {
	return i.id
}

// Status returns the current status
func (i *WASMInstance) Status() PluginStatus {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.status
}

// Start starts the plugin instance
func (i *WASMInstance) Start(ctx context.Context) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.status == StatusRunning {
		return nil
	}

	i.ctx, i.cancel = context.WithCancel(ctx)
	i.status = StatusRunning

	// Start event processing
	go i.processEvents()

	// Call plugin's start function if it exists
	startFunc, err := i.instance.Exports.GetFunction("start")
	if err == nil {
		_, err = startFunc()
		if err != nil {
			i.status = StatusError
			return fmt.Errorf("WASM plugin start failed: %w", err)
		}
	}

	i.logger.Debug("WASM plugin started", zap.String("plugin_id", i.id))
	return nil
}

// Stop stops the plugin instance
func (i *WASMInstance) Stop(ctx context.Context) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.status != StatusRunning {
		return nil
	}

	// Call plugin's stop function if it exists
	stopFunc, err := i.instance.Exports.GetFunction("stop")
	if err == nil {
		stopFunc()
	}

	if i.cancel != nil {
		i.cancel()
	}

	i.status = StatusTerminated
	i.logger.Debug("WASM plugin stopped", zap.String("plugin_id", i.id))
	return nil
}

// Call invokes a plugin function
func (i *WASMInstance) Call(ctx context.Context, method string, args interface{}) (interface{}, error) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	if i.status != StatusRunning {
		return nil, fmt.Errorf("plugin not running")
	}

	// Get the exported function
	fn, err := i.instance.Exports.GetFunction(method)
	if err != nil {
		return nil, fmt.Errorf("function not found: %s", method)
	}

	// For simplicity, assume no arguments for now
	result, err := fn()
	if err != nil {
		return nil, fmt.Errorf("WASM function call failed: %w", err)
	}

	return result, nil
}

// SendEvent sends an event to the plugin
func (i *WASMInstance) SendEvent(ctx context.Context, event *Event) error {
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
func (i *WASMInstance) GetMetrics() map[string]interface{} {
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
func (i *WASMInstance) Cleanup() error {
	if i.cancel != nil {
		i.cancel()
	}
	close(i.events)
	return nil
}

// initialize sets up the WASM plugin
func (i *WASMInstance) initialize() error {
	// Check for required exports
	requiredExports := []string{"on_event", "render"}
	for _, export := range requiredExports {
		if _, err := i.instance.Exports.GetFunction(export); err != nil {
			return fmt.Errorf("required export not found: %s", export)
		}
	}

	// Call initialization function if it exists
	initFunc, err := i.instance.Exports.GetFunction("init")
	if err == nil {
		_, err = initFunc()
		if err != nil {
			return fmt.Errorf("plugin initialization failed: %w", err)
		}
	}

	return nil
}

// processEvents handles incoming events
func (i *WASMInstance) processEvents() {
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
func (i *WASMInstance) handleEvent(event *Event) {
	i.mu.RLock()
	defer i.mu.RUnlock()

	// Get the event handler function
	onEventFunc, err := i.instance.Exports.GetFunction("on_event")
	if err != nil {
		return
	}

	// For now, just call the function without arguments
	// In a real implementation, you'd serialize the event data
	// and pass it to the WASM function
	_, err = onEventFunc()
	if err != nil {
		i.logger.Warn("WASM plugin event handler failed",
			zap.String("plugin_id", i.id),
			zap.String("event_type", string(event.Type)),
			zap.Error(err))
	}
}

// Memory and data exchange helpers

func (i *WASMInstance) writeMemory(data []byte, offset int32) error {
	memory, err := i.instance.Exports.GetMemory("memory")
	if err != nil {
		return err
	}

	memoryData := memory.Data()
	if int(offset)+len(data) > len(memoryData) {
		return fmt.Errorf("memory write out of bounds")
	}

	copy(memoryData[offset:], data)
	return nil
}

func (i *WASMInstance) readMemory(offset, length int32) ([]byte, error) {
	memory, err := i.instance.Exports.GetMemory("memory")
	if err != nil {
		return nil, err
	}

	memoryData := memory.Data()
	if int(offset)+int(length) > len(memoryData) {
		return nil, fmt.Errorf("memory read out of bounds")
	}

	data := make([]byte, length)
	copy(data, memoryData[offset:offset+length])
	return data, nil
}