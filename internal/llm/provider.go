package llm

import "context"

// Provider defines the interface for LLM providers
type Provider interface {
	// Generate sends a prompt and returns the response
	Generate(ctx context.Context, prompt string, opts ...Option) (string, error)

	// ValidateModel checks if a model name is valid for this provider
	// Returns error with helpful message if invalid
	ValidateModel(model string) error

	// DefaultModel returns the default model for this provider
	DefaultModel() string

	// Name returns the provider name (e.g., "claude", "gemini")
	Name() string
}
