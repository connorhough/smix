# Multi-Provider LLM Support Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor smix CLI to support multiple LLM providers (Claude via CLI, Gemini via SDK) with unified Provider interface and flexible per-command configuration.

**Architecture:** Provider interface abstraction with context support, global singleton factory with caching, GitHub CLI-style configuration resolution (flags > command config > global config), typed error handling with retry logic. PR command remains Claude-only interactive, ask/do support both providers.

**Tech Stack:** Go 1.21+, Cobra, Viper, google.golang.org/genai SDK, os/exec for Claude CLI wrapping

---

## Phase 0: Technical Validation & Proof of Concept

### Task 0.1: Prototype Claude CLI Wrapper

**Goal:** Validate that wrapping `claude` CLI works for non-interactive request-response.

**Files:**
- Create: `internal/llm/claude/prototype_test.go`

**Step 1: Write prototype test**

Create test file to validate CLI wrapping approach:

```go
package claude

import (
	"context"
	"os/exec"
	"testing"
)

func TestPrototypeCLIWrapper(t *testing.T) {
	// Skip if claude CLI not available
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude CLI not available")
	}

	ctx := context.Background()
	prompt := "Say 'hello' and nothing else"

	cmd := exec.CommandContext(ctx, "claude", "--model", "haiku", "-p", prompt)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("CLI execution failed: %v (output: %s)", err, output)
	}

	if len(output) == 0 {
		t.Fatal("Expected non-empty output")
	}

	t.Logf("CLI wrapper works! Output: %s", output)
}
```

**Step 2: Run prototype test**

Run: `go test -v ./internal/llm/claude -run TestPrototypeCLIWrapper`
Expected: PASS (if claude CLI installed) or SKIP

**Step 3: Document findings**

Create notes file with observations:
- Does CLI respect context cancellation?
- What error codes does it return?
- How does it handle invalid model names?
- Output format observations

**Step 4: Commit prototype**

```bash
git add internal/llm/claude/prototype_test.go
git commit -m "chore: add claude CLI wrapper prototype test"
```

---

### Task 0.2: Prototype Gemini SDK Integration

**Goal:** Validate Gemini SDK integration and understand error handling.

**Files:**
- Create: `internal/llm/gemini/prototype_test.go`

**Step 1: Add Gemini SDK dependency**

Run: `go get google.golang.org/genai`

**Step 2: Write prototype test**

```go
package gemini

import (
	"context"
	"os"
	"testing"

	"google.golang.org/genai"
)

func TestPrototypeGeminiSDK(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash")
	prompt := genai.Text("Say 'hello' and nothing else")

	resp, err := model.GenerateContent(ctx, prompt)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(resp.Candidates) == 0 {
		t.Fatal("Expected at least one candidate")
	}

	t.Logf("Gemini SDK works! Response: %v", resp.Candidates[0].Content)
}
```

**Step 3: Run prototype test**

Run: `GEMINI_API_KEY=your_key go test -v ./internal/llm/gemini -run TestPrototypeGeminiSDK`
Expected: PASS (with valid API key) or SKIP

**Step 4: Document findings**

Add notes about:
- Error format for invalid API key
- Error format for invalid model name
- Response structure
- Rate limiting behavior

**Step 5: Commit prototype**

```bash
git add internal/llm/gemini/prototype_test.go go.mod go.sum
git commit -m "chore: add gemini SDK prototype test and dependency"
```

---

### Task 0.3: Design Provider Interface

**Goal:** Finalize provider interface based on prototype learnings.

**Files:**
- Create: `docs/architecture/provider-interface.md`

**Step 1: Document interface design**

```markdown
# Provider Interface Design

## Interface Definition

```go
package llm

import "context"

// Provider defines the interface for LLM providers
type Provider interface {
	// Generate sends a prompt and returns the response
	Generate(ctx context.Context, prompt string, opts ...Option) (string, error)

	// ValidateModel checks if a model name is valid for this provider
	// Returns helpful error message if invalid
	ValidateModel(model string) error

	// DefaultModel returns the default model for this provider
	DefaultModel() string

	// Name returns the provider name (e.g., "claude", "gemini")
	Name() string
}

// Option configures provider behavior
type Option func(*GenerateOptions)

type GenerateOptions struct {
	Model string
}
```

## Design Decisions

1. **Context Support:** All operations take context.Context for cancellation/timeout
2. **Model Validation:** Let API/CLI fail naturally, wrap errors with helpful context
3. **Options Pattern:** Functional options for flexibility (model override, future params)
4. **No Streaming:** Simple request-response sufficient for ask/do commands

## Error Handling

Providers should wrap errors with typed errors defined in `errors.go`:
- `ErrProviderNotAvailable`: CLI not found or SDK client creation failed
- `ErrAuthenticationFailed`: API key invalid/missing
- `ErrRateLimitExceeded`: Provider rate limit hit
- `ErrModelNotFound`: Model name invalid (wrapped from API/CLI error)

## Retry Logic

Network errors trigger exponential backoff retry:
- Max retries: 3
- Initial delay: 1s
- Max delay: 30s
- Backoff factor: 2
```

**Step 2: Review and validate design**

Mentally walk through each command (ask, do) using this interface.
Ensure it handles all error cases from prototypes.

**Step 3: Commit design doc**

```bash
git add docs/architecture/provider-interface.md
git commit -m "docs: add provider interface design"
```

---

## Phase 1: Foundation

### Task 1.1: Define Error Types

**Files:**
- Create: `internal/llm/errors.go`
- Create: `internal/llm/errors_test.go`

**Step 1: Write failing test for error types**

```go
package llm

import (
	"errors"
	"testing"
)

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{
			name:    "provider not available",
			err:     ErrProviderNotAvailable("claude", errors.New("not found")),
			wantMsg: "provider 'claude' not available: not found",
		},
		{
			name:    "authentication failed",
			err:     ErrAuthenticationFailed("gemini", errors.New("invalid key")),
			wantMsg: "authentication failed for provider 'gemini': invalid key",
		},
		{
			name:    "rate limit exceeded",
			err:     ErrRateLimitExceeded("gemini"),
			wantMsg: "rate limit exceeded for provider 'gemini'",
		},
		{
			name:    "model not found",
			err:     ErrModelNotFound("invalid-model", "gemini"),
			wantMsg: "model 'invalid-model' not found for provider 'gemini'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.wantMsg {
				t.Errorf("got %q, want %q", tt.err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestErrorUnwrapping(t *testing.T) {
	underlying := errors.New("network error")
	err := ErrProviderNotAvailable("test", underlying)

	if !errors.Is(err, underlying) {
		t.Error("error should unwrap to underlying error")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/llm -run TestError`
