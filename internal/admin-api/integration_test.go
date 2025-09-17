// Copyright 2025 James Ross
//go:build integration
// +build integration

package adminapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	adminapi "github.com/flyingrobots/go-redis-work-queue/internal/admin-api"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Test fixtures
var (
	testJWTSecret = "test-secret-key-for-testing"
	testJWTToken  = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0QGV4YW1wbGUuY29tIiwicm9sZXMiOlsiYWRtaW4iXSwiZXhwIjo5OTk5OTk5OTk5LCJpYXQiOjE2MDk0NTkyMDB9.qg6LS-Y5frbTrGdZqvBhNXMQxgLqpLm1RqJvR_RfLpE"
)

type testSetup struct {
	server     *httptest.Server
	rdb        *redis.Client
	mr         *miniredis.Miniredis
	apiCfg     *adminapi.Config
	appCfg     *config.Config
	httpClient *http.Client
}

func setupIntegrationTest(t *testing.T) (*testSetup, func()) {
	// Create miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to create miniredis: %v", err)
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create configs
	appCfg := &config.Config{
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

	apiCfg := &adminapi.Config{
		JWTSecret:            testJWTSecret,
		RequireAuth:          false, // Disable for most tests
		DenyByDefault:        false,
		RateLimitEnabled:     true,
		RateLimitPerMinute:   1000,
		RateLimitBurst:       100,
		AuditEnabled:         true,
		AuditLogPath:         "/tmp/test-audit.log",
		RequireDoubleConfirm: true,
		ConfirmationPhrase:   "CONFIRM_DELETE",
	}

	// Create server
	logger := zap.NewNop()
	server, err := adminapi.NewServer(apiCfg, appCfg, rdb, logger)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Start test server
	mux := server.SetupRoutes()
	ts := httptest.NewServer(mux)

	setup := &testSetup{
		server:     ts,
		rdb:        rdb,
		mr:         mr,
		apiCfg:     apiCfg,
		appCfg:     appCfg,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	cleanup := func() {
		ts.Close()
		rdb.Close()
		mr.Close()
	}

	return setup, cleanup
}

func TestIntegrationStats(t *testing.T) {
	setup, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Add test data
	setup.mr.Lpush("jobqueue:high", "job1")
	setup.mr.Lpush("jobqueue:high", "job2")
	setup.mr.Lpush("jobqueue:high", "job3")
	setup.mr.Lpush("jobqueue:low", "job4")
	setup.mr.Lpush("jobqueue:low", "job5")
	setup.mr.Lpush("jobqueue:completed", "job6")
	setup.mr.Lpush("jobqueue:dead_letter", "job7")
	setup.mr.Lpush("jobqueue:dead_letter", "job8")

	// Request stats
	resp, err := setup.httpClient.Get(setup.server.URL + "/api/v1/stats")
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var stats adminapi.StatsResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify results
	if stats.Queues["high(jobqueue:high)"] != 3 {
		t.Errorf("Expected high queue to have 3 items, got %d", stats.Queues["high(jobqueue:high)"])
	}

	if stats.Queues["low(jobqueue:low)"] != 2 {
		t.Errorf("Expected low queue to have 2 items, got %d", stats.Queues["low(jobqueue:low)"])
	}

	if stats.Queues["completed(jobqueue:completed)"] != 1 {
		t.Errorf("Expected completed queue to have 1 item, got %d", stats.Queues["completed(jobqueue:completed)"])
	}

	if stats.Queues["dead_letter(jobqueue:dead_letter)"] != 2 {
		t.Errorf("Expected dead_letter queue to have 2 items, got %d", stats.Queues["dead_letter(jobqueue:dead_letter)"])
	}
}

func TestIntegrationStatsKeys(t *testing.T) {
	setup, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Add test data
	setup.mr.Lpush("jobqueue:high", "job1")
	setup.mr.Lpush("jobqueue:low", "job2")
	setup.mr.Set("jobqueue:rate_limit", "10")
	setup.mr.SetTTL("jobqueue:rate_limit", 60*time.Second)

	// Request stats/keys
	resp, err := setup.httpClient.Get(setup.server.URL + "/api/v1/stats/keys")
	if err != nil {
		t.Fatalf("Failed to get stats/keys: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var stats adminapi.StatsKeysResponse
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify results
	if stats.QueueLengths["high(jobqueue:high)"] != 1 {
		t.Errorf("Expected high queue to have 1 item, got %d", stats.QueueLengths["high(jobqueue:high)"])
	}

	if stats.RateLimitKey != "jobqueue:rate_limit" {
		t.Errorf("Expected rate limit key to be jobqueue:rate_limit, got %s", stats.RateLimitKey)
	}

	if stats.RateLimitTTL == "" {
		t.Error("Expected rate limit TTL to be set")
	}
}

func TestIntegrationPeek(t *testing.T) {
	setup, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Add test data
	jobs := []string{
		`{"id":"job1","filepath":"/test1.txt"}`,
		`{"id":"job2","filepath":"/test2.txt"}`,
		`{"id":"job3","filepath":"/test3.txt"}`,
	}
	for _, job := range jobs {
		setup.mr.Lpush("jobqueue:high", job)
	}

	// Test peek with count
	resp, err := setup.httpClient.Get(setup.server.URL + "/api/v1/queues/high/peek?count=2")
	if err != nil {
		t.Fatalf("Failed to peek queue: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var peek adminapi.PeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&peek); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify results
	if peek.Queue != "jobqueue:high" {
		t.Errorf("Expected queue to be jobqueue:high, got %s", peek.Queue)
	}

	if peek.Count != 2 {
		t.Errorf("Expected 2 items, got %d", peek.Count)
	}

	if len(peek.Items) != 2 {
		t.Errorf("Expected 2 items in array, got %d", len(peek.Items))
	}

	// Verify items are from the right end (last added)
	if !strings.Contains(peek.Items[0], "job3") {
		t.Errorf("Expected first item to be job3, got %s", peek.Items[0])
	}
}

func TestIntegrationPurgeDLQ(t *testing.T) {
	setup, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Add test data
	setup.mr.Lpush("jobqueue:dead_letter", "failed1")
	setup.mr.Lpush("jobqueue:dead_letter", "failed2")
	setup.mr.Lpush("jobqueue:dead_letter", "failed3")

	// Test with wrong confirmation
	wrongReq := adminapi.PurgeRequest{
		Confirmation: "WRONG",
		Reason:       "Test",
	}
	body, _ := json.Marshal(wrongReq)

	req, _ := http.NewRequest("DELETE", setup.server.URL+"/api/v1/queues/dlq", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := setup.httpClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to purge DLQ: %v", err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400 for wrong confirmation, got %d", resp.StatusCode)
	}

	// Test with correct confirmation
	correctReq := adminapi.PurgeRequest{
		Confirmation: "CONFIRM_DELETE",
		Reason:       "Integration test purge",
	}
	body, _ = json.Marshal(correctReq)

	req, _ = http.NewRequest("DELETE", setup.server.URL+"/api/v1/queues/dlq", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err = setup.httpClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to purge DLQ: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var purgeResp adminapi.PurgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&purgeResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify results
	if !purgeResp.Success {
		t.Error("Expected success to be true")
	}

	if purgeResp.ItemsDeleted != 3 {
		t.Errorf("Expected 3 items deleted, got %d", purgeResp.ItemsDeleted)
	}

	// Verify queue is empty
	if setup.mr.Exists("jobqueue:dead_letter") {
		t.Error("Dead letter queue should be deleted")
	}
}

func TestIntegrationPurgeAll(t *testing.T) {
	setup, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Add test data
	setup.mr.Lpush("jobqueue:high", "job1")
	setup.mr.Lpush("jobqueue:high", "job2")
	setup.mr.Lpush("jobqueue:low", "job3")
	setup.mr.Lpush("jobqueue:completed", "job4")
	setup.mr.Lpush("jobqueue:dead_letter", "job5")
	setup.mr.Set("jobqueue:rate_limit", "10")

	// Test with double confirmation
	req := adminapi.PurgeRequest{
		Confirmation: "CONFIRM_DELETE_ALL",
		Reason:       "Integration test full purge for testing",
	}
	body, _ := json.Marshal(req)

	httpReq, _ := http.NewRequest("DELETE", setup.server.URL+"/api/v1/queues/all", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := setup.httpClient.Do(httpReq)
	if err != nil {
		t.Fatalf("Failed to purge all: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp adminapi.ErrorResponse
		json.NewDecoder(resp.Body).Decode(&errResp)
		t.Errorf("Expected status 200, got %d: %s", resp.StatusCode, errResp.Error)
	}

	var purgeResp adminapi.PurgeResponse
	if err := json.NewDecoder(resp.Body).Decode(&purgeResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify results
	if !purgeResp.Success {
		t.Error("Expected success to be true")
	}

	if purgeResp.ItemsDeleted < 5 {
		t.Errorf("Expected at least 5 keys deleted, got %d", purgeResp.ItemsDeleted)
	}

	// Verify queues are empty
	for _, key := range []string{"jobqueue:high", "jobqueue:low", "jobqueue:completed", "jobqueue:dead_letter"} {
		if setup.mr.Exists(key) {
			t.Errorf("Queue %s should be deleted", key)
		}
	}
}

func TestIntegrationBenchmark(t *testing.T) {
	setup, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Run benchmark
	benchReq := adminapi.BenchRequest{
		Count:       50,
		Priority:    "high",
		Rate:        100,
		Timeout:     10,
		PayloadSize: 512,
	}
	body, _ := json.Marshal(benchReq)

	req, _ := http.NewRequest("POST", setup.server.URL+"/api/v1/bench", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := setup.httpClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to run benchmark: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var benchResp adminapi.BenchResponse
	if err := json.NewDecoder(resp.Body).Decode(&benchResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify results
	if benchResp.Count != 50 {
		t.Errorf("Expected count 50, got %d", benchResp.Count)
	}

	if benchResp.Duration == 0 {
		t.Error("Expected non-zero duration")
	}

	// Check that jobs were enqueued
	queueLen, _ := setup.rdb.LLen(context.Background(), "jobqueue:high").Result()
	if queueLen == 0 {
		t.Error("Expected jobs to be enqueued")
	}
}

func TestIntegrationRateLimiting(t *testing.T) {
	setup, cleanup := setupIntegrationTest(t)
	defer cleanup()

	// Configure strict rate limiting
	setup.apiCfg.RateLimitPerMinute = 5
	setup.apiCfg.RateLimitBurst = 2

	// Make burst requests
	successCount := 0
	rateLimitCount := 0

	for i := 0; i < 5; i++ {
		resp, err := setup.httpClient.Get(setup.server.URL + "/api/v1/stats")
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			successCount++
		} else if resp.StatusCode == http.StatusTooManyRequests {
			rateLimitCount++

			// Check rate limit headers
			if resp.Header.Get("X-RateLimit-Limit") == "" {
				t.Error("Missing X-RateLimit-Limit header")
			}
			if resp.Header.Get("X-RateLimit-Remaining") == "" {
				t.Error("Missing X-RateLimit-Remaining header")
			}
			if resp.Header.Get("X-RateLimit-Reset") == "" {
				t.Error("Missing X-RateLimit-Reset header")
			}
		}
	}

	// Should have some successful and some rate limited
	if successCount == 0 {
		t.Error("Expected some requests to succeed")
	}

	if rateLimitCount == 0 {
		t.Error("Expected some requests to be rate limited")
	}
}

func TestIntegrationHealthCheck(t *testing.T) {
	setup, cleanup := setupIntegrationTest(t)
	defer cleanup()

	resp, err := setup.httpClient.Get(setup.server.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to get health: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var health map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if health["status"] != "healthy" {
		t.Errorf("Expected status healthy, got %s", health["status"])
	}
}

func TestIntegrationOpenAPISpec(t *testing.T) {
	setup, cleanup := setupIntegrationTest(t)
	defer cleanup()

	resp, err := setup.httpClient.Get(setup.server.URL + "/api/v1/openapi.yaml")
	if err != nil {
		t.Fatalf("Failed to get OpenAPI spec: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check content type
	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/x-yaml" {
		t.Errorf("Expected Content-Type application/x-yaml, got %s", contentType)
	}

	// Read body
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	spec := buf.String()

	// Verify it contains OpenAPI content
	if !strings.Contains(spec, "openapi: 3.0.3") {
		t.Error("Response does not contain OpenAPI version")
	}

	if !strings.Contains(spec, "title: Redis Work Queue Admin API") {
		t.Error("Response does not contain API title")
	}

	// Verify required endpoints are documented
	requiredEndpoints := []string{
		"/stats",
		"/stats/keys",
		"/queues/{queue}/peek",
		"/queues/dlq",
		"/queues/all",
		"/bench",
	}

	for _, endpoint := range requiredEndpoints {
		if !strings.Contains(spec, endpoint) {
			t.Errorf("OpenAPI spec missing endpoint: %s", endpoint)
		}
	}
}

func TestIntegrationValidationErrors(t *testing.T) {
	setup, cleanup := setupIntegrationTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "Invalid peek count",
			method:         "GET",
			path:           "/api/v1/queues/high/peek?count=200",
			expectedStatus: http.StatusOK, // Count is clamped, not an error
		},
		{
			name:   "Missing confirmation",
			method: "DELETE",
			path:   "/api/v1/queues/dlq",
			body: adminapi.PurgeRequest{
				Reason: "Test",
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "CONFIRMATION_FAILED",
		},
		{
			name:   "Short reason",
			method: "DELETE",
			path:   "/api/v1/queues/dlq",
			body: adminapi.PurgeRequest{
				Confirmation: "CONFIRM_DELETE",
				Reason:       "X",
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "REASON_REQUIRED",
		},
		{
			name:   "Invalid benchmark count",
			method: "POST",
			path:   "/api/v1/bench",
			body: adminapi.BenchRequest{
				Count:    -1,
				Priority: "high",
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_COUNT",
		},
		{
			name:   "Invalid benchmark priority",
			method: "POST",
			path:   "/api/v1/bench",
			body: adminapi.BenchRequest{
				Count:    10,
				Priority: "invalid",
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_PRIORITY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var reqBody []byte
			if tt.body != nil {
				reqBody, _ = json.Marshal(tt.body)
			}

			req, _ := http.NewRequest(tt.method, setup.server.URL+tt.path, bytes.NewReader(reqBody))
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			resp, err := setup.httpClient.Do(req)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedStatus {
				var errResp adminapi.ErrorResponse
				json.NewDecoder(resp.Body).Decode(&errResp)
				t.Errorf("Expected status %d, got %d: %s", tt.expectedStatus, resp.StatusCode, errResp.Error)
			}

			if tt.expectedCode != "" && resp.StatusCode >= 400 {
				var errResp adminapi.ErrorResponse
				if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
					t.Fatalf("Failed to decode error response: %v", err)
				}

				if errResp.Code != tt.expectedCode {
					t.Errorf("Expected error code %s, got %s", tt.expectedCode, errResp.Code)
				}
			}
		})
	}
}
