package queue

import (
	"container/heap"
	"sync"
	"time"

	"github.com/amingeek/goworker/internal/task"
)

type priorityQueue []*task.Task

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].RunAt.Before(pq[j].RunAt)
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *priorityQueue) Push(x any) {
	n := len(*pq)
	item := x.(*task.Task)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *priorityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

type SafeQueue struct {
	mu sync.Mutex
	pq priorityQueue
}

func NewSafeQueue() *SafeQueue {
	sq := &SafeQueue{
		pq: make(priorityQueue, 0),
	}
	heap.Init(&sq.pq)
	return sq
}

func (q *SafeQueue) Push(t *task.Task) {
	q.mu.Lock()
	defer q.mu.Unlock()
	heap.Push(&q.pq, t)
}

func (q *SafeQueue) PopReady() *task.Task {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.pq) == 0 {
		return nil
	}

	if !q.pq[0].IsReady() {
		return nil
	}

	return heap.Pop(&q.pq).(*task.Task)
}

func (q *SafeQueue) NextRunAt() (time.Time, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.pq) == 0 {
		return time.Time{}, false
	}
	return q.pq[0].RunAt, true
}

func (q *SafeQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pq)
}
