package gca

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
	Type string `json:"type"`
	File string `json:"file"`
	Line int    `json:"line"`
	Body string `json:"body"`
}

// FetchReviews fetches gemini-code-assist feedback from a GitHub PR
func FetchReviews(ctx context.Context, client *github.Client, repoOwner, repoName string, prNumber int, outputDir string) error {
	if outputDir == "" {
		outputDir = fmt.Sprintf("./gemini_feedback_pr%d", prNumber)
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

			feedbackItems = append(feedbackItems, FeedbackItem{
				Type: "review_comment",
				File: *comment.Path,
				Line: line,
				Body: *comment.Body,
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
		var locationInfo string

		if item.File != "" {
			// Create sanitized filename
			filename := strings.ReplaceAll(strings.ReplaceAll(item.File, "/", "_"), ".", "_")
			outputFilePath = filepath.Join(outputDir, fmt.Sprintf("%d_%s_line%d.md", i+1, filename, item.Line))
			locationInfo = fmt.Sprintf("**File:** `%s`  \n**Line:** %d", item.File, item.Line)
		} else {
			outputFilePath = filepath.Join(outputDir, fmt.Sprintf("%d_general_comment.md", i+1))
			locationInfo = "**Type:** General PR comment"
		}

		// Generate the prompt file
		promptContent := generatePromptContent(repoOwner, repoName, prNumber, locationInfo, item.Body)
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

func generatePromptContent(repoOwner, repoName string, prNumber int, locationInfo, body string) string {
	repo := fmt.Sprintf("%s/%s", repoOwner, repoName)
	return fmt.Sprintf(`# Code Review Feedback Analysis

You are an expert software engineer reviewing code feedback from an AI code review tool (gemini-code-assist). Your task is to critically evaluate this feedback and determine the best course of action.

## Context

**Repository:** %s  
**Pull Request:** #%d  
%s

## Your Task

Analyze the feedback below and provide a structured response that includes:

1. **Critical Analysis**: 
   - Does the feedback identify a real issue?
   - Is the suggested solution the best approach?
   - Are there alternative solutions that might be better?
   - Does this align with the project's architecture and patterns?

2. **Decision**: Choose one of:
   - **ACCEPT**: Agree with the feedback and implement the suggested change (or your improved version)
   - **MODIFY**: Agree with the concern but propose a different/better solution
   - **REJECT**: Explain why the feedback doesn't apply or why the current implementation is correct

3. **Proposed Changes**: If accepting or modifying, provide:
   - The exact code changes to make
   - File path and line numbers
   - Any additional context needed

## Feedback from gemini-code-assist

severity:
%s

---
## Your Response Format

Please structure your response as follows:

%s
### Decision: [ACCEPT | MODIFY | REJECT]

### Reasoning
[Your critical analysis - what's the real issue? Is the suggestion optimal? What are the trade-offs?]

### Proposed Solution
[If ACCEPT or MODIFY: exact code changes, file paths, line numbers]
[If REJECT: explanation of why no changes are needed]

### Alternative Approaches Considered
[Other solutions you evaluated and why you chose (or didn't choose) them]

### Implementation Notes
[Any additional context, testing considerations, or follow-up items]
%s

---
**Important**: Your goal is to make the best technical decision for the codebase, not to blindly apply suggestions. Consider the broader context, existing patterns, and long-term maintainability.
`, repo, prNumber, locationInfo, body, "```", "```")
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
