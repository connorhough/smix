package gemini

import (
	"context"
	"os"
	"testing"
)

func TestGeminiProvider_Name(t *testing.T) {
	apiKey := "test-key"
	p, _ := NewProvider(apiKey)
	if got := p.Name(); got != "gemini" {
		t.Errorf("Name() = %q, want %q", got, "gemini")
	}
}

func TestGeminiProvider_DefaultModel(t *testing.T) {
	apiKey := "test-key"
	p, _ := NewProvider(apiKey)
	if got := p.DefaultModel(); got != ModelFlash {
		t.Errorf("DefaultModel() = %q, want %q", got, ModelFlash)
	}
}

func TestGeminiProvider_ValidateModel(t *testing.T) {
	apiKey := "test-key"
	p, _ := NewProvider(apiKey)

	// Gemini provider doesn't pre-validate - lets API fail naturally
	if err := p.ValidateModel("any-model"); err != nil {
		t.Errorf("ValidateModel() should return nil, got %v", err)
	}
}

func TestNewProvider_MissingAPIKey(t *testing.T) {
	_, err := NewProvider("")
	if err == nil {
		t.Error("expected error when API key is empty")
	}
}

// Integration test - only runs with GEMINI_API_KEY set
func TestGeminiProvider_Generate_Integration(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	p, err := NewProvider(apiKey)
	if err != nil {
		t.Fatalf("NewProvider failed: %v", err)
	}
	defer p.Close()

	ctx := context.Background()
	result, err := p.Generate(ctx, "Say 'hello' and nothing else")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result == "" {
		t.Error("expected non-empty result")
	}

	t.Logf("Generated response: %s", result)
}
