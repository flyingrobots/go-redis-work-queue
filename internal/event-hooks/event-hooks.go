// Copyright 2025 James Ross
package eventhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// EventBus manages event distribution to subscribers
type EventBus struct {
	config      EventBusConfig
	subscribers map[EventType][]EventSubscriber
	eventQueue  chan JobEvent
	retryQueue  chan *DeliveryAttempt
	dlhQueue    chan *DeadLetterHook
	metrics     *EventMetrics
	redis       *redis.Client
	logger      *slog.Logger

	// Shutdown coordination
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	mu        sync.RWMutex
	isRunning bool
}

// NewEventBus creates a new event bus instance
func NewEventBus(config EventBusConfig, redisClient *redis.Client, logger *slog.Logger) *EventBus {
	ctx, cancel := context.WithCancel(context.Background())

	bus := &EventBus{
		config:      config,
		subscribers: make(map[EventType][]EventSubscriber),
		eventQueue:  make(chan JobEvent, config.BufferSize),
		retryQueue:  make(chan *DeliveryAttempt, config.BufferSize/2),
		dlhQueue:    make(chan *DeadLetterHook, config.BufferSize/10),
		metrics:     &EventMetrics{SubscriptionHealth: make(map[string]float64)},
		redis:       redisClient,
		logger:      logger,
		ctx:         ctx,
		cancel:      cancel,
		isRunning:   false,
	}

	return bus
}

// Start begins event processing
func (eb *EventBus) Start() error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.isRunning {
		return fmt.Errorf("event bus is already running")
	}

	eb.logger.Info("starting event bus",
		"worker_pool_size", eb.config.WorkerPoolSize,
		"buffer_size", eb.config.BufferSize)

	// Start worker pool for event processing
	for i := 0; i < eb.config.WorkerPoolSize; i++ {
		eb.wg.Add(1)
		go eb.eventWorker(i)
	}

	// Start retry processor
	eb.wg.Add(1)
	go eb.retryProcessor()

	// Start DLH processor
	eb.wg.Add(1)
	go eb.dlhProcessor()

	// Start metrics collector
	eb.wg.Add(1)
	go eb.metricsCollector()

	eb.isRunning = true
	return nil
}

// Stop gracefully shuts down the event bus
func (eb *EventBus) Stop() error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if !eb.isRunning {
		return fmt.Errorf("event bus is not running")
	}

	eb.logger.Info("stopping event bus")
	eb.cancel()

	// Close event queue to signal workers
	close(eb.eventQueue)

	// Wait for all workers to finish
	eb.wg.Wait()

	eb.isRunning = false
	eb.logger.Info("event bus stopped")
	return nil
}

// Emit sends an event to all matching subscribers
func (eb *EventBus) Emit(event JobEvent) error {
	if !eb.isRunning {
		return ErrEventBusShutdown
	}

	// Set event timestamp if not provided
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Generate correlation ID if not provided
	if event.TraceID == "" {
		event.TraceID = uuid.New().String()
	}

	// Add deep links
	event.Links = eb.generateDeepLinks(event)

	select {
	case eb.eventQueue <- event:
		eb.metrics.EventsEmitted++
		return nil
	case <-eb.ctx.Done():
		return ErrEventBusShutdown
	default:
		// Queue is full, log warning and drop event
		eb.logger.Warn("event queue full, dropping event",
			"event_type", event.Event,
			"job_id", event.JobID)
		return fmt.Errorf("event queue full")
	}
}

// Subscribe adds a new subscriber to the event bus
func (eb *EventBus) Subscribe(subscriber EventSubscriber) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	filter := subscriber.GetFilter()
	for _, eventType := range filter.Events {
		eb.subscribers[eventType] = append(eb.subscribers[eventType], subscriber)
	}

	eb.logger.Info("subscriber added",
		"subscriber_id", subscriber.ID(),
		"subscriber_name", subscriber.Name(),
		"events", filter.Events)

	return nil
}

