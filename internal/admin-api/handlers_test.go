// Copyright 2025 James Ross
package adminapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type countingReader struct {
	reader io.Reader
	read   int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.read += int64(n)
	return n, err
}

func setupHandlerTest(t *testing.T) (*Handler, *miniredis.Miniredis, func()) {
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

func TestHandlerGetStats(t *testing.T) {
	handler, mr, cleanup := setupHandlerTest(t)
	defer cleanup()

	// Add test data
	mr.Lpush("jobqueue:high", "job1")
	mr.Lpush("jobqueue:high", "job2")
	mr.Lpush("jobqueue:low", "job3")
	mr.Lpush("jobqueue:completed", "job4")
	handler.cfg.Queue.OrderedQueuePattern = queuekeys.DefaultOrderedQueuePattern
	orderedQueue := queuekeys.Format(handler.cfg.Queue.OrderedQueuePattern, queuekeys.OrderingDigest("account:a"))
	mr.Lpush(orderedQueue, "ordered-1")
	mr.Lpush(orderedQueue, "ordered-2")

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
	if resp.OrderedPending != 2 {
		t.Errorf("Expected 2 ordered jobs pending, got %d", resp.OrderedPending)
	}
}

func TestEnqueueJobReturnsCreatedIDAndPreservesEnvelope(t *testing.T) {
	handler, mr, cleanup := setupHandlerTest(t)
	defer cleanup()
	handler.cfg.Queue.MaxPayloadSize = 64

	reqBody := EnqueueRequest{
		ID:            "http-job",
		Payload:       []byte(`{"x":1}`),
		PayloadSchema: "demo.v1",
		Priority:      "high",
		OrderingKey:   "account:42",
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/enqueue", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.EnqueueJob(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", w.Code, w.Body.String())
	}
	var response EnqueueResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.ID != "http-job" {
		t.Fatalf("response ID = %q", response.ID)
	}
	orderedQueue := queuekeys.Format(queuekeys.DefaultOrderedQueuePattern, queuekeys.OrderingDigest(reqBody.OrderingKey))
	items, err := mr.List(orderedQueue)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("ordered queue length = %d, want 1", len(items))
	}
	ready, err := mr.List(queuekeys.DefaultOrderedReadyList)
	if err != nil || len(ready) != 1 {
		t.Fatalf("ready tokens = %v (err=%v), want one", ready, err)
	}
	job, err := queue.UnmarshalJob(items[0])
	if err != nil {
		t.Fatal(err)
	}
	if job.ID != response.ID || !bytes.Equal(job.Payload, reqBody.Payload) || job.PayloadSchema != reqBody.PayloadSchema || job.OrderingKey != reqBody.OrderingKey {
		t.Fatalf("queued envelope changed: %#v", job)
	}
}

