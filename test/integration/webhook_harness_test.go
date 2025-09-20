//go:build integration_tests
// +build integration_tests

// Copyright 2025 James Ross
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// WebhookHarness provides a test HTTP server for webhook testing
type WebhookHarness struct {
	server   *httptest.Server
	handler  *WebhookTestHandler
	requests []WebhookRequest
	mu       sync.RWMutex
}

// WebhookTestHandler handles webhook test requests
type WebhookTestHandler struct {
	responses     map[string]WebhookResponse
	delays        map[string]time.Duration
	callbackCount map[string]int
	harness       *WebhookHarness
	customHandler func(w http.ResponseWriter, r *http.Request, defaultHandler func())
	mu            sync.RWMutex
}

// WebhookRequest captures details of received webhook requests
type WebhookRequest struct {
	Method    string
	URL       string
	Headers   http.Header
	Body      []byte
	Timestamp time.Time
	UserAgent string
	Signature string
}

// WebhookResponse defines how the test server should respond
type WebhookResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       string
	Delay      time.Duration
}

// WebhookDeliveryService simulates the webhook delivery system
type WebhookDeliveryService struct {
	client      *http.Client
	signer      *HMACSignatureService
	retryPolicy RetryPolicy
}

// HMACSignatureService handles webhook signature generation
type HMACSignatureService struct{}

// RetryPolicy defines retry behavior for webhook delivery
type RetryPolicy struct {
	MaxRetries   int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       bool
}

