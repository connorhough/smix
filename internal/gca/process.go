package gca

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProcessReviews processes gemini-code-assist feedback files with Crush
func ProcessReviews(ctx context.Context, feedbackDir string, interactive bool) error {
	// Check if feedback directory exists
	if _, err := os.Stat(feedbackDir); os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' does not exist", feedbackDir)
	}

	// Find all feedback markdown files (excluding INDEX.md)
	feedbackFiles, err := filepath.Glob(filepath.Join(feedbackDir, "*.md"))
	if err != nil {
		return fmt.Errorf("failed to find feedback files: %w", err)
	}

	// Filter out INDEX.md
	var filteredFiles []string
	for _, file := range feedbackFiles {
		if filepath.Base(file) != "INDEX.md" {
			filteredFiles = append(filteredFiles, file)
		}
	}

	// Sort files by name to process them in order
	// Note: This is a simple sort, not a version sort like the shell script
	// For a more sophisticated sort, we would need additional implementation

	if len(filteredFiles) == 0 {
		return fmt.Errorf("no feedback files found in %s", feedbackDir)
	}

	// Create output directory for results
	resultsDir := filepath.Join(feedbackDir, "crush_results")
	if err := os.MkdirAll(resultsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	fmt.Printf("Found %d feedback files to process\n", len(filteredFiles))
	if interactive {
		fmt.Println("Mode: Interactive")
	} else {
		fmt.Println("Mode: Automated")
	}
	fmt.Println()

	// Check if Crush is installed
	if _, err := exec.LookPath("crush"); err != nil {
		return fmt.Errorf("crush is not installed or not in your PATH. Please install it from https://github.com/charmbracelet/crush")
	}

	// Process each feedback file
	for i, feedbackFile := range filteredFiles {
		sessionID := getSessionID(feedbackFile)
		basename := filepath.Base(feedbackFile)

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("Processing [%d/%d]: %s\n", i+1, len(filteredFiles), basename)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		if interactive {
			// Interactive mode: launch full Crush session
			sessionDir := filepath.Join(resultsDir, fmt.Sprintf("%s_session", sessionID))
			if err := os.MkdirAll(sessionDir, 0o755); err != nil {
				return fmt.Errorf("failed to create session directory: %w", err)
			}

			// Copy feedback to session directory
			destFile := filepath.Join(sessionDir, "FEEDBACK.md")
			content, err := os.ReadFile(feedbackFile)
			if err != nil {
				return fmt.Errorf("failed to read feedback file %s: %w", feedbackFile, err)
			}

			if err := os.WriteFile(destFile, content, 0o644); err != nil {
				return fmt.Errorf("failed to copy feedback file to session directory: %w", err)
			}

			// Create a session info file
			sessionInfo := fmt.Sprintf(`# Review Session: %s

**Created:** %s
**Feedback File:** %s

## Quick Start

The feedback from gemini-code-assist is in 'FEEDBACK.md'.

Suggested prompts to get started:
%s
Read FEEDBACK.md and provide your critical analysis of the suggestion
%s

%s
Analyze FEEDBACK.md and implement the best solution
%s

When done, type 'exit' or press Ctrl+D to move to the next feedback item.
`, basename, basename, filepath.Base(feedbackFile), "```", "```", "```", "```")

			sessionInfoFile := filepath.Join(sessionDir, "SESSION_INFO.md")
			if err := os.WriteFile(sessionInfoFile, []byte(sessionInfo), 0o644); err != nil {
				return fmt.Errorf("failed to create session info file: %w", err)
			}

			fmt.Printf("Session directory: %s\n", sessionDir)
			fmt.Println()
			fmt.Println("Starting interactive Crush session...")
			fmt.Println("The feedback is in: FEEDBACK.md")
			fmt.Println()

			// Launch interactive crush in the session directory
			cmd := exec.Command("crush", "-c", sessionDir)
			cmd.Dir = sessionDir
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				fmt.Printf("⚠ Crush session encountered an error: %v\n", err)
			}

			fmt.Println()
			fmt.Println("✓ Session completed")
		} else {
			// Automated mode: use 'crush run' for non-interactive processing
			fmt.Println("Running automated analysis with Crush...")
			fmt.Println()

			// Read the feedback content
			feedbackContent, err := os.ReadFile(feedbackFile)
			if err != nil {
				return fmt.Errorf("failed to read feedback file %s: %w", feedbackFile, err)
			}

			// Create the prompt for crush
			prompt := fmt.Sprintf(`You are reviewing code feedback from an AI code review tool. Please analyze the following feedback critically and provide your recommendation.

%s

Please provide:
1. Your decision (ACCEPT/MODIFY/REJECT)
2. Your reasoning
3. Specific code changes if applicable
4. Any alternative approaches considered`, string(feedbackContent))

			// Create output file for this session
			outputFile := filepath.Join(resultsDir, fmt.Sprintf("%s_analysis.md", sessionID))

			// Run crush with the prompt and save output
			cmd := exec.Command("crush", "run", prompt)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Printf("⚠ Crush encountered an error: %v\n", err)
				fmt.Printf("Error logged to: %s\n", outputFile)
			}

			// Save output to file
			if err := os.WriteFile(outputFile, output, 0o644); err != nil {
				return fmt.Errorf("failed to save analysis to %s: %w", outputFile, err)
			}

			fmt.Printf("✓ Analysis saved to: %s\n", outputFile)
		}

		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✓ All feedback items processed!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	if interactive {
		fmt.Printf("Interactive sessions created in: %s\n", resultsDir)
		fmt.Println()
		fmt.Println("To review a session later:")
		fmt.Printf("  cd %s/<session-id>_session\n", resultsDir)
		fmt.Println("  crush")
	} else {
		fmt.Printf("Analysis results saved in: %s\n", resultsDir)
		fmt.Println()
		fmt.Println("Review files:")
		for _, feedbackFile := range filteredFiles {
			sessionID := getSessionID(feedbackFile)
			fmt.Printf("  - %s_analysis.md\n", sessionID)
		}
	}

	fmt.Println()
	fmt.Println("Sessions processed:")
	for _, feedbackFile := range filteredFiles {
		sessionID := getSessionID(feedbackFile)
		fmt.Printf("  - %s\n", sessionID)
	}

	return nil
}

func getSessionID(filename string) string {
	basename := filepath.Base(filename)
	ext := filepath.Ext(basename)
	return strings.TrimSuffix(basename, ext)
}
