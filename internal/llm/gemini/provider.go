// Package gemini implements the gemini Provider interface.
package gemini

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"google.golang.org/genai"

	"github.com/connorhough/smix/internal/llm"
)

const ProviderGemini = "gemini"

// Provider implements the llm.Provider interface for Gemini API
type Provider struct {
	client  *genai.Client
	apiKey  string
	cliPath string // Optional: path to gemini CLI for interactive mode
}

// Verify interface compliance at compile time
var (
	_ llm.Provider            = (*Provider)(nil)
	_ llm.InteractiveProvider = (*Provider)(nil)
)

// NewProvider creates a new Gemini provider
func NewProvider(ctx context.Context, apiKey string) (*Provider, error) {
	var client *genai.Client
	var err error

	// detect gemini CLI for interactive mode (non-fatal if not found)
	cliPath, _ := exec.LookPath("gemini")

	if apiKey == "" && cliPath == "" {
		return nil, llm.ErrAuthenticationFailed("gemini",
			fmt.Errorf("API key is required (set %s environment variable) or Gemini CLI must be installed", APIKeyEnvVar))
	}

	if apiKey != "" {
		client, err = genai.NewClient(ctx, &genai.ClientConfig{
			APIKey: apiKey,
		})
		if err != nil {
			return nil, llm.ErrProviderNotAvailable("gemini", err)
		}
	}

	return &Provider{
		client:  client,
		apiKey:  apiKey,
		cliPath: cliPath,
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return ProviderGemini
}

// DefaultModel returns the default model for Gemini
func (p *Provider) DefaultModel() string {
	return DefaultModel()
}

// ValidateModel checks if a model is valid
func (p *Provider) ValidateModel(model string) error {
	return nil // No pre-validation, let API handle it
}

// Generate sends a prompt to Gemini and returns the response
func (p *Provider) Generate(ctx context.Context, prompt string, opts ...llm.Option) (string, error) {
	options := llm.BuildOptions(opts)

	// Use provided model or default
	modelName := options.Model
	if modelName == "" {
		modelName = p.DefaultModel()
	}

	// Use CLI if no API client available
	if p.client == nil {
		return p.generateViaCLI(ctx, modelName, prompt)
	}

	// Execute with retry logic (API path)
	return llm.RetryWithBackoff(ctx, func(ctx context.Context) (string, error) {
		resp, err := p.client.Models.GenerateContent(ctx, modelName, genai.Text(prompt), nil)
		if err != nil {
			return "", p.wrapError(err, modelName)
		}

		// Extract text from response
		if len(resp.Candidates) == 0 {
			return "", fmt.Errorf("gemini API returned no candidates")
		}

		if resp.Candidates[0].Content == nil {
			return "", fmt.Errorf("gemini API returned nil content")
		}

		var result strings.Builder
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				result.WriteString(part.Text)
			}
		}

		output := strings.TrimSpace(result.String())
		if output == "" {
			return "", fmt.Errorf("gemini API returned empty response")
		}

		return output, nil
	})
}

// generateViaCLI runs the gemini CLI in non-interactive mode and returns the output.
// Command format: gemini --model {model} "{prompt}"
func (p *Provider) generateViaCLI(ctx context.Context, modelName, prompt string) (string, error) {
	if p.cliPath == "" {
		return "", fmt.Errorf("gemini CLI not available")
	}

	cmd := exec.CommandContext(ctx, p.cliPath, "--model", modelName, prompt)
	output, err := cmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("gemini CLI failed: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("gemini CLI failed: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "", fmt.Errorf("gemini CLI returned empty response")
	}

	return result, nil
}

// wrapError wraps Gemini API errors with appropriate typed errors
func (p *Provider) wrapError(err error, modelName string) error {
	var apiErr genai.APIError
	if errors.As(err, &apiErr) {
		// Prefer checking the structured Status field over string matching
		switch apiErr.Status {
		case "INVALID_ARGUMENT":
			if strings.Contains(apiErr.Message, "API key") {
				return llm.ErrAuthenticationFailed("gemini", err)
			}
		case "RESOURCE_EXHAUSTED":
			return llm.ErrRateLimitExceeded("gemini", err)
		case "NOT_FOUND":
			if strings.Contains(apiErr.Message, "model") {
				return llm.ErrModelNotFound(modelName, "gemini", err)
			}
		}
		// For other genai.APIError types, return a structured error message
		return fmt.Errorf("gemini API error: status %s, code %d, %w", apiErr.Status, apiErr.Code, err)
	}

	// Generic fallback for non-APIError types (e.g., network errors)
	return fmt.Errorf("gemini API error: %w", err)
}

// RunInteractive implements the llm.InteractiveProvider interface.
// It starts an interactive Gemini CLI session, connecting the provided IOStreams
// to the gemini CLI process. This allows the CLI to display output,
// handle user interaction, and show real-time streaming responses.
//
// Returns an error if the gemini CLI is not installed. Install with:
// npm install -g @google/gemini-cli
func (p *Provider) RunInteractive(ctx context.Context, streams *llm.IOStreams, prompt string, opts ...llm.Option) error {
	if p.cliPath == "" {
		return fmt.Errorf("gemini CLI not available for interactive mode; install with: npm install -g @google/gemini-cli")
	}

	options := llm.BuildOptions(opts)

	model := options.Model
	if model == "" {
		model = p.DefaultModel()
	}

	// Build command with model and prompt
	// gemini CLI uses -m for model and accepts prompt as positional argument
	cmd := exec.CommandContext(ctx, p.cliPath, "-m", model, prompt)

	// Connect provided streams to allow interactive mode
	cmd.Stdin = streams.In
	cmd.Stdout = streams.Out
	cmd.Stderr = streams.ErrOut

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gemini CLI interactive mode failed: %w", err)
	}

	return nil
}

// HasInteractiveSupport returns true if the gemini CLI is available for interactive sessions.
// This can be used to provide better error messages before attempting interactive mode.
func (p *Provider) HasInteractiveSupport() bool {
	return p.cliPath != ""
}
