// Package cmd provides the command-line interface for the smix application.
package cmd

import (
	"context"
	"fmt"
	"log/slog"
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
func Execute(ctx context.Context) error {
	if rootCmd == nil {
		rootCmd = NewRootCmd()
	}
	return rootCmd.ExecuteContext(ctx)
}

// NewRootCmd creates and returns the root command for smix
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "smix",
		Short:         "Hi I'm smix",
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
	rootCmd.AddCommand(newResumeCmd())

	// PersistentPreRun handles configuration initialization
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := initConfig(); err != nil {
			return err
		}

		setupLogging()
		return nil
	}

	return rootCmd
}

// initConfig reads in config file and ENV variables if set.
func initConfig() error {
	// Determine home directory for default paths
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig == "" {
		xdgConfig = filepath.Join(home, ".config")
	}

	// Determine config file path
	var configPath string
	if cfgFile != "" {
		// Use config file from the flag
		configPath = cfgFile
		viper.SetConfigFile(cfgFile)
	} else {
		// Check for existing config files in order of preference
		xdgPath := filepath.Join(xdgConfig, "smix", "config.yaml")
		dotPath := filepath.Join(home, ".smix.yaml")

		if _, err := os.Stat(xdgPath); err == nil {
			configPath = xdgPath
			viper.SetConfigFile(xdgPath)
		} else if _, err := os.Stat(dotPath); err == nil {
			configPath = dotPath
			viper.SetConfigFile(dotPath)
		} else {
			// Default to XDG path for new config creation
			configPath = xdgPath
			viper.SetConfigFile(xdgPath)
		}
	}

	// Ensure config file exists (create from template if needed)
	if err := config.EnsureConfigExists(configPath); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	slog.Debug("found config", "path", configPath)

	// Read in environment variables that match
	viper.SetEnvPrefix("SMIX")
	viper.AutomaticEnv()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to read config: %w", err)
	}

	slog.Debug("Config loaded successfully")

	return nil
}

func setupLogging() {
	level := slog.LevelInfo

	if debugFlag || viper.GetString("log_level") == "debug" {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})

	slog.SetDefault(slog.New(handler))
}