func TestEnqueueJobRejectsTrailingJSONWithoutQueueChange(t *testing.T) {
	handler, mr, cleanup := setupHandlerTest(t)
	defer cleanup()
	body := `{"payload":"YQ==","priority":"low"}{"payload":"Yg==","priority":"low"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/enqueue", strings.NewReader(body))
	w := httptest.NewRecorder()

	handler.EnqueueJob(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", w.Code, w.Body.String())
	}
	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Code != "INVALID_REQUEST" {
		t.Fatalf("error code = %q, want INVALID_REQUEST", response.Code)
	}
	if mr.Exists("jobqueue:low") {
		t.Fatal("request with trailing JSON enqueued the first document")
	}
}

func TestEnqueueJobOversizedPayloadReturnsTypedMessageWithoutQueueChange(t *testing.T) {
	handler, mr, cleanup := setupHandlerTest(t)
	defer cleanup()
	handler.cfg.Queue.MaxPayloadSize = 4
	reqBody := EnqueueRequest{Payload: []byte("12345"), Priority: "low"}
	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/enqueue", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.EnqueueJob(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413: %s", w.Code, w.Body.String())
	}
	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Code != "PAYLOAD_TOO_LARGE" || response.Error != "payload is 5 bytes; maximum is 4 bytes" {
		t.Fatalf("unexpected error response: %#v", response)
	}
	if mr.Exists("jobqueue:low") {
		items, listErr := mr.List("jobqueue:low")
		t.Fatalf("rejected enqueue changed queue: %v (err=%v)", items, listErr)
	}
}

func TestEnqueueJobBoundsEncodedBodyBeforeDecoding(t *testing.T) {
	handler, _, cleanup := setupHandlerTest(t)
	defer cleanup()
	handler.cfg.Queue.MaxPayloadSize = 4

	body := `{"payload":"` + strings.Repeat("A", 2<<20) + `"}`
	reader := &countingReader{reader: strings.NewReader(body)}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/enqueue", reader)
	w := httptest.NewRecorder()

	handler.EnqueueJob(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413: %s", w.Code, w.Body.String())
	}
	var response ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}
	if response.Code != "PAYLOAD_TOO_LARGE" {
		t.Fatalf("error code = %q, want PAYLOAD_TOO_LARGE", response.Code)
	}
	if reader.read > 70<<10 {
		t.Fatalf("handler read %d bytes before rejecting an oversized encoded body", reader.read)
	}
}

func TestSetupRoutesExposesEnqueueEndpoint(t *testing.T) {
	handler, _, cleanup := setupHandlerTest(t)
	defer cleanup()
	server := &Server{
		cfg:    handler.apiCfg,
		appCfg: handler.cfg,
		rdb:    handler.rdb,
		logger: handler.logger,
	}
	body, err := json.Marshal(EnqueueRequest{Payload: []byte("route"), Priority: "low"})
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/enqueue", bytes.NewReader(body))
	w := httptest.NewRecorder()

	server.SetupRoutes().ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201: %s", w.Code, w.Body.String())
	}
}

func TestDLQHandlersRoundTripSelectionHandles(t *testing.T) {
	handler, mr, cleanup := setupHandlerTest(t)
	defer cleanup()
	first := `{"id":"shared-id","payload":"Zmlyc3Q="}`
	second := `{"id":"shared-id","payload":"c2Vjb25k"}`
	mr.RPush(handler.cfg.Worker.DeadLetterList, first, second)

	listRequest := httptest.NewRequest(http.MethodGet, "/api/v1/dlq", nil)
	listResponse := httptest.NewRecorder()
	handler.ListDLQ(listResponse, listRequest)
	if listResponse.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200: %s", listResponse.Code, listResponse.Body.String())
	}
	var listed DLQListResponse
	if err := json.NewDecoder(listResponse.Body).Decode(&listed); err != nil {
		t.Fatal(err)
	}
	if len(listed.Items) != 2 || listed.Items[0].Handle == "" || listed.Items[0].Handle == listed.Items[1].Handle {
		t.Fatalf("listed items = %#v, want two distinct selection handles", listed.Items)
	}

	body, err := json.Marshal(DLQPurgeSelectionRequest{Handles: []string{listed.Items[1].Handle}})
	if err != nil {
		t.Fatal(err)
	}
	purgeRequest := httptest.NewRequest(http.MethodPost, "/api/v1/dlq/purge", bytes.NewReader(body))
	purgeResponse := httptest.NewRecorder()
	handler.PurgeDLQItems(purgeResponse, purgeRequest)
	if purgeResponse.Code != http.StatusOK {
		t.Fatalf("purge status = %d, want 200: %s", purgeResponse.Code, purgeResponse.Body.String())
	}
	var purged DLQPurgeSelectionResponse
	if err := json.NewDecoder(purgeResponse.Body).Decode(&purged); err != nil {
		t.Fatal(err)
	}
	if purged.Purged != 1 {
		t.Fatalf("purged = %d, want 1", purged.Purged)
	}
	if got, err := mr.List(handler.cfg.Worker.DeadLetterList); err != nil || len(got) != 1 || got[0] != first {
		t.Fatalf("dead-letter queue = %#v (err=%v), want only first delivery", got, err)
	}
}

func TestOpenAPISpecDocumentsEnqueueAsARealPath(t *testing.T) {
	var document struct {
		Paths      map[string]map[string]any `yaml:"paths"`
		Components struct {
			Schemas map[string]any `yaml:"schemas"`
		} `yaml:"components"`
	}
	if err := yaml.Unmarshal([]byte(openAPISpec), &document); err != nil {
		t.Fatalf("OpenAPI YAML is invalid: %v", err)
	}
	for path, method := range map[string]string{
		"/enqueue":     "post",
		"/dlq":         "get",
		"/dlq/requeue": "post",
		"/dlq/purge":   "post",
		"/workers":     "get",
	} {
		if _, ok := document.Paths[path][method]; !ok {
			t.Errorf("OpenAPI paths does not contain %s %s", strings.ToUpper(method), path)
		}
	}
	if _, ok := document.Components.Schemas["EnqueueRequest"]; !ok {
		t.Fatal("OpenAPI components does not contain EnqueueRequest")
	}
	if _, ok := document.Components.Schemas["EnqueueResponse"]; !ok {
		t.Fatal("OpenAPI components does not contain EnqueueResponse")
	}
	for _, schemaName := range []string{"DLQRequeueRequest", "DLQPurgeSelectionRequest"} {
		schema, ok := document.Components.Schemas[schemaName].(map[string]any)
		if !ok {
			t.Fatalf("OpenAPI schema %s is not an object", schemaName)
		}
		properties, ok := schema["properties"].(map[string]any)
		if !ok {
			t.Fatalf("OpenAPI schema %s properties are not an object", schemaName)
		}
		if _, ok := properties["handles"]; !ok {
			t.Errorf("OpenAPI schema %s does not expose selection handles", schemaName)
		}
		if _, ok := properties["ids"]; ok {
			t.Errorf("OpenAPI schema %s still exposes ambiguous job IDs", schemaName)
		}
	}
	enqueueOperation, ok := document.Paths["/enqueue"]["post"].(map[string]any)
	if !ok {
		t.Fatal("OpenAPI enqueue operation is not an object")
	}
	enqueueDescription, ok := enqueueOperation["description"].(string)
	if !ok {
		t.Fatal("OpenAPI enqueue operation does not contain a string description")
	}
	for _, claim := range []string{"strict FIFO", "ordering wins over priority"} {
		if !strings.Contains(enqueueDescription, claim) {
			t.Errorf("OpenAPI enqueue description does not document %q: %q", claim, enqueueDescription)
		}
	}
	if strings.Contains(enqueueDescription, "not enforced") {
		t.Errorf("OpenAPI enqueue description still claims ordering is not enforced: %q", enqueueDescription)
	}
	for name := range document.Components.Schemas {
		if strings.HasPrefix(name, "/") {
			t.Errorf("OpenAPI path %q is incorrectly nested under component schemas", name)
		}
	}
}

func TestPeekQueue(t *testing.T) {
	handler, mr, cleanup := setupHandlerTest(t)
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
	handler, mr, cleanup := setupHandlerTest(t)
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
	handler, _, cleanup := setupHandlerTest(t)
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
	handler, _, cleanup := setupHandlerTest(t)
	defer cleanup()

	// Create request
	reqBody := BenchRequest{
		Count:       10,
		Priority:    "high",
		Rate:        100,
		Timeout:     5,
		PayloadSize: 1024,
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
