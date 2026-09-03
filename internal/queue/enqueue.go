// Copyright 2026 James Ross
package queue

import (
	"context"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// DefaultMaxPayloadSize is the default decoded payload limit: one mebibyte.
const DefaultMaxPayloadSize = 1 << 20

// ErrPayloadTooLarge identifies enqueue attempts rejected by the payload guard.
var ErrPayloadTooLarge = errors.New("payload exceeds maximum size")

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

// Enqueue validates and atomically appends one job to the Redis list consumed
// by workers. A non-positive limit uses DefaultMaxPayloadSize so callers that
// construct Config values directly retain the safe default.
func Enqueue(ctx context.Context, rdb redis.Cmdable, queueName string, job Job, maxPayloadSize int) error {
	if maxPayloadSize <= 0 {
		maxPayloadSize = DefaultMaxPayloadSize
	}
	if len(job.Payload) > maxPayloadSize {
		return &PayloadTooLargeError{Size: len(job.Payload), Limit: maxPayloadSize}
	}

	encoded, err := job.Marshal()
	if err != nil {
		return fmt.Errorf("marshal job: %w", err)
	}
	if err := rdb.LPush(ctx, queueName, encoded).Err(); err != nil {
		return fmt.Errorf("enqueue job: %w", err)
	}
	return nil
}
