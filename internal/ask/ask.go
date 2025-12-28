// Package ask provides functionality for answering short questions using LLM providers.
package ask

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/connorhough/smix/internal/config"
	"github.com/connorhough/smix/internal/llm"
	"github.com/connorhough/smix/internal/providers"
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
func Answer(ctx context.Context, question string, cfg *config.ProviderConfig) (string, error) {
	slog.Debug("ask command config", "provider", cfg.Provider, "model", cfg.Model)

	// Get provider from factory
	provider, err := providers.GetProvider(cfg.Provider)
	if err != nil {
		return "", fmt.Errorf("failed to get provider: %w", err)
	}

	slog.Debug("resolved provider", "name", provider.Name())

	// Build prompt
	prompt := fmt.Sprintf(promptTemplate, question)
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
