// Package gca processes code review feedback files and launches Claude Code sessions for interactive refinement.
package gca

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// ProcessReviews processes gemini-code-assist feedback files and launches Claude Code sessions for each one
func ProcessReviews(feedbackDir string) error {
	if _, err := os.Stat(feedbackDir); os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' does not exist", feedbackDir)
	}

	// Find all feedback markdown files (excluding INDEX.md)
	feedbackFiles, err := filepath.Glob(filepath.Join(feedbackDir, "*.md"))
	if err != nil {
		return fmt.Errorf("failed to find feedback files: %w", err)
	}

	var filteredFiles []string
	for _, file := range feedbackFiles {
		if filepath.Base(file) != "INDEX.md" {
			filteredFiles = append(filteredFiles, file)
		}
	}

	if len(filteredFiles) == 0 {
		return fmt.Errorf("no feedback files found in %s", feedbackDir)
	}

	fmt.Printf("Found %d feedback files to process\n", len(filteredFiles))
	fmt.Println("Launching Claude Code sessions for each feedback item...")
	fmt.Println()

	for i, feedbackFile := range filteredFiles {
		basename := filepath.Base(feedbackFile)

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("Processing [%d/%d]: %s\n", i+1, len(filteredFiles), basename)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		// Launch Claude Code session
		fmt.Printf("Launching Claude Code...\n")
		if err := LaunchClaudeCode(feedbackFile); err != nil {
			fmt.Printf("Failed to launch Claude Code: %v\n", err)
		}

		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✓ All feedback items processed!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}

// LaunchClaudeCode opens Claude Code with a prompt to review the feedback and implement changes
func LaunchClaudeCode(feedbackFile string) error {
	prompt := fmt.Sprintf(
		"Act as a senior engineer reviewing code feedback. "+
			"Read %s and: "+
			"1) Critically evaluate if the gemini-code-assist feedback is valid and important "+
			"2) Decide to APPLY (as-is or modified) or REJECT it with reasoning "+
			"3) If applying, implement the changes directly in the codebase "+
			"4) Explain your decision and the changes made",
		feedbackFile,
	)

	// Launch Claude in interactive mode (without -p flag) with initial prompt
	cmd := exec.Command("claude", prompt)

	// Connect interactive session io to current terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