Expected: FAIL with "undefined: ErrProviderNotAvailable"

**Step 3: Implement error types**

```go
package llm

import "fmt"

// ProviderError represents a provider-specific error
type ProviderError struct {
	Provider string
	Msg      string
	Err      error
}

func (e *ProviderError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Msg, e.Err)
	}
	return e.Msg
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// ErrProviderNotAvailable indicates the provider is not available (CLI not found, SDK init failed)
func ErrProviderNotAvailable(provider string, err error) error {
	return &ProviderError{
		Provider: provider,
		Msg:      fmt.Sprintf("provider '%s' not available", provider),
		Err:      err,
	}
}

// ErrAuthenticationFailed indicates authentication failure (invalid API key, etc.)
func ErrAuthenticationFailed(provider string, err error) error {
	return &ProviderError{
		Provider: provider,
		Msg:      fmt.Sprintf("authentication failed for provider '%s'", provider),
		Err:      err,
	}
}

// ErrRateLimitExceeded indicates the provider's rate limit was hit
func ErrRateLimitExceeded(provider string) error {
	return &ProviderError{
		Provider: provider,
		Msg:      fmt.Sprintf("rate limit exceeded for provider '%s'", provider),
	}
}

// ErrModelNotFound indicates the specified model doesn't exist for the provider
func ErrModelNotFound(model, provider string) error {
	return &ProviderError{
		Provider: provider,
		Msg:      fmt.Sprintf("model '%s' not found for provider '%s'", model, provider),
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/llm -run TestError -v`
Expected: PASS

**Step 5: Commit error types**

```bash
git add internal/llm/errors.go internal/llm/errors_test.go
git commit -m "feat(llm): add typed error definitions"
```

---

### Task 1.2: Define Provider Interface and Types

**Files:**
- Create: `internal/llm/provider.go`
- Create: `internal/llm/options.go`

**Step 1: Create provider interface**

```go
package llm

import "context"

// Provider defines the interface for LLM providers
type Provider interface {
	// Generate sends a prompt and returns the response
	Generate(ctx context.Context, prompt string, opts ...Option) (string, error)

	// ValidateModel checks if a model name is valid for this provider
	// Returns error with helpful message if invalid
	ValidateModel(model string) error

	// DefaultModel returns the default model for this provider
	DefaultModel() string

	// Name returns the provider name (e.g., "claude", "gemini")
	Name() string
}
```

**Step 2: Create options types**

```go
package llm

// Option configures provider behavior
type Option func(*GenerateOptions)

// GenerateOptions holds configuration for Generate calls
type GenerateOptions struct {
	Model string
}

// WithModel overrides the model for this generation
func WithModel(model string) Option {
	return func(opts *GenerateOptions) {
		opts.Model = model
	}
}

// buildOptions constructs GenerateOptions from Option functions
func buildOptions(opts []Option) *GenerateOptions {
	options := &GenerateOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}
```

**Step 3: Commit interface and types**

```bash
git add internal/llm/provider.go internal/llm/options.go
git commit -m "feat(llm): define provider interface and options"
```

---

### Task 1.3: Implement Retry Logic with Exponential Backoff

**Files:**
- Create: `internal/llm/retry.go`
- Create: `internal/llm/retry_test.go`

**Step 1: Write failing test for retry logic**

```go
package llm

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryWithBackoff(t *testing.T) {
	t.Run("succeeds on first try", func(t *testing.T) {
		callCount := 0
		fn := func(ctx context.Context) (string, error) {
			callCount++
			return "success", nil
		}

		result, err := retryWithBackoff(context.Background(), fn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "success" {
			t.Errorf("got %q, want %q", result, "success")
		}
		if callCount != 1 {
			t.Errorf("got %d calls, want 1", callCount)
		}
	})

	t.Run("retries on transient error", func(t *testing.T) {
		callCount := 0
		fn := func(ctx context.Context) (string, error) {
			callCount++
			if callCount < 3 {
				return "", errors.New("network error")
			}
			return "success", nil
		}

		result, err := retryWithBackoff(context.Background(), fn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "success" {
			t.Errorf("got %q, want %q", result, "success")
		}
		if callCount != 3 {
			t.Errorf("got %d calls, want 3", callCount)
		}
	})

	t.Run("fails after max retries", func(t *testing.T) {
		callCount := 0
		testErr := errors.New("persistent error")
		fn := func(ctx context.Context) (string, error) {
			callCount++
			return "", testErr
		}

		_, err := retryWithBackoff(context.Background(), fn)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, testErr) {
			t.Errorf("expected error to wrap testErr")
		}
		if callCount != 3 {
			t.Errorf("got %d calls, want 3 (max retries)", callCount)
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		fn := func(ctx context.Context) (string, error) {
			return "", errors.New("should not retry")
		}

		_, err := retryWithBackoff(ctx, fn)
		if err == nil {
			t.Fatal("expected context error")
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/llm -run TestRetryWithBackoff`
Expected: FAIL with "undefined: retryWithBackoff"

**Step 3: Implement retry logic**

```go
package llm

import (
	"context"
	"fmt"
	"time"
)

const (
	maxRetries   = 3
	initialDelay = 1 * time.Second
	maxDelay     = 30 * time.Second
	backoffRate  = 2.0
)

// retryWithBackoff executes fn with exponential backoff retry logic
func retryWithBackoff(ctx context.Context, fn func(context.Context) (string, error)) (string, error) {
	var lastErr error
	delay := initialDelay

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check context before attempting
		if err := ctx.Err(); err != nil {
			return "", err
		}

		result, err := fn(ctx)
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't sleep after last attempt
		if attempt < maxRetries-1 {
			select {
			case <-time.After(delay):
				// Calculate next delay with exponential backoff
				delay = time.Duration(float64(delay) * backoffRate)
				if delay > maxDelay {
					delay = maxDelay
				}
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}

	return "", fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/llm -run TestRetryWithBackoff -v`
Expected: PASS

**Step 5: Commit retry logic**

```bash
git add internal/llm/retry.go internal/llm/retry_test.go
git commit -m "feat(llm): add retry logic with exponential backoff"
```

---

### Task 1.4: Update Config Package for Nested Configuration

**Files:**
- Modify: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Step 1: Write failing test for config resolution**

