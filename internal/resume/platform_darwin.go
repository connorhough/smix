//go:build darwin

package resume

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type darwinPlatform struct{}

func init() {
	CurrentPlatform = &darwinPlatform{}
}

func (p *darwinPlatform) GetActiveWindowTitle(ctx context.Context) (string, error) {
	// AppleScript to get the name of the frontmost window
	script := `tell application "System Events" to get name of window 1 of (first process whose frontmost is true)`
	out, err := runOsaScript(ctx, script)
	if err != nil {
		// Fallback to process name if window has no title or other error
		script = `tell application "System Events" to get name of first process whose frontmost is true`
		out2, err2 := runOsaScript(ctx, script)
		if err2 == nil {
			return strings.TrimSpace(out2), nil
		}
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (p *darwinPlatform) TypeString(ctx context.Context, msg string) error {
	// Escape string for AppleScript
	// We need to escape backslashes first, then quotes
	escaped := strings.ReplaceAll(msg, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")

	script := fmt.Sprintf(`tell application "System Events" to keystroke "%s"`, escaped)
	_, err := runOsaScript(ctx, script)
	return err
}

func (p *darwinPlatform) PressEnter(ctx context.Context) error {
	script := `tell application "System Events" to key code 36`
	_, err := runOsaScript(ctx, script)
	return err
}

func runOsaScript(ctx context.Context, script string) (string, error) {
	cmd := exec.CommandContext(ctx, "osascript", "-e", script)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("osascript failed: %w (stderr: %s)", err, string(exitErr.Stderr))
		}
		return "", err
	}
	return string(out), nil
}
