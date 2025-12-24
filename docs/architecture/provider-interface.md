# Provider Interface Design

## Interface Definition

```go
package llm

import "context"

// Provider defines the interface for LLM providers
type Provider interface {
	// Generate sends a prompt and returns the response
	Generate(ctx context.Context, prompt string, opts ...Option) (string, error)

	// ValidateModel checks if a model name is valid for this provider
	// Returns helpful error message if invalid
	ValidateModel(model string) error

	// DefaultModel returns the default model for this provider
	DefaultModel() string

	// Name returns the provider name (e.g., "claude", "gemini")
	Name() string
}

// Option configures provider behavior
type Option func(*GenerateOptions)

type GenerateOptions struct {
	Model string
}
```

## Design Decisions

1. **Context Support:** All operations take context.Context for cancellation/timeout
2. **Model Validation:** Let API/CLI fail naturally, wrap errors with helpful context
3. **Options Pattern:** Functional options for flexibility (model override, future params)
4. **No Streaming:** Simple request-response sufficient for ask/do commands

## Error Handling

Providers should wrap errors with typed errors defined in `errors.go`:
- `ErrProviderNotAvailable`: CLI not found or SDK client creation failed
- `ErrAuthenticationFailed`: API key invalid/missing
- `ErrRateLimitExceeded`: Provider rate limit hit
- `ErrModelNotFound`: Model name invalid (wrapped from API/CLI error)

## Retry Logic

Network errors trigger exponential backoff retry:
- Max retries: 3
- Initial delay: 1s
- Max delay: 30s
- Backoff factor: 2
