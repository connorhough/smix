// internal/pr/provider_test.go
package pr

import (
	"context"
	"strings"
	"testing"

	"github.com/connorhough/smix/internal/config"
	"github.com/connorhough/smix/internal/llm"
)

// mockInteractiveProvider for testing
type mockInteractiveProvider struct {
	lastPrompt  string
	lastModel   string
	lastStreams *llm.IOStreams
	callCount   int
}

func (m *mockInteractiveProvider) Generate(ctx context.Context, prompt string, opts ...llm.Option) (string, error) {
	return "should not be called", nil
}

func (m *mockInteractiveProvider) RunInteractive(ctx context.Context, streams *llm.IOStreams, prompt string, opts ...llm.Option) error {
	m.lastPrompt = prompt
	m.lastStreams = streams
	m.callCount++

	options := llm.BuildOptions(opts)
	m.lastModel = options.Model

	return nil
}

func (m *mockInteractiveProvider) ValidateModel(model string) error {
	return nil
}

func (m *mockInteractiveProvider) DefaultModel() string {
	return "default-model"
}

func (m *mockInteractiveProvider) Name() string {
	return "mock"
}

func TestLaunchInteractiveSession_Success(t *testing.T) {
	mock := &mockInteractiveProvider{}
	var provider llm.Provider = mock

	// Verify it implements InteractiveProvider
	interactive, ok := provider.(llm.InteractiveProvider)
	if !ok {
		t.Fatal("mock should implement InteractiveProvider")
	}

	// Create test streams (simulates TTY)
	streams, _, _ := llm.TestIOStreams()

	// Simulate the call pattern used in LaunchClaudeCode
	ctx := context.Background()
	feedbackFile := "test_feedback.md"
	targetFile := "main.go"
	cfg := &config.ProviderConfig{Provider: "mock", Model: ""}

	err := LaunchClaudeCode(ctx, provider, streams, feedbackFile, targetFile, 1, 3, cfg)
	if err != nil {
		t.Errorf("LaunchClaudeCode() error = %v", err)
	}

	if mock.callCount != 1 {
		t.Errorf("expected 1 call, got %d", mock.callCount)
	}

	if mock.lastStreams != streams {
		t.Error("expected streams to be passed through")
	}

	// Verify prompt contains feedback file
	if !strings.Contains(mock.lastPrompt, feedbackFile) {
		t.Errorf("expected prompt to contain feedback file %q", feedbackFile)
	}

	// Verify prompt contains target file
	if !strings.Contains(mock.lastPrompt, targetFile) {
		t.Errorf("expected prompt to contain target file %q", targetFile)
	}

	// Verify batch info is included
	if !strings.Contains(mock.lastPrompt, "1 of 3") {
		t.Error("expected prompt to contain batch info '1 of 3'")
	}

	// Test with interactive interface directly
	_ = interactive
}

func TestLaunchInteractiveSession_NonInteractiveProvider(t *testing.T) {
	// We need a wrapper that implements llm.Provider but not InteractiveProvider
	wrapper := &basicProviderWrapper{name: "basic"}
	streams, _, _ := llm.TestIOStreams()
	ctx := context.Background()
	cfg := &config.ProviderConfig{Provider: "basic", Model: ""}

	err := LaunchClaudeCode(ctx, wrapper, streams, "test.md", "main.go", 1, 1, cfg)
	if err == nil {
		t.Error("expected error for non-interactive provider")
	}

	if !strings.Contains(err.Error(), "does not support interactive mode") {
		t.Errorf("expected error about interactive mode, got: %v", err)
	}
}

// basicProviderWrapper implements only Provider, not InteractiveProvider
type basicProviderWrapper struct {
	name string
}

func (b *basicProviderWrapper) Generate(ctx context.Context, prompt string, opts ...llm.Option) (string, error) {
	return "generated response", nil
}

func (b *basicProviderWrapper) ValidateModel(model string) error {
	return nil
}

func (b *basicProviderWrapper) DefaultModel() string {
	return "default-model"
}

func (b *basicProviderWrapper) Name() string {
	return b.name
}

func TestLaunchInteractiveSession_NonInteractiveStreams(t *testing.T) {
	mock := &mockInteractiveProvider{}
	streams, _, _ := llm.TestIOStreamsNonInteractive()
	ctx := context.Background()
	cfg := &config.ProviderConfig{Provider: "mock", Model: ""}

	err := LaunchClaudeCode(ctx, mock, streams, "test.md", "main.go", 1, 1, cfg)
	if err == nil {
		t.Error("expected error for non-interactive streams")
	}

	if !strings.Contains(err.Error(), "requires a terminal") {
		t.Errorf("expected error about terminal requirement, got: %v", err)
	}

	// Should not have called the provider
	if mock.callCount != 0 {
		t.Errorf("expected 0 calls to provider, got %d", mock.callCount)
	}
}
