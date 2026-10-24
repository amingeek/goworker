package retry_test

import (
	"testing"
	"time"

	"github.com/amingeek/goworker/pkg/retry"
)

func TestCalculateBackoff_Boundaries(t *testing.T) {
	base := 100 * time.Millisecond
	max := 2 * time.Second

	for attempt := uint8(0); attempt <= 40; attempt++ {
		d := retry.CalculateBackoff(attempt, base, max)
		if d < 0 {
			t.Fatalf("attempt %d resulted in negative delay: %v", attempt, d)
		}
		if d > max {
			t.Fatalf("attempt %d exceeded maxDelay: got %v, max %v", attempt, d, max)
		}
	}
}

func TestCalculateBackoff_ZeroDelay(t *testing.T) {
	d := retry.CalculateBackoff(5, 0, 0)
	if d != 0 {
		t.Fatalf("expected 0 for zero inputs, got %v", d)
	}
}

func TestCalculateBackoff_JitterDistribution(t *testing.T) {
	base := 500 * time.Millisecond
	max := 5 * time.Second

	results := make(map[time.Duration]bool)
	for i := 0; i < 50; i++ {
		d := retry.CalculateBackoff(3, base, max)
		results[d] = true
	}

	if len(results) < 10 {
		t.Fatalf("jitter seems ineffective, only %d unique values out of 50 runs", len(results))
	}
}