```go
package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestResolveProviderConfig(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
provider: claude
model: sonnet

commands:
  ask:
    provider: gemini
    model: gemini-1.5-flash
  do:
    provider: gemini
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Initialize viper with test config
	viper.Reset()
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	tests := []struct {
		name         string
		command      string
		wantProvider string
		wantModel    string
	}{
		{
			name:         "ask command uses command-specific config",
			command:      "ask",
			wantProvider: "gemini",
			wantModel:    "gemini-1.5-flash",
		},
		{
			name:         "do command uses command-specific provider, inherits global model",
			command:      "do",
			wantProvider: "gemini",
			wantModel:    "sonnet", // Falls back to global
		},
		{
			name:         "pr command uses global defaults",
			command:      "pr",
			wantProvider: "claude",
			wantModel:    "sonnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ResolveProviderConfig(tt.command)
			if cfg.Provider != tt.wantProvider {
				t.Errorf("provider: got %q, want %q", cfg.Provider, tt.wantProvider)
			}
			if cfg.Model != tt.wantModel {
				t.Errorf("model: got %q, want %q", cfg.Model, tt.wantModel)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/config -run TestResolveProviderConfig`
Expected: FAIL with "undefined: ResolveProviderConfig"

**Step 3: Implement config resolution**

```go
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// ProviderConfig holds provider and model configuration
type ProviderConfig struct {
	Provider string
	Model    string
}

// ResolveProviderConfig resolves provider configuration for a command
// Precedence: command-specific config -> global config
// Flags are handled separately in command layer
func ResolveProviderConfig(commandName string) *ProviderConfig {
	cfg := &ProviderConfig{}

	// Try command-specific provider
	commandProviderKey := fmt.Sprintf("commands.%s.provider", commandName)
	if viper.IsSet(commandProviderKey) {
		cfg.Provider = viper.GetString(commandProviderKey)
	} else {
		// Fall back to global provider
		cfg.Provider = viper.GetString("provider")
	}

	// Try command-specific model
	commandModelKey := fmt.Sprintf("commands.%s.model", commandName)
	if viper.IsSet(commandModelKey) {
		cfg.Model = viper.GetString(commandModelKey)
	} else {
		// Fall back to global model
		cfg.Model = viper.GetString("model")
	}

	return cfg
}

// ApplyFlags applies flag overrides to config (called from command layer)
func (c *ProviderConfig) ApplyFlags(providerFlag, modelFlag string) {
	if providerFlag != "" {
		c.Provider = providerFlag
	}
	if modelFlag != "" {
		c.Model = modelFlag
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/config -run TestResolveProviderConfig -v`
Expected: PASS

**Step 5: Commit config updates**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat(config): add provider config resolution with precedence"
```

---

### Task 1.5: Add Config File Auto-Creation

**Files:**
- Create: `internal/config/init.go`
- Create: `internal/config/init_test.go`
- Create: `internal/config/template.go`

**Step 1: Write config template**

```go
package config

const configTemplate = `# smix configuration file
# Provider settings control which LLM provider to use

# Global default provider (claude or gemini)
provider: claude

# Global default model (optional, uses provider default if omitted)
# model: sonnet

# Provider-specific settings
providers:
  claude:
    # Path to claude CLI if not in PATH (optional)
    # cli_path: /usr/local/bin/claude
  gemini:
    # API key (prefer GEMINI_API_KEY environment variable)
    # api_key: ${GEMINI_API_KEY}

# Per-command overrides (optional)
# Uncomment and customize as needed
#commands:
#  ask:
#    provider: gemini
#    model: gemini-1.5-flash
#  do:
#    provider: gemini
#    model: gemini-1.5-flash
#  pr:
#    provider: claude
#    model: sonnet

# Observability settings
log_level: info  # debug, info, warn, error
`
```

**Step 2: Write failing test for config initialization**

```go
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureConfigExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Ensure config doesn't exist
	if _, err := os.Stat(configPath); err == nil {
		t.Fatal("config should not exist yet")
	}

	// Call EnsureConfigExists
	if err := EnsureConfigExists(configPath); err != nil {
		t.Fatalf("EnsureConfigExists failed: %v", err)
	}

	// Verify config was created
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read created config: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("config file is empty")
	}

	// Verify it contains expected content
	contentStr := string(content)
	expectedStrings := []string{
		"provider: claude",
		"providers:",
		"commands:",
	}

	for _, expected := range expectedStrings {
		if !contains(contentStr, expected) {
			t.Errorf("config missing expected content: %q", expected)
		}
	}
}

func TestEnsureConfigExistsIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create initial config
	if err := EnsureConfigExists(configPath); err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Call again - should not error
	if err := EnsureConfigExists(configPath); err != nil {
		t.Fatalf("second call should not error: %v", err)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || contains(s[1:], substr)))
}
```

**Step 3: Run test to verify it fails**

Run: `go test ./internal/config -run TestEnsureConfig`
Expected: FAIL with "undefined: EnsureConfigExists"

**Step 4: Implement config initialization**

```go
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureConfigExists creates a config file with template if it doesn't exist
func EnsureConfigExists(configPath string) error {
	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // Config exists, nothing to do
	}

	// Create config directory if needed
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write template config
	if err := os.WriteFile(configPath, []byte(configTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write config template: %w", err)
	}

	return nil
}
```

**Step 5: Run test to verify it passes**

Run: `go test ./internal/config -run TestEnsureConfig -v`
Expected: PASS

**Step 6: Commit config initialization**

```bash
git add internal/config/init.go internal/config/init_test.go internal/config/template.go
git commit -m "feat(config): add auto-creation of config file with template"
```

---

## Phase 2: Provider Implementations

### Task 2.1: Define Model Constants for Claude

**Files:**
- Create: `internal/llm/claude/models.go`

**Step 1: Create model constants**

```go
package claude

// Model name constants for Claude CLI
// Claude CLI accepts short model names
const (
	ModelHaiku  = "haiku"
	ModelSonnet = "sonnet"
	ModelOpus   = "opus"
)

// DefaultModel returns the default Claude model
func DefaultModel() string {
	return ModelSonnet
}
```

**Step 2: Commit model constants**

```bash
git add internal/llm/claude/models.go
git commit -m "feat(llm/claude): add model name constants"
```

---

### Task 2.2: Implement Claude Provider

**Files:**
- Create: `internal/llm/claude/provider.go`
- Create: `internal/llm/claude/provider_test.go`

**Step 1: Write failing test for Claude provider**

```go
package claude

