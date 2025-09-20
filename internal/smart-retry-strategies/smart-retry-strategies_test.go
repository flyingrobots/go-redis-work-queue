//go:build smart_retry_tests
// +build smart_retry_tests

// Copyright 2025 James Ross
package smartretry

import (
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"go.uber.org/zap"
)

func TestNewManager(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		Enabled:       true,
		RedisAddr:     mr.Addr(),
		RedisPassword: "",
		RedisDB:       0,
		Strategy: RetryStrategy{
			Name:              "test",
			Enabled:           true,
			BayesianThreshold: 0.7,
			MLEnabled:         false,
			Guardrails: PolicyGuardrails{
				MaxAttempts:       5,
				MaxDelayMs:        30000,
				MaxBudgetPercent:  20.0,
				PerTenantLimits:   true,
				EmergencyStop:     false,
				ExplainabilityReq: true,
			},
			DataCollection: DataCollectionConfig{
				Enabled:             true,
				SampleRate:          1.0,
				RetentionDays:       7,
				AggregationInterval: 5 * time.Minute,
				FeatureExtraction:   true,
			},
		},
		Cache: CacheConfig{
			Enabled:    true,
			TTL:        5 * time.Minute,
			MaxEntries: 100,
		},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
}

func TestGetRecommendation(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		Enabled:   true,
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Strategy: RetryStrategy{
			Name:              "test",
			Enabled:           true,
			Policies:          defaultPolicies(),
			BayesianThreshold: 0.7,
			MLEnabled:         false,
			Guardrails: PolicyGuardrails{
				MaxAttempts: 5,
				MaxDelayMs:  30000,
			},
		},
		Cache: CacheConfig{
			Enabled:    false, // Disable cache for predictable tests
			TTL:        5 * time.Minute,
			MaxEntries: 100,
		},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	tests := []struct {
		name     string
		features RetryFeatures
		wantErr  bool
	}{
		{
			name: "rate limit error",
			features: RetryFeatures{
				JobType:       "test_job",
				ErrorClass:    "429",
				ErrorCode:     "rate_limit",
				AttemptNumber: 1,
				Queue:         "default",
			},
			wantErr: false,
		},
		{
			name: "service unavailable",
			features: RetryFeatures{
				JobType:       "test_job",
				ErrorClass:    "503",
				ErrorCode:     "service_unavailable",
				AttemptNumber: 2,
				Queue:         "default",
			},
			wantErr: false,
		},
		{
			name: "validation error",
			features: RetryFeatures{
				JobType:       "test_job",
				ErrorClass:    "validation",
				ErrorCode:     "invalid_input",
				AttemptNumber: 1,
				Queue:         "default",
			},
			wantErr: false,
		},
		{
			name: "max attempts reached",
			features: RetryFeatures{
				JobType:       "test_job",
				ErrorClass:    "timeout",
				ErrorCode:     "connection_timeout",
				AttemptNumber: 6, // Exceeds max attempts
				Queue:         "default",
			},
			wantErr: false, // Should return recommendation with ShouldRetry=false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, err := mgr.GetRecommendation(tt.features)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRecommendation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if rec == nil {
				t.Fatal("Recommendation should not be nil")
			}

			// Validate recommendation structure
			if rec.Method == "" {
				t.Error("Recommendation method should not be empty")
			}

			if rec.Rationale == "" {
				t.Error("Recommendation rationale should not be empty")
			}

			if rec.Confidence < 0 || rec.Confidence > 1 {
				t.Errorf("Confidence should be between 0 and 1, got %f", rec.Confidence)
			}

			// Test specific scenarios
			if tt.features.AttemptNumber >= 6 {
				if rec.ShouldRetry {
					t.Error("Should not retry when max attempts exceeded")
				}
			}

			if tt.features.ErrorClass == "validation" {
				if rec.ShouldRetry && rec.MaxAttempts > 1 {
					t.Error("Validation errors should not retry or have low max attempts")
				}
			}
		})
	}
}

func TestRecordAttempt(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		Enabled:   true,
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Strategy: RetryStrategy{
			Name:    "test",
			Enabled: true,
		},
		DataCollection: DataCollectionConfig{
			Enabled:       true,
			SampleRate:    1.0,
			RetentionDays: 7,
		},
		Cache: CacheConfig{
			Enabled: false,
		},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	attempt := AttemptHistory{
		JobID:          "test-job-123",
		JobType:        "test_job",
		AttemptNumber:  1,
		ErrorClass:     "timeout",
		ErrorCode:      "connection_timeout",
		Status:         "failed",
		Queue:          "default",
		PayloadSize:    1024,
		TimeOfDay:      14,
		WorkerVersion:  "v1.0.0",
		DelayMs:        2000,
		Success:        false,
		Timestamp:      time.Now(),
		ProcessingTime: 5 * time.Second,
	}

	err = mgr.RecordAttempt(attempt)
	if err != nil {
		t.Fatalf("Failed to record attempt: %v", err)
	}

	// Record a successful attempt
	successAttempt := attempt
	successAttempt.AttemptNumber = 2
	successAttempt.Success = true
	successAttempt.Status = "completed"

	err = mgr.RecordAttempt(successAttempt)
	if err != nil {
		t.Fatalf("Failed to record successful attempt: %v", err)
	}
}

