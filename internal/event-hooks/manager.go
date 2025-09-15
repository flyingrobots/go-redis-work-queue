// Copyright 2025 James Ross
package eventhooks

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

// Manager coordinates all event hooks components
type Manager struct {
	config           EventBusConfig
	eventBus         *EventBus
	configManager    *ConfigManager
	webhookDeliverer *WebhookDeliverer
	natsDeliverer    *NATSDeliverer
	apiService       *EventHooksService
	redis            *redis.Client
	logger           *slog.Logger

	// State management
	mu        sync.RWMutex
	isRunning bool
}

// NewManager creates a new event hooks manager
func NewManager(config EventBusConfig, redisClient *redis.Client, logger *slog.Logger) *Manager {
	configManager := NewConfigManager(redisClient, logger)
	webhookDeliverer := NewWebhookDeliverer(logger)
	natsDeliverer := NewNATSDeliverer("", logger) // NATS URL would be configured
	eventBus := NewEventBus(config, redisClient, logger)

	apiService := NewEventHooksService(
		configManager,
		eventBus,
		webhookDeliverer,
		natsDeliverer,
		logger,
	)

	return &Manager{
		config:           config,
		eventBus:         eventBus,
		configManager:    configManager,
		webhookDeliverer: webhookDeliverer,
		natsDeliverer:    natsDeliverer,
		apiService:       apiService,
		redis:            redisClient,
		logger:           logger,
		isRunning:        false,
	}
}

// Start initializes and starts all event hooks components
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return fmt.Errorf("event hooks manager is already running")
	}

	m.logger.Info("starting event hooks manager")

	// Start event bus
	if err := m.eventBus.Start(); err != nil {
		return fmt.Errorf("failed to start event bus: %w", err)
	}

	// Load existing webhook subscriptions and resubscribe
	if err := m.loadExistingSubscriptions(ctx); err != nil {
		m.logger.Warn("failed to load existing subscriptions", "error", err)
		// Continue anyway - not critical
	}

	m.isRunning = true
	m.logger.Info("event hooks manager started successfully")
	return nil
}

// Stop gracefully shuts down all event hooks components
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning {
		return fmt.Errorf("event hooks manager is not running")
	}

	m.logger.Info("stopping event hooks manager")

	// Stop event bus
	if err := m.eventBus.Stop(); err != nil {
		m.logger.Warn("failed to stop event bus", "error", err)
	}

	// Close webhook deliverer
	webhookSubs := m.webhookDeliverer.ListSubscribers()
	for id := range webhookSubs {
		if err := m.webhookDeliverer.RemoveSubscription(id); err != nil {
			m.logger.Warn("failed to remove webhook subscription",
				"subscription_id", id, "error", err)
		}
	}

	// Close NATS deliverer
	if err := m.natsDeliverer.Close(); err != nil {
		m.logger.Warn("failed to close NATS deliverer", "error", err)
	}

	m.isRunning = false
	m.logger.Info("event hooks manager stopped")
	return nil
}

// EmitEvent emits an event to all subscribers
func (m *Manager) EmitEvent(event JobEvent) error {
	if !m.isRunning {
		return ErrEventBusShutdown
	}

	return m.eventBus.Emit(event)
}

// RegisterAPIRoutes registers HTTP API routes
func (m *Manager) RegisterAPIRoutes(router *mux.Router) {
	m.apiService.RegisterRoutes(router)
}

// GetMetrics returns current metrics
func (m *Manager) GetMetrics() EventMetrics {
	return m.eventBus.GetMetrics()
}

// IsRunning returns whether the manager is currently running
func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.isRunning
}

// loadExistingSubscriptions reloads webhook subscriptions from storage
func (m *Manager) loadExistingSubscriptions(ctx context.Context) error {
	m.logger.Info("loading existing webhook subscriptions")

	subscriptions, err := m.configManager.ListWebhookSubscriptions(ctx)
	if err != nil {
		return fmt.Errorf("failed to list webhook subscriptions: %w", err)
	}

	loaded := 0
	for _, subscription := range subscriptions {
		if subscription.Disabled {
			m.logger.Debug("skipping disabled subscription",
				"subscription_id", subscription.ID,
				"name", subscription.Name)
			continue
		}

		// Add to webhook deliverer
		webhookSub := m.webhookDeliverer.AddSubscription(subscription)

		// Subscribe to event bus
		if err := m.eventBus.Subscribe(webhookSub); err != nil {
			m.logger.Warn("failed to subscribe to event bus",
				"subscription_id", subscription.ID,
				"name", subscription.Name,
				"error", err)
			continue
		}

		loaded++
		m.logger.Debug("loaded webhook subscription",
			"subscription_id", subscription.ID,
			"name", subscription.Name,
			"url", subscription.URL)
	}

	m.logger.Info("loaded webhook subscriptions",
		"total", len(subscriptions),
		"loaded", loaded,
		"skipped", len(subscriptions)-loaded)

	return nil
}

// CreateWebhookSubscription creates a new webhook subscription
func (m *Manager) CreateWebhookSubscription(ctx context.Context, req CreateWebhookRequest) (*WebhookSubscription, error) {
	if !m.isRunning {
		return nil, ErrEventBusShutdown
	}

	// Create subscription
	subscription, err := m.configManager.CreateWebhookSubscription(ctx, req)
	if err != nil {
		return nil, err
	}

	// Add to deliverer and event bus
	webhookSub := m.webhookDeliverer.AddSubscription(subscription)
	if err := m.eventBus.Subscribe(webhookSub); err != nil {
		// Clean up on failure
		m.webhookDeliverer.RemoveSubscription(subscription.ID)
		m.configManager.DeleteWebhookSubscription(ctx, subscription.ID)
		return nil, fmt.Errorf("failed to subscribe to event bus: %w", err)
	}

	return subscription, nil
}

