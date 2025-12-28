// Package gemini implements the gemini Provider interface.
package gemini

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/connorhough/smix/internal/llm"
)

const ProviderGemini = "gemini"

// Provider implements the llm.Provider interface for Gemini API
type Provider struct {
	client *genai.Client
	apiKey string
}

// Verify interface compliance at compile time
var _ llm.Provider = (*Provider)(nil)

// NewProvider creates a new Gemini provider
func NewProvider(apiKey string) (*Provider, error) {
	if apiKey == "" {
		return nil, llm.ErrAuthenticationFailed("gemini",
			fmt.Errorf("API key is required (set %s environment variable)", APIKeyEnvVar))
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, llm.ErrProviderNotAvailable("gemini", err)
	}

	return &Provider{
		client: client,
		apiKey: apiKey,
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

	// Execute with retry logic
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