// WebhookEvent represents an event to be delivered via webhook
type WebhookEvent struct {
	ID        string                 `json:"id"`
	Event     string                 `json:"event"`
	Timestamp time.Time              `json:"timestamp"`
	JobID     string                 `json:"job_id"`
	Queue     string                 `json:"queue"`
	Priority  int                    `json:"priority"`
	Data      map[string]interface{} `json:"data,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
}

// NewWebhookHarness creates a new webhook test harness
func NewWebhookHarness() *WebhookHarness {
	handler := &WebhookTestHandler{
		responses:     make(map[string]WebhookResponse),
		delays:        make(map[string]time.Duration),
		callbackCount: make(map[string]int),
	}

	harness := &WebhookHarness{
		handler: handler,
	}

	handler.harness = harness

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler.ServeHTTP)

	harness.server = httptest.NewServer(mux)

	return harness
}

// Close shuts down the test server
func (h *WebhookHarness) Close() {
	h.server.Close()
}

// URL returns the base URL of the test server
func (h *WebhookHarness) URL() string {
	return h.server.URL
}

// SetResponse configures how the server responds to a specific path
func (h *WebhookHarness) SetResponse(path string, response WebhookResponse) {
	h.handler.mu.Lock()
	defer h.handler.mu.Unlock()
	h.handler.responses[path] = response
}

// SetDelay configures a delay for a specific path
func (h *WebhookHarness) SetDelay(path string, delay time.Duration) {
	h.handler.mu.Lock()
	defer h.handler.mu.Unlock()
	h.handler.delays[path] = delay
}

// SetCustomHandler sets a custom request handler
func (h *WebhookHarness) SetCustomHandler(handler func(w http.ResponseWriter, r *http.Request, defaultHandler func())) {
	h.handler.mu.Lock()
	defer h.handler.mu.Unlock()
	h.handler.customHandler = handler
}

// GetRequests returns all received webhook requests
func (h *WebhookHarness) GetRequests() []WebhookRequest {
	h.mu.RLock()
	defer h.mu.RUnlock()
	requests := make([]WebhookRequest, len(h.requests))
	copy(requests, h.requests)
	return requests
}

// GetRequestCount returns the number of requests received for a path
func (h *WebhookHarness) GetRequestCount(path string) int {
	h.handler.mu.RLock()
	defer h.handler.mu.RUnlock()
	return h.handler.callbackCount[path]
}

// ClearRequests clears the request history
func (h *WebhookHarness) ClearRequests() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.requests = nil

	h.handler.mu.Lock()
	defer h.handler.mu.Unlock()
	h.handler.callbackCount = make(map[string]int)
}

// ServeHTTP handles incoming webhook requests
func (h *WebhookTestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	// Record the request
	request := WebhookRequest{
		Method:    r.Method,
		URL:       r.URL.String(),
		Headers:   r.Header.Clone(),
		Body:      body,
		Timestamp: time.Now(),
		UserAgent: r.UserAgent(),
		Signature: r.Header.Get("X-Webhook-Signature"),
	}

	h.harness.mu.Lock()
	h.harness.requests = append(h.harness.requests, request)
	h.harness.mu.Unlock()

	// Update callback count
	h.mu.Lock()
	h.callbackCount[r.URL.Path]++
	h.mu.Unlock()

	// Check for custom handler
	h.mu.RLock()
	customHandler := h.customHandler
	h.mu.RUnlock()

	defaultHandler := func() {
		// Apply configured delay
		if delay, exists := h.delays[r.URL.Path]; exists {
			time.Sleep(delay)
		}

		// Send configured response
		h.mu.RLock()
		response, exists := h.responses[r.URL.Path]
		h.mu.RUnlock()

		if !exists {
			response = WebhookResponse{
				StatusCode: http.StatusOK,
				Body:       `{"status": "received"}`,
			}
		}

		// Set response headers
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}

		w.WriteHeader(response.StatusCode)
		if response.Body != "" {
			w.Write([]byte(response.Body))
		}
	}

	if customHandler != nil {
		customHandler(w, r, defaultHandler)
	} else {
		defaultHandler()
	}
}

// NewHMACSignatureService creates a new signature service
func NewHMACSignatureService() *HMACSignatureService {
	return &HMACSignatureService{}
}

// SignPayload generates HMAC signature for webhook payload
func (s *HMACSignatureService) SignPayload(payload []byte, secret string) string {
	// Implementation would use actual HMAC signing
	// For testing, we'll use a simplified approach
	return fmt.Sprintf("sha256=%x", len(payload)+len(secret))
}

// VerifySignature validates HMAC signature
func (s *HMACSignatureService) VerifySignature(payload []byte, signature, secret string) bool {
	expected := s.SignPayload(payload, secret)
	return signature == expected
}

// NewWebhookDeliveryService creates a new webhook delivery service
func NewWebhookDeliveryService(retryPolicy RetryPolicy) *WebhookDeliveryService {
	return &WebhookDeliveryService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		signer:      NewHMACSignatureService(),
		retryPolicy: retryPolicy,
	}
}

// DeliverWebhook delivers a webhook event with retries
func (s *WebhookDeliveryService) DeliverWebhook(ctx context.Context, url, secret string, event WebhookEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= s.retryPolicy.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := s.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payload))
		if err != nil {
			lastErr = err
			continue
		}

		// Add signature header
		signature := s.signer.SignPayload(payload, secret)
		req.Header.Set("X-Webhook-Signature", signature)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "WebhookDeliveryService/1.0")

		resp, err := s.client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		resp.Body.Close()

		// Success for 2xx status codes
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		// For 4xx errors (except 429), don't retry
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != 429 {
			return fmt.Errorf("webhook delivery failed with status %d", resp.StatusCode)
		}

		lastErr = fmt.Errorf("webhook delivery failed with status %d", resp.StatusCode)
	}

	return fmt.Errorf("webhook delivery failed after %d attempts: %w", s.retryPolicy.MaxRetries+1, lastErr)
}

// calculateDelay calculates retry delay with exponential backoff
func (s *WebhookDeliveryService) calculateDelay(attempt int) time.Duration {
	delay := s.retryPolicy.InitialDelay
	for i := 1; i < attempt; i++ {
		delay = time.Duration(float64(delay) * s.retryPolicy.Multiplier)
		if delay > s.retryPolicy.MaxDelay {
			delay = s.retryPolicy.MaxDelay
			break
		}
	}

	// Add jitter if enabled
	if s.retryPolicy.Jitter && delay > 0 {
		jitterAmount := time.Duration(float64(delay) * 0.1) // 10% jitter
		delay += jitterAmount
	}

	return delay
}

// Integration Tests

func TestWebhookHarness_BasicDelivery(t *testing.T) {
	harness := NewWebhookHarness()
	defer harness.Close()

	// Configure successful response
	harness.SetResponse("/webhook", WebhookResponse{
		StatusCode: http.StatusOK,
		Body:       `{"status": "success"}`,
	})

	// Create delivery service
	retryPolicy := RetryPolicy{
		MaxRetries:   3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}
	service := NewWebhookDeliveryService(retryPolicy)

	// Create test event
	event := WebhookEvent{
		ID:        "evt_123",
		Event:     "job_failed",
		Timestamp: time.Now(),
		JobID:     "job_456",
		Queue:     "test_queue",
		Priority:  5,
		Data:      map[string]interface{}{"error": "Connection timeout"},
		TraceID:   "trace_789",
	}

	// Deliver webhook
	ctx := context.Background()
	webhookURL := harness.URL() + "/webhook"
	secret := "test_secret_key"

	err := service.DeliverWebhook(ctx, webhookURL, secret, event)
	assert.NoError(t, err)

	// Verify request was received
	requests := harness.GetRequests()
	require.Len(t, requests, 1)

	request := requests[0]
	assert.Equal(t, "POST", request.Method)
	assert.Equal(t, "/webhook", request.URL)
	assert.Equal(t, "application/json", request.Headers.Get("Content-Type"))
	assert.Equal(t, "WebhookDeliveryService/1.0", request.Headers.Get("User-Agent"))
	assert.NotEmpty(t, request.Headers.Get("X-Webhook-Signature"))

	// Verify payload
	var receivedEvent WebhookEvent
	err = json.Unmarshal(request.Body, &receivedEvent)
	assert.NoError(t, err)
	assert.Equal(t, event.ID, receivedEvent.ID)
	assert.Equal(t, event.Event, receivedEvent.Event)
	assert.Equal(t, event.JobID, receivedEvent.JobID)
}

func TestWebhookHarness_RetryOnFailure(t *testing.T) {
	harness := NewWebhookHarness()
	defer harness.Close()

	// Configure to fail first 2 attempts, then succeed
	attemptCount := 0
	harness.SetCustomHandler(func(w http.ResponseWriter, r *http.Request, defaultHandler func()) {
		attemptCount++
		if attemptCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "temporary failure"}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "success"}`))
		}
	})

	retryPolicy := RetryPolicy{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond, // Fast for testing
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}
	service := NewWebhookDeliveryService(retryPolicy)

	event := WebhookEvent{
		ID:    "evt_retry_test",
		Event: "job_failed",
		JobID: "job_retry",
	}

	ctx := context.Background()
	webhookURL := harness.URL() + "/webhook"
	secret := "test_secret"

	// Should eventually succeed after retries
	err := service.DeliverWebhook(ctx, webhookURL, secret, event)
	assert.NoError(t, err)

	// Verify 3 attempts were made
	requests := harness.GetRequests()
	assert.Len(t, requests, 3)
}

