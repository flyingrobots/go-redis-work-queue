// Copyright 2025 James Ross
package smartretry

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestShadowMode_IntegrationWithRedis(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	manager := setupManager(t, rdb)

	tests := []struct {
		name                string
		features            RetryFeatures
		expectedShadowLog   bool
		expectedPrimary     bool
		primaryStrategy     string
		shadowStrategy      string
	}{
		{
			name: "shadow mode comparison - rule vs bayesian",
			features: RetryFeatures{
				JobType:       "payment",
				ErrorClass:    "timeout",
				AttemptNumber: 2,
				Queue:         "critical",
			},
			expectedShadowLog: true,
			expectedPrimary:   true,
			primaryStrategy:   "rules",
			shadowStrategy:    "bayesian",
		},
		{
			name: "shadow mode with high confidence bayesian",
			features: RetryFeatures{
				JobType:       "notification",
				ErrorClass:    "503_error",
				AttemptNumber: 1,
				PayloadSize:   1024,
			},
			expectedShadowLog: true,
			expectedPrimary:   true,
			primaryStrategy:   "rules",
			shadowStrategy:    "bayesian",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			recommendation, err := manager.GetRecommendation(tt.features)
			if err != nil {
				t.Fatalf("GetRecommendation failed: %v", err)
			}

			if recommendation.ShouldRetry != tt.expectedPrimary {
				t.Errorf("expected primary ShouldRetry=%v, got %v",
					tt.expectedPrimary, recommendation.ShouldRetry)
			}

			if recommendation.Method != tt.primaryStrategy {
				t.Errorf("expected primary method=%s, got %s",
					tt.primaryStrategy, recommendation.Method)
			}

			shadowKey := "shadow_recommendations:*"
			keys, err := rdb.Keys(context.Background(), shadowKey).Result()
			if err != nil {
				t.Fatalf("failed to check shadow logs: %v", err)
			}

			if tt.expectedShadowLog && len(keys) == 0 {
				t.Error("expected shadow recommendation to be logged")
			}

			if len(keys) > 0 {
				shadowData, err := rdb.HGetAll(context.Background(), keys[0]).Result()
				if err == nil {
					if shadowData["method"] != tt.shadowStrategy {
						t.Errorf("expected shadow method=%s, got %s",
							tt.shadowStrategy, shadowData["method"])
					}
				}
			}
		})
	}
}

func TestABTesting_TrafficSplitting(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	manager := setupManager(t, rdb)

	abConfig := ABTestConfig{
		Name:           "rules_vs_bayesian",
		TrafficPercent: 50.0,
		ControlMethod:  "rules",
		TestMethod:     "bayesian",
		StartTime:      time.Now(),
		EndTime:        time.Now().Add(24 * time.Hour),
		Enabled:        true,
	}

	err := manager.StartABTest(abConfig)
	if err != nil {
		t.Fatalf("failed to start A/B test: %v", err)
	}

	features := RetryFeatures{
		JobType:       "email",
		ErrorClass:    "timeout",
		AttemptNumber: 1,
		Queue:         "default",
	}

	controlCount := 0
	testCount := 0
	totalRuns := 100

	for i := 0; i < totalRuns; i++ {
		recommendation, err := manager.GetRecommendation(features)
		if err != nil {
			t.Fatalf("GetRecommendation failed: %v", err)
		}

		switch recommendation.Method {
		case "rules":
			controlCount++
		case "bayesian":
			testCount++
		}
	}

	expectedControl := int(float64(totalRuns) * 0.5)
	tolerance := int(float64(totalRuns) * 0.2)

	if abs(controlCount-expectedControl) > tolerance {
		t.Errorf("control group traffic outside tolerance: expected ~%d, got %d",
			expectedControl, controlCount)
	}

	if abs(testCount-expectedControl) > tolerance {
		t.Errorf("test group traffic outside tolerance: expected ~%d, got %d",
			expectedControl, testCount)
	}

	stats, err := manager.GetABTestStats("rules_vs_bayesian")
	if err != nil {
		t.Fatalf("failed to get A/B test stats: %v", err)
	}

	if stats.TotalRequests != int64(totalRuns) {
		t.Errorf("expected %d total requests, got %d", totalRuns, stats.TotalRequests)
	}

	if stats.ControlGroup.RequestCount == 0 || stats.TestGroup.RequestCount == 0 {
		t.Error("both control and test groups should have received requests")
	}
}