// Unsubscribe removes a subscriber from the event bus
func (eb *EventBus) Unsubscribe(subscriberID string) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	for eventType, subscribers := range eb.subscribers {
		for i, subscriber := range subscribers {
			if subscriber.ID() == subscriberID {
				// Remove subscriber from slice
				eb.subscribers[eventType] = append(subscribers[:i], subscribers[i+1:]...)

				// Close subscriber
				if err := subscriber.Close(); err != nil {
					eb.logger.Warn("error closing subscriber",
						"subscriber_id", subscriberID,
						"error", err)
				}

				eb.logger.Info("subscriber removed", "subscriber_id", subscriberID)
				return nil
			}
		}
	}

	return fmt.Errorf("subscriber not found: %s", subscriberID)
}

// GetMetrics returns current event bus metrics
func (eb *EventBus) GetMetrics() EventMetrics {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	// Create a copy to avoid race conditions
	metrics := *eb.metrics
	healthCopy := make(map[string]float64)
	for k, v := range eb.metrics.SubscriptionHealth {
		healthCopy[k] = v
	}
	metrics.SubscriptionHealth = healthCopy

	return metrics
}

// eventWorker processes events from the queue
func (eb *EventBus) eventWorker(workerID int) {
	defer eb.wg.Done()

	eb.logger.Debug("event worker started", "worker_id", workerID)

	for {
		select {
		case event, ok := <-eb.eventQueue:
			if !ok {
				// Queue closed, exit worker
				eb.logger.Debug("event worker stopping", "worker_id", workerID)
				return
			}

			eb.processEvent(event, workerID)

		case <-eb.ctx.Done():
			eb.logger.Debug("event worker stopping due to context cancellation", "worker_id", workerID)
			return
		}
	}
}

// processEvent sends an event to all matching subscribers
func (eb *EventBus) processEvent(event JobEvent, workerID int) {
	eb.mu.RLock()
	subscribers := eb.subscribers[event.Event]
	eb.mu.RUnlock()

	if len(subscribers) == 0 {
		return
	}

	eb.logger.Debug("processing event",
		"worker_id", workerID,
		"event_type", event.Event,
		"job_id", event.JobID,
		"subscriber_count", len(subscribers))

	for _, subscriber := range subscribers {
		// Check if subscriber is healthy
		if !subscriber.IsHealthy() {
			eb.logger.Warn("skipping unhealthy subscriber",
				"subscriber_id", subscriber.ID(),
				"event_type", event.Event)
			continue
		}

		// Check if event matches subscriber filter
		filter := subscriber.GetFilter()
		if !filter.Matches(event) {
			continue
		}

		// Process event in goroutine to avoid blocking
		go func(sub EventSubscriber, evt JobEvent) {
			err := sub.ProcessEvent(evt)
			if err != nil {
				eb.handleDeliveryError(sub, evt, err)
			} else {
				eb.handleDeliverySuccess(sub, evt)
			}
		}(subscriber, event)
	}
}

// handleDeliveryError processes delivery failures
func (eb *EventBus) handleDeliveryError(subscriber EventSubscriber, event JobEvent, err error) {
	eb.logger.Warn("delivery failed",
		"subscriber_id", subscriber.ID(),
		"event_type", event.Event,
		"job_id", event.JobID,
		"error", err)

	eb.metrics.WebhookFailures++

	// For webhook subscribers, handle retries
	if webhookSub, ok := subscriber.(*WebhookSubscriber); ok {
		attempt := &DeliveryAttempt{
			ID:             uuid.New().String(),
			SubscriptionID: webhookSub.ID(),
			Event:          event,
			AttemptNumber:  1,
			ScheduledAt:    time.Now(),
			Success:        false,
			ErrorMessage:   err.Error(),
		}

		// Check if error is retryable
		if IsRetryableError(err) && webhookSub.subscription.MaxRetries > 0 {
			select {
			case eb.retryQueue <- attempt:
				eb.logger.Debug("scheduled retry",
					"subscription_id", webhookSub.ID(),
					"attempt_id", attempt.ID)
			default:
				eb.logger.Warn("retry queue full, sending to DLH",
					"subscription_id", webhookSub.ID())
				eb.sendToDLH(webhookSub.subscription, event, []*DeliveryAttempt{attempt}, err.Error())
			}
		} else {
			// Non-retryable error or max retries reached, send to DLH
			eb.sendToDLH(webhookSub.subscription, event, []*DeliveryAttempt{attempt}, err.Error())
		}
	}
}

