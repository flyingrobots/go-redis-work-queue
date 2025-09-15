// Copyright 2025 James Ross
package smartretry

import (
	"math"
	"testing"
	"time"
)

func TestBayesianBucket_ProbabilityCalculation(t *testing.T) {
	tests := []struct {
		name        string
		successes   int
		failures    int
		expectProb  float64
		expectUpper float64
		expectLower float64
	}{
		{
			name:        "equal successes and failures",
			successes:   10,
			failures:    10,
			expectProb:  0.5,
			expectUpper: 0.7, // Approximate 95% upper bound
			expectLower: 0.3, // Approximate 95% lower bound
		},
		{
			name:        "high success rate",
			successes:   90,
			failures:    10,
			expectProb:  0.9,
			expectUpper: 0.95,
			expectLower: 0.83,
		},
		{
			name:        "low success rate",
			successes:   10,
			failures:    90,
			expectProb:  0.1,
			expectUpper: 0.17,
			expectLower: 0.05,
		},
		{
			name:        "no data",
			successes:   0,
			failures:    0,
			expectProb:  0.5, // Prior assumption
			expectUpper: 1.0,
			expectLower: 0.0,
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

			// 95% confidence intervals (approximate)
			variance := (alpha * beta) / ((alpha + beta) * (alpha + beta) * (alpha + beta + 1))
			stddev := math.Sqrt(variance)
			bucket.UpperBound = math.Min(1.0, probability+1.96*stddev)
			bucket.LowerBound = math.Max(0.0, probability-1.96*stddev)

			if math.Abs(bucket.Probability-tt.expectProb) > 0.02 {
				t.Errorf("expected probability %.2f, got %.2f", tt.expectProb, bucket.Probability)
			}

			if math.Abs(bucket.UpperBound-tt.expectUpper) > 0.05 {
				t.Errorf("expected upper bound %.2f, got %.2f", tt.expectUpper, bucket.UpperBound)
			}

			if math.Abs(bucket.LowerBound-tt.expectLower) > 0.05 {
				t.Errorf("expected lower bound %.2f, got %.2f", tt.expectLower, bucket.LowerBound)
			}
		})
	}
}

func TestBayesianModel_DelayRecommendation(t *testing.T) {
	model := &BayesianModel{
		JobType:    "payment",
		ErrorClass: "timeout",
		Buckets: []BayesianBucket{
			{DelayMinMs: 0, DelayMaxMs: 1000, Successes: 5, Failures: 15, Probability: 0.25},
			{DelayMinMs: 1000, DelayMaxMs: 5000, Successes: 15, Failures: 5, Probability: 0.75},
			{DelayMinMs: 5000, DelayMaxMs: 10000, Successes: 20, Failures: 5, Probability: 0.80},
			{DelayMinMs: 10000, DelayMaxMs: 30000, Successes: 25, Failures: 5, Probability: 0.83},
		},
		Confidence: 0.85,
	}

	tests := []struct {
		name            string
		threshold       float64
		expectedDelayMs int64
		expectFound     bool
	}{
		{
			name:            "low threshold - first bucket",
			threshold:       0.20,
			expectedDelayMs: 500, // Mid-point of first bucket
			expectFound:     true,
		},
		{
			name:            "medium threshold - third bucket",
			threshold:       0.78,
			expectedDelayMs: 7500, // Mid-point of third bucket
			expectFound:     true,
		},
		{
			name:            "high threshold - last bucket",
			threshold:       0.82,
			expectedDelayMs: 20000, // Mid-point of last bucket
			expectFound:     true,
		},
		{
			name:            "impossible threshold",
			threshold:       0.95,
			expectedDelayMs: 0,
			expectFound:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delayMs, found := findOptimalDelay(model, tt.threshold)

			if found != tt.expectFound {
				t.Errorf("expected found=%v, got found=%v", tt.expectFound, found)
			}

			if tt.expectFound {
				tolerance := int64(2000) // 2 second tolerance
				if math.Abs(float64(delayMs-tt.expectedDelayMs)) > float64(tolerance) {
					t.Errorf("expected delay %dms (±%dms), got %dms",
						tt.expectedDelayMs, tolerance, delayMs)
				}
			}
		})
	}
}