import (
	"context"
	"os/exec"
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
	if got := p.DefaultModel(); got != ModelSonnet {
		t.Errorf("DefaultModel() = %q, want %q", got, ModelSonnet)
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

func TestNewProvider_CLINotFound(t *testing.T) {
	// Save original PATH and restore after test
	originalPath := exec.LookPath
	defer func() { exec.LookPath = originalPath }()

	// This test validates the constructor checks for CLI availability
	// We'll implement the actual check in the next step
	_, err := NewProvider()
	if err == nil {
		t.Error("expected error when claude CLI not found")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/llm/claude -run TestClaudeProvider`
Expected: FAIL with "undefined: Provider"

**Step 3: Implement Claude provider**

```go
package claude

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/connorhough/smix/internal/llm"
)

// Provider implements the llm.Provider interface for Claude CLI
type Provider struct {
	cliPath string
}

// NewProvider creates a new Claude provider
func NewProvider() (*Provider, error) {
	// Check if claude CLI is available
	cliPath, err := exec.LookPath("claude")
	if err != nil {
		return nil, llm.ErrProviderNotAvailable("claude", err)
	}

	return &Provider{
		cliPath: cliPath,
	}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return "claude"
}

// DefaultModel returns the default model for Claude
func (p *Provider) DefaultModel() string {
	return DefaultModel()
}

// ValidateModel checks if a model is valid
// Claude CLI will fail with helpful error if model is invalid,
// so we let it fail naturally and wrap the error
func (p *Provider) ValidateModel(model string) error {
	return nil // No pre-validation, let CLI handle it
}

// Generate sends a prompt to Claude and returns the response
func (p *Provider) Generate(ctx context.Context, prompt string, opts ...llm.Option) (string, error) {
	options := llm.BuildOptions(opts)

	// Use provided model or default
	model := options.Model
	if model == "" {
		model = p.DefaultModel()
	}

	// Build command
	cmd := exec.CommandContext(ctx, p.cliPath, "--model", model, "-p", prompt)

	// Execute with retry logic
	return llm.RetryWithBackoff(ctx, func(ctx context.Context) (string, error) {
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Wrap CLI errors with context
			return "", fmt.Errorf("claude CLI failed: %w (output: %s)", err, output)
		}

		result := strings.TrimSpace(string(output))
		if result == "" {
			return "", fmt.Errorf("claude CLI returned empty response")
		}

		return result, nil
	})
}
```

**Step 4: Update options to export BuildOptions helper**

Modify `internal/llm/options.go`:

```go
// BuildOptions constructs GenerateOptions from Option functions
// Exported for use by provider implementations
func BuildOptions(opts []Option) *GenerateOptions {
	options := &GenerateOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}
```

**Step 5: Export RetryWithBackoff for provider use**

Modify `internal/llm/retry.go` - change function name from lowercase to uppercase:

```go
// RetryWithBackoff executes fn with exponential backoff retry logic
// Exported for use by provider implementations
func RetryWithBackoff(ctx context.Context, fn func(context.Context) (string, error)) (string, error) {
	// ... existing implementation
}
```

**Step 6: Update retry tests**

Modify `internal/llm/retry_test.go` - change `retryWithBackoff` to `RetryWithBackoff` in all tests.

**Step 7: Run test to verify it passes**

Run: `go test ./internal/llm/claude -v`
Expected: PASS (or SKIP if claude not installed)

**Step 8: Commit Claude provider**

```bash
git add internal/llm/claude/provider.go internal/llm/claude/provider_test.go internal/llm/options.go internal/llm/retry.go internal/llm/retry_test.go
git commit -m "feat(llm/claude): implement Claude CLI provider"
```

---

### Task 2.3: Define Model Constants for Gemini

**Files:**
- Create: `internal/llm/gemini/models.go`

**Step 1: Create model constants**

```go
package gemini

// Model name constants for Gemini API
// Gemini requires full model names
const (
	ModelFlash     = "gemini-1.5-flash"
	ModelFlash8B   = "gemini-1.5-flash-8b"
	ModelPro       = "gemini-1.5-pro"
	ModelProLatest = "gemini-1.5-pro-latest"
)

// DefaultModel returns the default Gemini model
func DefaultModel() string {
	return ModelFlash
}
```

**Step 2: Commit model constants**

```bash
git add internal/llm/gemini/models.go
git commit -m "feat(llm/gemini): add model name constants"
```

---

### Task 2.4: Implement Gemini Provider

**Files:**
- Create: `internal/llm/gemini/provider.go`
- Create: `internal/llm/gemini/provider_test.go`

**Step 1: Write failing test for Gemini provider**

```go
package gemini

import (
	"context"
	"os"
	"testing"

	"github.com/connorhough/smix/internal/llm"
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
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/llm/gemini -run TestGeminiProvider`
Expected: FAIL with "undefined: Provider"

**Step 3: Implement Gemini provider**

```go
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
func (p *Provider) Close() error {
	if p.client != nil {
		return p.client.Close()
	}
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
		model := p.client.GenerativeModel(modelName)
		resp, err := model.GenerateContent(ctx, genai.Text(prompt))
		if err != nil {
			return "", p.wrapError(err)
		}

		// Extract text from response
		if len(resp.Candidates) == 0 {
			return "", fmt.Errorf("gemini API returned no candidates")
		}

		var result strings.Builder
		for _, part := range resp.Candidates[0].Content.Parts {
			if text, ok := part.(genai.Text); ok {
				result.WriteString(string(text))
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
	errMsg := err.Error()

	// Check for common error patterns
	if strings.Contains(errMsg, "API key") || strings.Contains(errMsg, "authentication") {
		return llm.ErrAuthenticationFailed("gemini", err)
	}

	if strings.Contains(errMsg, "rate limit") || strings.Contains(errMsg, "quota") {
		return llm.ErrRateLimitExceeded("gemini")
	}

	if strings.Contains(errMsg, "model") && strings.Contains(errMsg, "not found") {
		return llm.ErrModelNotFound("unknown", "gemini")
	}

	// Generic error
	return fmt.Errorf("gemini API error: %w", err)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/llm/gemini -v`
Expected: PASS for unit tests, SKIP for integration test (unless GEMINI_API_KEY set)

**Step 5: Commit Gemini provider**

```bash
git add internal/llm/gemini/provider.go internal/llm/gemini/provider_test.go
git commit -m "feat(llm/gemini): implement Gemini SDK provider"
```

---

### Task 2.5: Implement Provider Factory with Caching

**Files:**
- Create: `internal/llm/factory.go`
- Create: `internal/llm/factory_test.go`

**Step 1: Write failing test for factory**

```go
package llm

import (
	"os"
	"sync"
	"testing"
)

func TestFactory_GetProvider(t *testing.T) {
	factory := NewFactory()

	t.Run("creates claude provider", func(t *testing.T) {
		provider, err := factory.GetProvider("claude", "")
		if err != nil {
			// Skip if claude CLI not available
			if _, ok := err.(*ProviderError); ok {
				t.Skip("claude CLI not available")
			}
			t.Fatalf("unexpected error: %v", err)
		}

		if provider.Name() != "claude" {
			t.Errorf("got provider %q, want %q", provider.Name(), "claude")
		}
	})

	t.Run("creates gemini provider with API key", func(t *testing.T) {
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			apiKey = "test-key-for-unit-test"
		}

		provider, err := factory.GetProvider("gemini", apiKey)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if provider.Name() != "gemini" {
			t.Errorf("got provider %q, want %q", provider.Name(), "gemini")
		}
	})

	t.Run("returns cached provider", func(t *testing.T) {
		// Get provider twice
		p1, err := factory.GetProvider("claude", "")
		if err != nil {
			t.Skip("claude CLI not available")
		}

		p2, err := factory.GetProvider("claude", "")
		if err != nil {
			t.Fatalf("unexpected error on second call: %v", err)
		}

		// Should be same instance (pointer equality)
		if p1 != p2 {
			t.Error("expected cached provider instance")
		}
	})

	t.Run("fails for unknown provider", func(t *testing.T) {
		_, err := factory.GetProvider("unknown", "")
		if err == nil {
			t.Error("expected error for unknown provider")
		}
	})

	t.Run("thread-safe concurrent access", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = factory.GetProvider("claude", "")
			}()
		}
		wg.Wait()
		// If we get here without panic, thread safety works
	})
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/llm -run TestFactory`
Expected: FAIL with "undefined: NewFactory"

**Step 3: Implement factory**

```go
package llm

import (
	"fmt"
	"sync"

	"github.com/connorhough/smix/internal/llm/claude"
	"github.com/connorhough/smix/internal/llm/gemini"
)

// Factory creates and caches provider instances
type Factory struct {
	cache map[string]Provider
	mu    sync.RWMutex
}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	return &Factory{
		cache: make(map[string]Provider),
	}
}

// GetProvider returns a provider by name, creating and caching it if needed
func (f *Factory) GetProvider(name, apiKey string) (Provider, error) {
	// Try to get from cache first (read lock)
	f.mu.RLock()
	if provider, ok := f.cache[name]; ok {
		f.mu.RUnlock()
		return provider, nil
	}
	f.mu.RUnlock()

	// Not in cache, create new provider (write lock)
	f.mu.Lock()
	defer f.mu.Unlock()

	// Double-check after acquiring write lock (another goroutine might have created it)
	if provider, ok := f.cache[name]; ok {
		return provider, nil
	}

	// Create provider based on name
	var provider Provider
	var err error

	switch name {
	case "claude":
		provider, err = claude.NewProvider()
	case "gemini":
		provider, err = gemini.NewProvider(apiKey)
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}

	if err != nil {
		return nil, err
	}

	// Cache the provider
	f.cache[name] = provider

	return provider, nil
}

// Global factory instance
var globalFactory = NewFactory()

// GetProvider is a convenience function that uses the global factory
func GetProvider(name, apiKey string) (Provider, error) {
	return globalFactory.GetProvider(name, apiKey)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/llm -run TestFactory -v`
Expected: PASS (some tests may SKIP if dependencies not available)

**Step 5: Commit factory**

```bash
git add internal/llm/factory.go internal/llm/factory_test.go
git commit -m "feat(llm): implement provider factory with caching"
```

---

## Phase 3: Command Refactoring (Simple Commands)

### Task 3.1: Add Debug Flag to Root Command

**Files:**
- Modify: `cmd/root.go`

**Step 1: Add persistent flags to root command**

Find the root command initialization in `cmd/root.go` and add:

```go
var (
	debugFlag    bool
	providerFlag string
	modelFlag    string
)

func init() {
	// Add persistent flags available to all commands
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug output")
	rootCmd.PersistentFlags().StringVar(&providerFlag, "provider", "", "Override LLM provider (claude, gemini)")
	rootCmd.PersistentFlags().StringVar(&modelFlag, "model", "", "Override model name")
}
```

**Step 2: Add debug logging helper**

Add to `cmd/root.go`:

```go
// debugLog prints debug information if debug flag is enabled
func debugLog(format string, args ...interface{}) {
	if debugFlag {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}
```

**Step 3: Commit flag additions**

```bash
git add cmd/root.go
git commit -m "feat(cmd): add debug, provider, and model flags to root command"
```

---

### Task 3.2: Wire Config Initialization into Root Command

**Files:**
- Modify: `cmd/root.go`

**Step 1: Add config initialization to root command init**

Find the `initConfig()` function in `cmd/root.go` and update it to ensure config exists:

```go
func initConfig() {
	// ... existing viper config code ...

	// Ensure config file exists (create from template if needed)
	configFile := viper.ConfigFileUsed()
	if configFile == "" {
		// Use default XDG config path
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			xdgConfig = filepath.Join(home, ".config")
		}

		configFile = filepath.Join(xdgConfig, "smix", "config.yaml")
	}

	// Create config if it doesn't exist
	if err := config.EnsureConfigExists(configFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}

	// Read config
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading config: %v\n", err)
		os.Exit(1)
	}

	debugLog("Using config file: %s", viper.ConfigFileUsed())
}
```

**Step 2: Test config initialization**

Run: `go build && ./builds/smix --help`
Expected: Should create config file if it doesn't exist

**Step 3: Commit config initialization**

```bash
git add cmd/root.go
git commit -m "feat(cmd): add config auto-creation to root command init"
```

---

### Task 3.3: Refactor `ask` Command to Use Provider Interface

**Files:**
- Modify: `internal/ask/ask.go`
- Modify: `cmd/ask.go`

**Step 1: Update ask.go to use provider**

Replace the current implementation in `internal/ask/ask.go`:

```go
package ask

import (
	"context"
	"fmt"

	"github.com/connorhough/smix/internal/config"
	"github.com/connorhough/smix/internal/llm"
)

const promptTemplate = `You are a helpful technical assistant that provides concise, accurate answers to user questions.

Requirements:
1. Provide clear, direct answers without unnecessary elaboration
2. Focus on accuracy and practical information
3. Use plain text formatting (no markdown, code blocks, or special formatting)
4. Keep responses brief but informative (2-4 sentences typically)
5. For technical topics, include key details but avoid overwhelming the user
6. If the question is ambiguous, answer the most common interpretation

Examples:
User: "what is FastAPI"
Output: FastAPI is a modern Python web framework for building APIs. It's known for high performance, automatic API documentation, and type hints for data validation. It uses Python type annotations and is built on Starlette and Pydantic.

User: "does the mv command overwrite duplicate files"
Output: Yes, mv overwrites files by default without prompting. If a file with the same name exists in the destination, it will be replaced. Use mv -i for interactive mode to get a confirmation prompt before overwriting, or mv -n to prevent overwriting entirely.

User's Question: %s`

// Answer processes a user's question and returns a concise answer
func Answer(ctx context.Context, question string, cfg *config.ProviderConfig, debugFn func(string, ...interface{})) (string, error) {
	debugFn("ask command config: provider=%s, model=%s", cfg.Provider, cfg.Model)

	// Get API key for Gemini if needed
	apiKey := ""
	if cfg.Provider == "gemini" {
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return "", fmt.Errorf("GEMINI_API_KEY environment variable required for Gemini provider")
		}
	}

	// Get provider from factory
	provider, err := llm.GetProvider(cfg.Provider, apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	debugFn("Using provider: %s", provider.Name())

	// Build prompt
	prompt := fmt.Sprintf(promptTemplate, question)
	debugFn("Prompt length: %d characters", len(prompt))

	// Generate response
	var opts []llm.Option
	if cfg.Model != "" {
		opts = append(opts, llm.WithModel(cfg.Model))
		debugFn("Using model: %s", cfg.Model)
	} else {
		debugFn("Using default model: %s", provider.DefaultModel())
	}

	return provider.Generate(ctx, prompt, opts...)
}
```

**Step 2: Add os import**

Add to imports in `internal/ask/ask.go`:

```go
import (
	"context"
	"fmt"
	"os"

	"github.com/connorhough/smix/internal/config"
	"github.com/connorhough/smix/internal/llm"
)
```

**Step 3: Update cmd/ask.go to use new signature**

Modify the command handler in `cmd/ask.go`:

```go
RunE: func(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("question required")
	}
	question := args[0]

	// Resolve configuration
	cfg := config.ResolveProviderConfig("ask")
	cfg.ApplyFlags(providerFlag, modelFlag)

	debugLog("Resolved config for 'ask': provider=%s, model=%s", cfg.Provider, cfg.Model)

	// Create context
	ctx := context.Background()

	// Get answer
	answer, err := ask.Answer(ctx, question, cfg, debugLog)
	if err != nil {
		return err
	}

	fmt.Println(answer)
	return nil
},
```

**Step 4: Test ask command**

Run: `go build && ./builds/smix ask "what is FastAPI"`
Expected: Should work with default provider from config

Run: `./builds/smix ask --provider gemini --debug "what is FastAPI"`
Expected: Should use Gemini provider and show debug output

**Step 5: Commit ask refactor**

```bash
git add internal/ask/ask.go cmd/ask.go
git commit -m "refactor(ask): use provider interface for multi-provider support"
```

---

### Task 3.4: Refactor `do` Command to Use Provider Interface

**Files:**
- Modify: `internal/do/translate.go`
- Modify: `cmd/do.go`

**Step 1: Update translate.go to use provider**

Replace the current implementation in `internal/do/translate.go`:

```go
package do

import (
	"context"
	"fmt"
	"os"

	"github.com/connorhough/smix/internal/config"
	"github.com/connorhough/smix/internal/llm"
)

const promptTemplate = `You are a shell command expert for Unix-like systems (Linux, macOS).
Your sole purpose is to translate the user's request into a single, functional, and secure shell command.

Requirements:
1. Output ONLY the raw command with no explanations, preambles, or markdown formatting
2. Ensure commands are safe and won't cause damage to the system
3. Prefer POSIX-compliant commands when possible
4. For complex tasks, chain commands with pipes and logical operators
5. Handle errors gracefully within the command (e.g., using || for fallbacks)
6. Use absolute paths when necessary
7. For process killing, prefer safer methods like fuser over kill with lsof
8. Commands should be one-liners that can be directly executed or piped

Examples:
User: "find all files larger than 50MB in my home directory"
Output: find ~ -type f -size +50M

User: "list the 10 largest files in the current directory"
Output: du -ah . | sort -rh | head -n 10

User: "kill the process listening on port 3000"
Output: fuser -k 3000/tcp

User's Request: %s`

// Translate converts natural language to shell commands
func Translate(ctx context.Context, taskDescription string, cfg *config.ProviderConfig, debugFn func(string, ...interface{})) (string, error) {
	debugFn("do command config: provider=%s, model=%s", cfg.Provider, cfg.Model)

	// Get API key for Gemini if needed
	apiKey := ""
	if cfg.Provider == "gemini" {
		apiKey = os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			return "", fmt.Errorf("GEMINI_API_KEY environment variable required for Gemini provider")
		}
	}

	// Get provider from factory
	provider, err := llm.GetProvider(cfg.Provider, apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	debugFn("Using provider: %s", provider.Name())

	// Build prompt
	prompt := fmt.Sprintf(promptTemplate, taskDescription)
	debugFn("Prompt length: %d characters", len(prompt))

	// Generate response
	var opts []llm.Option
	if cfg.Model != "" {
		opts = append(opts, llm.WithModel(cfg.Model))
		debugFn("Using model: %s", cfg.Model)
	} else {
		debugFn("Using default model: %s", provider.DefaultModel())
	}

	return provider.Generate(ctx, prompt, opts...)
}
```

**Step 2: Update cmd/do.go to use new signature**

Modify the command handler in `cmd/do.go`:

```go
RunE: func(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("task description required")
	}
	taskDescription := args[0]

	// Resolve configuration
	cfg := config.ResolveProviderConfig("do")
	cfg.ApplyFlags(providerFlag, modelFlag)

	debugLog("Resolved config for 'do': provider=%s, model=%s", cfg.Provider, cfg.Model)

	// Create context
	ctx := context.Background()

	// Translate
	command, err := do.Translate(ctx, taskDescription, cfg, debugLog)
	if err != nil {
		return err
	}

	fmt.Println(command)
	return nil
},
```

**Step 3: Test do command**

Run: `go build && ./builds/smix do "list all files in current directory"`
Expected: Should work with default provider from config

Run: `./builds/smix do --provider gemini --debug "list all files in current directory"`
Expected: Should use Gemini provider and show debug output

**Step 4: Commit do refactor**

```bash
git add internal/do/translate.go cmd/do.go
git commit -m "refactor(do): use provider interface for multi-provider support"
```

---

## Phase 4: Documentation & Cleanup

### Task 4.1: Update README.md

**Files:**
- Modify: `README.md`

**Step 1: Update README with multi-provider information**

Add section after installation:

```markdown
## Configuration

