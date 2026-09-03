// Copyright 2026 James Ross
package queue

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queuekeys"
	"github.com/redis/go-redis/v9"
)

func testOrderingLayout() OrderingLayout {
	return OrderingLayout{
		ReadyList:    "test:ordered:ready",
		ActiveSet:    "test:ordered:active",
		QueuePattern: "test:ordered:queue:%s",
		LeasePattern: "test:ordered:lease:%s",
	}
}

func TestOrderingLayoutRejectsIdenticalQueueAndLeasePatterns(t *testing.T) {
	layout := testOrderingLayout()
	layout.LeasePattern = layout.QueuePattern
	if err := layout.Validate(); err == nil {
		t.Fatal("expected identical ordered queue and lease patterns to fail validation")
	}
}

func TestOrderingLayoutRejectsIdenticalReadyAndActiveKeys(t *testing.T) {
	layout := testOrderingLayout()
	layout.ActiveSet = layout.ReadyList
	if err := layout.Validate(); err == nil {
		t.Fatal("expected identical ordered ready and active keys to fail validation")
	}
}

func TestUnorderedEnqueueRetainsLegacyKeyLayout(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	job := NewJob("plain", "", 0, "low", "", "")
	if err := EnqueueWithOrdering(context.Background(), rdb, "test:low", job, DefaultMaxPayloadSize, testOrderingLayout()); err != nil {
		t.Fatal(err)
	}

	keys, err := rdb.Keys(context.Background(), "*").Result()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(keys, []string{"test:low"}) {
		t.Fatalf("unordered enqueue keys = %v, want only legacy queue", keys)
	}
	got, err := rdb.RPop(context.Background(), "test:low").Result()
	if err != nil {
		t.Fatal(err)
	}
	want, _ := job.Marshal()
	if got != want {
		t.Fatalf("unordered envelope changed: got %q, want %q", got, want)
	}
}

func TestOrderedEnqueueClaimAndComplete(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()
	layout := testOrderingLayout()

	job := NewJob("ordered", "", 0, "high", "", "")
	job.OrderingKey = "repo:{main}:目录/file.go"
	if err := EnqueueWithOrdering(ctx, rdb, "test:high", job, DefaultMaxPayloadSize, layout); err != nil {
		t.Fatal(err)
	}
	digest := queuekeys.OrderingDigest(job.OrderingKey)
	queueKey := queuekeys.Format(layout.QueuePattern, digest)
	leaseKey := queuekeys.Format(layout.LeasePattern, digest)
	gotList, err := mr.List(queueKey)
	if err != nil || len(gotList) != 1 {
		t.Fatalf("ordered queue length = %d, want 1 (err=%v)", len(gotList), err)
	}
	if mr.Exists("test:high") {
		t.Fatal("ordered enqueue touched the ordinary priority list")
	}

	delivery, ok, err := ClaimOrdered(ctx, rdb, layout, "test:processing", "test:heartbeat", "worker-1", time.Second)
	if err != nil || !ok {
		t.Fatalf("claim = (%#v, %v, %v)", delivery, ok, err)
	}
	if delivery.Digest != digest || delivery.QueueKey != queueKey || delivery.LeaseKey != leaseKey {
		t.Fatalf("unexpected delivery metadata: %#v", delivery)
	}
	if got, err := rdb.Get(ctx, leaseKey).Result(); err != nil || got != "worker-1" {
		t.Fatalf("lease owner = %q (err=%v)", got, err)
	}
	if got, err := rdb.LLen(ctx, "test:processing").Result(); err != nil || got != 1 {
		t.Fatalf("processing length = %d (err=%v)", got, err)
	}

	owned, err := RenewOrderedLease(ctx, rdb, leaseKey, "worker-1", 2*time.Second)
	if err != nil || !owned {
		t.Fatalf("renew = (%v, %v)", owned, err)
	}
	transitioned, err := TransitionOrdered(ctx, rdb, layout, delivery, "test:processing", "test:heartbeat", "worker-1", "test:completed", delivery.Payload, OrderedComplete)
	if err != nil || !transitioned {
		t.Fatalf("complete = (%v, %v)", transitioned, err)
	}
	if got, err := rdb.LLen(ctx, "test:completed").Result(); err != nil || got != 1 {
		t.Fatalf("completed length = %d (err=%v)", got, err)
	}
	for _, key := range []string{queueKey, leaseKey, layout.ReadyList, layout.ActiveSet, "test:processing", "test:heartbeat"} {
		if mr.Exists(key) {
			t.Errorf("completed transition left key %q", key)
		}
	}
}