func TestBayesianModel_UpdateWithAttempt(t *testing.T) {
	model := &BayesianModel{
		JobType:    "payment",
		ErrorClass: "503_error",
		Buckets: []BayesianBucket{
			{DelayMinMs: 0, DelayMaxMs: 1000, Successes: 10, Failures: 10},
			{DelayMinMs: 1000, DelayMaxMs: 5000, Successes: 15, Failures: 5},
		},
		SampleCount: 40,
	}

	tests := []struct {
		name        string
		delayMs     int64
		success     bool
		bucketIndex int
		expectSuccess int
		expectFailure int
	}{
		{
			name:          "successful attempt in first bucket",
			delayMs:       500,
			success:       true,
			bucketIndex:   0,
			expectSuccess: 11,
			expectFailure: 10,
		},
		{
			name:          "failed attempt in second bucket",
			delayMs:       3000,
			success:       false,
			bucketIndex:   1,
			expectSuccess: 15,
			expectFailure: 6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			originalModel := cloneBayesianModel(model)

			attempt := AttemptHistory{
				DelayMs:   tt.delayMs,
				Success:   tt.success,
				Timestamp: time.Now(),
			}

			updateBayesianModel(model, attempt)

			bucket := &model.Buckets[tt.bucketIndex]
			if bucket.Successes != tt.expectSuccess {
				t.Errorf("expected %d successes, got %d", tt.expectSuccess, bucket.Successes)
			}

			if bucket.Failures != tt.expectFailure {
				t.Errorf("expected %d failures, got %d", tt.expectFailure, bucket.Failures)
			}

			if model.SampleCount != originalModel.SampleCount+1 {
				t.Errorf("expected sample count %d, got %d",
					originalModel.SampleCount+1, model.SampleCount)
			}

			if !model.LastUpdated.After(originalModel.LastUpdated) {
				t.Error("LastUpdated should have been updated")
			}
		})
	}
}

func TestBayesianModel_ConfidenceCalculation(t *testing.T) {
	tests := []struct {
		name           string
		sampleCount    int
		expectedMinConf float64
		expectedMaxConf float64
	}{
		{
			name:            "low sample count",
			sampleCount:     5,
			expectedMinConf: 0.0,
			expectedMaxConf: 0.3,
		},
		{
			name:            "medium sample count",
			sampleCount:     50,
			expectedMinConf: 0.4,
			expectedMaxConf: 0.8,
		},
		{
			name:            "high sample count",
			sampleCount:     200,
			expectedMinConf: 0.8,
			expectedMaxConf: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := &BayesianModel{
				SampleCount: tt.sampleCount,
			}

			confidence := calculateModelConfidence(model)

			if confidence < tt.expectedMinConf {
				t.Errorf("confidence %.2f below minimum %.2f",
					confidence, tt.expectedMinConf)
			}

			if confidence > tt.expectedMaxConf {
				t.Errorf("confidence %.2f above maximum %.2f",
					confidence, tt.expectedMaxConf)
			}
		})
	}
}

// Helper functions for testing

func findOptimalDelay(model *BayesianModel, threshold float64) (int64, bool) {
	for _, bucket := range model.Buckets {
		if bucket.Probability >= threshold {
			// Return mid-point of bucket
			return (bucket.DelayMinMs + bucket.DelayMaxMs) / 2, true
		}
	}
	return 0, false
}

func cloneBayesianModel(original *BayesianModel) *BayesianModel {
	clone := *original
	clone.Buckets = make([]BayesianBucket, len(original.Buckets))
	copy(clone.Buckets, original.Buckets)
	return &clone
}

