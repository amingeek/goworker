# GoWorker ⚡

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-Passing-success)](./internal/queue/queue_test.go)

**GoWorker** is a high-performance, resilient, in-memory distributed task scheduling and priority queue engine built from scratch in Go. 

It is designed to handle asynchronous background execution, delayed jobs, intelligent exponential backoff with full jitter, and persistent crash recovery using SQLite (WAL mode).

---

## 🏗 Architecture Overview

```text
┌────────────────────────────────────────────────────────┐
│                      Client / API                      │
│            Enqueues: Task{Type, Payload, RunAt}        │
└───────────────────────────┬────────────────────────────┘
                            │
                            ▼
┌────────────────────────────────────────────────────────┐
│                    Engine / Scheduler                  │
│  - Persists initial state to DB (Status: PENDING)      │
│  - Inserts task into in-memory priority queue (RunAt)  │
└───────────────┬────────────────────────┬───────────────┘
                │                        │
                ▼                        ▼
     ┌────────────────────┐   ┌──────────────────────────┐
     │ In-Memory Min-Heap │   │       SQLite (WAL)       │
     │   (Priority Queue) │   │  (Durability & Recovery) │
     └──────────┬─────────┘   └──────────────────────────┘
                │
                ▼
┌────────────────────────────────────────────────────────┐
│                      Worker Pool                       │
│  - N Concurrent Workers processing tasks               │
│  - Decoupled execution via Go Channels & Context       │
│  - Handlers execution & dynamic error propagation      │
│  - On Failure: Exponential Backoff + Jitter reschedule │
│  - On Exhaustion: Move to Dead Letter Queue (DLQ)      │
└────────────────────────────────────────────────────────┘
```

---

## ✨ Key Features

- **In-Memory Priority Queue:** Thread-safe, sub-millisecond task indexing powered by a Min-Heap (`container/heap`) ordered by execution timestamp (`RunAt`).
- **Delayed & Scheduled Jobs:** Precise execution of future tasks without busy-waiting or thread starvation.
- **Worker Pool with Graceful Shutdown:** Controlled concurrency utilizing Go channels and `context.Context` cancellation propagation.
- **Resilience & Fault Tolerance:**
  - **Exponential Backoff with Full Jitter:** Prevents thundering herd problems when interacting with downstream services.
  - **Dead Letter Queue (DLQ):** Isolates permanently failed tasks for manual inspection without stalling the main queue.
- **Crash Recovery & Durability:** SQLite Write-Ahead Logging (WAL) state persistence ensures zero task loss across unexpected server shutdowns.
- **Zero Third-Party Core Runtime Dependencies:** Written purely using Go standard libraries for the critical scheduling and queueing paths.

---

## 📁 Project Structure

```text
goworker/
├── cmd/
│   └── worker/
│       └── main.go          # Application entrypoint & graceful shutdown
├── internal/
│   ├── queue/              # Thread-safe Min-Heap Priority Queue
│   │   ├── heap.go
│   │   └── queue_test.go   # Concurrency & race condition tests
│   ├── scheduler/          # Worker pool and timing loop
│   ├── storage/            # SQLite WAL persistence layer
│   └── task/               # Task domain model & state definitions
├── pkg/
│   └── retry/              # Exponential backoff and jitter algorithms
├── go.mod
└── README.md
```

---

## 🚦 Roadmap & Implementation Phases

- [x] **Phase 1: In-Memory Priority Queue Core**
  - [x] Task domain model & lifecycle state machine (`PENDING`, `RUNNING`, `COMPLETED`, `FAILED`, `DLQ`).
  - [x] Concurrency-safe Min-Heap implementation (`SafeQueue`).
  - [x] Race condition tests passed (`go test -race`).
- [ ] **Phase 2: Retry Engine & Backoff**
  - [ ] Exponential backoff calculation with Full Jitter.
  - [ ] Dead Letter Queue (DLQ) dispatching.
- [ ] **Phase 3: Worker Pool & Engine Dispatcher**
  - [ ] Goroutine worker pool with channel buffering.
  - [ ] Context cancellation and OS signal trapping (`SIGINT`, `SIGTERM`).
- [ ] **Phase 4: SQLite WAL Persistence & Crash Recovery**
  - [ ] Outbox pattern implementation for local durability.
  - [ ] Recovery on boot for uncompleted tasks.
- [ ] **Phase 5: CLI & Prometheus Observability**

---

## 🧪 Testing

Run all unit tests and verify thread safety with Go's race detector:

```bash
go test -v -race ./...
```

---

## 👤 Author

**Seyyed Amin Pourseyyedy**
- GitHub: [@amingeek](https://github.com/amingeek)
