package cmd

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/connorhough/smix/internal/resume"
)

// Re-implement mocks for cmd package tests since internal test files aren't exported.

type MockPlatform struct {
	WindowTitle  string
	Typed        []string
	EnterPressed bool
}

func (m *MockPlatform) GetActiveWindowTitle(ctx context.Context) (string, error) {
	return m.WindowTitle, nil
}

func (m *MockPlatform) TypeString(ctx context.Context, msg string) error {
	m.Typed = append(m.Typed, msg)
	return nil
}

func (m *MockPlatform) PressEnter(ctx context.Context) error {
	m.EnterPressed = true
	return nil
}

type MockClock struct {
	CurrentTime time.Time
}

func (m *MockClock) Now() time.Time {
	return m.CurrentTime
}

func (m *MockClock) Sleep(ctx context.Context, d time.Duration) error {
	return nil
}

func TestResumeCmd_Validation(t *testing.T) {
	cmd := newResumeCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))

	// Case 1: Missing --at
	err := cmd.Execute()
	if err == nil {
		t.Error("Expected error for missing --at flag, got nil")
	} else if err.Error() != "--at flag is required (e.g., '14:30', '2:30PM')" {
		t.Errorf("Unexpected error: %v", err)
	}

	// Case 2: Invalid --at format
	cmd = newResumeCmd()
	cmd.SetOut(new(bytes.Buffer))
	cmd.SetErr(new(bytes.Buffer))
	cmd.SetArgs([]string{"--at", "invalid-time"})

	err = cmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid time format, got nil")
	} else if err.Error() == "" {
		t.Error("Expected error message, got empty")
	}
}

func TestResumeCmd_Success(t *testing.T) {
	// Setup Mocks
	mockPlat := &MockPlatform{WindowTitle: "Terminal"}
	resume.CurrentPlatform = mockPlat
	defer func() { resume.CurrentPlatform = nil }()

	start := time.Date(2023, 1, 1, 10, 0, 0, 0, time.Local)
	mockClock := &MockClock{CurrentTime: start}
	originalClock := resume.SystemClock
	resume.SystemClock = mockClock
	defer func() { resume.SystemClock = originalClock }()

	// Execute
	cmd := newResumeCmd()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	// We want to resume at 10:05.
	// 10:05 is 5 minutes from 10:00.
	cmd.SetArgs([]string{"--at", "10:05", "--message", "hello"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}

	// Verify Platform
	if len(mockPlat.Typed) != 1 {
		t.Errorf("Expected 1 typed message, got %d", len(mockPlat.Typed))
	} else if mockPlat.Typed[0] != "hello" {
		t.Errorf("Expected message 'hello', got %q", mockPlat.Typed[0])
	}
}

func TestResumeCmd_Formats(t *testing.T) {
	// Test various time formats
	formats := []string{"14:30", "2:30PM", "2:30 pm", "14:30:05"}

	mockPlat := &MockPlatform{WindowTitle: "Terminal"}
	resume.CurrentPlatform = mockPlat
	defer func() { resume.CurrentPlatform = nil }()

	mockClock := &MockClock{CurrentTime: time.Date(2023, 1, 1, 10, 0, 0, 0, time.Local)}
	originalClock := resume.SystemClock
	resume.SystemClock = mockClock
	defer func() { resume.SystemClock = originalClock }()

	for _, f := range formats {
		cmd := newResumeCmd()
		cmd.SetOut(new(bytes.Buffer))
		cmd.SetErr(new(bytes.Buffer))
		cmd.SetArgs([]string{"--at", f})

		if err := cmd.Execute(); err != nil {
			t.Errorf("Format %q failed: %v", f, err)
		}
	}
}
