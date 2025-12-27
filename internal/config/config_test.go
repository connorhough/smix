package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

func TestApplyFlags(t *testing.T) {
	tests := []struct {
		name         string
		initial      *ProviderConfig
		providerFlag string
		modelFlag    string
		wantProvider string
		wantModel    string
	}{
		{
			name:         "both flags override config",
			initial:      &ProviderConfig{Provider: "claude", Model: "sonnet"},
			providerFlag: "gemini",
			modelFlag:    "gemini-1.5-flash",
			wantProvider: "gemini",
			wantModel:    "gemini-1.5-flash",
		},
		{
			name:         "only provider flag overrides",
			initial:      &ProviderConfig{Provider: "claude", Model: "sonnet"},
			providerFlag: "gemini",
			modelFlag:    "",
			wantProvider: "gemini",
			wantModel:    "sonnet",
		},
		{
			name:         "only model flag overrides",
			initial:      &ProviderConfig{Provider: "claude", Model: "sonnet"},
			providerFlag: "",
			modelFlag:    "opus",
			wantProvider: "claude",
			wantModel:    "opus",
		},
		{
			name:         "empty flags don't override",
			initial:      &ProviderConfig{Provider: "claude", Model: "sonnet"},
			providerFlag: "",
			modelFlag:    "",
			wantProvider: "claude",
			wantModel:    "sonnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.ApplyFlags(tt.providerFlag, tt.modelFlag)
			if tt.initial.Provider != tt.wantProvider {
				t.Errorf("provider: got %q, want %q", tt.initial.Provider, tt.wantProvider)
			}
			if tt.initial.Model != tt.wantModel {
				t.Errorf("model: got %q, want %q", tt.initial.Model, tt.wantModel)
			}
		})
	}
}

func TestResolveProviderConfig(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
provider: claude
model: sonnet

commands:
  ask:
    provider: gemini
    model: gemini-1.5-flash
  do:
    provider: gemini
`

	if err := os.WriteFile(configFile, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	// Initialize viper with test config
	viper.Reset()
	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("failed to read config: %v", err)
	}

	tests := []struct {
		name         string
		command      string
		wantProvider string
		wantModel    string
	}{
		{
			name:         "ask command uses command-specific config",
			command:      "ask",
			wantProvider: "gemini",
			wantModel:    "gemini-1.5-flash",
		},
		{
			name:         "do command uses command-specific provider, inherits global model",
			command:      "do",
			wantProvider: "gemini",
			wantModel:    "sonnet", // Falls back to global
		},
		{
			name:         "pr command uses global defaults",
			command:      "pr",
			wantProvider: "claude",
			wantModel:    "sonnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := ResolveProviderConfig(tt.command)
			if cfg.Provider != tt.wantProvider {
				t.Errorf("provider: got %q, want %q", cfg.Provider, tt.wantProvider)
			}
			if cfg.Model != tt.wantModel {
				t.Errorf("model: got %q, want %q", cfg.Model, tt.wantModel)
			}
		})
	}
}
