package llm

import (
	"context"
	"fmt"
	"time"
)

const (
	maxRetries   = 3
	initialDelay = 1 * time.Second
	maxDelay     = 30 * time.Second
	backoffRate  = 2.0
)

// RetryWithBackoff executes a function with exponential backoff retry logic.
// It retries up to maxRetries times with exponential backoff starting
// at initialDelay and capping at maxDelay. The delay increases by a factor of
// backoffRate after each failed attempt
//
// Context cancellation is respected at two points:
// 1. Before each attempt
// 2. During the sleep delay between attempts
//
// Returns the last error wrapped with retry count if all attempts fail.
func RetryWithBackoff(ctx context.Context, fn func(context.Context) (string, error)) (string, error) {
	var lastErr error
	delay := initialDelay

	for attempt := range maxRetries {
		if err := ctx.Err(); err != nil {
			return "", err
		}

		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't sleep after last attempt
		if attempt < maxRetries-1 {
			select {
			case <-time.After(delay):
				delay = min(
					time.Duration(float64(delay)*backoffRate),
					maxDelay,
				)
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}

	return "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