smix supports multiple LLM providers. Configuration is stored in `~/.config/smix/config.yaml` (or `$XDG_CONFIG_HOME/smix/config.yaml`).

On first run, a template configuration file is automatically created with sensible defaults.

### Provider Setup

#### Claude (Default)
- **Requires:** Claude Code CLI installed and authenticated
- **Install:** Visit https://claude.ai/code
- **Models:** `haiku`, `sonnet`, `opus`

#### Gemini
- **Requires:** Google AI Studio API key
- **Setup:** Set `GEMINI_API_KEY` environment variable
- **Get API Key:** https://aistudio.google.com/apikey
- **Models:** `gemini-1.5-flash`, `gemini-1.5-pro`

### Configuration Examples

**Global default (all commands use Claude Sonnet):**
```yaml
provider: claude
model: sonnet
```

**Per-command customization (fast/cheap for ask/do, smart for pr):**
```yaml
provider: claude  # global default

commands:
  ask:
    provider: gemini
    model: gemini-1.5-flash
  do:
    provider: gemini
    model: gemini-1.5-flash
  pr:
    provider: claude
    model: sonnet
```

**Override with flags:**
```bash
smix ask --provider gemini --model gemini-1.5-pro "what is FastAPI"
smix do --provider claude --model haiku "list all files"
```