func TestOrderedEnqueueWrongTypeControlKeyIsAtomic(t *testing.T) {
	for _, control := range []string{"ready", "active"} {
		t.Run(control, func(t *testing.T) {
			mr := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })
			ctx := context.Background()
			layout := testOrderingLayout()

			corruptKey := layout.ReadyList
			otherKey := layout.ActiveSet
			if control == "active" {
				corruptKey, otherKey = layout.ActiveSet, layout.ReadyList
			}
			if err := rdb.Set(ctx, corruptKey, "not-a-control-structure", 0).Err(); err != nil {
				t.Fatal(err)
			}

			job := NewJob("wrong-type-"+control, "", 0, "low", "", "")
			job.OrderingKey = "account:" + control
			if err := EnqueueWithOrdering(ctx, rdb, "ignored", job, DefaultMaxPayloadSize, layout); err == nil {
				t.Fatal("ordered enqueue succeeded with wrong-type control key")
			}

			queueKey := queuekeys.Format(layout.QueuePattern, queuekeys.OrderingDigest(job.OrderingKey))
			if exists := rdb.Exists(ctx, queueKey).Val(); exists != 0 {
				t.Fatalf("failed enqueue left ordered queue %q", queueKey)
			}
			if exists := rdb.Exists(ctx, otherKey).Val(); exists != 0 {
				t.Fatalf("failed enqueue mutated other control key %q", otherKey)
			}
			if got := rdb.Get(ctx, corruptKey).Val(); got != "not-a-control-structure" {
				t.Fatalf("corrupt control value = %q", got)
			}
		})
	}
}

func TestOrderedIntakeRejectsPerJobKeyAliasesBeforeWriting(t *testing.T) {
	type aliasCase struct {
		name      string
		configure func(*OrderingLayout, string)
	}
	aliases := []aliasCase{
		{
			name: "queue-ready",
			configure: func(layout *OrderingLayout, digest string) {
				layout.ReadyList = queuekeys.Format(layout.QueuePattern, digest)
			},
		},
		{
			name: "queue-active",
			configure: func(layout *OrderingLayout, digest string) {
				layout.ActiveSet = queuekeys.Format(layout.QueuePattern, digest)
			},
		},
		{
			name: "lease-ready",
			configure: func(layout *OrderingLayout, digest string) {
				layout.ReadyList = queuekeys.Format(layout.LeasePattern, digest)
			},
		},
		{
			name: "lease-active",
			configure: func(layout *OrderingLayout, digest string) {
				layout.ActiveSet = queuekeys.Format(layout.LeasePattern, digest)
			},
		},
		{
			name: "queue-lease",
			configure: func(layout *OrderingLayout, digest string) {
				layout.QueuePattern = "%s" + digest
				layout.LeasePattern = digest + "%s"
			},
		},
	}
	intakes := []struct {
		name string
		run  func(context.Context, redis.Cmdable, Job, string, OrderingLayout) error
	}{
		{
			name: "single",
			run: func(ctx context.Context, rdb redis.Cmdable, job Job, encoded string, layout OrderingLayout) error {
				return AppendEncoded(ctx, rdb, "ignored", job, encoded, layout)
			},
		},
		{
			name: "batch",
			run: func(ctx context.Context, rdb redis.Cmdable, job Job, encoded string, layout OrderingLayout) error {
				return AppendEncodedBatch(ctx, rdb, []PreparedEnqueue{{
					QueueName: "ignored",
					Job:       job,
					Encoded:   encoded,
				}}, layout)
			},
		},
	}

	for _, alias := range aliases {
		for _, intake := range intakes {
			t.Run(alias.name+"/"+intake.name, func(t *testing.T) {
				mr := miniredis.RunT(t)
				rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
				t.Cleanup(func() { _ = rdb.Close() })
				ctx := context.Background()
				job := NewJob("alias-"+alias.name, "", 0, "low", "", "")
				job.OrderingKey = "account:" + alias.name
				digest := queuekeys.OrderingDigest(job.OrderingKey)
				layout := testOrderingLayout()
				alias.configure(&layout, digest)
				if err := layout.Validate(); err != nil {
					t.Fatalf("alias fixture must pass static layout validation: %v", err)
				}
				encoded, err := job.Marshal()
				if err != nil {
					t.Fatal(err)
				}

				if err := intake.run(ctx, rdb, job, encoded, layout); err == nil {
					t.Fatal("ordered intake accepted aliased per-job keys")
				}
				if keys := rdb.Keys(ctx, "*").Val(); len(keys) != 0 {
					t.Fatalf("rejected ordered intake changed Redis keys: %v", keys)
				}
			})
		}
	}
}