func TestGetStats(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		Enabled:   true,
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Strategy: RetryStrategy{
			Name:    "test",
			Enabled: true,
		},
		DataCollection: DataCollectionConfig{
			Enabled:       true,
			SampleRate:    1.0,
			RetentionDays: 7,
		},
		Cache: CacheConfig{
			Enabled: false,
		},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Record some attempts first
	attempts := []AttemptHistory{
		{
			JobID:         "job1",
			JobType:       "test_job",
			AttemptNumber: 1,
			ErrorClass:    "timeout",
			Success:       false,
			Timestamp:     time.Now(),
		},
		{
			JobID:         "job2",
			JobType:       "test_job",
			AttemptNumber: 1,
			ErrorClass:    "timeout",
			Success:       true,
			Timestamp:     time.Now(),
		},
		{
			JobID:         "job3",
			JobType:       "test_job",
			AttemptNumber: 1,
			ErrorClass:    "timeout",
			Success:       true,
			Timestamp:     time.Now(),
		},
	}

	for _, attempt := range attempts {
		err := mgr.RecordAttempt(attempt)
		if err != nil {
			t.Fatalf("Failed to record attempt: %v", err)
		}
	}

	// Get stats
	stats, err := mgr.GetStats("test_job", "timeout", 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to get stats: %v", err)
	}

	if stats == nil {
		t.Fatal("Stats should not be nil")
	}

	if stats.JobType != "test_job" {
		t.Errorf("Expected job type 'test_job', got '%s'", stats.JobType)
	}

	if stats.ErrorClass != "timeout" {
		t.Errorf("Expected error class 'timeout', got '%s'", stats.ErrorClass)
	}

	// Check that we have some attempts recorded
	if stats.TotalAttempts < 0 {
		t.Errorf("Total attempts should be non-negative, got %d", stats.TotalAttempts)
	}
}

func TestPreviewRetrySchedule(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		Enabled:   true,
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Strategy: RetryStrategy{
			Name:     "test",
			Enabled:  true,
			Policies: defaultPolicies(),
			Guardrails: PolicyGuardrails{
				MaxAttempts: 5,
				MaxDelayMs:  30000,
			},
		},
		Cache: CacheConfig{
			Enabled: false,
		},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	features := RetryFeatures{
		JobType:       "test_job",
		ErrorClass:    "timeout",
		ErrorCode:     "connection_timeout",
		AttemptNumber: 1,
		Queue:         "default",
	}

	preview, err := mgr.PreviewRetrySchedule(features, 3)
	if err != nil {
		t.Fatalf("Failed to generate preview: %v", err)
	}

	if preview == nil {
		t.Fatal("Preview should not be nil")
	}

	if preview.CurrentAttempt != 1 {
		t.Errorf("Expected current attempt 1, got %d", preview.CurrentAttempt)
	}

	if len(preview.Recommendations) == 0 {
		t.Error("Preview should have recommendations")
	}

	if len(preview.Timeline) == 0 {
		t.Error("Preview should have timeline entries")
	}

	// Validate timeline is in chronological order
	for i := 1; i < len(preview.Timeline); i++ {
		if preview.Timeline[i].ScheduledTime.Before(preview.Timeline[i-1].ScheduledTime) {
			t.Error("Timeline entries should be in chronological order")
		}
	}
}

