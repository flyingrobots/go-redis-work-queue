// Copyright 2026 James Ross
package queuekeys

import (
	"strings"
	"testing"
)

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
	if !IsOrderingDigest(OrderingDigest(key)) {
		t.Fatal("canonical ordering digest was rejected")
	}
	for _, invalid := range []string{"", strings.Repeat("a", 63), strings.Repeat("A", 64), strings.Repeat("g", 64)} {
		if IsOrderingDigest(invalid) {
			t.Fatalf("invalid ordering digest %q was accepted", invalid)
		}
	}
}

func TestIsWorkerIDRecognizesOnlyCanonicalGeneratedShape(t *testing.T) {
	for _, valid := range []string{
		"host-123-456-1a2b-0",
		"host-with-hyphens-1-9223372036854775807-0000-15",
		"host:port-42-123456789-ffff-0",
	} {
		if !IsWorkerID(valid) {
			t.Errorf("canonical worker ID %q was rejected", valid)
		}
	}
	for _, invalid := range []string{
		"",
		"w1",
		"-1-2-abcd-0",
		"host-0-2-abcd-0",
		"host-01-2-abcd-0",
		"host-1-0-abcd-0",
		"host-1-02-abcd-0",
		"host-1-2-ABCd-0",
		"host-1-2-abc-0",
		"host-1-2-abcz-0",
		"host-1-2-abcd-00",
		"host-1-2-abcd--1",
	} {
		if IsWorkerID(invalid) {
			t.Errorf("non-canonical worker ID %q was accepted", invalid)
		}
	}
}

func TestDLQGenerationKeyFollowsConfiguredList(t *testing.T) {
	if got, want := DLQGenerationKey("tenant:{jobs}:dead"), "tenant:{jobs}:dead:generation"; got != want {
		t.Fatalf("DLQ generation key = %q, want %q", got, want)
	}
}

func TestOrderedClaimsKeyFollowsConfiguredActiveSet(t *testing.T) {
	if got, want := OrderedClaimsKey("tenant:{jobs}:ordered:active"), "tenant:{jobs}:ordered:active:claims"; got != want {
		t.Fatalf("ordered claims key = %q, want %q", got, want)
	}
}

func TestPatternsOverlap(t *testing.T) {
	tests := []struct {
		name   string
		first  string
		second string
		want   bool
	}{
		{name: "identical", first: "tenant:%s", second: "tenant:%s", want: true},
		{name: "nested prefix", first: "tenant:%s", second: "tenant:heartbeat:%s", want: true},
		{name: "nested suffix", first: "%s:state", second: "%s:heartbeat:state", want: true},
		{name: "nested prefix and suffix", first: "tenant:%s:state", second: "tenant:worker:%s:heartbeat:state", want: true},
		{name: "disjoint prefix", first: "processing:%s", second: "heartbeat:%s", want: false},
		{name: "disjoint suffix", first: "%s:processing", second: "%s:heartbeat", want: false},
		{name: "disjoint prefix and suffix", first: "a:%s:x", second: "b:%s:y", want: false},
		{name: "invalid pattern", first: "tenant", second: "tenant:%s", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PatternsOverlap(tt.first, tt.second); got != tt.want {
				t.Fatalf("PatternsOverlap(%q, %q) = %v, want %v", tt.first, tt.second, got, tt.want)
			}
		})
	}
}

func TestMatchesOrderingDigestRequiresCanonicalGeneratedKey(t *testing.T) {
	pattern := "custom:ordered:%s:jobs"
	digest := OrderingDigest("account:42")
	tests := []struct {
		name string
		key  string
		want bool
	}{
		{name: "generated digest", key: Format(pattern, digest), want: true},
		{name: "uppercase digest", key: Format(pattern, strings.ToUpper(digest))},
		{name: "non-hex token", key: Format(pattern, strings.Repeat("g", 64))},
		{name: "short token", key: Format(pattern, digest[:63])},
		{name: "wrong suffix", key: "custom:ordered:" + digest + ":leases"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchesOrderingDigest(pattern, tt.key); got != tt.want {
				t.Fatalf("MatchesOrderingDigest(%q, %q) = %v, want %v", pattern, tt.key, got, tt.want)
			}
		})
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
