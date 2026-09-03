// Copyright 2026 James Ross
package queuekeys

import "testing"

func TestDefaultCoreKeysMatchExistingRedisLayout(t *testing.T) {
	tests := map[string]string{
		"namespace":        Namespace,
		"high":             DefaultHighPriorityQueue,
		"low":              DefaultLowPriorityQueue,
		"processing":       DefaultProcessingListPattern,
		"heartbeat":        DefaultHeartbeatKeyPattern,
		"completed":        DefaultCompletedList,
		"dead letter":      DefaultDeadLetterList,
		"producer limiter": DefaultProducerRateLimitKey,
	}
	want := map[string]string{
		"namespace":        "jobqueue:",
		"high":             "jobqueue:high_priority",
		"low":              "jobqueue:low_priority",
		"processing":       "jobqueue:worker:%s:processing",
		"heartbeat":        "jobqueue:processing:worker:%s",
		"completed":        "jobqueue:completed",
		"dead letter":      "jobqueue:dead_letter",
		"producer limiter": "jobqueue:rate_limit:producer",
	}
	for name, got := range tests {
		if got != want[name] {
			t.Errorf("%s key = %q, want %q", name, got, want[name])
		}
	}
}

func TestFormatPatternHelpersSupportCustomLayouts(t *testing.T) {
	pattern := "custom:{worker:%s}:active"
	if got := ScanPattern(pattern); got != "custom:{worker:*}:active" {
		t.Fatalf("scan pattern = %q", got)
	}
	key := Format(pattern, "alpha:β")
	if key != "custom:{worker:alpha:β}:active" {
		t.Fatalf("formatted key = %q", key)
	}
	id, ok := Extract(pattern, key)
	if !ok || id != "alpha:β" {
		t.Fatalf("extracted (%q, %v), want (%q, true)", id, ok, "alpha:β")
	}
}