// handleDeliverySuccess processes successful deliveries
func (eb *EventBus) handleDeliverySuccess(subscriber EventSubscriber, event JobEvent) {
	eb.logger.Debug("delivery succeeded",
		"subscriber_id", subscriber.ID(),
		"event_type", event.Event,
		"job_id", event.JobID)

	eb.metrics.WebhookDeliveries++
}

// retryProcessor handles retry attempts
func (eb *EventBus) retryProcessor() {
	defer eb.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case attempt := <-eb.retryQueue:
			eb.processRetryAttempt(attempt)

		case <-ticker.C:
			// Process scheduled retries
			eb.processScheduledRetries()

		case <-eb.ctx.Done():
			return
		}
	}
}

// processRetryAttempt schedules a retry attempt
func (eb *EventBus) processRetryAttempt(attempt *DeliveryAttempt) {
	// Calculate retry delay using exponential backoff
	policy := DefaultRetryPolicy()
	delay := eb.calculateRetryDelay(policy, attempt.AttemptNumber)

	attempt.ScheduledAt = time.Now().Add(delay)

	// Store retry attempt in Redis
	key := fmt.Sprintf("event_hooks:retry:%s", attempt.ID)
	data, err := json.Marshal(attempt)
	if err != nil {
		eb.logger.Error("failed to marshal retry attempt", "error", err)
		return
	}

	err = eb.redis.Set(eb.ctx, key, data, delay+time.Minute).Err()
	if err != nil {
		eb.logger.Error("failed to store retry attempt", "error", err)
		return
	}

	eb.logger.Debug("retry scheduled",
		"attempt_id", attempt.ID,
		"delay", delay,
		"scheduled_at", attempt.ScheduledAt)
}

// calculateRetryDelay calculates the next retry delay
func (eb *EventBus) calculateRetryDelay(policy RetryPolicy, attempt int) time.Duration {
	var delay time.Duration

	switch policy.Strategy {
	case "exponential":
		delay = time.Duration(float64(policy.InitialDelay) *
			(math.Pow(policy.Multiplier, float64(attempt-1))))
	case "linear":
		delay = time.Duration(float64(policy.InitialDelay) * float64(attempt))
	case "fixed":
		delay = policy.InitialDelay
	default:
		delay = policy.InitialDelay
	}

	// Apply maximum delay cap
	if delay > policy.MaxDelay {
		delay = policy.MaxDelay
	}

	// Add jitter to prevent thundering herd
	if policy.Jitter && delay > 0 {
		jitter := time.Duration(rand.Int63n(int64(delay / 4)))
		delay += jitter
	}

	return delay
}

// processScheduledRetries processes retry attempts that are due
func (eb *EventBus) processScheduledRetries() {
	// Scan for ready retry attempts
	pattern := "event_hooks:retry:*"
	iter := eb.redis.Scan(eb.ctx, 0, pattern, 100).Iterator()

	for iter.Next(eb.ctx) {
		key := iter.Val()

		// Get retry attempt data
		data, err := eb.redis.Get(eb.ctx, key).Result()
		if err != nil {
			if err != redis.Nil {
				eb.logger.Warn("failed to get retry attempt", "key", key, "error", err)
			}
			continue
		}

		var attempt DeliveryAttempt
		err = json.Unmarshal([]byte(data), &attempt)
		if err != nil {
			eb.logger.Warn("failed to unmarshal retry attempt", "key", key, "error", err)
			continue
		}

		// Check if retry is due
		if time.Now().After(attempt.ScheduledAt) {
			// Remove from Redis
			eb.redis.Del(eb.ctx, key)

			// Execute retry
			go eb.executeRetry(&attempt)
		}
	}
}

