// Package do provides functionality for translating natural language descriptions
// into executable shell commands using LLM providers.
package do

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/connorhough/smix/internal/config"
	"github.com/connorhough/smix/internal/llm"
	"github.com/connorhough/smix/internal/providers"
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
func Translate(ctx context.Context, taskDescription string, cfg *config.ProviderConfig) (string, error) {
	slog.Debug("do command config: provider=%s, model=%s", cfg.Provider, cfg.Model)

	provider, err := providers.GetProvider(cfg.Provider)
	if err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	slog.Debug("using provider", "name", provider.Name())

	prompt := fmt.Sprintf(promptTemplate, taskDescription)
	slog.Debug("prompt constructed", "length", len(prompt))

	// Generate response
	var opts []llm.Option
	resolvedModel := cfg.Model
	if resolvedModel == "" {
		resolvedModel = provider.DefaultModel()
	} else {
		opts = append(opts, llm.WithModel(resolvedModel))
	}

	slog.Debug("resolved model", "model", resolvedModel)

	return provider.Generate(ctx, prompt, opts...)
}
