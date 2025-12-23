// Package pr fetches code review feedback from GitHub pull requests and saves them as individual prompt files for further processing.
package pr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/go-github/github"
)

// FeedbackItem represents a single feedback item from gemini-code-assist
type FeedbackItem struct {
	Type      string `json:"type"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Body      string `json:"body"`
	DiffHunk  string `json:"diff_hunk"`  // New field
	CommentID int64  `json:"comment_id"` // New field
}

// FetchReviews fetches gemini-code-assist feedback from a GitHub PR
func FetchReviews(ctx context.Context, client *github.Client, repoOwner, repoName string, prNumber int, outputDir string) error {
	if outputDir == "" {
		outputDir = fmt.Sprintf("./pr_feedback_pr%d", prNumber)
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Verify that the PR is accessible
	pr, _, err := client.PullRequests.Get(ctx, repoOwner, repoName, prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR #%d in %s/%s: %w", prNumber, repoOwner, repoName, err)
	}
	fmt.Printf("Successfully fetched PR #%d: %s\n", prNumber, pr.GetTitle())

	// Fetch PR files to get diff hunks
	prFiles, _, err := client.PullRequests.ListFiles(ctx, repoOwner, repoName, prNumber, &github.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to fetch PR files: %w", err)
	}
	fmt.Printf("Fetched %d changed files\n", len(prFiles))

	// Create a map of file paths to diff patches for quick lookup
	fileDiffs := make(map[string]string)
	for _, file := range prFiles {
		if file.Filename != nil && file.Patch != nil {
			fileDiffs[*file.Filename] = *file.Patch
		}
	}

	// Fetch review comments (inline code comments)
	reviewComments, _, err := client.PullRequests.ListComments(ctx, repoOwner, repoName, prNumber, &github.PullRequestListCommentsOptions{})
	if err != nil {
		return fmt.Errorf("failed to fetch review comments: %w", err)
	}
	fmt.Printf("Fetched %d review comments\n", len(reviewComments))

	// Fetch issue comments (general PR comments)
	issueComments, _, err := client.Issues.ListComments(ctx, repoOwner, repoName, prNumber, &github.IssueListCommentsOptions{})
	if err != nil {
		return fmt.Errorf("failed to fetch issue comments: %w", err)
	}
	fmt.Printf("Fetched %d issue comments\n", len(issueComments))

	// Filter comments from gemini-code-assist bot
	var feedbackItems []FeedbackItem

	// Process review comments
	for _, comment := range reviewComments {
		if comment.User != nil && comment.User.Login != nil && strings.Contains(*comment.User.Login, "gemini-code-assist") {
			line := 0
			if comment.Position != nil {
				line = *comment.Position
			} else if comment.OriginalPosition != nil {
				line = *comment.OriginalPosition
			}

			file := *comment.Path
			diffHunk := ""
			if patch, ok := fileDiffs[file]; ok {
				diffHunk = patch
			}

			commentID := int64(0)
			if comment.ID != nil {
				commentID = *comment.ID
			}

			feedbackItems = append(feedbackItems, FeedbackItem{
				Type:      "review_comment",
				File:      file,
				Line:      line,
				Body:      *comment.Body,
				DiffHunk:  diffHunk,
				CommentID: commentID,
			})
		}
	}

	// Process issue comments, excluding summaries
	for _, comment := range issueComments {
		if comment.User != nil && comment.User.Login != nil && strings.Contains(*comment.User.Login, "gemini-code-assist") {
			body := *comment.Body
			// Exclude summary comments
			if !strings.HasPrefix(body, "## Code Review") && !strings.HasPrefix(body, "## Summary") {
				feedbackItems = append(feedbackItems, FeedbackItem{
					Type: "issue_comment",
					File: "",
					Line: 0,
					Body: body,
				})
			}
		}
	}

	if len(feedbackItems) == 0 {
		fmt.Printf("No gemini-code-assist feedback found for PR #%d\n", prNumber)
		return nil
	}

	fmt.Printf("Found %d feedback items\n", len(feedbackItems))
	fmt.Printf("Creating individual prompt files in: %s\n", outputDir)

	// Process each comment and create individual files
	for i, item := range feedbackItems {
		var outputFilePath string
		var fileContent string

		if item.File != "" {
			// Create sanitized filename
			filename := strings.ReplaceAll(strings.ReplaceAll(item.File, "/", "_"), ".", "_")
			outputFilePath = filepath.Join(outputDir, fmt.Sprintf("%d_%s_line%d.md", i+1, filename, item.Line))

			// Fetch the file content for context
			file, _, _, err := client.Repositories.GetContents(ctx, repoOwner, repoName, item.File, &github.RepositoryContentGetOptions{Ref: pr.GetHead().GetSHA()})
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to fetch content of %s: %v\n", item.File, err)
			} else if file != nil {
				content, err := file.GetContent()
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning: failed to decode content of %s: %v\n", item.File, err)
				} else {
					fileContent = content
				}
			}
		} else {
			outputFilePath = filepath.Join(outputDir, fmt.Sprintf("%d_general_comment.md", i+1))
		}

		// Determine start line for the snippet
		startLine := 1
		if item.Line > 0 {
			// Try to show context around the comment line
			startLine = item.Line
			if startLine > 10 {
				startLine -= 10 // Show 10 lines before
			} else {
				startLine = 1
			}
		}

		// Create a code snippet of ~40 lines
		var snippet string
		if fileContent != "" {
			lines := strings.Split(fileContent, "\n")
			startIdx := startLine - 1
			if startIdx < 0 {
				startIdx = 0
			}
			endIdx := startIdx + 40
			if endIdx > len(lines) {
				endIdx = len(lines)
			}
			if startIdx < endIdx {
				snippet = strings.Join(lines[startIdx:endIdx], "\n")
			}
		}

		// Generate comment URL
		commentURL := ""
		if item.CommentID > 0 {
			commentURL = fmt.Sprintf("https://github.com/%s/%s/pull/%d#discussion_r%d",
				repoOwner, repoName, prNumber, item.CommentID)
		}

		// Generate the prompt file with enhanced context
		promptContent := generatePatchPrompt(
			repoOwner, repoName, prNumber,
			item.File, item.Body, snippet,
			startLine, item.DiffHunk, commentURL,
		)
		if err := os.WriteFile(outputFilePath, []byte(promptContent), 0o644); err != nil {
			return fmt.Errorf("failed to create prompt file %s: %w", outputFilePath, err)
		}

		fmt.Printf("Created: %s\n", outputFilePath)
	}

	// Create an index file
	indexFilePath := filepath.Join(outputDir, "INDEX.md")
	indexContent := generateIndexContent(repoOwner, repoName, prNumber, feedbackItems)
	if err := os.WriteFile(indexFilePath, []byte(indexContent), 0o644); err != nil {
		return fmt.Errorf("failed to create index file: %w", err)
	}

	fmt.Printf("\n✓ Created %d prompt files in: %s\n", len(feedbackItems), outputDir)
	fmt.Printf("✓ Index file created: %s\n", indexFilePath)

	return nil
}

func generatePatchPrompt(repoOwner, repoName string, prNumber int, file, comment, codeSnippet string, startLine int, diffHunk, commentURL string) string {
	repo := fmt.Sprintf("%s/%s", repoOwner, repoName)
	language := inferLanguage(file)

	// Add line numbers to code snippet
	numberedCode := addLineNumbers(codeSnippet, startLine)

	var prompt strings.Builder

	// Write metadata section
	fmt.Fprintf(&prompt, `# PR Feedback Review Task

## Metadata
- **Repository:** %s
- **Pull Request:** #%d
- **Target File:** `+"`%s`"+`
- **Reviewer:** gemini-code-assist[bot]`, repo, prNumber, file)

	if commentURL != "" {
		fmt.Fprintf(&prompt, "\n- **Feedback Link:** %s", commentURL)
	}

	// Write feedback section
	fmt.Fprintf(&prompt, `

## Reviewer Feedback

%s
`, comment)

	// Add diff context if available
	if diffHunk != "" {
		fmt.Fprintf(&prompt, `
## PR Diff (relevant changes)

> This shows what changed in the PR. Use this to understand the context of the feedback.

%sdiff
%s
%s
`, "```", diffHunk, "```")
	}

	// Add file snapshot with line numbers
	if numberedCode != "" {
		fmt.Fprintf(&prompt, `
## Current File Snapshot (starting at line %d)

> ⚠️ **WARNING:** This is a READ-ONLY snapshot for context.
> You MUST edit the actual file at: `+"`%s`"+`

%s%s
%s
%s
`, startLine, file, "```", language, numberedCode, "```")
	}

	// Write task instructions
	fmt.Fprintf(&prompt, `
## Your Task

1. **Evaluate** the feedback above against:
   - **Correctness:** Is the suggestion technically accurate?
   - **Relevance:** Does it apply to this codebase's patterns and conventions?
   - **Priority:** Is this a bug fix, security issue, style nit, or premature optimization?

2. **Decide** one of:
   - **APPLY:** Implement the suggestion (verbatim or with modifications)
   - **REJECT:** The feedback is incorrect, inapplicable, or low-value
   - **SKIP:** The target file doesn't exist or the feedback is no longer applicable

3. **Act** on your decision:
   - If APPLY: Edit the file at `+"`%s`"+` and run relevant formatters (e.g., gofmt, prettier)
   - If REJECT or SKIP: Do not modify any files

4. **Document** your decision in this format:

---
## Decision: [APPLY | REJECT | SKIP]

### Reasoning
[Your explanation of why you made this decision]

### Changes Made
[Summary of edits, or "None" if rejected]
---

**Project Conventions:** If a CONVENTIONS.md, .editorconfig, or style guide exists in the repo root, consult it before deciding.
`, file)

	return prompt.String()
}

// inferLanguage returns the language identifier for syntax highlighting based on filename
func inferLanguage(file string) string {
	base := filepath.Base(file)
	ext := filepath.Ext(file)

	// Handle extensionless files
	switch base {
	case "Dockerfile":
		return "dockerfile"
	case "Makefile":
		return "makefile"
	case "Jenkinsfile":
		return "jenkinsfile"
	case "go.mod", "go.sum":
		return "go"
	case ".editorconfig":
		return "editorconfig"
	}

	// Handle extensions
	if len(ext) > 1 {
		return ext[1:] // Remove leading dot
	}

	return "text"
}

// addLineNumbers adds line numbers to code snippets starting from startLine
func addLineNumbers(code string, startLine int) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(strings.TrimRight(code, "\n"), "\n")
	var numbered strings.Builder

	for i, line := range lines {
		if i > 0 {
			numbered.WriteString("\n")
		}
		numbered.WriteString(fmt.Sprintf("%d: %s", startLine+i, line))
	}

	return numbered.String()
}

func generateIndexContent(repoOwner, repoName string, prNumber int, feedbackItems []FeedbackItem) string {
	repo := fmt.Sprintf("%s/%s", repoOwner, repoName)
	content := fmt.Sprintf(`# Gemini Code Assist Feedback - PR #%d

**Repository:** %s  
**Total Feedback Items:** %d  
**Generated:** %s

## Feedback Files

`, prNumber, repo, len(feedbackItems), time.Now().Format("2006-01-02 15:04:05"))

	for i, item := range feedbackItems {
		if item.File != "" {
			filename := strings.ReplaceAll(strings.ReplaceAll(item.File, "/", "_"), ".", "_")
			promptFile := fmt.Sprintf("%d_%s_line%d.md", i+1, filename, item.Line)
			content += fmt.Sprintf("%d. [`%s:%d`](./%s)\n", i+1, item.File, item.Line, promptFile)
		} else {
			promptFile := fmt.Sprintf("%d_general_comment.md", i+1)
			content += fmt.Sprintf("%d. [General PR Comment](./%s)\n", i+1, promptFile)
		}
	}

	content += `
## Usage

Each file contains a complete, self-contained prompt that you can feed to an AI coding agent (Claude, GPT-4, etc.) for analysis and recommendations.

Process these files individually to get thoughtful, context-aware feedback on each suggestion.
`
	return content
}
