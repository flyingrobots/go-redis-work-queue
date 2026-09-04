// Copyright 2026 James Ross
// Package queuekeys defines the Redis key layout shared by producers,
// workers, admin tools, and public clients.
package queuekeys

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
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

const (
	dlqGenerationSuffix = ":generation"
	orderedClaimsSuffix = ":claims"
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

// PatternsOverlap reports whether two valid single-placeholder patterns can
// generate the same key for some non-empty identifiers. Because identifiers
// are unbounded, compatible fixed prefixes and compatible fixed suffixes are
// both necessary and sufficient for the generated keyspaces to intersect.
func PatternsOverlap(first, second string) bool {
	firstPrefix, firstSuffix, firstOK := SplitPattern(first)
	secondPrefix, secondSuffix, secondOK := SplitPattern(second)
	if !firstOK || !secondOK {
		return false
	}
	prefixesCompatible := strings.HasPrefix(firstPrefix, secondPrefix) ||
		strings.HasPrefix(secondPrefix, firstPrefix)
	suffixesCompatible := strings.HasSuffix(firstSuffix, secondSuffix) ||
		strings.HasSuffix(secondSuffix, firstSuffix)
	return prefixesCompatible && suffixesCompatible
}

// MatchesOrderingDigest reports whether key is produced by pattern with the
// canonical lowercase SHA-256 token used for an ordering key.
func MatchesOrderingDigest(pattern, key string) bool {
	identifier, ok := Extract(pattern, key)
	return ok && IsOrderingDigest(identifier)
}

// IsOrderingDigest reports whether value is a canonical lowercase SHA-256
// token produced by OrderingDigest.
func IsOrderingDigest(value string) bool {
	return len(value) == sha256.Size*2 && isLowerHex(value)
}

// IsWorkerID reports whether value has the canonical identifier shape emitted
// by Worker: hostname-pid-startUnixNano-randomHex-index. Parsing from the right
// preserves hostnames containing hyphens.
func IsWorkerID(value string) bool {
	parts := strings.Split(value, "-")
	if len(parts) < 5 {
		return false
	}
	suffix := len(parts) - 4
	if strings.Join(parts[:suffix], "-") == "" {
		return false
	}
	return isCanonicalUnsigned(parts[suffix], false) &&
		isCanonicalUnsigned(parts[suffix+1], false) &&
		len(parts[suffix+2]) == 4 && isLowerHex(parts[suffix+2]) &&
		isCanonicalUnsigned(parts[suffix+3], true)
}

func isLowerHex(value string) bool {
	for _, char := range value {
		if (char < '0' || char > '9') && (char < 'a' || char > 'f') {
			return false
		}
	}
	return true
}

func isCanonicalUnsigned(value string, allowZero bool) bool {
	if value == "" || (len(value) > 1 && value[0] == '0') {
		return false
	}
	number, err := strconv.ParseUint(value, 10, 64)
	return err == nil && (allowZero || number > 0)
}

// DLQGenerationKey returns the metadata key that advances on every supported
// mutation of a dead-letter list. The list length is also included in selection
// versions so unsupported size-changing writes invalidate existing handles.
func DLQGenerationKey(deadLetterList string) string {
	return deadLetterList + dlqGenerationSuffix
}

// OrderedClaimsKey returns the hash that durably maps worker processing lists
// to their in-flight ordering digest. Deriving it from the active-set key keeps
// custom ordered layouts self-contained without another configuration field.
func OrderedClaimsKey(orderedActiveSet string) string {
	return orderedActiveSet + orderedClaimsSuffix
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
