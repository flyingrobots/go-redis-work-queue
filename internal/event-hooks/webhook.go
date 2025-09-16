// Copyright 2025 James Ross
package eventhooks

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"golang.org/x/time/rate"
)

// WebhookSubscriber implements EventSubscriber for HTTP webhooks
type WebhookSubscriber struct {
	subscription *WebhookSubscription
	client       *http.Client
	rateLimiter  *rate.Limiter
	filter       EventFilter
	logger       *slog.Logger
	mu           sync.RWMutex
	healthy      bool
}

// NewWebhookSubscriber creates a new webhook subscriber
func NewWebhookSubscriber(subscription *WebhookSubscription, logger *slog.Logger) *WebhookSubscriber {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: subscription.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
			MaxIdleConnsPerHost: 2,
		},
	}

	// Create rate limiter
	var rateLimiter *rate.Limiter
	if subscription.RateLimit > 0 {
		rateLimiter = rate.NewLimiter(rate.Limit(subscription.RateLimit)/60, subscription.RateLimit)
	}

	// Create event filter
	filter := EventFilter{
		Events:      subscription.Events,
		Queues:      subscription.Queues,
		MinPriority: subscription.MinPriority,
	}

	return &WebhookSubscriber{
		subscription: subscription,
		client:       client,
		rateLimiter:  rateLimiter,
		filter:       filter,
		logger:       logger,
		healthy:      true,
	}
}

// ID returns the subscriber ID
func (ws *WebhookSubscriber) ID() string {
	return ws.subscription.ID
}

// Name returns the subscriber name
func (ws *WebhookSubscriber) Name() string {
	return ws.subscription.Name
}

// GetFilter returns the event filter for this subscriber
func (ws *WebhookSubscriber) GetFilter() EventFilter {
	return ws.filter
}

// IsHealthy returns the health status of the subscriber
func (ws *WebhookSubscriber) IsHealthy() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	// Check if subscription is disabled
	if ws.subscription.Disabled {
		return false
	}

	// Check consecutive failure threshold
	if ws.subscription.FailureCount > 10 {
		return false
	}

	return ws.healthy
}

// ProcessEvent delivers an event via webhook
func (ws *WebhookSubscriber) ProcessEvent(event JobEvent) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Check rate limit
	if ws.rateLimiter != nil {
		if !ws.rateLimiter.Allow() {
			return NewDeliveryError(ws.subscription.ID, event.JobID, 1, 429,
				"rate limit exceeded", true, ErrRateLimitExceeded)
		}
	}

	// Prepare payload
	payload, err := ws.preparePayload(event)
	if err != nil {
		return NewDeliveryError(ws.subscription.ID, event.JobID, 1, 0,
			"payload preparation failed", false, err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", ws.subscription.URL, bytes.NewBuffer(payload))
	if err != nil {
		return NewDeliveryError(ws.subscription.ID, event.JobID, 1, 0,
			"request creation failed", false, err)
	}

	// Set headers
	err = ws.setRequestHeaders(req, payload, event)
	if err != nil {
		return NewDeliveryError(ws.subscription.ID, event.JobID, 1, 0,
			"header setting failed", false, err)
	}

	// Execute request
	start := time.Now()
	resp, err := ws.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		ws.handleDeliveryFailure(event, 0, err.Error())
		return NewDeliveryError(ws.subscription.ID, event.JobID, 1, 0,
			"request failed", true, err)
	}
	defer resp.Body.Close()

	// Read response body (limited)
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		ws.logger.Warn("failed to read response body", "error", err)
	}

	// Check response status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		ws.handleDeliverySuccess(event, resp.StatusCode, duration)
		ws.logger.Debug("webhook delivery successful",
			"subscription_id", ws.subscription.ID,
			"event_type", event.Event,
			"job_id", event.JobID,
			"status_code", resp.StatusCode,
			"duration", duration)
		return nil
	}

	// Handle error response
	errorMsg := fmt.Sprintf("HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	ws.handleDeliveryFailure(event, resp.StatusCode, errorMsg)

	retryable := IsTemporaryError(resp.StatusCode)
	return NewDeliveryError(ws.subscription.ID, event.JobID, 1, resp.StatusCode,
		errorMsg, retryable, nil)
}

// Close shuts down the webhook subscriber
func (ws *WebhookSubscriber) Close() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	ws.healthy = false
	if ws.client != nil {
		ws.client.CloseIdleConnections()
	}
	return nil
}

