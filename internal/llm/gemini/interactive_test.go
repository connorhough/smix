// internal/llm/gemini/interactive_test.go
package gemini

import (
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
