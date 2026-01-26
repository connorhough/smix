package resume

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"
)

// Run executes the resume wait and type logic.
func Run(ctx context.Context, targetTime time.Time, msg string, windowFilter string) error {
	if CurrentPlatform == nil {
		return fmt.Errorf("resume command is not supported on this platform")
	}

	now := SystemClock.Now()
	// If targetTime is in the past, we assume the user meant the next occurrence (tomorrow).
	if targetTime.Before(now) {
		targetTime = targetTime.Add(24 * time.Hour)
	}

	waitDuration := targetTime.Sub(now)
	// Using 15:04:05 for clear 24h time display
	fmt.Printf("Waiting until %s to resume...\n", targetTime.Format("15:04:05"))

	if waitDuration > 10*time.Second {
		// Sleep until 10s before
		if err := SystemClock.Sleep(ctx, waitDuration-10*time.Second); err != nil {
			return err
		}
		fmt.Println("Resuming in 10 seconds...")
		if err := SystemClock.Sleep(ctx, 10*time.Second); err != nil {
			return err
		}
	} else {
		if err := SystemClock.Sleep(ctx, waitDuration); err != nil {
			return err
		}
	}

	// Safety Check: Window Filter
	if windowFilter != "" {
		title, err := CurrentPlatform.GetActiveWindowTitle(ctx)
		if err != nil {
			return fmt.Errorf("failed to get active window title: %w", err)
		}

		slog.Debug("Checking active window", "title", title, "filter", windowFilter)
		// Case-insensitive check
		if !strings.Contains(strings.ToLower(title), strings.ToLower(windowFilter)) {
			return fmt.Errorf("active window '%s' does not match filter '%s'", title, windowFilter)
		}
	}

	slog.Info("Typing resume message")
	if err := CurrentPlatform.TypeString(ctx, msg); err != nil {
		return fmt.Errorf("failed to type message: %w", err)
	}

	if err := CurrentPlatform.PressEnter(ctx); err != nil {
		return fmt.Errorf("failed to press enter: %w", err)
	}

	return nil
}