func TestWebhookHarness_NonRetriableError(t *testing.T) {
	harness := NewWebhookHarness()
	defer harness.Close()

	// Configure 400 Bad Request (non-retriable)
	harness.SetResponse("/webhook", WebhookResponse{
		StatusCode: http.StatusBadRequest,
		Body:       `{"error": "invalid payload"}`,
	})

	retryPolicy := RetryPolicy{
		MaxRetries:   3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}
	service := NewWebhookDeliveryService(retryPolicy)

	event := WebhookEvent{
		ID:    "evt_bad_request",
		Event: "job_failed",
		JobID: "job_bad",
	}

	ctx := context.Background()
	webhookURL := harness.URL() + "/webhook"
	secret := "test_secret"

	// Should fail immediately without retries
	err := service.DeliverWebhook(ctx, webhookURL, secret, event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook delivery failed with status 400")

	// Verify only 1 attempt was made
	requests := harness.GetRequests()
	assert.Len(t, requests, 1)
}

func TestWebhookHarness_Timeout(t *testing.T) {
	harness := NewWebhookHarness()
	defer harness.Close()

	// Configure delay longer than timeout
	harness.SetDelay("/webhook", 2*time.Second)
	harness.SetResponse("/webhook", WebhookResponse{
		StatusCode: http.StatusOK,
		Body:       `{"status": "success"}`,
	})

	retryPolicy := RetryPolicy{
		MaxRetries:   1,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}
	service := NewWebhookDeliveryService(retryPolicy)

	// Override client timeout
	service.client.Timeout = 500 * time.Millisecond

	event := WebhookEvent{
		ID:    "evt_timeout",
		Event: "job_failed",
		JobID: "job_timeout",
	}

	ctx := context.Background()
	webhookURL := harness.URL() + "/webhook"
	secret := "test_secret"

	// Should fail due to timeout
	err := service.DeliverWebhook(ctx, webhookURL, secret, event)
	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "timeout")
}

