package queue_test

import (
	"sync"
	"testing"
	"time"

	"github.com/amingeek/goworker/internal/queue"
	"github.com/amingeek/goworker/internal/task"
)

func TestSafeQueue_PriorityOrder(t *testing.T) {
	q := queue.NewSafeQueue()
	now := time.Now()

	tLate := &task.Task{ID: "late", RunAt: now.Add(500 * time.Millisecond)}
	tUrgent := &task.Task{ID: "urgent", RunAt: now.Add(-100 * time.Millisecond)}
	tNow := &task.Task{ID: "now", RunAt: now}

	q.Push(tLate)
	q.Push(tUrgent)
	q.Push(tNow)

	popped := q.PopReady()
	if popped == nil || popped.ID != "urgent" {
		t.Fatalf("expected urgent, got %v", popped)
	}

	popped = q.PopReady()
	if popped == nil || popped.ID != "now" {
		t.Fatalf("expected now, got %v", popped)
	}

	popped = q.PopReady()
	if popped != nil {
		t.Fatalf("expected nil for future task, got %v", popped)
	}

	time.Sleep(550 * time.Millisecond)
	popped = q.PopReady()
	if popped == nil || popped.ID != "late" {
		t.Fatalf("expected late after wait, got %v", popped)
	}
}

func TestSafeQueue_ConcurrentPush(t *testing.T) {
	q := queue.NewSafeQueue()
	var wg sync.WaitGroup
	count := 2000

	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(id int) {
			defer wg.Done()
			q.Push(&task.Task{
				ID:    "task",
				RunAt: time.Now(),
			})
		}(i)
	}

	wg.Wait()

	if q.Len() != count {
		t.Fatalf("expected queue length %d, got %d", count, q.Len())
	}
}