### Configuration Precedence

1. CLI flags (`--provider`, `--model`)
2. Command-specific config (`commands.ask.provider`)
3. Global config (`provider`)

### Debug Mode

Use `--debug` flag to see provider selection and configuration resolution:

```bash
smix ask --debug "what is FastAPI"
```
```

**Step 2: Update command examples**

Update the existing `ask`, `do`, and `pr` examples to mention provider support:

```markdown
### smix ask

Answers short technical questions using your configured LLM provider.

```bash
smix ask "what is the difference between TCP and UDP"

# Use specific provider
smix ask --provider gemini "explain docker volumes"
```

### smix do

Translates natural language to shell commands using your configured LLM provider.

```bash
smix do "list all files larger than 100MB in the current directory"

# Use specific model
smix do --model haiku "kill process on port 3000"
```

### smix pr

**Note:** The `pr` command currently only supports Claude provider in interactive mode.

Processes code review feedback from gemini-code-assist bot.

```bash
smix pr review owner/repo 123
```
```

**Step 3: Test README accuracy**

Manually verify all examples work as documented.

**Step 4: Commit README updates**

```bash
git add README.md
git commit -m "docs: update README with multi-provider configuration guide"
```

---

### Task 4.2: Update Command Help Text

**Files:**
- Modify: `cmd/ask.go`
- Modify: `cmd/do.go`
- Modify: `cmd/pr.go`

**Step 1: Update ask command help**

In `cmd/ask.go`, update the command definition:

```go
var askCmd = &cobra.Command{
	Use:   "ask [question]",
	Short: "Answer short technical questions",
	Long: `Answer short technical questions using your configured LLM provider.

Supports multiple providers (Claude, Gemini) with per-command configuration.

Examples:
  smix ask "what is FastAPI"
  smix ask --provider gemini "how do I list files in linux"
  smix ask --model haiku "what is docker"`,
	// ... rest of command
}
```

**Step 2: Update do command help**

In `cmd/do.go`, update the command definition:

```go
var doCmd = &cobra.Command{
	Use:   "do [task description]",
	Short: "Translate natural language to shell commands",
	Long: `Translate natural language task descriptions into executable shell commands
using your configured LLM provider.

Supports multiple providers (Claude, Gemini) with per-command configuration.

Examples:
  smix do "list all files in current directory"
  smix do --provider gemini "find files larger than 100MB"
  smix do --model haiku "kill process on port 3000"`,
	// ... rest of command
}
```

**Step 3: Update pr command help**

In `cmd/pr.go`, update the command definition to clarify Claude-only:

```go
var reviewCmd = &cobra.Command{
	Use:   "review [owner/repo] [pr_number]",
	Short: "Process code review feedback from gemini-code-assist bot",
	Long: `Fetches code review feedback from gemini-code-assist bot on GitHub PRs
and launches interactive Claude Code sessions to analyze and implement suggestions.

Note: This command currently only supports Claude provider in interactive mode.

Examples:
  smix pr review owner/repo 123
  smix pr review --dir ./gca_review_pr123`,
	// ... rest of command
}
```

**Step 4: Test help text**

Run: `./builds/smix ask --help`
Run: `./builds/smix do --help`
Run: `./builds/smix pr --help`

Expected: All help text displays correctly

**Step 5: Commit help text updates**

```bash
git add cmd/ask.go cmd/do.go cmd/pr.go
git commit -m "docs: update command help text for multi-provider support"
```

---

### Task 4.3: Remove Old Internal/Claude Package (if unused)

**Files:**
- Remove: `internal/claude/` (if no longer needed)

**Step 1: Check if internal/claude is still used**

Run: `grep -r "github.com/connorhough/smix/internal/claude" --include="*.go" .`

**Step 2: If only used by pr command, keep it**

The `pr` command still uses `internal/claude/cli.go` for the `CheckCLI` function.
We should keep this since PR stays Claude-only.

**Step 3: Update internal/claude package docs**

Modify `internal/claude/cli.go` to clarify its purpose:

```go
// Package claude provides utilities for interacting with the Claude Code CLI.
// This package is primarily used by the pr command which remains Claude-only
// for interactive session support. For general LLM provider usage, see internal/llm.
package claude
```

**Step 4: Commit package documentation**

```bash
git add internal/claude/cli.go
git commit -m "docs: clarify internal/claude package scope"
```

---

### Task 4.4: Update CLAUDE.md Project Instructions

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Add LLM architecture documentation**

Add new section to CLAUDE.md after "Architecture":

```markdown
### LLM Provider Architecture

