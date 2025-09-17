// Copyright 2025 James Ross
package producer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/obs"
	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Producer struct {
	cfg *config.Config
	rdb *redis.Client
	log *zap.Logger
}

func New(cfg *config.Config, rdb *redis.Client, log *zap.Logger) *Producer {
	return &Producer{cfg: cfg, rdb: rdb, log: log}
}

func (p *Producer) Run(ctx context.Context) error {
	root := p.cfg.Producer.ScanDir
	absRoot, errAbs := filepath.Abs(root)
	if errAbs != nil {
		return errAbs
	}
	include := p.cfg.Producer.IncludeGlobs
	exclude := p.cfg.Producer.ExcludeGlobs

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		// Safety: ensure file under root
		abs, err2 := filepath.Abs(path)
		if err2 != nil {
			return nil
		}
		if !strings.HasPrefix(abs, absRoot+string(os.PathSeparator)) && abs != absRoot {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		// include check
		incMatch := len(include) == 0
		for _, g := range include {
			if ok, _ := doublestar.PathMatch(g, rel); ok {
				incMatch = true
				break
			}
		}
		if !incMatch {
			return nil
		}
		for _, g := range exclude {
			if ok, _ := doublestar.PathMatch(g, rel); ok {
				return nil
			}
		}

		// Per-file enqueue (streaming)
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if err := p.rateLimit(ctx); err != nil {
			return err
		}
		fi, err := os.Stat(path)
		if err != nil {
			return nil
		}
		prio := p.priorityForExt(filepath.Ext(path))
		id := randID()

		// Start enqueue span for tracing
		enqCtx, enqSpan := obs.StartEnqueueSpan(ctx, p.cfg.Worker.Queues[prio], prio)

		// Get trace and span IDs from the current context
		traceID, spanID := obs.GetTraceAndSpanID(enqCtx)
		if traceID == "" {
			// Fallback to random IDs if no active trace
			traceID, spanID = randTraceAndSpan()
		}

		j := queue.NewJob(id, abs, fi.Size(), prio, traceID, spanID)

		// Add span attributes
		obs.AddSpanAttributes(enqCtx,
			obs.KeyValue("job.id", j.ID),
			obs.KeyValue("job.filepath", abs),
			obs.KeyValue("job.filesize", fi.Size()),
			obs.KeyValue("job.priority", prio),
		)

		payload, _ := j.Marshal()
		key := p.cfg.Worker.Queues[prio]
		if key == "" {
			key = p.cfg.Worker.Queues[p.cfg.Producer.DefaultPriority]
		}

		// Add event before enqueue
		obs.AddEvent(enqCtx, "enqueueing_job",
			obs.KeyValue("queue", key),
			obs.KeyValue("job_id", j.ID),
		)

		if err := p.rdb.LPush(enqCtx, key, payload).Err(); err != nil {
			obs.RecordError(enqCtx, err)
			enqSpan.End()
			return err
		}

		// Mark span as successful
		obs.SetSpanSuccess(enqCtx)
		obs.AddEvent(enqCtx, "job_enqueued",
			obs.KeyValue("queue", key),
			obs.KeyValue("job_id", j.ID),
		)
		enqSpan.End()

		obs.JobsProduced.Inc()
		p.log.Info("enqueued job", obs.String("id", j.ID), obs.String("queue", key), obs.String("trace_id", j.TraceID), obs.String("span_id", j.SpanID))
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (p *Producer) priorityForExt(ext string) string {
	ext = strings.ToLower(ext)
	for _, e := range p.cfg.Producer.HighPriorityExts {
		if strings.ToLower(e) == ext {
			return "high"
		}
	}
	return p.cfg.Producer.DefaultPriority
}

func (p *Producer) rateLimit(ctx context.Context) error {
	if p.cfg.Producer.RateLimitPerSec <= 0 {
		return nil
	}
	key := p.cfg.Producer.RateLimitKey
	// Fixed-window limiter with TTL-based sleep and jitter
	n, err := p.rdb.Incr(ctx, key).Result()
	if err != nil {
		return err
	}
	if n == 1 {
		// First hit in window: set expiry
		_ = p.rdb.Expire(ctx, key, time.Second).Err()
	}
	if int(n) > p.cfg.Producer.RateLimitPerSec {
		ttl, err := p.rdb.TTL(ctx, key).Result()
		if err == nil && ttl > 0 {
			// Add jitter up to 50ms
			jitter := time.Duration(randUint32()%50) * time.Millisecond
			select {
			case <-ctx.Done():
			case <-time.After(ttl + jitter):
			}
		} else {
			time.Sleep(200 * time.Millisecond)
		}
	}
	return nil
}

func randUint32() uint32 {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
}

func randID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

func randTraceAndSpan() (string, string) {
	var tb [16]byte
	var sb [8]byte
	_, _ = rand.Read(tb[:])
	_, _ = rand.Read(sb[:])
	return hex.EncodeToString(tb[:]), hex.EncodeToString(sb[:])
}
