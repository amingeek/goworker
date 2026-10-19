package task

import (
	"context"
	"time"
)

type State string

const (
	Pending   State = "PENDING"
	Running   State = "RUNNING"
	Completed State = "COMPLETED"
	Failed    State = "FAILED"
	DLQ       State = "DLQ"
)

type Handler func(ctx context.Context, payload []byte) error

type Task struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Payload    []byte    `json:"payload"`
	State      State     `json:"state"`
	MaxRetries uint8     `json:"max_retries"`
	Retries    uint8     `json:"retries"`
	RunAt      time.Time `json:"run_at"`
	CreatedAt  time.Time `json:"created_at"`
	LastError  string    `json:"last_error,omitempty"`
	Index      int       `json:"index"`
}

func (t *Task) IsReady() bool {
	now := time.Now()
	return now.After(t.RunAt) || now.Equal(t.RunAt)
}
