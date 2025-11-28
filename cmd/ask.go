package cmd

import (
	"fmt"

	"github.com/connorhough/smix/internal/ask"
	"github.com/spf13/cobra"
)

// NewAskCmd creates and returns the ask command
func NewAskCmd() *cobra.Command {
	askCmd := &cobra.Command{
		Use:   "ask \"your question\"",
		Short: "Ask short technical questions and get concise answers",
		Long: `Ask short technical questions and get concise answers using Claude Code CLI.

Great for quick lookups like:
- "what is FastAPI"
- "does the mv command overwrite duplicate files"
- "how do I check if a port is open"

Requirements:
- claude CLI installed (Claude Code from https://claude.ai/code)`,
		Args: cobra.ExactArgs(1),
		RunE: runAsk,
	}

	return askCmd
}

func runAsk(cmd *cobra.Command, args []string) error {
	question := args[0]

	// Use internal package to get answer
	answer, err := ask.Answer(question)
	if err != nil {
		return err
	}

	// Print the answer
	fmt.Fprintln(cmd.OutOrStdout(), answer)

	return nil
}
