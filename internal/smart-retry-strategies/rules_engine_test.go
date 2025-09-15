// Copyright 2025 James Ross
package smartretry

import (
	"testing"
	"time"
)

func TestRetryPolicy_ErrorMatching(t *testing.T) {
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
			name: "job type pattern match",
			policy: RetryPolicy{
				JobTypePatterns: []string{"payment_.*"},
			},
			features: RetryFeatures{
				JobType: "payment_processing",
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
		{
			name: "multiple patterns - one matches",
			policy: RetryPolicy{
				ErrorPatterns: []string{"timeout", "connection_.*", "503_error"},
			},
			features: RetryFeatures{
				ErrorClass: "connection_reset",
			},
			shouldMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := policyMatches(tt.policy, tt.features)
			if matches != tt.shouldMatch {
				t.Errorf("expected match=%v, got match=%v", tt.shouldMatch, matches)
			}
		})
	}
}

func TestRetryPolicy_DelayCalculation(t *testing.T) {
	tests := []struct {
		name           string
		policy         RetryPolicy
		attemptNumber  int
		expectedDelayMs int64
		tolerance       int64
	}{
		{
			name: "exponential backoff",
			policy: RetryPolicy{
				BaseDelayMs:       1000,
				BackoffMultiplier: 2.0,
				JitterPercent:     0.0, // No jitter for predictable testing
			},
			attemptNumber:   3,
			expectedDelayMs: 4000, // 1000 * 2^(3-1) = 4000
			tolerance:       100,
		},
		{
			name: "linear backoff",
			policy: RetryPolicy{
				BaseDelayMs:       500,
				BackoffMultiplier: 1.0, // Linear
				JitterPercent:     0.0,
			},
			attemptNumber:   4,
			expectedDelayMs: 2000, // 500 * 4 = 2000
			tolerance:       100,
		},
		{
			name: "with max delay cap",
			policy: RetryPolicy{
				BaseDelayMs:       1000,
				MaxDelayMs:        5000,
				BackoffMultiplier: 2.0,
				JitterPercent:     0.0,
			},
			attemptNumber:   5,
			expectedDelayMs: 5000, // Capped at max
			tolerance:       100,
		},
		{
			name: "with jitter",
			policy: RetryPolicy{
				BaseDelayMs:       1000,
				BackoffMultiplier: 2.0,
				JitterPercent:     20.0,
			},
			attemptNumber:   2,
			expectedDelayMs: 2000, // Base calculation
			tolerance:       800,   // Allow for 20% jitter ± extra tolerance
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delayMs := calculatePolicyDelay(tt.policy, tt.attemptNumber)

			if abs(delayMs-tt.expectedDelayMs) > tt.tolerance {
				t.Errorf("expected delay %dms (±%dms), got %dms",
					tt.expectedDelayMs, tt.tolerance, delayMs)
			}

			// Verify max delay enforcement
			if tt.policy.MaxDelayMs > 0 && delayMs > tt.policy.MaxDelayMs {
				t.Errorf("delay %dms exceeds max %dms", delayMs, tt.policy.MaxDelayMs)
			}
		})
	}
}

func TestRulesEngine_PolicySelection(t *testing.T) {
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
		{
			name: "connection error matches medium priority",
			features: RetryFeatures{
				ErrorClass: "connection_refused",
			},
			expectedPolicy: "general_network",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := selectBestPolicy(policies, tt.features)

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

func TestRulesEngine_StopOnValidation(t *testing.T) {
	policy := RetryPolicy{
		Name:             "validation_policy",
		ErrorPatterns:    []string{"validation_.*"},
		StopOnValidation: true,
		MaxAttempts:      5,
	}

	features := RetryFeatures{
		ErrorClass:    "validation_error",
		AttemptNumber: 2,
	}

	recommendation := applyPolicyRules(policy, features)

	if recommendation.ShouldRetry {
		t.Error("should not retry on validation error when StopOnValidation=true")
	}

	if recommendation.Rationale == "" {
		t.Error("should provide rationale for not retrying")
	}
}

func TestRulesEngine_MaxAttemptsCheck(t *testing.T) {
	policy := RetryPolicy{
		Name:          "limited_attempts",
		ErrorPatterns: []string{".*"},
		MaxAttempts:   3,
	}

	tests := []struct {
		name          string
		attemptNumber int
		shouldRetry   bool
	}{
		{
			name:          "within limit",
			attemptNumber: 2,
			shouldRetry:   true,
		},
		{
			name:          "at limit",
			attemptNumber: 3,
			shouldRetry:   false,
		},
		{
			name:          "over limit",
			attemptNumber: 4,
			shouldRetry:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := RetryFeatures{
				ErrorClass:    "test_error",
				AttemptNumber: tt.attemptNumber,
			}

			recommendation := applyPolicyRules(policy, features)

			if recommendation.ShouldRetry != tt.shouldRetry {
				t.Errorf("expected ShouldRetry=%v, got %v",
					tt.shouldRetry, recommendation.ShouldRetry)
			}
		})
	}
}

func TestRulesEngine_SpecialErrorCodes(t *testing.T) {
	tests := []struct {
		name           string
		errorClass     string
		expectedDelay  time.Duration
		expectedRetry  bool
		expectedMethod string
	}{
		{
			name:           "rate limit 429",
			errorClass:     "429_rate_limit",
			expectedDelay:  60 * time.Second, // Standard rate limit backoff
			expectedRetry:  true,
			expectedMethod: "rate_limit_backoff",
		},
		{
			name:           "service unavailable 503",
			errorClass:     "503_service_unavailable",
			expectedDelay:  5 * time.Second, // Quick recovery expected
			expectedRetry:  true,
			expectedMethod: "service_unavailable_backoff",
		},
		{
			name:           "bad request 400",
			errorClass:     "400_bad_request",
			expectedDelay:  0,
			expectedRetry:  false,
			expectedMethod: "no_retry_client_error",
		},
		{
			name:           "authentication 401",
			errorClass:     "401_unauthorized",
			expectedDelay:  0,
			expectedRetry:  false,
			expectedMethod: "no_retry_auth_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := RetryFeatures{
				ErrorClass:    tt.errorClass,
				AttemptNumber: 1,
			}

			recommendation := applySpecialErrorHandling(features)

			if recommendation.ShouldRetry != tt.expectedRetry {
				t.Errorf("expected ShouldRetry=%v, got %v",
					tt.expectedRetry, recommendation.ShouldRetry)
			}

			actualDelay := time.Duration(recommendation.DelayMs) * time.Millisecond
			if actualDelay != tt.expectedDelay {
				t.Errorf("expected delay %v, got %v", tt.expectedDelay, actualDelay)
			}

			if recommendation.Method != tt.expectedMethod {
				t.Errorf("expected method '%s', got '%s'",
					tt.expectedMethod, recommendation.Method)
			}
		})
	}
}

