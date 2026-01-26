//go:build linux

package resume

import (
	"context"
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
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func (p *linuxPlatform) TypeString(ctx context.Context, msg string) error {
	// --delay 50 sets a 50ms delay between keystrokes to ensure reliability
	cmd := exec.CommandContext(ctx, "xdotool", "type", "--delay", "50", msg)
	return cmd.Run()
}

func (p *linuxPlatform) PressEnter(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "xdotool", "key", "Return")
	return cmd.Run()
}
