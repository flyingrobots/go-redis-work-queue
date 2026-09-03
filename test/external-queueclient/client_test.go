// Copyright 2026 James Ross
//
// This fixture is deliberately a separate Go module. Before ROADMAP Item 3,
// the only queue packages lived under internal/, so this import surface did
// not exist and an external module could not legally enqueue work.
package queueclientsmoke

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queueclient"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queueworker"
	"github.com/redis/go-redis/v9"
)

func TestExternalModuleCanEnqueueAndPeek(t *testing.T) {
	mr := miniredis.RunT(t)
	client, err := queueclient.New(&redis.Options{Addr: mr.Addr()}, queueclient.Config{
		Queues:          map[string]string{"low": "external:low"},
		DefaultPriority: "low",
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = client.Close() })

	id, err := client.Enqueue(context.Background(), queueclient.Job{
		Payload:       []byte(`{"from":"external-module"}`),
		PayloadSchema: "external.v1",
	})
	if err != nil {
		t.Fatal(err)
	}
	peek, err := client.Peek(context.Background(), "low", 1)
	if err != nil {
		t.Fatal(err)
	}
	if id == "" || len(peek.Items) != 1 {
		t.Fatalf("enqueue/peek failed: id=%q peek=%#v", id, peek)
	}
}

func TestExternalModuleCanRunApplicationHandler(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	queueCfg := queueclient.DefaultConfig()
	queueCfg.Queues = map[string]string{"low": "external:worker:low"}
	queueCfg.DefaultPriority = "low"
	client, err := queueclient.NewWithClient(rdb, queueCfg)
	if err != nil {
		t.Fatal(err)
	}

	workerCfg := queueworker.DefaultConfig()
	workerCfg.Count = 1
	workerCfg.Priorities = []string{"low"}
	workerCfg.QueueConfig = queueCfg
	workerCfg.BRPopLPushTimeout = 10 * time.Millisecond
	received := make(chan queueclient.Job, 1)
	wrk, err := queueworker.NewWithClient(rdb, workerCfg, func(_ context.Context, job queueclient.Job) error {
		received <- job
		return nil
	}, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = wrk.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- wrk.Run(ctx) }()

	id, err := client.Enqueue(context.Background(), queueclient.Job{Payload: []byte("external-handler")})
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	select {
	case job := <-received:
		if job.ID != id || string(job.Payload) != "external-handler" {
			t.Fatalf("handler received unexpected job: %#v", job)
		}
	case <-time.After(2 * time.Second):
		cancel()
		t.Fatal("external handler did not receive the job")
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		completed, countErr := rdb.LLen(context.Background(), queueCfg.CompletedList).Result()
		if countErr == nil && completed == 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if completed, countErr := rdb.LLen(context.Background(), queueCfg.CompletedList).Result(); countErr != nil || completed != 1 {
		cancel()
		t.Fatalf("external handler completion count = %d, err = %v", completed, countErr)
	}

	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("worker stopped with error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not stop")
	}
}
