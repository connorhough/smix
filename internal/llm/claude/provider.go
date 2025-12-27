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
