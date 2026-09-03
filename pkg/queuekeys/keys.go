// Copyright 2026 James Ross
// Package queuekeys defines the Redis key layout shared by producers,
// workers, admin tools, and public clients.
package queuekeys

import (
	"crypto/sha256"
	"encoding/hex"
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
	DefaultOrderedReadyList      = Namespace + "ordered:ready"
	DefaultOrderedActiveSet      = Namespace + "ordered:active"
	DefaultOrderedQueuePattern   = Namespace + "ordered:queue:%s"
	DefaultOrderedLeasePattern   = Namespace + "ordered:lease:%s"
)

// IsReservedQueueAlias reports names owned by terminal queue views rather
// than configurable priority queues.
func IsReservedQueueAlias(alias string) bool {
	switch strings.ToLower(strings.TrimSpace(alias)) {
	case "completed", "dead_letter", "dlq":
		return true
	default:
		return false
	}
}

// Format substitutes an identifier into a configured key pattern. Percent
// sequences other than the single %s placeholder remain literal key bytes.
func Format(pattern, identifier string) string {
	return strings.Replace(pattern, "%s", identifier, 1)
}

// ScanPattern converts a configured %s key pattern into a Redis glob pattern
// while preserving all glob metacharacters in its fixed prefix and suffix.
func ScanPattern(pattern string) string {
	prefix, suffix, ok := SplitPattern(pattern)
	if !ok {
		return escapeRedisGlob(pattern)
	}
	return escapeRedisGlob(prefix) + "*" + escapeRedisGlob(suffix)
}

func escapeRedisGlob(value string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`*`, `\*`,
		`?`, `\?`,
		`[`, `\[`,
		`]`, `\]`,
	).Replace(value)
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

// OrderingDigest maps exact ordering-key bytes to a bounded Redis-safe token.
// The original ordering key remains in the durable job envelope.
func OrderingDigest(orderingKey string) string {
	sum := sha256.Sum256([]byte(orderingKey))
	return hex.EncodeToString(sum[:])
}

// SplitPattern returns the fixed prefix and suffix around a pattern's sole %s
// placeholder. It is useful for Lua scripts that format a key after claiming
// an identifier inside Redis.
func SplitPattern(pattern string) (prefix, suffix string, ok bool) {
	if strings.Count(pattern, "%s") != 1 {
		return "", "", false
	}
	prefix, suffix, _ = strings.Cut(pattern, "%s")
	return prefix, suffix, true
}
