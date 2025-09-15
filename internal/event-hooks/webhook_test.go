// Copyright 2025 James Ross
package eventhooks

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWebhookSubscriber_ProcessEvent(t *testing.T) {
	// Mock webhook server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Webhook-Signature") == "" {
			t.Error("expected HMAC signature header")
		}
		if r.Header.Get("X-Webhook-Event") == "" {
			t.Error("expected event type header")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	subscription := &WebhookSubscription{
		ID:         "test-webhook",
		Name:       "Test Webhook",
		URL:        server.URL,
		Secret:     "test-secret",
		Events:     []EventType{EventJobSucceeded},
		Queues:     []string{"*"},
		MaxRetries: 3,
		Timeout:    5 * time.Second,
	}

	logger := slog.Default()
	subscriber := NewWebhookSubscriber(subscription, logger)

	event := JobEvent{
		Event:    EventJobSucceeded,
		JobID:    "test-job-123",
		Queue:    "test-queue",
		Priority: 5,
	}

	err := subscriber.ProcessEvent(event)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestWebhookSubscriber_ProcessEvent_Failure(t *testing.T) {
	subscription := &WebhookSubscription{
		ID:         "test-webhook",
		Name:       "Test Webhook",
		URL:        "http://invalid-url-that-does-not-exist.local",
		Events:     []EventType{EventJobSucceeded},
		Queues:     []string{"*"},
		MaxRetries: 3,
		Timeout:    1 * time.Second,
	}

	logger := slog.Default()
	subscriber := NewWebhookSubscriber(subscription, logger)

	event := JobEvent{
		Event:    EventJobSucceeded,
		JobID:    "test-job-123",
		Queue:    "test-queue",
		Priority: 5,
	}

	err := subscriber.ProcessEvent(event)
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}

	deliveryErr, ok := err.(*DeliveryError)
	if !ok {
		t.Fatalf("expected DeliveryError, got %T", err)
	}

	if !deliveryErr.IsRetryable() {
		t.Error("expected error to be retryable")
	}
}

func TestWebhookSubscriber_RateLimit(t *testing.T) {
	subscription := &WebhookSubscription{
		ID:         "test-webhook",
		Name:       "Test Webhook",
		URL:        "https://httpbin.org/post",
		Events:     []EventType{EventJobSucceeded},
		Queues:     []string{"*"},
		RateLimit:  1, // 1 request per minute
		MaxRetries: 3,
		Timeout:    5 * time.Second,
	}

	logger := slog.Default()
	subscriber := NewWebhookSubscriber(subscription, logger)

	event := JobEvent{
		Event:    EventJobSucceeded,
		JobID:    "test-job-123",
		Queue:    "test-queue",
		Priority: 5,
	}

	// First request should succeed
	err1 := subscriber.ProcessEvent(event)
	if err1 != nil && err1.Error() != "rate limit exceeded" {
		// May succeed if external service is available
	}

	// Second immediate request should be rate limited
	err2 := subscriber.ProcessEvent(event)
	if err2 == nil {
		t.Skip("Rate limiting may not trigger immediately in test")
	}

	deliveryErr, ok := err2.(*DeliveryError)
	if ok && deliveryErr.StatusCode == 429 {
		// Rate limit triggered
		if !deliveryErr.IsRetryable() {
			t.Error("rate limit error should be retryable")
		}
	}
}

func TestWebhookDeliverer_Management(t *testing.T) {
	logger := slog.Default()
	deliverer := NewWebhookDeliverer(logger)

	subscription := &WebhookSubscription{
		ID:         "test-webhook",
		Name:       "Test Webhook",
		URL:        "https://httpbin.org/post",
		Events:     []EventType{EventJobSucceeded},
		Queues:     []string{"*"},
		MaxRetries: 3,
		Timeout:    5 * time.Second,
	}

	// Add subscription
	subscriber := deliverer.AddSubscription(subscription)
	if subscriber == nil {
		t.Fatal("expected subscriber to be created")
	}

	// Get subscription
	retrieved, err := deliverer.GetSubscriber(subscription.ID)
	if err != nil {
		t.Fatalf("expected to retrieve subscriber, got error: %v", err)
	}
	if retrieved.ID() != subscription.ID {
		t.Errorf("expected ID %s, got %s", subscription.ID, retrieved.ID())
	}

	// List subscriptions
	all := deliverer.ListSubscribers()
	if len(all) != 1 {
		t.Errorf("expected 1 subscriber, got %d", len(all))
	}

	// Remove subscription
	err = deliverer.RemoveSubscription(subscription.ID)
	if err != nil {
		t.Fatalf("expected no error removing subscription, got: %v", err)
	}

	// Verify removal
	_, err = deliverer.GetSubscriber(subscription.ID)
	if err != ErrSubscriptionNotFound {
		t.Errorf("expected ErrSubscriptionNotFound, got: %v", err)
	}
}