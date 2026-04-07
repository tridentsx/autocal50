package core

import (
	"context"
	"fmt"
	"time"
)

// Retry calls fn up to maxAttempts times with exponential backoff.
// Returns the first successful result or the last error.
func Retry[T any](ctx context.Context, maxAttempts int, fn func() (T, error)) (T, error) {
	var lastErr error
	var zero T
	delay := 500 * time.Millisecond

	for i := 0; i < maxAttempts; i++ {
		result, err := fn()
		if err == nil {
			return result, nil
		}
		lastErr = err
		if i < maxAttempts-1 {
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(delay):
			}
			delay *= 2
			if delay > 5*time.Second {
				delay = 5 * time.Second
			}
		}
	}
	return zero, fmt.Errorf("after %d attempts: %w", maxAttempts, lastErr)
}

// RetryVoid is Retry for functions that return only an error.
func RetryVoid(ctx context.Context, maxAttempts int, fn func() error) error {
	_, err := Retry(ctx, maxAttempts, func() (struct{}, error) {
		return struct{}{}, fn()
	})
	return err
}
