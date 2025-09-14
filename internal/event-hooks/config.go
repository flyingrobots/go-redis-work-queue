// Copyright 2025 James Ross
package eventhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// Redis key prefixes
	webhookSubscriptionPrefix = "event_hooks:webhook:"
	natsSubscriptionPrefix    = "event_hooks:nats:"
	subscriptionIndexKey      = "event_hooks:subscriptions"
	dlhPrefix                 = "event_hooks:dlh:"
	dlhIndexPrefix            = "event_hooks:dlh_index:"
)

// ConfigManager handles persistence and management of subscriptions
type ConfigManager struct {
	redis  *redis.Client
	logger *slog.Logger
	mu     sync.RWMutex
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(redisClient *redis.Client, logger *slog.Logger) *ConfigManager {
	return &ConfigManager{
		redis:  redisClient,
		logger: logger,
	}
}

// CreateWebhookSubscription creates a new webhook subscription
func (cm *ConfigManager) CreateWebhookSubscription(ctx context.Context, req CreateWebhookRequest) (*WebhookSubscription, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Validate request
	if err := cm.validateWebhookRequest(req); err != nil {
		return nil, err
	}

	// Check for duplicate name
	existing, err := cm.findSubscriptionByName(ctx, req.Name)
	if err != nil && err != ErrSubscriptionNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicateSubscription
	}

	// Create subscription
	subscription := &WebhookSubscription{
		ID:             uuid.New().String(),
		Name:           req.Name,
		URL:            req.URL,
		Secret:         req.Secret,
		Events:         req.Events,
		Queues:         req.Queues,
		MinPriority:    req.MinPriority,
		MaxRetries:     req.MaxRetries,
		Timeout:        req.Timeout,
		RateLimit:      req.RateLimit,
		Headers:        req.Headers,
		IncludePayload: req.IncludePayload,
		PayloadFields:  req.PayloadFields,
		RedactFields:   req.RedactFields,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Disabled:       false,
	}

	// Set defaults
	if subscription.MaxRetries == 0 {
		subscription.MaxRetries = 5
	}
	if subscription.Timeout == 0 {
		subscription.Timeout = 30 * time.Second
	}
	if subscription.RateLimit == 0 {
		subscription.RateLimit = 60 // requests per minute
	}

	// Store in Redis
	err = cm.storeWebhookSubscription(ctx, subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to store subscription: %w", err)
	}

	cm.logger.Info("webhook subscription created",
		"subscription_id", subscription.ID,
		"name", subscription.Name,
		"url", subscription.URL)

	return subscription, nil
}

// GetWebhookSubscription retrieves a webhook subscription by ID
func (cm *ConfigManager) GetWebhookSubscription(ctx context.Context, id string) (*WebhookSubscription, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	key := webhookSubscriptionPrefix + id
	data, err := cm.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, ErrSubscriptionNotFound
		}
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}

	var subscription WebhookSubscription
	err = json.Unmarshal([]byte(data), &subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal subscription: %w", err)
	}

	return &subscription, nil
}

// ListWebhookSubscriptions returns all webhook subscriptions
func (cm *ConfigManager) ListWebhookSubscriptions(ctx context.Context) ([]*WebhookSubscription, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Get all subscription IDs
	pattern := webhookSubscriptionPrefix + "*"
	keys, err := cm.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list subscription keys: %w", err)
	}

	var subscriptions []*WebhookSubscription
	for _, key := range keys {
		id := strings.TrimPrefix(key, webhookSubscriptionPrefix)
		subscription, err := cm.GetWebhookSubscription(ctx, id)
		if err != nil {
			cm.logger.Warn("failed to get subscription",
				"subscription_id", id, "error", err)
			continue
		}
		subscriptions = append(subscriptions, subscription)
	}

	return subscriptions, nil
}

