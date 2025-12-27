# GCA Prompt Improvements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Enhance the `gca review` command prompts to provide clearer instructions, better context, and more structured decision-making for Claude Code sessions.

**Architecture:** Improve prompt generation in `internal/gca/fetch.go` and `internal/gca/process.go` to include git diff context, line numbers, explicit file paths, structured output format, and better autonomy constraints. Add robust language detection and edge case handling.

**Tech Stack:** Go, GitHub API (go-github), git diff parsing

---

## Context

The current `gca review` implementation has several issues:
- Ambiguous file references in prompts
- Missing line numbers in code snippets
- No git diff context from the PR
- Vague evaluation instructions
- Generic persona without clear decision framework
- No explicit target file for modifications
- No structured output format
- Naive language detection

## Task 1: Improve Language Detection

**Files:**
- Modify: `internal/gca/fetch.go:149-179`

**Step 1: Write test for improved language detection**

Create: `internal/gca/fetch_test.go`

```go
package gca

import "testing"

func TestInferLanguage(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"main.go", "go"},
		{"app.tsx", "tsx"},
		{"Dockerfile", "dockerfile"},
		{"Makefile", "makefile"},
		{"Jenkinsfile", "jenkinsfile"},
		{"go.mod", "go"},
		{"go.sum", "go"},
		{".editorconfig", "editorconfig"},
		{"script.sh", "sh"},
		{"unknown", "text"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := inferLanguage(tt.filename)
			if got != tt.want {
				t.Errorf("inferLanguage(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/connor/Src/smix && go test ./internal/gca -run TestInferLanguage -v`
Expected: FAIL with "undefined: inferLanguage"

**Step 3: Implement inferLanguage function**

Add to `internal/gca/fetch.go` (after line 179, before `generateIndexContent`):

```go
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
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/connor/Src/smix && go test ./internal/gca -run TestInferLanguage -v`
Expected: PASS

**Step 5: Update generatePatchPrompt to use inferLanguage**

In `internal/gca/fetch.go`, replace lines 152-157 with:

```go
// Determine file extension for syntax highlighting
language := inferLanguage(file)
```

**Step 6: Run tests to verify still passing**

Run: `cd /Users/connor/Src/smix && go test ./internal/gca -v`
Expected: All PASS

**Step 7: Commit**

```bash
git add internal/gca/fetch.go internal/gca/fetch_test.go
git commit -m "feat(gca): improve language detection for syntax highlighting"
```

---

## Task 2: Add Diff Context Support

**Files:**
- Modify: `internal/gca/fetch.go:15-21` (FeedbackItem struct)
- Modify: `internal/gca/fetch.go:23-147` (FetchReviews function)

**Step 1: Extend FeedbackItem struct**

Update the `FeedbackItem` struct in `internal/gca/fetch.go` (lines 15-21):

```go
// FeedbackItem represents a single feedback item from gemini-code-assist
type FeedbackItem struct {
	Type      string `json:"type"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Body      string `json:"body"`
	DiffHunk  string `json:"diff_hunk"`  // New field
	CommentID int64  `json:"comment_id"` // New field
}
```

**Step 2: Fetch PR files with diff hunks**

In `internal/gca/fetch.go`, after line 39 (after PR fetch), add:

```go
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
```

**Step 3: Populate DiffHunk and CommentID in review comments**

In `internal/gca/fetch.go`, update the review comments loop (lines 58-75):

```go
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
```

**Step 4: Build and test manually**

Run: `cd /Users/connor/Src/smix && make build`
Expected: Build successful

**Step 5: Commit**

```bash
git add internal/gca/fetch.go
git commit -m "feat(gca): add diff hunk and comment ID to feedback items"
```

---

## Task 3: Add Line Numbers to Code Snippets

**Files:**
- Modify: `internal/gca/fetch.go:102-134` (feedback file generation)
- Modify: `internal/gca/fetch.go:149-179` (generatePatchPrompt signature)

**Step 1: Write test for line numbering**

Add to `internal/gca/fetch_test.go`:

```go
func TestAddLineNumbers(t *testing.T) {
	input := `func main() {
	fmt.Println("Hello")
}`
	want := `1: func main() {
2: 	fmt.Println("Hello")
3: }`

	got := addLineNumbers(input, 1)
	if got != want {
		t.Errorf("addLineNumbers() =\n%s\nwant:\n%s", got, want)
	}
}

