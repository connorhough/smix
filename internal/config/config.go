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