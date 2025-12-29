package claude

import (
	"context"
	"errors"
	"testing"

	"github.com/connorhough/smix/internal/llm"
)

func TestClaudeProvider_Name(t *testing.T) {
	p := &Provider{}
	if got := p.Name(); got != "claude" {
		t.Errorf("Name() = %q, want %q", got, "claude")
	}
}

func TestClaudeProvider_DefaultModel(t *testing.T) {
	p := &Provider{}
	if got := p.DefaultModel(); got != ModelHaiku {
		t.Errorf("DefaultModel() = %q, want %q", got, ModelHaiku)
	}
}

func TestClaudeProvider_ValidateModel(t *testing.T) {
	p := &Provider{}

	// Claude provider doesn't validate models - lets CLI fail
	// Just ensure method exists and returns nil
	if err := p.ValidateModel("any-model"); err != nil {
		t.Errorf("ValidateModel() should return nil, got %v", err)
	}
}

func TestNewProvider(t *testing.T) {
	// This test validates the constructor checks for CLI availability
	// We can't reliably test the error case without mocking exec.LookPath
	// So we just verify the constructor exists and returns a provider
	p, err := NewProvider()
	if err != nil {
		// If CLI not found, ensure we get the right error type
		var providerErr *llm.ProviderError
		if !errors.As(err, &providerErr) {
			t.Errorf("expected ProviderError, got: %v", err)
		}
		t.Skipf("claude CLI not available: %v", err)
	}

	if p == nil {
		t.Error("expected non-nil provider")
		t.FailNow()
	}

	if p.cliPath == "" {
		t.Error("expected non-empty cliPath")
	}
}

func TestClaudeProvider_Generate_Integration(t *testing.T) {
	// Integration test - only runs if claude CLI is available
	p, err := NewProvider()
	if err != nil {
		t.Skipf("claude CLI not available: %v", err)
	}

	ctx := context.Background()
	result, err := p.Generate(ctx, "Say 'test' and nothing else", llm.WithModel(ModelHaiku))
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if result == "" {
		t.Error("expected non-empty result")
	}

	t.Logf("Claude response: %s", result)
}
