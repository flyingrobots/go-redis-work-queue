// Copyright 2026 James Ross
package queue

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestEnqueueRoundTripsPayload(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	payload := append([]byte("hello, 世界\nembedded JSON: {\"nested\":[1,2,3]}\n"), 0x00, 0xff)
	job := NewJob("payload-job", "", 0, "customer-priority", "trace", "span")
	job.Payload = payload
	job.PayloadSchema = "example.message.v1"

	if err := Enqueue(context.Background(), rdb, "jobs", job, len(payload)); err != nil {
		t.Fatal(err)
	}
	encoded, err := rdb.RPop(context.Background(), "jobs").Result()
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalJob(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got.Payload, payload) {
		t.Fatalf("payload changed in Redis round-trip:\nwant: %v\n got: %v", payload, got.Payload)
	}
	if got.PayloadSchema != job.PayloadSchema {
		t.Fatalf("payload schema changed: want %q, got %q", job.PayloadSchema, got.PayloadSchema)
	}
	if got.Priority != "customer-priority" {
		t.Fatalf("caller-defined priority changed: got %q", got.Priority)
	}
}

func TestEnqueuePayloadSizeBoundary(t *testing.T) {
	tests := []struct {
		name       string
		payload    []byte
		wantErr    bool
		wantQueued int64
	}{
		{name: "empty payload", payload: nil, wantQueued: 1},
		{name: "exactly at limit", payload: bytes.Repeat([]byte{'x'}, 8), wantQueued: 1},
		{name: "one byte over", payload: bytes.Repeat([]byte{'x'}, 9), wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mr := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })

			job := NewJob("boundary-job", "", 0, "low", "", "")
			job.Payload = tt.payload
			err := Enqueue(context.Background(), rdb, "jobs", job, 8)

			if tt.wantErr {
				var sizeErr *PayloadTooLargeError
				if !errors.As(err, &sizeErr) {
					t.Fatalf("expected PayloadTooLargeError, got %v", err)
				}
				if !errors.Is(err, ErrPayloadTooLarge) {
					t.Fatalf("expected errors.Is(err, ErrPayloadTooLarge), got %v", err)
				}
				if sizeErr.Size != len(tt.payload) || sizeErr.Limit != 8 {
					t.Fatalf("unexpected size error: %#v", sizeErr)
				}
			} else if err != nil {
				t.Fatal(err)
			}

			queued, err := rdb.LLen(context.Background(), "jobs").Result()
			if err != nil {
				t.Fatal(err)
			}
			if queued != tt.wantQueued {
				t.Fatalf("expected queue length %d, got %d", tt.wantQueued, queued)
			}
		})
	}
}

func TestEnqueueAllowsEmptyPayloadSchema(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	job := NewJob("schema-optional", "", 0, "low", "", "")
	job.Payload = []byte(`{"ok":true}`)

	if err := Enqueue(context.Background(), rdb, "jobs", job, DefaultMaxPayloadSize); err != nil {
		t.Fatal(err)
	}
	encoded, err := rdb.RPop(context.Background(), "jobs").Result()
	if err != nil {
		t.Fatal(err)
	}
	got, err := UnmarshalJob(encoded)
	if err != nil {
		t.Fatal(err)
	}
	if got.PayloadSchema != "" {
		t.Fatalf("expected empty schema, got %q", got.PayloadSchema)
	}
}
