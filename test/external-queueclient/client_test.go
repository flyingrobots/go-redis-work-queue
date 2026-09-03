// Copyright 2026 James Ross
//
// This fixture is deliberately a separate Go module. Before ROADMAP Item 3,
// the only queue packages lived under internal/, so this import surface did
// not exist and an external module could not legally enqueue work.
package queueclientsmoke

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queueclient"
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