// UpdateWebhookSubscription updates an existing webhook subscription
func (cm *ConfigManager) UpdateWebhookSubscription(ctx context.Context, id string, req UpdateWebhookRequest) (*WebhookSubscription, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Get existing subscription
	subscription, err := cm.GetWebhookSubscription(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields
	if req.Name != nil {
		subscription.Name = *req.Name
	}
	if req.URL != nil {
		subscription.URL = *req.URL
	}
	if req.Secret != nil {
		subscription.Secret = *req.Secret
	}
	if req.Events != nil {
		subscription.Events = req.Events
	}
	if req.Queues != nil {
		subscription.Queues = req.Queues
	}
	if req.MinPriority != nil {
		subscription.MinPriority = req.MinPriority
	}
	if req.MaxRetries != nil {
		subscription.MaxRetries = *req.MaxRetries
	}
	if req.Timeout != nil {
		subscription.Timeout = *req.Timeout
	}
	if req.RateLimit != nil {
		subscription.RateLimit = *req.RateLimit
	}
	if req.Headers != nil {
		subscription.Headers = req.Headers
	}
	if req.IncludePayload != nil {
		subscription.IncludePayload = *req.IncludePayload
	}
	if req.PayloadFields != nil {
		subscription.PayloadFields = req.PayloadFields
	}
	if req.RedactFields != nil {
		subscription.RedactFields = req.RedactFields
	}
	if req.Disabled != nil {
		subscription.Disabled = *req.Disabled
	}

	subscription.UpdatedAt = time.Now()

	// Store updated subscription
	err = cm.storeWebhookSubscription(ctx, subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	cm.logger.Info("webhook subscription updated",
		"subscription_id", subscription.ID,
		"name", subscription.Name)

	return subscription, nil
}

// DeleteWebhookSubscription deletes a webhook subscription
func (cm *ConfigManager) DeleteWebhookSubscription(ctx context.Context, id string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if subscription exists
	_, err := cm.GetWebhookSubscription(ctx, id)
	if err != nil {
		return err
	}

	// Delete from Redis
	key := webhookSubscriptionPrefix + id
	err = cm.redis.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete subscription: %w", err)
	}

	// Remove from index
	cm.redis.SRem(ctx, subscriptionIndexKey, id)

	cm.logger.Info("webhook subscription deleted", "subscription_id", id)
	return nil
}

// CreateNATSSubscription creates a new NATS subscription
func (cm *ConfigManager) CreateNATSSubscription(ctx context.Context, req CreateNATSRequest) (*NATSSubscription, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Validate request
	if err := cm.validateNATSRequest(req); err != nil {
		return nil, err
	}

	// Check for duplicate name
	existing, err := cm.findNATSSubscriptionByName(ctx, req.Name)
	if err != nil && err != ErrSubscriptionNotFound {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDuplicateSubscription
	}

	// Create subscription
	subscription := &NATSSubscription{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Subject:   req.Subject,
		Events:    req.Events,
		Queues:    req.Queues,
		Headers:   req.Headers,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Disabled:  false,
	}

	// Store in Redis
	err = cm.storeNATSSubscription(ctx, subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to store NATS subscription: %w", err)
	}

	cm.logger.Info("NATS subscription created",
		"subscription_id", subscription.ID,
		"name", subscription.Name,
		"subject", subscription.Subject)

	return subscription, nil
}

// storeWebhookSubscription stores a webhook subscription in Redis
func (cm *ConfigManager) storeWebhookSubscription(ctx context.Context, subscription *WebhookSubscription) error {
	key := webhookSubscriptionPrefix + subscription.ID
	data, err := json.Marshal(subscription)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription: %w", err)
	}

	err = cm.redis.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to store subscription: %w", err)
	}

	// Add to index
	cm.redis.SAdd(ctx, subscriptionIndexKey, subscription.ID)

	return nil
}

// storeNATSSubscription stores a NATS subscription in Redis
func (cm *ConfigManager) storeNATSSubscription(ctx context.Context, subscription *NATSSubscription) error {
	key := natsSubscriptionPrefix + subscription.ID
	data, err := json.Marshal(subscription)
	if err != nil {
		return fmt.Errorf("failed to marshal NATS subscription: %w", err)
	}

	err = cm.redis.Set(ctx, key, data, 0).Err()
	if err != nil {
		return fmt.Errorf("failed to store NATS subscription: %w", err)
	}

	return nil
}

// findSubscriptionByName finds a webhook subscription by name
func (cm *ConfigManager) findSubscriptionByName(ctx context.Context, name string) (*WebhookSubscription, error) {
	subscriptions, err := cm.ListWebhookSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	for _, sub := range subscriptions {
		if sub.Name == name {
			return sub, nil
		}
	}

	return nil, ErrSubscriptionNotFound
}

// findNATSSubscriptionByName finds a NATS subscription by name
func (cm *ConfigManager) findNATSSubscriptionByName(ctx context.Context, name string) (*NATSSubscription, error) {
	pattern := natsSubscriptionPrefix + "*"
	keys, err := cm.redis.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to list NATS subscription keys: %w", err)
	}

	for _, key := range keys {
		data, err := cm.redis.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var subscription NATSSubscription
		err = json.Unmarshal([]byte(data), &subscription)
		if err != nil {
			continue
		}

		if subscription.Name == name {
			return &subscription, nil
		}
	}

	return nil, ErrSubscriptionNotFound
}

