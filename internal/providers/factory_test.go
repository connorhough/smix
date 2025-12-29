package providers

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/connorhough/smix/internal/llm"
	"github.com/connorhough/smix/internal/llm/gemini"
)

func TestFactory_GetProvider(t *testing.T) {
	factory := NewFactory()
	ctx := context.Background()

	t.Run("creates claude provider", func(t *testing.T) {
		provider, err := factory.GetProvider(ctx, "claude")
		if err != nil {
			// Skip if claude CLI not available
			if _, ok := err.(*llm.ProviderError); ok {
				t.Skip("claude CLI not available")
			}
			t.Fatalf("unexpected error: %v", err)
		}

		if provider.Name() != "claude" {
			t.Errorf("got provider %q, want %q", provider.Name(), "claude")
		}
	})

	t.Run("creates gemini provider with API key", func(t *testing.T) {
		apiKey := os.Getenv(gemini.APIKeyEnvVar)
		if apiKey == "" {
			apiKey = "test-key-for-unit-test"
			os.Setenv(gemini.APIKeyEnvVar, apiKey)
			defer os.Unsetenv(gemini.APIKeyEnvVar)
		}

		provider, err := factory.GetProvider(ctx, "gemini")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if provider.Name() != "gemini" {
			t.Errorf("got provider %q, want %q", provider.Name(), "gemini")
		}
	})

	t.Run("retrieves API key from environment for gemini", func(t *testing.T) {
		// Set up environment
		os.Setenv(gemini.APIKeyEnvVar, "test-api-key-from-env")
		defer os.Unsetenv(gemini.APIKeyEnvVar)

		// Create fresh factory to avoid cache
		newFactory := NewFactory()

		provider, err := newFactory.GetProvider(ctx, "gemini")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if provider.Name() != "gemini" {
			t.Errorf("got provider %q, want %q", provider.Name(), "gemini")
		}
	})

	t.Run("fails for gemini without API key or CLI", func(t *testing.T) {
		// Ensure env var is not set
		os.Unsetenv(gemini.APIKeyEnvVar)

		// Create fresh factory to avoid cache
		newFactory := NewFactory()

		provider, err := newFactory.GetProvider(ctx, "gemini")

		// If CLI is available, provider creation succeeds
		if err == nil {
			if provider.Name() != "gemini" {
				t.Errorf("got provider %q, want %q", provider.Name(), "gemini")
			}
			// Valid: CLI-only mode - test passes
			return
		}

		// If error, should be a ProviderError (neither API key nor CLI)
		if _, ok := err.(*llm.ProviderError); !ok {
			t.Errorf("expected ProviderError, got %T: %v", err, err)
		}
	})

	t.Run("returns cached provider", func(t *testing.T) {
		// Get provider twice
		p1, err := factory.GetProvider(ctx, "claude")
		if err != nil {
			t.Skip("claude CLI not available")
		}

		p2, err := factory.GetProvider(ctx, "claude")
		if err != nil {
			t.Fatalf("unexpected error on second call: %v", err)
		}

		// Should be same instance (pointer equality)
		if p1 != p2 {
			t.Error("expected cached provider instance")
		}
	})

	t.Run("fails for unknown provider", func(t *testing.T) {
		_, err := factory.GetProvider(ctx, "unknown")
		if err == nil {
			t.Error("expected error for unknown provider")
		}
	})

	t.Run("thread-safe concurrent access", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = factory.GetProvider(ctx, "claude")
			}()
		}
		wg.Wait()
		// If we get here without panic, thread safety works
	})
}