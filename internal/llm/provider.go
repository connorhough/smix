// Package llm specifies the provider interface and contains implementations for supported providers
package llm

import "context"

// Provider defines the interface for LLM providers
type Provider interface {
	// Generate sends a prompt and returns the response
	Generate(ctx context.Context, prompt string, opts ...Option) (string, error)

	// ValidateModel can be used to check if a model name is valid for this provider.
	// Implementations may choose to perform no-op validation and let the Generate
	// method handle invalid model errors.
	ValidateModel(model string) error

	// DefaultModel returns the default model for this provider
	DefaultModel() string

	// Name returns the provider name (e.g., "claude", "gemini")
	Name() string
}
