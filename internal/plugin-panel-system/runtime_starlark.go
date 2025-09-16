// Copyright 2025 James Ross
package pluginpanel

import (
	"context"
	"fmt"
	"strings"
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
		methods:  make(map[string]interface{}),
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
	methods  map[string]interface{}
	globals  map[string]interface{}
	scope    map[string]interface{}
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
	if _, exists := i.methods["start"]; exists {
		// Simulate calling start function
		_, err := i.callMethod("start", nil)
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
	if _, exists := i.methods["stop"]; exists {
		// Simulate calling stop function
		i.callMethod("stop", nil)
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

	_, exists := i.methods[method]
	if !exists {
		return nil, fmt.Errorf("method not found: %s", method)
	}

	// Simulate method call (simplified)
	result, err := i.callMethod(method, args)
	if err != nil {
		return nil, fmt.Errorf("plugin method call failed: %w", err)
	}

	return result, nil
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
	// Initialize simplified environment
	i.metrics = make(map[string]interface{})

	// Create built-in functions for the plugin
	// Create built-in functions for the plugin
	i.scope = map[string]interface{}{
		"log":         i.builtinLog,
		"render":      i.builtinRender,
		"subscribe":   i.builtinSubscribe,
		"get_stats":   i.builtinGetStats,
		"enqueue":     i.builtinEnqueue,
		"get_jobs":    i.builtinGetJobs,
		"get_queues":  i.builtinGetQueues,
		"requeue_job": i.builtinRequeueJob,
		"show_dialog": i.builtinShowDialog,
	}

	// Parse and validate the plugin code (simplified)
	if err := i.parseCode(); err != nil {
		return fmt.Errorf("failed to parse plugin code: %w", err)
	}

	// Extract plugin methods (simplified detection)
	i.extractMethods()

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
	_, exists := i.methods["on_event"]
	i.mu.RUnlock()

	if !exists {
		return
	}

	// Convert event to Go map
	eventMap := map[string]interface{}{
		"type":      string(event.Type),
		"timestamp": event.Timestamp.Format(time.RFC3339),
		"source":    event.Source,
	}

	for key, value := range event.Data {
		eventMap[key] = value
	}

	// Call the event handler
	_, err := i.callMethod("on_event", eventMap)
	if err != nil {
		i.logger.Warn("Plugin event handler failed",
			zap.String("plugin_id", i.id),
			zap.String("event_type", string(event.Type)),
			zap.Error(err))
	}
}

// Built-in functions for plugins

func (i *StarlarkInstance) builtinLog(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("log requires 2 arguments: level and message")
	}

	level, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("log level must be string")
	}

	message, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("log message must be string")
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

func (i *StarlarkInstance) builtinRender(args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("render requires 1 argument: text")
	}

	text, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("render text must be string")
	}

	// This would integrate with the panel renderer
	i.logger.Debug("Plugin render request",
		zap.String("plugin_id", i.id),
		zap.String("text", text))

	return nil, nil
}

func (i *StarlarkInstance) builtinSubscribe(args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("subscribe requires 1 argument: event_type")
	}

	eventType, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("event type must be string")
	}

	i.logger.Debug("Plugin subscribing to event",
		zap.String("plugin_id", i.id),
		zap.String("event_type", eventType))

	return nil, nil
}

func (i *StarlarkInstance) builtinGetStats(args ...interface{}) (interface{}, error) {
	// This would integrate with the stats system
	stats := map[string]interface{}{
		"queues": 5,
		"jobs":   150,
	}

	return stats, nil
}

func (i *StarlarkInstance) builtinEnqueue(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("enqueue requires 2 arguments: queue and payload")
	}

	queue, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("queue must be string")
	}

	_ = args[1] // payload

	// This would require capability checking and integration with the job system
	i.logger.Debug("Plugin enqueue request",
		zap.String("plugin_id", i.id),
		zap.String("queue", queue))

	return nil, nil
}

// Additional built-in functions for plugins

func (i *StarlarkInstance) builtinGetJobs(args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("get_jobs requires 1 argument: queue")
	}

	queue, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("queue must be string")
	}

	limit := 50
	if len(args) > 1 {
		if l, ok := args[1].(int); ok {
			limit = l
		}
	}

	// Mock job data
	jobs := make([]map[string]interface{}, 0, limit)
	for i := 0; i < 3; i++ {
		jobs = append(jobs, map[string]interface{}{
			"id":     fmt.Sprintf("job_%s_%d", queue, i),
			"status": "failed",
			"error":  "Connection timeout",
		})
	}

	i.logger.Debug("Plugin get_jobs request",
		zap.String("plugin_id", i.id),
		zap.String("queue", queue),
		zap.Int("limit", limit))

	return jobs, nil
}

func (i *StarlarkInstance) builtinGetQueues(args ...interface{}) (interface{}, error) {
	// Mock queue data
	queues := []string{"default", "high_priority", "background"}

	i.logger.Debug("Plugin get_queues request",
		zap.String("plugin_id", i.id))

	return queues, nil
}

func (i *StarlarkInstance) builtinRequeueJob(args ...interface{}) (interface{}, error) {
	if len(args) < 1 {
		return nil, fmt.Errorf("requeue_job requires 1 argument: job_id")
	}

	jobID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("job_id must be string")
	}

	i.logger.Debug("Plugin requeue_job request",
		zap.String("plugin_id", i.id),
		zap.String("job_id", jobID))

	return nil, nil
}

func (i *StarlarkInstance) builtinShowDialog(args ...interface{}) (interface{}, error) {
	if len(args) < 2 {
		return nil, fmt.Errorf("show_dialog requires 2 arguments: title and message")
	}

	title, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("title must be string")
	}

	message, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("message must be string")
	}

	i.logger.Debug("Plugin show_dialog request",
		zap.String("plugin_id", i.id),
		zap.String("title", title),
		zap.String("message", message))

	return nil, nil
}

// Helper methods for simplified code parsing

func (i *StarlarkInstance) parseCode() error {
	// Simplified code validation - check for function definitions
	lines := strings.Split(i.code, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "def ") && strings.Contains(line, ":") {
			// Extract function name
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				funcName := strings.TrimSuffix(parts[1], "(")
				if strings.Contains(funcName, "(") {
					funcName = strings.Split(funcName, "(")[0]
				}
				i.methods[funcName] = true
			}
		}
	}
	return nil
}

func (i *StarlarkInstance) extractMethods() {
	// Methods are extracted during parseCode
	i.logger.Debug("Extracted plugin methods",
		zap.String("plugin_id", i.id),
		zap.Int("method_count", len(i.methods)))
}

func (i *StarlarkInstance) callMethod(method string, args interface{}) (interface{}, error) {
	// Simplified method calling - just log and return mock data
	i.logger.Debug("Plugin method call",
		zap.String("plugin_id", i.id),
		zap.String("method", method))

	switch method {
	case "render":
		return "Plugin rendered content", nil
	case "on_event":
		return nil, nil
	default:
		return fmt.Sprintf("Method %s called", method), nil
	}
}