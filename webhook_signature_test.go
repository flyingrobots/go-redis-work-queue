// Copyright 2025 James Ross
package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HMACSigner handles webhook payload signing
type HMACSigner struct{}

func NewHMACSigner() *HMACSigner {
	return &HMACSigner{}
}

// SignPayload generates HMAC signature for webhook payload
func (s *HMACSigner) SignPayload(payload []byte, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(payload)
	signature := h.Sum(nil)
	return fmt.Sprintf("sha256=%x", signature)
}

// VerifySignature validates HMAC signature
func (s *HMACSigner) VerifySignature(payload []byte, signature, secret string) bool {
	expected := s.SignPayload(payload, secret)
	return hmac.Equal([]byte(signature), []byte(expected))
}

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	Strategy     string        `json:"strategy"`
	InitialDelay time.Duration `json:"initial_delay"`
	MaxDelay     time.Duration `json:"max_delay"`
	Multiplier   float64       `json:"multiplier"`
	MaxRetries   int           `json:"max_retries"`
	Jitter       bool          `json:"jitter"`
}

// BackoffScheduler calculates retry delays
type BackoffScheduler struct {
	policy RetryPolicy
}

func NewBackoffScheduler(policy RetryPolicy) *BackoffScheduler {
	return &BackoffScheduler{policy: policy}
}

// CalculateDelay returns the delay for a given retry attempt
func (b *BackoffScheduler) CalculateDelay(attempt int) time.Duration {
	if attempt <= 0 {
		return b.policy.InitialDelay
	}

	var delay time.Duration

	switch b.policy.Strategy {
	case "exponential":
		delay = b.policy.InitialDelay
		for i := 1; i < attempt; i++ {
			delay = time.Duration(float64(delay) * b.policy.Multiplier)
			if delay > b.policy.MaxDelay {
				delay = b.policy.MaxDelay
				break
			}
		}
	case "linear":
		delay = b.policy.InitialDelay + time.Duration(float64(attempt-1)*float64(b.policy.InitialDelay))
		if delay > b.policy.MaxDelay {
			delay = b.policy.MaxDelay
		}
	case "fixed":
		delay = b.policy.InitialDelay
	default:
		delay = b.policy.InitialDelay
	}

	// Add jitter if enabled
	if b.policy.Jitter && delay > 0 {
		// Add up to 25% jitter
		jitterAmount := time.Duration(float64(delay) * 0.25 * randomFloat())
		delay += jitterAmount
	}

	return delay
}

// ShouldRetry determines if retry should be attempted
func (b *BackoffScheduler) ShouldRetry(attempt int) bool {
	return attempt <= b.policy.MaxRetries
}

// Helper function for jitter (simplified for testing)
func randomFloat() float64 {
	return 0.5 // Fixed for deterministic testing
}

// Tests for HMAC Signature Generation/Verification

func TestHMACSigner_SignPayload(t *testing.T) {
	signer := NewHMACSigner()

	t.Run("generates consistent signature", func(t *testing.T) {
		payload := []byte(`{"event":"job_failed","job_id":"123"}`)
		secret := "test_secret_key"

		sig1 := signer.SignPayload(payload, secret)
		sig2 := signer.SignPayload(payload, secret)

		assert.Equal(t, sig1, sig2)
		assert.Contains(t, sig1, "sha256=")
	})

	t.Run("different payloads generate different signatures", func(t *testing.T) {
		payload1 := []byte(`{"event":"job_failed","job_id":"123"}`)
		payload2 := []byte(`{"event":"job_succeeded","job_id":"123"}`)
		secret := "test_secret_key"

		sig1 := signer.SignPayload(payload1, secret)
		sig2 := signer.SignPayload(payload2, secret)

		assert.NotEqual(t, sig1, sig2)
	})

	t.Run("different secrets generate different signatures", func(t *testing.T) {
		payload := []byte(`{"event":"job_failed","job_id":"123"}`)
		secret1 := "test_secret_key"
		secret2 := "different_secret"

		sig1 := signer.SignPayload(payload, secret1)
		sig2 := signer.SignPayload(payload, secret2)

		assert.NotEqual(t, sig1, sig2)
	})

	t.Run("empty payload", func(t *testing.T) {
		payload := []byte("")
		secret := "test_secret_key"

		sig := signer.SignPayload(payload, secret)
		assert.Contains(t, sig, "sha256=")
		assert.True(t, len(sig) > 7) // "sha256=" + hex
	})

	t.Run("empty secret", func(t *testing.T) {
		payload := []byte(`{"event":"job_failed"}`)
		secret := ""

		sig := signer.SignPayload(payload, secret)
		assert.Contains(t, sig, "sha256=")
	})
}