func TestWebhookHarness_ConcurrentDeliveries(t *testing.T) {
	harness := NewWebhookHarness()
	defer harness.Close()

	// Configure successful response
	harness.SetResponse("/webhook", WebhookResponse{
		StatusCode: http.StatusOK,
		Body:       `{"status": "success"}`,
	})

	retryPolicy := RetryPolicy{
		MaxRetries:   1,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
		Multiplier:   2.0,
	}
	service := NewWebhookDeliveryService(retryPolicy)

	// Send multiple webhooks concurrently
	const numWebhooks = 10
	var wg sync.WaitGroup
	errors := make([]error, numWebhooks)

	for i := 0; i < numWebhooks; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			event := WebhookEvent{
				ID:    fmt.Sprintf("evt_%d", index),
				Event: "job_completed",
				JobID: fmt.Sprintf("job_%d", index),
			}

			ctx := context.Background()
			webhookURL := harness.URL() + "/webhook"
			secret := "test_secret"

			errors[index] = service.DeliverWebhook(ctx, webhookURL, secret, event)
		}(i)
	}

	wg.Wait()

	// All deliveries should succeed
	for i, err := range errors {
		assert.NoError(t, err, "Delivery %d should succeed", i)
	}

	// Verify all requests were received
	requests := harness.GetRequests()
	assert.Len(t, requests, numWebhooks)

	// Verify unique event IDs
	eventIDs := make(map[string]bool)
	for _, request := range requests {
		var event WebhookEvent
		err := json.Unmarshal(request.Body, &event)
		require.NoError(t, err)
		assert.False(t, eventIDs[event.ID], "Event ID %s should be unique", event.ID)
		eventIDs[event.ID] = true
	}
}

func TestWebhookHarness_SignatureValidation(t *testing.T) {
	harness := NewWebhookHarness()
	defer harness.Close()

	// Configure to validate signatures
	secret := "test_secret_key"
	signer := NewHMACSignatureService()

	harness.SetResponse("/webhook", WebhookResponse{
		StatusCode: http.StatusOK,
		Body:       `{"status": "success"}`,
	})

	retryPolicy := RetryPolicy{MaxRetries: 1, InitialDelay: 10 * time.Millisecond}
	service := NewWebhookDeliveryService(retryPolicy)

	event := WebhookEvent{
		ID:    "evt_signature_test",
		Event: "job_failed",
		JobID: "job_signature",
		Data:  map[string]interface{}{"test": true},
	}

	ctx := context.Background()
	webhookURL := harness.URL() + "/webhook"

	err := service.DeliverWebhook(ctx, webhookURL, secret, event)
	assert.NoError(t, err)

	requests := harness.GetRequests()
	require.Len(t, requests, 1)

	request := requests[0]
	signature := request.Headers.Get("X-Webhook-Signature")
	assert.NotEmpty(t, signature)

	// Verify signature is correct
	isValid := signer.VerifySignature(request.Body, signature, secret)
	assert.True(t, isValid, "Signature should be valid")
}

func TestWebhookHarness_RequestHeaders(t *testing.T) {
	harness := NewWebhookHarness()
	defer harness.Close()

	harness.SetResponse("/webhook", WebhookResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"X-Response-Time": "123ms",
			"X-Server-ID":     "test-server-1",
		},
		Body: `{"status": "processed"}`,
	})

	retryPolicy := RetryPolicy{MaxRetries: 1, InitialDelay: 10 * time.Millisecond}
	service := NewWebhookDeliveryService(retryPolicy)

	event := WebhookEvent{
		ID:    "evt_headers_test",
		Event: "job_completed",
		JobID: "job_headers",
	}

	ctx := context.Background()
	webhookURL := harness.URL() + "/webhook"
	secret := "test_secret"

	err := service.DeliverWebhook(ctx, webhookURL, secret, event)
	assert.NoError(t, err)

	requests := harness.GetRequests()
	require.Len(t, requests, 1)

	request := requests[0]

	// Verify standard headers
	assert.Equal(t, "application/json", request.Headers.Get("Content-Type"))
	assert.Equal(t, "WebhookDeliveryService/1.0", request.Headers.Get("User-Agent"))
	assert.NotEmpty(t, request.Headers.Get("X-Webhook-Signature"))

	// Verify request was properly formatted
	assert.Equal(t, "POST", request.Method)
	assert.Equal(t, "/webhook", request.URL)
	assert.True(t, len(request.Body) > 0)
}

// Benchmark Tests

func BenchmarkWebhookHarness_SingleDelivery(b *testing.B) {
	harness := NewWebhookHarness()
	defer harness.Close()

	harness.SetResponse("/webhook", WebhookResponse{
		StatusCode: http.StatusOK,
		Body:       `{"status": "success"}`,
	})

	retryPolicy := RetryPolicy{
		MaxRetries:   1,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
	}
	service := NewWebhookDeliveryService(retryPolicy)

	event := WebhookEvent{
		ID:    "benchmark_event",
		Event: "job_completed",
		JobID: "benchmark_job",
		Data:  map[string]interface{}{"benchmark": true},
	}

	ctx := context.Background()
	webhookURL := harness.URL() + "/webhook"
	secret := "benchmark_secret"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		harness.ClearRequests()
		service.DeliverWebhook(ctx, webhookURL, secret, event)
	}
}
