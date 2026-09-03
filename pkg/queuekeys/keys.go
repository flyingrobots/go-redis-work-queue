// Copyright 2026 James Ross
// Package queuekeys defines the Redis key layout shared by producers,
// workers, admin tools, and public clients.
package queuekeys

import (
	"fmt"
	"strings"
)

const (
	// Namespace prefixes the default queue-owned Redis keys.
	Namespace = "jobqueue:"

	DefaultHighPriorityQueue     = Namespace + "high_priority"
	DefaultLowPriorityQueue      = Namespace + "low_priority"
	DefaultProcessingListPattern = Namespace + "worker:%s:processing"
	DefaultHeartbeatKeyPattern   = Namespace + "processing:worker:%s"
	DefaultCompletedList         = Namespace + "completed"
	DefaultDeadLetterList        = Namespace + "dead_letter"
	DefaultProducerRateLimitKey  = Namespace + "rate_limit:producer"
)

// Format substitutes an identifier into a configured key pattern.
func Format(pattern, identifier string) string {
	return fmt.Sprintf(pattern, identifier)
}

// ScanPattern converts configured %s key patterns into Redis glob patterns.
func ScanPattern(pattern string) string {
	return strings.ReplaceAll(pattern, "%s", "*")
}

// Extract returns the identifier embedded in a key generated from pattern.
// Patterns with anything other than one %s placeholder are not reversible.
func Extract(pattern, key string) (string, bool) {
	if strings.Count(pattern, "%s") != 1 {
		return "", false
	}
	prefix, suffix, _ := strings.Cut(pattern, "%s")
	if !strings.HasPrefix(key, prefix) || !strings.HasSuffix(key, suffix) {
		return "", false
	}
	end := len(key) - len(suffix)
	if end < len(prefix) {
		return "", false
	}
	return key[len(prefix):end], true
}