func updateBayesianModel(model *BayesianModel, attempt AttemptHistory) {
	// Find appropriate bucket
	for i := range model.Buckets {
		bucket := &model.Buckets[i]
		if attempt.DelayMs >= bucket.DelayMinMs && attempt.DelayMs < bucket.DelayMaxMs {
			if attempt.Success {
				bucket.Successes++
			} else {
				bucket.Failures++
			}

			// Update probability
			alpha := float64(bucket.Successes + 1)
			beta := float64(bucket.Failures + 1)
			bucket.Probability = alpha / (alpha + beta)

			break
		}
	}

	model.SampleCount++
	model.LastUpdated = time.Now()
	model.Confidence = calculateModelConfidence(model)
}

func calculateModelConfidence(model *BayesianModel) float64 {
	// Simple confidence based on sample count
	if model.SampleCount < 10 {
		return 0.1
	}
	if model.SampleCount < 50 {
		return 0.5
	}
	if model.SampleCount < 100 {
		return 0.7
	}
	return 0.9
}

func TestBayesianBucket_ConfidenceIntervals(t *testing.T) {
	bucket := &BayesianBucket{
		Successes: 80,
		Failures:  20,
	}

	alpha := float64(bucket.Successes + 1)
	beta := float64(bucket.Failures + 1)

	// Beta distribution mean
	mean := alpha / (alpha + beta)

	// Variance calculation
	variance := (alpha * beta) / ((alpha + beta) * (alpha + beta) * (alpha + beta + 1))
	stddev := math.Sqrt(variance)

	// 95% confidence interval (approximate)
	lower := math.Max(0.0, mean-1.96*stddev)
	upper := math.Min(1.0, mean+1.96*stddev)

	if mean < 0.75 || mean > 0.85 {
		t.Errorf("expected mean around 0.8, got %.3f", mean)
	}

	if upper-lower > 0.15 {
		t.Errorf("confidence interval too wide: [%.3f, %.3f]", lower, upper)
	}
}

func BenchmarkBayesianModel_DelayRecommendation(b *testing.B) {
	model := &BayesianModel{
		JobType:    "test",
		ErrorClass: "timeout",
		Buckets: make([]BayesianBucket, 10),
	}

	// Initialize buckets with test data
	for i := range model.Buckets {
		model.Buckets[i] = BayesianBucket{
			DelayMinMs:  int64(i * 1000),
			DelayMaxMs:  int64((i + 1) * 1000),
			Successes:   10 + i*2,
			Failures:    20 - i,
			Probability: float64(10+i*2) / float64(30+i),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = findOptimalDelay(model, 0.7)
	}
}

func TestBayesianModel_EdgeCases(t *testing.T) {
	t.Run("empty model", func(t *testing.T) {
		model := &BayesianModel{
			JobType:    "test",
			ErrorClass: "error",
			Buckets:    []BayesianBucket{},
		}

		_, found := findOptimalDelay(model, 0.5)
		if found {
			t.Error("should not find delay in empty model")
		}
	})

	t.Run("zero threshold", func(t *testing.T) {
		model := &BayesianModel{
			Buckets: []BayesianBucket{
				{DelayMinMs: 0, DelayMaxMs: 1000, Probability: 0.1},
			},
		}

		delayMs, found := findOptimalDelay(model, 0.0)
		if !found {
			t.Error("should find delay with zero threshold")
		}
		if delayMs != 500 {
			t.Errorf("expected 500ms, got %dms", delayMs)
		}
	})

	t.Run("perfect success rate", func(t *testing.T) {
		bucket := &BayesianBucket{
			Successes: 100,
			Failures:  0,
		}

		alpha := float64(bucket.Successes + 1)
		beta := float64(bucket.Failures + 1)
		probability := alpha / (alpha + beta)

		if probability < 0.99 {
			t.Errorf("expected high probability, got %.3f", probability)
		}
	})
}