func TestAddLineNumbersWithOffset(t *testing.T) {
	input := `fmt.Println("Hello")
return nil`
	want := `42: fmt.Println("Hello")
43: return nil`

	got := addLineNumbers(input, 42)
	if got != want {
		t.Errorf("addLineNumbers() =\n%s\nwant:\n%s", got, want)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/connor/Src/smix && go test ./internal/gca -run TestAddLineNumbers -v`
Expected: FAIL with "undefined: addLineNumbers"

**Step 3: Implement addLineNumbers function**

Add to `internal/gca/fetch.go` (after `inferLanguage`):

```go
// addLineNumbers adds line numbers to code snippets starting from startLine
func addLineNumbers(code string, startLine int) string {
	if code == "" {
		return ""
	}

	lines := strings.Split(code, "\n")
	var numbered strings.Builder

	for i, line := range lines {
		if i > 0 {
			numbered.WriteString("\n")
		}
		numbered.WriteString(fmt.Sprintf("%d: %s", startLine+i, line))
	}

	return numbered.String()
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/connor/Src/smix && go test ./internal/gca -run TestAddLineNumbers -v`
Expected: PASS

**Step 5: Update generatePatchPrompt signature and usage**

Change `generatePatchPrompt` signature to include `startLine`, `diffHunk`, and `commentURL`:

```go
func generatePatchPrompt(repoOwner, repoName string, prNumber int, file, comment, codeSnippet string, startLine int, diffHunk, commentURL string) string {
```

**Step 6: Update generatePatchPrompt implementation**

Replace the entire `generatePatchPrompt` function (lines 149-179):

```go
func generatePatchPrompt(repoOwner, repoName string, prNumber int, file, comment, codeSnippet string, startLine int, diffHunk, commentURL string) string {
	repo := fmt.Sprintf("%s/%s", repoOwner, repoName)
	language := inferLanguage(file)

	// Add line numbers to code snippet
	numberedCode := addLineNumbers(codeSnippet, startLine)

	prompt := fmt.Sprintf(`# PR Feedback Review Task

## Metadata
- **Repository:** %s
- **Pull Request:** #%d
- **Target File:** `+"`%s`"+`
- **Reviewer:** gemini-code-assist[bot]`, repo, prNumber, file)

	if commentURL != "" {
		prompt += fmt.Sprintf("\n- **Feedback Link:** %s", commentURL)
	}

	prompt += fmt.Sprintf(`

## Reviewer Feedback

%s
`, comment)

	// Add diff context if available
	if diffHunk != "" {
		prompt += fmt.Sprintf(`
## PR Diff (relevant changes)

> This shows what changed in the PR. Use this to understand the context of the feedback.

%sdiff
%s
%s
`, "```", diffHunk, "```")
	}

	// Add file snapshot with line numbers
	if numberedCode != "" {
		prompt += fmt.Sprintf(`
## Current File Snapshot (starting at line %d)

> ⚠️ **WARNING:** This is a READ-ONLY snapshot for context.
> You MUST edit the actual file at: `+"`%s`"+`

%s%s
%s
%s
`, startLine, file, "```", language, numberedCode, "```")
	}

	prompt += `
## Your Task

1. **Evaluate** the feedback above against:
   - **Correctness:** Is the suggestion technically accurate?
   - **Relevance:** Does it apply to this codebase's patterns and conventions?
   - **Priority:** Is this a bug fix, security issue, style nit, or premature optimization?

2. **Decide** one of:
   - **APPLY:** Implement the suggestion (verbatim or with modifications)
   - **REJECT:** The feedback is incorrect, inapplicable, or low-value

3. **Act** on your decision:
   - If APPLY: Edit the file at `+"`%s`"+`
   - If REJECT: Do not modify any files

4. **Document** your decision in this format:

---
## Decision: [APPLY | REJECT]

### Reasoning
[Your explanation of why you made this decision]

### Changes Made
[Summary of edits, or "None" if rejected]
---

**Project Conventions:** If a CONVENTIONS.md, .editorconfig, or style guide exists in the repo root, consult it before deciding.
`

	return fmt.Sprintf(prompt, file)
}
```

**Step 7: Update the caller in FetchReviews**

In `internal/gca/fetch.go`, update the call to `generatePatchPrompt` (around line 128):

```go
// Determine start line (default to 1 if not available)
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

// Generate comment URL
commentURL := ""
if item.CommentID > 0 {
	commentURL = fmt.Sprintf("https://github.com/%s/%s/pull/%d#discussion_r%d",
		repoOwner, repoName, prNumber, item.CommentID)
}

// Generate the prompt file with enhanced context
promptContent := generatePatchPrompt(
	repoOwner, repoName, prNumber,
	item.File, item.Body, fileContent,
	startLine, item.DiffHunk, commentURL,
)
```

**Step 8: Run tests**

Run: `cd /Users/connor/Src/smix && go test ./internal/gca -v`
Expected: All PASS

**Step 9: Commit**

```bash
git add internal/gca/fetch.go internal/gca/fetch_test.go
git commit -m "feat(gca): add line numbers, diff context, and comment links to prompts"
```

---

## Task 4: Improve LaunchClaudeCode with Target File

**Files:**
- Modify: `internal/gca/process.go:11-60` (ProcessReviews)
- Modify: `internal/gca/process.go:62-83` (LaunchClaudeCode)

**Step 1: Update LaunchClaudeCode signature**

Change `LaunchClaudeCode` to accept `targetFile` and `totalCount`/`currentIndex`:

```go
// LaunchClaudeCode opens Claude Code with a prompt to review the feedback and implement changes
func LaunchClaudeCode(feedbackFile, targetFile string, currentIndex, totalCount int) error {
```

**Step 2: Enhance the prompt with explicit instructions**

Replace the `LaunchClaudeCode` implementation (lines 62-83):

```go
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
```

**Step 3: Update ProcessReviews to pass target file**

The challenge here is that we need to extract the target file from the feedback filename. Update `ProcessReviews`:

```go
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
```

**Step 4: Implement extractTargetFile helper**

Add after `LaunchClaudeCode`:

```go
// extractTargetFile extracts the target file path from a feedback markdown file
func extractTargetFile(feedbackFile string) string {
	content, err := os.ReadFile(feedbackFile)
	if err != nil {
		return ""
	}

	// Look for "**Target File:** `path/to/file`" in the markdown
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "**Target File:**") {
			// Extract content between backticks
			start := strings.Index(line, "`")
			if start == -1 {
				continue
			}
			end := strings.Index(line[start+1:], "`")
			if end == -1 {
				continue
			}
			return line[start+1 : start+1+end]
		}
	}

	return ""
}
```

**Step 5: Write test for extractTargetFile**

Create: `internal/gca/process_test.go`

```go
package gca

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractTargetFile(t *testing.T) {
	// Create temporary file with feedback content
	tmpDir := t.TempDir()
	feedbackFile := filepath.Join(tmpDir, "test_feedback.md")

	content := `# PR Feedback Review Task

## Metadata
- **Repository:** owner/repo
- **Pull Request:** #123
- **Target File:** ` + "`internal/gca/fetch.go`" + `
- **Reviewer:** gemini-code-assist[bot]

## Reviewer Feedback

Some feedback here.
`

	err := os.WriteFile(feedbackFile, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	got := extractTargetFile(feedbackFile)
	want := "internal/gca/fetch.go"

	if got != want {
		t.Errorf("extractTargetFile() = %q, want %q", got, want)
	}
}

func TestExtractTargetFileNotFound(t *testing.T) {
	// Create temporary file without target file
	tmpDir := t.TempDir()
	feedbackFile := filepath.Join(tmpDir, "test_feedback.md")

	content := `# Some content without target file`

	err := os.WriteFile(feedbackFile, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	got := extractTargetFile(feedbackFile)
	want := ""

	if got != want {
		t.Errorf("extractTargetFile() = %q, want %q", got, want)
	}
}
```

**Step 6: Run tests**

Run: `cd /Users/connor/Src/smix && go test ./internal/gca -v`
Expected: All PASS

**Step 7: Build and test manually**

Run: `cd /Users/connor/Src/smix && make build`
Expected: Build successful

**Step 8: Commit**

```bash
git add internal/gca/process.go internal/gca/process_test.go
git commit -m "feat(gca): improve Claude Code prompt with explicit target file and constraints"
```

---

## Task 5: Integration Testing

**Files:**
- None (manual testing)

**Step 1: Test with a real GitHub PR**

If you have access to a PR with gemini-code-assist comments, test the full workflow:

```bash
cd /Users/connor/Src/smix
make build
./builds/smix gca review owner/repo PR_NUMBER
```

**Expected behavior:**
- Fetches PR and diff context
- Creates feedback files with:
  - Line-numbered code snippets
  - Diff hunks
  - Comment permalinks
  - Clear structure
- Launches Claude Code with:
  - Explicit target file
  - Clear constraints
  - Batch awareness (X of Y)

**Step 2: Test with --dir flag on existing directory**

```bash
./builds/smix gca review --dir gca_review_pr123
```

**Expected:**
- Processes existing feedback files
- Shows proper batch numbering

**Step 3: Verify language detection**

Check generated feedback files to ensure proper syntax highlighting for:
- `.go` files → `go`
- `Dockerfile` → `dockerfile`
- `Makefile` → `makefile`

**Step 4: Document any issues found**

If issues are discovered, create follow-up tasks or fix immediately.

---

## Task 6: Update Documentation

**Files:**
- Modify: `CLAUDE.md`

**Step 1: Update gca command documentation**

In `CLAUDE.md`, update the `gca (Gemini Code Assist)` section (around line 66):

```markdown
### gca (Gemini Code Assist)

Fetches code review feedback from gemini-code-assist bot on GitHub PRs and launches interactive Claude Code sessions to analyze and implement the suggested changes.

```bash
smix gca review owner/repo pr_number
smix gca review --dir gca_review_pr123  # Process existing feedback directory
```

**Requirements:**
- `GITHUB_TOKEN` env var (optional, increases rate limits)
- `claude` CLI installed (Claude Code)

**Workflow:**
1. Fetches review comments and diff context from GitHub PR
2. Filters for gemini-code-assist bot comments
3. Creates individual prompt files with:
   - Line-numbered code snapshots
   - Git diff hunks showing PR changes
   - Direct links to comment threads
   - Structured decision format (APPLY/REJECT)
4. For each feedback item:
   - Launches Claude Code session with explicit target file and constraints
   - Claude evaluates feedback against codebase patterns and correctness
   - Implements changes with clear reasoning and documentation
   - Provides batch progress (item X of Y)

**Prompt Features:**
- Explicit target file paths for modifications
- Line-numbered code snippets for precise navigation
- PR diff context to understand what changed
- Structured output format (Decision/Reasoning/Changes)
- Clear autonomy constraints (no tests, no commits)
- Project conventions awareness (.editorconfig, CONVENTIONS.md)
```

**Step 2: Commit**

```bash
git add CLAUDE.md
git commit -m "docs: update gca command documentation with new features"
```

---

## Task 7: Final Verification

**Files:**
- None (verification)

**Step 1: Run all tests**

```bash
cd /Users/connor/Src/smix
go test ./...
```

Expected: All PASS

**Step 2: Run linter**

```bash
cd /Users/connor/Src/smix
go vet ./...
```

Expected: No issues

**Step 3: Build all targets**

```bash
cd /Users/connor/Src/smix
make clean
make build
make build-darwin-arm64
make build-linux-amd64
```

Expected: All builds successful

**Step 4: Review all commits**

```bash
git log --oneline
```

Expected commits:
1. feat(gca): improve language detection for syntax highlighting
2. feat(gca): add diff hunk and comment ID to feedback items
3. feat(gca): add line numbers, diff context, and comment links to prompts
4. feat(gca): improve Claude Code prompt with explicit target file and constraints
5. docs: update gca command documentation with new features

**Step 5: Final commit if needed**

If any cleanup or final adjustments were made:

```bash
git add .
git commit -m "chore: final cleanup and verification"
```

---

## Summary

This plan implements all the key improvements from the feedback:

### generatePatchPrompt improvements:
✅ Explicit file path with warning that snapshot is READ-ONLY
✅ Line numbers in code snippets
✅ Git diff context from PR
✅ Structured decision format (APPLY/REJECT with reasoning)
✅ Robust language detection (Dockerfile, Makefile, etc.)
✅ Comment permalinks for reference
✅ Project conventions awareness

### LaunchClaudeCode improvements:
✅ Specific persona ("autonomous code review agent")
✅ Explicit target file parameter
✅ Clear execution protocol
✅ Structured output format
✅ Edge case handling (missing files → SKIP)
✅ Autonomy constraints (no tests, no commits, single file)
✅ Batch awareness (item X of Y)

### Testing:
✅ Unit tests for language detection
✅ Unit tests for line numbering
✅ Unit tests for target file extraction
✅ Integration testing guidance
✅ Verification steps

The implementation follows TDD principles with bite-sized steps, frequent commits, and clear acceptance criteria for each task.
