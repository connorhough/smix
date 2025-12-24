package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureConfigExists creates a config file with template if it doesn't exist
func EnsureConfigExists(configPath string) error {
	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // Config exists, nothing to do
	}

	// Create config directory if needed
	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write template config
	if err := os.WriteFile(configPath, []byte(configTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write config template: %w", err)
	}

	return nil
}
