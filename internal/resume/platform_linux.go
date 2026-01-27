//go:build linux

package resume

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type linuxPlatform struct{}

func init() {
	CurrentPlatform = &linuxPlatform{}
}

func (p *linuxPlatform) GetActiveWindowTitle(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "xdotool", "getwindowfocus", "getwindowname")
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("xdotool getwindowname failed: %w (stderr: %s)", err, string(exitErr.Stderr))
		}
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (p *linuxPlatform) TypeString(ctx context.Context, msg string) error {
	// --delay 50 sets a 50ms delay between keystrokes to ensure reliability
	cmd := exec.CommandContext(ctx, "xdotool", "type", "--delay", "50", msg)
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("xdotool type failed: %w (stderr: %s)", err, string(exitErr.Stderr))
		}
		return err
	}
	return nil
}

func (p *linuxPlatform) PressEnter(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "xdotool", "key", "Return")
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("xdotool key failed: %w (stderr: %s)", err, string(exitErr.Stderr))
		}
		return err
	}
	return nil
}
