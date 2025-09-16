// Copyright 2025 James Ross
package smartretry

import (
	"math"
	"regexp"
	"testing"
	"time"
)

func TestRetryPolicy_BasicErrorMatching(t *testing.T) {
	tests := []struct {
		name         string
		policy       RetryPolicy
		features     RetryFeatures
		shouldMatch  bool
	}{
		{
			name: "exact error class match",
			policy: RetryPolicy{
				ErrorPatterns: []string{"timeout"},
			},
			features: RetryFeatures{
				ErrorClass: "timeout",
			},
			shouldMatch: true,
		},
		{
			name: "regex error pattern match",
			policy: RetryPolicy{
				ErrorPatterns: []string{"5\\d\\d_error"},
			},
			features: RetryFeatures{
				ErrorClass: "503_error",
			},
			shouldMatch: true,
		},
		{
			name: "no match",
			policy: RetryPolicy{
				ErrorPatterns: []string{"network_error"},
			},
			features: RetryFeatures{
				ErrorClass: "validation_error",
			},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := basicPolicyMatches(tt.policy, tt.features)
			if matches != tt.shouldMatch {
				t.Errorf("expected match=%v, got match=%v", tt.shouldMatch, matches)
			}
		})
	}
}

func TestBayesianBucket_BasicProbabilityCalculation(t *testing.T) {
	tests := []struct {
		name        string
		successes   int
		failures    int
		expectProb  float64
		tolerance   float64
	}{
		{
			name:        "equal successes and failures",
			successes:   10,
			failures:    10,
			expectProb:  0.5,
			tolerance:   0.05,
		},
		{
			name:        "high success rate",
			successes:   90,
			failures:    10,
			expectProb:  0.9,
			tolerance:   0.05,
		},
		{
			name:        "low success rate",
			successes:   10,
			failures:    90,
			expectProb:  0.1,
			tolerance:   0.05,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket := &BayesianBucket{
				Successes: tt.successes,
				Failures:  tt.failures,
			}

			// Calculate Beta distribution parameters
			alpha := float64(bucket.Successes + 1) // Add prior
			beta := float64(bucket.Failures + 1)   // Add prior

			// Mean of Beta distribution
			probability := alpha / (alpha + beta)
			bucket.Probability = probability

			if math.Abs(bucket.Probability-tt.expectProb) > tt.tolerance {
				t.Errorf("expected probability %.2f (±%.2f), got %.2f",
					tt.expectProb, tt.tolerance, bucket.Probability)
			}
		})
	}
}

func TestDelayCalculation_Basic(t *testing.T) {
	tests := []struct {
		name           string
		baseDelayMs    int64
		multiplier     float64
		attemptNumber  int
		expectedDelayMs int64
		tolerance       int64
	}{
		{
			name:            "exponential backoff",
			baseDelayMs:     1000,
			multiplier:      2.0,
			attemptNumber:   3,
			expectedDelayMs: 4000, // 1000 * 2^(3-1) = 4000
			tolerance:       100,
		},
		{
			name:            "linear backoff",
			baseDelayMs:     500,
			multiplier:      1.0,
			attemptNumber:   4,
			expectedDelayMs: 2000, // 500 * 4 = 2000
			tolerance:       100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delayMs := basicCalculateDelay(tt.baseDelayMs, tt.multiplier, tt.attemptNumber)

			if abs64(delayMs-tt.expectedDelayMs) > tt.tolerance {
				t.Errorf("expected delay %dms (±%dms), got %dms",
					tt.expectedDelayMs, tt.tolerance, delayMs)
			}
		})
	}
}

func TestPolicySelection_Basic(t *testing.T) {
	policies := []RetryPolicy{
		{
			Name:          "high_priority_timeout",
			Priority:      100,
			ErrorPatterns: []string{"timeout"},
			MaxAttempts:   5,
		},
		{
			Name:          "general_network",
			Priority:      50,
			ErrorPatterns: []string{"network_.*", "connection_.*"},
			MaxAttempts:   3,
		},
		{
			Name:          "fallback",
			Priority:      1,
			ErrorPatterns: []string{".*"}, // Matches everything
			MaxAttempts:   2,
		},
	}

	tests := []struct {
		name           string
		features       RetryFeatures
		expectedPolicy string
	}{
		{
			name: "timeout matches high priority",
			features: RetryFeatures{
				ErrorClass: "timeout",
			},
			expectedPolicy: "high_priority_timeout",
		},
		{
			name: "network error matches medium priority",
			features: RetryFeatures{
				ErrorClass: "network_unreachable",
			},
			expectedPolicy: "general_network",
		},
		{
			name: "unknown error matches fallback",
			features: RetryFeatures{
				ErrorClass: "unknown_error",
			},
			expectedPolicy: "fallback",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := basicSelectBestPolicy(policies, tt.features)

			if policy == nil {
				t.Fatal("expected to find a matching policy")
			}

			if policy.Name != tt.expectedPolicy {
				t.Errorf("expected policy '%s', got '%s'",
					tt.expectedPolicy, policy.Name)
			}
		})
	}
}

