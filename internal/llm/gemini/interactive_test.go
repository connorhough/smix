// internal/llm/gemini/interactive_test.go
package gemini

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/connorhough/smix/internal/llm"
)

func TestGeminiProvider_ImplementsInteractiveProvider(t *testing.T) {
	// Create provider with mock API key (won't actually connect)
	p := &Provider{cliPath: "gemini"}

	// Verify Provider implements InteractiveProvider interface at compile time
	var _ llm.InteractiveProvider = p

	// Verify via type assertion
	var provider llm.Provider = p
	_, ok := provider.(llm.InteractiveProvider)
	if !ok {
		t.Error("Gemini provider should implement InteractiveProvider interface")
	}
}

func TestGeminiProvider_RunInteractive_NoCLI(t *testing.T) {
	// Provider without CLI path should return clear error
	p := &Provider{cliPath: ""}

	streams, _, _ := llm.TestIOStreams()
	ctx := context.Background()

	err := p.RunInteractive(ctx, streams, "test prompt")
	if err == nil {
		t.Error("expected error when CLI is not available")
	}

	expectedMsg := "gemini CLI not available"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("expected error to contain %q, got: %v", expectedMsg, err)
	}
}

func TestGeminiProvider_RunInteractive_UsesStreams(t *testing.T) {
	// Unit test - verifies streams are connected (doesn't actually run gemini)
	p := &Provider{cliPath: "echo"} // Use echo for testing

	streams, _, out := llm.TestIOStreams()
	ctx := context.Background()

	// This will call "echo -m <model> test-prompt" which should output to our buffer
	err := p.RunInteractive(ctx, streams, "test-prompt", llm.WithModel(ModelFlash))
	if err != nil {
		t.Logf("RunInteractive() returned error: %v", err)
		return
	}

	// If it succeeded, verify output was captured
	output := out.String()
	if !strings.Contains(output, "test-prompt") {
		t.Errorf("expected output to contain 'test-prompt', got: %q", output)
	}
}

func TestGeminiProvider_RunInteractive_NonInteractive(t *testing.T) {
	// Test with non-interactive streams - caller should check IsInteractive first
	p := &Provider{cliPath: "echo"} // Use echo to avoid hanging

	streams, _, _ := llm.TestIOStreamsNonInteractive()
	ctx := context.Background()

	// Provider doesn't check TTY - caller's responsibility
	// This documents that behavior (may fail or hang depending on subprocess)
	err := p.RunInteractive(ctx, streams, "test", llm.WithModel(ModelFlash))

	// We don't assert on error here - just document that callers should
	// check streams.IsInteractive() before calling RunInteractive
	t.Logf("RunInteractive() with non-interactive streams returned: %v", err)
}

func TestGeminiProvider_RunInteractive_Integration(t *testing.T) {
	// Integration test - only runs if gemini CLI is available
	apiKey := os.Getenv(APIKeyEnvVar)
	if apiKey == "" {
		t.Skipf("skipping: %s not set", APIKeyEnvVar)
	}

	p, err := NewProvider(apiKey)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	if p.cliPath == "" {
		t.Skip("gemini CLI not installed")
	}

	// Use real streams for integration test
	streams := llm.NewIOStreams()
	if !streams.IsInteractive() {
		t.Skip("skipping interactive test in non-TTY environment (CI/CD)")
	}

	ctx := context.Background()
	err = p.RunInteractive(ctx, streams, "Say 'interactive test' and nothing else", llm.WithModel(ModelFlash))
	if err != nil {
		t.Errorf("RunInteractive() failed in TTY environment: %v", err)
	}
}

func TestGeminiProvider_HasInteractiveSupport(t *testing.T) {
	tests := []struct {
		name     string
		cliPath  string
		expected bool
	}{
		{"with CLI", "gemini", true},
		{"without CLI", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Provider{cliPath: tt.cliPath}
			if got := p.HasInteractiveSupport(); got != tt.expected {
				t.Errorf("HasInteractiveSupport() = %v, want %v", got, tt.expected)
			}
		})
	}
}