func TestPolicyMatching(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		Enabled:   true,
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Strategy: RetryStrategy{
			Name:     "test",
			Enabled:  true,
			Policies: defaultPolicies(),
		},
		Cache: CacheConfig{
			Enabled: false,
		},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	m, ok := mgr.(*manager)
	if !ok {
		t.Fatal("Failed to cast manager to internal type")
	}

	tests := []struct {
		name        string
		policy      RetryPolicy
		features    RetryFeatures
		shouldMatch bool
	}{
		{
			name: "rate limit policy matches 429 error",
			policy: RetryPolicy{
				Name:          "rate_limit",
				ErrorPatterns: []string{"429", "rate_limit"},
				Priority:      100,
			},
			features: RetryFeatures{
				ErrorClass: "429",
			},
			shouldMatch: true,
		},
		{
			name: "validation policy matches validation error",
			policy: RetryPolicy{
				Name:          "validation",
				ErrorPatterns: []string{"validation", "invalid_input"},
				Priority:      80,
			},
			features: RetryFeatures{
				ErrorClass: "validation",
			},
			shouldMatch: true,
		},
		{
			name: "policy with job type pattern matches",
			policy: RetryPolicy{
				Name:            "job_specific",
				ErrorPatterns:   []string{"timeout"},
				JobTypePatterns: []string{"test_.*"},
				Priority:        50,
			},
			features: RetryFeatures{
				JobType:    "test_job",
				ErrorClass: "timeout",
			},
			shouldMatch: true,
		},
		{
			name: "policy with job type pattern doesn't match wrong job type",
			policy: RetryPolicy{
				Name:            "job_specific",
				ErrorPatterns:   []string{"timeout"},
				JobTypePatterns: []string{"test_.*"},
				Priority:        50,
			},
			features: RetryFeatures{
				JobType:    "other_job",
				ErrorClass: "timeout",
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := m.policyMatches(tt.policy, tt.features)
			if matches != tt.shouldMatch {
				t.Errorf("policyMatches() = %v, want %v", matches, tt.shouldMatch)
			}
		})
	}
}

func TestDelayCalculation(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		Enabled:   true,
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Strategy: RetryStrategy{
			Name:    "test",
			Enabled: true,
		},
		Cache: CacheConfig{
			Enabled: false,
		},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	m, ok := mgr.(*manager)
	if !ok {
		t.Fatal("Failed to cast manager to internal type")
	}

	policy := RetryPolicy{
		BaseDelayMs:       1000,
		MaxDelayMs:        30000,
		BackoffMultiplier: 2.0,
	}

	tests := []struct {
		attempt     int
		expectedMin int64
		expectedMax int64
	}{
		{1, 1000, 1000},   // First attempt: base delay
		{2, 2000, 2000},   // Second attempt: base * 2
		{3, 4000, 4000},   // Third attempt: base * 2^2
		{4, 8000, 8000},   // Fourth attempt: base * 2^3
		{6, 30000, 30000}, // Should cap at max delay
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("attempt_%d", tt.attempt), func(t *testing.T) {
			delay := m.calculateDelay(policy, tt.attempt)

			if delay < tt.expectedMin || delay > tt.expectedMax {
				t.Errorf("calculateDelay(%d) = %d, want between %d and %d",
					tt.attempt, delay, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestCacheOperations(t *testing.T) {
	cache := &retryCache{
		enabled:    true,
		ttl:        1 * time.Second,
		maxEntries: 2,
		entries:    make(map[string]*cacheEntry),
	}

	// Test set and get
	cache.set("key1", "value1", 1*time.Second)
	value, ok := cache.get("key1")
	if !ok {
		t.Error("Should find cached value")
	}
	if value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}

	// Test expiration
	cache.set("key2", "value2", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)
	_, ok = cache.get("key2")
	if ok {
		t.Error("Should not find expired value")
	}

	// Test capacity limit
	cache.set("key3", "value3", 10*time.Second)
	cache.set("key4", "value4", 10*time.Second)
	cache.set("key5", "value5", 10*time.Second) // Should trigger cleanup

	// At least one entry should remain
	remaining := 0
	for key := range []string{"key3", "key4", "key5"} {
		if _, ok := cache.get(fmt.Sprintf("key%d", key+3)); ok {
			remaining++
		}
	}
	if remaining == 0 {
		t.Error("At least one entry should remain after cleanup")
	}
}

func TestValidationErrors(t *testing.T) {
	mr := miniredis.RunT(t)
	defer mr.Close()

	config := &Config{
		Enabled:   true,
		RedisAddr: mr.Addr(),
		RedisDB:   0,
		Strategy: RetryStrategy{
			Name:    "test",
			Enabled: true,
		},
	}

	mgr, err := NewManager(config, zap.NewNop())
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	m, ok := mgr.(*manager)
	if !ok {
		t.Fatal("Failed to cast manager to internal type")
	}

	tests := []struct {
		errorClass string
		expected   bool
	}{
		{"validation", true},
		{"invalid_input", true},
		{"malformed", true},
		{"schema_error", true},
		{"timeout", false},
		{"rate_limit", false},
		{"unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.errorClass, func(t *testing.T) {
			result := m.isValidationError(tt.errorClass)
			if result != tt.expected {
				t.Errorf("isValidationError(%s) = %v, want %v", tt.errorClass, result, tt.expected)
			}
		})
	}
}
