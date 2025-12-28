// Package llm specifies the provider interface and contains implementations for supported providers
package llm

import (
	"context"
)

// Provider defines the interface for LLM providers
type Provider interface {
	// Generate sends a prompt and returns the response
	Generate(ctx context.Context, prompt string, opts ...Option) (string, error)

	// ValidateModel can be used to check if a model name is valid for this provider.
	ValidateModel(model string) error

	// DefaultModel returns the default model name for this provider
	DefaultModel() string

	// Name returns the provider name (e.g., "claude", "gemini")
	Name() string
}

// InteractiveProvider is an optional interface for providers that support
// yielding control of I/O streams for stateful, interactive sessions.
//
// Implementations can either:
//   - Delegate to an external CLI tool (e.g., claude code) by spawning a subprocess
//     and connecting the provided IOStreams
//   - Implement a REPL loop directly in Go against an API (e.g., streaming Gemini API)
//     by reading from streams.In and writing to streams.Out
//
// IMPORTANT: The caller is responsible for checking streams.IsInteractive() before
// calling RunInteractive. Providers should NOT check TTY status themselves.
type InteractiveProvider interface {
	// RunInteractive starts an interactive session using the provided IOStreams.
	// The prompt parameter is passed to the provider as the initial input.
	// Options like WithModel can be used to configure the session.
	//
	// The caller MUST verify streams.IsInteractive() returns true before calling
	// this method. If called with non-interactive streams (pipes, redirects, CI/CD),
	// behavior is undefined (the session may fail or hang).
	//
	// Implementations should connect their subprocess or REPL to:
	// - streams.In for user input
	// - streams.Out for normal output
	// - streams.ErrOut for error messages
	//
	// Returns an error if:
	// - The provider cannot start an interactive session
	// - The session fails or is cancelled by the user
	// - Context is cancelled during the session
	//
	// Example usage:
	//   streams := llm.NewIOStreams()
	//   if !streams.IsInteractive() {
	//       return fmt.Errorf("interactive mode requires a terminal")
	//   }
	//   if ip, ok := provider.(llm.InteractiveProvider); ok {
	//       return ip.RunInteractive(ctx, streams, prompt, opts...)
	//   }
	RunInteractive(ctx context.Context, streams *IOStreams, prompt string, opts ...Option) error
}
