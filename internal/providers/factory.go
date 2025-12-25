package providers

import (
	"fmt"
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

// GetProvider returns a provider by name, creating and caching it if needed
func (f *Factory) GetProvider(name, apiKey string) (llm.Provider, error) {
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
func GetProvider(name, apiKey string) (llm.Provider, error) {
	return globalFactory.GetProvider(name, apiKey)
}
