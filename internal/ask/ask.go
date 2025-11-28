// Package ask provides functionality for answering short questions using Claude Code CLI.
package ask

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/connorhough/smix/internal/claude"
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

// Answer processes a user's question and returns a concise answer using Claude Code CLI
func Answer(question string) (string, error) {
	// Check if claude CLI is available
	if err := claude.CheckCLI(); err != nil {
		return "", err
	}

	prompt := fmt.Sprintf(promptTemplate, question)

	cmd := exec.Command("claude", "-p", prompt)

	outputBytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claude CLI failed: %w (output: %s)", err, string(outputBytes))
	}

	// Return trimmed output
	output := strings.TrimSpace(string(outputBytes))
	if output == "" {
		return "", fmt.Errorf("claude CLI returned empty response")
	}

	return output, nil
}
