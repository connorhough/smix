// Package cmd provides the command-line interface for the smix application.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/connorhough/smix/internal/config"
	"github.com/connorhough/smix/internal/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile      string
	rootCmd      *cobra.Command
	debugFlag    bool
	providerFlag string
	modelFlag    string
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
	rootCmd.PersistentFlags().BoolVar(&debugFlag, "debug", false, "Enable debug output")
	rootCmd.PersistentFlags().StringVar(&providerFlag, "provider", "", "Override LLM provider (claude, gemini)")
	rootCmd.PersistentFlags().StringVar(&modelFlag, "model", "", "Override model name")

	// Add subcommands
	rootCmd.AddCommand(newConfigCmd())
	rootCmd.AddCommand(newPRCmd())
	rootCmd.AddCommand(NewDoCmd())
	rootCmd.AddCommand(NewAskCmd())

	// PersistentPreRun handles configuration initialization
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		return initConfig()
	}

	return rootCmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	// Determine config file path
	var configPath string
	if cfgFile != "" {
		// Use config file from the flag
		configPath = cfgFile
		viper.SetConfigFile(cfgFile)
	} else {
		// Determine default config path
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get user home directory: %w", err)
		}

		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig == "" {
			xdgConfig = filepath.Join(home, ".config")
		}

		configPath = filepath.Join(xdgConfig, "smix", "config.yaml")

		// Set up viper search paths
		viper.AddConfigPath(filepath.Join(xdgConfig, "smix"))
		viper.AddConfigPath(filepath.Join(home, ".config", "smix"))
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Ensure config file exists (create from template if needed)
	if err := config.EnsureConfigExists(configPath); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	debugLog("Using config file: %s", configPath)

	// Read in environment variables that match
	viper.SetEnvPrefix("SMIX")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	debugLog("Config loaded successfully")

	return nil
}

// debugLog prints debug information if debug flag is enabled
func debugLog(format string, args ...interface{}) {
	if debugFlag {
		fmt.Fprintf(os.Stderr, "[DEBUG] "+format+"\n", args...)
	}
}
