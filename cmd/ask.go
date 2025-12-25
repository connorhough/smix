package cmd

import (
	"context"
	"fmt"

	"github.com/connorhough/smix/internal/ask"
	"github.com/connorhough/smix/internal/config"
	"github.com/spf13/cobra"
)

// NewAskCmd creates and returns the ask command
func NewAskCmd() *cobra.Command {
	askCmd := &cobra.Command{
		Use:   "ask \"your question\"",
		Short: "Ask short technical questions and get concise answers",
		Long: `Ask short technical questions and get concise answers using your configured LLM provider.

Supports multiple providers (Claude, Gemini) with per-command configuration.

Great for quick lookups like:
- "what is FastAPI"
- "does the mv command overwrite duplicate files"
- "how do I check if a port is open"`,
		Args: cobra.ExactArgs(1),
		RunE: runAsk,
	}

	return askCmd
}

func runAsk(cmd *cobra.Command, args []string) error {
	question := args[0]

	// Resolve configuration
	cfg := config.ResolveProviderConfig("ask")
	cfg.ApplyFlags(providerFlag, modelFlag)

	debugLog("Resolved config for 'ask': provider=%s, model=%s", cfg.Provider, cfg.Model)

	// Create context
	ctx := context.Background()

	// Get answer
	answer, err := ask.Answer(ctx, question, cfg, debugLog)
	if err != nil {
		return err
	}

	// Print the answer
	fmt.Fprintln(cmd.OutOrStdout(), answer)

	return nil
}