func TestRetryStrategy_PolicyUpdateIntegration(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	manager := setupManager(t, rdb)

	newPolicy := RetryPolicy{
		Name:              "integration_test_policy",
		ErrorPatterns:     []string{"integration_.*"},
		MaxAttempts:       5,
		BaseDelayMs:       2000,
		BackoffMultiplier: 1.5,
		Priority:          200,
	}

	err := manager.AddPolicy(newPolicy)
	if err != nil {
		t.Fatalf("failed to add policy: %v", err)
	}

	features := RetryFeatures{
		JobType:       "test",
		ErrorClass:    "integration_timeout",
		AttemptNumber: 1,
	}

	recommendation, err := manager.GetRecommendation(features)
	if err != nil {
		t.Fatalf("GetRecommendation failed: %v", err)
	}

	if recommendation.MaxAttempts != 5 {
		t.Errorf("expected max attempts 5, got %d", recommendation.MaxAttempts)
	}

	expectedDelay := int64(2000)
	tolerance := int64(500)
	if abs64(recommendation.DelayMs-expectedDelay) > tolerance {
		t.Errorf("expected delay ~%dms, got %dms", expectedDelay, recommendation.DelayMs)
	}

	err = manager.RemovePolicy("integration_test_policy")
	if err != nil {
		t.Fatalf("failed to remove policy: %v", err)
	}

	recommendation2, err := manager.GetRecommendation(features)
	if err != nil {
		t.Fatalf("GetRecommendation failed after policy removal: %v", err)
	}

	if recommendation2.MaxAttempts == 5 {
		t.Error("policy should no longer affect recommendations after removal")
	}
}

func TestBayesianModel_LearningIntegration(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	manager := setupManager(t, rdb)

	jobType := "learning_test"
	errorClass := "network_error"

	attempts := []AttemptHistory{
		{JobType: jobType, ErrorClass: errorClass, DelayMs: 1000, Success: false, Timestamp: time.Now()},
		{JobType: jobType, ErrorClass: errorClass, DelayMs: 1500, Success: false, Timestamp: time.Now()},
		{JobType: jobType, ErrorClass: errorClass, DelayMs: 2000, Success: true, Timestamp: time.Now()},
		{JobType: jobType, ErrorClass: errorClass, DelayMs: 2500, Success: true, Timestamp: time.Now()},
		{JobType: jobType, ErrorClass: errorClass, DelayMs: 3000, Success: true, Timestamp: time.Now()},
	}

	for _, attempt := range attempts {
		err := manager.RecordAttempt(attempt)
		if err != nil {
			t.Fatalf("failed to record attempt: %v", err)
		}
	}

	err := manager.UpdateBayesianModel(jobType, errorClass)
	if err != nil {
		t.Fatalf("failed to update Bayesian model: %v", err)
	}

	features := RetryFeatures{
		JobType:       jobType,
		ErrorClass:    errorClass,
		AttemptNumber: 1,
	}

	recommendation, err := manager.GetRecommendation(features)
	if err != nil {
		t.Fatalf("GetRecommendation failed: %v", err)
	}

	if recommendation.Method != "bayesian" {
		t.Errorf("expected bayesian method, got %s", recommendation.Method)
	}

	if recommendation.DelayMs < 2000 {
		t.Errorf("expected delay >= 2000ms based on learning, got %dms", recommendation.DelayMs)
	}

	if recommendation.Confidence < 0.3 {
		t.Errorf("expected reasonable confidence, got %.2f", recommendation.Confidence)
	}
}