// Helper functions for basic testing

func basicPolicyMatches(policy RetryPolicy, features RetryFeatures) bool {
	// Check error patterns
	for _, pattern := range policy.ErrorPatterns {
		if matched, _ := regexp.MatchString(pattern, features.ErrorClass); matched {
			return true
		}
	}

	// Check job type patterns
	for _, pattern := range policy.JobTypePatterns {
		if matched, _ := regexp.MatchString(pattern, features.JobType); matched {
			return true
		}
	}

	return false
}

func basicCalculateDelay(baseDelayMs int64, multiplier float64, attemptNumber int) int64 {
	delay := float64(baseDelayMs)

	if multiplier > 1.0 {
		// Exponential backoff: base * multiplier^(attempt-1)
		delay *= math.Pow(multiplier, float64(attemptNumber-1))
	} else {
		// Linear backoff: base * attempt
		delay *= float64(attemptNumber)
	}

	return int64(delay)
}

func basicSelectBestPolicy(policies []RetryPolicy, features RetryFeatures) *RetryPolicy {
	var bestPolicy *RetryPolicy
	highestPriority := -1

	for i := range policies {
		policy := &policies[i]
		if basicPolicyMatches(*policy, features) && policy.Priority > highestPriority {
			bestPolicy = policy
			highestPriority = policy.Priority
		}
	}

	return bestPolicy
}

func abs64(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func BenchmarkBasicPolicyMatching(b *testing.B) {
	policies := []RetryPolicy{
		{ErrorPatterns: []string{"timeout", "connection_.*", "network_.*"}},
		{ErrorPatterns: []string{"5\\d\\d_.*", "server_.*"}},
		{ErrorPatterns: []string{"validation_.*", "400_.*", "401_.*"}},
		{ErrorPatterns: []string{".*"}}, // Catch-all
	}

	features := RetryFeatures{
		ErrorClass: "connection_timeout",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = basicSelectBestPolicy(policies, features)
	}
}

func TestRetryRecommendation_Structure(t *testing.T) {
	recommendation := RetryRecommendation{
		ShouldRetry:      true,
		DelayMs:          2000,
		MaxAttempts:      5,
		Confidence:       0.8,
		Method:           "rules",
		Rationale:        "Exponential backoff for timeout error",
		EstimatedSuccess: 0.7,
		NextEvaluation:   time.Now().Add(2 * time.Second),
	}

	if !recommendation.ShouldRetry {
		t.Error("expected ShouldRetry to be true")
	}

	if recommendation.DelayMs != 2000 {
		t.Errorf("expected DelayMs 2000, got %d", recommendation.DelayMs)
	}

	if recommendation.Confidence < 0.0 || recommendation.Confidence > 1.0 {
		t.Errorf("confidence should be between 0 and 1, got %.2f", recommendation.Confidence)
	}

	if recommendation.Method == "" {
		t.Error("method should not be empty")
	}
}

func TestRetryFeatures_Structure(t *testing.T) {
	features := RetryFeatures{
		JobType:           "payment",
		ErrorClass:        "timeout",
		AttemptNumber:     2,
		Queue:             "critical",
		PayloadSize:       1024,
		TimeOfDay:         14,
		SinceLastFailure:  5 * time.Minute,
		RecentFailures:    3,
		AvgProcessingTime: 2 * time.Second,
	}

	if features.JobType == "" {
		t.Error("JobType should not be empty")
	}

	if features.AttemptNumber < 1 {
		t.Error("AttemptNumber should be >= 1")
	}

	if features.TimeOfDay < 0 || features.TimeOfDay > 23 {
		t.Errorf("TimeOfDay should be 0-23, got %d", features.TimeOfDay)
	}
}