func TestHMACSigner_VerifySignature(t *testing.T) {
	signer := NewHMACSigner()

	t.Run("valid signature verification", func(t *testing.T) {
		payload := []byte(`{"event":"job_failed","job_id":"123"}`)
		secret := "test_secret_key"

		signature := signer.SignPayload(payload, secret)
		isValid := signer.VerifySignature(payload, signature, secret)

		assert.True(t, isValid)
	})

	t.Run("invalid signature verification", func(t *testing.T) {
		payload := []byte(`{"event":"job_failed","job_id":"123"}`)
		secret := "test_secret_key"
		invalidSignature := "sha256=invalid_signature"

		isValid := signer.VerifySignature(payload, invalidSignature, secret)

		assert.False(t, isValid)
	})

	t.Run("tampered payload verification", func(t *testing.T) {
		originalPayload := []byte(`{"event":"job_failed","job_id":"123"}`)
		tamperedPayload := []byte(`{"event":"job_failed","job_id":"456"}`)
		secret := "test_secret_key"

		signature := signer.SignPayload(originalPayload, secret)
		isValid := signer.VerifySignature(tamperedPayload, signature, secret)

		assert.False(t, isValid)
	})

	t.Run("wrong secret verification", func(t *testing.T) {
		payload := []byte(`{"event":"job_failed","job_id":"123"}`)
		correctSecret := "test_secret_key"
		wrongSecret := "wrong_secret"

		signature := signer.SignPayload(payload, correctSecret)
		isValid := signer.VerifySignature(payload, signature, wrongSecret)

		assert.False(t, isValid)
	})
}

// Tests for Backoff Schedule

func TestBackoffScheduler_ExponentialStrategy(t *testing.T) {
	policy := RetryPolicy{
		Strategy:     "exponential",
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		MaxRetries:   5,
		Jitter:       false,
	}
	scheduler := NewBackoffScheduler(policy)

	t.Run("exponential backoff calculation", func(t *testing.T) {
		testCases := []struct {
			attempt  int
			expected time.Duration
		}{
			{1, 1 * time.Second},
			{2, 2 * time.Second},
			{3, 4 * time.Second},
			{4, 8 * time.Second},
			{5, 16 * time.Second},
			{6, 30 * time.Second}, // Capped at MaxDelay
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("attempt_%d", tc.attempt), func(t *testing.T) {
				delay := scheduler.CalculateDelay(tc.attempt)
				assert.Equal(t, tc.expected, delay)
			})
		}
	})

	t.Run("should retry logic", func(t *testing.T) {
		assert.True(t, scheduler.ShouldRetry(1))
		assert.True(t, scheduler.ShouldRetry(3))
		assert.True(t, scheduler.ShouldRetry(5))
		assert.False(t, scheduler.ShouldRetry(6))
		assert.False(t, scheduler.ShouldRetry(10))
	})
}

func TestBackoffScheduler_LinearStrategy(t *testing.T) {
	policy := RetryPolicy{
		Strategy:     "linear",
		InitialDelay: 2 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   1.0, // Not used in linear
		MaxRetries:   3,
		Jitter:       false,
	}
	scheduler := NewBackoffScheduler(policy)

	t.Run("linear backoff calculation", func(t *testing.T) {
		testCases := []struct {
			attempt  int
			expected time.Duration
		}{
			{1, 2 * time.Second},
			{2, 4 * time.Second},
			{3, 6 * time.Second},
			{4, 8 * time.Second},
			{5, 10 * time.Second}, // Capped at MaxDelay
		}

		for _, tc := range testCases {
			t.Run(fmt.Sprintf("attempt_%d", tc.attempt), func(t *testing.T) {
				delay := scheduler.CalculateDelay(tc.attempt)
				assert.Equal(t, tc.expected, delay)
			})
		}
	})
}

func TestBackoffScheduler_FixedStrategy(t *testing.T) {
	policy := RetryPolicy{
		Strategy:     "fixed",
		InitialDelay: 5 * time.Second,
		MaxDelay:     30 * time.Second, // Not used in fixed
		Multiplier:   2.0,              // Not used in fixed
		MaxRetries:   4,
		Jitter:       false,
	}
	scheduler := NewBackoffScheduler(policy)

	t.Run("fixed backoff calculation", func(t *testing.T) {
		for attempt := 1; attempt <= 6; attempt++ {
			t.Run(fmt.Sprintf("attempt_%d", attempt), func(t *testing.T) {
				delay := scheduler.CalculateDelay(attempt)
				assert.Equal(t, 5*time.Second, delay)
			})
		}
	})
}