func TestGuardrails_IntegrationEnforcement(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	manager := setupManager(t, rdb)

	guardrails := PolicyGuardrails{
		MaxAttempts:      3,
		MaxDelayMs:       10000,
		MaxBudgetPercent: 50.0,
		EmergencyStop:    false,
	}

	err := manager.UpdateGuardrails(guardrails)
	if err != nil {
		t.Fatalf("failed to update guardrails: %v", err)
	}

	features := RetryFeatures{
		JobType:       "guardrail_test",
		ErrorClass:    "timeout",
		AttemptNumber: 4,
		Queue:         "default",
	}

	recommendation, err := manager.GetRecommendation(features)
	if err != nil {
		t.Fatalf("GetRecommendation failed: %v", err)
	}

	if recommendation.ShouldRetry {
		t.Error("should not retry when attempt exceeds guardrail max attempts")
	}

	if len(recommendation.PolicyGuardrails) == 0 {
		t.Error("expected policy guardrails to be populated")
	}

	features2 := RetryFeatures{
		JobType:       "guardrail_test",
		ErrorClass:    "timeout",
		AttemptNumber: 2,
	}

	recommendation2, err := manager.GetRecommendation(features2)
	if err != nil {
		t.Fatalf("GetRecommendation failed: %v", err)
	}

	if recommendation2.DelayMs > 10000 {
		t.Errorf("delay %dms exceeds guardrail max %dms",
			recommendation2.DelayMs, guardrails.MaxDelayMs)
	}
}

func TestDataCollection_SamplingIntegration(t *testing.T) {
	rdb := setupTestRedis(t)
	defer rdb.Close()

	manager := setupManager(t, rdb)

	dataConfig := DataCollectionConfig{
		Enabled:           true,
		SampleRate:        0.5,
		RetentionDays:     7,
		FeatureExtraction: true,
	}

	err := manager.UpdateDataCollection(dataConfig)
	if err != nil {
		t.Fatalf("failed to update data collection config: %v", err)
	}

	attempts := []AttemptHistory{}
	for i := 0; i < 100; i++ {
		attempt := AttemptHistory{
			JobID:         fmt.Sprintf("job_%d", i),
			JobType:       "sampling_test",
			ErrorClass:    "test_error",
			AttemptNumber: 1,
			DelayMs:       1000,
			Success:       i%2 == 0,
			Timestamp:     time.Now(),
		}
		attempts = append(attempts, attempt)
	}

	recordedCount := 0
	for _, attempt := range attempts {
		err := manager.RecordAttempt(attempt)
		if err == nil {
			recordedCount++
		}
	}

	expectedSampled := int(float64(len(attempts)) * dataConfig.SampleRate)
	tolerance := int(float64(len(attempts)) * 0.2)

	if abs(recordedCount-expectedSampled) > tolerance {
		t.Errorf("sampling outside tolerance: expected ~%d, recorded %d",
			expectedSampled, recordedCount)
	}

	stats, err := manager.GetStats("sampling_test", "test_error", 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.TotalAttempts == 0 {
		t.Error("expected some attempts to be recorded")
	}
}

// Helper functions

func setupTestRedis(t *testing.T) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15,
	})

	// Clean up test database
	ctx := context.Background()
	rdb.FlushDB(ctx)

	// Test connection
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}

	return rdb
}

func setupManager(t *testing.T, rdb *redis.Client) Manager {
	config := Config{
		Enabled:   true,
		RedisAddr: "localhost:6379",
		RedisDB:   15,
		Strategy: RetryStrategy{
			Name:              "test_strategy",
			Enabled:           true,
			BayesianThreshold: 0.7,
			MLEnabled:         false,
			Policies: []RetryPolicy{
				{
					Name:              "default_policy",
					ErrorPatterns:     []string{".*"},
					MaxAttempts:       3,
					BaseDelayMs:       1000,
					BackoffMultiplier: 2.0,
					Priority:          1,
				},
			},
			Guardrails: PolicyGuardrails{
				MaxAttempts: 5,
				MaxDelayMs:  30000,
			},
		},
		DataCollection: DataCollectionConfig{
			Enabled:       true,
			SampleRate:    1.0,
			RetentionDays: 7,
		},
	}

	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("failed to create manager: %v", err)
	}

	return manager
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

// Additional integration test types

type ABTestConfig struct {
	Name           string
	TrafficPercent float64
	ControlMethod  string
	TestMethod     string
	StartTime      time.Time
	EndTime        time.Time
	Enabled        bool
}

type ABTestStats struct {
	Name          string
	TotalRequests int64
	ControlGroup  ABGroupStats
	TestGroup     ABGroupStats
}

type ABGroupStats struct {
	RequestCount int64
	SuccessRate  float64
	AvgDelayMs   float64
}

// Extended Manager interface for integration testing
type ExtendedManager interface {
	Manager
	StartABTest(config ABTestConfig) error
	GetABTestStats(name string) (*ABTestStats, error)
	UpdateDataCollection(config DataCollectionConfig) error
}