func TestClaimOrderedWrongTypeKeyPreservesReadyJob(t *testing.T) {
	for _, target := range []string{"processing", "queue", "lease"} {
		t.Run(target, func(t *testing.T) {
			mr := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })
			ctx := context.Background()
			layout := testOrderingLayout()

			job := NewJob("wrong-type-claim-"+target, "", 0, "low", "", "")
			job.OrderingKey = "account:claim:" + target
			if err := EnqueueWithOrdering(ctx, rdb, "ignored", job, DefaultMaxPayloadSize, layout); err != nil {
				t.Fatal(err)
			}
			digest := queuekeys.OrderingDigest(job.OrderingKey)
			queueKey := queuekeys.Format(layout.QueuePattern, digest)
			leaseKey := queuekeys.Format(layout.LeasePattern, digest)
			processingKey := "test:processing"
			switch target {
			case "processing":
				if err := rdb.Set(ctx, processingKey, "not-a-list", 0).Err(); err != nil {
					t.Fatal(err)
				}
			case "queue":
				if err := rdb.Set(ctx, queueKey, "not-a-list", 0).Err(); err != nil {
					t.Fatal(err)
				}
			case "lease":
				if err := rdb.LPush(ctx, leaseKey, "not-a-string").Err(); err != nil {
					t.Fatal(err)
				}
			}

			if delivery, ok, err := ClaimOrdered(ctx, rdb, layout, processingKey, "test:heartbeat", "worker-1", time.Second); err == nil || ok {
				t.Fatalf("claim = (%#v, %v, %v), want wrong-type error", delivery, ok, err)
			}
			if ready := rdb.LRange(ctx, layout.ReadyList, 0, -1).Val(); !slices.Equal(ready, []string{digest}) {
				t.Fatalf("ready ring after failed claim = %#v", ready)
			}
			if active := rdb.SIsMember(ctx, layout.ActiveSet, digest).Val(); !active {
				t.Fatal("failed claim removed the active digest")
			}
			if target != "queue" {
				want, _ := job.Marshal()
				if queued := rdb.LRange(ctx, queueKey, 0, -1).Val(); !slices.Equal(queued, []string{want}) {
					t.Fatalf("ordered queue after failed claim = %#v", queued)
				}
			}
			if target != "lease" && rdb.Exists(ctx, leaseKey).Val() != 0 {
				t.Fatal("failed claim created a lease")
			}
		})
	}
}

func TestTransitionOrderedWrongTypeDestinationPreservesProcessingDelivery(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()
	layout := testOrderingLayout()

	job := NewJob("wrong-type-terminal", "", 0, "low", "", "")
	job.OrderingKey = "account:terminal"
	if err := EnqueueWithOrdering(ctx, rdb, "ignored", job, DefaultMaxPayloadSize, layout); err != nil {
		t.Fatal(err)
	}
	delivery, ok, err := ClaimOrdered(ctx, rdb, layout, "test:processing", "test:heartbeat", "worker-1", time.Second)
	if err != nil || !ok {
		t.Fatalf("claim = (%#v, %v, %v)", delivery, ok, err)
	}
	if err := rdb.Set(ctx, "test:completed", "not-a-list", 0).Err(); err != nil {
		t.Fatal(err)
	}

	transitioned, err := TransitionOrdered(ctx, rdb, layout, delivery, "test:processing", "test:heartbeat", "worker-1", "test:completed", delivery.Payload, OrderedComplete)
	if err == nil || transitioned {
		t.Fatalf("complete = (%v, %v), want wrong-type error", transitioned, err)
	}
	if got := rdb.LRange(ctx, "test:processing", 0, -1).Val(); len(got) != 1 || got[0] != delivery.Payload {
		t.Fatalf("processing delivery after failed transition = %#v", got)
	}
	if owner := rdb.Get(ctx, delivery.LeaseKey).Val(); owner != "worker-1" {
		t.Fatalf("lease owner after failed transition = %q", owner)
	}
}

