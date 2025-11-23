package utils

import (
	"context"
	"math/rand"
	"time"
)

// RetryFunc is the function to be retried.
type RetryFunc func() error

// RetryConfig defines the configuration for retry mechanism.
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Multiplier     float64
	Jitter         float64
}

// DefaultRetryConfig returns a default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     2 * time.Second,
		Multiplier:     2.0,
		Jitter:         0.1,
	}
}

// Retry executes the function with retry logic.
func Retry(ctx context.Context, fn RetryFunc, cfg RetryConfig) error {
	var err error
	backoff := cfg.InitialBackoff

	for i := 0; i <= cfg.MaxRetries; i++ {
		if err = fn(); err == nil {
			return nil
		}

		if i == cfg.MaxRetries {
			break
		}

		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(backoff):
		}

		// Calculate next backoff
		nextBackoff := float64(backoff) * cfg.Multiplier
		if cfg.Jitter > 0 {
			jitter := (rand.Float64()*2 - 1) * cfg.Jitter * nextBackoff
			nextBackoff += jitter
		}

		backoff = time.Duration(nextBackoff)
		if backoff > cfg.MaxBackoff {
			backoff = cfg.MaxBackoff
		}
	}

	return err
}
