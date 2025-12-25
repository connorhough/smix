package providers

import (
	"os"
	"sync"
	"testing"

	"github.com/connorhough/smix/internal/llm"
)

func TestFactory_GetProvider(t *testing.T) {
	factory := NewFactory()

	t.Run("creates claude provider", func(t *testing.T) {
		provider, err := factory.GetProvider("claude", "")
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
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			apiKey = "test-key-for-unit-test"
		}

		provider, err := factory.GetProvider("gemini", apiKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if provider.Name() != "gemini" {
			t.Errorf("got provider %q, want %q", provider.Name(), "gemini")
		}
	})

	// TODO: The following tests use the new GetProvider signature (single parameter)
	// and will fail to compile until Task 2 refactors the factory to retrieve
	// API keys from environment variables internally. This is intentional TDD.

	t.Run("retrieves API key from environment for gemini", func(t *testing.T) {
		// Set up environment
		os.Setenv("GEMINI_API_KEY", "test-api-key-from-env")
		defer os.Unsetenv("GEMINI_API_KEY")

		// Create fresh factory to avoid cache
		newFactory := NewFactory()

		provider, err := newFactory.GetProvider("gemini")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if provider.Name() != "gemini" {
			t.Errorf("got provider %q, want %q", provider.Name(), "gemini")
		}
	})

	t.Run("fails for gemini without API key", func(t *testing.T) {
		// Ensure env var is not set
		os.Unsetenv("GEMINI_API_KEY")

		// Create fresh factory to avoid cache
		newFactory := NewFactory()

		_, err := newFactory.GetProvider("gemini")
		if err == nil {
			t.Error("expected error when GEMINI_API_KEY not set")
		}

		// Should be authentication error
		if _, ok := err.(*llm.ProviderError); !ok {
			t.Errorf("expected ProviderError, got %T", err)
		}
	})

	t.Run("returns cached provider", func(t *testing.T) {
		// Get provider twice
		p1, err := factory.GetProvider("claude", "")
		if err != nil {
			t.Skip("claude CLI not available")
		}

		p2, err := factory.GetProvider("claude", "")
		if err != nil {
			t.Fatalf("unexpected error on second call: %v", err)
		}

		// Should be same instance (pointer equality)
		if p1 != p2 {
			t.Error("expected cached provider instance")
		}
	})

	t.Run("fails for unknown provider", func(t *testing.T) {
		_, err := factory.GetProvider("unknown", "")
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
				_, _ = factory.GetProvider("claude", "")
			}()
		}
		wg.Wait()
		// If we get here without panic, thread safety works
	})
}
