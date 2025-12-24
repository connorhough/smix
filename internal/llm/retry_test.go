package llm

import (
	"context"
	"errors"
	"testing"
)

func TestRetryWithBackoff(t *testing.T) {
	t.Run("succeeds on first try", func(t *testing.T) {
		callCount := 0
		fn := func(ctx context.Context) (string, error) {
			callCount++
			return "success", nil
		}

		result, err := RetryWithBackoff(context.Background(), fn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "success" {
			t.Errorf("got %q, want %q", result, "success")
		}
		if callCount != 1 {
			t.Errorf("got %d calls, want 1", callCount)
		}
	})

	t.Run("retries on transient error", func(t *testing.T) {
		callCount := 0
		fn := func(ctx context.Context) (string, error) {
			callCount++
			if callCount < 3 {
				return "", errors.New("network error")
			}
			return "success", nil
		}

		result, err := RetryWithBackoff(context.Background(), fn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "success" {
			t.Errorf("got %q, want %q", result, "success")
		}
		if callCount != 3 {
			t.Errorf("got %d calls, want 3", callCount)
		}
	})

	t.Run("fails after max retries", func(t *testing.T) {
		callCount := 0
		testErr := errors.New("persistent error")
		fn := func(ctx context.Context) (string, error) {
			callCount++
			return "", testErr
		}

		_, err := RetryWithBackoff(context.Background(), fn)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, testErr) {
			t.Errorf("expected error to wrap testErr")
		}
		if callCount != 3 {
			t.Errorf("got %d calls, want 3 (max retries)", callCount)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		fn := func(ctx context.Context) (string, error) {
			return "", errors.New("should not retry")
		}

		_, err := RetryWithBackoff(ctx, fn)
		if err == nil {
			t.Fatal("expected context error")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}
