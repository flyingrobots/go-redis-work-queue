// Copyright 2025 James Ross
package eventhooks

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSPublisher implements EventSubscriber for NATS message publishing
type NATSPublisher struct {
	subscription *NATSSubscription
	conn         *nats.Conn
	js           nats.JetStreamContext
	filter       EventFilter
	logger       *slog.Logger
	mu           sync.RWMutex
	healthy      bool
}

// NewNATSPublisher creates a new NATS publisher
func NewNATSPublisher(subscription *NATSSubscription, natsURL string, logger *slog.Logger) (*NATSPublisher, error) {
	// Connect to NATS
	conn, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := conn.JetStream()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	// Create event filter
	filter := EventFilter{
		Events: subscription.Events,
		Queues: subscription.Queues,
	}

	return &NATSPublisher{
		subscription: subscription,
		conn:         conn,
		js:           js,
		filter:       filter,
		logger:       logger,
		healthy:      true,
	}, nil
}

// ID returns the subscriber ID
func (np *NATSPublisher) ID() string {
	return np.subscription.ID
}

// Name returns the subscriber name
func (np *NATSPublisher) Name() string {
	return np.subscription.Name
}

// GetFilter returns the event filter for this subscriber
func (np *NATSPublisher) GetFilter() EventFilter {
	return np.filter
}

// IsHealthy returns the health status of the subscriber
func (np *NATSPublisher) IsHealthy() bool {
	np.mu.RLock()
	defer np.mu.RUnlock()

	// Check if subscription is disabled
	if np.subscription.Disabled {
		return false
	}

	// Check NATS connection
	if np.conn == nil || !np.conn.IsConnected() {
		return false
	}

	return np.healthy
}

// ProcessEvent publishes an event to NATS
func (np *NATSPublisher) ProcessEvent(event JobEvent) error {
	np.mu.Lock()
	defer np.mu.Unlock()

	// Generate subject name
	subject := np.generateSubject(event)

	// Prepare message payload
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create NATS message
	msg := &nats.Msg{
		Subject: subject,
		Data:    payload,
		Header:  make(nats.Header),
	}

	// Add headers
	msg.Header.Set("Event-Type", string(event.Event))
	msg.Header.Set("Job-ID", event.JobID)
	msg.Header.Set("Queue", event.Queue)
	msg.Header.Set("Timestamp", event.Timestamp.Format(time.RFC3339))

	if event.TraceID != "" {
		msg.Header.Set("Trace-ID", event.TraceID)
	}
	if event.RequestID != "" {
		msg.Header.Set("Request-ID", event.RequestID)
	}

	// Add custom headers
	for key, value := range np.subscription.Headers {
		msg.Header.Set(key, value)
	}

	// Publish message
	_, err = np.js.PublishMsg(msg)
	if err != nil {
		np.logger.Warn("NATS publish failed",
			"subscription_id", np.subscription.ID,
			"subject", subject,
			"event_type", event.Event,
			"job_id", event.JobID,
			"error", err)
		return fmt.Errorf("NATS publish failed: %w", err)
	}

	np.logger.Debug("NATS publish successful",
		"subscription_id", np.subscription.ID,
		"subject", subject,
		"event_type", event.Event,
		"job_id", event.JobID)

	return nil
}

// Close shuts down the NATS publisher
func (np *NATSPublisher) Close() error {
	np.mu.Lock()
	defer np.mu.Unlock()

	np.healthy = false
	if np.conn != nil {
		np.conn.Close()
		np.conn = nil
	}

	np.logger.Info("NATS publisher closed", "subscription_id", np.subscription.ID)
	return nil
}

// generateSubject creates a NATS subject based on event and subscription config
func (np *NATSPublisher) generateSubject(event JobEvent) string {
	// Use configured subject as template
	subject := np.subscription.Subject

	// Replace placeholders with actual values
	// Support templates like: events.{queue}.{event_type}
	if len(subject) == 0 {
		// Default subject pattern
		subject = fmt.Sprintf("events.%s.%s", event.Queue, event.Event)
	} else {
		// Replace template variables
		subject = fmt.Sprintf(subject, event.Queue, event.Event, event.Priority)
	}

	return subject
}

