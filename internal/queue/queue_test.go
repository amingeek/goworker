package queue_test

import (
	"sync"
	"testing"
	"time"

	"github.com/amingeek/goworker/internal/queue"
	"github.com/amingeek/goworker/internal/task"
)

// TestSafeQueue_PriorityOrder تست می‌کند که آیا تسک‌ها دقیقاً بر اساس زمان اجرا خارج می‌شوند یا نه
func TestSafeQueue_PriorityOrder(t *testing.T) {
	q := queue.NewSafeQueue()
	now := time.Now()

	// سه تسک با زمان‌های متفاوت
	tLate := &task.Task{ID: "late", RunAt: now.Add(500 * time.Millisecond)}
	tUrgent := &task.Task{ID: "urgent", RunAt: now.Add(-100 * time.Millisecond)}
	tNow := &task.Task{ID: "now", RunAt: now}

	// درج بدون ترتیب در صف
	q.Push(tLate)
	q.Push(tUrgent)
	q.Push(tNow)

	// ۱. اولین خروجی حتماً باید urgent باشد (چون زمانش قبل‌تر از همه بوده)
	popped := q.PopReady()
	if popped == nil || popped.ID != "urgent" {
		t.Fatalf("expected urgent, got %v", popped)
	}

	// ۲. دومین خروجی باید now باشد
	popped = q.PopReady()
	if popped == nil || popped.ID != "now" {
		t.Fatalf("expected now, got %v", popped)
	}

	// ۳. تسک late هنوز وقتش نرسیده؛ پس PopReady باید nil بدهد
	popped = q.PopReady()
	if popped != nil {
		t.Fatalf("expected nil for future task, got %v", popped)
	}

	// ۴. اگر به اندازه کافی صبر کنیم تا زمان late برسد، باید با موفقیت خارج شود
	time.Sleep(550 * time.Millisecond)
	popped = q.PopReady()
	if popped == nil || popped.ID != "late" {
		t.Fatalf("expected late after wait, got %v", popped)
	}
}

// TestSafeQueue_ConcurrentPush تست همروندی و عدم وجود Data Race زیر بار همزمان
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
