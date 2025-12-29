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

// ProcessReviews processes pull request feedback files and launches interactive provider sessions for each one.
// Requires a provider that implements InteractiveProvider and a TTY.
// If providerOverride is not empty, it will be used instead of the configured provider.
// If modelOverride is not empty, it will be used instead of the configured model.
func ProcessReviews(feedbackDir string, providerOverride string, modelOverride string) error {
	if _, err := os.Stat(feedbackDir); os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' does not exist", feedbackDir)
	}

	// Create IOStreams for interactive mode
	streams := llm.NewIOStreams()

	if !streams.IsInteractive() {
		return fmt.Errorf("pr review command requires an interactive terminal (TTY). This command cannot run in CI/CD pipelines or with redirected stdin")
	}

	// Get configured provider and model for pr command, default to claude code
	var providerName, model string
	if providerOverride != "" {
		providerName = providerOverride
	} else {
		cfg := config.ResolveProviderConfig("pr")
		providerName = cfg.Provider
		if providerName == "" {
			providerName = "claude"
		}
	}

	if modelOverride != "" {
		model = modelOverride
	} else {
		cfg := config.ResolveProviderConfig("pr")
		model = cfg.Model
		// Model can be empty - provider will use its default
	}

	provider, err := providers.GetProvider(providerName)
	if err != nil {
		return fmt.Errorf("failed to get %s provider: %w", providerName, err)
	}

	// Verify provider supports interactive mode
	if _, ok := provider.(llm.InteractiveProvider); !ok {
		return fmt.Errorf("provider %q does not support interactive mode (required for pr command). Interactive mode requires a provider that can yield control of stdin/stdout/stderr", provider.Name())
	}

	// Find all feedback markdown files excluding INDEX.md
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
	fmt.Printf("Using interactive provider: %s\n", provider.Name())
	fmt.Println("Launching interactive sessions for each feedback item...")
	fmt.Println()

	ctx := context.Background()

	for i, feedbackFile := range filteredFiles {
		basename := filepath.Base(feedbackFile)

		fmt.Println("--------")
		fmt.Printf("Processing [%d/%d]: %s\n", i+1, totalCount, basename)
		fmt.Println("--------")
		fmt.Println()

		targetFile := extractTargetFile(feedbackFile)

		fmt.Printf("Launching interactive session...\n")
		if err := LaunchClaudeCode(ctx, provider, streams, feedbackFile, targetFile, i+1, totalCount, model); err != nil {
			fmt.Printf("Failed to launch interactive session: %v\n", err)
		}

		fmt.Println()
	}

	fmt.Println("--------")
	fmt.Println("All feedback items processed!")
	fmt.Println("--------")

	return nil
}

// LaunchClaudeCode opens an interactive session with the provider to review feedback and implement changes.
// If model is not empty, it will be passed to the provider via WithModel option.
func LaunchClaudeCode(ctx context.Context, provider llm.Provider, streams *llm.IOStreams, feedbackFile, targetFile string, currentIndex, totalCount int, model string) error {
	// Verify streams are interactive
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

	prompt := fmt.Sprintf(`You are a Senior Software Engineer tasked with triaging and applying automated code review feedback.

**Feedback Context:**
%s%s%s

## Your Agent Protocol
You have access to file system tools and linters. You must execute the following steps in order:

1.  **Investigate:**
    * Read the feedback context provided above to understand the issue.
    * **Tool Use:** Use your file reading tool to load the *current* version of the target file from disk. Do not rely solely on the feedback snippet.
    * If the file does not exist, stop and report "File not found."

2.  **Evaluate (Think Step):**
    * Analyze the code. Is the feedback technically correct?
    * Is this actionable? (Ignore low-value style nits unless they fix a linter violation).
    * Does the fix introduce security risks or break existing logic?

3.  **Execute (If Applying):**
		* **Tool Use:** Plan out the changes that need to be made to address the feedback
    * **Tool Use:** Apply the fix to the file using your editing tool if the changes are minimal. If the fix requires medium or large changes, dispatch sub-agents to solve each task, or prompt the user to manually launch sub-agents according to the tasks laid out in the plan.
    * **Tool Use:** Run the appropriate linter/formatter for this file type (e.g., 'gofmt', 'eslint', 'black') to ensure the new code is valid.
    * If the linter fails, attempt to self-correct or revert the changes.

4.  **Execute (If Rejecting):**
    * Do not modify any files.

## Final Report (Output to User)
After completing your actions, provide a concise summary in the following format:

**STATUS:** [APPLIED | REJECTED | FAILED]
**FILE:** [File Path]
**ACTION TAKEN:** [One sentence summary of what you did, e.g., "Updated regex to fix ReDoS vulnerability and ran gofmt."]
**REASONING:** [Brief explanation of why you made this decision.]
`, feedbackFile, targetFileInfo, batchInfo)

	// Use provider's interactive mode with injected streams
	var opts []llm.Option
	if model != "" {
		opts = append(opts, llm.WithModel(model))
	}
	return interactive.RunInteractive(ctx, streams, prompt, opts...)
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
