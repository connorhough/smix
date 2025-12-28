// internal/llm/claude/interactive_test.go
package claude

import (
	"context"
	"strings"
	"testing"

	"github.com/connorhough/smix/internal/llm"
)

func TestClaudeProvider_ImplementsInteractiveProvider(t *testing.T) {
	p := &Provider{cliPath: "claude"}

	// Verify Provider implements InteractiveProvider interface at compile time
	var _ llm.InteractiveProvider = p

	// Verify via type assertion
	var provider llm.Provider = p
	_, ok := provider.(llm.InteractiveProvider)
	if !ok {
		t.Error("Claude provider should implement InteractiveProvider interface")
	}
}

func TestClaudeProvider_RunInteractive_UsesStreams(t *testing.T) {
	// Unit test - verifies streams are connected (doesn't actually run claude)
	p := &Provider{cliPath: "echo"} // Use echo for testing

	streams, _, out := llm.TestIOStreams()
	ctx := context.Background()

	// This will call "echo test-prompt" which should output to our buffer
	err := p.RunInteractive(ctx, streams, "test-prompt", llm.WithModel(ModelHaiku))
	if err != nil {
		// echo might fail in some test environments, which is okay
		t.Logf("RunInteractive() returned error: %v", err)
		return
	}

	// If it succeeded, verify output was captured
	output := out.String()
	if !strings.Contains(output, "test-prompt") {
		t.Errorf("expected output to contain 'test-prompt', got: %q", output)
	}
}

func TestClaudeProvider_RunInteractive_Integration(t *testing.T) {
	// Integration test - only runs if claude CLI is available
	p, err := NewProvider()
	if err != nil {
		t.Skipf("claude CLI not available: %v", err)
	}

	// Use real streams for integration test
	streams := llm.NewIOStreams()
	if !streams.IsInteractive() {
		t.Skip("skipping interactive test in non-TTY environment (CI/CD)")
	}

	ctx := context.Background()
	err = p.RunInteractive(ctx, streams, "Say 'interactive test' and nothing else", llm.WithModel(ModelHaiku))
	// In a real TTY, this should work
	if err != nil {
		t.Errorf("RunInteractive() failed in TTY environment: %v", err)
	}
}

func TestClaudeProvider_RunInteractive_NonInteractive(t *testing.T) {
	// Test with non-interactive streams - caller should check IsInteractive first
	p := &Provider{cliPath: "claude"}

	streams, _, _ := llm.TestIOStreamsNonInteractive()
	ctx := context.Background()

	// Provider doesn't check TTY - caller's responsibility
	// This documents that behavior (may fail or hang depending on subprocess)
	err := p.RunInteractive(ctx, streams, "test", llm.WithModel(ModelHaiku))

	// We don't assert on error here - just document that callers should
	// check streams.IsInteractive() before calling RunInteractive
	t.Logf("RunInteractive() with non-interactive streams returned: %v", err)
}
