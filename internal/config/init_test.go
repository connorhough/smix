package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureConfigExists(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Ensure config doesn't exist
	if _, err := os.Stat(configPath); err == nil {
		t.Fatal("config should not exist yet")
	}

	// Call EnsureConfigExists
	if err := EnsureConfigExists(configPath); err != nil {
		t.Fatalf("EnsureConfigExists failed: %v", err)
	}

	// Verify config was created
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read created config: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("config file is empty")
	}

	// Verify it contains expected content
	contentStr := string(content)
	expectedStrings := []string{
		"provider: claude",
		"providers:",
		"commands:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("config missing expected content: %q", expected)
		}
	}
}

func TestEnsureConfigExistsIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	// Create initial config
	if err := EnsureConfigExists(configPath); err != nil {
		t.Fatalf("first call failed: %v", err)
	}

	// Modify the file
	customContent := []byte("# custom config\nprovider: custom")
	if err := os.WriteFile(configPath, customContent, 0644); err != nil {
		t.Fatalf("failed to write custom content: %v", err)
	}

	// Call again - should not overwrite
	if err := EnsureConfigExists(configPath); err != nil {
		t.Fatalf("second call should not error: %v", err)
	}

	// Verify file was NOT overwritten
	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config: %v", err)
	}
	if string(content) != string(customContent) {
		t.Error("EnsureConfigExists overwrote existing config")
	}
}
