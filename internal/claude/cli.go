// Package claude provides shared utilities for interacting with the Claude Code CLI.
package claude

import (
	"fmt"
	"os/exec"
)

// CheckCLI verifies that the claude CLI is available in the system PATH.
// Returns an error if the CLI is not found.
func CheckCLI() error {
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude CLI not found in PATH. Please install Claude Code from https://claude.ai/code")
	}
	return nil
}
