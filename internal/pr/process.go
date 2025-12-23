// Package pr processes code review feedback files and launches Claude Code sessions for interactive refinement.
package pr

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

	totalCount := len(filteredFiles)
	fmt.Printf("Found %d feedback files to process\n", totalCount)
	fmt.Println("Launching Claude Code sessions for each feedback item...")
	fmt.Println()

	for i, feedbackFile := range filteredFiles {
		basename := filepath.Base(feedbackFile)

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("Processing [%d/%d]: %s\n", i+1, totalCount, basename)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		// Extract target file from feedback file
		targetFile := extractTargetFile(feedbackFile)

		// Launch Claude Code session
		fmt.Printf("Launching Claude Code...\n")
		if err := LaunchClaudeCode(feedbackFile, targetFile, i+1, totalCount); err != nil {
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
func LaunchClaudeCode(feedbackFile, targetFile string, currentIndex, totalCount int) error {
	batchInfo := ""
	if totalCount > 1 {
		batchInfo = fmt.Sprintf("\n\n**Note:** This is feedback item %d of %d in this PR. Focus only on this item.", currentIndex, totalCount)
	}

	targetFileInfo := ""
	if targetFile != "" {
		targetFileInfo = fmt.Sprintf("\n**Target file to modify (if applying):** `%s`", targetFile)
	}

	prompt := fmt.Sprintf(`You are an autonomous code review agent. Your task is to process feedback from an automated reviewer and decide whether to apply it.

**Read the feedback file:** %s%s%s

## Execution Protocol

1. **Read** the feedback file to understand the suggestion and context.
2. **Read** the actual target file from disk (not just the snapshot in the feedback file).
3. **Evaluate** the feedback:
   - Is it technically correct?
   - Does it align with the patterns already in this codebase?
   - Is it high-value (bugs, security, correctness) or low-value (style nits, micro-optimizations)?
4. **Make your decision:**
   - If APPLY: Edit the target file directly. Run any relevant linter/formatter if available (e.g., gofmt, eslint).
   - If REJECT: Do not modify any files.
5. **Output** your reasoning using the format specified in the feedback file.

## Constraints

- **DO NOT** run tests unless explicitly asked
- **DO NOT** commit changes
- **DO NOT** modify files other than the target file
- **DO NOT** add features beyond what the feedback requests
- If the target file does not exist, output "SKIP: File not found" and explain

## Decision Format

Follow the format specified in the feedback file for consistency.
`, feedbackFile, targetFileInfo, batchInfo)

	// Launch Claude in interactive mode (without -p flag) with initial prompt
	cmd := exec.Command("claude", prompt)

	// Connect interactive session io to current terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// extractTargetFile extracts the target file path from a feedback markdown file
func extractTargetFile(feedbackFile string) string {
	content, err := os.ReadFile(feedbackFile)
	if err != nil {
		return ""
	}

	// Look for "**Target File:** `path/to/file`" in the markdown
	re := regexp.MustCompile(`(?m)^- \*\*Target File:\*\* ` + "`" + `([^` + "`" + `]+)` + "`" + `$`)
	matches := re.FindStringSubmatch(string(content))
	if len(matches) > 1 {
		return matches[1]
	}

	return ""
}
