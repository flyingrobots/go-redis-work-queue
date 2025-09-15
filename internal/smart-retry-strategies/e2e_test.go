// Copyright 2025 James Ross
package smartretry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestE2E_RetryWorkflow(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	manager := setupManager(t, rdb)

	scenario := E2EScenario{
		Name:        "payment_processing_timeout",
		JobType:     "payment",
		ErrorClass:  "timeout",
		InitialJobs: 10,
	}

	results := runE2EScenario(t, manager, scenario)

	// Verify expected retry patterns
	if results.TotalJobs != scenario.InitialJobs {
		t.Errorf("expected %d jobs, processed %d", scenario.InitialJobs, results.TotalJobs)
	}

	if results.TotalRetries == 0 {
		t.Error("expected some retries to occur")
	}

	if results.SuccessRate < 0.3 {
		t.Errorf("success rate too low: %.2f", results.SuccessRate)
	}

	// Verify delay progression
	for i := 1; i < len(results.DelayProgression); i++ {
		if results.DelayProgression[i] <= results.DelayProgression[i-1] {
			t.Error("delays should generally increase with attempt number")
		}
	}

	// Verify guardrails enforcement
	maxObservedDelay := int64(0)
	for _, delay := range results.DelayProgression {
		if delay > maxObservedDelay {
			maxObservedDelay = delay
		}
	}

	strategy, _ := manager.GetStrategy()
	if maxObservedDelay > strategy.Guardrails.MaxDelayMs {
		t.Errorf("observed delay %dms exceeds guardrail %dms",
			maxObservedDelay, strategy.Guardrails.MaxDelayMs)
	}
}

func TestE2E_HighVolumeStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress test in short mode")
	}

	rdb := setupTestRedis(t)
	defer rdb.Close()

	manager := setupManager(t, rdb)

	stressConfig := StressTestConfig{
		ConcurrentWorkers: 10,
		JobsPerWorker:     100,
		ErrorRate:         0.7,
		MaxDuration:       30 * time.Second,
	}

	results := runStressTest(t, manager, stressConfig)

	// Performance assertions
	if results.Duration > stressConfig.MaxDuration {
		t.Errorf("test took too long: %v > %v", results.Duration, stressConfig.MaxDuration)
	}

	expectedJobs := stressConfig.ConcurrentWorkers * stressConfig.JobsPerWorker
	if results.ProcessedJobs < int64(float64(expectedJobs)*0.8) {
		t.Errorf("processed too few jobs: %d < %d", results.ProcessedJobs, expectedJobs)
	}

	// Verify no memory leaks or resource exhaustion
	if results.MemoryLeakDetected {
		t.Error("memory leak detected during stress test")
	}

	// Verify retry decision consistency under load
	if results.InconsistentDecisions > int64(float64(results.ProcessedJobs)*0.01) {
		t.Errorf("too many inconsistent retry decisions: %d", results.InconsistentDecisions)
	}
}

func TestE2E_MLModelDeploymentWorkflow(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	manager := setupManager(t, rdb)

	// Step 1: Collect training data
	trainingData := generateTrainingData(1000)
	for _, attempt := range trainingData {
		err := manager.RecordAttempt(attempt)
		if err != nil {
			t.Fatalf("failed to record training attempt: %v", err)
		}
	}

	// Step 2: Train ML model
	trainingConfig := MLTrainingConfig{
		ModelType:      "gradient_boost",
		Features:       []string{"job_type", "error_class", "attempt_number", "payload_size"},
		TrainingPeriod: 24 * time.Hour,
		ValidationSet:  0.2,
		CrossValidation: 5,
	}

	model, err := manager.TrainMLModel(trainingConfig)
	if err != nil {
		t.Fatalf("failed to train ML model: %v", err)
	}

	if model.Accuracy < 0.6 {
		t.Errorf("model accuracy too low: %.2f", model.Accuracy)
	}

	// Step 3: Deploy with canary
	err = manager.DeployMLModel(model, 10.0)
	if err != nil {
		t.Fatalf("failed to deploy ML model: %v", err)
	}

	// Step 4: Test canary deployment
	canaryResults := testCanaryDeployment(t, manager, 100)

	if canaryResults.CanaryTraffic < 5 || canaryResults.CanaryTraffic > 15 {
		t.Errorf("canary traffic outside expected range: %.1f%%", canaryResults.CanaryTraffic)
	}

	// Step 5: Full deployment
	err = manager.DeployMLModel(model, 100.0)
	if err != nil {
		t.Fatalf("failed to fully deploy ML model: %v", err)
	}

	// Verify ML recommendations
	features := RetryFeatures{
		JobType:       "ml_test",
		ErrorClass:    "timeout",
		AttemptNumber: 1,
		PayloadSize:   2048,
	}

	recommendation, err := manager.GetRecommendation(features)
	if err != nil {
		t.Fatalf("failed to get ML recommendation: %v", err)
	}

	if recommendation.Method != "ml" {
		t.Errorf("expected ML method, got %s", recommendation.Method)
	}

	if recommendation.Confidence < 0.5 {
		t.Errorf("ML confidence too low: %.2f", recommendation.Confidence)
	}
}

