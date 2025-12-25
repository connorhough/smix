package gemini

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/genai"

	"github.com/connorhough/smix/internal/llm"
)

// Provider implements the llm.Provider interface for Gemini API
type Provider struct {
	client *genai.Client
	apiKey string
}

// NewProvider creates a new Gemini provider
func NewProvider(apiKey string) (*Provider, error) {
	if apiKey == "" {
		return nil, llm.ErrAuthenticationFailed("gemini",
			fmt.Errorf("API key is required (set GEMINI_API_KEY environment variable)"))
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

// Close cleans up the provider's resources
// Note: genai.Client does not have a Close method
func (p *Provider) Close() error {
	return nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "gemini"
}

// DefaultModel returns the default model for Gemini
func (p *Provider) DefaultModel() string {
	return DefaultModel()
}

// ValidateModel checks if a model is valid
// Gemini API will fail with helpful error if model is invalid,
// so we let it fail naturally and wrap the error
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
			return "", p.wrapError(err)
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
func (p *Provider) wrapError(err error) error {
	// Check for genai.APIError using type assertion
	if apiErr, ok := err.(genai.APIError); ok {
		// Check status first (most reliable)
		switch apiErr.Status {
		case "INVALID_ARGUMENT":
			if strings.Contains(apiErr.Message, "API key") {
				return llm.ErrAuthenticationFailed("gemini", err)
			}
		case "RESOURCE_EXHAUSTED":
			// Rate limit or quota exceeded
			return llm.ErrRateLimitExceeded("gemini")
		case "NOT_FOUND":
			// Model not found
			return llm.ErrModelNotFound("unknown", "gemini")
		}

		// Fallback to HTTP status code
		switch apiErr.Code {
		case 400:
			if strings.Contains(apiErr.Message, "API key") {
				return llm.ErrAuthenticationFailed("gemini", err)
			}
		case 429:
			return llm.ErrRateLimitExceeded("gemini")
		case 404:
			return llm.ErrModelNotFound("unknown", "gemini")
		}
	}

	// Fallback to string matching for non-APIError types
	errMsg := err.Error()
	if strings.Contains(errMsg, "API key") || strings.Contains(errMsg, "authentication") {
		return llm.ErrAuthenticationFailed("gemini", err)
	}

	if strings.Contains(errMsg, "rate limit") || strings.Contains(errMsg, "quota") || strings.Contains(errMsg, "RESOURCE_EXHAUSTED") {
		return llm.ErrRateLimitExceeded("gemini")
	}

	if strings.Contains(errMsg, "not found") && strings.Contains(errMsg, "model") {
		return llm.ErrModelNotFound("unknown", "gemini")
	}

	// Generic error
	return fmt.Errorf("gemini API error: %w", err)
}
