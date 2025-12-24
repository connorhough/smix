package claude

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/connorhough/smix/internal/llm"
)

// Provider implements the llm.Provider interface for Claude CLI
type Provider struct {
	cliPath string
}

// NewProvider creates a new Claude provider
func NewProvider() (*Provider, error) {
	// Check if claude CLI is available
	cliPath, err := exec.LookPath("claude")
	if err != nil {
		return nil, llm.ErrProviderNotAvailable("claude", err)
	}

	return &Provider{
		cliPath: cliPath,
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "claude"
}

// DefaultModel returns the default model for Claude
func (p *Provider) DefaultModel() string {
	return DefaultModel()
}

// ValidateModel checks if a model is valid
// Claude CLI will fail with helpful error if model is invalid,
// so we let it fail naturally and wrap the error
func (p *Provider) ValidateModel(model string) error {
	return nil // No pre-validation, let CLI handle it
}

// Generate sends a prompt to Claude and returns the response
func (p *Provider) Generate(ctx context.Context, prompt string, opts ...llm.Option) (string, error) {
	options := llm.BuildOptions(opts)

	// Use provided model or default
	model := options.Model
	if model == "" {
		model = p.DefaultModel()
	}

	// Execute with retry logic
	return llm.RetryWithBackoff(ctx, func(ctx context.Context) (string, error) {
		// Create fresh command for each attempt
		cmd := exec.CommandContext(ctx, p.cliPath, "--model", model, "-p", prompt)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Wrap CLI errors with context
			return "", fmt.Errorf("claude CLI failed: %w (output: %s)", err, output)
		}

		result := strings.TrimSpace(string(output))
		if result == "" {
			return "", fmt.Errorf("claude CLI returned empty response")
		}

		return result, nil
	})
}