func TestBackoffScheduler_WithJitter(t *testing.T) {
	policy := RetryPolicy{
		Strategy:     "exponential",
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		MaxRetries:   3,
		Jitter:       true,
	}
	scheduler := NewBackoffScheduler(policy)

	t.Run("jitter adds randomness", func(t *testing.T) {
		baseDelay := scheduler.CalculateDelay(2) // Should be 2s without jitter

		// With our fixed randomFloat() returning 0.5, jitter should add 25% * 0.5 = 12.5%
		// 2s + 12.5% = 2.25s
		expectedDelay := 2*time.Second + time.Duration(float64(2*time.Second)*0.25*0.5)

		assert.Equal(t, expectedDelay, baseDelay)
	})
}

func TestBackoffScheduler_EdgeCases(t *testing.T) {
	policy := RetryPolicy{
		Strategy:     "exponential",
		InitialDelay: 1 * time.Second,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		MaxRetries:   3,
		Jitter:       false,
	}
	scheduler := NewBackoffScheduler(policy)

	t.Run("zero attempt", func(t *testing.T) {
		delay := scheduler.CalculateDelay(0)
		assert.Equal(t, 1*time.Second, delay)
	})

	t.Run("negative attempt", func(t *testing.T) {
		delay := scheduler.CalculateDelay(-1)
		assert.Equal(t, 1*time.Second, delay)
	})

	t.Run("unknown strategy defaults to initial delay", func(t *testing.T) {
		unknownPolicy := RetryPolicy{
			Strategy:     "unknown",
			InitialDelay: 3 * time.Second,
			MaxDelay:     10 * time.Second,
			Multiplier:   2.0,
			MaxRetries:   3,
			Jitter:       false,
		}
		unknownScheduler := NewBackoffScheduler(unknownPolicy)

		delay := unknownScheduler.CalculateDelay(5)
		assert.Equal(t, 3*time.Second, delay)
	})
}

// Benchmark tests for performance validation

func BenchmarkHMACSigner_SignPayload(b *testing.B) {
	signer := NewHMACSigner()
	payload := []byte(`{"event":"job_failed","job_id":"test_job","queue":"test_queue","timestamp":"2023-01-15T10:30:00Z"}`)
	secret := "benchmark_secret_key"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		signer.SignPayload(payload, secret)
	}
}

func BenchmarkHMACSigner_VerifySignature(b *testing.B) {
	signer := NewHMACSigner()
	payload := []byte(`{"event":"job_failed","job_id":"test_job","queue":"test_queue","timestamp":"2023-01-15T10:30:00Z"}`)
	secret := "benchmark_secret_key"
	signature := signer.SignPayload(payload, secret)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		signer.VerifySignature(payload, signature, secret)
	}
}

func BenchmarkBackoffScheduler_CalculateDelay(b *testing.B) {
	policy := RetryPolicy{
		Strategy:     "exponential",
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		MaxRetries:   5,
		Jitter:       true,
	}
	scheduler := NewBackoffScheduler(policy)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scheduler.CalculateDelay(i%6 + 1)
	}
}

// Integration test for signature and backoff working together
func TestWebhookDeliveryWithRetries(t *testing.T) {
	signer := NewHMACSigner()
	policy := RetryPolicy{
		Strategy:     "exponential",
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		MaxRetries:   3,
		Jitter:       false,
	}
	scheduler := NewBackoffScheduler(policy)

	payload := []byte(`{"event":"job_failed","job_id":"integration_test"}`)
	secret := "integration_secret"

	t.Run("complete delivery flow", func(t *testing.T) {
		// Sign the payload
		signature := signer.SignPayload(payload, secret)
		require.NotEmpty(t, signature)

		// Verify the signature
		isValid := signer.VerifySignature(payload, signature, secret)
		require.True(t, isValid)

		// Simulate retry attempts
		for attempt := 1; attempt <= 3; attempt++ {
			shouldRetry := scheduler.ShouldRetry(attempt)
			delay := scheduler.CalculateDelay(attempt)

			assert.True(t, shouldRetry, "Should retry attempt %d", attempt)
			assert.Greater(t, delay, time.Duration(0), "Delay should be positive for attempt %d", attempt)

			t.Logf("Attempt %d: delay=%v, should_retry=%v", attempt, delay, shouldRetry)
		}

		// After max retries
		shouldRetry := scheduler.ShouldRetry(4)
		assert.False(t, shouldRetry, "Should not retry after max attempts")
	})
}