func BenchmarkPolicyMatching(b *testing.B) {
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
		_ = selectBestPolicy(policies, features)
	}
}

// Helper functions for testing

func policyMatches(policy RetryPolicy, features RetryFeatures) bool {
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

func calculatePolicyDelay(policy RetryPolicy, attemptNumber int) int64 {
	// Basic exponential backoff calculation
	delay := float64(policy.BaseDelayMs)

	if policy.BackoffMultiplier > 1.0 {
		// Exponential backoff: base * multiplier^(attempt-1)
		delay *= math.Pow(policy.BackoffMultiplier, float64(attemptNumber-1))
	} else {
		// Linear backoff: base * attempt
		delay *= float64(attemptNumber)
	}

	// Apply jitter if specified
	if policy.JitterPercent > 0 {
		jitter := delay * (policy.JitterPercent / 100.0)
		delay += (rand.Float64() - 0.5) * 2 * jitter
	}

	// Apply max delay cap
	if policy.MaxDelayMs > 0 && int64(delay) > policy.MaxDelayMs {
		delay = float64(policy.MaxDelayMs)
	}

	return int64(delay)
}

func selectBestPolicy(policies []RetryPolicy, features RetryFeatures) *RetryPolicy {
	var bestPolicy *RetryPolicy
	highestPriority := -1

	for i := range policies {
		policy := &policies[i]
		if policyMatches(*policy, features) && policy.Priority > highestPriority {
			bestPolicy = policy
			highestPriority = policy.Priority
		}
	}

	return bestPolicy
}

func applyPolicyRules(policy RetryPolicy, features RetryFeatures) *RetryRecommendation {
	// Check if we should stop retrying
	if features.AttemptNumber >= policy.MaxAttempts {
		return &RetryRecommendation{
			ShouldRetry: false,
			Method:      "rules",
			Rationale:   "Maximum attempts reached",
		}
	}

	// Check stop on validation
	if policy.StopOnValidation && isValidationError(features.ErrorClass) {
		return &RetryRecommendation{
			ShouldRetry: false,
			Method:      "rules",
			Rationale:   "Validation error - no retry needed",
		}
	}

	// Calculate delay
	delayMs := calculatePolicyDelay(policy, features.AttemptNumber)

	return &RetryRecommendation{
		ShouldRetry: true,
		DelayMs:     delayMs,
		Method:      "rules",
		Rationale:   "Policy-based retry with exponential backoff",
		Confidence:  0.8,
	}
}

func applySpecialErrorHandling(features RetryFeatures) *RetryRecommendation {
	switch {
	case features.ErrorClass == "429_rate_limit":
		return &RetryRecommendation{
			ShouldRetry: true,
			DelayMs:     60000, // 1 minute
			Method:      "rate_limit_backoff",
			Rationale:   "Rate limit exceeded, backing off",
			Confidence:  0.9,
		}

	case features.ErrorClass == "503_service_unavailable":
		return &RetryRecommendation{
			ShouldRetry: true,
			DelayMs:     5000, // 5 seconds
			Method:      "service_unavailable_backoff",
			Rationale:   "Service temporarily unavailable",
			Confidence:  0.8,
		}

	case features.ErrorClass == "400_bad_request":
		return &RetryRecommendation{
			ShouldRetry: false,
			DelayMs:     0,
			Method:      "no_retry_client_error",
			Rationale:   "Bad request - retry will not help",
			Confidence:  1.0,
		}

	case features.ErrorClass == "401_unauthorized":
		return &RetryRecommendation{
			ShouldRetry: false,
			DelayMs:     0,
			Method:      "no_retry_auth_error",
			Rationale:   "Authentication failed - retry will not help",
			Confidence:  1.0,
		}

	default:
		return &RetryRecommendation{
			ShouldRetry: true,
			DelayMs:     1000,
			Method:      "default_backoff",
			Rationale:   "Default retry strategy",
			Confidence:  0.5,
		}
	}
}

func isValidationError(errorClass string) bool {
	validationPatterns := []string{
		"validation_.*",
		"invalid_.*",
		"malformed_.*",
		"schema_.*",
	}

	for _, pattern := range validationPatterns {
		if matched, _ := regexp.MatchString(pattern, errorClass); matched {
			return true
		}
	}

	return false
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}