// executeRetry executes a retry attempt
func (eb *EventBus) executeRetry(attempt *DeliveryAttempt) {
	eb.logger.Debug("executing retry",
		"attempt_id", attempt.ID,
		"subscription_id", attempt.SubscriptionID,
		"attempt_number", attempt.AttemptNumber)

	// Find the webhook subscriber
	var webhookSub *WebhookSubscriber
	eb.mu.RLock()
	for _, subscribers := range eb.subscribers {
		for _, sub := range subscribers {
			if sub.ID() == attempt.SubscriptionID {
				if ws, ok := sub.(*WebhookSubscriber); ok {
					webhookSub = ws
					break
				}
			}
		}
		if webhookSub != nil {
			break
		}
	}
	eb.mu.RUnlock()

	if webhookSub == nil {
		eb.logger.Warn("webhook subscriber not found for retry",
			"subscription_id", attempt.SubscriptionID)
		return
	}

	// Attempt delivery
	err := webhookSub.ProcessEvent(attempt.Event)
	if err != nil {
		eb.logger.Warn("retry attempt failed",
			"attempt_id", attempt.ID,
			"error", err)

		// Check if we should retry again
		if attempt.AttemptNumber < webhookSub.subscription.MaxRetries {
			attempt.AttemptNumber++
			select {
			case eb.retryQueue <- attempt:
				eb.metrics.RetryAttempts++
			default:
				// Retry queue full, send to DLH
				eb.sendToDLH(webhookSub.subscription, attempt.Event, []*DeliveryAttempt{attempt}, err.Error())
			}
		} else {
			// Max retries exceeded, send to DLH
			eb.sendToDLH(webhookSub.subscription, attempt.Event, []*DeliveryAttempt{attempt}, err.Error())
		}
	} else {
		eb.logger.Info("retry succeeded",
			"attempt_id", attempt.ID,
			"subscription_id", attempt.SubscriptionID)
		eb.metrics.WebhookDeliveries++
	}
}

// sendToDLH sends failed deliveries to the dead letter hooks queue
func (eb *EventBus) sendToDLH(subscription *WebhookSubscription, event JobEvent, attempts []*DeliveryAttempt, finalError string) {
	dlh := &DeadLetterHook{
		ID:             uuid.New().String(),
		SubscriptionID: subscription.ID,
		Event:          event,
		FinalError:     finalError,
		CreatedAt:      time.Now(),
	}

	// Convert delivery attempts
	for _, attempt := range attempts {
		dlh.Attempts = append(dlh.Attempts, *attempt)
	}

	select {
	case eb.dlhQueue <- dlh:
		eb.logger.Info("sent to dead letter hooks",
			"dlh_id", dlh.ID,
			"subscription_id", subscription.ID,
			"event_type", event.Event,
			"job_id", event.JobID)
	default:
		eb.logger.Error("DLH queue full, dropping dead letter hook",
			"subscription_id", subscription.ID,
			"event_type", event.Event,
			"job_id", event.JobID)
	}
}

// dlhProcessor handles dead letter hooks
func (eb *EventBus) dlhProcessor() {
	defer eb.wg.Done()

	for {
		select {
		case dlh := <-eb.dlhQueue:
			eb.storeDLH(dlh)

		case <-eb.ctx.Done():
			return
		}
	}
}