// validateWebhookRequest validates a webhook creation/update request
func (cm *ConfigManager) validateWebhookRequest(req CreateWebhookRequest) error {
	if req.Name == "" {
		return NewValidationError("name", "name is required", req.Name)
	}
	if req.URL == "" {
		return NewValidationError("url", "URL is required", req.URL)
	}
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		return NewValidationError("url", "URL must be HTTP or HTTPS", req.URL)
	}
	if len(req.Events) == 0 {
		return NewValidationError("events", "at least one event type is required", req.Events)
	}
	if len(req.Queues) == 0 {
		return NewValidationError("queues", "at least one queue is required", req.Queues)
	}
	if req.MaxRetries < 0 || req.MaxRetries > 20 {
		return NewValidationError("max_retries", "max_retries must be between 0 and 20", req.MaxRetries)
	}
	if req.Timeout < 0 || req.Timeout > 5*time.Minute {
		return NewValidationError("timeout", "timeout must be between 0 and 5 minutes", req.Timeout)
	}
	if req.RateLimit < 0 || req.RateLimit > 1000 {
		return NewValidationError("rate_limit", "rate_limit must be between 0 and 1000", req.RateLimit)
	}

	// Validate event types
	for _, event := range req.Events {
		if !cm.isValidEventType(event) {
			return NewValidationError("events", "invalid event type", event)
		}
	}

	return nil
}

// validateNATSRequest validates a NATS subscription request
func (cm *ConfigManager) validateNATSRequest(req CreateNATSRequest) error {
	if req.Name == "" {
		return NewValidationError("name", "name is required", req.Name)
	}
	if req.Subject == "" {
		return NewValidationError("subject", "NATS subject is required", req.Subject)
	}
	if len(req.Events) == 0 {
		return NewValidationError("events", "at least one event type is required", req.Events)
	}
	if len(req.Queues) == 0 {
		return NewValidationError("queues", "at least one queue is required", req.Queues)
	}

	// Validate event types
	for _, event := range req.Events {
		if !cm.isValidEventType(event) {
			return NewValidationError("events", "invalid event type", event)
		}
	}

	return nil
}

// isValidEventType checks if an event type is valid
func (cm *ConfigManager) isValidEventType(eventType EventType) bool {
	switch eventType {
	case EventJobEnqueued, EventJobStarted, EventJobSucceeded,
		 EventJobFailed, EventJobDLQ, EventJobRetried:
		return true
	default:
		return false
	}
}

// Request types for API operations
type CreateWebhookRequest struct {
	Name           string        `json:"name" validate:"required"`
	URL            string        `json:"url" validate:"required,url"`
	Secret         string        `json:"secret"`
	Events         []EventType   `json:"events" validate:"required,min=1"`
	Queues         []string      `json:"queues" validate:"required,min=1"`
	MinPriority    *int          `json:"min_priority,omitempty"`
	MaxRetries     int           `json:"max_retries"`
	Timeout        time.Duration `json:"timeout"`
	RateLimit      int           `json:"rate_limit"`
	Headers        []HeaderPair  `json:"headers"`
	IncludePayload bool          `json:"include_payload"`
	PayloadFields  []string      `json:"payload_fields,omitempty"`
	RedactFields   []string      `json:"redact_fields,omitempty"`
}

type UpdateWebhookRequest struct {
	Name           *string       `json:"name,omitempty"`
	URL            *string       `json:"url,omitempty"`
	Secret         *string       `json:"secret,omitempty"`
	Events         []EventType   `json:"events,omitempty"`
	Queues         []string      `json:"queues,omitempty"`
	MinPriority    *int          `json:"min_priority,omitempty"`
	MaxRetries     *int          `json:"max_retries,omitempty"`
	Timeout        *time.Duration `json:"timeout,omitempty"`
	RateLimit      *int          `json:"rate_limit,omitempty"`
	Headers        []HeaderPair  `json:"headers,omitempty"`
	IncludePayload *bool         `json:"include_payload,omitempty"`
	PayloadFields  []string      `json:"payload_fields,omitempty"`
	RedactFields   []string      `json:"redact_fields,omitempty"`
	Disabled       *bool         `json:"disabled,omitempty"`
}

type CreateNATSRequest struct {
	Name    string            `json:"name" validate:"required"`
	Subject string            `json:"subject" validate:"required"`
	Events  []EventType       `json:"events" validate:"required,min=1"`
	Queues  []string          `json:"queues" validate:"required,min=1"`
	Headers map[string]string `json:"headers"`
}