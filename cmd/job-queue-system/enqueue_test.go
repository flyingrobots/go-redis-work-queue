// Copyright 2026 James Ross
package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queueclient"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
)

func writeEnqueueCLIConfig(t *testing.T, redisAddr string, maxPayload int) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	contents := fmt.Sprintf(`redis:
  addr: %q
  dial_timeout: 100ms
  read_timeout: 100ms
  write_timeout: 100ms
  max_retries: 0
worker:
  priorities: ["high", "low"]
  queues:
    high: "jobqueue:cli:high"
    low: "jobqueue:cli:low"
producer:
  default_priority: "low"
queue:
  max_payload_size: %d
  ordered_ready_list: "jobqueue:cli:ordered:ready"
  ordered_active_set: "jobqueue:cli:ordered:active"
  ordered_queue_pattern: "jobqueue:cli:ordered:queue:%%s"
  ordered_lease_pattern: "jobqueue:cli:ordered:lease:%%s"
`, redisAddr, maxPayload)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunEnqueueReadsStdinAndPrintsJobID(t *testing.T) {
	mr := miniredis.RunT(t)
	configPath := writeEnqueueCLIConfig(t, mr.Addr(), 1024)
	payload := []byte(`{"x":1}`)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	err := runEnqueue([]string{
		"--config", configPath,
		"--schema", "demo.v1",
		"--priority", "high",
		"--ordering-key", "account:42",
	}, bytes.NewReader(payload), &stdout, &stderr)
	if err != nil {
		t.Fatalf("run enqueue: %v (stderr=%q)", err, stderr.String())
	}
	id := strings.TrimSpace(stdout.String())
	if id == "" || strings.Contains(id, "\n") {
		t.Fatalf("stdout must contain one job ID, got %q", stdout.String())
	}

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	orderedQueue := fmt.Sprintf("jobqueue:cli:ordered:queue:%s", queuekeys.OrderingDigest("account:42"))
	raw, err := rdb.RPop(context.Background(), orderedQueue).Result()
	if err != nil {
		t.Fatal(err)
	}
	job, err := queue.UnmarshalJob(raw)
	if err != nil {
		t.Fatal(err)
	}
	if job.ID != id || !bytes.Equal(job.Payload, payload) || job.PayloadSchema != "demo.v1" || job.OrderingKey != "account:42" {
		t.Fatalf("unexpected queued job: %#v", job)
	}
}

func TestRunEnqueueReadsPayloadFile(t *testing.T) {
	mr := miniredis.RunT(t)
	configPath := writeEnqueueCLIConfig(t, mr.Addr(), 1024)
	payloadPath := filepath.Join(t.TempDir(), "payload.bin")
	want := []byte{0x00, 0x01, 0xfe, 0xff}
	if err := os.WriteFile(payloadPath, want, 0o600); err != nil {
		t.Fatal(err)
	}
	var stdout bytes.Buffer

	if err := runEnqueue([]string{"--config", configPath, "--payload-file", payloadPath}, strings.NewReader("ignored"), &stdout, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	raw, err := mr.List("jobqueue:cli:low")
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 1 {
		t.Fatalf("queued items = %d, want 1", len(raw))
	}
	job, err := queue.UnmarshalJob(raw[0])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(job.Payload, want) {
		t.Fatalf("file payload changed: got %v, want %v", job.Payload, want)
	}
}

func TestRunEnqueueAllowsEmptyStdin(t *testing.T) {
	mr := miniredis.RunT(t)
	configPath := writeEnqueueCLIConfig(t, mr.Addr(), 1024)
	var stdout bytes.Buffer

	if err := runEnqueue([]string{"--config", configPath}, strings.NewReader(""), &stdout, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(stdout.String()) == "" {
		t.Fatal("empty payload enqueue did not print an ID")
	}
	raw, err := mr.List("jobqueue:cli:low")
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 1 {
		t.Fatalf("empty payload queued items = %d, want 1", len(raw))
	}
	job, err := queue.UnmarshalJob(raw[0])
	if err != nil {
		t.Fatal(err)
	}
	if len(job.Payload) != 0 {
		t.Fatalf("empty stdin became %d payload bytes", len(job.Payload))
	}
}

func TestRunEnqueueBoundsStdinByConfiguredPayloadLimit(t *testing.T) {
	mr := miniredis.RunT(t)
	configPath := writeEnqueueCLIConfig(t, mr.Addr(), 4)
	input := bytes.Repeat([]byte{'x'}, 1024)
	stdin := bytes.NewReader(input)

	err := runEnqueue([]string{"--config", configPath}, stdin, &bytes.Buffer{}, &bytes.Buffer{})
	var sizeErr *queueclient.PayloadTooLargeError
	if !errors.As(err, &sizeErr) {
		t.Fatalf("expected PayloadTooLargeError, got %v", err)
	}
	if consumed := len(input) - stdin.Len(); consumed > 5 {
		t.Fatalf("stdin reader consumed %d bytes, want at most limit+1 (5)", consumed)
	}
	if sizeErr.Size != 5 || sizeErr.Limit != 4 {
		t.Fatalf("payload size error = %#v, want observed limit+1 and limit", sizeErr)
	}
	if keys := mr.Keys(); len(keys) != 0 {
		t.Fatalf("oversized stdin changed Redis: %v", keys)
	}
}

func TestRunEnqueueBoundsPayloadFileByConfiguredPayloadLimit(t *testing.T) {
	mr := miniredis.RunT(t)
	configPath := writeEnqueueCLIConfig(t, mr.Addr(), 4)
	payloadPath := filepath.Join(t.TempDir(), "oversized.bin")
	if err := os.WriteFile(payloadPath, bytes.Repeat([]byte{'x'}, 1024), 0o600); err != nil {
		t.Fatal(err)
	}

	err := runEnqueue([]string{
		"--config", configPath,
		"--payload-file", payloadPath,
	}, strings.NewReader("ignored"), &bytes.Buffer{}, &bytes.Buffer{})
	var sizeErr *queueclient.PayloadTooLargeError
	if !errors.As(err, &sizeErr) {
		t.Fatalf("expected PayloadTooLargeError, got %v", err)
	}
	if sizeErr.Size != 5 || sizeErr.Limit != 4 {
		t.Fatalf("payload size error = %#v, want observed limit+1 and limit", sizeErr)
	}
	if keys := mr.Keys(); len(keys) != 0 {
		t.Fatalf("oversized payload file changed Redis: %v", keys)
	}
}

func TestRunEnqueueRejectsUnknownPriorityBeforeRedis(t *testing.T) {
	mr := miniredis.RunT(t)
	configPath := writeEnqueueCLIConfig(t, mr.Addr(), 1024)
	var stdout bytes.Buffer

	err := runEnqueue([]string{"--config", configPath, "--priority", "urgent"}, strings.NewReader("payload"), &stdout, &bytes.Buffer{})
	var priorityErr *queueclient.UnknownPriorityError
	if !errors.As(err, &priorityErr) {
		t.Fatalf("expected UnknownPriorityError, got %v", err)
	}
	if stdout.Len() != 0 {
		t.Fatalf("failed enqueue wrote stdout: %q", stdout.String())
	}
	if keys := mr.Keys(); len(keys) != 0 {
		t.Fatalf("invalid priority changed Redis: %v", keys)
	}
}