// storeDLH stores a dead letter hook in Redis
func (eb *EventBus) storeDLH(dlh *DeadLetterHook) {
	key := fmt.Sprintf("event_hooks:dlh:%s", dlh.ID)
	data, err := json.Marshal(dlh)
	if err != nil {
		eb.logger.Error("failed to marshal DLH", "error", err)
		return
	}

	// Store with 30-day expiration
	err = eb.redis.Set(eb.ctx, key, data, 30*24*time.Hour).Err()
	if err != nil {
		eb.logger.Error("failed to store DLH", "error", err)
		return
	}

	// Add to DLH index for subscription
	indexKey := fmt.Sprintf("event_hooks:dlh_index:%s", dlh.SubscriptionID)
	eb.redis.LPush(eb.ctx, indexKey, dlh.ID)
	eb.redis.Expire(eb.ctx, indexKey, 30*24*time.Hour)

	eb.metrics.DLHSize++

	eb.logger.Info("DLH stored",
		"dlh_id", dlh.ID,
		"subscription_id", dlh.SubscriptionID)
}

// metricsCollector periodically updates metrics
func (eb *EventBus) metricsCollector() {
	defer eb.wg.Done()

	ticker := time.NewTicker(eb.config.MetricsInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			eb.updateMetrics()

		case <-eb.ctx.Done():
			return
		}
	}
}

// updateMetrics updates collected metrics
func (eb *EventBus) updateMetrics() {
	// Update DLH size
	pattern := "event_hooks:dlh:*"
	keys, err := eb.redis.Keys(eb.ctx, pattern).Result()
	if err == nil {
		eb.metrics.DLHSize = int64(len(keys))
	}

	// Update subscription health metrics
	eb.mu.RLock()
	for _, subscribers := range eb.subscribers {
		for _, sub := range subscribers {
			if webhookSub, ok := sub.(*WebhookSubscriber); ok {
				health := eb.calculateSubscriptionHealth(webhookSub)
				eb.metrics.SubscriptionHealth[webhookSub.ID()] = health
			}
		}
	}
	eb.mu.RUnlock()
}

// calculateSubscriptionHealth calculates health score for a subscription
func (eb *EventBus) calculateSubscriptionHealth(sub *WebhookSubscriber) float64 {
	sub.subscription.mu.RLock()
	defer sub.subscription.mu.RUnlock()

	// Simple health calculation based on recent failure rate
	if sub.subscription.FailureCount == 0 {
		return 1.0
	}

	// Health decreases with consecutive failures
	health := 1.0 - (float64(sub.subscription.FailureCount) / 10.0)
	if health < 0 {
		health = 0
	}

	return health
}

// generateDeepLinks creates TUI deep links for an event
func (eb *EventBus) generateDeepLinks(event JobEvent) map[string]string {
	baseURL := "queue://localhost:8080" // TODO: Make configurable

	links := map[string]string{
		"job_details":     fmt.Sprintf("%s/jobs/%s", baseURL, event.JobID),
		"queue_dashboard": fmt.Sprintf("%s/queues/%s", baseURL, event.Queue),
	}

	if event.Event == EventJobFailed || event.Event == EventJobDLQ {
		links["retry_job"] = fmt.Sprintf("%s/jobs/%s/retry", baseURL, event.JobID)
		links["dlq_browser"] = fmt.Sprintf("%s/dlq/%s", baseURL, event.Queue)
	}

	return links
}

// GetDLHEntries returns dead letter hook entries for a subscription
func (eb *EventBus) GetDLHEntries(subscriptionID string, limit int) ([]*DeadLetterHook, error) {
	indexKey := fmt.Sprintf("event_hooks:dlh_index:%s", subscriptionID)
	ids, err := eb.redis.LRange(eb.ctx, indexKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	var entries []*DeadLetterHook
	for _, id := range ids {
		key := fmt.Sprintf("event_hooks:dlh:%s", id)
		data, err := eb.redis.Get(eb.ctx, key).Result()
		if err != nil {
			if err != redis.Nil {
				eb.logger.Warn("failed to get DLH entry", "id", id, "error", err)
			}
			continue
		}

		var dlh DeadLetterHook
		err = json.Unmarshal([]byte(data), &dlh)
		if err != nil {
			eb.logger.Warn("failed to unmarshal DLH entry", "id", id, "error", err)
			continue
		}

		entries = append(entries, &dlh)
	}

	return entries, nil
}