package scheduler_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amingeek/goworker/internal/scheduler"
	"github.com/amingeek/goworker/internal/task"
)

func TestWorkerPool_SuccessTask(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool := scheduler.NewWorkerPool(2, 10, 10*time.Millisecond, 50*time.Millisecond, nil, nil)

	var executed atomic.Bool
	pool.RegisterHandler("email", func(ctx context.Context, payload []byte) error {
		executed.Store(true)
		return nil
	})

	pool.Start(ctx)

	job := &task.Task{
		ID:         "job-1",
		Type:       "email",
		MaxRetries: 3,
	}

	pool.Submit(job)

	time.Sleep(50 * time.Millisecond)
	pool.Stop()

	if !executed.Load() {
		t.Fatal("expected handler to be executed")
	}
	if job.State != task.Completed {
		t.Fatalf("expected task state COMPLETED, got %v", job.State)
	}
}

func TestWorkerPool_RetryAndDLQ(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var retriedCount atomic.Int32
	var dlqCount atomic.Int32
	var wg sync.WaitGroup
	wg.Add(1)

	onRetry := func(tk *task.Task) {
		retriedCount.Add(1)
	}

	onDLQ := func(tk *task.Task) {
		dlqCount.Add(1)
		wg.Done()
	}

	pool := scheduler.NewWorkerPool(2, 10, 5*time.Millisecond, 20*time.Millisecond, onRetry, onDLQ)

	pool.RegisterHandler("faulty", func(ctx context.Context, payload []byte) error {
		return errors.New("network error")
	})

	pool.Start(ctx)

	job := &task.Task{
		ID:         "job-faulty",
		Type:       "faulty",
		MaxRetries: 1,
	}

	pool.Submit(job)

	wg.Wait()
	pool.Stop()

	if dlqCount.Load() != 1 {
		t.Fatalf("expected 1 task in DLQ, got %d", dlqCount.Load())
	}
	if job.State != task.DLQ {
		t.Fatalf("expected state DLQ, got %v", job.State)
	}
	if job.LastError != "network error" {
		t.Fatalf("expected error message 'network error', got %v", job.LastError)
	}
}