// preparePayload creates the webhook payload
func (ws *WebhookSubscriber) preparePayload(event JobEvent) ([]byte, error) {
	// Create a copy of the event to modify
	payload := event

	// Apply payload filtering if configured
	if !ws.subscription.IncludePayload {
		payload.Payload = nil
	} else if len(ws.subscription.PayloadFields) > 0 {
		// Filter payload fields (would need reflection or type assertion)
		// For now, keep the full payload
	}

	// Apply field redaction
	if len(ws.subscription.RedactFields) > 0 {
		payload = ws.redactFields(payload, ws.subscription.RedactFields)
	}

	return json.Marshal(payload)
}

// redactFields removes sensitive fields from the payload
func (ws *WebhookSubscriber) redactFields(event JobEvent, redactFields []string) JobEvent {
	// Create a copy
	redacted := event

	// Redact fields based on field names
	for _, field := range redactFields {
		switch field {
		case "user_id":
			redacted.UserID = "[REDACTED]"
		case "trace_id":
			redacted.TraceID = "[REDACTED]"
		case "request_id":
			redacted.RequestID = "[REDACTED]"
		case "payload":
			redacted.Payload = "[REDACTED]"
		}
	}

	return redacted
}

// setRequestHeaders sets the appropriate HTTP headers
func (ws *WebhookSubscriber) setRequestHeaders(req *http.Request, payload []byte, event JobEvent) error {
	// Set content type
	req.Header.Set("Content-Type", "application/json")

	// Set user agent
	req.Header.Set("User-Agent", "go-redis-work-queue/1.0")

	// Generate delivery ID
	deliveryID := uuid.New().String()
	req.Header.Set("X-Webhook-Delivery", deliveryID)

	// Set event metadata
	req.Header.Set("X-Webhook-Event", string(event.Event))
	req.Header.Set("X-Webhook-Timestamp", strconv.FormatInt(event.Timestamp.Unix(), 10))
	req.Header.Set("X-Webhook-Job-ID", event.JobID)
	req.Header.Set("X-Webhook-Queue", event.Queue)

	// Add trace headers if available
	if event.TraceID != "" {
		req.Header.Set("X-Trace-ID", event.TraceID)
	}
	if event.RequestID != "" {
		req.Header.Set("X-Request-ID", event.RequestID)
	}

	// Generate HMAC signature
	if ws.subscription.Secret != "" {
		signature := ws.generateSignature(payload, ws.subscription.Secret)
		req.Header.Set("X-Webhook-Signature", signature)
	}

	// Add custom headers
	for _, header := range ws.subscription.Headers {
		req.Header.Set(header.Key, header.Value)
	}

	return nil
}

// generateSignature creates an HMAC signature for the payload
func (ws *WebhookSubscriber) generateSignature(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	signature := h.Sum(nil)
	return fmt.Sprintf("sha256=%x", signature)
}

// handleDeliverySuccess updates subscription stats for successful delivery
func (ws *WebhookSubscriber) handleDeliverySuccess(event JobEvent, statusCode int, duration time.Duration) {
	ws.subscription.mu.Lock()
	defer ws.subscription.mu.Unlock()

	now := time.Now()
	ws.subscription.LastSuccess = &now
	ws.subscription.FailureCount = 0 // Reset failure count on success
	ws.subscription.UpdatedAt = now

	ws.logger.Debug("webhook delivery success recorded",
		"subscription_id", ws.subscription.ID,
		"status_code", statusCode,
		"duration", duration)
}

// handleDeliveryFailure updates subscription stats for failed delivery
func (ws *WebhookSubscriber) handleDeliveryFailure(event JobEvent, statusCode int, errorMsg string) {
	ws.subscription.mu.Lock()
	defer ws.subscription.mu.Unlock()

	now := time.Now()
	ws.subscription.LastFailure = &now
	ws.subscription.FailureCount++
	ws.subscription.UpdatedAt = now

	// Mark as unhealthy if too many consecutive failures
	if ws.subscription.FailureCount > 5 {
		ws.healthy = false
	}

	ws.logger.Warn("webhook delivery failure recorded",
		"subscription_id", ws.subscription.ID,
		"status_code", statusCode,
		"error", errorMsg,
		"failure_count", ws.subscription.FailureCount)
}

// UpdateSubscription updates the subscription configuration
func (ws *WebhookSubscriber) UpdateSubscription(updated *WebhookSubscription) error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	// Update subscription
	ws.subscription = updated

	// Update rate limiter if needed
	if updated.RateLimit > 0 {
		ws.rateLimiter = rate.NewLimiter(rate.Limit(updated.RateLimit)/60, updated.RateLimit)
	} else {
		ws.rateLimiter = nil
	}

	// Update filter
	ws.filter = EventFilter{
		Events:      updated.Events,
		Queues:      updated.Queues,
		MinPriority: updated.MinPriority,
	}

	// Update HTTP client timeout
	ws.client.Timeout = updated.Timeout

	// Reset health if subscription was re-enabled
	if !updated.Disabled {
		ws.healthy = true
	}

	ws.logger.Info("webhook subscription updated",
		"subscription_id", updated.ID,
		"url", updated.URL,
		"events", updated.Events)

	return nil
}

