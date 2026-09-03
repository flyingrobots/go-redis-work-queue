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
		"ordered ready":    DefaultOrderedReadyList,
		"ordered active":   DefaultOrderedActiveSet,
		"ordered queue":    DefaultOrderedQueuePattern,
		"ordered lease":    DefaultOrderedLeasePattern,
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
		"ordered ready":    "jobqueue:ordered:ready",
		"ordered active":   "jobqueue:ordered:active",
		"ordered queue":    "jobqueue:ordered:queue:%s",
		"ordered lease":    "jobqueue:ordered:lease:%s",
	}
	for name, got := range tests {
		if got != want[name] {
			t.Errorf("%s key = %q, want %q", name, got, want[name])
		}
	}
}

func TestOrderingDigestIsStableAndRedisSafe(t *testing.T) {
	key := "repo:{main}:目录/hello.go"
	const want = "043a1248cb072ea65d9eff97c65664eb2709d0345e9a77a6b42b5a39b707b297"
	if got := OrderingDigest(key); got != want {
		t.Fatalf("digest = %q, want %q", got, want)
	}
	if got := len(OrderingDigest(key)); got != 64 {
		t.Fatalf("digest length = %d, want 64", got)
	}
}

func TestSplitPattern(t *testing.T) {
	prefix, suffix, ok := SplitPattern("custom:{ordered}:%s:jobs")
	if !ok || prefix != "custom:{ordered}:" || suffix != ":jobs" {
		t.Fatalf("split = (%q, %q, %v)", prefix, suffix, ok)
	}
	if _, _, ok := SplitPattern("missing-placeholder"); ok {
		t.Fatal("pattern without placeholder unexpectedly accepted")
	}
	if _, _, ok := SplitPattern("%s:twice:%s"); ok {
		t.Fatal("pattern with two placeholders unexpectedly accepted")
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

func TestFormatTreatsOtherPercentSequencesLiterally(t *testing.T) {
	pattern := "custom:100%:%s:%q"
	key := Format(pattern, "worker-1")
	if key != "custom:100%:worker-1:%q" {
		t.Fatalf("formatted key = %q", key)
	}
	if got := ScanPattern(pattern); got != "custom:100%:*:%q" {
		t.Fatalf("scan pattern = %q", got)
	}
	identifier, ok := Extract(pattern, key)
	if !ok || identifier != "worker-1" {
		t.Fatalf("extracted (%q, %v), want (%q, true)", identifier, ok, "worker-1")
	}
}

func TestScanPatternEscapesLiteralRedisGlobCharacters(t *testing.T) {
	pattern := `tenant[1]*?:ordered:\%s:suffix[?]`
	want := `tenant\[1\]\*\?:ordered:\\*:suffix\[\?\]`
	if got := ScanPattern(pattern); got != want {
		t.Fatalf("scan pattern = %q, want %q", got, want)
	}
}
