// internal/llm/provider_test.go
package llm

import (
	"context"
	"testing"
)

// mockInteractiveProvider implements both Provider and InteractiveProvider
type mockInteractiveProvider struct {
	interactiveCalled bool
	generateCalled    bool
	lastStreams       *IOStreams
	lastPrompt        string
}

func (m *mockInteractiveProvider) Generate(ctx context.Context, prompt string, opts ...Option) (string, error) {
	m.generateCalled = true
	return "generated response", nil
}

func (m *mockInteractiveProvider) ValidateModel(model string) error {
	return nil
}

func (m *mockInteractiveProvider) DefaultModel() string {
	return "test-model"
}

func (m *mockInteractiveProvider) Name() string {
	return "mock"
}

func (m *mockInteractiveProvider) RunInteractive(ctx context.Context, streams *IOStreams, prompt string, opts ...Option) error {
	m.interactiveCalled = true
	m.lastStreams = streams
	m.lastPrompt = prompt
	return nil
}

func TestInteractiveProvider_TypeAssertion(t *testing.T) {
	// Test that we can detect InteractiveProvider capability
	var provider Provider = &mockInteractiveProvider{}

	interactive, ok := provider.(InteractiveProvider)
	if !ok {
		t.Error("expected provider to implement InteractiveProvider")
	}

	ctx := context.Background()
	streams, _, _ := TestIOStreams()

	if err := interactive.RunInteractive(ctx, streams, "test prompt"); err != nil {
		t.Errorf("RunInteractive() error = %v", err)
	}

	mock := provider.(*mockInteractiveProvider)
	if !mock.interactiveCalled {
		t.Error("expected RunInteractive to be called")
	}

	if mock.lastStreams != streams {
		t.Error("expected streams to be passed through")
	}

	if mock.lastPrompt != "test prompt" {
		t.Errorf("expected prompt 'test prompt', got %q", mock.lastPrompt)
	}
}

func TestProvider_WithoutInteractiveCapability(t *testing.T) {
	// Test that providers without InteractiveProvider gracefully fall back
	// Create a struct that embeds Provider interface but NOT InteractiveProvider
	provider := struct {
		Provider
	}{
		// Anonymous struct that satisfies Provider but NOT InteractiveProvider
	}

	// This should fail the type assertion since the struct
	// doesn't have RunInteractive method
	var i interface{} = provider
	_, ok := i.(InteractiveProvider)
	if ok {
		t.Error("non-interactive provider should not implement InteractiveProvider")
	}
}
