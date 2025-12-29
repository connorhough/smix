package gemini

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/connorhough/smix/internal/llm"
)

func TestGeminiProvider_Name(t *testing.T) {
	ctx := context.Background()
	apiKey := "test-key"
	p, _ := NewProvider(ctx, apiKey)
	if got := p.Name(); got != "gemini" {
		t.Errorf("Name() = %q, want %q", got, "gemini")
	}
}

func TestGeminiProvider_DefaultModel(t *testing.T) {
	ctx := context.Background()
	apiKey := "test-key"
	p, _ := NewProvider(ctx, apiKey)
	if got := p.DefaultModel(); got != ModelFlash {
		t.Errorf("DefaultModel() = %q, want %q", got, ModelFlash)
	}
}

func TestGeminiProvider_ValidateModel(t *testing.T) {
	ctx := context.Background()
	apiKey := "test-key"
	p, _ := NewProvider(ctx, apiKey)

	// Gemini provider doesn't pre-validate - lets API fail naturally
	if err := p.ValidateModel("any-model"); err != nil {
		t.Errorf("ValidateModel() should return nil, got %v", err)
	}
}

func TestNewProvider_MissingAPIKey(t *testing.T) {
	ctx := context.Background()
	p, err := NewProvider(ctx, "")

	// Provider creation succeeds if CLI is available
	if err == nil {
		if p.cliPath == "" {
			t.Error("provider created without API key or CLI - should have failed")
		}
		// Valid: CLI-only mode
		return
	}

	// If error, should be authentication error (neither API key nor CLI)
	if !strings.Contains(err.Error(), "API key is required") {
		t.Errorf("unexpected error: %v", err)
	}
}

// Integration test - only runs with API key set
func TestGeminiProvider_Generate_Integration(t *testing.T) {
	apiKey := os.Getenv(APIKeyEnvVar)
	if apiKey == "" {
		t.Skip("API key not set")
	}

	ctx := context.Background()
	p, err := NewProvider(ctx, apiKey)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}

	result, err := p.Generate(ctx, "Say 'hello' and nothing else")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty result")
	}

	t.Logf("Generated response: %s", result)
}

func TestGeminiProvider_Generate_ViaCLI(t *testing.T) {
	// Create provider with only CLI (no API client)
	p := &Provider{
		client:  nil,
		cliPath: "echo", // Use echo to simulate CLI output
	}

	ctx := context.Background()
	result, err := p.Generate(ctx, "test-prompt", llm.WithModel("test-model"))
	if err != nil {
		t.Fatalf("Generate via CLI failed: %v", err)
	}

	// echo will output: --model test-model test-prompt
	if !strings.Contains(result, "test-prompt") {
		t.Errorf("expected output to contain 'test-prompt', got: %q", result)
	}
}
