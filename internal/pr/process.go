// Package pr processes code review feedback files and launches Claude Code sessions for interactive refinement.
package pr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/connorhough/smix/internal/config"
	"github.com/connorhough/smix/internal/llm"
	"github.com/connorhough/smix/internal/providers"
)

// ProcessReviews processes gemini-code-assist feedback files and launches interactive provider sessions for each one.
// Requires a provider that implements InteractiveProvider (e.g., Claude) and a TTY (interactive terminal).
func ProcessReviews(feedbackDir string) error {
	if _, err := os.Stat(feedbackDir); os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' does not exist", feedbackDir)
	}

	// Create IOStreams for interactive mode
	// This connects to os.Stdin/Stdout/Stderr and includes TTY detection
	streams := llm.NewIOStreams()

	// Early check: pr command requires interactive terminal
	if !streams.IsInteractive() {
		return fmt.Errorf("pr review command requires an interactive terminal (TTY). This command cannot run in CI/CD pipelines or with redirected stdin")
	}

	// Get configured provider for pr command
	// Defaults to "claude" if not explicitly configured
	cfg := config.ResolveProviderConfig("pr")
	providerName := cfg.Provider
	if providerName == "" {
		providerName = "claude" // Default to claude for backward compatibility
	}

	provider, err := providers.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("failed to get %s provider: %w", providerName, err)
	}

	// Verify it supports interactive mode
	// This type assertion allows any provider to implement InteractiveProvider,
	// whether by shelling out to a CLI or implementing a REPL loop against an API
	if _, ok := provider.(llm.InteractiveProvider); !ok {
		return fmt.Errorf("provider %q does not support interactive mode (required for pr command). Interactive mode requires a provider that can yield control of stdin/stdout/stderr", provider.Name())
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
	fmt.Printf("Using provider: %s (interactive mode)\n", provider.Name())
	fmt.Println("Launching interactive sessions for each feedback item...")
	fmt.Println()

	ctx := context.Background()

	for i, feedbackFile := range filteredFiles {
		basename := filepath.Base(feedbackFile)

		fmt.Println("--------")
		fmt.Printf("Processing [%d/%d]: %s\n", i+1, totalCount, basename)
		fmt.Println("--------")
		fmt.Println()

		// Extract target file from feedback file
		targetFile := extractTargetFile(feedbackFile)

		// Launch interactive session with streams
		fmt.Printf("Launching interactive session...\n")
		if err := LaunchClaudeCode(ctx, provider, streams, feedbackFile, targetFile, i+1, totalCount); err != nil {
			fmt.Printf("Failed to launch interactive session: %v\n", err)
			// Continue processing other files even if one fails
		}

		fmt.Println()
	}

	fmt.Println("--------")
	fmt.Println("All feedback items processed!")
	fmt.Println("--------")

	return nil
}

// LaunchClaudeCode opens an interactive session with the provider to review feedback and implement changes.
// The provider must implement llm.InteractiveProvider, and streams must be interactive (TTY).
//
// This function performs TTY validation at the call site (not inside the provider) following
// the IOStreams dependency injection pattern.
func LaunchClaudeCode(ctx context.Context, provider llm.Provider, streams *llm.IOStreams, feedbackFile, targetFile string, currentIndex, totalCount int) error {
	// Verify streams are interactive (TTY check at call site, not in provider)
	if !streams.IsInteractive() {
		return fmt.Errorf("interactive mode requires a terminal (TTY), but stdin is not a terminal. This can happen when running in CI/CD pipelines or when stdin is redirected")
	}

	// Verify provider supports interactive mode
	interactive, ok := provider.(llm.InteractiveProvider)
	if !ok {
		return fmt.Errorf("provider %q does not support interactive mode", provider.Name())
	}

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

	// Use provider's interactive mode with injected streams
	return interactive.RunInteractive(ctx, streams, prompt)
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
