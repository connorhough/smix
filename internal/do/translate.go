package do

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// Translate converts natural language to shell commands using Cerebras API
func Translate(taskDescription string) (string, error) {
	// Get API key from environment
	apiKey := os.Getenv("CEREBRAS_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("CEREBRAS_API_KEY environment variable not set")
	}

	// Create authenticated client
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.cerebras.ai/v1"
	client := openai.NewClientWithConfig(config)

	// Craft an improved prompt for Cerebras
	systemPrompt := `You are a shell command expert for Unix-like systems (Linux, macOS). 
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
Output: fuser -k 3000/tcp`

	// Create chat completion request
	resp, err := client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: "qwen-3-coder-480b",
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("User's Request: %s", taskDescription),
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to get inference from Cerebras: %w", err)
	}

	// Return trimmed output
	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}