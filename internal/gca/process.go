package gca

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

	if len(filteredFiles) == 0 {
		return fmt.Errorf("no feedback files found in %s", feedbackDir)
	}

	// Create output directory for results
	resultsDir := filepath.Join(feedbackDir, "results")
	if err := os.MkdirAll(resultsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create results directory: %w", err)
	}

	fmt.Printf("Found %d feedback files to process\n", len(filteredFiles))
	fmt.Println("Launching Crush sessions for each patch...")
	fmt.Println()

	// Process each feedback file
	for i, feedbackFile := range filteredFiles {
		basename := filepath.Base(feedbackFile)
		sessionID := getSessionID(basename)

		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Printf("Processing [%d/%d]: %s\n", i+1, len(filteredFiles), basename)
		fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		fmt.Println()

		// Read the feedback content
		feedbackContent, err := os.ReadFile(feedbackFile)
		if err != nil {
			return fmt.Errorf("failed to read feedback file %s: %w", feedbackFile, err)
		}

		// Generate patch using LLM
		patch, err := generatePatch(string(feedbackContent))
		if err != nil {
			return fmt.Errorf("failed to generate patch for %s: %w", feedbackFile, err)
		}

		// Check if the patch is rejected
		if after, found := strings.CutPrefix(patch.Content, "REJECT:"); found {
			fmt.Printf("⚠ Patch rejected: %s\n", after)
			fmt.Println()
			continue
		}

		// Save the patch to a file
		outputFile := filepath.Join(resultsDir, fmt.Sprintf("%s_patch.diff", sessionID))
		if err := os.WriteFile(outputFile, []byte(patch.Content), 0o644); err != nil {
			return fmt.Errorf("failed to save patch to %s: %w", outputFile, err)
		}
		fmt.Printf("✓ Patch saved to: %s\n", outputFile)

		// Put a prompt with the output file path in clipboard and launch Crush in interactive mode
		fmt.Printf("Launching Crush...\n")
		if err := clipboard.Init(); err != nil {
			fmt.Printf("Failed to initialize clipboard: %v\n", err)
		} else {
			clipboardContent := "Review and apply " + outputFile
			clipboard.Write(clipboard.FmtText, []byte(clipboardContent))
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
	// Get API key from environment
	apiKey := os.Getenv("CEREBRAS_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("CEREBRAS_API_KEY environment variable not set")
	}

	// Create authenticated client
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.cerebras.ai/v1"
	client := openai.NewClientWithConfig(config)

	// Craft improved prompt for patch generation
	systemPrompt := `You are an expert software engineer tasked with applying code review suggestions.
Your goal is to generate precise git-style diffs that can be automatically applied to code files.

Requirements:
1. Analyze the code feedback carefully in context
2. Generate a git-style diff with proper formatting
3. Include 2-3 lines of context before and after changes
4. Use '-' for lines to be removed and '+' for lines to be added
5. If the feedback is invalid or no change is needed, respond with "REJECT: <reason>"

Example response format:
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

Or if rejecting:
REJECT: The feedback suggestion is not applicable because...`

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
	// Create the command
	cmd := exec.Command("crush")

	// Connect the interactive session to the current terminal
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