func TestRecoverOrderedWrongTypeDestinationPreservesProcessingDelivery(t *testing.T) {
	for _, target := range []string{"queue", "ready", "active"} {
		t.Run(target, func(t *testing.T) {
			mr := miniredis.RunT(t)
			rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			t.Cleanup(func() { _ = rdb.Close() })
			ctx := context.Background()
			layout := testOrderingLayout()
			processingKey := "test:processing"
			heartbeatKey := "test:heartbeat"

			job := NewJob("wrong-type-recover-"+target, "", 0, "low", "", "")
			job.OrderingKey = "account:recover:" + target
			if err := EnqueueWithOrdering(ctx, rdb, "ignored", job, DefaultMaxPayloadSize, layout); err != nil {
				t.Fatal(err)
			}
			delivery, ok, err := ClaimOrdered(ctx, rdb, layout, processingKey, heartbeatKey, "worker-1", time.Second)
			if err != nil || !ok {
				t.Fatalf("claim = (%#v, %v, %v)", delivery, ok, err)
			}
			if err := rdb.Del(ctx, heartbeatKey, delivery.LeaseKey).Err(); err != nil {
				t.Fatal(err)
			}

			corruptKey := delivery.QueueKey
			switch target {
			case "ready":
				corruptKey = layout.ReadyList
			case "active":
				corruptKey = layout.ActiveSet
				if err := rdb.Del(ctx, corruptKey).Err(); err != nil {
					t.Fatal(err)
				}
			}
			if err := rdb.Set(ctx, corruptKey, "not-a-destination", 0).Err(); err != nil {
				t.Fatal(err)
			}

			if recovered, err := RecoverOrdered(ctx, rdb, layout, delivery, processingKey, heartbeatKey); err == nil || recovered {
				t.Fatalf("recover = (%v, %v), want wrong-type error", recovered, err)
			}
			if processing := rdb.LRange(ctx, processingKey, 0, -1).Val(); !slices.Equal(processing, []string{delivery.Payload}) {
				t.Fatalf("processing delivery after failed recovery = %#v", processing)
			}
			if target != "queue" && rdb.Exists(ctx, delivery.QueueKey).Val() != 0 {
				t.Fatalf("failed recovery wrote ordered queue %q", delivery.QueueKey)
			}
			if got := rdb.Get(ctx, corruptKey).Val(); got != "not-a-destination" {
				t.Fatalf("corrupt destination value = %q", got)
			}
		})
	}
}

func TestOrderedRetryAndRecoveryStayAheadOfLaterJobs(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	ctx := context.Background()
	layout := testOrderingLayout()

	jobs := []Job{
		{ID: "first", OrderingKey: "same", Priority: "low", CreationTime: time.Now().UTC().Format(time.RFC3339Nano)},
		{ID: "later", OrderingKey: "same", Priority: "high", CreationTime: time.Now().UTC().Format(time.RFC3339Nano)},
	}
	for _, job := range jobs {
		if err := EnqueueWithOrdering(ctx, rdb, "ignored", job, DefaultMaxPayloadSize, layout); err != nil {
			t.Fatal(err)
		}
	}

	first, ok, err := ClaimOrdered(ctx, rdb, layout, "test:processing", "test:heartbeat", "worker-1", time.Second)
	if err != nil || !ok {
		t.Fatalf("first claim = (%v, %v)", ok, err)
	}
	firstJob, _ := UnmarshalJob(first.Payload)
	firstJob.Retries++
	retryPayload, _ := firstJob.Marshal()
	transitioned, err := TransitionOrdered(ctx, rdb, layout, first, "test:processing", "test:heartbeat", "worker-1", "", retryPayload, OrderedRetry)
	if err != nil || !transitioned {
		t.Fatalf("retry = (%v, %v)", transitioned, err)
	}

	retry, ok, err := ClaimOrdered(ctx, rdb, layout, "test:processing", "test:heartbeat", "worker-2", time.Second)
	if err != nil || !ok {
		t.Fatalf("retry claim = (%v, %v)", ok, err)
	}
	retriedJob, _ := UnmarshalJob(retry.Payload)
	if retriedJob.ID != "first" || retriedJob.Retries != 1 {
		t.Fatalf("retry did not stay first: %#v", retriedJob)
	}

	// Simulate a crash: ownership expires, then the reaper must put this same
	// envelope ahead of the still-pending later job.
	if err := rdb.Del(ctx, "test:heartbeat", retry.LeaseKey).Err(); err != nil {
		t.Fatal(err)
	}
	recovered, err := RecoverOrdered(ctx, rdb, layout, retry, "test:processing", "test:heartbeat")
	if err != nil || !recovered {
		t.Fatalf("recover = (%v, %v)", recovered, err)
	}
	redelivered, ok, err := ClaimOrdered(ctx, rdb, layout, "test:processing", "test:heartbeat", "worker-3", time.Second)
	if err != nil || !ok {
		t.Fatalf("redelivery claim = (%v, %v)", ok, err)
	}
	redeliveredJob, _ := UnmarshalJob(redelivered.Payload)
	if redeliveredJob.ID != "first" {
		t.Fatalf("recovered job = %q, want first", redeliveredJob.ID)
	}
}
