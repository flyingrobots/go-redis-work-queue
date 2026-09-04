// Copyright 2026 James Ross
package queue

import (
	"context"
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/redis/go-redis/v9"
)

// DefaultMaxPayloadSize is the default decoded payload limit: one mebibyte.
const DefaultMaxPayloadSize = 1 << 20

var (
	// ErrPayloadTooLarge identifies enqueue attempts rejected by the payload guard.
	ErrPayloadTooLarge = errors.New("payload exceeds maximum size")
	// ErrInvalidOrderingKey identifies ordering keys that JSON cannot preserve byte-exactly.
	ErrInvalidOrderingKey = errors.New("ordering key must be valid UTF-8")
)

// PayloadTooLargeError reports the observed and configured payload sizes.
type PayloadTooLargeError struct {
	Size  int
	Limit int
}

func (e *PayloadTooLargeError) Error() string {
	return fmt.Sprintf("payload is %d bytes; maximum is %d bytes", e.Size, e.Limit)
}

// Unwrap supports errors.Is(err, ErrPayloadTooLarge).
func (e *PayloadTooLargeError) Unwrap() error {
	return ErrPayloadTooLarge
}

// ValidatePayloadSize applies the enqueue payload guard without modifying
// Redis. A non-positive limit uses DefaultMaxPayloadSize.
func ValidatePayloadSize(payload []byte, maxPayloadSize int) error {
	if maxPayloadSize <= 0 {
		maxPayloadSize = DefaultMaxPayloadSize
	}
	if len(payload) > maxPayloadSize {
		return &PayloadTooLargeError{Size: len(payload), Limit: maxPayloadSize}
	}
	return nil
}

// EncodeForEnqueue applies the payload guard and returns the durable job JSON
// without modifying Redis.
func EncodeForEnqueue(job Job, maxPayloadSize int) (string, error) {
	if !utf8.ValidString(job.OrderingKey) {
		return "", ErrInvalidOrderingKey
	}
	if err := ValidatePayloadSize(job.Payload, maxPayloadSize); err != nil {
		return "", err
	}

	encoded, err := job.Marshal()
	if err != nil {
		return "", fmt.Errorf("marshal job: %w", err)
	}
	return encoded, nil
}

// Enqueue validates and atomically appends one job to the Redis list consumed
// by workers. A non-positive limit uses DefaultMaxPayloadSize so callers that
// construct Config values directly retain the safe default.
func Enqueue(ctx context.Context, rdb redis.Cmdable, queueName string, job Job, maxPayloadSize int) error {
	return EnqueueWithOrdering(ctx, rdb, queueName, job, maxPayloadSize, DefaultOrderingLayout())
}

// EnqueueWithOrdering validates and appends one job using the supplied ordered
// key layout. Empty ordering keys retain the original single-LPUSH path.
func EnqueueWithOrdering(ctx context.Context, rdb redis.Cmdable, queueName string, job Job, maxPayloadSize int, layout OrderingLayout) error {
	encoded, err := EncodeForEnqueue(job, maxPayloadSize)
	if err != nil {
		return err
	}
	return AppendEncoded(ctx, rdb, queueName, job, encoded, layout)
}
