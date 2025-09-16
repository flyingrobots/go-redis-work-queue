// Copyright 2025 James Ross
package exactlyonce

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrIdempotencyKeyExists(t *testing.T) {
	t.Run("with tenant", func(t *testing.T) {
		err := &ErrIdempotencyKeyExists{
			Key:           "test-key",
			ExistingValue: "cached-result",
			QueueName:     "test-queue",
			TenantID:      "test-tenant",
		}

		expected := "idempotency key 'test-key' already exists in queue 'test-queue' for tenant 'test-tenant'"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("without tenant", func(t *testing.T) {
		err := &ErrIdempotencyKeyExists{
			Key:       "test-key",
			QueueName: "test-queue",
		}

		expected := "idempotency key 'test-key' already exists in queue 'test-queue'"
		assert.Equal(t, expected, err.Error())
	})
}

func TestErrIdempotencyKeyNotFound(t *testing.T) {
	t.Run("with tenant", func(t *testing.T) {
		err := &ErrIdempotencyKeyNotFound{
			Key:       "test-key",
			QueueName: "test-queue",
			TenantID:  "test-tenant",
		}

		expected := "idempotency key 'test-key' not found in queue 'test-queue' for tenant 'test-tenant'"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("without tenant", func(t *testing.T) {
		err := &ErrIdempotencyKeyNotFound{
			Key:       "test-key",
			QueueName: "test-queue",
		}

		expected := "idempotency key 'test-key' not found in queue 'test-queue'"
		assert.Equal(t, expected, err.Error())
	})
}

func TestErrStorageUnavailable(t *testing.T) {
	t.Run("with cause", func(t *testing.T) {
		cause := errors.New("connection refused")
		err := &ErrStorageUnavailable{
			StorageType: "redis",
			Cause:       cause,
		}

		expected := "storage backend 'redis' is unavailable: connection refused"
		assert.Equal(t, expected, err.Error())
		assert.Equal(t, cause, errors.Unwrap(err))
	})

	t.Run("without cause", func(t *testing.T) {
		err := &ErrStorageUnavailable{
			StorageType: "memory",
		}

		expected := "storage backend 'memory' is unavailable"
		assert.Equal(t, expected, err.Error())
		assert.Nil(t, errors.Unwrap(err))
	})
}

func TestErrInvalidConfiguration(t *testing.T) {
	err := &ErrInvalidConfiguration{
		Field:  "default_ttl",
		Value:  -1,
		Reason: "must be positive",
	}

	expected := "invalid configuration for field 'default_ttl' (value: -1): must be positive"
	assert.Equal(t, expected, err.Error())
}

func TestErrOutboxEventNotFound(t *testing.T) {
	err := &ErrOutboxEventNotFound{
		EventID: "event-123",
	}

	expected := "outbox event 'event-123' not found"
	assert.Equal(t, expected, err.Error())
}

func TestErrPublisherNotFound(t *testing.T) {
	err := &ErrPublisherNotFound{
		PublisherName: "kafka-publisher",
	}

	expected := "publisher 'kafka-publisher' not found"
	assert.Equal(t, expected, err.Error())
}

func TestErrMaxRetriesExceeded(t *testing.T) {
	t.Run("with last error", func(t *testing.T) {
		lastError := errors.New("network timeout")
		err := &ErrMaxRetriesExceeded{
			Operation:    "publish_event",
			MaxRetries:   3,
			CurrentRetry: 3,
			LastError:    lastError,
		}

		expected := "max retries (3) exceeded for operation 'publish_event' at retry 3: network timeout"
		assert.Equal(t, expected, err.Error())
		assert.Equal(t, lastError, errors.Unwrap(err))
	})

	t.Run("without last error", func(t *testing.T) {
		err := &ErrMaxRetriesExceeded{
			Operation:    "publish_event",
			MaxRetries:   3,
			CurrentRetry: 3,
		}

		expected := "max retries (3) exceeded for operation 'publish_event' at retry 3"
		assert.Equal(t, expected, err.Error())
		assert.Nil(t, errors.Unwrap(err))
	})
}

func TestErrTransactionFailed(t *testing.T) {
	t.Run("with cause", func(t *testing.T) {
		cause := errors.New("deadlock detected")
		err := &ErrTransactionFailed{
			Operation: "store_outbox_event",
			Cause:     cause,
		}

		expected := "transaction failed for operation 'store_outbox_event': deadlock detected"
		assert.Equal(t, expected, err.Error())
		assert.Equal(t, cause, errors.Unwrap(err))
	})

	t.Run("without cause", func(t *testing.T) {
		err := &ErrTransactionFailed{
			Operation: "store_outbox_event",
		}

		expected := "transaction failed for operation 'store_outbox_event'"
		assert.Equal(t, expected, err.Error())
		assert.Nil(t, errors.Unwrap(err))
	})
}

func TestErrInvalidIdempotencyKey(t *testing.T) {
	err := &ErrInvalidIdempotencyKey{
		Key:    "invalid@key",
		Reason: "contains invalid characters",
	}

	expected := "invalid idempotency key 'invalid@key': contains invalid characters"
	assert.Equal(t, expected, err.Error())
}

func TestErrStorageCorruption(t *testing.T) {
	err := &ErrStorageCorruption{
		Key:         "corrupt-key",
		StorageType: "redis",
		Details:     "invalid JSON format",
	}

	expected := "storage corruption detected in redis for key 'corrupt-key': invalid JSON format"
	assert.Equal(t, expected, err.Error())
}

func TestErrCircuitBreakerOpen(t *testing.T) {
	err := &ErrCircuitBreakerOpen{
		Operation: "redis_check",
		Duration:  "30s",
	}

	expected := "circuit breaker is open for operation 'redis_check' (for 30s)"
	assert.Equal(t, expected, err.Error())
}

func TestErrTenantNotAllowed(t *testing.T) {
	err := &ErrTenantNotAllowed{
		TenantID:  "blocked-tenant",
		Operation: "idempotency_check",
	}

	expected := "tenant 'blocked-tenant' is not allowed to perform operation 'idempotency_check'"
	assert.Equal(t, expected, err.Error())
}

func TestErrQuotaExceeded(t *testing.T) {
	t.Run("with tenant", func(t *testing.T) {
		err := &ErrQuotaExceeded{
			TenantID:  "quota-tenant",
			QueueName: "processing-queue",
			Quota:     1000,
			Current:   1001,
			QuotaType: "requests",
		}

		expected := "requests quota exceeded for tenant 'quota-tenant' in queue 'processing-queue': 1001/1000"
		assert.Equal(t, expected, err.Error())
	})

	t.Run("without tenant", func(t *testing.T) {
		err := &ErrQuotaExceeded{
			QueueName: "processing-queue",
			Quota:     500,
			Current:   501,
			QuotaType: "storage",
		}

		expected := "storage quota exceeded for queue 'processing-queue': 501/500"
		assert.Equal(t, expected, err.Error())
	})
}

func TestCommonErrors(t *testing.T) {
	// Test that common error variables are defined and have expected messages
	assert.Equal(t, "idempotency checking is disabled", ErrIdempotencyDisabled.Error())
	assert.Equal(t, "outbox pattern is disabled", ErrOutboxDisabled.Error())
	assert.Equal(t, "metrics collection is disabled", ErrMetricsDisabled.Error())
	assert.Equal(t, "idempotency key cannot be empty", ErrEmptyKey.Error())
	assert.Equal(t, "queue name cannot be empty", ErrEmptyQueueName.Error())
	assert.Equal(t, "TTL must be positive", ErrInvalidTTL.Error())
}

func TestErrorTypes(t *testing.T) {
	// Test that all custom errors implement the error interface
	var _ error = &ErrIdempotencyKeyExists{}
	var _ error = &ErrIdempotencyKeyNotFound{}
	var _ error = &ErrStorageUnavailable{}
	var _ error = &ErrInvalidConfiguration{}
	var _ error = &ErrOutboxEventNotFound{}
	var _ error = &ErrPublisherNotFound{}
	var _ error = &ErrMaxRetriesExceeded{}
	var _ error = &ErrTransactionFailed{}
	var _ error = &ErrInvalidIdempotencyKey{}
	var _ error = &ErrStorageCorruption{}
	var _ error = &ErrCircuitBreakerOpen{}
	var _ error = &ErrTenantNotAllowed{}
	var _ error = &ErrQuotaExceeded{}
}

func TestErrorWrapping(t *testing.T) {
	t.Run("ErrStorageUnavailable wraps cause", func(t *testing.T) {
		cause := errors.New("connection failed")
		err := &ErrStorageUnavailable{
			StorageType: "redis",
			Cause:       cause,
		}

		assert.True(t, errors.Is(err, cause))
		assert.Equal(t, cause, errors.Unwrap(err))
	})

	t.Run("ErrMaxRetriesExceeded wraps last error", func(t *testing.T) {
		lastError := errors.New("timeout")
		err := &ErrMaxRetriesExceeded{
			Operation: "test",
			LastError: lastError,
		}

		assert.True(t, errors.Is(err, lastError))
		assert.Equal(t, lastError, errors.Unwrap(err))
	})

	t.Run("ErrTransactionFailed wraps cause", func(t *testing.T) {
		cause := errors.New("constraint violation")
		err := &ErrTransactionFailed{
			Operation: "insert",
			Cause:     cause,
		}

		assert.True(t, errors.Is(err, cause))
		assert.Equal(t, cause, errors.Unwrap(err))
	})
}

// Test error creation helpers (if they existed)
func TestErrorCreationPatterns(t *testing.T) {
	t.Run("consistent error formatting", func(t *testing.T) {
		// Test that similar errors follow consistent formatting patterns
		keyExists := &ErrIdempotencyKeyExists{
			Key:       "test",
			QueueName: "queue",
		}
		keyNotFound := &ErrIdempotencyKeyNotFound{
			Key:       "test",
			QueueName: "queue",
		}

		// Both should mention the key and queue in a consistent way
		assert.Contains(t, keyExists.Error(), "test")
		assert.Contains(t, keyExists.Error(), "queue")
		assert.Contains(t, keyNotFound.Error(), "test")
		assert.Contains(t, keyNotFound.Error(), "queue")
	})
}