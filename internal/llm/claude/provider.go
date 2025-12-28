// Package claude implements the claude Provider interface.
package claude

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/connorhough/smix/internal/llm"
)

const ProviderClaude = "claude"

// Provider implements the llm.Provider interface for Claude CLI
type Provider struct {
	cliPath string
}

// NewProvider creates a new Claude provider
func NewProvider() (*Provider, error) {
	cliPath, err := exec.LookPath(ProviderClaude)
	if err != nil {
		return nil, llm.ErrProviderNotAvailable(ProviderClaude, err)
	}

	return &Provider{
		cliPath: cliPath,
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return ProviderClaude
}

// DefaultModel returns the default model for Claude
func (p *Provider) DefaultModel() string {
	return DefaultModel()
}

// ValidateModel checks if a model is valid
func (p *Provider) ValidateModel(model string) error {
	return nil // No pre-validation, let CLI handle it
}

// Generate sends a prompt to Claude and returns the response
func (p *Provider) Generate(ctx context.Context, prompt string, opts ...llm.Option) (string, error) {
	options := llm.BuildOptions(opts)

	model := options.Model
	if model == "" {
		model = p.DefaultModel()
	}

	cmd := exec.CommandContext(ctx, p.cliPath, "--model", model, "-p", prompt)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claude CLI failed: %w (output: %s)", err, output)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", fmt.Errorf("claude CLI returned empty response")
	}

	return result, nil
}

// Verify interface compliance at compile time
var _ llm.InteractiveProvider = (*Provider)(nil)

// RunInteractive implements the llm.InteractiveProvider interface.
// It starts an interactive Claude session, connecting the provided IOStreams
// to the claude CLI process. This allows the CLI to display colored output,
// handle user interaction, and show real-time streaming responses.
//
// IMPORTANT: The caller MUST verify streams.IsInteractive() returns true before
// calling this method. This implementation does NOT check for TTY - that is the
// caller's responsibility.
//
// This method is suitable for interactive workflows like the pr command where
// the user needs to see formatted output and potentially interact with Claude.
// It should NOT be used for commands that need clean, parseable output.
func (p *Provider) RunInteractive(ctx context.Context, streams *llm.IOStreams, prompt string, opts ...llm.Option) error {
	options := llm.BuildOptions(opts)

	model := options.Model
	if model == "" {
		model = p.DefaultModel()
	}

	// Build command with model and prompt
	// Note: Using bare prompt as argument (not -p flag) triggers interactive mode
	// The --model flag selects which Claude model to use for the session
	cmd := exec.CommandContext(ctx, p.cliPath, "--model", model, prompt)

	// Connect provided streams to allow interactive mode
	// These may be os.Stdin/Stdout/Stderr (production) or buffers (testing)
	cmd.Stdin = streams.In
	cmd.Stdout = streams.Out
	cmd.Stderr = streams.ErrOut

	// Run and wait for completion
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("claude CLI interactive mode failed: %w", err)
	}

	return nil
}
