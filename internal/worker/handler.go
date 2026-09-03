// Copyright 2026 James Ross
package worker

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/queue"
	"go.uber.org/zap"
)

// Handler executes one job. A Worker may invoke a Handler concurrently, so
// handlers must protect mutable state and honor context cancellation.
type Handler func(ctx context.Context, job queue.Job) error

// ErrHandlerRequired prevents a production worker from acknowledging jobs
// without application logic. BenchHandler must be selected explicitly.
var ErrHandlerRequired = errors.New("worker handler is required; select BenchHandler explicitly for benchmark workloads")

// ErrBenchJobFailed is returned by BenchHandler when the legacy filepath
// failure marker is present.
var ErrBenchJobFailed = errors.New("bench job filepath contains failure marker")

// BenchHandler preserves the original benchmark worker behavior. It sleeps for
// up to one second based on the legacy FileSize field and treats a FilePath
// containing "fail" as an error. Application payload bytes are never inspected.
func BenchHandler(ctx context.Context, job queue.Job) error {
	duration := time.Duration(min64(job.FileSize/1024, 1000)) * time.Millisecond
	if duration > 0 {
		timer := time.NewTimer(duration)
		defer func() {
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		}()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	} else {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}

	if strings.Contains(strings.ToLower(job.FilePath), "fail") {
		return ErrBenchJobFailed
	}
	return nil
}

// Handle installs the application handler used for future jobs. Passing nil
// clears the handler; it never enables BenchHandler implicitly. Handle is safe
// to call while the Worker is running.
func (w *Worker) Handle(handler Handler) *Worker {
	w.handlerMu.Lock()
	w.handler = handler
	w.handlerMu.Unlock()
	return w
}

func (w *Worker) selectedHandler() Handler {
	w.handlerMu.RLock()
	handler := w.handler
	w.handlerMu.RUnlock()
	return handler
}

func (w *Worker) invokeHandler(ctx context.Context, job queue.Job) (err error) {
	handler := w.selectedHandler()
	if handler == nil {
		return ErrHandlerRequired
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			stack := debug.Stack()
			w.log.Error("job handler panicked",
				zap.String("job_id", job.ID),
				zap.Any("panic", recovered),
				zap.ByteString("stack", stack),
			)
			err = fmt.Errorf("job handler panicked: %v", recovered)
		}
	}()
	return handler(ctx, job)
}
