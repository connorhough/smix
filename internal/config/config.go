// Package config provides configuration management functionality for the smix application.
package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// GetValue retrieves a configuration value by key
func GetValue(key string) (string, error) {
	if !viper.IsSet(key) {
		return "", fmt.Errorf("key '%s' not found in configuration", key)
	}
	return viper.GetString(key), nil
}

// SetValue sets a configuration value by key and persists it to the config file
func SetValue(key string, value string) error {
	viper.Set(key, value)
	return viper.WriteConfig()
}

// ProviderConfig holds provider and model configuration
type ProviderConfig struct {
	Provider string
	Model    string
}

// ResolveProviderConfig resolves provider configuration for a command
// Precedence: command-specific config -> global config
// Flags are handled separately in command layer
func ResolveProviderConfig(commandName string) *ProviderConfig {
	cfg := &ProviderConfig{}

	// Try command-specific provider
	commandProviderKey := fmt.Sprintf("commands.%s.provider", commandName)
	if viper.IsSet(commandProviderKey) {
		cfg.Provider = viper.GetString(commandProviderKey)
	} else {
		// Fall back to global provider
		cfg.Provider = viper.GetString("provider")
	}

	// Try command-specific model
	commandModelKey := fmt.Sprintf("commands.%s.model", commandName)
	if viper.IsSet(commandModelKey) {
		cfg.Model = viper.GetString(commandModelKey)
	} else {
		// Fall back to global model
		cfg.Model = viper.GetString("model")
	}

	return cfg
}

// ApplyFlags applies flag overrides to config (called from command layer)
func (c *ProviderConfig) ApplyFlags(providerFlag, modelFlag string) {
	if providerFlag != "" {
		c.Provider = providerFlag
	}
	if modelFlag != "" {
		c.Model = modelFlag
	}
}
