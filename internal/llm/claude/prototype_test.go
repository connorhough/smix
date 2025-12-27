package claude

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestPrototypeCLIWrapper(t *testing.T) {
	// Skip if claude CLI not available
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude CLI not available")
	}

	ctx := context.Background()
	prompt := "Say 'hello' and nothing else"

	cmd := exec.CommandContext(ctx, "claude", "--model", "haiku", "-p", prompt)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("CLI execution failed: %v (output: %s)", err, output)
	}

	if len(output) == 0 {
		t.Fatal("Expected non-empty output")
	}

	t.Logf("CLI wrapper works! Output: %s", output)
}

// TestContextCancellation verifies if the CLI respects context cancellation
func TestContextCancellation(t *testing.T) {
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude CLI not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Use a prompt that might take longer to process
	prompt := "Write a detailed essay about the history of computing"

	cmd := exec.CommandContext(ctx, "claude", "--model", "haiku", "-p", prompt)
	_, err := cmd.CombinedOutput()

	if err == nil {
		t.Log("Command completed before timeout (expected for fast responses)")
		return
	}

	// Check if error is context-related
	if ctx.Err() == context.DeadlineExceeded {
		t.Logf("Context cancellation works! Error: %v", err)
	} else {
		t.Logf("Command failed with non-context error: %v", err)
	}
}

// TestInvalidModelName verifies error handling for invalid model names
func TestInvalidModelName(t *testing.T) {
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude CLI not available")
	}

	ctx := context.Background()
	prompt := "Say hello"

	cmd := exec.CommandContext(ctx, "claude", "--model", "invalid-model-name-xyz", "-p", prompt)
	output, err := cmd.CombinedOutput()

	if err != nil {
		exitErr, ok := err.(*exec.ExitError)
		if ok {
			t.Logf("Invalid model returns exit code: %d", exitErr.ExitCode())
			t.Logf("Error output: %s", output)
		} else {
			t.Logf("Invalid model returns error: %v (output: %s)", err, output)
		}
	} else {
		t.Logf("Invalid model name did not produce error. Output: %s", output)
	}
}

// TestOutputFormat examines the structure of CLI output
func TestOutputFormat(t *testing.T) {
	if _, err := exec.LookPath("claude"); err != nil {
		t.Skip("claude CLI not available")
	}

	ctx := context.Background()
	prompt := "Say 'hello' and nothing else"

	cmd := exec.CommandContext(ctx, "claude", "--model", "haiku", "-p", prompt)
	output, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("CLI execution failed: %v", err)
	}

	outputStr := string(output)

	t.Logf("Output length: %d bytes", len(output))
	t.Logf("Output contains newline: %v", strings.Contains(outputStr, "\n"))
	t.Logf("Output starts with whitespace: %v", len(outputStr) > 0 && (outputStr[0] == ' ' || outputStr[0] == '\n'))
	t.Logf("Output ends with newline: %v", strings.HasSuffix(outputStr, "\n"))
	t.Logf("Raw output: %q", outputStr)
}
