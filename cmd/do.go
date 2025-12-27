package cmd

import (
	"context"
	"fmt"

	"github.com/connorhough/smix/internal/config"
	"github.com/connorhough/smix/internal/do"
	"github.com/spf13/cobra"
)

// NewDoCmd creates and returns the do command
func NewDoCmd() *cobra.Command {
	doCmd := &cobra.Command{
		Use:   "do \"natural language task description\"",
		Short: "Translate natural language to shell commands",
		Long: `Translate natural language task descriptions into executable shell commands
using your configured LLM provider.

Supports multiple providers (Claude, Gemini) with per-command configuration.`,
		Args: cobra.ExactArgs(1),
		RunE: runDo,
	}

	return doCmd
}

func runDo(cmd *cobra.Command, args []string) error {
	taskDescription := args[0]

	// Resolve configuration
	cfg := config.ResolveProviderConfig("do")
	cfg.ApplyFlags(providerFlag, modelFlag)

	debugLog("Resolved config for 'do': provider=%s, model=%s", cfg.Provider, cfg.Model)

	// Create context
	ctx := context.Background()

	// Translate
	shellCommand, err := do.Translate(ctx, taskDescription, cfg, debugLog)
	if err != nil {
		return err
	}

	// Print the resulting shell command
	fmt.Println(shellCommand)

	return nil
}
