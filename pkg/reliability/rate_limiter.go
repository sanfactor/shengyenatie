package reliability

import (
	"context"
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting
type RateLimiter struct {
	rate       float64
	bucketSize int
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate float64, bucketSize int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		bucketSize: bucketSize,
		tokens:     float64(bucketSize),
		lastRefill: time.Now(),
	}
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.tokens = min(float64(rl.bucketSize), rl.tokens+elapsed*rl.rate)
	rl.lastRefill = now

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}

// WaitN waits for n tokens to become available
func (rl *RateLimiter) WaitN(ctx context.Context, n int) error {
	for n > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if rl.Allow() {
				n--
			} else {
				time.Sleep(time.Second / time.Duration(rl.rate))
			}
		}
	}
	return nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
