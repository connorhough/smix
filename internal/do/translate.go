// Package do provides functionality for translating natural language descriptions
// into executable shell commands using Claude Code CLI.
package do

import (
	"fmt"
	"os/exec"
	"strings"
)

// Translate converts natural language to shell commands using Claude Code CLI
func Translate(taskDescription string) (string, error) {
	// Check if claude CLI is available
	if _, err := exec.LookPath("claude"); err != nil {
		return "", fmt.Errorf("claude CLI not found in PATH. Please install Claude Code from https://claude.ai/code")
	}

	prompt := fmt.Sprintf(`You are a shell command expert for Unix-like systems (Linux, macOS).
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

User's Request: %s`, taskDescription)

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
