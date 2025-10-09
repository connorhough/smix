// Package gca processes code review feedback files by generating patches using an LLM and launching Crush sessions for interactive refinement.
package gca

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/connorhough/smix/internal/llm"
	"github.com/sashabaranov/go-openai"
	"golang.design/x/clipboard"
)

// Patch represents a code change to be applied
type Patch struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

// ProcessReviews processes gemini-code-assist feedback files with direct LLM API calls and launches Crush sessions
func ProcessReviews(ctx context.Context, feedbackDir string) error {
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

	resultsDir := filepath.Join(feedbackDir, "results")
	if err := os.MkdirAll(resultsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	fmt.Printf("Found %d feedback files to process\n", len(filteredFiles))
	fmt.Println("Launching Crush sessions for each patch...")
	fmt.Println()

	for i, feedbackFile := range filteredFiles {
		basename := filepath.Base(feedbackFile)
		sessionID := getSessionID(basename)

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("Processing [%d/%d]: %s\n", i+1, len(filteredFiles), basename)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		feedbackContent, err := os.ReadFile(feedbackFile)
		if err != nil {
			return fmt.Errorf("failed to read feedback file %s: %w", feedbackFile, err)
		}

		patch, err := generatePatch(string(feedbackContent))
		if err != nil {
			return fmt.Errorf("failed to generate patch for %s: %w", feedbackFile, err)
		}

		// Check if the patch is rejected
		content := patch.Content
		if strings.HasPrefix(content, "```diff") {
			content = strings.TrimPrefix(content, "```diff")
			content = strings.TrimSuffix(content, "```")
			content = strings.TrimSpace(content)
		}
		if after, found := strings.CutPrefix(content, "REJECT:"); found {
			fmt.Printf("⚠ Patch rejected: %s\n", after)
			fmt.Println()
			continue
		}

		outputFile := filepath.Join(resultsDir, fmt.Sprintf("%s_patch.diff", sessionID))
		if err := os.WriteFile(outputFile, []byte(patch.Content), 0o644); err != nil {
			return fmt.Errorf("failed to save patch to %s: %w", outputFile, err)
		}
		fmt.Printf("✓ Patch saved to: %s\n", outputFile)

		// Copy prompt to clipboard and launch crush session
		fmt.Printf("Launching Crush...\n")
		if err := clipboard.Init(); err != nil {
			fmt.Printf("Failed to initialize clipboard: %v\n", err)
		} else {
			systemPrompt := fmt.Sprintf(
				"Review the gemini-code-review feedback in %s and the proposed solution in %s. "+
					"Explain the suggested changes and their benefits to the author. "+
					"Then help implement the changes they approve, adapting the solutions as needed based on their input.",
				feedbackFile,
				outputFile,
			)
			clipboard.Write(clipboard.FmtText, []byte(systemPrompt))
		}
		if err := LaunchCrush(); err != nil {
			fmt.Printf("Failed to launch Crush: %v\n", err)
		}

		fmt.Println()
	}

	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✓ All feedback items processed!")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	return nil
}

func getSessionID(filename string) string {
	basename := filepath.Base(filename)
	ext := filepath.Ext(basename)
	return strings.TrimSuffix(basename, ext)
}

// generatePatch generates a patch for the given feedback using a direct LLM API call
func generatePatch(feedbackContent string) (*Patch, error) {
	client, err := llm.NewCerebrasClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Cerebras client: %w", err)
	}

	systemPrompt := `You are an expert software engineer tasked with applying code review suggestions.
Your goal is to generate precise git-style diffs that can be automatically applied to code files.

Requirements:
1. Analyze the code feedback carefully in context
2. Generate a git-style diff with proper formatting
3. Include 2-3 lines of context before and after changes
4. Use '-' for lines to be removed and '+' for lines to be added
5. Preserve exact indentation and whitespace in context lines
6. If the feedback should be rejected, respond with "REJECT: <reason>"

REJECT the feedback if ANY of these conditions apply:
- The feedback is vague, unclear, or lacks specific actionable changes
- The suggested change would introduce bugs or break functionality
- The feedback targets code that doesn't exist in the provided file
- The suggestion contradicts language best practices or project conventions
- The feedback is a question, observation, or praise without requesting a change
- The change would require modifications to other files not provided
- Multiple interpretations exist and the correct one is ambiguous
- The current code is already correct and feedback is mistaken

Example response format for accepted changes:
diff --git a/path/to/file.go b/path/to/file.go
index abc123..def456 100644
--- a/path/to/file.go
+++ b/path/to/file.go
@@ -10,7 +10,7 @@
 func example() {
     // Context lines
-    oldLine
+    newLine
     // More context
 }

Example response format for rejected changes:
REJECT: The feedback suggests changing the authentication method, but this would break compatibility with the existing API contract and requires changes to the database schema in files not provided.`

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: llm.CerebrasProModel,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: fmt.Sprintf("Code Review Feedback:\n%s", feedbackContent),
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get patch from LLM: %w", err)
	}

	return &Patch{
		Content: strings.TrimSpace(resp.Choices[0].Message.Content),
	}, nil
}

// LaunchCrush opens Charm Crush in interactive mode
func LaunchCrush() error {
	cmd := exec.Command("crush")

	// Connect interactive session io to current terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
