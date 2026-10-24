package retry

import (
	"math/rand/v2"
	"time"
)

func CalculateBackoff(attempt uint8, baseDelay, maxDelay time.Duration) time.Duration {
	if attempt > 30 {
		attempt = 30
	}

	multiplier := time.Duration(1 << attempt)
	maxSleep := baseDelay * multiplier

	if maxSleep > maxDelay || maxSleep <= 0 {
		maxSleep = maxDelay
	}

	if maxSleep <= 0 {
		return 0
	}

	return time.Duration(rand.Int64N(int64(maxSleep) + 1))
}