The CLI supports multiple LLM providers through a unified `Provider` interface defined in `internal/llm/`.

**Provider Interface** (`internal/llm/provider.go`):
- `Generate(ctx, prompt, opts)` - Request-response generation
- `ValidateModel(model)` - Model validation (may be no-op)
- `DefaultModel()` - Provider default model
- `Name()` - Provider identifier

**Implementations**:
- **Claude Provider** (`internal/llm/claude/`) - Wraps `claude` CLI binary via `os/exec`
- **Gemini Provider** (`internal/llm/gemini/`) - Uses `google.golang.org/genai` SDK

**Factory Pattern** (`internal/llm/factory.go`):
- Global singleton factory with provider caching
- Thread-safe concurrent access
- `GetProvider(name, apiKey)` returns cached or new provider instance

**Error Handling** (`internal/llm/errors.go`):
- Typed errors: `ErrProviderNotAvailable`, `ErrAuthenticationFailed`, `ErrRateLimitExceeded`, `ErrModelNotFound`
- Retry logic with exponential backoff (3 attempts, 1s initial delay, 30s max)

**Configuration Resolution**:
- Precedence: CLI flags → command config → global config
- Per-command customization supported (`commands.ask.provider`, etc.)
- Debug mode (`--debug` flag) shows resolution logic

