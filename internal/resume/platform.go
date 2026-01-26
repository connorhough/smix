package resume

import "context"

// Platform defines the interface for OS-specific operations required by the resume command.
type Platform interface {
	// GetActiveWindowTitle returns the title of the currently focused window.
	GetActiveWindowTitle(ctx context.Context) (string, error)

	// TypeString types the given string into the active window.
	TypeString(ctx context.Context, msg string) error

	// PressEnter simulates pressing the Enter key.
	PressEnter(ctx context.Context) error
}

// CurrentPlatform holds the implementation for the current operating system.
// It is initialized by the platform-specific files (e.g., platform_darwin.go).
var CurrentPlatform Platform
