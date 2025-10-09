package cmd

import (
	"fmt"

	"github.com/connorhough/smix/internal/do"
	"github.com/spf13/cobra"
)

// NewDoCmd creates and returns the do command
func NewDoCmd() *cobra.Command {
	doCmd := &cobra.Command{
		Use:   "do \"natural language task description\"",
		Short: "Translate natural language to shell commands",
		Long: `Translate natural language task descriptions into functional shell commands.

The CEREBRAS_API_KEY environment variable is required to be set before running this command.`,
		Args: cobra.ExactArgs(1),
		RunE: runDo,
	}

	return doCmd
}

func runDo(cmd *cobra.Command, args []string) error {
	taskDescription := args[0]

	// Use internal package to translate natural language to shell command
	shellCommand, err := do.Translate(taskDescription)
	if err != nil {
		return err
	}

	// Print the resulting shell command
	fmt.Println(shellCommand)

	return nil
}
