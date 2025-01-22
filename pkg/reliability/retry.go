package reliability

import (
	"context"
	"math"
	"time"
)

// RetryConfig configures the retry behavior
type RetryConfig struct {
	MaxAttempts      int
	InitialDelay     time.Duration
	MaxDelay         time.Duration
	BackoffFactor    float64
	RetryableErrors  []error
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      10 * time.Second,
		BackoffFactor: 2.0,
	}
}

// Retry executes the function with retry logic
func Retry(ctx context.Context, fn func() error, config RetryConfig) error {
	var err error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !isRetryable(err, config.RetryableErrors) {
			return err
		}

		if attempt == config.MaxAttempts {
			break
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return err
}

// isRetryable checks if an error should be retried
func isRetryable(err error, retryableErrors []error) bool {
	if len(retryableErrors) == 0 {
		return true
	}
	for _, retryableErr := range retryableErrors {
		if err == retryableErr {
			return true
		}
	}
	return false
}

// Error types
type ReliabilityError string

func (e ReliabilityError) Error() string { return string(e) }

const (
	ErrCircuitOpen = ReliabilityError("circuit breaker is open")
	ErrRateLimited = ReliabilityError("rate limit exceeded")
)
