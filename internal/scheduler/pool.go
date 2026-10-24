package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/amingeek/goworker/internal/task"
	"github.com/amingeek/goworker/pkg/retry"
)

type WorkerPool struct {
	workers   int
	taskQueue chan *task.Task
	handlers  map[string]task.Handler
	mu        sync.RWMutex
	wg        sync.WaitGroup
	baseDelay time.Duration
	maxDelay  time.Duration
	onRetry   func(t *task.Task)
	onDLQ     func(t *task.Task)
}

func NewWorkerPool(workers, queueSize int, baseDelay, maxDelay time.Duration, onRetry, onDLQ func(t *task.Task)) *WorkerPool {
	return &WorkerPool{
		workers:   workers,
		taskQueue: make(chan *task.Task, queueSize),
		handlers:  make(map[string]task.Handler),
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
		onRetry:   onRetry,
		onDLQ:     onDLQ,
	}
}

func (p *WorkerPool) RegisterHandler(taskType string, h task.Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.handlers[taskType] = h
}

func (p *WorkerPool) Submit(t *task.Task) {
	p.taskQueue <- t
}

func (p *WorkerPool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.startWorker(ctx)
	}
}

func (p *WorkerPool) Stop() {
	close(p.taskQueue)
	p.wg.Wait()
}

func (p *WorkerPool) startWorker(ctx context.Context) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-p.taskQueue:
			if !ok {
				return
			}
			p.execute(ctx, t)
		}
	}
}

func (p *WorkerPool) execute(ctx context.Context, t *task.Task) {
	p.mu.RLock()
	handler, exists := p.handlers[t.Type]
	p.mu.RUnlock()

	if !exists {
		t.State = task.DLQ
		t.LastError = fmt.Sprintf("no handler registered for task type: %s", t.Type)
		if p.onDLQ != nil {
			p.onDLQ(t)
		}
		return
	}

	t.State = task.Running
	err := handler(ctx, t.Payload)

	if err == nil {
		t.State = task.Completed
		t.LastError = ""
		return
	}

	t.Retries++
	t.LastError = err.Error()

	if t.Retries >= t.MaxRetries {
		t.State = task.DLQ
		if p.onDLQ != nil {
			p.onDLQ(t)
		}
		return
	}

	t.State = task.Failed
	delay := retry.CalculateBackoff(t.Retries, p.baseDelay, p.maxDelay)
	t.RunAt = time.Now().Add(delay)

	if p.onRetry != nil {
		p.onRetry(t)
	}
}