// TestDelivery sends a test event to verify webhook configuration
func (ws *WebhookSubscriber) TestDelivery() error {
	testEvent := JobEvent{
		Event:     EventJobSucceeded,
		Timestamp: time.Now(),
		JobID:     "test-job-" + uuid.New().String(),
		Queue:     "test-queue",
		Priority:  5,
		Attempt:   1,
		Duration:  func() *time.Duration { d := 1500 * time.Millisecond; return &d }(),
		Worker:    "test-worker",
		Links: map[string]string{
			"test": "This is a test webhook delivery",
		},
	}

	return ws.ProcessEvent(testEvent)
}

// GetHealthStatus returns detailed health information
func (ws *WebhookSubscriber) GetHealthStatus() SubscriptionHealthStatus {
	ws.mu.RLock()
	defer ws.mu.RUnlock()

	status := SubscriptionHealthStatus{
		SubscriptionID:      ws.subscription.ID,
		ConsecutiveFailures: ws.subscription.FailureCount,
	}

	if ws.subscription.LastSuccess != nil {
		status.LastSuccess = ws.subscription.LastSuccess
	}
	if ws.subscription.LastFailure != nil {
		status.LastFailure = ws.subscription.LastFailure
	}

	// Calculate success rate (simplified)
	if ws.subscription.FailureCount == 0 {
		status.SuccessRate = 1.0
	} else {
		// Simple calculation - would be more sophisticated in real implementation
		status.SuccessRate = 1.0 - (float64(ws.subscription.FailureCount) / 100.0)
		if status.SuccessRate < 0 {
			status.SuccessRate = 0
		}
	}

	return status
}

// WebhookDeliverer manages multiple webhook subscribers
type WebhookDeliverer struct {
	subscribers map[string]*WebhookSubscriber
	logger      *slog.Logger
	mu          sync.RWMutex
}

// NewWebhookDeliverer creates a new webhook deliverer
func NewWebhookDeliverer(logger *slog.Logger) *WebhookDeliverer {
	return &WebhookDeliverer{
		subscribers: make(map[string]*WebhookSubscriber),
		logger:      logger,
	}
}

// AddSubscription adds a new webhook subscription
func (wd *WebhookDeliverer) AddSubscription(subscription *WebhookSubscription) *WebhookSubscriber {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	subscriber := NewWebhookSubscriber(subscription, wd.logger)
	wd.subscribers[subscription.ID] = subscriber

	wd.logger.Info("webhook subscription added",
		"subscription_id", subscription.ID,
		"name", subscription.Name,
		"url", subscription.URL)

	return subscriber
}

// RemoveSubscription removes a webhook subscription
func (wd *WebhookDeliverer) RemoveSubscription(subscriptionID string) error {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	subscriber, exists := wd.subscribers[subscriptionID]
	if !exists {
		return ErrSubscriptionNotFound
	}

	// Close the subscriber
	subscriber.Close()

	// Remove from map
	delete(wd.subscribers, subscriptionID)

	wd.logger.Info("webhook subscription removed", "subscription_id", subscriptionID)
	return nil
}

// GetSubscriber returns a webhook subscriber by ID
func (wd *WebhookDeliverer) GetSubscriber(subscriptionID string) (*WebhookSubscriber, error) {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	subscriber, exists := wd.subscribers[subscriptionID]
	if !exists {
		return nil, ErrSubscriptionNotFound
	}

	return subscriber, nil
}

// ListSubscribers returns all webhook subscribers
func (wd *WebhookDeliverer) ListSubscribers() map[string]*WebhookSubscriber {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	// Return a copy to prevent concurrent map access
	result := make(map[string]*WebhookSubscriber)
	for id, sub := range wd.subscribers {
		result[id] = sub
	}

	return result
}

// UpdateSubscription updates an existing webhook subscription
func (wd *WebhookDeliverer) UpdateSubscription(subscription *WebhookSubscription) error {
	wd.mu.Lock()
	defer wd.mu.Unlock()

	subscriber, exists := wd.subscribers[subscription.ID]
	if !exists {
		return ErrSubscriptionNotFound
	}

	return subscriber.UpdateSubscription(subscription)
}

// GetHealthStatuses returns health status for all subscriptions
func (wd *WebhookDeliverer) GetHealthStatuses() []SubscriptionHealthStatus {
	wd.mu.RLock()
	defer wd.mu.RUnlock()

	var statuses []SubscriptionHealthStatus
	for _, subscriber := range wd.subscribers {
		statuses = append(statuses, subscriber.GetHealthStatus())
	}

	return statuses
}