package scheduler_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/amingeek/goworker/internal/queue"
	"github.com/amingeek/goworker/internal/scheduler"
	"github.com/amingeek/goworker/internal/task"
)

func TestDispatcher_ImmediateTaskExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := queue.NewSafeQueue()
	pool := scheduler.NewWorkerPool(2, 10, 10*time.Millisecond, 50*time.Millisecond, nil, nil)
	dispatcher := scheduler.NewDispatcher(q, pool, 20*time.Millisecond)

	var executed atomic.Bool
	pool.RegisterHandler("immediate", func(ctx context.Context, payload []byte) error {
		executed.Store(true)
		return nil
	})

	pool.Start(ctx)
	dispatcher.Start(ctx)

	job := &task.Task{
		ID:         "job-now",
		Type:       "immediate",
		RunAt:      time.Now(),
		MaxRetries: 1,
	}
	q.Push(job)

	time.Sleep(50 * time.Millisecond)

	dispatcher.Stop()
	pool.Stop()

	if !executed.Load() {
		t.Fatal("expected immediate task to be dispatched and executed")
	}
}

func TestDispatcher_DelayedTaskExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	q := queue.NewSafeQueue()
	pool := scheduler.NewWorkerPool(2, 10, 10*time.Millisecond, 50*time.Millisecond, nil, nil)
	dispatcher := scheduler.NewDispatcher(q, pool, 20*time.Millisecond)

	var executed atomic.Bool
	pool.RegisterHandler("delayed", func(ctx context.Context, payload []byte) error {
		executed.Store(true)
		return nil
	})

	pool.Start(ctx)
	dispatcher.Start(ctx)

	job := &task.Task{
		ID:         "job-delayed",
		Type:       "delayed",
		RunAt:      time.Now().Add(100 * time.Millisecond),
		MaxRetries: 1,
	}
	q.Push(job)

	time.Sleep(40 * time.Millisecond)
	if executed.Load() {
		t.Fatal("task was executed prematurely!")
	}

	time.Sleep(90 * time.Millisecond)
	if !executed.Load() {
		t.Fatal("task was expected to execute after its delay")
	}

	dispatcher.Stop()
	pool.Stop()
}