// UpdateSubscription updates the NATS subscription configuration
func (np *NATSPublisher) UpdateSubscription(updated *NATSSubscription) error {
	np.mu.Lock()
	defer np.mu.Unlock()

	np.subscription = updated

	// Update filter
	np.filter = EventFilter{
		Events: updated.Events,
		Queues: updated.Queues,
	}

	// Reset health if subscription was re-enabled
	if !updated.Disabled {
		np.healthy = true
	}

	np.logger.Info("NATS subscription updated",
		"subscription_id", updated.ID,
		"subject", updated.Subject,
		"events", updated.Events)

	return nil
}

// NATSDeliverer manages multiple NATS publishers
type NATSDeliverer struct {
	publishers map[string]*NATSPublisher
	natsURL    string
	logger     *slog.Logger
	mu         sync.RWMutex
}

// NewNATSDeliverer creates a new NATS deliverer
func NewNATSDeliverer(natsURL string, logger *slog.Logger) *NATSDeliverer {
	return &NATSDeliverer{
		publishers: make(map[string]*NATSPublisher),
		natsURL:    natsURL,
		logger:     logger,
	}
}

// AddSubscription adds a new NATS subscription
func (nd *NATSDeliverer) AddSubscription(subscription *NATSSubscription) (*NATSPublisher, error) {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	publisher, err := NewNATSPublisher(subscription, nd.natsURL, nd.logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS publisher: %w", err)
	}

	nd.publishers[subscription.ID] = publisher

	nd.logger.Info("NATS subscription added",
		"subscription_id", subscription.ID,
		"name", subscription.Name,
		"subject", subscription.Subject)

	return publisher, nil
}

// RemoveSubscription removes a NATS subscription
func (nd *NATSDeliverer) RemoveSubscription(subscriptionID string) error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	publisher, exists := nd.publishers[subscriptionID]
	if !exists {
		return ErrSubscriptionNotFound
	}

	// Close the publisher
	publisher.Close()

	// Remove from map
	delete(nd.publishers, subscriptionID)

	nd.logger.Info("NATS subscription removed", "subscription_id", subscriptionID)
	return nil
}

// GetPublisher returns a NATS publisher by ID
func (nd *NATSDeliverer) GetPublisher(subscriptionID string) (*NATSPublisher, error) {
	nd.mu.RLock()
	defer nd.mu.RUnlock()

	publisher, exists := nd.publishers[subscriptionID]
	if !exists {
		return nil, ErrSubscriptionNotFound
	}

	return publisher, nil
}

// ListPublishers returns all NATS publishers
func (nd *NATSDeliverer) ListPublishers() map[string]*NATSPublisher {
	nd.mu.RLock()
	defer nd.mu.RUnlock()

	// Return a copy to prevent concurrent map access
	result := make(map[string]*NATSPublisher)
	for id, pub := range nd.publishers {
		result[id] = pub
	}

	return result
}

// UpdateSubscription updates an existing NATS subscription
func (nd *NATSDeliverer) UpdateSubscription(subscription *NATSSubscription) error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	publisher, exists := nd.publishers[subscription.ID]
	if !exists {
		return ErrSubscriptionNotFound
	}

	return publisher.UpdateSubscription(subscription)
}

// Close shuts down all NATS publishers
func (nd *NATSDeliverer) Close() error {
	nd.mu.Lock()
	defer nd.mu.Unlock()

	for id, publisher := range nd.publishers {
		if err := publisher.Close(); err != nil {
			nd.logger.Warn("error closing NATS publisher",
				"subscription_id", id, "error", err)
		}
	}

	clear(nd.publishers)
	nd.logger.Info("NATS deliverer closed")
	return nil
}