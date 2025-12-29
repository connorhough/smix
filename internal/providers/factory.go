// Package providers implements the provider factory.
package providers

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/connorhough/smix/internal/llm"
	"github.com/connorhough/smix/internal/llm/claude"
	"github.com/connorhough/smix/internal/llm/gemini"
)

// Factory creates and caches provider instances
type Factory struct {
	cache map[string]llm.Provider
	mu    sync.RWMutex
}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	return &Factory{
		cache: make(map[string]llm.Provider),
	}
}

// GetProvider returns a provider by name
func (f *Factory) GetProvider(ctx context.Context, name string) (llm.Provider, error) {
	f.mu.RLock()
	if provider, ok := f.cache[name]; ok {
		f.mu.RUnlock()
		return provider, nil
	}
	f.mu.RUnlock()

	f.mu.Lock()
	defer f.mu.Unlock()

	if provider, ok := f.cache[name]; ok {
		return provider, nil
	}

	var provider llm.Provider
	var err error

	switch name {
	case claude.ProviderClaude:
		provider, err = claude.NewProvider()
	case gemini.ProviderGemini:
		apiKey := os.Getenv(gemini.APIKeyEnvVar)
		provider, err = gemini.NewProvider(ctx, apiKey)
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}

	if err != nil {
		return nil, err
	}

	f.cache[name] = provider

	return provider, nil
}

// Global factory instance
var globalFactory = NewFactory()

// GetProvider is a convenience function that uses the global factory
func GetProvider(ctx context.Context, name string) (llm.Provider, error) {
	return globalFactory.GetProvider(ctx, name)
}