// UpdateWebhookSubscription updates an existing webhook subscription
func (m *Manager) UpdateWebhookSubscription(ctx context.Context, id string, req UpdateWebhookRequest) (*WebhookSubscription, error) {
	if !m.isRunning {
		return nil, ErrEventBusShutdown
	}

	subscription, err := m.configManager.UpdateWebhookSubscription(ctx, id, req)
	if err != nil {
		return nil, err
	}

	// Update webhook deliverer
	if err := m.webhookDeliverer.UpdateSubscription(subscription); err != nil {
		m.logger.Warn("failed to update webhook deliverer",
			"subscription_id", id, "error", err)
	}

	return subscription, nil
}

// DeleteWebhookSubscription removes a webhook subscription
func (m *Manager) DeleteWebhookSubscription(ctx context.Context, id string) error {
	if !m.isRunning {
		return ErrEventBusShutdown
	}

	// Remove from deliverer and event bus
	m.webhookDeliverer.RemoveSubscription(id)
	m.eventBus.Unsubscribe(id)

	// Delete from storage
	return m.configManager.DeleteWebhookSubscription(ctx, id)
}

// GetWebhookSubscription retrieves a webhook subscription
func (m *Manager) GetWebhookSubscription(ctx context.Context, id string) (*WebhookSubscription, error) {
	return m.configManager.GetWebhookSubscription(ctx, id)
}

// ListWebhookSubscriptions lists all webhook subscriptions
func (m *Manager) ListWebhookSubscriptions(ctx context.Context) ([]*WebhookSubscription, error) {
	return m.configManager.ListWebhookSubscriptions(ctx)
}

// TestWebhookDelivery tests a webhook subscription
func (m *Manager) TestWebhookDelivery(id string) error {
	if !m.isRunning {
		return ErrEventBusShutdown
	}

	webhookSub, err := m.webhookDeliverer.GetSubscriber(id)
	if err != nil {
		return err
	}

	return webhookSub.TestDelivery()
}

// GetDeadLetterHooks retrieves dead letter hooks for a subscription
func (m *Manager) GetDeadLetterHooks(subscriptionID string, limit int) ([]*DeadLetterHook, error) {
	if !m.isRunning {
		return nil, ErrEventBusShutdown
	}

	return m.eventBus.GetDLHEntries(subscriptionID, limit)
}

// GetSubscriptionHealthStatuses returns health status for all subscriptions
func (m *Manager) GetSubscriptionHealthStatuses() []SubscriptionHealthStatus {
	if !m.isRunning {
		return []SubscriptionHealthStatus{}
	}

	return m.webhookDeliverer.GetHealthStatuses()
}

// EmitJobEvent is a convenience method for emitting job lifecycle events
func (m *Manager) EmitJobEvent(eventType EventType, jobID, queue string, priority, attempt int) error {
	event := JobEvent{
		Event:     eventType,
		JobID:     jobID,
		Queue:     queue,
		Priority:  priority,
		Attempt:   attempt,
	}

	return m.EmitEvent(event)
}

// EmitJobEnqueued emits a job enqueued event
func (m *Manager) EmitJobEnqueued(jobID, queue string, priority int, payload interface{}) error {
	event := JobEvent{
		Event:    EventJobEnqueued,
		JobID:    jobID,
		Queue:    queue,
		Priority: priority,
		Attempt:  1,
		Payload:  payload,
	}

	return m.EmitEvent(event)
}

// EmitJobStarted emits a job started event
func (m *Manager) EmitJobStarted(jobID, queue, workerID string, priority, attempt int) error {
	event := JobEvent{
		Event:    EventJobStarted,
		JobID:    jobID,
		Queue:    queue,
		Priority: priority,
		Attempt:  attempt,
		Worker:   workerID,
	}

	return m.EmitEvent(event)
}

// EmitJobSucceeded emits a job succeeded event
func (m *Manager) EmitJobSucceeded(jobID, queue, workerID string, priority, attempt int, duration *time.Duration) error {
	event := JobEvent{
		Event:    EventJobSucceeded,
		JobID:    jobID,
		Queue:    queue,
		Priority: priority,
		Attempt:  attempt,
		Worker:   workerID,
		Duration: duration,
	}

	return m.EmitEvent(event)
}

// EmitJobFailed emits a job failed event
func (m *Manager) EmitJobFailed(jobID, queue, workerID string, priority, attempt int, errorMsg string, duration *time.Duration) error {
	event := JobEvent{
		Event:    EventJobFailed,
		JobID:    jobID,
		Queue:    queue,
		Priority: priority,
		Attempt:  attempt,
		Worker:   workerID,
		Error:    errorMsg,
		Duration: duration,
	}

	return m.EmitEvent(event)
}

// EmitJobDLQ emits a job moved to dead letter queue event
func (m *Manager) EmitJobDLQ(jobID, queue string, priority, attempt int, errorMsg string) error {
	event := JobEvent{
		Event:    EventJobDLQ,
		JobID:    jobID,
		Queue:    queue,
		Priority: priority,
		Attempt:  attempt,
		Error:    errorMsg,
	}

	return m.EmitEvent(event)
}

// EmitJobRetried emits a job retry event
func (m *Manager) EmitJobRetried(jobID, queue string, priority, attempt int) error {
	event := JobEvent{
		Event:    EventJobRetried,
		JobID:    jobID,
		Queue:    queue,
		Priority: priority,
		Attempt:  attempt,
	}

	return m.EmitEvent(event)
}