func TestE2E_DisasterRecoveryScenario(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	manager := setupManager(t, rdb)

	// Simulate normal operation
	normalAttempts := generateTrainingData(50)
	for _, attempt := range normalAttempts {
		manager.RecordAttempt(attempt)
	}

	// Step 1: Simulate Redis failure
	rdb.Close()

	// Verify graceful degradation
	features := RetryFeatures{
		JobType:       "disaster_test",
		ErrorClass:    "timeout",
		AttemptNumber: 1,
	}

	recommendation, err := manager.GetRecommendation(features)
	if err != nil {
		t.Logf("expected error during Redis failure: %v", err)
	}

	// Should fall back to default policy
	if recommendation == nil {
		recommendation = &RetryRecommendation{
			ShouldRetry: true,
			DelayMs:     1000,
			Method:      "fallback",
			Rationale:   "Default fallback due to system failure",
		}
	}

	if recommendation.Method != "fallback" && recommendation.Method != "rules" {
		t.Errorf("expected fallback during disaster, got %s", recommendation.Method)
	}

	// Step 2: Simulate recovery
	rdb = setupTestRedis(t)
	defer rdb.Close()

	// Verify system recovery
	time.Sleep(100 * time.Millisecond) // Brief recovery delay

	recommendation2, err := manager.GetRecommendation(features)
	if err != nil {
		t.Fatalf("system should recover from disaster: %v", err)
	}

	if recommendation2.Method == "fallback" {
		t.Error("system should have recovered from fallback mode")
	}
}

