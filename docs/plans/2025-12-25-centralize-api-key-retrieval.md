# Centralize API Key Retrieval Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Eliminate duplicated API key retrieval logic by centralizing it in the provider factory.

**Architecture:** Move API key retrieval from business logic packages (ask, do) into the providers.GetProvider factory function. The factory will internally retrieve the GEMINI_API_KEY from environment variables when creating Gemini providers, making business logic packages completely agnostic of provider-specific requirements.

**Tech Stack:** Go 1.x, os package for environment variables, existing provider architecture

---

## Task 1: Write failing test for API key retrieval in factory

**Files:**
- Modify: `internal/providers/factory_test.go:29-43`

**Step 1: Write the failing test**

Add new test case for API key retrieval from environment:

```go
// Add to factory_test.go after line 43
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
```

**Step 2: Run test to verify it fails**

Run: `go test -v ./internal/providers -run TestFactory_GetProvider`

Expected: FAIL - `GetProvider` has wrong signature (expects 2 parameters, test provides 1)

**Step 3: Commit**

```bash
git add internal/providers/factory_test.go
git commit -m "test: add failing tests for centralized API key retrieval"
```

---

## Task 2: Update factory signature and implementation

**Files:**
- Modify: `internal/providers/factory.go:25-73`

**Step 1: Write minimal implementation**

Update GetProvider to accept single parameter and retrieve API key internally:

```go
// GetProvider returns a provider by name, creating and caching it if needed
// API keys are retrieved from environment variables automatically
func (f *Factory) GetProvider(name string) (llm.Provider, error) {
	// Try to get from cache first (read lock)
	f.mu.RLock()
	if provider, ok := f.cache[name]; ok {
		f.mu.RUnlock()
		return provider, nil
	}
	f.mu.RUnlock()

	// Not in cache, create new provider (write lock)
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have created it)
	if provider, ok := f.cache[name]; ok {
		return provider, nil
	}

	// Create provider based on name
	var provider llm.Provider
	var err error

	switch name {
	case "claude":
		provider, err = claude.NewProvider()
	case "gemini":
		// Retrieve API key from environment
		apiKey := os.Getenv("GEMINI_API_KEY")
		provider, err = gemini.NewProvider(apiKey)
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}

	if err != nil {
		return nil, err
	}

	// Cache the provider
	f.cache[name] = provider

	return provider, nil
}

// Global factory instance
var globalFactory = NewFactory()

// GetProvider is a convenience function that uses the global factory
func GetProvider(name string) (llm.Provider, error) {
	return globalFactory.GetProvider(name)
}
```

**Step 2: Add os import**

Update imports at top of file:

```go
import (
	"fmt"
	"os"
	"sync"

	"github.com/connorhough/smix/internal/llm"
	"github.com/connorhough/smix/internal/llm/claude"
	"github.com/connorhough/smix/internal/llm/gemini"
)
```

**Step 3: Run tests to verify they pass**

Run: `go test -v ./internal/providers -run TestFactory_GetProvider`

Expected: PASS

**Step 4: Commit**

```bash
git add internal/providers/factory.go
git commit -m "refactor: centralize API key retrieval in provider factory"
```

---

## Task 3: Update ask package to use new factory signature

**Files:**
- Modify: `internal/ask/ask.go:37-50`

**Step 1: Write failing test**

Since no tests exist for ask package, we'll manually verify by building:

Run: `go build ./internal/ask`

Expected: FAIL - compilation error due to wrong number of arguments to GetProvider

**Step 2: Remove duplicated API key logic**

Replace lines 37-50 with simplified version:

```go
// Answer processes a user's question and returns a concise answer
func Answer(ctx context.Context, question string, cfg *config.ProviderConfig, debugFn func(string, ...interface{})) (string, error) {
	debugFn("ask command config: provider=%s, model=%s", cfg.Provider, cfg.Model)

	// Get provider from factory (API keys handled internally)
	provider, err := providers.GetProvider(cfg.Provider)
	if err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	debugFn("Using provider: %s", provider.Name())
```

**Step 3: Remove unused os import**

Update imports at top of file (remove "os"):

```go
import (
	"context"
	"fmt"

	"github.com/connorhough/smix/internal/config"
	"github.com/connorhough/smix/internal/llm"
	"github.com/connorhough/smix/internal/providers"
)
```

**Step 4: Verify it builds**

Run: `go build ./internal/ask`

Expected: SUCCESS

