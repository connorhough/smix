// Package cmd provides the command-line interface for the smix application.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/connorhough/smix/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	rootCmd *cobra.Command
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.go. It only needs to happen once to the rootCmd.
func Execute() error {
	if rootCmd == nil {
		rootCmd = NewRootCmd()
	}
	return rootCmd.Execute()
}

// NewRootCmd creates and returns the root command for smix
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "smix",
		Short:         "I'm smix",
		Long:          `Hi I'm smix!`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version.String(),
	}

	// Add persistent flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default locations: $XDG_CONFIG_HOME/smix/config.yaml, ~/.config/smix/config.yaml, or ~/.smix.yaml)")

	// Add subcommands
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newExampleCmd())

	// PersistentPreRun handles configuration initialization
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return initConfig()
	}

	return rootCmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find config file in standard locations
		if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
			viper.AddConfigPath(filepath.Join(xdgConfigHome, "smix"))
		} else {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get user home directory: %w", err)
			}
			viper.AddConfigPath(filepath.Join(home, ".config", "smix"))
			viper.AddConfigPath(home)
		}
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("SMIX")
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		// Config file not found; ignore error if desired
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	return nil
}