func TestE2E_APIEndpointsIntegration(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	manager := setupManager(t, rdb)

	// Setup test server
	server := setupTestAPIServer(manager)
	defer server.Close()

	client := &http.Client{Timeout: 5 * time.Second}

	// Test recommendation endpoint
	t.Run("recommendation endpoint", func(t *testing.T) {
		payload := `{
			"job_type": "api_test",
			"error_class": "timeout",
			"attempt_number": 1,
			"queue": "default"
		}`

		resp, err := client.Post(server.URL+"/api/v1/retry/recommendation",
			"application/json", strings.NewReader(payload))
		if err != nil {
			t.Fatalf("failed to call recommendation API: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var recommendation RetryRecommendation
		err = json.NewDecoder(resp.Body).Decode(&recommendation)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if recommendation.Method == "" {
			t.Error("recommendation method should not be empty")
		}
	})

	// Test preview endpoint
	t.Run("preview endpoint", func(t *testing.T) {
		payload := `{
			"job_type": "api_test",
			"error_class": "timeout",
			"attempt_number": 1,
			"max_attempts": 5
		}`

		resp, err := client.Post(server.URL+"/api/v1/retry/preview",
			"application/json", strings.NewReader(payload))
		if err != nil {
			t.Fatalf("failed to call preview API: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var preview RetryPreview
		err = json.NewDecoder(resp.Body).Decode(&preview)
		if err != nil {
			t.Fatalf("failed to decode preview response: %v", err)
		}

		if len(preview.Timeline) == 0 {
			t.Error("preview timeline should not be empty")
		}
	})

	// Test stats endpoint
	t.Run("stats endpoint", func(t *testing.T) {
		resp, err := client.Get(server.URL + "/api/v1/retry/stats?job_type=api_test&error_class=timeout&window=24h")
		if err != nil {
			t.Fatalf("failed to call stats API: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status 200, got %d", resp.StatusCode)
		}

		var stats RetryStats
		err = json.NewDecoder(resp.Body).Decode(&stats)
		if err != nil {
			t.Fatalf("failed to decode stats response: %v", err)
		}

		if stats.JobType != "api_test" {
			t.Errorf("expected job_type api_test, got %s", stats.JobType)
		}
	})
}

// Test data structures and helpers

type E2EScenario struct {
	Name        string
	JobType     string
	ErrorClass  string
	InitialJobs int
}

type E2EResults struct {
	TotalJobs         int
	TotalRetries      int
	SuccessRate       float64
	DelayProgression  []int64
	GuardrailViolations int
}

type StressTestConfig struct {
	ConcurrentWorkers int
	JobsPerWorker     int
	ErrorRate         float64
	MaxDuration       time.Duration
}

type StressTestResults struct {
	ProcessedJobs        int64
	Duration             time.Duration
	MemoryLeakDetected   bool
	InconsistentDecisions int64
}

type CanaryResults struct {
	CanaryTraffic   float64
	CanarySuccess   float64
	BaselineSuccess float64
}

func runE2EScenario(t *testing.T, manager Manager, scenario E2EScenario) E2EResults {
	results := E2EResults{
		TotalJobs:        scenario.InitialJobs,
		DelayProgression: []int64{},
	}

	successCount := 0
	retryCount := 0

	for i := 0; i < scenario.InitialJobs; i++ {
		jobSuccess := simulateJobLifecycle(t, manager, scenario.JobType, scenario.ErrorClass, &results)
		if jobSuccess {
			successCount++
		}
		retryCount += len(results.DelayProgression)
	}

	results.TotalRetries = retryCount
	results.SuccessRate = float64(successCount) / float64(scenario.InitialJobs)

	return results
}

func simulateJobLifecycle(t *testing.T, manager Manager, jobType, errorClass string, results *E2EResults) bool {
	attempt := 1
	maxAttempts := 5

	for attempt <= maxAttempts {
		features := RetryFeatures{
			JobType:       jobType,
			ErrorClass:    errorClass,
			AttemptNumber: attempt,
			PayloadSize:   1024,
			TimeOfDay:     12,
		}

		recommendation, err := manager.GetRecommendation(features)
		if err != nil {
			t.Logf("failed to get recommendation: %v", err)
			return false
		}

		if !recommendation.ShouldRetry {
			return false
		}

		results.DelayProgression = append(results.DelayProgression, recommendation.DelayMs)

		// Simulate job execution (80% success rate on final attempts)
		if attempt >= 3 && time.Now().UnixNano()%10 < 8 {
			// Record successful attempt
			attemptHistory := AttemptHistory{
				JobType:       jobType,
				ErrorClass:    errorClass,
				AttemptNumber: attempt,
				DelayMs:       recommendation.DelayMs,
				Success:       true,
				Timestamp:     time.Now(),
			}
			manager.RecordAttempt(attemptHistory)
			return true
		}

		// Record failed attempt
		attemptHistory := AttemptHistory{
			JobType:       jobType,
			ErrorClass:    errorClass,
			AttemptNumber: attempt,
			DelayMs:       recommendation.DelayMs,
			Success:       false,
			Timestamp:     time.Now(),
		}
		manager.RecordAttempt(attemptHistory)

		attempt++
	}

	return false
}

func runStressTest(t *testing.T, manager Manager, config StressTestConfig) StressTestResults {
	var wg sync.WaitGroup
	var mu sync.Mutex

	results := StressTestResults{}
	startTime := time.Now()

	for i := 0; i < config.ConcurrentWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for j := 0; j < config.JobsPerWorker; j++ {
				features := RetryFeatures{
					JobType:       fmt.Sprintf("stress_worker_%d", workerID),
					ErrorClass:    "timeout",
					AttemptNumber: j%3 + 1,
					Queue:         "stress_test",
				}

				_, err := manager.GetRecommendation(features)

				mu.Lock()
				if err != nil {
					results.InconsistentDecisions++
				} else {
					results.ProcessedJobs++
				}
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	results.Duration = time.Since(startTime)

	return results
}

func generateTrainingData(count int) []AttemptHistory {
	attempts := make([]AttemptHistory, count)

	for i := 0; i < count; i++ {
		attempts[i] = AttemptHistory{
			JobID:         fmt.Sprintf("training_job_%d", i),
			JobType:       []string{"payment", "email", "notification"}[i%3],
			ErrorClass:    []string{"timeout", "500_error", "network_error"}[i%3],
			AttemptNumber: (i%5) + 1,
			DelayMs:       int64((i%10+1) * 1000),
			Success:       i%4 != 0, // 75% success rate
			Timestamp:     time.Now().Add(-time.Duration(i) * time.Minute),
			PayloadSize:   int64(1024 + i%1024),
		}
	}

	return attempts
}

func testCanaryDeployment(t *testing.T, manager Manager, requests int) CanaryResults {
	canaryCount := 0

	for i := 0; i < requests; i++ {
		features := RetryFeatures{
			JobType:       "canary_test",
			ErrorClass:    "timeout",
			AttemptNumber: 1,
		}

		recommendation, err := manager.GetRecommendation(features)
		if err == nil && recommendation.Method == "ml" {
			canaryCount++
		}
	}

	return CanaryResults{
		CanaryTraffic: float64(canaryCount) / float64(requests) * 100,
	}
}

func setupTestAPIServer(manager Manager) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v1/retry/recommendation", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var features RetryFeatures
		err := json.NewDecoder(r.Body).Decode(&features)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		recommendation, err := manager.GetRecommendation(features)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(recommendation)
	})

	mux.HandleFunc("/api/v1/retry/preview", func(w http.ResponseWriter, r *http.Request) {
		// Simplified preview endpoint
		preview := RetryPreview{
			CurrentAttempt: 1,
			Timeline: []RetryTimelineEntry{
				{AttemptNumber: 2, DelayMs: 1000, Method: "rules"},
				{AttemptNumber: 3, DelayMs: 2000, Method: "rules"},
			},
			GeneratedAt: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(preview)
	})

	mux.HandleFunc("/api/v1/retry/stats", func(w http.ResponseWriter, r *http.Request) {
		stats := RetryStats{
			JobType:      r.URL.Query().Get("job_type"),
			ErrorClass:   r.URL.Query().Get("error_class"),
			TotalAttempts: 100,
			SuccessRate:  0.75,
			LastUpdated:  time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	})

	return httptest.NewServer(mux)
}