**Command Integration**:
- `ask` and `do` commands support all providers
- `pr` command remains Claude-only (interactive mode not supported by API providers)
```

**Step 2: Update "Adding New Commands" section**

Update the example to use the new provider interface:

```markdown
## Adding New Commands

Follow the established pattern to maintain separation of concerns:

1. Create business logic in `internal/newcommand/newcommand.go`:
```go
package newcommand

import (
    "context"
    "fmt"

    "github.com/connorhough/smix/internal/config"
    "github.com/connorhough/smix/internal/llm"
)

func Run(ctx context.Context, input string, cfg *config.ProviderConfig, debugFn func(string, ...interface{})) (string, error) {
    // Get provider
    apiKey := ""
    if cfg.Provider == "gemini" {
        apiKey = os.Getenv("GEMINI_API_KEY")
    }

    provider, err := llm.GetProvider(cfg.Provider, apiKey)
    if err != nil {
        return "", err
    }

    // Use provider
    opts := []llm.Option{}
    if cfg.Model != "" {
        opts = append(opts, llm.WithModel(cfg.Model))
    }

    return provider.Generate(ctx, "your prompt here", opts...)
}
```

2. Create CLI wiring in `cmd/newcommand.go`:
```go
package cmd

func newNewCommandCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "newcommand [input]",
        Short: "Description",
        RunE: func(cmd *cobra.Command, args []string) error {
            cfg := config.ResolveProviderConfig("newcommand")
            cfg.ApplyFlags(providerFlag, modelFlag)

            ctx := context.Background()
            return newcommand.Run(ctx, args[0], cfg, debugLog)
        },
    }
    return cmd
}

func init() {
    rootCmd.AddCommand(newNewCommandCmd())
}
```

3. Commands MUST return errors rather than calling os.Exit() directly (main.go handles exit codes)
4. Use the provider interface for LLM interactions (see `internal/ask/ask.go` for reference)
```

**Step 3: Commit CLAUDE.md updates**

```bash
git add CLAUDE.md
git commit -m "docs: update CLAUDE.md with LLM provider architecture"
```

---

### Task 4.5: Final Integration Testing

**Files:**
- None (manual testing)

**Step 1: Test all commands with different provider combinations**

```bash
# Build fresh binary
make build

# Test ask with Claude
./builds/smix ask "what is FastAPI"

# Test ask with Gemini (requires GEMINI_API_KEY)
export GEMINI_API_KEY="your-key"
./builds/smix ask --provider gemini "what is FastAPI"

# Test do with different models
./builds/smix do --model haiku "list files"
./builds/smix do --provider gemini --model gemini-1.5-flash "find large files"

# Test debug mode
./builds/smix ask --debug "test question"

# Test pr command (should still work, Claude-only)
./builds/smix pr review --help
```

**Step 2: Test config file creation**

```bash
# Remove config file
rm ~/.config/smix/config.yaml

# Run command - should auto-create config
./builds/smix ask "test"

# Verify config was created
cat ~/.config/smix/config.yaml
```

**Step 3: Test configuration precedence**

Edit config file to set global provider to `claude`, then:

```bash
# Should use claude from config
./builds/smix ask "test"

# Should override to gemini
./builds/smix ask --provider gemini "test"
```

Edit config to add command-specific config:

```yaml
commands:
  ask:
    provider: gemini
```

```bash
# Should use gemini from command config
./builds/smix ask "test"

# Should override to claude
./builds/smix ask --provider claude "test"
```

**Step 4: Test error handling**

```bash
# Test invalid provider
./builds/smix ask --provider invalid "test"
# Expected: Error about unknown provider

# Test missing Gemini API key
unset GEMINI_API_KEY
./builds/smix ask --provider gemini "test"
# Expected: Error about missing API key

# Test invalid model (let provider/API fail naturally)
./builds/smix ask --provider claude --model invalid-model "test"
# Expected: Wrapped error from Claude CLI
```

**Step 5: Document test results**

Create test report:

```bash
cat > docs/testing/integration-test-results.md <<EOF
# Integration Test Results - Multi-Provider LLM

## Test Date
$(date)

## Test Environment
- OS: $(uname -s)
- Go Version: $(go version)
- Claude CLI: $(claude --version || echo "not installed")
- Gemini API Key: $([ -n "$GEMINI_API_KEY" ] && echo "set" || echo "not set")

## Test Results

### Config Auto-Creation: ✓ PASS
### Claude Provider: ✓ PASS
### Gemini Provider: ✓ PASS (with API key)
### Debug Mode: ✓ PASS
### Flag Overrides: ✓ PASS
### Error Handling: ✓ PASS

## Notes
[Add any observations or issues found during testing]
EOF
```

**Step 6: Commit test results**

```bash
git add docs/testing/integration-test-results.md
git commit -m "test: add integration test results for multi-provider support"
```

---

## Completion Checklist

Before marking this work as complete, verify:

- [ ] All unit tests pass: `make test`
- [ ] All commands work with both Claude and Gemini providers
- [ ] Config file auto-creation works
- [ ] Debug mode shows helpful information
- [ ] Error messages are clear and actionable
- [ ] README is accurate and up-to-date
- [ ] Command help text is accurate
- [ ] CLAUDE.md reflects new architecture
- [ ] No dead code remaining (except intentional)
- [ ] All commits follow conventional commit format

**Final commit:**

```bash
git add .
git commit -m "feat: complete multi-provider LLM support

- Add provider interface with Claude and Gemini implementations
- Implement config resolution with precedence (flags > command > global)
- Add provider factory with caching
- Refactor ask/do commands to use providers
- Add --debug, --provider, --model flags
- Auto-create config file from template
- Add retry logic with exponential backoff
- Update documentation (README, CLAUDE.md, help text)
- PR command remains Claude-only interactive

BREAKING CHANGE: Configuration file format updated to support multiple providers"
```

---

## Execution Handoff

Plan complete and saved to `docs/plans/2025-12-23-multi-provider-llm.md`.

Two execution options:

**1. Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

**2. Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

Which approach?