**Step 5: Run formatting**

Run: `gofmt -w internal/ask/ask.go`

Expected: File formatted

**Step 6: Commit**

```bash
git add internal/ask/ask.go
git commit -m "refactor(ask): remove duplicated API key retrieval logic"
```

---

## Task 4: Update do package to use new factory signature

**Files:**
- Modify: `internal/do/translate.go:44-57`

**Step 1: Write failing test**

Build to verify compilation error:

Run: `go build ./internal/do`

Expected: FAIL - compilation error due to wrong number of arguments to GetProvider

**Step 2: Remove duplicated API key logic**

Replace lines 44-57 with simplified version:

```go
// Translate converts natural language to shell commands
func Translate(ctx context.Context, taskDescription string, cfg *config.ProviderConfig, debugFn func(string, ...interface{})) (string, error) {
	debugFn("do command config: provider=%s, model=%s", cfg.Provider, cfg.Model)

	// Get provider from factory (API keys handled internally)
	provider, err := providers.GetProvider(cfg.Provider)
	if err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	debugFn("Using provider: %s", provider.Name())
```

**Step 3: Remove unused os import**

Update imports at top of file (remove "os"):

```go
import (
	"context"
	"fmt"

	"github.com/connorhough/smix/internal/config"
	"github.com/connorhough/smix/internal/llm"
	"github.com/connorhough/smix/internal/providers"
)
```

**Step 4: Verify it builds**

Run: `go build ./internal/do`

Expected: SUCCESS

**Step 5: Run formatting**

Run: `gofmt -w internal/do/translate.go`

Expected: File formatted

**Step 6: Commit**

```bash
git add internal/do/translate.go
git commit -m "refactor(do): remove duplicated API key retrieval logic"
```

---

## Task 5: Update all existing factory tests

**Files:**
- Modify: `internal/providers/factory_test.go:14-82`

**Step 1: Update existing test cases**

Update all existing GetProvider calls to use single parameter:

```go
func TestFactory_GetProvider(t *testing.T) {
	factory := NewFactory()

	t.Run("creates claude provider", func(t *testing.T) {
		provider, err := factory.GetProvider("claude")
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

	t.Run("creates gemini provider with API key from env", func(t *testing.T) {
		// Set up test API key in environment
		os.Setenv("GEMINI_API_KEY", "test-key-for-unit-test")
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
		p1, err := factory.GetProvider("claude")
		if err != nil {
			t.Skip("claude CLI not available")
		}

		p2, err := factory.GetProvider("claude")
		if err != nil {
			t.Fatalf("unexpected error on second call: %v", err)
		}

		// Should be same instance (pointer equality)
		if p1 != p2 {
			t.Error("expected cached provider instance")
		}
	})

	t.Run("fails for unknown provider", func(t *testing.T) {
		_, err := factory.GetProvider("unknown")
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
				_, _ = factory.GetProvider("claude")
			}()
		}
		wg.Wait()
		// If we get here without panic, thread safety works
	})
}
```

**Step 2: Run all tests**

Run: `go test -v ./internal/providers`

Expected: PASS (all tests pass)

**Step 3: Commit**

```bash
git add internal/providers/factory_test.go
git commit -m "test: update factory tests for centralized API key retrieval"
```

---

## Task 6: Build entire project and run all tests

**Files:**
- Verify: All packages

**Step 1: Build the project**

Run: `make build`

Expected: SUCCESS - binary created in builds/ directory

**Step 2: Run all tests**

Run: `make test`

Expected: PASS (all tests pass)

**Step 3: Run linter**

Run: `make lint`

Expected: PASS (no lint errors)

**Step 4: Verify commands still work**

If GEMINI_API_KEY is set:
```bash
./builds/smix ask "what is Go"
./builds/smix do "list files"
```

Expected: Commands execute successfully

**Step 5: Final commit if any fixes needed**

```bash
# Only if fixes were needed
git add .
git commit -m "fix: address build/test issues"
```

---

## Summary

This refactoring:
1. **Eliminates duplication** - API key retrieval logic exists in one place (factory)
2. **Improves separation of concerns** - Business logic packages (ask, do) don't know about provider-specific requirements
3. **Maintains backward compatibility** - Error messages and behavior remain the same
4. **Follows DRY principle** - Single source of truth for API key retrieval
5. **Simplifies future additions** - New providers with API keys can be added by updating only the factory

The factory now handles all provider-specific setup, making it easy to add more providers in the future.
