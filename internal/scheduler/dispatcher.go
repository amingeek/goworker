package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/amingeek/goworker/internal/queue"
)

type Dispatcher struct {
	queue      *queue.SafeQueue
	pool       *WorkerPool
	pollPeriod time.Duration
	wg         sync.WaitGroup
	stopCh     chan struct{}
}

func NewDispatcher(q *queue.SafeQueue, pool *WorkerPool, pollPeriod time.Duration) *Dispatcher {
	if pollPeriod <= 0 {
		pollPeriod = 50 * time.Millisecond
	}
	return &Dispatcher{
		queue:      q,
		pool:       pool,
		pollPeriod: pollPeriod,
		stopCh:     make(chan struct{}),
	}
}

func (d *Dispatcher) Start(ctx context.Context) {
	d.wg.Add(1)
	go d.dispatchLoop(ctx)
}

func (d *Dispatcher) dispatchLoop(ctx context.Context) {
	defer d.wg.Done()

	for {
		for {
			t := d.queue.PopReady()
			if t == nil {
				break
			}
			d.pool.Submit(t)
		}

		sleepDuration := d.pollPeriod
		if nextRun, ok := d.queue.NextRunAt(); ok {
			timeUntil := time.Until(nextRun)
			if timeUntil > 0 && timeUntil < sleepDuration {
				sleepDuration = timeUntil
			}
		}

		timer := time.NewTimer(sleepDuration)

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-d.stopCh:
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}

func (d *Dispatcher) Stop() {
	close(d.stopCh)
	d.wg.Wait()
}
