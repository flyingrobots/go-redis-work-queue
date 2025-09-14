// Copyright 2025 James Ross
package adminapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func setupTestHandler(t *testing.T) (*Handler, *miniredis.Miniredis, func()) {
	// Create mini redis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create config
	cfg := &config.Config{
		Worker: config.Worker{
			Queues: map[string]string{
				"high": "jobqueue:high",
				"low":  "jobqueue:low",
			},
			CompletedList:  "jobqueue:completed",
			DeadLetterList: "jobqueue:dead_letter",
		},
		Producer: config.Producer{
			RateLimitKey: "jobqueue:rate_limit",
		},
	}

	apiCfg := &Config{
		ConfirmationPhrase: "CONFIRM_DELETE",
	}

	logger := zap.NewNop()

	handler := NewHandler(cfg, apiCfg, rdb, logger, nil)

	cleanup := func() {
		rdb.Close()
		mr.Close()
	}

	return handler, mr, cleanup
}

func TestGetStats(t *testing.T) {
	handler, mr, cleanup := setupTestHandler(t)
	defer cleanup()

	// Add test data
	mr.Lpush("jobqueue:high", "job1")
	mr.Lpush("jobqueue:high", "job2")
	mr.Lpush("jobqueue:low", "job3")
	mr.Lpush("jobqueue:completed", "job4")

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/stats", nil)
	w := httptest.NewRecorder()

	// Execute handler
	handler.GetStats(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp StatsResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify queue counts
	if resp.Queues["high(jobqueue:high)"] != 2 {
		t.Errorf("Expected high queue to have 2 items, got %d", resp.Queues["high(jobqueue:high)"])
	}

	if resp.Queues["low(jobqueue:low)"] != 1 {
		t.Errorf("Expected low queue to have 1 item, got %d", resp.Queues["low(jobqueue:low)"])
	}
}

func TestPeekQueue(t *testing.T) {
	handler, mr, cleanup := setupTestHandler(t)
	defer cleanup()

	// Add test data
	mr.Lpush("jobqueue:high", "job1")
	mr.Lpush("jobqueue:high", "job2")
	mr.Lpush("jobqueue:high", "job3")

	// Create request
	req := httptest.NewRequest("GET", "/api/v1/queues/high/peek?count=2", nil)
	w := httptest.NewRecorder()

	// Execute handler
	handler.PeekQueue(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp PeekResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify peek results
	if resp.Queue != "jobqueue:high" {
		t.Errorf("Expected queue 'jobqueue:high', got %s", resp.Queue)
	}

	if resp.Count != 2 {
		t.Errorf("Expected 2 items, got %d", resp.Count)
	}

	if len(resp.Items) != 2 {
		t.Errorf("Expected 2 items in array, got %d", len(resp.Items))
	}
}

func TestPurgeDLQ(t *testing.T) {
	handler, mr, cleanup := setupTestHandler(t)
	defer cleanup()

	// Add test data
	mr.Lpush("jobqueue:dead_letter", "failed1")
	mr.Lpush("jobqueue:dead_letter", "failed2")

	// Create request with proper confirmation
	reqBody := PurgeRequest{
		Confirmation: "CONFIRM_DELETE",
		Reason:       "Test purge operation",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("DELETE", "/api/v1/queues/dlq", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute handler
	handler.PurgeDLQ(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp PurgeResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify purge results
	if !resp.Success {
		t.Error("Expected success to be true")
	}

	if resp.ItemsDeleted != 2 {
		t.Errorf("Expected 2 items deleted, got %d", resp.ItemsDeleted)
	}

	// Verify queue is empty
	if mr.Exists("jobqueue:dead_letter") {
		t.Error("Dead letter queue should be deleted")
	}
}

func TestPurgeDLQInvalidConfirmation(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create request with wrong confirmation
	reqBody := PurgeRequest{
		Confirmation: "WRONG_PHRASE",
		Reason:       "Test",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("DELETE", "/api/v1/queues/dlq", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute handler
	handler.PurgeDLQ(w, req)

	// Should fail with bad request
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var resp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Code != "CONFIRMATION_FAILED" {
		t.Errorf("Expected error code CONFIRMATION_FAILED, got %s", resp.Code)
	}
}

func TestBenchmark(t *testing.T) {
	handler, _, cleanup := setupTestHandler(t)
	defer cleanup()

	// Create request
	reqBody := BenchRequest{
		Count:    10,
		Priority: "high",
		Rate:     100,
		Timeout:  5,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/v1/bench", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Execute handler
	handler.RunBenchmark(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp BenchResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify benchmark results
	if resp.Count != 10 {
		t.Errorf("Expected count 10, got %d", resp.Count)
	}

	if resp.Duration == 0 {
		t.Error("Expected non-zero duration")
	}
}

func TestRateLimiting(t *testing.T) {
	bucket := &rateBucket{
		tokens:    3,
		lastFill:  time.Now(),
		maxTokens: 3,
		fillRate:  1.0,
	}

	// Should allow first 3 requests
	for i := 0; i < 3; i++ {
		if !bucket.consume() {
			t.Errorf("Request %d should have been allowed", i+1)
		}
	}

	// 4th request should be denied
	if bucket.consume() {
		t.Error("4th request should have been denied")
	}

	// Wait for refill
	time.Sleep(2 * time.Second)

	// Should allow again after refill
	if !bucket.consume() {
		t.Error("Request should be allowed after refill")
	}
}

func TestJWTValidation(t *testing.T) {
	secret := "test-secret"

	tests := []struct {
		name        string
		token       string
		shouldError bool
	}{
		{
			name:        "Invalid format",
			token:       "invalid",
			shouldError: true,
		},
		{
			name:        "Missing parts",
			token:       "header.payload",
			shouldError: true,
		},
		{
			name:        "Invalid base64",
			token:       "invalid!.base64!.here!",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validateJWT(tt.token, secret)
			if tt.shouldError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}