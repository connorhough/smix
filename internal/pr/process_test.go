package pr

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
- **Target File:** ` + "`internal/pr/fetch.go`" + `
- **Reviewer:** gemini-code-assist[bot]

## Reviewer Feedback

Some feedback here.
`

	err := os.WriteFile(feedbackFile, []byte(content), 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	got := extractTargetFile(feedbackFile)
	want := "internal/pr/fetch.go"

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
